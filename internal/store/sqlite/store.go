package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type Store struct {
	db *sql.DB
}

func New(dsn string) (*Store, error) {
	// Activa foreign_keys en CADA conexión del pool; de otro modo se
	// perdería al reciclarse una conexión (p. ej. por ConnMaxLifetime).
	db, err := sql.Open("sqlite", withForeignKeysPragma(dsn))
	if err != nil {
		return nil, fmt.Errorf("sqlite: open: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Hour)

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("sqlite: migrate: %w", err)
	}
	return s, nil
}

// withForeignKeysPragma agrega el pragma por conexión al DSN, respetando
// parámetros previos si el DSN ya era una URI con query string.
func withForeignKeysPragma(dsn string) string {
	if strings.Contains(dsn, "?") {
		return dsn + "&_pragma=foreign_keys(1)"
	}
	return dsn + "?_pragma=foreign_keys(1)"
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) DB() *sql.DB {
	return s.db
}

func (s *Store) migrate() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Registro de versiones aplicadas (idempotente sobre bases nuevas y heredadas).
	if _, err := s.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version TEXT PRIMARY KEY,
		applied_at TEXT NOT NULL DEFAULT (datetime('now'))
	)`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	// v1: esquema inicial; las bases heredadas ya pueden tenerlo.
	var count int
	if err := s.db.QueryRowContext(ctx, "SELECT count(*) FROM sqlite_master WHERE type='table' AND name='empleados'").Scan(&count); err != nil {
		return fmt.Errorf("check tables: %w", err)
	}
	if count == 0 {
		migration, err := migrationsFS.ReadFile("migrations/000001_initial_schema.up.sql")
		if err != nil {
			return fmt.Errorf("read migration 000001: %w", err)
		}
		if _, err := s.db.ExecContext(ctx, string(migration)); err != nil {
			return fmt.Errorf("exec migration 000001: %w", err)
		}
	}
	if err := s.recordMigration(ctx, "000001_initial_schema"); err != nil {
		return err
	}

	// Parche heredado v2: columna tipo en historial_estimulos (bases viejas).
	var tipoExists int
	if err := s.db.QueryRowContext(ctx, "SELECT count(*) FROM pragma_table_info('historial_estimulos') WHERE name='tipo'").Scan(&tipoExists); err == nil && tipoExists == 0 {
		if _, err := s.db.ExecContext(ctx, "ALTER TABLE historial_estimulos ADD COLUMN tipo TEXT NOT NULL DEFAULT ''"); err != nil {
			return fmt.Errorf("add tipo column: %w", err)
		}
	}

	// v2: constraints de integridad (FK, CHECK, índices) con backfill previo.
	if err := s.applyPending(ctx, "000002_hardening"); err != nil {
		return err
	}

	return nil
}

func (s *Store) recordMigration(ctx context.Context, version string) error {
	if _, err := s.db.ExecContext(ctx, `INSERT OR IGNORE INTO schema_migrations (version) VALUES (?)`, version); err != nil {
		return fmt.Errorf("record migration %s: %w", version, err)
	}
	return nil
}

// applyPending ejecuta una migración forward solo si su versión no está
// registrada y valida con PRAGMA foreign_key_check que no queden FKs rotas.
func (s *Store) applyPending(ctx context.Context, version string) error {
	var applied int
	if err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM schema_migrations WHERE version=?`, version).Scan(&applied); err != nil {
		return fmt.Errorf("check migration %s: %w", version, err)
	}
	if applied > 0 {
		return nil
	}

	migration, err := migrationsFS.ReadFile("migrations/" + version + ".up.sql")
	if err != nil {
		return fmt.Errorf("read migration %s: %w", version, err)
	}
	if _, err := s.db.ExecContext(ctx, string(migration)); err != nil {
		return fmt.Errorf("exec migration %s: %w", version, err)
	}

	rows, err := s.db.QueryContext(ctx, "PRAGMA foreign_key_check")
	if err != nil {
		return fmt.Errorf("foreign_key_check %s: %w", version, err)
	}
	var broken []string
	for rows.Next() {
		var table string
		var rowid int64
		var parent string
		var fkid int64
		if err := rows.Scan(&table, &rowid, &parent, &fkid); err != nil {
			rows.Close()
			return fmt.Errorf("scan foreign_key_check: %w", err)
		}
		broken = append(broken, fmt.Sprintf("%s(rowid=%d) ref %s(fk#%d)", table, rowid, parent, fkid))
	}
	rows.Close()
	if len(broken) > 0 {
		return fmt.Errorf("migration %s left broken foreign keys: %v", version, broken)
	}

	return s.recordMigration(ctx, version)
}

func (s *Store) WithTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}

// execer es el subconjunto de *sql.DB y *sql.Tx usado para escrituras,
// lo que permite compartir la lógica de inserción entre ambos.
type execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// queryer es el subconjunto de *sql.DB y *sql.Tx usado para lecturas.
type queryer interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

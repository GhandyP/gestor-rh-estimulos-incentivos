package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type Store struct {
	db *sql.DB
}

func New(dsn string) (*Store, error) {
	db, err := sql.Open("sqlite", dsn)
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

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) DB() *sql.DB {
	return s.db
}

func (s *Store) migrate() error {
	migration, err := migrationsFS.ReadFile("migrations/000001_initial_schema.up.sql")
	if err != nil {
		return fmt.Errorf("read migration: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var count int
	if err := s.db.QueryRowContext(ctx, "SELECT count(*) FROM sqlite_master WHERE type='table' AND name='empleados'").Scan(&count); err != nil {
		return fmt.Errorf("check tables: %w", err)
	}

	if count == 0 {
		if _, err := s.db.ExecContext(ctx, string(migration)); err != nil {
			return fmt.Errorf("exec migration: %w", err)
		}
	}

	// v2: add tipo column to historial_estimulos for effectiveness analysis
	var tipoExists int
	if err := s.db.QueryRowContext(ctx, "SELECT count(*) FROM pragma_table_info('historial_estimulos') WHERE name='tipo'").Scan(&tipoExists); err == nil && tipoExists == 0 {
		s.db.ExecContext(ctx, "ALTER TABLE historial_estimulos ADD COLUMN tipo TEXT NOT NULL DEFAULT ''")
	}

	return nil
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

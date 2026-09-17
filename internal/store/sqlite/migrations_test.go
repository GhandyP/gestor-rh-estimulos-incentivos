package sqlite

import (
	"context"
	"database/sql"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"estimulos-incentivos/internal/domain"
	"estimulos-incentivos/internal/store"
)

// openLegacy abre una base con el esquema 000001 únicamente (sin hardening),
// como existía antes de la migración 000002.
func openLegacy(t *testing.T, dsn string) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatalf("open legacy db: %v", err)
	}
	up, err := migrationsFS.ReadFile("migrations/000001_initial_schema.up.sql")
	if err != nil {
		t.Fatalf("read 000001: %v", err)
	}
	if _, err := db.Exec(string(up)); err != nil {
		t.Fatalf("exec 000001: %v", err)
	}
	return db
}

func TestMigrate_AppliesOnceAndIsIdempotent(t *testing.T) {
	dsn := filepath.Join(t.TempDir(), "idem.db")

	s1, err := New(dsn)
	if err != nil {
		t.Fatalf("first New: %v", err)
	}
	versions := []string{}
	rows, err := s1.db.Query("SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		t.Fatalf("list versions: %v", err)
	}
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			t.Fatalf("scan version: %v", err)
		}
		versions = append(versions, v)
	}
	rows.Close()
	if len(versions) != 3 || versions[0] != "000001_initial_schema" || versions[1] != "000002_hardening" || versions[2] != "000003_nudge_target_depto" {
		t.Fatalf("applied versions = %v, want [000001_initial_schema 000002_hardening 000003_nudge_target_depto]", versions)
	}
	s1.Close()

	// Reabrir: no debe fallar ni reaplicar.
	s2, err := New(dsn)
	if err != nil {
		t.Fatalf("reopen New: %v", err)
	}
	defer s2.Close()

	var fk int
	if err := s2.db.QueryRow("PRAGMA foreign_keys").Scan(&fk); err != nil {
		t.Fatalf("read foreign_keys pragma: %v", err)
	}
	if fk != 1 {
		t.Fatalf("foreign_keys pragma = %d, want 1 (activado por conexión)", fk)
	}
	var still int
	if err := s2.db.QueryRow("SELECT count(*) FROM schema_migrations WHERE version='000002_hardening'").Scan(&still); err != nil {
		t.Fatalf("count versions: %v", err)
	}
	if still != 1 {
		t.Fatalf("000002_hardening applied twice? count = %d, want 1", still)
	}
}

func TestMigrate_EnforcesForeignKeys(t *testing.T) {
	dsn := filepath.Join(t.TempDir(), "fk.db")
	ctx := context.Background()

	// RED: sobre el esquema heredado (FKs desactivadas) el insert huérfano entra.
	legacy := openLegacy(t, dsn)
	if _, err := legacy.ExecContext(ctx, `INSERT INTO perfiles_map
		(empleado_id, motivacion, habilidad, prompt, sensibilidad, confiabilidad, updated_at)
		VALUES (999, 0.5, 0.5, 0.5, 'desarrollo', 0.3, '2026-01-01T00:00:00Z')`); err != nil {
		t.Fatalf("premise: legacy insert with missing parent failed: %v", err)
	}
	legacy.Close()

	// GREEN: tras la migración la misma operación es rechazada por la FK.
	s, err := New(dsn)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer s.Close()

	if _, err := s.db.ExecContext(ctx, `INSERT INTO perfiles_map
		(empleado_id, motivacion, habilidad, prompt, sensibilidad, confiabilidad, updated_at)
		VALUES (998, 0.5, 0.5, 0.5, 'desarrollo', 0.3, '2026-01-01T00:00:00Z')`); err == nil {
		t.Fatal("post-migration insert with missing parent succeeded, want FOREIGN KEY error")
	} else if !strings.Contains(err.Error(), "FOREIGN KEY") {
		t.Fatalf("post-migration error = %q, want FOREIGN KEY constraint error", err.Error())
	}
}

func TestMigrate_EnforcesRangeChecks(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	e := testEmpleado()
	if err := s.CreateEmpleado(ctx, e); err != nil {
		t.Fatalf("create empleado: %v", err)
	}

	if _, err := s.db.ExecContext(ctx, `INSERT INTO umbrales
		(empleado_id, umbral_absoluto, umbral_diferencial, updated_at)
		VALUES (?, 2.5, 0.15, '2026-01-01T00:00:00Z')`, e.ID); err == nil {
		t.Fatal("umbral_absoluto=2.5 accepted, want CHECK constraint error")
	}
}

func TestMigrate_BackfillRemovesOrphansKeepsValidRows(t *testing.T) {
	dsn := filepath.Join(t.TempDir(), "backfill.db")
	ctx := context.Background()

	legacy := openLegacy(t, dsn)
	// Fila válida que debe sobrevivir.
	if _, err := legacy.ExecContext(ctx, `INSERT INTO empleados
		(nombre, email, cargo, departamento, fecha_ingreso, activo)
		VALUES ('Ana López', 'ana@empresa.com', 'Tech Lead', 'Ingeniería', '2026-01-01T00:00:00Z', 1)`); err != nil {
		t.Fatalf("seed valid empleado: %v", err)
	}
	// Huérfanos (padres inexistentes) que el backfill debe eliminar.
	for _, q := range []string{
		`INSERT INTO perfiles_map (empleado_id, motivacion, habilidad, prompt, sensibilidad, confiabilidad, updated_at)
		 VALUES (999, 0.5, 0.5, 0.5, 'desarrollo', 0.3, '2026-01-01T00:00:00Z')`,
		`INSERT INTO umbrales (empleado_id, umbral_absoluto, umbral_diferencial, updated_at)
		 VALUES (999, 0.3, 0.15, '2026-01-01T00:00:00Z')`,
		`INSERT INTO estimulos (empleado_id, tipo, contenido, intensidad, canal, estado, fecha_ideal, created_at)
		 VALUES (999, 'tipo', 'contenido', 0.5, 'email', 'pendiente', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z')`,
	} {
		if _, err := legacy.ExecContext(ctx, q); err != nil {
			t.Fatalf("seed orphan row: %v", err)
		}
	}
	legacy.Close()

	s, err := New(dsn)
	if err != nil {
		t.Fatalf("New (runs migration with backfill): %v", err)
	}
	defer s.Close()

	for _, orphan := range []struct {
		table string
		id    int64
	}{
		{"perfiles_map", 999},
		{"umbrales", 999},
		{"estimulos", 999},
	} {
		var n int
		if err := s.db.QueryRowContext(ctx,
			"SELECT count(*) FROM "+orphan.table+" WHERE empleado_id=?", orphan.id).Scan(&n); err != nil {
			t.Fatalf("count orphans %s: %v", orphan.table, err)
		}
		if n != 0 {
			t.Fatalf("orphan %s(empleado_id=%d) survived backfill: %d rows", orphan.table, orphan.id, n)
		}
	}

	var valid int
	if err := s.db.QueryRowContext(ctx,
		"SELECT count(*) FROM empleados WHERE email='ana@empresa.com'").Scan(&valid); err != nil {
		t.Fatalf("count valid rows: %v", err)
	}
	if valid != 1 {
		t.Fatalf("valid empleado lost by backfill: %d rows, want 1", valid)
	}
}

func TestMigrate_DownPreservesDataAndDropsConstraints(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	// Datos creados por el flujo real (tx-aware).
	e := testEmpleado()
	if err := s.WithTx(ctx, func(tx store.Transaction) error {
		if err := tx.CreateEmpleado(ctx, e); err != nil {
			return err
		}
		if err := tx.CreatePerfilMAP(ctx, testPerfil(e.ID)); err != nil {
			return err
		}
		return tx.CreateUmbral(ctx, testUmbral(e.ID))
	}); err != nil {
		t.Fatalf("seed workflow: %v", err)
	}

	down, err := migrationsFS.ReadFile("migrations/000002_hardening.down.sql")
	if err != nil {
		t.Fatalf("read down migration: %v", err)
	}
	if _, err := s.db.ExecContext(ctx, string(down)); err != nil {
		t.Fatalf("exec down migration: %v", err)
	}

	// Índices nuevos eliminados.
	for _, idx := range []string{"idx_historial_umbral_fecha", "idx_estimulos_estado_fecha_ideal", "idx_estimulos_empleado_estado"} {
		var n int
		if err := s.db.QueryRowContext(ctx,
			"SELECT count(*) FROM sqlite_master WHERE type='index' AND name=?", idx).Scan(&n); err != nil {
			t.Fatalf("check index %s: %v", idx, err)
		}
		if n != 0 {
			t.Fatalf("index %s survived down migration", idx)
		}
	}

	// CHECK eliminado: un valor fuera de rango vuelve a ser aceptado.
	other := testEmpleado()
	other.Email = "otro@empresa.com"
	if err := s.CreateEmpleado(ctx, other); err != nil {
		t.Fatalf("create segundo empleado: %v", err)
	}
	if _, err := s.db.ExecContext(ctx, `INSERT INTO umbrales
		(empleado_id, umbral_absoluto, umbral_diferencial, updated_at)
		VALUES (?, 2.5, 0.15, '2026-01-01T00:00:00Z')`, other.ID); err != nil {
		t.Fatalf("post-down out-of-range insert failed: %v (CHECK debería haber desaparecido)", err)
	}

	// Datos preservados.
	var empleados, perfiles, umbrales int
	s.db.QueryRowContext(ctx, "SELECT count(*) FROM empleados").Scan(&empleados)
	s.db.QueryRowContext(ctx, "SELECT count(*) FROM perfiles_map").Scan(&perfiles)
	s.db.QueryRowContext(ctx, "SELECT count(*) FROM umbrales").Scan(&umbrales)
	if empleados != 2 || perfiles != 1 || umbrales != 2 {
		t.Fatalf("data after down = empleados:%d perfiles:%d umbrales:%d, want 2/1/2", empleados, perfiles, umbrales)
	}
}

var _ = domain.SensibilidadDesarrollo
var _ = time.Now

package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"estimulos-incentivos/internal/domain"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	dsn := filepath.Join(t.TempDir(), "test.db")
	s, err := New(dsn)
	if err != nil {
		t.Fatalf("New(%s): %v", dsn, err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func testEmpleado() *domain.Empleado {
	return &domain.Empleado{
		Nombre:       "María García",
		Email:        "maria@empresa.com",
		Cargo:        "Senior Developer",
		Departamento: domain.Departamento("Ingeniería"),
		Activo:       true,
	}
}

func testPerfil(empleadoID int64) *domain.PerfilMAP {
	return &domain.PerfilMAP{
		EmpleadoID:    empleadoID,
		Motivacion:    0.5,
		Habilidad:     0.5,
		Prompt:        0.6,
		Sensibilidad:  domain.SensibilidadDesarrollo,
		Confiabilidad: 0.3,
	}
}

func testUmbral(empleadoID int64) *domain.Umbral {
	return &domain.Umbral{
		EmpleadoID:        empleadoID,
		UmbralAbsoluto:    0.30,
		UmbralDiferencial: 0.15,
	}
}

func TestWithTxCommitPersists(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	e := testEmpleado()
	if err := s.WithTx(ctx, func(tx *sql.Tx) error {
		return s.CreateEmpleadoTx(ctx, tx, e)
	}); err != nil {
		t.Fatalf("WithTx: %v", err)
	}

	got, err := s.GetEmpleado(ctx, e.ID)
	if err != nil || got == nil {
		t.Fatalf("GetEmpleado after commit = %v, %v; want persisted", got, err)
	}
	if got.Email != e.Email {
		t.Fatalf("GetEmpleado.Email = %q, want %q", got.Email, e.Email)
	}
}

func TestWithTxRollbackOnError(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	boom := errors.New("injected failure")
	err := s.WithTx(ctx, func(tx *sql.Tx) error {
		e := testEmpleado()
		if err := s.CreateEmpleadoTx(ctx, tx, e); err != nil {
			return err
		}
		if err := s.CreatePerfilMAPTx(ctx, tx, testPerfil(e.ID)); err != nil {
			return err
		}
		return boom
	})
	if !errors.Is(err, boom) {
		t.Fatalf("WithTx error = %v, want %v", err, boom)
	}

	var count int
	if err := s.db.QueryRowContext(ctx, "SELECT count(*) FROM empleados").Scan(&count); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 0 {
		t.Fatalf("empleados after rollback = %d, want 0", count)
	}
}

func TestCreateEmpleadoConPerfilRollsBackOnPerfilFailure(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	if _, err := s.db.ExecContext(ctx, `CREATE TRIGGER fail_perfil
		BEFORE INSERT ON perfiles_map BEGIN SELECT RAISE(ABORT, 'injected'); END;`); err != nil {
		t.Fatalf("create trigger: %v", err)
	}

	err := s.CreateEmpleadoConPerfil(ctx, testEmpleado(), testPerfil(0))
	if err == nil {
		t.Fatal("CreateEmpleadoConPerfil() = nil, want injected error")
	}

	var count int
	if err := s.db.QueryRowContext(ctx, "SELECT count(*) FROM empleados").Scan(&count); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 0 {
		t.Fatalf("empleados after failed workflow = %d, want 0 (no partial state)", count)
	}
}

func TestTxWorkflowPersistsAllEntities(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	e := testEmpleado()
	if err := s.WithTx(ctx, func(tx *sql.Tx) error {
		if err := s.CreateEmpleadoTx(ctx, tx, e); err != nil {
			return err
		}
		if err := s.CreatePerfilMAPTx(ctx, tx, testPerfil(e.ID)); err != nil {
			return err
		}
		return s.CreateUmbralTx(ctx, tx, testUmbral(e.ID))
	}); err != nil {
		t.Fatalf("WithTx workflow: %v", err)
	}

	perfil, err := s.GetPerfilMAP(ctx, e.ID)
	if err != nil || perfil == nil {
		t.Fatalf("perfil not persisted: %v, %v", perfil, err)
	}
	umbral, err := s.GetUmbral(ctx, e.ID)
	if err != nil || umbral == nil {
		t.Fatalf("umbral not persisted: %v, %v", umbral, err)
	}
	if umbral.UmbralAbsoluto != 0.30 {
		t.Fatalf("umbral.UmbralAbsoluto = %v, want 0.30", umbral.UmbralAbsoluto)
	}
}

func TestTransitionEstimuloTxConditional(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	e := testEmpleado()
	if err := s.CreateEmpleado(ctx, e); err != nil {
		t.Fatalf("create empleado: %v", err)
	}
	est := &domain.Estimulo{
		EmpleadoID:          e.ID,
		Tipo:                "recomendacion_personalizada",
		Contenido:           "Oportunidad de desarrollo profesional",
		Intensidad:          0.6,
		Canal:               domain.CanalEmail,
		Estado:              domain.EstadoPendiente,
		FechaIdeal:          time.Now(),
		OrigenRecomendacion: "motor MAP",
	}
	if err := s.CreateEstimulo(ctx, est); err != nil {
		t.Fatalf("create estimulo: %v", err)
	}

	// First transition wins.
	var first bool
	if err := s.WithTx(ctx, func(tx *sql.Tx) error {
		var err error
		first, err = s.TransitionEstimuloTx(ctx, tx, est.ID, time.Now().UTC())
		return err
	}); err != nil {
		t.Fatalf("first transition: %v", err)
	}
	if !first {
		t.Fatal("first TransitionEstimuloTx = false, want true")
	}

	got, err := s.GetEstimulo(ctx, est.ID)
	if err != nil || got == nil {
		t.Fatalf("get estimulo: %v, %v", got, err)
	}
	if got.Estado != domain.EstadoAplicado {
		t.Fatalf("estimulo estado = %q, want %q", got.Estado, domain.EstadoAplicado)
	}
	if got.FechaAplicado == nil {
		t.Fatal("estimulo fecha_aplicado = nil, want set")
	}

	// Second transition loses.
	var second bool
	if err := s.WithTx(ctx, func(tx *sql.Tx) error {
		var err error
		second, err = s.TransitionEstimuloTx(ctx, tx, est.ID, time.Now().UTC())
		return err
	}); err != nil {
		t.Fatalf("second transition: %v", err)
	}
	if second {
		t.Fatal("second TransitionEstimuloTx = true, want false (already applied)")
	}
}

func TestTransitionEstimuloTxMissingStimulus(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	var applied bool
	if err := s.WithTx(ctx, func(tx *sql.Tx) error {
		var err error
		applied, err = s.TransitionEstimuloTx(ctx, tx, 999, time.Now().UTC())
		return err
	}); err != nil {
		t.Fatalf("transition: %v", err)
	}
	if applied {
		t.Fatal("TransitionEstimuloTx for missing id = true, want false")
	}
}

func TestUmbralHistorialTxVariants(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	e := testEmpleado()
	var umbralID int64
	if err := s.WithTx(ctx, func(tx *sql.Tx) error {
		if err := s.CreateEmpleadoTx(ctx, tx, e); err != nil {
			return err
		}
		u := testUmbral(e.ID)
		if err := s.CreateUmbralTx(ctx, tx, u); err != nil {
			return err
		}
		umbralID = u.ID
		if err := s.AddHistorialTx(ctx, tx, u.ID, domain.PuntoHistorial{
			Fecha:        time.Now(),
			Intensidad:   0.4,
			RespuestaMAP: 0.7,
			Tipo:         "recomendacion_personalizada",
		}); err != nil {
			return err
		}
		u2, err := s.GetUmbralTx(ctx, tx, e.ID)
		if err != nil {
			return err
		}
		u2.UltimoEstimulo = 0.4
		now := time.Now()
		u2.FechaUltimoEstimulo = &now
		return s.UpdateUmbralTx(ctx, tx, u2)
	}); err != nil {
		t.Fatalf("tx workflow: %v", err)
	}

	historial, err := s.GetHistorial(ctx, umbralID)
	if err != nil {
		t.Fatalf("get historial: %v", err)
	}
	if len(historial) != 1 {
		t.Fatalf("historial length = %d, want 1", len(historial))
	}
	if historial[0].Intensidad != 0.4 || historial[0].RespuestaMAP != 0.7 {
		t.Fatalf("historial[0] = %+v, want intensidad 0.4, respuesta 0.7", historial[0])
	}

	umbral, err := s.GetUmbral(ctx, e.ID)
	if err != nil || umbral == nil {
		t.Fatalf("get umbral: %v, %v", umbral, err)
	}
	if umbral.UltimoEstimulo != 0.4 {
		t.Fatalf("umbral.UltimoEstimulo = %v, want 0.4", umbral.UltimoEstimulo)
	}
}

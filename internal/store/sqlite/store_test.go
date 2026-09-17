package sqlite

import (
	"context"
	"errors"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"estimulos-incentivos/internal/domain"
	"estimulos-incentivos/internal/store"
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
	if err := s.WithTx(ctx, func(tx store.Transaction) error {
		return tx.CreateEmpleado(ctx, e)
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
	err := s.WithTx(ctx, func(tx store.Transaction) error {
		e := testEmpleado()
		if err := tx.CreateEmpleado(ctx, e); err != nil {
			return err
		}
		if err := tx.CreatePerfilMAP(ctx, testPerfil(e.ID)); err != nil {
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
	if err := s.WithTx(ctx, func(tx store.Transaction) error {
		if err := tx.CreateEmpleado(ctx, e); err != nil {
			return err
		}
		if err := tx.CreatePerfilMAP(ctx, testPerfil(e.ID)); err != nil {
			return err
		}
		return tx.CreateUmbral(ctx, testUmbral(e.ID))
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
	if err := s.WithTx(ctx, func(tx store.Transaction) error {
		var err error
		first, err = tx.TransitionEstimulo(ctx, est.ID, time.Now().UTC())
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
	if err := s.WithTx(ctx, func(tx store.Transaction) error {
		var err error
		second, err = tx.TransitionEstimulo(ctx, est.ID, time.Now().UTC())
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
	if err := s.WithTx(ctx, func(tx store.Transaction) error {
		var err error
		applied, err = tx.TransitionEstimulo(ctx, 999, time.Now().UTC())
		return err
	}); err != nil {
		t.Fatalf("transition: %v", err)
	}
	if applied {
		t.Fatal("TransitionEstimuloTx for missing id = true, want false")
	}
}

// Spec: "Concurrent application has one winner" — N goroutines intentan la
// transición del MISMO estímulo en transacciones propias (una conexión
// serializa los writers): exactamente una gana, el resto reporta conflicto,
// y no hay historial duplicado. El setup es real SQLite (temp DB con
// migraciones), no mocks.
func TestTransitionEstimuloTxConcurrentOneWinner(t *testing.T) {
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
	u := testUmbral(e.ID)
	if err := s.CreateUmbral(ctx, u); err != nil {
		t.Fatalf("create umbral: %v", err)
	}

	const workers = 8
	start := make(chan struct{})
	results := make([]bool, workers)
	errs := make([]error, workers)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			results[i], errs[i] = func() (bool, error) {
				var applied bool
				err := s.WithTx(ctx, func(tx store.Transaction) error {
					var err error
					applied, err = tx.TransitionEstimulo(ctx, est.ID, time.Now().UTC())
					if err != nil {
						return err
					}
					if !applied {
						return nil
					}
					// Solo el ganador escribe el historial, como hace el service.
					return tx.AddHistorial(ctx, u.ID, domain.PuntoHistorial{
						Fecha:        time.Now(),
						Intensidad:   0.6,
						RespuestaMAP: 0.7,
						Tipo:         est.Tipo,
					})
				})
				return applied, err
			}()
		}(i)
	}
	close(start)
	wg.Wait()

	winners, conflicts, failures := 0, 0, 0
	for i := 0; i < workers; i++ {
		switch {
		case errs[i] != nil:
			failures++
		case results[i]:
			winners++
		default:
			conflicts++
		}
	}
	if failures != 0 {
		t.Fatalf("concurrent transitions errors = %d (%v), want 0", failures, errs)
	}
	if winners != 1 {
		t.Fatalf("winners = %d, want exactly 1", winners)
	}
	if conflicts != workers-1 {
		t.Fatalf("conflicts = %d, want %d", conflicts, workers-1)
	}

	got, err := s.GetEstimulo(ctx, est.ID)
	if err != nil || got == nil {
		t.Fatalf("get estimulo: %v, %v", got, err)
	}
	if got.Estado != domain.EstadoAplicado {
		t.Fatalf("estimulo estado = %q, want %q", got.Estado, domain.EstadoAplicado)
	}

	historial, err := s.GetHistorial(ctx, u.ID)
	if err != nil {
		t.Fatalf("get historial: %v", err)
	}
	if len(historial) != 1 {
		t.Fatalf("historial = %d entries, want exactly 1 (los perdedores no escriben)", len(historial))
	}
}

// Spec: "SQLite failures prove rollback" — el flujo completo (empleado +
// perfil + umbral) falla DESPUÉS de tres escrituras exitosas; al reabrir la
// base no queda NINGÚN estado parcial en ninguna de las tres tablas.
func TestMultiEntityWorkflowRollbackLeavesNoPartialState(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	boom := errors.New("injected failure after third write")
	err := s.WithTx(ctx, func(tx store.Transaction) error {
		e := testEmpleado()
		if err := tx.CreateEmpleado(ctx, e); err != nil {
			return err
		}
		if err := tx.CreatePerfilMAP(ctx, testPerfil(e.ID)); err != nil {
			return err
		}
		if err := tx.CreateUmbral(ctx, testUmbral(e.ID)); err != nil {
			return err
		}
		return boom
	})
	if !errors.Is(err, boom) {
		t.Fatalf("WithTx error = %v, want %v", err, boom)
	}

	for _, table := range []string{"empleados", "perfiles_map", "umbrales"} {
		var count int
		if err := s.db.QueryRowContext(ctx, "SELECT count(*) FROM "+table).Scan(&count); err != nil {
			t.Fatalf("count %s: %v", table, err)
		}
		if count != 0 {
			t.Fatalf("%s after failed workflow = %d, want 0 (no partial state)", table, count)
		}
	}
}

func TestUmbralHistorialTxVariants(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	e := testEmpleado()
	var umbralID int64
	if err := s.WithTx(ctx, func(tx store.Transaction) error {
		if err := tx.CreateEmpleado(ctx, e); err != nil {
			return err
		}
		u := testUmbral(e.ID)
		if err := tx.CreateUmbral(ctx, u); err != nil {
			return err
		}
		umbralID = u.ID
		if err := tx.AddHistorial(ctx, u.ID, domain.PuntoHistorial{
			Fecha:        time.Now(),
			Intensidad:   0.4,
			RespuestaMAP: 0.7,
			Tipo:         "recomendacion_personalizada",
		}); err != nil {
			return err
		}
		u2, err := tx.GetUmbral(ctx, e.ID)
		if err != nil {
			return err
		}
		u2.UltimoEstimulo = 0.4
		now := time.Now()
		u2.FechaUltimoEstimulo = &now
		return tx.UpdateUmbral(ctx, u2)
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

func TestCountEmpleados(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	count, err := s.CountEmpleados(ctx)
	if err != nil {
		t.Fatalf("CountEmpleados on empty store: %v", err)
	}
	if count != 0 {
		t.Fatalf("CountEmpleados on empty store = %d, want 0", count)
	}

	first := testEmpleado()
	second := testEmpleado()
	second.Email = "juan@empresa.com"
	for _, e := range []*domain.Empleado{first, second} {
		if err := s.CreateEmpleado(ctx, e); err != nil {
			t.Fatalf("CreateEmpleado: %v", err)
		}
	}

	count, err = s.CountEmpleados(ctx)
	if err != nil {
		t.Fatalf("CountEmpleados: %v", err)
	}
	if count != 2 {
		t.Fatalf("CountEmpleados = %d, want 2", count)
	}
}

func TestUpdateNudgePersistsChanges(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	n := &domain.Nudge{
		Nombre:      "Original",
		Descripcion: "Descripción original",
		Tipo:        domain.NudgeDefaults,
		Ambito:      domain.AmbitoGlobal,
		Activo:      true,
	}
	if err := s.CreateNudge(ctx, n); err != nil {
		t.Fatalf("CreateNudge: %v", err)
	}

	n.Nombre = "Actualizado"
	n.Descripcion = "Descripción actualizada"
	n.Activo = false
	if err := s.UpdateNudge(ctx, n); err != nil {
		t.Fatalf("UpdateNudge: %v", err)
	}

	got, err := s.GetNudge(ctx, n.ID)
	if err != nil || got == nil {
		t.Fatalf("GetNudge after update = %v, %v", got, err)
	}
	if got.Nombre != n.Nombre || got.Descripcion != n.Descripcion || got.Activo != n.Activo {
		t.Fatalf("updated nudge = %+v, want name=%q description=%q active=%t", got, n.Nombre, n.Descripcion, n.Activo)
	}
}

func TestNudgeTargetDeptoRoundtrip(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	n := &domain.Nudge{
		Nombre:      "Defaults Ventas",
		Descripcion: "Defaults del departamento Ventas",
		Tipo:        domain.NudgeDefaults,
		Ambito:      domain.AmbitoDepartamento,
		TargetDepto: "Ventas",
		Activo:      true,
	}
	if err := s.CreateNudge(ctx, n); err != nil {
		t.Fatalf("CreateNudge: %v", err)
	}

	got, err := s.GetNudge(ctx, n.ID)
	if err != nil || got == nil {
		t.Fatalf("GetNudge = %v, %v", got, err)
	}
	if got.TargetDepto != "Ventas" || got.Ambito != domain.AmbitoDepartamento {
		t.Fatalf("GetNudge TargetDepto = %q Ambito = %q, want Ventas/departamento", got.TargetDepto, got.Ambito)
	}

	listed, err := s.ListNudges(ctx)
	if err != nil {
		t.Fatalf("ListNudges: %v", err)
	}
	if len(listed) != 1 || listed[0].TargetDepto != "Ventas" {
		t.Fatalf("ListNudges = %+v, want TargetDepto Ventas", listed)
	}

	n.TargetDepto = "Marketing"
	if err := s.UpdateNudge(ctx, n); err != nil {
		t.Fatalf("UpdateNudge: %v", err)
	}
	got, err = s.GetNudge(ctx, n.ID)
	if err != nil || got == nil {
		t.Fatalf("GetNudge after update = %v, %v", got, err)
	}
	if got.TargetDepto != "Marketing" {
		t.Fatalf("TargetDepto after update = %q, want Marketing", got.TargetDepto)
	}
}

func TestListEstimulosFiltersAndPreservesOrdering(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	empleado := testEmpleado()
	if err := s.CreateEmpleado(ctx, empleado); err != nil {
		t.Fatalf("CreateEmpleado: %v", err)
	}

	base := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	appliedAt := base.Add(5 * time.Hour)
	pendingLate := &domain.Estimulo{
		EmpleadoID: empleado.ID,
		Tipo:       "pending-late",
		Contenido:  "Late",
		Intensidad: 0.4,
		Canal:      domain.CanalEmail,
		Estado:     domain.EstadoPendiente,
		FechaIdeal: base.Add(3 * time.Hour),
	}
	pendingEarly := &domain.Estimulo{
		EmpleadoID: empleado.ID,
		Tipo:       "pending-early",
		Contenido:  "Early",
		Intensidad: 0.5,
		Canal:      domain.CanalEmail,
		Estado:     domain.EstadoPendiente,
		FechaIdeal: base.Add(1 * time.Hour),
	}
	applied := &domain.Estimulo{
		EmpleadoID:    empleado.ID,
		Tipo:          "applied",
		Contenido:     "Applied",
		Intensidad:    0.6,
		Canal:         domain.CanalEmail,
		Estado:        domain.EstadoAplicado,
		FechaIdeal:    base.Add(4 * time.Hour),
		FechaAplicado: &appliedAt,
	}
	stimuli := []*domain.Estimulo{pendingLate, pendingEarly, applied}
	createdAt := []time.Time{base.Add(1 * time.Hour), base.Add(2 * time.Hour), base.Add(3 * time.Hour)}
	for i, stimulus := range stimuli {
		if err := s.CreateEstimulo(ctx, stimulus); err != nil {
			t.Fatalf("CreateEstimulo(%s): %v", stimulus.Tipo, err)
		}
		if _, err := s.db.ExecContext(ctx, "UPDATE estimulos SET created_at=? WHERE id=?", createdAt[i].Format(time.RFC3339), stimulus.ID); err != nil {
			t.Fatalf("set created_at for %s: %v", stimulus.Tipo, err)
		}
	}

	checks := []struct {
		estado string
		want   []int64
	}{
		{"pendiente", []int64{pendingEarly.ID, pendingLate.ID}},
		{"aplicado", []int64{applied.ID}},
		{"todos", []int64{applied.ID, pendingEarly.ID, pendingLate.ID}},
		{"", []int64{applied.ID, pendingEarly.ID, pendingLate.ID}},
	}
	for _, check := range checks {
		got, err := s.ListEstimulos(ctx, check.estado)
		if err != nil {
			t.Fatalf("ListEstimulos(%q): %v", check.estado, err)
		}
		if len(got) != len(check.want) {
			t.Fatalf("ListEstimulos(%q) length = %d, want %d", check.estado, len(got), len(check.want))
		}
		for i, wantID := range check.want {
			if got[i].ID != wantID {
				t.Fatalf("ListEstimulos(%q)[%d].ID = %d, want %d", check.estado, i, got[i].ID, wantID)
			}
		}
	}
}

func TestTransactionStimulusTransitionRollsBack(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	empleado := testEmpleado()
	if err := s.CreateEmpleado(ctx, empleado); err != nil {
		t.Fatalf("CreateEmpleado: %v", err)
	}
	stimulus := &domain.Estimulo{
		EmpleadoID: empleado.ID,
		Tipo:       "recomendacion_personalizada",
		Contenido:  "Oportunidad de desarrollo profesional",
		Intensidad: 0.6,
		Canal:      domain.CanalEmail,
		Estado:     domain.EstadoPendiente,
		FechaIdeal: time.Now(),
	}
	if err := s.CreateEstimulo(ctx, stimulus); err != nil {
		t.Fatalf("CreateEstimulo: %v", err)
	}

	boom := errors.New("injected transition failure")
	var transitioned bool
	err := s.WithTx(ctx, func(tx store.Transaction) error {
		var err error
		transitioned, err = tx.TransitionEstimulo(ctx, stimulus.ID, time.Now().UTC())
		if err != nil {
			return err
		}
		return boom
	})
	if !errors.Is(err, boom) {
		t.Fatalf("WithTx error = %v, want %v", err, boom)
	}
	if !transitioned {
		t.Fatal("transaction transition = false, want true before rollback")
	}

	got, err := s.GetEstimulo(ctx, stimulus.ID)
	if err != nil || got == nil {
		t.Fatalf("GetEstimulo after rollback = %v, %v", got, err)
	}
	if got.Estado != domain.EstadoPendiente {
		t.Fatalf("stimulus state after rollback = %q, want %q", got.Estado, domain.EstadoPendiente)
	}
	if got.FechaAplicado != nil {
		t.Fatal("stimulus fecha_aplicado after rollback is set, want nil")
	}
}

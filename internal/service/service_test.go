package service

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"estimulos-incentivos/internal/domain"
	"estimulos-incentivos/internal/store/sqlite"
)

func newTestService(t *testing.T) (*Service, *sqlite.Store) {
	t.Helper()
	dsn := filepath.Join(t.TempDir(), "test.db")
	store, err := sqlite.New(dsn)
	if err != nil {
		t.Fatalf("sqlite.New(%s): %v", dsn, err)
	}
	t.Cleanup(func() { store.Close() })
	return New(store), store
}

func countRows(t *testing.T, s *sqlite.Store, table string) int {
	t.Helper()
	var n int
	if err := s.DB().QueryRow("SELECT count(*) FROM " + table).Scan(&n); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	return n
}

func seedPendiente(t *testing.T, s *sqlite.Store) int64 {
	t.Helper()
	ctx := context.Background()
	e := &domain.Empleado{
		Nombre:       "María García",
		Email:        "maria@empresa.com",
		Cargo:        "Senior Developer",
		Departamento: domain.Departamento("Ingeniería"),
		Activo:       true,
	}
	if err := s.CreateEmpleado(ctx, e); err != nil {
		t.Fatalf("seed empleado: %v", err)
	}
	if err := s.CreatePerfilMAP(ctx, &domain.PerfilMAP{
		EmpleadoID:    e.ID,
		Motivacion:    0.5,
		Habilidad:     0.5,
		Prompt:        0.6,
		Sensibilidad:  domain.SensibilidadDesarrollo,
		Confiabilidad: 0.3,
	}); err != nil {
		t.Fatalf("seed perfil: %v", err)
	}
	if err := s.CreateUmbral(ctx, &domain.Umbral{
		EmpleadoID:        e.ID,
		UmbralAbsoluto:    0.30,
		UmbralDiferencial: 0.15,
	}); err != nil {
		t.Fatalf("seed umbral: %v", err)
	}
	est := &domain.Estimulo{
		EmpleadoID:          e.ID,
		Tipo:                "recomendacion_personalizada",
		Contenido:           "Oportunidad de desarrollo profesional",
		Intensidad:          0.6,
		Canal:               domain.CanalEmail,
		Estado:              domain.EstadoPendiente,
		FechaIdeal:          time.Now().Add(24 * time.Hour),
		OrigenRecomendacion: "motor MAP",
	}
	if err := s.CreateEstimulo(ctx, est); err != nil {
		t.Fatalf("seed estimulo: %v", err)
	}
	return est.ID
}

func TestCreateEmpleadoRollsBackFullyOnIntermediateFailure(t *testing.T) {
	svc, store := newTestService(t)
	ctx := context.Background()

	// La falla intermedia es REAL: el umbral no se puede insertar.
	if _, err := store.DB().ExecContext(ctx, `CREATE TRIGGER fail_umbral
		BEFORE INSERT ON umbrales BEGIN SELECT RAISE(ABORT, 'injected'); END;`); err != nil {
		t.Fatalf("create trigger: %v", err)
	}

	_, err := svc.CreateEmpleado(ctx, "María García", "maria@empresa.com", "Senior Developer", "Ingeniería")
	if err == nil {
		t.Fatal("CreateEmpleado() = nil, want injected failure")
	}

	if n := countRows(t, store, "empleados"); n != 0 {
		t.Fatalf("empleados after failed workflow = %d, want 0 (rollback total)", n)
	}
	if n := countRows(t, store, "perfiles_map"); n != 0 {
		t.Fatalf("perfiles_map after failed workflow = %d, want 0", n)
	}
}

func TestCreateEmpleadoRejectsInvalidInputBeforePersist(t *testing.T) {
	svc, store := newTestService(t)
	ctx := context.Background()

	_, err := svc.CreateEmpleado(ctx, "", "email-sin-arroba", "Senior Developer", "Ingeniería")
	if err == nil {
		t.Fatal("CreateEmpleado(invalid) = nil, want validation error")
	}

	if n := countRows(t, store, "empleados"); n != 0 {
		t.Fatalf("empleados = %d, want 0 (validation before persistence)", n)
	}
}

func TestCreateEmpleadoPersistsAtomically(t *testing.T) {
	svc, store := newTestService(t)
	ctx := context.Background()

	e, err := svc.CreateEmpleado(ctx, "María García", "maria@empresa.com", "Senior Developer", "Ingeniería")
	if err != nil {
		t.Fatalf("CreateEmpleado: %v", err)
	}
	if e.ID == 0 {
		t.Fatal("CreateEmpleado returned id 0")
	}

	perfil, err := store.GetPerfilMAP(ctx, e.ID)
	if err != nil || perfil == nil {
		t.Fatalf("perfil not persisted: %v, %v", perfil, err)
	}
	umbral, err := store.GetUmbral(ctx, e.ID)
	if err != nil || umbral == nil {
		t.Fatalf("umbral not persisted: %v, %v", umbral, err)
	}
	if umbral.UmbralAbsoluto == 0 {
		t.Fatalf("umbral.UmbralAbsoluto = 0, want default inicial")
	}
}

func TestApplyEstimuloRepeatedHasOneWinner(t *testing.T) {
	svc, store := newTestService(t)
	ctx := context.Background()
	id := seedPendiente(t, store)

	first, err := svc.ApplyEstimulo(ctx, id, 0.7)
	if err != nil {
		t.Fatalf("first ApplyEstimulo: %v", err)
	}
	if !first.Applied || first.Conflict {
		t.Fatalf("first result = %+v, want Applied=true Conflict=false", first)
	}

	second, err := svc.ApplyEstimulo(ctx, id, 0.7)
	if err != nil {
		t.Fatalf("second ApplyEstimulo: %v", err)
	}
	if second.Applied || !second.Conflict {
		t.Fatalf("second result = %+v, want Applied=false Conflict=true", second)
	}

	// Exactamente un historial y una recalibración.
	est, err := store.GetEstimulo(ctx, id)
	if err != nil || est == nil {
		t.Fatalf("get estimulo: %v, %v", est, err)
	}
	if est.Estado != domain.EstadoAplicado {
		t.Fatalf("estimulo estado = %q, want %q", est.Estado, domain.EstadoAplicado)
	}
	if est.FechaAplicado == nil {
		t.Fatal("estimulo fecha_aplicado = nil, want set")
	}

	umbral, err := store.GetUmbral(ctx, est.EmpleadoID)
	if err != nil || umbral == nil {
		t.Fatalf("get umbral: %v, %v", umbral, err)
	}
	if umbral.UltimoEstimulo != est.Intensidad {
		t.Fatalf("umbral.UltimoEstimulo = %v, want %v (recalibrado una vez)", umbral.UltimoEstimulo, est.Intensidad)
	}
	if n := countRows(t, store, "historial_estimulos"); n != 1 {
		t.Fatalf("historial_estimulos = %d, want 1", n)
	}
}

func TestApplyEstimuloConcurrentHasExactlyOneWinner(t *testing.T) {
	svc, store := newTestService(t)
	ctx := context.Background()
	id := seedPendiente(t, store)

	const workers = 8
	start := make(chan struct{})
	results := make([]ApplyResult, workers)
	errs := make([]error, workers)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			results[i], errs[i] = svc.ApplyEstimulo(ctx, id, 0.7)
		}(i)
	}
	close(start)
	wg.Wait()

	applied, conflicts, failures := 0, 0, 0
	for i := 0; i < workers; i++ {
		switch {
		case errs[i] != nil:
			failures++
		case results[i].Applied && !results[i].Conflict:
			applied++
		case !results[i].Applied && results[i].Conflict:
			conflicts++
		}
	}
	if failures != 0 {
		t.Fatalf("concurrent apply errors = %d (%v), want 0", failures, errs)
	}
	if applied != 1 {
		t.Fatalf("winners = %d, want exactly 1", applied)
	}
	if conflicts != workers-1 {
		t.Fatalf("conflicts = %d, want %d", conflicts, workers-1)
	}
	if n := countRows(t, store, "historial_estimulos"); n != 1 {
		t.Fatalf("historial_estimulos = %d, want 1 (un solo ganador)", n)
	}
}

func TestApplyEstimuloRejectsInvalidRespuesta(t *testing.T) {
	svc, store := newTestService(t)
	ctx := context.Background()
	id := seedPendiente(t, store)

	if _, err := svc.ApplyEstimulo(ctx, id, 1.5); err == nil {
		t.Fatal("ApplyEstimulo(respuesta=1.5) = nil, want validation error")
	}

	est, err := store.GetEstimulo(ctx, id)
	if err != nil || est == nil {
		t.Fatalf("get estimulo: %v, %v", est, err)
	}
	if est.Estado != domain.EstadoPendiente {
		t.Fatalf("estimulo estado = %q, want %q (sin cambios)", est.Estado, domain.EstadoPendiente)
	}
	if n := countRows(t, store, "historial_estimulos"); n != 0 {
		t.Fatalf("historial_estimulos = %d, want 0", n)
	}
}

func TestApplyEstimuloMissingStimulusFails(t *testing.T) {
	svc, _ := newTestService(t)
	ctx := context.Background()

	if _, err := svc.ApplyEstimulo(ctx, 999, 0.7); err == nil {
		t.Fatal("ApplyEstimulo(missing id) = nil, want error")
	}
}

func TestSeedPersistsAllDataAndIsIdempotent(t *testing.T) {
	svc, store := newTestService(t)
	ctx := context.Background()

	if err := svc.Seed(ctx); err != nil {
		t.Fatalf("first Seed: %v", err)
	}

	want := []struct {
		table string
		count int
	}{
		{"empleados", 6},
		{"perfiles_map", 6},
		{"umbrales", 6},
		{"incentivos", 8},
		{"nudges", 8},
	}
	for _, check := range want {
		if got := countRows(t, store, check.table); got != check.count {
			t.Fatalf("%s after first Seed = %d, want %d", check.table, got, check.count)
		}
	}

	if err := svc.Seed(ctx); err != nil {
		t.Fatalf("second Seed: %v", err)
	}
	for _, check := range want {
		if got := countRows(t, store, check.table); got != check.count {
			t.Fatalf("%s after idempotent Seed = %d, want %d", check.table, got, check.count)
		}
	}
}

func TestSeedRollsBackEverySeedTableAfterIntermediateFailure(t *testing.T) {
	svc, store := newTestService(t)
	ctx := context.Background()

	// The trigger fires after all employees, MAP profiles, thresholds, and
	// incentives have already been written by the seed transaction.
	if _, err := store.DB().ExecContext(ctx, `CREATE TRIGGER fail_seed_nudge
		BEFORE INSERT ON nudges BEGIN SELECT RAISE(ABORT, 'injected seed failure'); END;`); err != nil {
		t.Fatalf("create trigger: %v", err)
	}

	if err := svc.Seed(ctx); err == nil {
		t.Fatal("Seed() = nil, want injected failure")
	}

	for _, table := range []string{"empleados", "perfiles_map", "umbrales", "incentivos", "nudges"} {
		if got := countRows(t, store, table); got != 0 {
			t.Fatalf("%s after failed Seed = %d, want 0 (complete rollback)", table, got)
		}
	}
}

func TestSeedConcurrentCallsCreateOneSeedSet(t *testing.T) {
	svc, store := newTestService(t)
	ctx := context.Background()

	const callers = 8
	start := make(chan struct{})
	errs := make([]error, callers)
	var wg sync.WaitGroup
	for i := 0; i < callers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			errs[i] = svc.Seed(ctx)
		}(i)
	}
	close(start)
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("concurrent Seed call %d: %v", i, err)
		}
	}
	for _, check := range []struct {
		table string
		count int
	}{
		{"empleados", 6},
		{"perfiles_map", 6},
		{"umbrales", 6},
		{"incentivos", 8},
		{"nudges", 8},
	} {
		if got := countRows(t, store, check.table); got != check.count {
			t.Fatalf("%s after concurrent Seed = %d, want %d", check.table, got, check.count)
		}
	}
}

func TestServiceProductionSourceStaysAbovePersistenceBoundary(t *testing.T) {
	_, testFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller() failed")
	}

	source, err := os.ReadFile(filepath.Join(filepath.Dir(testFile), "service.go"))
	if err != nil {
		t.Fatalf("read service.go: %v", err)
	}

	for _, forbidden := range []string{
		"database/sql",
		"internal/store/sqlite",
		"DB()",
		"*sql.Tx",
		"ExecContext",
		"QueryContext",
		"QueryRowContext",
		"SELECT ",
		"INSERT ",
		"UPDATE ",
		"DELETE ",
	} {
		if strings.Contains(string(source), forbidden) {
			t.Errorf("service.go contains persistence-boundary token %q", forbidden)
		}
	}
}

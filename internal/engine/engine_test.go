package engine

import (
	"testing"
	"time"

	"estimulos-incentivos/internal/domain"
)

func TestCalibrarUmbral_EmptyHistory(t *testing.T) {
	u := domain.Umbral{
		EmpleadoID: 1,
	}
	result := CalibrarUmbral(u, nil)

	if result.UmbralAbsoluto != defaultAbsoluteThreshold {
		t.Errorf("expected default absolute threshold %.2f, got %.2f", defaultAbsoluteThreshold, result.UmbralAbsoluto)
	}
	if result.UmbralDiferencial != defaultWeberConstant {
		t.Errorf("expected default weber constant %.2f, got %.2f", defaultWeberConstant, result.UmbralDiferencial)
	}
}

func TestCalibrarUmbral_WithHistory(t *testing.T) {
	now := time.Now()
	historial := []domain.PuntoHistorial{
		{Fecha: now.Add(-72 * time.Hour), Intensidad: 0.3, RespuestaMAP: 0.05},
		{Fecha: now.Add(-48 * time.Hour), Intensidad: 0.4, RespuestaMAP: 0.08},
		{Fecha: now.Add(-24 * time.Hour), Intensidad: 0.5, RespuestaMAP: 0.01},
	}

	u := domain.Umbral{
		EmpleadoID:     1,
		UmbralAbsoluto: 0.30,
	}
	result := CalibrarUmbral(u, historial)

	if result.UltimoEstimulo != 0.5 {
		t.Errorf("expected ultimo_estimulo=0.5, got %.2f", result.UltimoEstimulo)
	}
	if result.UmbralDiferencial <= 0 {
		t.Errorf("expected positive weber differential, got %.2f", result.UmbralDiferencial)
	}
	if result.FechaUltimoEstimulo == nil {
		t.Error("expected FechaUltimoEstimulo to be set")
	}
}

func TestDetectarRiesgo_AllAboveCurve(t *testing.T) {
	empleados := []domain.Empleado{
		{ID: 1, Nombre: "Ana", Activo: true},
		{ID: 2, Nombre: "Bob", Activo: true},
	}
	perfiles := []domain.PerfilMAP{
		{EmpleadoID: 1, Motivacion: 0.8, Habilidad: 0.8},
		{EmpleadoID: 2, Motivacion: 0.7, Habilidad: 0.9},
	}

	riesgos := DetectarRiesgo(empleados, perfiles, 0.25)

	if len(riesgos) != 0 {
		t.Errorf("expected 0 risks, got %d", len(riesgos))
	}
}

func TestDetectarRiesgo_BelowCurve(t *testing.T) {
	empleados := []domain.Empleado{
		{ID: 1, Nombre: "Ana", Activo: true},
		{ID: 2, Nombre: "Bob", Activo: true},
	}
	perfiles := []domain.PerfilMAP{
		{EmpleadoID: 1, Motivacion: 0.8, Habilidad: 0.8},
		{EmpleadoID: 2, Motivacion: 0.2, Habilidad: 0.3},
	}

	riesgos := DetectarRiesgo(empleados, perfiles, 0.25)

	if len(riesgos) != 1 {
		t.Fatalf("expected 1 risk, got %d", len(riesgos))
	}

	r := riesgos[0]
	if r.EmpleadoID != 2 {
		t.Errorf("expected empleado 2, got %d", r.EmpleadoID)
	}
	if r.Nombre != "Bob" {
		t.Errorf("expected Bob, got %s", r.Nombre)
	}
}

func TestDetectarRiesgo_InactiveSkip(t *testing.T) {
	empleados := []domain.Empleado{
		{ID: 1, Nombre: "Inactivo", Activo: false},
	}
	perfiles := []domain.PerfilMAP{
		{EmpleadoID: 1, Motivacion: 0.1, Habilidad: 0.1},
	}

	riesgos := DetectarRiesgo(empleados, perfiles, 0.25)

	if len(riesgos) != 0 {
		t.Errorf("expected 0 risks for inactive employee, got %d", len(riesgos))
	}
}

func TestRecomendar_AboveCurve_GeneratesStimulus(t *testing.T) {
	e := domain.Empleado{ID: 1, Nombre: "Ana", Departamento: "IT"}
	p := domain.PerfilMAP{
		EmpleadoID:   1,
		Motivacion:   0.8,
		Habilidad:    0.9,
		Sensibilidad: domain.SensibilidadDesarrollo,
	}
	u := domain.Umbral{
		EmpleadoID:     1,
		UmbralAbsoluto: 0.30,
	}

	result := Recomendar(e, p, nil, nil, u)

	if result.EstimuloRecomendado == nil {
		t.Fatal("expected stimulus recommendation when above curve")
	}
	if result.EstimuloRecomendado.Intensidad < 0.3 {
		t.Errorf("expected intensity >= 0.3, got %.2f", result.EstimuloRecomendado.Intensidad)
	}
	if result.EstimuloRecomendado.EmpleadoID != 1 {
		t.Errorf("expected empleado_id=1, got %d", result.EstimuloRecomendado.EmpleadoID)
	}
}

func TestRecomendar_BelowCurve_RecommendsInterventions(t *testing.T) {
	e := domain.Empleado{ID: 2, Nombre: "Bob", Departamento: "IT"}
	p := domain.PerfilMAP{
		EmpleadoID:   2,
		Motivacion:   0.30,
		Habilidad:    0.40,
		Sensibilidad: domain.SensibilidadReconocimiento,
	}
	u := domain.Umbral{
		EmpleadoID:     2,
		UmbralAbsoluto: 0.30,
	}

	incentivos := []domain.Incentivo{
		{ID: 1, Nombre: "Reconocimiento", Tipo: domain.IncentivoIdentidad, Intensidad: 0.8, Activo: true, Disponibilidad: domain.DisponibilidadPermanente},
	}

	result := Recomendar(e, p, incentivos, nil, u)

	// Should NOT generate stimulus when below curve
	if result.EstimuloRecomendado != nil {
		t.Error("expected no stimulus when below curve")
	}
	// Should recommend incentives
	if len(result.IncentivosRecomendados) == 0 {
		t.Error("expected incentive recommendations for low motivation")
	}
}

func TestRecomendar_MatchesSensibilidadType(t *testing.T) {
	e := domain.Empleado{ID: 3, Nombre: "Carlos", Departamento: "Ventas"}
	p := domain.PerfilMAP{
		EmpleadoID:   3,
		Motivacion:   0.30,
		Habilidad:    0.50,
		Sensibilidad: domain.SensibilidadEconomico,
	}
	u := domain.Umbral{
		EmpleadoID:     3,
		UmbralAbsoluto: 0.30,
	}

	incentivos := []domain.Incentivo{
		{ID: 1, Nombre: "Bono", Tipo: domain.IncentivoBeneficios, Intensidad: 0.9, Activo: true, Disponibilidad: domain.DisponibilidadPermanente},
		{ID: 2, Nombre: "Capacitación", Tipo: domain.IncentivoFormacion, Intensidad: 0.7, Activo: true, Disponibilidad: domain.DisponibilidadPermanente},
	}

	result := Recomendar(e, p, incentivos, nil, u)

	// Should only recommend beneficios (matches SensibilidadEconomico)
	if len(result.IncentivosRecomendados) == 0 {
		t.Fatal("expected incentive recommendations")
	}
	for _, inc := range result.IncentivosRecomendados {
		if inc.Tipo != domain.IncentivoBeneficios {
			t.Errorf("expected only beneficios for economico sensitivity, got %s", inc.Tipo)
		}
	}
}

func TestUmbralInicial(t *testing.T) {
	u := UmbralInicial(42)

	if u.EmpleadoID != 42 {
		t.Errorf("expected empleado_id=42, got %d", u.EmpleadoID)
	}
	if u.UmbralAbsoluto != defaultAbsoluteThreshold {
		t.Errorf("expected absoluto=%.2f, got %.2f", defaultAbsoluteThreshold, u.UmbralAbsoluto)
	}
	if u.UmbralDiferencial != defaultWeberConstant {
		t.Errorf("expected diferencial=%.2f, got %.2f", defaultWeberConstant, u.UmbralDiferencial)
	}
}

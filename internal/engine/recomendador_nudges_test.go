package engine

import (
	"testing"

	"estimulos-incentivos/internal/domain"
)

func nudgeFixtureDepto(t *testing.T) (domain.Empleado, domain.PerfilMAP, domain.Umbral) {
	t.Helper()
	e := domain.Empleado{ID: 2, Nombre: "Bob", Departamento: "IT"}
	p := domain.PerfilMAP{
		EmpleadoID:   2,
		Motivacion:   0.30,
		Habilidad:    0.40,
		Sensibilidad: domain.SensibilidadReconocimiento,
	}
	u := domain.Umbral{EmpleadoID: 2, UmbralAbsoluto: 0.30}
	return e, p, u
}

func TestRecomendar_NudgeDepartamento_CoincideDepto(t *testing.T) {
	e, p, u := nudgeFixtureDepto(t)
	nudges := []domain.Nudge{
		{ID: 1, Nombre: "Defaults IT", Tipo: domain.NudgeDefaults, Ambito: domain.AmbitoDepartamento, TargetDepto: "IT", Activo: true},
	}

	result := Recomendar(e, p, nil, nudges, u)

	if len(result.NudgesRecomendados) != 1 {
		t.Fatalf("esperado 1 nudge de departamento recomendado, obtenidos %d", len(result.NudgesRecomendados))
	}
	if result.NudgesRecomendados[0].ID != 1 {
		t.Errorf("nudge recomendado ID = %d, want 1", result.NudgesRecomendados[0].ID)
	}
}

func TestRecomendar_NudgeDepartamento_OtroDeptoExcluido(t *testing.T) {
	e, p, u := nudgeFixtureDepto(t)
	nudges := []domain.Nudge{
		{ID: 1, Nombre: "Defaults RRHH", Tipo: domain.NudgeDefaults, Ambito: domain.AmbitoDepartamento, TargetDepto: "RRHH", Activo: true},
	}

	result := Recomendar(e, p, nil, nudges, u)

	if len(result.NudgesRecomendados) != 0 {
		t.Errorf("nudge de otro departamento no debe recomendarse, obtenidos %d", len(result.NudgesRecomendados))
	}
}

func TestRecomendar_NudgesAmbitoPreservado(t *testing.T) {
	e, p, u := nudgeFixtureDepto(t)
	nudges := []domain.Nudge{
		{ID: 1, Nombre: "Global", Tipo: domain.NudgeFraming, Ambito: domain.AmbitoGlobal, Activo: true},
		{ID: 2, Nombre: "Individual propio", Tipo: domain.NudgeFriccion, Ambito: domain.AmbitoIndividual, TargetID: 2, Activo: true},
		{ID: 3, Nombre: "Individual ajeno", Tipo: domain.NudgeFriccion, Ambito: domain.AmbitoIndividual, TargetID: 99, Activo: true},
		{ID: 4, Nombre: "Depto propio", Tipo: domain.NudgeDefaults, Ambito: domain.AmbitoDepartamento, TargetDepto: "IT", Activo: true},
		{ID: 5, Nombre: "Depto ajeno", Tipo: domain.NudgeDefaults, Ambito: domain.AmbitoDepartamento, TargetDepto: "RRHH", Activo: true},
		{ID: 6, Nombre: "Inactivo", Tipo: domain.NudgeDefaults, Ambito: domain.AmbitoGlobal, Activo: false},
	}

	result := Recomendar(e, p, nil, nudges, u)

	gotIDs := map[int64]bool{}
	for _, n := range result.NudgesRecomendados {
		gotIDs[n.ID] = true
	}
	want := map[int64]bool{1: true, 2: true, 4: true}
	if len(gotIDs) != len(want) {
		t.Fatalf("nudges recomendados = %v, want %v", gotIDs, want)
	}
	for id := range want {
		if !gotIDs[id] {
			t.Errorf("nudge %d debe recomendarse y no fue incluido", id)
		}
	}
}

package engine

import (
	"math"
	"testing"

	"estimulos-incentivos/internal/domain"
)

// fixtureAnalisis arma empleados en tres departamentos y perfiles sobre la
// curva de acción, de modo que el análisis no tenga empleados en riesgo.
func fixtureAnalisis() ([]domain.Empleado, []domain.PerfilMAP) {
	empleados := []domain.Empleado{
		{ID: 2, Nombre: "Ana", Departamento: "comercial", Activo: true},
		{ID: 3, Nombre: "Beto", Departamento: "comercial", Activo: true},
		{ID: 4, Nombre: "Caro", Departamento: "operaciones", Activo: true},
		{ID: 5, Nombre: "Dani", Departamento: "rrhh", Activo: true},
		{ID: 6, Nombre: "Eli", Departamento: "rrhh", Activo: false}, // inactivo: fuera de distribuciones
	}
	perfiles := []domain.PerfilMAP{
		{EmpleadoID: 2, Motivacion: 0.5, Habilidad: 0.75, Prompt: 0.5},
		{EmpleadoID: 3, Motivacion: 0.75, Habilidad: 0.75, Prompt: 0.5},
		{EmpleadoID: 4, Motivacion: 0.75, Habilidad: 0.75, Prompt: 0.25},
		{EmpleadoID: 5, Motivacion: 0.75, Habilidad: 0.75, Prompt: 0.75},
	}
	return empleados, perfiles
}

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}

func TestAnalizar_DistribucionesOrdenadasPorDepto(t *testing.T) {
	empleados, perfiles := fixtureAnalisis()
	result := Analizar(empleados, perfiles, nil)

	if len(result.Distribuciones) != 3 {
		t.Fatalf("Distribuciones: esperadas 3, obtenidas %d", len(result.Distribuciones))
	}

	wantOrden := []string{"comercial", "operaciones", "rrhh"}
	for i, depto := range wantOrden {
		got := result.Distribuciones[i]
		if got.Depto != depto {
			t.Fatalf("Distribuciones[%d].Depto = %q, esperado %q (orden alfabético)", i, got.Depto, depto)
		}
	}

	// comercial: M=(0.5+0.75)/2=0.625, A=(0.75+0.75)/2=0.75, P=0.5, MAP=0.46875, count=2
	comercial := result.Distribuciones[0]
	if !almostEqual(comercial.PromedioM, 0.625) || !almostEqual(comercial.PromedioA, 0.75) || !almostEqual(comercial.PromedioP, 0.5) || !almostEqual(comercial.MAP, 0.46875) || comercial.TotalEmpleados != 2 {
		t.Errorf("comercial = %+v, esperado promedioM=0.625 promedioA=0.75 promedioP=0.5 map=0.46875 count=2", comercial)
	}

	// operaciones: count=1, M=A=0.75, MAP=0.5625
	operaciones := result.Distribuciones[1]
	if !almostEqual(operaciones.MAP, 0.5625) || operaciones.TotalEmpleados != 1 {
		t.Errorf("operaciones = %+v, esperado map=0.5625 count=1", operaciones)
	}

	// rrhh: count=1, M=A=0.75, MAP=0.5625
	rrhh := result.Distribuciones[2]
	if !almostEqual(rrhh.MAP, 0.5625) || rrhh.TotalEmpleados != 1 {
		t.Errorf("rrhh = %+v, esperado map=0.5625 count=1", rrhh)
	}
}

func TestAnalizar_EfectividadOrdenadaPorTipo(t *testing.T) {
	historial := map[int64][]domain.PuntoHistorial{
		2: {
			{Tipo: "framing", RespuestaMAP: 0.4},
			{Tipo: "framing", RespuestaMAP: 0},
			{Tipo: "defaults", RespuestaMAP: 0.2},
		},
		3: {
			{Tipo: "social_proof", RespuestaMAP: 0.3},
			{Tipo: "social_proof", RespuestaMAP: 0},
		},
		4: {
			{Tipo: "defaults", RespuestaMAP: 0},
		},
	}

	result := Analizar(nil, nil, historial)

	wantOrden := []string{"defaults", "framing", "social_proof"}
	if len(result.Efectividad) != len(wantOrden) {
		t.Fatalf("Efectividad: esperadas %d filas, obtenidas %d", len(wantOrden), len(result.Efectividad))
	}
	for i, tipo := range wantOrden {
		if result.Efectividad[i].Tipo != tipo {
			t.Fatalf("Efectividad[%d].Tipo = %q, esperado %q (orden alfabético)", i, result.Efectividad[i].Tipo, tipo)
		}
	}

	// defaults: total=2, exitos=1 → 0.5
	defaults := result.Efectividad[0]
	if defaults.TotalAplicados != 2 || defaults.TotalExitosos != 1 || !almostEqual(defaults.TasaExito, 0.5) {
		t.Errorf("defaults = %+v, esperado total=2 exitos=1 tasa=0.5", defaults)
	}
	// framing: total=2, exitos=1 → 0.5
	framing := result.Efectividad[1]
	if framing.TotalAplicados != 2 || framing.TotalExitosos != 1 || !almostEqual(framing.TasaExito, 0.5) {
		t.Errorf("framing = %+v, esperado total=2 exitos=1 tasa=0.5", framing)
	}
	// social_proof: total=2, exitos=1 → 0.5
	socialProof := result.Efectividad[2]
	if socialProof.TotalAplicados != 2 || socialProof.TotalExitosos != 1 || !almostEqual(socialProof.TasaExito, 0.5) {
		t.Errorf("social_proof = %+v, esperado total=2 exitos=1 tasa=0.5", socialProof)
	}
}

func TestAnalizar_TotalesYRiesgo(t *testing.T) {
	empleados, perfiles := fixtureAnalisis()
	result := Analizar(empleados, perfiles, nil)

	if result.TotalEmpleados != 5 {
		t.Errorf("TotalEmpleados = %d, esperado 5 (incluye inactivos)", result.TotalEmpleados)
	}
	if result.EmpleadosEnRiesgo != 0 {
		t.Errorf("EmpleadosEnRiesgo = %d, esperado 0 (todos sobre la curva)", result.EmpleadosEnRiesgo)
	}
	if len(result.ZonaRiesgo) != 0 {
		t.Errorf("ZonaRiesgo = %+v, esperada vacía", result.ZonaRiesgo)
	}
}

func TestAnalizar_RiesgoPassthrough(t *testing.T) {
	empleados, perfiles := fixtureAnalisis()
	// Bajar M×A de Dani (rrhh) por debajo del umbral 0.25: 0.5*0.3=0.15.
	perfiles[3].Motivacion = 0.5
	perfiles[3].Habilidad = 0.3

	result := Analizar(empleados, perfiles, nil)
	if result.EmpleadosEnRiesgo != 1 || len(result.ZonaRiesgo) != 1 {
		t.Fatalf("esperado 1 empleado en riesgo, obtenidos %d (zona %+v)", result.EmpleadosEnRiesgo, result.ZonaRiesgo)
	}
	if result.ZonaRiesgo[0].EmpleadoID != 5 {
		t.Errorf("ZonaRiesgo[0].EmpleadoID = %d, esperado 5", result.ZonaRiesgo[0].EmpleadoID)
	}
}

func TestAnalizar_DeterministaEntreEjecuciones(t *testing.T) {
	empleados, perfiles := fixtureAnalisis()
	historial := map[int64][]domain.PuntoHistorial{
		2: {{Tipo: "framing", RespuestaMAP: 0.4}, {Tipo: "defaults", RespuestaMAP: 0.2}},
		3: {{Tipo: "social_proof", RespuestaMAP: 0.3}, {Tipo: "framing", RespuestaMAP: 0}},
		4: {{Tipo: "defaults", RespuestaMAP: 0}, {Tipo: "social_proof", RespuestaMAP: 0.1}},
	}

	first := Analizar(empleados, perfiles, historial)
	for i := 0; i < 20; i++ {
		got := Analizar(empleados, perfiles, historial)
		if !sameAnalisis(first, got) {
			t.Fatalf("ejecución %d divergió del resultado inicial", i)
		}
	}
}

func sameAnalisis(a, b AnalisisResult) bool {
	if a.TotalEmpleados != b.TotalEmpleados || a.EmpleadosEnRiesgo != b.EmpleadosEnRiesgo {
		return false
	}
	if len(a.Distribuciones) != len(b.Distribuciones) || len(a.Efectividad) != len(b.Efectividad) || len(a.ZonaRiesgo) != len(b.ZonaRiesgo) {
		return false
	}
	for i := range a.Distribuciones {
		if a.Distribuciones[i] != b.Distribuciones[i] {
			return false
		}
	}
	for i := range a.Efectividad {
		if a.Efectividad[i] != b.Efectividad[i] {
			return false
		}
	}
	for i := range a.ZonaRiesgo {
		if a.ZonaRiesgo[i] != b.ZonaRiesgo[i] {
			return false
		}
	}
	return true
}

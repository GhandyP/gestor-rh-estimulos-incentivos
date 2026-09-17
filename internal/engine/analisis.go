package engine

import (
	"sort"

	"estimulos-incentivos/internal/domain"
)

type DistribucionMAP struct {
	Depto          string  `json:"depto"`
	PromedioM      float64 `json:"promedio_m"`
	PromedioA      float64 `json:"promedio_a"`
	PromedioP      float64 `json:"promedio_p"`
	MAP            float64 `json:"map"`
	TotalEmpleados int     `json:"total_empleados"`
}

type EfectividadRow struct {
	Tipo           string  `json:"tipo"`
	TotalAplicados int     `json:"total_aplicados"`
	TotalExitosos  int     `json:"total_exitosos"`
	TasaExito      float64 `json:"tasa_exito"`
}

type AnalisisResult struct {
	Distribuciones    []DistribucionMAP `json:"distribuciones"`
	Efectividad       []EfectividadRow  `json:"efectividad"`
	ZonaRiesgo        []RiesgoResult    `json:"zona_riesgo"`
	TotalEmpleados    int               `json:"total_empleados"`
	EmpleadosEnRiesgo int               `json:"empleados_en_riesgo"`
}

// Analizar genera el análisis descriptivo del capital humano.
func Analizar(
	empleados []domain.Empleado,
	perfiles []domain.PerfilMAP,
	historialEstimulos map[int64][]domain.PuntoHistorial,
) AnalisisResult {
	result := AnalisisResult{
		Distribuciones: make([]DistribucionMAP, 0),
		Efectividad:    make([]EfectividadRow, 0),
		ZonaRiesgo:     make([]RiesgoResult, 0),
	}

	perfilMap := make(map[int64]domain.PerfilMAP)
	for _, p := range perfiles {
		perfilMap[p.EmpleadoID] = p
	}

	// Distribución por departamento
	deptos := make(map[string]struct {
		sumM, sumA, sumP float64
		count            int
	})

	for _, e := range empleados {
		if !e.Activo {
			continue
		}
		p, ok := perfilMap[e.ID]
		if !ok {
			continue
		}
		d := deptos[string(e.Departamento)]
		d.sumM += p.Motivacion
		d.sumA += p.Habilidad
		d.sumP += p.Prompt
		d.count++
		deptos[string(e.Departamento)] = d
	}

	for depto, d := range deptos {
		if d.count == 0 {
			continue
		}
		avgM := d.sumM / float64(d.count)
		avgA := d.sumA / float64(d.count)
		result.Distribuciones = append(result.Distribuciones, DistribucionMAP{
			Depto:          depto,
			PromedioM:      avgM,
			PromedioA:      avgA,
			PromedioP:      d.sumP / float64(d.count),
			MAP:            avgM * avgA,
			TotalEmpleados: d.count,
		})
	}
	// Orden determinista: la iteración de mapas es aleatoria en Go.
	sort.Slice(result.Distribuciones, func(i, j int) bool {
		return result.Distribuciones[i].Depto < result.Distribuciones[j].Depto
	})

	// Zona de riesgo
	result.ZonaRiesgo = DetectarRiesgo(empleados, perfiles, 0.25)

	// Efectividad por tipo de estímulo
	tipoStats := make(map[string]struct{ total, exitos int })
	for _, puntos := range historialEstimulos {
		for _, p := range puntos {
			t := p.Tipo
			if t == "" {
				t = "estímulo"
			}
			s := tipoStats[t]
			s.total++
			if p.RespuestaMAP > 0 {
				s.exitos++
			}
			tipoStats[t] = s
		}
	}
	for tipo, s := range tipoStats {
		if s.total == 0 {
			continue
		}
		result.Efectividad = append(result.Efectividad, EfectividadRow{
			Tipo:           tipo,
			TotalAplicados: s.total,
			TotalExitosos:  s.exitos,
			TasaExito:      float64(s.exitos) / float64(s.total),
		})
	}
	// Orden determinista: la iteración de mapas es aleatoria en Go.
	sort.Slice(result.Efectividad, func(i, j int) bool {
		return result.Efectividad[i].Tipo < result.Efectividad[j].Tipo
	})

	// Totales
	result.TotalEmpleados = len(empleados)
	result.EmpleadosEnRiesgo = len(result.ZonaRiesgo)

	return result
}

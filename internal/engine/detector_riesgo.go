package engine

import (
	"sort"

	"estimulos-incentivos/internal/domain"
)

type RiesgoResult struct {
	EmpleadoID  int64   `json:"empleado_id"`
	Nombre      string  `json:"nombre"`
	MAPProducto float64 `json:"map_producto"`
	Severidad   string  `json:"severidad"`
	Motivo      string  `json:"motivo"`
}

// DetectarRiesgo identifica empleados bajo la curva de acción de Fogg.
func DetectarRiesgo(empleados []domain.Empleado, perfiles []domain.PerfilMAP, defaultThreshold float64) []RiesgoResult {
	perfilMap := make(map[int64]domain.PerfilMAP)
	for _, p := range perfiles {
		perfilMap[p.EmpleadoID] = p
	}

	riesgos := make([]RiesgoResult, 0)
	for _, e := range empleados {
		if !e.Activo {
			continue
		}
		p, ok := perfilMap[e.ID]
		if !ok {
			continue
		}

		threshold := defaultThreshold
		if threshold == 0 {
			threshold = 0.25
		}

		if !p.CurvaAccion(threshold) {
			producto := p.MAPProducto()
			severidad := "baja"
			motivo := describeRiesgo(p, threshold)

			if producto < threshold*0.5 {
				severidad = "alta"
			} else if producto < threshold*0.8 {
				severidad = "media"
			}

			riesgos = append(riesgos, RiesgoResult{
				EmpleadoID:  e.ID,
				Nombre:      e.Nombre,
				MAPProducto: producto,
				Severidad:   severidad,
				Motivo:      motivo,
			})
		}
	}

	sort.Slice(riesgos, func(i, j int) bool {
		return riesgos[i].MAPProducto < riesgos[j].MAPProducto
	})

	return riesgos
}

func describeRiesgo(p domain.PerfilMAP, threshold float64) string {
	if p.Motivacion < threshold && p.Habilidad < threshold {
		return "Baja motivación y baja habilidad — requiere intervención integral"
	}
	if p.Motivacion < threshold {
		return "Baja motivación — revisar incentivos y alineación con intereses"
	}
	if p.Habilidad < threshold {
		return "Baja habilidad percibida — revisar nudges y eliminar fricciones"
	}
	return "Combinación de factores — revisión individual recomendada"
}

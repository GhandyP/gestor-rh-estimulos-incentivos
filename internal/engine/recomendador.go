package engine

import (
	"fmt"
	"time"

	"estimulos-incentivos/internal/domain"
)

type RecomendacionResult struct {
	Diagnostico            string             `json:"diagnostico"`
	IncentivosRecomendados []domain.Incentivo `json:"incentivos_recomendados"`
	NudgesRecomendados     []domain.Nudge     `json:"nudges_recomendados"`
	EstimuloRecomendado    *domain.Estimulo   `json:"estimulo_recomendado,omitempty"`
}

// Recomendar orquesta las tres capas conductuales para un empleado específico.
func Recomendar(
	empleado domain.Empleado,
	perfil domain.PerfilMAP,
	incentivos []domain.Incentivo,
	nudges []domain.Nudge,
	umbral domain.Umbral,
) RecomendacionResult {
	result := RecomendacionResult{
		IncentivosRecomendados: make([]domain.Incentivo, 0),
		NudgesRecomendados:     make([]domain.Nudge, 0),
	}
	threshold := 0.25

	sobreCurva := perfil.CurvaAccion(threshold)

	if !sobreCurva {
		result.Diagnostico = fmt.Sprintf(
			"Empleado bajo la curva de acción (M=%.2f, A=%.2f, umbral=%.2f). Se necesita intervención.",
			perfil.Motivacion, perfil.Habilidad, threshold,
		)

		// Capa 1: Recomendar incentivos que eleven M
		if perfil.Motivacion < 0.5 {
			result.IncentivosRecomendados = recomendarIncentivos(incentivos, perfil)
		}

		// Capa 2: Recomendar nudges que reduzcan fricción (mejoren A)
		if perfil.Habilidad < 0.5 {
			result.NudgesRecomendados = recomendarNudges(nudges, empleado)
		}
	} else {
		result.Diagnostico = fmt.Sprintf(
			"Empleado sobre la curva de acción (M=%.2f, A=%.2f). Listo para estímulo calibrado.",
			perfil.Motivacion, perfil.Habilidad,
		)

		// Capa 3: Estímulo calibrado
		intensidad := umbral.IntensidadMinimaPerceptible()
		if intensidad < 0.4 {
			intensidad = 0.4
		}

		canal := domain.CanalEmail
		if perfil.Sensibilidad == domain.SensibilidadReconocimiento {
			canal = domain.CanalPresencial
		}

		result.EstimuloRecomendado = &domain.Estimulo{
			EmpleadoID:          empleado.ID,
			Tipo:                "recomendacion_personalizada",
			Contenido:           generarContenidoEstimulo(empleado, perfil),
			Intensidad:          intensidad,
			Canal:               canal,
			Estado:              domain.EstadoPendiente,
			FechaIdeal:          time.Now().Add(24 * time.Hour),
			OrigenRecomendacion: fmt.Sprintf("Motor MAP: motivacion=%.2f, habilidad=%.2f", perfil.Motivacion, perfil.Habilidad),
		}
	}

	return result
}

func recomendarIncentivos(incentivos []domain.Incentivo, perfil domain.PerfilMAP) []domain.Incentivo {
	var recomendados []domain.Incentivo
	for _, inc := range incentivos {
		if !inc.Activo || !inc.HayCupos() {
			continue
		}
		if coincideTipo(inc.Tipo, perfil.Sensibilidad) {
			recomendados = append(recomendados, inc)
		}
	}
	if len(recomendados) > 3 {
		recomendados = recomendados[:3]
	}
	return recomendados
}

func recomendarNudges(nudges []domain.Nudge, empleado domain.Empleado) []domain.Nudge {
	var recomendados []domain.Nudge
	for _, n := range nudges {
		if !n.Activo {
			continue
		}
		// Solo nudges globales o que apliquen al depto del empleado
		if n.Ambito == domain.AmbitoGlobal ||
			(n.Ambito == domain.AmbitoIndividual && n.TargetID == empleado.ID) {
			recomendados = append(recomendados, n)
		}
	}
	if len(recomendados) > 3 {
		recomendados = recomendados[:3]
	}
	return recomendados
}

func coincideTipo(tipoInc domain.TipoIncentivo, sens domain.Sensibilidad) bool {
	mapping := map[domain.TipoIncentivo]domain.Sensibilidad{
		domain.IncentivoIdentidad:           domain.SensibilidadReconocimiento,
		domain.IncentivoBeneficios:          domain.SensibilidadEconomico,
		domain.IncentivoFormacion:           domain.SensibilidadDesarrollo,
		domain.IncentivoProyectoCorporativo: domain.SensibilidadDesarrollo,
	}
	if s, ok := mapping[tipoInc]; ok {
		return s == sens
	}
	return true
}

func generarContenidoEstimulo(e domain.Empleado, p domain.PerfilMAP) string {
	switch p.Sensibilidad {
	case domain.SensibilidadReconocimiento:
		return fmt.Sprintf("Reconocimiento a %s por su contribución en %s", e.Nombre, e.Departamento)
	case domain.SensibilidadDesarrollo:
		return fmt.Sprintf("Oportunidad de desarrollo profesional para %s", e.Nombre)
	case domain.SensibilidadEconomico:
		return fmt.Sprintf("Incentivo por desempeño destacado - %s", e.Nombre)
	default:
		return fmt.Sprintf("Recordatorio de bienestar para %s", e.Nombre)
	}
}

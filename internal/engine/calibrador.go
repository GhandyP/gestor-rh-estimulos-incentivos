package engine

import (
	"math"
	"sort"
	"time"

	"estimulos-incentivos/internal/domain"
)

const (
	defaultAbsoluteThreshold = 0.30
	defaultWeberConstant    = 0.15
	minConfidence           = 0.30
	historyWindowDays       = 90
)

func CalibrarUmbral(u domain.Umbral, historial []domain.PuntoHistorial) domain.Umbral {
	if len(historial) == 0 {
		if u.UmbralAbsoluto == 0 {
			u.UmbralAbsoluto = defaultAbsoluteThreshold
		}
		if u.UmbralDiferencial == 0 {
			u.UmbralDiferencial = defaultWeberConstant
		}
		return u
	}

	// Calculate Weber constant from history: intensity perception ratio
	var ratios []float64
	for i := 1; i < len(historial); i++ {
		prev := historial[i-1].Intensidad
		curr := historial[i].Intensidad
		if prev > 0.01 && curr > prev {
			ratio := (curr - prev) / prev
			ratios = append(ratios, ratio)
		}
	}

	if len(ratios) > 0 {
		sort.Float64s(ratios)
		// Use median for robustness
		u.UmbralDiferencial = ratios[len(ratios)/2]
		if u.UmbralDiferencial < 0.05 {
			u.UmbralDiferencial = 0.05
		}
		if u.UmbralDiferencial > 0.50 {
			u.UmbralDiferencial = 0.50
		}
	}

	// Adjust absolute threshold based on response patterns
	var positiveResponses int
	for _, h := range historial {
		if h.RespuestaMAP > 0 {
			positiveResponses++
		}
	}
	successRate := float64(positiveResponses) / float64(len(historial))

	if successRate > 0.8 {
		// Most stimuli effective → current threshold is appropriate, small fine-tuning
		u.UmbralAbsoluto = math.Max(0.10, u.UmbralAbsoluto-0.02)
	} else if successRate < 0.3 {
		// Few stimuli effective → threshold may be too low
		u.UmbralAbsoluto = math.Min(0.70, u.UmbralAbsoluto+0.05)
	}

	u.UltimoEstimulo = historial[len(historial)-1].Intensidad
	ultimo := historial[len(historial)-1].Fecha
	u.FechaUltimoEstimulo = &ultimo
	u.UpdatedAt = time.Now().UTC()

	return u
}

func UmbralInicial(empleadoID int64) domain.Umbral {
	return domain.Umbral{
		EmpleadoID:       empleadoID,
		UmbralAbsoluto:   defaultAbsoluteThreshold,
		UmbralDiferencial: defaultWeberConstant,
		UpdatedAt:        time.Now().UTC(),
	}
}

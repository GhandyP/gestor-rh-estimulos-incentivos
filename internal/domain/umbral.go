package domain

import "time"

// Umbral modela los umbrales psicofísicos de un empleado frente a estímulos.
type Umbral struct {
	ID               int64     `json:"id"`
	EmpleadoID       int64     `json:"empleado_id"`
	UmbralAbsoluto   float64   `json:"umbral_absoluto"`   // Intensidad mínima para ser detectado (0.0 - 1.0)
	UmbralDiferencial float64  `json:"umbral_diferencial"` // Constante de Weber personal (ej: 0.15)
	UltimoEstimulo   float64   `json:"ultimo_estimulo"`   // Intensidad del último estímulo recibido
	FechaUltimoEstimulo *time.Time `json:"fecha_ultimo_estimulo,omitempty"`
	Historial        []PuntoHistorial `json:"historial,omitempty"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// PuntoHistorial registra un estímulo aplicado y la respuesta observada.
type PuntoHistorial struct {
	Fecha        time.Time `json:"fecha"`
	Intensidad   float64   `json:"intensidad"`
	RespuestaMAP float64   `json:"respuesta_map"`
	Tipo         string    `json:"tipo,omitempty"`
}

// IntensidadMinimaPerceptible calcula cuánta intensidad necesita el próximo estímulo
// para ser percibido como mejora sobre el anterior (Ley de Weber).
// Si no hay estímulo previo, retorna el umbral absoluto.
func (u Umbral) IntensidadMinimaPerceptible() float64 {
	if u.UltimoEstimulo == 0 {
		return u.UmbralAbsoluto
	}
	intensidad := u.UltimoEstimulo * (1 + u.UmbralDiferencial)
	if intensidad < u.UmbralAbsoluto {
		return u.UmbralAbsoluto
	}
	if intensidad > 1.0 {
		return 1.0
	}
	return intensidad
}

// Fatiga detecta si el empleado está recibiendo estímulos con demasiada frecuencia.
// Retorna true si hay riesgo de fatiga (umbral subiendo peligrosamente).
func (u Umbral) Fatiga(frecuenciaMaxima int, ventanaDias int) bool {
	if len(u.Historial) == 0 {
		return false
	}
	limite := time.Now().AddDate(0, 0, -ventanaDias)
	count := 0
	for _, p := range u.Historial {
		if p.Fecha.After(limite) {
			count++
		}
	}
	return count >= frecuenciaMaxima
}

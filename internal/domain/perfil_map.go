package domain

import "time"

// Sensibilidad indica a qué tipo de motivador es más receptivo el empleado.
type Sensibilidad string

const (
	SensibilidadEconomico     Sensibilidad = "economico"
	SensibilidadReconocimiento Sensibilidad = "reconocimiento"
	SensibilidadDesarrollo    Sensibilidad = "desarrollo"
	SensibilidadBienestar     Sensibilidad = "bienestar"
)

// PerfilMAP modela las tres dimensiones del modelo de Fogg (B = M × A × P).
// Todas las puntuaciones van de 0.0 a 1.0.
type PerfilMAP struct {
	ID           int64     `json:"id"`
	EmpleadoID   int64     `json:"empleado_id"`
	Motivacion   float64   `json:"motivacion"`    // 0.0 - 1.0
	Habilidad    float64   `json:"habilidad"`     // 0.0 - 1.0
	Prompt       float64   `json:"prompt"`        // Sensibilidad general al detonante, 0.0 - 1.0
	Sensibilidad Sensibilidad `json:"sensibilidad"` // Tipo de motivador preferido
	Confiabilidad float64  `json:"confiabilidad"` // 0.0 - 1.0: qué tan confiable es este perfil (basado en cantidad de datos)
	UpdatedAt    time.Time `json:"updated_at"`
}

// CurvaAccion calcula si el empleado está sobre la curva de acción de Fogg.
// Retorna true si M × A ≥ threshold (el prompt puede funcionar).
func (p PerfilMAP) CurvaAccion(threshold float64) bool {
	return (p.Motivacion * p.Habilidad) >= threshold
}

// MAPProducto es el producto M × A, usado para comparaciones rápidas.
func (p PerfilMAP) MAPProducto() float64 {
	return p.Motivacion * p.Habilidad
}

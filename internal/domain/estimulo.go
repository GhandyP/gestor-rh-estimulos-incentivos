package domain

import "time"

// CanalEstímulo define el medio por el cual se entrega el estímulo.
type CanalEstimulo string

const (
	CanalEmail     CanalEstimulo = "email"
	CanalSlack     CanalEstimulo = "slack"
	CanalPresencial CanalEstimulo = "presencial"
	CanalDashboard  CanalEstimulo = "dashboard"
)

// EstadoEstímulo indica si el estímulo fue aplicado o está pendiente.
type EstadoEstimulo string

const (
	EstadoPendiente EstadoEstimulo = "pendiente"
	EstadoAplicado  EstadoEstimulo = "aplicado"
	EstadoFallido   EstadoEstimulo = "fallido"
)

// Estimulo es un trigger táctico calibrado para un empleado específico.
type Estimulo struct {
	ID          int64          `json:"id"`
	EmpleadoID  int64          `json:"empleado_id"`
	Tipo        string         `json:"tipo"`         // Mensaje, invitación, alerta, recordatorio
	Contenido   string         `json:"contenido"`
	Intensidad  float64        `json:"intensidad"`   // Calibrada al umbral del empleado (0.0 - 1.0)
	Canal       CanalEstimulo  `json:"canal"`
	Estado      EstadoEstimulo `json:"estado"`
	FechaIdeal  time.Time      `json:"fecha_ideal"`  // Timing óptimo según el motor
	FechaAplicado *time.Time   `json:"fecha_aplicado,omitempty"`
	OrigenRecomendacion string  `json:"origen_recomendacion"` // Por qué el motor recomendó esto
	CreatedAt   time.Time      `json:"created_at"`
}

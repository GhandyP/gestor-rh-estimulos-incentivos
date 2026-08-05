package domain

import (
	"fmt"
	"strings"
	"time"
)

// CanalEstímulo define el medio por el cual se entrega el estímulo.
type CanalEstimulo string

const (
	CanalEmail      CanalEstimulo = "email"
	CanalSlack      CanalEstimulo = "slack"
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
	ID                  int64          `json:"id"`
	EmpleadoID          int64          `json:"empleado_id"`
	Tipo                string         `json:"tipo"` // Mensaje, invitación, alerta, recordatorio
	Contenido           string         `json:"contenido"`
	Intensidad          float64        `json:"intensidad"` // Calibrada al umbral del empleado (0.0 - 1.0)
	Canal               CanalEstimulo  `json:"canal"`
	Estado              EstadoEstimulo `json:"estado"`
	FechaIdeal          time.Time      `json:"fecha_ideal"` // Timing óptimo según el motor
	FechaAplicado       *time.Time     `json:"fecha_aplicado,omitempty"`
	OrigenRecomendacion string         `json:"origen_recomendacion"` // Por qué el motor recomendó esto
	CreatedAt           time.Time      `json:"created_at"`
}

// Validate comprueba las invariantes de un estímulo antes de persistirlo.
func (e *Estimulo) Validate() error {
	if e.EmpleadoID <= 0 {
		return fmt.Errorf("empleado_id es requerido")
	}
	if strings.TrimSpace(e.Tipo) == "" {
		return fmt.Errorf("tipo es requerido")
	}
	if strings.TrimSpace(e.Contenido) == "" {
		return fmt.Errorf("contenido es requerido")
	}
	if e.Intensidad < 0 || e.Intensidad > 1 {
		return fmt.Errorf("intensidad debe estar entre 0 y 1")
	}
	if !canalValido(e.Canal) {
		return fmt.Errorf("canal inválido: %q", e.Canal)
	}
	if !estadoValido(e.Estado) {
		return fmt.Errorf("estado inválido: %q", e.Estado)
	}
	if (e.Estado == EstadoAplicado) != (e.FechaAplicado != nil) {
		return fmt.Errorf("estado 'aplicado' y fecha_aplicado deben registrarse juntos")
	}
	return nil
}

func canalValido(c CanalEstimulo) bool {
	switch c {
	case CanalEmail, CanalSlack, CanalPresencial, CanalDashboard:
		return true
	default:
		return false
	}
}

func estadoValido(s EstadoEstimulo) bool {
	switch s {
	case EstadoPendiente, EstadoAplicado, EstadoFallido:
		return true
	default:
		return false
	}
}

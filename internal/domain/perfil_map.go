package domain

import (
	"fmt"
	"time"
)

// Sensibilidad indica a qué tipo de motivador es más receptivo el empleado.
type Sensibilidad string

const (
	SensibilidadEconomico      Sensibilidad = "economico"
	SensibilidadReconocimiento Sensibilidad = "reconocimiento"
	SensibilidadDesarrollo     Sensibilidad = "desarrollo"
	SensibilidadBienestar      Sensibilidad = "bienestar"
)

// PerfilMAP modela las tres dimensiones del modelo de Fogg (B = M × A × P).
// Todas las puntuaciones van de 0.0 a 1.0.
type PerfilMAP struct {
	ID            int64        `json:"id"`
	EmpleadoID    int64        `json:"empleado_id"`
	Motivacion    float64      `json:"motivacion"`    // 0.0 - 1.0
	Habilidad     float64      `json:"habilidad"`     // 0.0 - 1.0
	Prompt        float64      `json:"prompt"`        // Sensibilidad general al detonante, 0.0 - 1.0
	Sensibilidad  Sensibilidad `json:"sensibilidad"`  // Tipo de motivador preferido
	Confiabilidad float64      `json:"confiabilidad"` // 0.0 - 1.0: qué tan confiable es este perfil (basado en cantidad de datos)
	UpdatedAt     time.Time    `json:"updated_at"`
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

// Validate comprueba las invariantes de un perfil MAP antes de persistirlo.
func (p *PerfilMAP) Validate() error {
	if p.EmpleadoID <= 0 {
		return fmt.Errorf("empleado_id es requerido")
	}
	if p.Motivacion < 0 || p.Motivacion > 1 {
		return fmt.Errorf("motivacion debe estar entre 0 y 1")
	}
	if p.Habilidad < 0 || p.Habilidad > 1 {
		return fmt.Errorf("habilidad debe estar entre 0 y 1")
	}
	if p.Prompt < 0 || p.Prompt > 1 {
		return fmt.Errorf("prompt debe estar entre 0 y 1")
	}
	if p.Confiabilidad < 0 || p.Confiabilidad > 1 {
		return fmt.Errorf("confiabilidad debe estar entre 0 y 1")
	}
	if !sensibilidadValida(p.Sensibilidad) {
		return fmt.Errorf("sensibilidad inválida: %q", p.Sensibilidad)
	}
	return nil
}

func sensibilidadValida(s Sensibilidad) bool {
	switch s {
	case SensibilidadEconomico, SensibilidadReconocimiento, SensibilidadDesarrollo, SensibilidadBienestar:
		return true
	default:
		return false
	}
}

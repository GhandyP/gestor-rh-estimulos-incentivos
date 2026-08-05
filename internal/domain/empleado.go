package domain

import (
	"fmt"
	"strings"
	"time"
)

// Departamento representa una unidad organizacional.
type Departamento string

// Empleado es la entidad central del sistema.
type Empleado struct {
	ID           int64        `json:"id"`
	Nombre       string       `json:"nombre"`
	Email        string       `json:"email"`
	Cargo        string       `json:"cargo"`
	Departamento Departamento `json:"departamento"`
	FechaIngreso time.Time    `json:"fecha_ingreso"`
	Activo       bool         `json:"activo"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

// Validate comprueba las invariantes de un empleado antes de persistirlo.
// FechaIngreso puede ser cero: los flujos actuales la omiten y se persiste el
// valor por defecto del store.
func (e *Empleado) Validate() error {
	if strings.TrimSpace(e.Nombre) == "" {
		return fmt.Errorf("nombre es requerido")
	}
	if strings.TrimSpace(e.Email) == "" {
		return fmt.Errorf("email es requerido")
	}
	if !strings.Contains(e.Email, "@") {
		return fmt.Errorf("email inválido: %q", e.Email)
	}
	if strings.TrimSpace(e.Cargo) == "" {
		return fmt.Errorf("cargo es requerido")
	}
	if strings.TrimSpace(string(e.Departamento)) == "" {
		return fmt.Errorf("departamento es requerido")
	}
	return nil
}

package domain

import "time"

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

package domain

import "time"

// TipoIncentivo clasifica los incentivos según el eje motivacional que activan.
type TipoIncentivo string

const (
	IncentivoIdentidad          TipoIncentivo = "identidad"
	IncentivoBeneficios         TipoIncentivo = "beneficios"
	IncentivoFormacion          TipoIncentivo = "formacion"
	IncentivoProyectoCorporativo TipoIncentivo = "proyecto_corporativo"
)

// Disponibilidad indica si un incentivo tiene cupos limitados o es permanente.
type Disponibilidad string

const (
	DisponibilidadLimitado   Disponibilidad = "limitado"
	DisponibilidadRecurrente Disponibilidad = "recurrente"
	DisponibilidadPermanente Disponibilidad = "permanente"
)

// Incentivo representa un motivador estratégico del catálogo de RRHH.
type Incentivo struct {
	ID             int64          `json:"id"`
	Nombre         string         `json:"nombre"`
	Descripcion    string         `json:"descripcion"`
	Tipo           TipoIncentivo  `json:"tipo"`
	Intensidad     float64        `json:"intensidad"`   // Cuánta motivación aporta (0.0 - 1.0)
	Costo          float64        `json:"costo"`        // Costo estimado para la empresa
	Disponibilidad Disponibilidad `json:"disponibilidad"`
	Cupos          int            `json:"cupos"`        // Solo si es limitado
	CuposUsados    int            `json:"cupos_usados"`
	Activo         bool           `json:"activo"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

// HayCupos retorna true si el incentivo aún tiene disponibilidad.
func (i Incentivo) HayCupos() bool {
	if i.Disponibilidad == DisponibilidadPermanente || i.Disponibilidad == DisponibilidadRecurrente {
		return true
	}
	return i.CuposUsados < i.Cupos
}

// Elegibilidad representa un criterio que un empleado debe cumplir para acceder al incentivo.
type Elegibilidad struct {
	ID          int64  `json:"id"`
	IncentivoID int64  `json:"incentivo_id"`
	Campo       string `json:"campo"`  // Ej: "departamento", "cargo", "antiguedad_meses"
	Operador    string `json:"operador"` // "eq", "gte", "lte"
	Valor       string `json:"valor"`
}

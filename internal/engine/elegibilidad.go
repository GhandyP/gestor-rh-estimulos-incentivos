package engine

import (
	"strconv"
	"strings"
	"time"

	"estimulos-incentivos/internal/domain"
)

// EvaluarElegibilidad verifica si un empleado cumple con todos los criterios de elegibilidad.
// Retorna true solo si el empleado satisface cada criterio.
func EvaluarElegibilidad(empleado domain.Empleado, criterios []domain.Elegibilidad) bool {
	if len(criterios) == 0 {
		return true // sin criterios, todos son elegibles
	}
	for _, c := range criterios {
		if !evaluarCriterio(empleado, c) {
			return false
		}
	}
	return true
}

// evaluarCriterio evalúa un único criterio de elegibilidad contra un empleado.
func evaluarCriterio(empleado domain.Empleado, c domain.Elegibilidad) bool {
	switch c.Campo {
	case "departamento":
		return evaluarString(string(empleado.Departamento), c.Operador, c.Valor)
	case "cargo":
		return evaluarString(empleado.Cargo, c.Operador, c.Valor)
	case "antiguedad_meses":
		meses := mesesDesde(empleado.FechaIngreso)
		valor, err := strconv.Atoi(c.Valor)
		if err != nil {
			return false
		}
		return evaluarNumero(meses, c.Operador, valor)
	default:
		return false
	}
}

// evaluarString evalúa operadores eq (=) y neq (!=) para campos de texto.
func evaluarString(campo, operador, valor string) bool {
	switch operador {
	case "eq":
		return strings.EqualFold(campo, valor)
	case "neq":
		return !strings.EqualFold(campo, valor)
	default:
		return false
	}
}

// evaluarNumero evalúa operadores eq, gte, lte, gt, lt para campos numéricos.
func evaluarNumero(campo int, operador string, valor int) bool {
	switch operador {
	case "eq":
		return campo == valor
	case "gte":
		return campo >= valor
	case "lte":
		return campo <= valor
	case "gt":
		return campo > valor
	case "lt":
		return campo < valor
	default:
		return false
	}
}

// mesesDesde calcula la cantidad de meses transcurridos desde una fecha hasta ahora.
func mesesDesde(fecha time.Time) int {
	ahora := time.Now()
	anios := ahora.Year() - fecha.Year()
	meses := int(ahora.Month()) - int(fecha.Month())
	total := anios*12 + meses
	if ahora.Day() < fecha.Day() {
		total--
	}
	if total < 0 {
		return 0
	}
	return total
}

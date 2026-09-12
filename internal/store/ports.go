package store

import (
	"context"
	"time"

	"estimulos-incentivos/internal/domain"
)

// Transaction exposes the domain operations that may participate in a
// repository transaction. Concrete database transaction types stay behind
// this boundary.
type Transaction interface {
	CreateEmpleado(context.Context, *domain.Empleado) error
	CreatePerfilMAP(context.Context, *domain.PerfilMAP) error
	CreateUmbral(context.Context, *domain.Umbral) error

	TransitionEstimulo(context.Context, int64, time.Time) (bool, error)
	GetUmbral(context.Context, int64) (*domain.Umbral, error)
	AddHistorial(context.Context, int64, domain.PuntoHistorial) error
	GetHistorial(context.Context, int64) ([]domain.PuntoHistorial, error)
	UpdateUmbral(context.Context, *domain.Umbral) error
}

// Repository is the persistence port consumed by the service layer.
type Repository interface {
	WithTx(context.Context, func(Transaction) error) error

	CreateEmpleado(context.Context, *domain.Empleado) error
	GetEmpleado(context.Context, int64) (*domain.Empleado, error)
	ListEmpleados(context.Context) ([]domain.Empleado, error)
	UpdateEmpleado(context.Context, *domain.Empleado) error
	DeleteEmpleado(context.Context, int64) error
	CountEmpleados(context.Context) (int, error)

	CreatePerfilMAP(context.Context, *domain.PerfilMAP) error
	GetPerfilMAP(context.Context, int64) (*domain.PerfilMAP, error)
	ListPerfilesMAP(context.Context) ([]domain.PerfilMAP, error)
	UpdatePerfilMAP(context.Context, *domain.PerfilMAP) error
	GetUmbral(context.Context, int64) (*domain.Umbral, error)
	CreateUmbral(context.Context, *domain.Umbral) error
	GetHistorial(context.Context, int64) ([]domain.PuntoHistorial, error)

	CreateIncentivo(context.Context, *domain.Incentivo) error
	GetIncentivo(context.Context, int64) (*domain.Incentivo, error)
	ListIncentivos(context.Context) ([]domain.Incentivo, error)
	UpdateIncentivo(context.Context, *domain.Incentivo) error
	DeleteIncentivo(context.Context, int64) error
	CreateElegibilidad(context.Context, *domain.Elegibilidad) error
	ListElegibilidades(context.Context, int64) ([]domain.Elegibilidad, error)
	DeleteElegibilidad(context.Context, int64) error

	CreateNudge(context.Context, *domain.Nudge) error
	GetNudge(context.Context, int64) (*domain.Nudge, error)
	ListNudges(context.Context) ([]domain.Nudge, error)
	UpdateNudge(context.Context, *domain.Nudge) error

	CreateEstimulo(context.Context, *domain.Estimulo) error
	GetEstimulo(context.Context, int64) (*domain.Estimulo, error)
	ListEstimulos(context.Context, string) ([]domain.Estimulo, error)
	ListEstimulosPorEmpleado(context.Context, int64) ([]domain.Estimulo, error)
}

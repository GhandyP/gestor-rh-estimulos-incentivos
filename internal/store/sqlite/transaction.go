package sqlite

import (
	"context"
	"database/sql"
	"time"

	"estimulos-incentivos/internal/domain"
	"estimulos-incentivos/internal/store"
)

// transaction adapts the concrete SQLite transaction to the neutral store
// transaction port. SQL remains private to this package.
type transaction struct {
	tx *sql.Tx
}

var _ store.Transaction = (*transaction)(nil)

func (t *transaction) CreateEmpleado(ctx context.Context, e *domain.Empleado) error {
	return createEmpleado(ctx, t.tx, e)
}

func (t *transaction) CreatePerfilMAP(ctx context.Context, p *domain.PerfilMAP) error {
	return createPerfilMAP(ctx, t.tx, p)
}

func (t *transaction) CreateUmbral(ctx context.Context, u *domain.Umbral) error {
	return createUmbral(ctx, t.tx, u)
}

func (t *transaction) TransitionEstimulo(ctx context.Context, id int64, at time.Time) (bool, error) {
	return transitionEstimulo(ctx, t.tx, id, at)
}

func (t *transaction) GetUmbral(ctx context.Context, empleadoID int64) (*domain.Umbral, error) {
	return getUmbral(ctx, t.tx, empleadoID)
}

func (t *transaction) AddHistorial(ctx context.Context, umbralID int64, p domain.PuntoHistorial) error {
	return addHistorial(ctx, t.tx, umbralID, p)
}

func (t *transaction) GetHistorial(ctx context.Context, umbralID int64) ([]domain.PuntoHistorial, error) {
	return getHistorial(ctx, t.tx, umbralID)
}

func (t *transaction) UpdateUmbral(ctx context.Context, u *domain.Umbral) error {
	return updateUmbral(ctx, t.tx, u)
}

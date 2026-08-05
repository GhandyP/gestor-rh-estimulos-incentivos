package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"estimulos-incentivos/internal/domain"
)

func (s *Store) CreateEstimulo(ctx context.Context, e *domain.Estimulo) error {
	now := time.Now().UTC()
	var fechaAplicado *string
	if e.FechaAplicado != nil {
		f := e.FechaAplicado.Format(time.RFC3339)
		fechaAplicado = &f
	}
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO estimulos (empleado_id, tipo, contenido, intensidad, canal, estado, fecha_ideal, fecha_aplicado, origen_recomendacion, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		e.EmpleadoID, e.Tipo, e.Contenido, e.Intensidad, e.Canal, e.Estado, e.FechaIdeal.Format(time.RFC3339), fechaAplicado, e.OrigenRecomendacion, now,
	)
	if err != nil {
		return fmt.Errorf("create estimulo: %w", err)
	}
	id, _ := res.LastInsertId()
	e.ID = id
	e.CreatedAt = now
	return nil
}

func (s *Store) GetEstimulo(ctx context.Context, id int64) (*domain.Estimulo, error) {
	e := &domain.Estimulo{}
	var fechaIdeal, createdAt string
	var fechaAplicado sql.NullString
	err := s.db.QueryRowContext(ctx,
		`SELECT id, empleado_id, tipo, contenido, intensidad, canal, estado, fecha_ideal, fecha_aplicado, origen_recomendacion, created_at
		 FROM estimulos WHERE id = ?`, id,
	).Scan(&e.ID, &e.EmpleadoID, &e.Tipo, &e.Contenido, &e.Intensidad, &e.Canal, &e.Estado, &fechaIdeal, &fechaAplicado, &e.OrigenRecomendacion, &createdAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get estimulo: %w", err)
	}
	e.FechaIdeal, _ = time.Parse(time.RFC3339, fechaIdeal)
	if fechaAplicado.Valid {
		t, _ := time.Parse(time.RFC3339, fechaAplicado.String)
		e.FechaAplicado = &t
	}
	e.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	return e, nil
}

func (s *Store) ListEstimulosPorEmpleado(ctx context.Context, empleadoID int64) ([]domain.Estimulo, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, empleado_id, tipo, contenido, intensidad, canal, estado, fecha_ideal, fecha_aplicado, origen_recomendacion, created_at
		 FROM estimulos WHERE empleado_id = ? ORDER BY created_at DESC`, empleadoID,
	)
	if err != nil {
		return nil, fmt.Errorf("list estimulos: %w", err)
	}
	defer rows.Close()

	var estimulos []domain.Estimulo
	for rows.Next() {
		var e domain.Estimulo
		var fechaIdeal, createdAt string
		var fechaAplicado sql.NullString
		if err := rows.Scan(&e.ID, &e.EmpleadoID, &e.Tipo, &e.Contenido, &e.Intensidad, &e.Canal, &e.Estado, &fechaIdeal, &fechaAplicado, &e.OrigenRecomendacion, &createdAt); err != nil {
			return nil, fmt.Errorf("scan estimulo: %w", err)
		}
		e.FechaIdeal, _ = time.Parse(time.RFC3339, fechaIdeal)
		if fechaAplicado.Valid {
			t, _ := time.Parse(time.RFC3339, fechaAplicado.String)
			e.FechaAplicado = &t
		}
		e.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		estimulos = append(estimulos, e)
	}
	return estimulos, rows.Err()
}

func (s *Store) ListEstimulosPendientes(ctx context.Context) ([]domain.Estimulo, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, empleado_id, tipo, contenido, intensidad, canal, estado, fecha_ideal, fecha_aplicado, origen_recomendacion, created_at
		 FROM estimulos WHERE estado = 'pendiente' ORDER BY fecha_ideal`)
	if err != nil {
		return nil, fmt.Errorf("list pendientes: %w", err)
	}
	defer rows.Close()

	var estimulos []domain.Estimulo
	for rows.Next() {
		var e domain.Estimulo
		var fechaIdeal, createdAt string
		var fechaAplicado sql.NullString
		if err := rows.Scan(&e.ID, &e.EmpleadoID, &e.Tipo, &e.Contenido, &e.Intensidad, &e.Canal, &e.Estado, &fechaIdeal, &fechaAplicado, &e.OrigenRecomendacion, &createdAt); err != nil {
			return nil, fmt.Errorf("scan estimulo: %w", err)
		}
		e.FechaIdeal, _ = time.Parse(time.RFC3339, fechaIdeal)
		if fechaAplicado.Valid {
			t, _ := time.Parse(time.RFC3339, fechaAplicado.String)
			e.FechaAplicado = &t
		}
		e.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		estimulos = append(estimulos, e)
	}
	return estimulos, rows.Err()
}

// TransitionEstimuloTx marca un estímulo como aplicado de forma condicional
// dentro de la transacción. Solo transiciona estímulos pendientes; retorna
// true si la transición ocurrió y false si el estímulo ya no estaba pendiente.
func (s *Store) TransitionEstimuloTx(ctx context.Context, tx *sql.Tx, id int64, at time.Time) (bool, error) {
	res, err := tx.ExecContext(ctx,
		`UPDATE estimulos SET estado='aplicado', fecha_aplicado=? WHERE id=? AND estado='pendiente'`,
		at.Format(time.RFC3339), id,
	)
	if err != nil {
		return false, fmt.Errorf("transition estimulo: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("transition estimulo rows: %w", err)
	}
	return n > 0, nil
}

// Umbrales

func (s *Store) CreateUmbral(ctx context.Context, u *domain.Umbral) error {
	return createUmbral(ctx, s.db, u)
}

func (s *Store) CreateUmbralTx(ctx context.Context, tx *sql.Tx, u *domain.Umbral) error {
	return createUmbral(ctx, tx, u)
}

func createUmbral(ctx context.Context, db execer, u *domain.Umbral) error {
	now := time.Now().UTC()
	var fechaUlt *string
	if u.FechaUltimoEstimulo != nil {
		f := u.FechaUltimoEstimulo.Format(time.RFC3339)
		fechaUlt = &f
	}
	res, err := db.ExecContext(ctx,
		`INSERT INTO umbrales (empleado_id, umbral_absoluto, umbral_diferencial, ultimo_estimulo, fecha_ultimo_estimulo, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		u.EmpleadoID, u.UmbralAbsoluto, u.UmbralDiferencial, u.UltimoEstimulo, fechaUlt, now,
	)
	if err != nil {
		return fmt.Errorf("create umbral: %w", err)
	}
	id, _ := res.LastInsertId()
	u.ID = id
	u.UpdatedAt = now
	return nil
}

func (s *Store) GetUmbral(ctx context.Context, empleadoID int64) (*domain.Umbral, error) {
	return getUmbral(ctx, s.db, empleadoID)
}

func (s *Store) GetUmbralTx(ctx context.Context, tx *sql.Tx, empleadoID int64) (*domain.Umbral, error) {
	return getUmbral(ctx, tx, empleadoID)
}

func getUmbral(ctx context.Context, db queryer, empleadoID int64) (*domain.Umbral, error) {
	u := &domain.Umbral{}
	var updatedAt string
	var fechaUltimo sql.NullString
	err := db.QueryRowContext(ctx,
		`SELECT id, empleado_id, umbral_absoluto, umbral_diferencial, ultimo_estimulo, fecha_ultimo_estimulo, updated_at
		 FROM umbrales WHERE empleado_id = ?`, empleadoID,
	).Scan(&u.ID, &u.EmpleadoID, &u.UmbralAbsoluto, &u.UmbralDiferencial, &u.UltimoEstimulo, &fechaUltimo, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get umbral: %w", err)
	}
	if fechaUltimo.Valid {
		t, _ := time.Parse(time.RFC3339, fechaUltimo.String)
		u.FechaUltimoEstimulo = &t
	}
	u.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	return u, nil
}

func (s *Store) UpdateUmbral(ctx context.Context, u *domain.Umbral) error {
	return updateUmbral(ctx, s.db, u)
}

func (s *Store) UpdateUmbralTx(ctx context.Context, tx *sql.Tx, u *domain.Umbral) error {
	return updateUmbral(ctx, tx, u)
}

func updateUmbral(ctx context.Context, db execer, u *domain.Umbral) error {
	now := time.Now().UTC()
	var fechaUlt *string
	if u.FechaUltimoEstimulo != nil {
		f := u.FechaUltimoEstimulo.Format(time.RFC3339)
		fechaUlt = &f
	}
	_, err := db.ExecContext(ctx,
		`UPDATE umbrales SET umbral_absoluto=?, umbral_diferencial=?, ultimo_estimulo=?, fecha_ultimo_estimulo=?, updated_at=?
		 WHERE id=?`,
		u.UmbralAbsoluto, u.UmbralDiferencial, u.UltimoEstimulo, fechaUlt, now, u.ID,
	)
	if err != nil {
		return fmt.Errorf("update umbral: %w", err)
	}
	u.UpdatedAt = now
	return nil
}

// Historial

func (s *Store) AddHistorial(ctx context.Context, umbralID int64, p domain.PuntoHistorial) error {
	return addHistorial(ctx, s.db, umbralID, p)
}

func (s *Store) AddHistorialTx(ctx context.Context, tx *sql.Tx, umbralID int64, p domain.PuntoHistorial) error {
	return addHistorial(ctx, tx, umbralID, p)
}

func addHistorial(ctx context.Context, db execer, umbralID int64, p domain.PuntoHistorial) error {
	_, err := db.ExecContext(ctx,
		`INSERT INTO historial_estimulos (umbral_id, fecha, intensidad, respuesta_map, tipo) VALUES (?, ?, ?, ?, ?)`,
		umbralID, p.Fecha.Format(time.RFC3339), p.Intensidad, p.RespuestaMAP, p.Tipo,
	)
	return err
}

func (s *Store) GetHistorial(ctx context.Context, umbralID int64) ([]domain.PuntoHistorial, error) {
	return getHistorial(ctx, s.db, umbralID)
}

func (s *Store) GetHistorialTx(ctx context.Context, tx *sql.Tx, umbralID int64) ([]domain.PuntoHistorial, error) {
	return getHistorial(ctx, tx, umbralID)
}

func getHistorial(ctx context.Context, db queryer, umbralID int64) ([]domain.PuntoHistorial, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT fecha, intensidad, respuesta_map, COALESCE(tipo, '') FROM historial_estimulos WHERE umbral_id = ? ORDER BY fecha`, umbralID,
	)
	if err != nil {
		return nil, fmt.Errorf("get historial: %w", err)
	}
	defer rows.Close()

	var historial []domain.PuntoHistorial
	for rows.Next() {
		var p domain.PuntoHistorial
		var fecha string
		if err := rows.Scan(&fecha, &p.Intensidad, &p.RespuestaMAP, &p.Tipo); err != nil {
			return nil, fmt.Errorf("scan historial: %w", err)
		}
		p.Fecha, _ = time.Parse(time.RFC3339, fecha)
		historial = append(historial, p)
	}
	return historial, rows.Err()
}

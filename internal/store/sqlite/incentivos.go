package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"estimulos-incentivos/internal/domain"
)

func (s *Store) CreateIncentivo(ctx context.Context, i *domain.Incentivo) error {
	now := time.Now().UTC()
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO incentivos (nombre, descripcion, tipo, intensidad, costo, disponibilidad, cupos, cupos_usados, activo, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		i.Nombre, i.Descripcion, i.Tipo, i.Intensidad, i.Costo, i.Disponibilidad, i.Cupos, i.CuposUsados, i.Activo, now, now,
	)
	if err != nil {
		return fmt.Errorf("create incentivo: %w", err)
	}
	id, _ := res.LastInsertId()
	i.ID = id
	i.CreatedAt = now
	i.UpdatedAt = now
	return nil
}

func (s *Store) GetIncentivo(ctx context.Context, id int64) (*domain.Incentivo, error) {
	i := &domain.Incentivo{}
	var createdAt, updatedAt string
	err := s.db.QueryRowContext(ctx,
		`SELECT id, nombre, descripcion, tipo, intensidad, costo, disponibilidad, cupos, cupos_usados, activo, created_at, updated_at
		 FROM incentivos WHERE id = ?`, id,
	).Scan(&i.ID, &i.Nombre, &i.Descripcion, &i.Tipo, &i.Intensidad, &i.Costo, &i.Disponibilidad, &i.Cupos, &i.CuposUsados, &i.Activo, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get incentivo: %w", err)
	}
	i.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	i.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	return i, nil
}

func (s *Store) ListIncentivos(ctx context.Context) ([]domain.Incentivo, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, nombre, descripcion, tipo, intensidad, costo, disponibilidad, cupos, cupos_usados, activo, created_at, updated_at
		 FROM incentivos WHERE activo = 1 ORDER BY tipo, nombre`)
	if err != nil {
		return nil, fmt.Errorf("list incentivos: %w", err)
	}
	defer rows.Close()

	var incentivos []domain.Incentivo
	for rows.Next() {
		var i domain.Incentivo
		var createdAt, updatedAt string
		if err := rows.Scan(&i.ID, &i.Nombre, &i.Descripcion, &i.Tipo, &i.Intensidad, &i.Costo, &i.Disponibilidad, &i.Cupos, &i.CuposUsados, &i.Activo, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("scan incentivo: %w", err)
		}
		i.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		i.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		incentivos = append(incentivos, i)
	}
	return incentivos, rows.Err()
}

func (s *Store) UpdateIncentivo(ctx context.Context, i *domain.Incentivo) error {
	now := time.Now().UTC()
	_, err := s.db.ExecContext(ctx,
		`UPDATE incentivos SET nombre=?, descripcion=?, tipo=?, intensidad=?, costo=?, disponibilidad=?, cupos=?, cupos_usados=?, activo=?, updated_at=?
		 WHERE id=?`,
		i.Nombre, i.Descripcion, i.Tipo, i.Intensidad, i.Costo, i.Disponibilidad, i.Cupos, i.CuposUsados, i.Activo, now, i.ID,
	)
	if err != nil {
		return fmt.Errorf("update incentivo: %w", err)
	}
	i.UpdatedAt = now
	return nil
}

func (s *Store) DeleteIncentivo(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM incentivos WHERE id=?`, id)
	return err
}

// Elegibilidades

func (s *Store) CreateElegibilidad(ctx context.Context, e *domain.Elegibilidad) error {
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO elegibilidades (incentivo_id, campo, operador, valor) VALUES (?, ?, ?, ?)`,
		e.IncentivoID, e.Campo, e.Operador, e.Valor,
	)
	if err != nil {
		return fmt.Errorf("create elegibilidad: %w", err)
	}
	id, _ := res.LastInsertId()
	e.ID = id
	return nil
}

func (s *Store) ListElegibilidades(ctx context.Context, incentivoID int64) ([]domain.Elegibilidad, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, incentivo_id, campo, operador, valor FROM elegibilidades WHERE incentivo_id = ?`, incentivoID,
	)
	if err != nil {
		return nil, fmt.Errorf("list elegibilidades: %w", err)
	}
	defer rows.Close()

	var elegs []domain.Elegibilidad
	for rows.Next() {
		var e domain.Elegibilidad
		if err := rows.Scan(&e.ID, &e.IncentivoID, &e.Campo, &e.Operador, &e.Valor); err != nil {
			return nil, fmt.Errorf("scan elegibilidad: %w", err)
		}
		elegs = append(elegs, e)
	}
	return elegs, rows.Err()
}

func (s *Store) DeleteElegibilidad(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM elegibilidades WHERE id=?`, id)
	return err
}

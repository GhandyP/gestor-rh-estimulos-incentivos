package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"estimulos-incentivos/internal/domain"
)

func (s *Store) CreateNudge(ctx context.Context, n *domain.Nudge) error {
	now := time.Now().UTC()
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO nudges (nombre, descripcion, tipo, ambito, target_id, activo, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		n.Nombre, n.Descripcion, n.Tipo, n.Ambito, n.TargetID, n.Activo, now, now,
	)
	if err != nil {
		return fmt.Errorf("create nudge: %w", err)
	}
	id, _ := res.LastInsertId()
	n.ID = id
	n.CreatedAt = now
	n.UpdatedAt = now
	return nil
}

func (s *Store) GetNudge(ctx context.Context, id int64) (*domain.Nudge, error) {
	n := &domain.Nudge{}
	var createdAt, updatedAt string
	err := s.db.QueryRowContext(ctx,
		`SELECT id, nombre, descripcion, tipo, ambito, target_id, activo, created_at, updated_at
		 FROM nudges WHERE id = ?`, id,
	).Scan(&n.ID, &n.Nombre, &n.Descripcion, &n.Tipo, &n.Ambito, &n.TargetID, &n.Activo, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get nudge: %w", err)
	}
	n.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	n.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	return n, nil
}

func (s *Store) ListNudges(ctx context.Context) ([]domain.Nudge, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, nombre, descripcion, tipo, ambito, target_id, activo, created_at, updated_at
		 FROM nudges WHERE activo = 1 ORDER BY tipo, nombre`)
	if err != nil {
		return nil, fmt.Errorf("list nudges: %w", err)
	}
	defer rows.Close()

	var nudges []domain.Nudge
	for rows.Next() {
		var n domain.Nudge
		var createdAt, updatedAt string
		if err := rows.Scan(&n.ID, &n.Nombre, &n.Descripcion, &n.Tipo, &n.Ambito, &n.TargetID, &n.Activo, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("scan nudge: %w", err)
		}
		n.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		n.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		nudges = append(nudges, n)
	}
	return nudges, rows.Err()
}

func (s *Store) UpdateNudge(ctx context.Context, n *domain.Nudge) error {
	now := time.Now().UTC()
	_, err := s.db.ExecContext(ctx,
		`UPDATE nudges SET nombre=?, descripcion=?, tipo=?, ambito=?, target_id=?, activo=?, updated_at=?
		 WHERE id=?`,
		n.Nombre, n.Descripcion, n.Tipo, n.Ambito, n.TargetID, n.Activo, now, n.ID,
	)
	return err
}

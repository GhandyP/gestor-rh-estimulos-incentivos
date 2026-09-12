package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"estimulos-incentivos/internal/domain"
	"estimulos-incentivos/internal/store"
)

func (s *Store) CreateEmpleado(ctx context.Context, e *domain.Empleado) error {
	return createEmpleado(ctx, s.db, e)
}

func createEmpleado(ctx context.Context, db execer, e *domain.Empleado) error {
	now := time.Now().UTC()
	res, err := db.ExecContext(ctx,
		`INSERT INTO empleados (nombre, email, cargo, departamento, fecha_ingreso, activo, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		e.Nombre, e.Email, e.Cargo, e.Departamento, e.FechaIngreso.Format(time.RFC3339), e.Activo, now, now,
	)
	if err != nil {
		return fmt.Errorf("create empleado: %w", err)
	}
	id, _ := res.LastInsertId()
	e.ID = id
	e.CreatedAt = now
	e.UpdatedAt = now
	return nil
}

func (s *Store) CreateEmpleadoConPerfil(ctx context.Context, e *domain.Empleado, p *domain.PerfilMAP) error {
	return s.WithTx(ctx, func(tx store.Transaction) error {
		if err := tx.CreateEmpleado(ctx, e); err != nil {
			return err
		}
		p.EmpleadoID = e.ID
		return tx.CreatePerfilMAP(ctx, p)
	})
}

func (s *Store) GetEmpleado(ctx context.Context, id int64) (*domain.Empleado, error) {
	e := &domain.Empleado{}
	var fechaIng, createdAt, updatedAt string
	err := s.db.QueryRowContext(ctx,
		`SELECT id, nombre, email, cargo, departamento, fecha_ingreso, activo, created_at, updated_at
		 FROM empleados WHERE id = ?`, id,
	).Scan(&e.ID, &e.Nombre, &e.Email, &e.Cargo, &e.Departamento, &fechaIng, &e.Activo, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get empleado: %w", err)
	}
	e.FechaIngreso, _ = time.Parse(time.RFC3339, fechaIng)
	e.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	e.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	return e, nil
}

func (s *Store) ListEmpleados(ctx context.Context) ([]domain.Empleado, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, nombre, email, cargo, departamento, fecha_ingreso, activo, created_at, updated_at
		 FROM empleados ORDER BY nombre`)
	if err != nil {
		return nil, fmt.Errorf("list empleados: %w", err)
	}
	defer rows.Close()

	var empleados []domain.Empleado
	for rows.Next() {
		var e domain.Empleado
		var fechaIng, createdAt, updatedAt string
		if err := rows.Scan(&e.ID, &e.Nombre, &e.Email, &e.Cargo, &e.Departamento, &fechaIng, &e.Activo, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("scan empleado: %w", err)
		}
		e.FechaIngreso, _ = time.Parse(time.RFC3339, fechaIng)
		e.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		e.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		empleados = append(empleados, e)
	}
	return empleados, rows.Err()
}

func (s *Store) CountEmpleados(ctx context.Context) (int, error) {
	var count int
	if err := s.db.QueryRowContext(ctx, "SELECT count(*) FROM empleados").Scan(&count); err != nil {
		return 0, fmt.Errorf("count empleados: %w", err)
	}
	return count, nil
}

func (s *Store) UpdateEmpleado(ctx context.Context, e *domain.Empleado) error {
	now := time.Now().UTC()
	_, err := s.db.ExecContext(ctx,
		`UPDATE empleados SET nombre=?, email=?, cargo=?, departamento=?, fecha_ingreso=?, activo=?, updated_at=?
		 WHERE id=?`,
		e.Nombre, e.Email, e.Cargo, e.Departamento, e.FechaIngreso.Format(time.RFC3339), e.Activo, now, e.ID,
	)
	if err != nil {
		return fmt.Errorf("update empleado: %w", err)
	}
	e.UpdatedAt = now
	return nil
}

func (s *Store) DeleteEmpleado(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM empleados WHERE id=?`, id)
	if err != nil {
		return fmt.Errorf("delete empleado: %w", err)
	}
	return nil
}

// --- PerfilMAP ---

func (s *Store) CreatePerfilMAP(ctx context.Context, p *domain.PerfilMAP) error {
	return createPerfilMAP(ctx, s.db, p)
}

func createPerfilMAP(ctx context.Context, db execer, p *domain.PerfilMAP) error {
	now := time.Now().UTC()
	res, err := db.ExecContext(ctx,
		`INSERT INTO perfiles_map (empleado_id, motivacion, habilidad, prompt, sensibilidad, confiabilidad, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		p.EmpleadoID, p.Motivacion, p.Habilidad, p.Prompt, p.Sensibilidad, p.Confiabilidad, now,
	)
	if err != nil {
		return fmt.Errorf("create perfil: %w", err)
	}
	id, _ := res.LastInsertId()
	p.ID = id
	p.UpdatedAt = now
	return nil
}

func (s *Store) GetPerfilMAP(ctx context.Context, empleadoID int64) (*domain.PerfilMAP, error) {
	p := &domain.PerfilMAP{}
	var updatedAt string
	err := s.db.QueryRowContext(ctx,
		`SELECT id, empleado_id, motivacion, habilidad, prompt, sensibilidad, confiabilidad, updated_at
		 FROM perfiles_map WHERE empleado_id = ?`, empleadoID,
	).Scan(&p.ID, &p.EmpleadoID, &p.Motivacion, &p.Habilidad, &p.Prompt, &p.Sensibilidad, &p.Confiabilidad, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get perfil: %w", err)
	}
	p.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	return p, nil
}

func (s *Store) ListPerfilesMAP(ctx context.Context) ([]domain.PerfilMAP, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, empleado_id, motivacion, habilidad, prompt, sensibilidad, confiabilidad, updated_at
		 FROM perfiles_map ORDER BY empleado_id`)
	if err != nil {
		return nil, fmt.Errorf("list perfiles: %w", err)
	}
	defer rows.Close()

	var perfiles []domain.PerfilMAP
	for rows.Next() {
		var p domain.PerfilMAP
		var updatedAt string
		if err := rows.Scan(&p.ID, &p.EmpleadoID, &p.Motivacion, &p.Habilidad, &p.Prompt, &p.Sensibilidad, &p.Confiabilidad, &updatedAt); err != nil {
			return nil, fmt.Errorf("scan perfil: %w", err)
		}
		p.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		perfiles = append(perfiles, p)
	}
	return perfiles, rows.Err()
}

func (s *Store) UpdatePerfilMAP(ctx context.Context, p *domain.PerfilMAP) error {
	now := time.Now().UTC()
	_, err := s.db.ExecContext(ctx,
		`UPDATE perfiles_map SET motivacion=?, habilidad=?, prompt=?, sensibilidad=?, confiabilidad=?, updated_at=?
		 WHERE id=?`,
		p.Motivacion, p.Habilidad, p.Prompt, p.Sensibilidad, p.Confiabilidad, now, p.ID,
	)
	if err != nil {
		return fmt.Errorf("update perfil: %w", err)
	}
	p.UpdatedAt = now
	return nil
}

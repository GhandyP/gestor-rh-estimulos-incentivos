package service

import (
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"time"

	"estimulos-incentivos/internal/domain"
	"estimulos-incentivos/internal/engine"
	"estimulos-incentivos/internal/store/sqlite"
)

type Service struct {
	store *sqlite.Store
}

func New(store *sqlite.Store) *Service {
	return &Service{store: store}
}

// Empleados

func (s *Service) CreateEmpleado(ctx context.Context, nombre, email, cargo, departamento string) (*domain.Empleado, error) {
	e := &domain.Empleado{
		Nombre:       nombre,
		Email:        email,
		Cargo:        cargo,
		Departamento: domain.Departamento(departamento),
		Activo:       true,
	}

	if err := s.store.CreateEmpleado(ctx, e); err != nil {
		return nil, err
	}

	// Crear perfil MAP y umbral por defecto
	perfil := &domain.PerfilMAP{
		EmpleadoID:    e.ID,
		Motivacion:    0.50,
		Habilidad:     0.50,
		Prompt:        0.60,
		Sensibilidad:  domain.SensibilidadDesarrollo,
		Confiabilidad: 0.30,
	}
	if err := s.store.CreatePerfilMAP(ctx, perfil); err != nil {
		return nil, err
	}

	umbral := engine.UmbralInicial(e.ID)
	if err := s.store.CreateUmbral(ctx, &umbral); err != nil {
		return nil, err
	}

	return e, nil
}

func (s *Service) GetEmpleado(ctx context.Context, id int64) (*domain.Empleado, error) {
	return s.store.GetEmpleado(ctx, id)
}

func (s *Service) ListEmpleados(ctx context.Context) ([]domain.Empleado, error) {
	return s.store.ListEmpleados(ctx)
}

func (s *Service) EmpleadoConPerfil(ctx context.Context, id int64) (*domain.Empleado, *domain.PerfilMAP, *domain.Umbral, error) {
	e, err := s.store.GetEmpleado(ctx, id)
	if err != nil || e == nil {
		return nil, nil, nil, err
	}
	p, err := s.store.GetPerfilMAP(ctx, id)
	if err != nil {
		return nil, nil, nil, err
	}
	u, err := s.store.GetUmbral(ctx, id)
	if err != nil {
		return nil, nil, nil, err
	}
	return e, p, u, nil
}

// Perfil MAP

func (s *Service) UpdatePerfilMAP(ctx context.Context, empleadoID int64, motivacion, habilidad, prompt float64, sensibilidad string, confiabilidad float64) error {
	p, err := s.store.GetPerfilMAP(ctx, empleadoID)
	if err != nil || p == nil {
		return fmt.Errorf("perfil no encontrado: %w", err)
	}
	p.Motivacion = motivacion
	p.Habilidad = habilidad
	p.Prompt = prompt
	p.Sensibilidad = domain.Sensibilidad(sensibilidad)
	p.Confiabilidad = confiabilidad
	return s.store.UpdatePerfilMAP(ctx, p)
}

// Incentivos

func (s *Service) CreateIncentivo(ctx context.Context, nombre, descripcion string, tipo domain.TipoIncentivo, intensidad, costo float64, disponibilidad domain.Disponibilidad, cupos int) (*domain.Incentivo, error) {
	i := &domain.Incentivo{
		Nombre:         nombre,
		Descripcion:    descripcion,
		Tipo:           tipo,
		Intensidad:     intensidad,
		Costo:          costo,
		Disponibilidad: disponibilidad,
		Cupos:          cupos,
		Activo:         true,
	}
	if err := s.store.CreateIncentivo(ctx, i); err != nil {
		return nil, err
	}
	return i, nil
}

func (s *Service) ListIncentivos(ctx context.Context) ([]domain.Incentivo, error) {
	return s.store.ListIncentivos(ctx)
}

// Nudges

func (s *Service) CreateNudge(ctx context.Context, nombre, descripcion string, tipo domain.TipoNudge, ambito domain.AmbitoNudge, targetID int64) (*domain.Nudge, error) {
	n := &domain.Nudge{
		Nombre:      nombre,
		Descripcion: descripcion,
		Tipo:        tipo,
		Ambito:      ambito,
		TargetID:    targetID,
		Activo:      true,
	}
	if err := s.store.CreateNudge(ctx, n); err != nil {
		return nil, err
	}
	return n, nil
}

func (s *Service) ListNudges(ctx context.Context) ([]domain.Nudge, error) {
	return s.store.ListNudges(ctx)
}

// Recomendación

func (s *Service) Recomendar(ctx context.Context, empleadoID int64) (*engine.RecomendacionResult, error) {
	e, err := s.store.GetEmpleado(ctx, empleadoID)
	if err != nil || e == nil {
		return nil, fmt.Errorf("empleado no encontrado")
	}

	p, err := s.store.GetPerfilMAP(ctx, empleadoID)
	if err != nil || p == nil {
		return nil, fmt.Errorf("perfil no encontrado")
	}

	u, err := s.store.GetUmbral(ctx, empleadoID)
	if err != nil || u == nil {
		return nil, fmt.Errorf("umbral no encontrado")
	}

	incentivos, err := s.store.ListIncentivos(ctx)
	if err != nil {
		return nil, err
	}

	nudges, err := s.store.ListNudges(ctx)
	if err != nil {
		return nil, err
	}

	result := engine.Recomendar(*e, *p, incentivos, nudges, *u)

	// Si hay estímulo recomendado, persistirlo
	if result.EstimuloRecomendado != nil {
		if err := s.store.CreateEstimulo(ctx, result.EstimuloRecomendado); err != nil {
			return nil, fmt.Errorf("guardar estimulo: %w", err)
		}
	}

	return &result, nil
}

func (s *Service) Analizar(ctx context.Context) (*engine.AnalisisResult, error) {
	empleados, err := s.store.ListEmpleados(ctx)
	if err != nil {
		return nil, err
	}

	perfiles, err := s.store.ListPerfilesMAP(ctx)
	if err != nil {
		return nil, err
	}

	historial := make(map[int64][]domain.PuntoHistorial)
	for _, e := range empleados {
		u, err := s.store.GetUmbral(ctx, e.ID)
		if err != nil || u == nil {
			continue
		}
		h, err := s.store.GetHistorial(ctx, u.ID)
		if err != nil {
			continue
		}
		historial[e.ID] = h
	}

	result := engine.Analizar(empleados, perfiles, historial)
	return &result, nil
}

// Estimulos pendientes

func (s *Service) ListEstimulosPendientes(ctx context.Context) ([]domain.Estimulo, error) {
	return s.store.ListEstimulosPendientes(ctx)
}

func (s *Service) ApplyEstimulo(ctx context.Context, estimuloID int64, respuesta float64) error {
	if err := s.store.ApplyEstimulo(ctx, estimuloID); err != nil {
		return fmt.Errorf("apply estimulo: %w", err)
	}

	estimulo, err := s.store.GetEstimulo(ctx, estimuloID)
	if err != nil || estimulo == nil {
		return fmt.Errorf("get estimulo %d: %w", estimuloID, err)
	}

	umbral, err := s.store.GetUmbral(ctx, estimulo.EmpleadoID)
	if err != nil || umbral == nil {
		return fmt.Errorf("get umbral for empleado %d: %w", estimulo.EmpleadoID, err)
	}

	if err := s.store.AddHistorial(ctx, umbral.ID, domain.PuntoHistorial{
		Fecha:        time.Now(),
		Intensidad:   estimulo.Intensidad,
		RespuestaMAP: respuesta,
		Tipo:         estimulo.Tipo,
	}); err != nil {
		return fmt.Errorf("add historial: %w", err)
	}

	historial, err := s.store.GetHistorial(ctx, umbral.ID)
	if err != nil {
		return fmt.Errorf("get historial: %w", err)
	}

	calibrado := engine.CalibrarUmbral(*umbral, historial)

	if err := s.store.UpdateUmbral(ctx, &calibrado); err != nil {
		return fmt.Errorf("update umbral: %w", err)
	}

	return nil
}

// EmpleadoDetail agrupa toda la información relevante de un empleado.
type EmpleadoDetail struct {
	Empleado   domain.Empleado         `json:"empleado"`
	Perfil     *domain.PerfilMAP       `json:"perfil"`
	Umbral     *domain.Umbral          `json:"umbral"`
	Historial  []domain.PuntoHistorial `json:"historial"`
	Estimulos  []domain.Estimulo       `json:"estimulos"`
	SobreCurva bool                    `json:"sobre_curva"`
}

// GetEmpleadoDetail obtiene toda la información compuesta de un empleado.
func (s *Service) GetEmpleadoDetail(ctx context.Context, id int64) (*EmpleadoDetail, error) {
	e, err := s.store.GetEmpleado(ctx, id)
	if err != nil || e == nil {
		return nil, fmt.Errorf("empleado %d no encontrado", id)
	}

	p, err := s.store.GetPerfilMAP(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("perfil MAP: %w", err)
	}

	u, err := s.store.GetUmbral(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("umbral: %w", err)
	}

	var historial []domain.PuntoHistorial
	var sobreCurva bool
	if u != nil {
		historial, err = s.store.GetHistorial(ctx, u.ID)
		if err != nil {
			historial = []domain.PuntoHistorial{}
		}
		if p != nil {
			sobreCurva = p.CurvaAccion(0.25)
		}
	}

	estimulos, err := s.store.ListEstimulosPorEmpleado(ctx, id)
	if err != nil {
		estimulos = []domain.Estimulo{}
	}

	return &EmpleadoDetail{
		Empleado:   *e,
		Perfil:     p,
		Umbral:     u,
		Historial:  historial,
		Estimulos:  estimulos,
		SobreCurva: sobreCurva,
	}, nil
}

// UpdateEmpleado actualiza los datos de un empleado.
func (s *Service) UpdateEmpleado(ctx context.Context, e *domain.Empleado) error {
	return s.store.UpdateEmpleado(ctx, e)
}

// DeleteEmpleado elimina un empleado por ID.
func (s *Service) DeleteEmpleado(ctx context.Context, id int64) error {
	return s.store.DeleteEmpleado(ctx, id)
}

// ImportCSV importa empleados desde un reader CSV.
// Espera columnas: nombre, email, cargo, departamento
// Retorna cantidad de creados y lista de errores por fila.
func (s *Service) ImportCSV(ctx context.Context, reader io.Reader) (int, []string, error) {
	r := csv.NewReader(reader)
	r.TrimLeadingSpace = true
	records, err := r.ReadAll()
	if err != nil {
		return 0, nil, fmt.Errorf("leer CSV: %w", err)
	}

	if len(records) < 2 {
		return 0, []string{"El CSV debe tener al menos una fila de datos además del encabezado"}, nil
	}

	// Saltar encabezado
	creados := 0
	var errores []string

	for i, row := range records[1:] {
		rowNum := i + 2 // 1-indexed + header
		if len(row) < 4 {
			errores = append(errores, fmt.Sprintf("fila %d: columnas insuficientes", rowNum))
			continue
		}
		nombre := strings.TrimSpace(row[0])
		email := strings.TrimSpace(row[1])
		cargo := strings.TrimSpace(row[2])
		departamento := strings.TrimSpace(row[3])

		if nombre == "" || email == "" || cargo == "" || departamento == "" {
			errores = append(errores, fmt.Sprintf("fila %d: campos vacíos", rowNum))
			continue
		}

		_, err := s.CreateEmpleado(ctx, nombre, email, cargo, departamento)
		if err != nil {
			errores = append(errores, fmt.Sprintf("fila %d (%s): %v", rowNum, nombre, err))
			continue
		}
		creados++
	}

	return creados, errores, nil
}

// --- Incentivos extendidos ---

// GetIncentivo obtiene un incentivo por ID.
func (s *Service) GetIncentivo(ctx context.Context, id int64) (*domain.Incentivo, error) {
	return s.store.GetIncentivo(ctx, id)
}

// UpdateIncentivo actualiza un incentivo.
func (s *Service) UpdateIncentivo(ctx context.Context, i *domain.Incentivo) error {
	return s.store.UpdateIncentivo(ctx, i)
}

// DeleteIncentivo elimina un incentivo.
func (s *Service) DeleteIncentivo(ctx context.Context, id int64) error {
	return s.store.DeleteIncentivo(ctx, id)
}

// AddElegibilidad agrega un criterio de elegibilidad a un incentivo.
func (s *Service) AddElegibilidad(ctx context.Context, incentivoID int64, campo, operador, valor string) (*domain.Elegibilidad, error) {
	e := &domain.Elegibilidad{
		IncentivoID: incentivoID,
		Campo:       campo,
		Operador:    operador,
		Valor:       valor,
	}
	if err := s.store.CreateElegibilidad(ctx, e); err != nil {
		return nil, err
	}
	return e, nil
}

// DeleteElegibilidad elimina un criterio de elegibilidad.
func (s *Service) DeleteElegibilidad(ctx context.Context, id int64) error {
	return s.store.DeleteElegibilidad(ctx, id)
}

// GetIncentivosElegibles retorna los incentivos para los cuales un empleado es elegible.
func (s *Service) GetIncentivosElegibles(ctx context.Context, empleadoID int64) ([]domain.Incentivo, error) {
	e, err := s.store.GetEmpleado(ctx, empleadoID)
	if err != nil || e == nil {
		return nil, fmt.Errorf("empleado no encontrado")
	}

	incentivos, err := s.store.ListIncentivos(ctx)
	if err != nil {
		return nil, err
	}

	var elegibles []domain.Incentivo
	for _, inc := range incentivos {
		criterios, err := s.store.ListElegibilidades(ctx, inc.ID)
		if err != nil {
			continue
		}
		if engine.EvaluarElegibilidad(*e, criterios) {
			elegibles = append(elegibles, inc)
		}
	}
	return elegibles, nil
}

// --- Nudges extendidos ---

// GetNudge obtiene un nudge por ID.
func (s *Service) GetNudge(ctx context.Context, id int64) (*domain.Nudge, error) {
	return s.store.GetNudge(ctx, id)
}

// UpdateNudge actualiza un nudge.
func (s *Service) UpdateNudge(ctx context.Context, n *domain.Nudge) error {
	// El store actual no tiene UpdateNudge. Usamos DB directa.
	now := time.Now().UTC()
	_, err := s.store.DB().ExecContext(ctx,
		`UPDATE nudges SET nombre=?, descripcion=?, tipo=?, ambito=?, target_id=?, activo=?, updated_at=?
		 WHERE id=?`,
		n.Nombre, n.Descripcion, n.Tipo, n.Ambito, n.TargetID, n.Activo, now, n.ID,
	)
	return err
}

// --- Estímulos extendidos ---

// GetEstimulo obtiene un estímulo por ID.
func (s *Service) GetEstimulo(ctx context.Context, id int64) (*domain.Estimulo, error) {
	return s.store.GetEstimulo(ctx, id)
}

// ListEstimulos lista estímulos con filtro por estado.
// estado: "pendiente", "aplicado", "todos" o vacío (todos).
func (s *Service) ListEstimulos(ctx context.Context, estado string) ([]domain.Estimulo, error) {
	if estado == "pendiente" {
		return s.store.ListEstimulosPendientes(ctx)
	}

	var rows *sql.Rows
	var err error
	if estado == "aplicado" {
		rows, err = s.store.DB().QueryContext(ctx,
			`SELECT id, empleado_id, tipo, contenido, intensidad, canal, estado, fecha_ideal, fecha_aplicado, origen_recomendacion, created_at
			 FROM estimulos WHERE estado='aplicado' ORDER BY created_at DESC`)
	} else {
		rows, err = s.store.DB().QueryContext(ctx,
			`SELECT id, empleado_id, tipo, contenido, intensidad, canal, estado, fecha_ideal, fecha_aplicado, origen_recomendacion, created_at
			 FROM estimulos ORDER BY created_at DESC`)
	}
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

// --- Operaciones adicionales ---

// ToggleNudge invierte el estado activo de un nudge y retorna el nudge actualizado.
func (s *Service) ToggleNudge(ctx context.Context, id int64) (*domain.Nudge, error) {
	n, err := s.store.GetNudge(ctx, id)
	if err != nil || n == nil {
		return nil, fmt.Errorf("nudge no encontrado")
	}
	n.Activo = !n.Activo
	if err := s.UpdateNudge(ctx, n); err != nil {
		return nil, err
	}
	return n, nil
}

// IncentivoDetailData agrupa un incentivo con sus elegibilidades y empleados elegibles.
type IncentivoDetailData struct {
	Incentivo      *domain.Incentivo      `json:"incentivo"`
	Elegibilidades []domain.Elegibilidad  `json:"elegibilidades"`
	Elegibles      []domain.Empleado      `json:"elegibles"`
}

// GetIncentivoDetail obtiene el detalle completo de un incentivo.
func (s *Service) GetIncentivoDetail(ctx context.Context, id int64) (*IncentivoDetailData, error) {
	inc, err := s.store.GetIncentivo(ctx, id)
	if err != nil || inc == nil {
		return nil, fmt.Errorf("incentivo %d no encontrado", id)
	}

	elegibilidades, err := s.store.ListElegibilidades(ctx, id)
	if err != nil {
		elegibilidades = []domain.Elegibilidad{}
	}

	empleados, err := s.store.ListEmpleados(ctx)
	if err != nil {
		empleados = []domain.Empleado{}
	}

	var elegibles []domain.Empleado
	for _, e := range empleados {
		if e.Activo && engine.EvaluarElegibilidad(e, elegibilidades) {
			elegibles = append(elegibles, e)
		}
	}

	return &IncentivoDetailData{
		Incentivo:      inc,
		Elegibilidades: elegibilidades,
		Elegibles:      elegibles,
	}, nil
}

// Seed data for demo

func (s *Service) Seed(ctx context.Context) error {
	var count int
	s.store.DB().QueryRowContext(ctx, "SELECT count(*) FROM empleados").Scan(&count)
	if count > 0 {
		return nil
	}

	type seedEmpleado struct {
		nombre, email, cargo, depto string
		m, a, p                     float64
		sensibilidad                domain.Sensibilidad
	}

	empleados := []seedEmpleado{
		{"María García", "maria@empresa.com", "Senior Developer", "Ingeniería", 0.75, 0.85, 0.80, domain.SensibilidadDesarrollo},
		{"Juan Pérez", "juan@empresa.com", "Junior Developer", "Ingeniería", 0.40, 0.90, 0.50, domain.SensibilidadReconocimiento},
		{"Ana López", "ana@empresa.com", "Tech Lead", "Ingeniería", 0.85, 0.90, 0.85, domain.SensibilidadReconocimiento},
		{"Carlos Ruiz", "carlos@empresa.com", "Sales Manager", "Ventas", 0.50, 0.60, 0.55, domain.SensibilidadEconomico},
		{"Laura Díaz", "laura@empresa.com", "HR Specialist", "RRHH", 0.70, 0.40, 0.65, domain.SensibilidadBienestar},
		{"Pedro Torres", "pedro@empresa.com", "UX Designer", "Diseño", 0.30, 0.85, 0.35, domain.SensibilidadDesarrollo},
	}

	for _, se := range empleados {
		e := &domain.Empleado{
			Nombre:       se.nombre,
			Email:        se.email,
			Cargo:        se.cargo,
			Departamento: domain.Departamento(se.depto),
			Activo:       true,
		}
		if err := s.store.CreateEmpleado(ctx, e); err != nil {
			return err
		}

		p := &domain.PerfilMAP{
			EmpleadoID:    e.ID,
			Motivacion:    se.m,
			Habilidad:     se.a,
			Prompt:        se.p,
			Sensibilidad:  se.sensibilidad,
			Confiabilidad: 0.70,
		}
		if err := s.store.CreatePerfilMAP(ctx, p); err != nil {
			return err
		}

		u := engine.UmbralInicial(e.ID)
		if err := s.store.CreateUmbral(ctx, &u); err != nil {
			return err
		}
	}

	// Incentivos de ejemplo
	incentivos := []struct {
		nombre, desc string
		tipo         domain.TipoIncentivo
		intensidad   float64
		costo        float64
	}{
		{"Insignia Senior Craftsmanship", "Reconocimiento de excelencia técnica", domain.IncentivoIdentidad, 0.8, 100},
		{"Embajador de Cultura", "Representar valores de la empresa", domain.IncentivoIdentidad, 0.7, 50},
		{"Días Extra de Vacaciones", "3 días adicionales de vacaciones", domain.IncentivoBeneficios, 0.9, 500},
		{"Presupuesto Home Office", "$500 para equipamiento", domain.IncentivoBeneficios, 0.6, 500},
		{"Beca para Conferencia", "Entrada + viaje a conferencia tech", domain.IncentivoFormacion, 0.85, 1500},
		{"Presupuesto de Libros", "$200 en libros técnicos", domain.IncentivoFormacion, 0.5, 200},
		{"Liderar Comité de Innovación", "Liderar iniciativa estratégica", domain.IncentivoProyectoCorporativo, 0.9, 300},
		{"Representante en Evento", "Representar empresa en feria", domain.IncentivoProyectoCorporativo, 0.7, 1000},
	}

	for _, inc := range incentivos {
		i := &domain.Incentivo{
			Nombre:         inc.nombre,
			Descripcion:    inc.desc,
			Tipo:           inc.tipo,
			Intensidad:     inc.intensidad,
			Costo:          inc.costo,
			Disponibilidad: domain.DisponibilidadPermanente,
			Activo:         true,
		}
		if err := s.store.CreateIncentivo(ctx, i); err != nil {
			return err
		}
	}

	// Nudges de ejemplo
	nudges := []struct {
		nombre, desc string
		tipo         domain.TipoNudge
		ambito       domain.AmbitoNudge
	}{
		{"Evaluación 360° Pre-agendada", "Evaluación automática trimestral", domain.NudgeDefaults, domain.AmbitoGlobal},
		{"Onboarding Automático", "Proceso de inducción sin pasos manuales", domain.NudgeDefaults, domain.AmbitoGlobal},
		{"Team Completion Rate", "El 80% de tu equipo ya completó el plan", domain.NudgeSocialProof, domain.AmbitoGlobal},
		{"Feedback como Oportunidad", "Enmarcar feedback positivamente", domain.NudgeFraming, domain.AmbitoGlobal},
		{"One-click Capacitación", "Solicitar capacitación en un click", domain.NudgeFriccion, domain.AmbitoGlobal},
		{"Formularios Auto-completados", "Pre-llenar solicitudes frecuentes", domain.NudgeFriccion, domain.AmbitoGlobal},
	}

	for _, nd := range nudges {
		n := &domain.Nudge{
			Nombre:      nd.nombre,
			Descripcion: nd.desc,
			Tipo:        nd.tipo,
			Ambito:      nd.ambito,
			Activo:      true,
		}
		if err := s.store.CreateNudge(ctx, n); err != nil {
			return err
		}
	}

	return nil
}

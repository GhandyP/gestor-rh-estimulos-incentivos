package handler

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"

	"estimulos-incentivos/internal/domain"
	"estimulos-incentivos/internal/service"
)

// templateFuncs son funciones auxiliares disponibles en todos los templates.
var templateFuncs = template.FuncMap{
	"percent": func(v float64) string { return fmt.Sprintf("%.0f", v*100) },
}

// Handler maneja rutas HTTP para el sistema de Estímulos e Incentivos.
type Handler struct {
	Svc *service.Service
}

// New crea un nuevo Handler.
func New(svc *service.Service) *Handler {
	return &Handler{Svc: svc}
}

// RegisterRoutes registra todas las rutas HTML y API en el mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// --- Páginas HTML ---
	mux.HandleFunc("GET /", h.Dashboard)
	mux.HandleFunc("GET /empleados", h.EmpleadosPage)
	mux.HandleFunc("GET /empleados/{id}", h.EmpleadoDetailPage)
	mux.HandleFunc("GET /incentivos", h.IncentivosPage)
	mux.HandleFunc("GET /nudges", h.NudgesPage)
	mux.HandleFunc("GET /estimulos", h.EstimulosPage)

	// --- API existente (preservada) ---
	mux.HandleFunc("GET /api/empleados", h.ListEmpleadosAPI)
	mux.HandleFunc("POST /api/empleados", h.CreateEmpleadoAPI)
	mux.HandleFunc("GET /api/analisis", h.AnalisisAPI)
	mux.HandleFunc("GET /api/riesgos", h.RiesgosAPI)
	mux.HandleFunc("POST /api/recomendar/{id}", h.RecomendarAPI)
	mux.HandleFunc("GET /api/incentivos", h.ListIncentivosAPI)
	mux.HandleFunc("POST /api/incentivos", h.CreateIncentivoAPI)
	mux.HandleFunc("GET /api/nudges", h.ListNudgesAPI)
	mux.HandleFunc("POST /api/nudges", h.CreateNudgeAPI)
	mux.HandleFunc("POST /api/estimulos/{id}/aplicar", h.ApplyEstimuloAPI)

	// --- API extendida: empleados ---
	mux.HandleFunc("GET /api/empleados/{id}", h.GetEmpleadoAPI)
	mux.HandleFunc("GET /api/empleados/{id}/detail", h.EmpleadoDetailAPI)
	mux.HandleFunc("PUT /api/empleados/{id}", h.UpdateEmpleadoAPI)
	mux.HandleFunc("PUT /api/empleados/{id}/perfil", h.UpdatePerfilMAPAPI)
	mux.HandleFunc("GET /api/empleados/{id}/perfil-form", h.PerfilFormAPI)
	mux.HandleFunc("GET /api/empleados/new-form", h.EmpleadoFormAPI)
	mux.HandleFunc("POST /api/empleados/import", h.ImportCSVAPI)
	mux.HandleFunc("GET /api/empleados/template", h.CSVTemplatesAPI)

	// --- API extendida: incentivos ---
	mux.HandleFunc("GET /api/incentivos/{id}", h.GetIncentivoAPI)
	mux.HandleFunc("PUT /api/incentivos/{id}", h.UpdateIncentivoAPI)
	mux.HandleFunc("DELETE /api/incentivos/{id}", h.DeleteIncentivoAPI)
	mux.HandleFunc("GET /api/incentivos/elegibles/{empleadoId}", h.IncentivosElegiblesAPI)
	mux.HandleFunc("POST /api/incentivos/{id}/elegibilidades", h.AddElegibilidadAPI)
	mux.HandleFunc("DELETE /api/elegibilidades/{id}", h.DeleteElegibilidadAPI)

	// --- API extendida: nudges ---
	mux.HandleFunc("GET /api/nudges/{id}", h.GetNudgeAPI)
	mux.HandleFunc("PUT /api/nudges/{id}", h.UpdateNudgeAPI)

	// --- API extendida: estímulos ---
	mux.HandleFunc("GET /api/estimulos", h.ListEstimulosAPI)
	mux.HandleFunc("GET /api/estimulos/{id}", h.GetEstimuloAPI)
	mux.HandleFunc("GET /api/estimulos/{id}/apply-form", h.EstimuloApplyFormAPI)

	// --- Dashboard partials (HTMX) ---
	mux.HandleFunc("GET /api/dashboard/stats", h.DashboardStatsAPI)
	mux.HandleFunc("GET /api/dashboard/riesgos", h.DashboardRiesgosAPI)
	mux.HandleFunc("GET /api/dashboard/distribucion", h.DashboardDistribucionAPI)
	mux.HandleFunc("GET /api/dashboard/efectividad", h.DashboardEfectividadAPI)

	// --- Formularios parciales adicionales ---
	mux.HandleFunc("GET /api/empleados/import-form", h.ImportFormAPI)
	mux.HandleFunc("GET /api/incentivos/form", h.IncentivoFormAPI)
	mux.HandleFunc("GET /api/nudges/form", h.NudgeFormAPI)

	// --- Nudge toggle ---
	mux.HandleFunc("PUT /api/nudges/{id}/toggle", h.ToggleNudgeAPI)

	// --- Eliminar empleado ---
	mux.HandleFunc("DELETE /api/empleados/{id}", h.DeleteEmpleadoAPI)

	// --- Páginas de detalle adicionales ---
	mux.HandleFunc("GET /incentivos/{id}", h.IncentivoDetailPage)
	mux.HandleFunc("GET /nudges/{id}", h.NudgeDetailPage)
}

// =============================================================================
// Páginas HTML
// =============================================================================

// Dashboard renderiza la página principal con análisis y métricas.
func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	result, _ := h.Svc.Analizar(r.Context())
	empleados, _ := h.Svc.ListEmpleados(r.Context())
	incentivos, _ := h.Svc.ListIncentivos(r.Context())
	nudges, _ := h.Svc.ListNudges(r.Context())

	data := map[string]interface{}{
		"Analisis":   result,
		"Empleados":  empleados,
		"Incentivos": incentivos,
		"Nudges":     nudges,
	}

	tmpl, err := template.New("").Funcs(templateFuncs).ParseFiles("web/templates/base.html", "web/templates/dashboard.html")
	if err != nil {
		http.Error(w, "Error al cargar templates: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.ExecuteTemplate(w, "base", data)
}

// EmpleadosPage renderiza la lista de empleados.
func (h *Handler) EmpleadosPage(w http.ResponseWriter, r *http.Request) {
	empleados, _ := h.Svc.ListEmpleados(r.Context())
	data := map[string]interface{}{
		"Empleados": empleados,
	}
	tmpl, err := template.New("").Funcs(templateFuncs).ParseFiles("web/templates/base.html", "web/templates/empleados/list.html")
	if err != nil {
		http.Error(w, "Error al cargar templates: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.ExecuteTemplate(w, "base", data)
}

// EmpleadoDetailPage renderiza el detalle de un empleado.
func (h *Handler) EmpleadoDetailPage(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	detail, err := h.Svc.GetEmpleadoDetail(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	tmpl, err := template.New("").Funcs(templateFuncs).ParseFiles("web/templates/base.html", "web/templates/empleados/detail.html")
	if err != nil {
		http.Error(w, "Error al cargar templates: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.ExecuteTemplate(w, "base", detail)
}

// IncentivosPage renderiza la lista de incentivos.
func (h *Handler) IncentivosPage(w http.ResponseWriter, r *http.Request) {
	incentivos, _ := h.Svc.ListIncentivos(r.Context())
	data := map[string]interface{}{
		"Incentivos": incentivos,
	}
	tmpl, err := template.New("").Funcs(templateFuncs).ParseFiles("web/templates/base.html", "web/templates/incentivos/list.html")
	if err != nil {
		http.Error(w, "Error al cargar templates: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.ExecuteTemplate(w, "base", data)
}

// NudgesPage renderiza la lista de nudges.
func (h *Handler) NudgesPage(w http.ResponseWriter, r *http.Request) {
	nudges, _ := h.Svc.ListNudges(r.Context())
	data := map[string]interface{}{
		"Nudges": nudges,
	}
	tmpl, err := template.New("").Funcs(templateFuncs).ParseFiles("web/templates/base.html", "web/templates/nudges/list.html", "web/templates/nudges/_card.html")
	if err != nil {
		http.Error(w, "Error al cargar templates: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.ExecuteTemplate(w, "base", data)
}

// EstimulosPage renderiza la lista de estímulos con filtro por estado.
// Si es HTMX, retorna solo la tabla parcial con tabs.
func (h *Handler) EstimulosPage(w http.ResponseWriter, r *http.Request) {
	estado := r.URL.Query().Get("estado")
	if estado == "" {
		estado = "todos"
	}
	estimulos, _ := h.Svc.ListEstimulos(r.Context(), estado)
	data := map[string]interface{}{
		"Estimulos": estimulos,
		"Estado":    estado,
	}

	// HTMX: retornar solo la tabla con tabs
	if r.Header.Get("HX-Request") == "true" {
		tmpl, err := template.New("").Funcs(templateFuncs).ParseFiles("web/templates/estimulos/_table.html")
		if err != nil {
			http.Error(w, "Error al cargar template", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, data)
		return
	}

	tmpl, err := template.New("").Funcs(templateFuncs).ParseFiles("web/templates/base.html", "web/templates/estimulos/list.html")
	if err != nil {
		http.Error(w, "Error al cargar templates: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.ExecuteTemplate(w, "base", data)
}

// =============================================================================
// API: Empleados (extendida)
// =============================================================================

// GetEmpleadoAPI retorna un empleado por ID en JSON.
func (h *Handler) GetEmpleadoAPI(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}
	e, err := h.Svc.GetEmpleado(r.Context(), id)
	if err != nil || e == nil {
		http.Error(w, "Empleado no encontrado", http.StatusNotFound)
		return
	}
	writeJSON(w, e)
}

// EmpleadoDetailAPI retorna el detalle compuesto del empleado en JSON.
func (h *Handler) EmpleadoDetailAPI(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}
	detail, err := h.Svc.GetEmpleadoDetail(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, detail)
}

// UpdateEmpleadoAPI actualiza los datos de un empleado.
func (h *Handler) UpdateEmpleadoAPI(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	var input struct {
		Nombre       string `json:"nombre"`
		Email        string `json:"email"`
		Cargo        string `json:"cargo"`
		Departamento string `json:"departamento"`
		Activo       *bool  `json:"activo"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	e, err := h.Svc.GetEmpleado(r.Context(), id)
	if err != nil || e == nil {
		http.Error(w, "Empleado no encontrado", http.StatusNotFound)
		return
	}

	if input.Nombre != "" {
		e.Nombre = input.Nombre
	}
	if input.Email != "" {
		e.Email = input.Email
	}
	if input.Cargo != "" {
		e.Cargo = input.Cargo
	}
	if input.Departamento != "" {
		e.Departamento = domain.Departamento(input.Departamento)
	}
	if input.Activo != nil {
		e.Activo = *input.Activo
	}

	if err := h.Svc.UpdateEmpleado(r.Context(), e); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, e)
}

// UpdatePerfilMAPAPI actualiza el perfil MAP de un empleado.
// Si es HTMX, redirige a la página de detalle del empleado.
func (h *Handler) UpdatePerfilMAPAPI(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	empleadoID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	var input struct {
		Motivacion    float64 `json:"motivacion"`
		Habilidad     float64 `json:"habilidad"`
		Prompt        float64 `json:"prompt"`
		Sensibilidad  string  `json:"sensibilidad"`
		Confiabilidad float64 `json:"confiabilidad"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	if err := h.Svc.UpdatePerfilMAP(r.Context(), empleadoID,
		input.Motivacion, input.Habilidad, input.Prompt,
		input.Sensibilidad, input.Confiabilidad); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// HTMX: redirigir al detalle del empleado para refrescar
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("X-Toast", "✓ Perfil MAP actualizado")
		w.Header().Set("HX-Redirect", fmt.Sprintf("/empleados/%d", empleadoID))
		w.WriteHeader(http.StatusOK)
		return
	}

	// Retornar el perfil actualizado
	p, _ := h.Svc.GetEmpleadoDetail(r.Context(), empleadoID)
	writeJSON(w, p)
}

// EmpleadoFormAPI retorna un formulario HTML parcial para crear empleado.
func (h *Handler) EmpleadoFormAPI(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("").Funcs(templateFuncs).ParseFiles("web/templates/empleados/_form.html")
	if err != nil {
		http.Error(w, "Formulario no disponible", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

// PerfilFormAPI retorna un formulario HTML parcial para editar el perfil MAP.
func (h *Handler) PerfilFormAPI(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	detail, err := h.Svc.GetEmpleadoDetail(r.Context(), id)
	if err != nil || detail.Perfil == nil {
		http.Error(w, "Perfil no encontrado", http.StatusNotFound)
		return
	}

	tmpl, err := template.New("").Funcs(templateFuncs).ParseFiles("web/templates/empleados/_perfil_form.html")
	if err != nil {
		http.Error(w, "Formulario no disponible", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, detail.Perfil)
}

// ImportCSVAPI importa empleados desde un archivo CSV.
// Si es HTMX, retorna un resumen HTML del resultado.
func (h *Handler) ImportCSVAPI(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB
		http.Error(w, "Error al procesar formulario", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Archivo no encontrado en el campo 'file'", http.StatusBadRequest)
		return
	}
	defer file.Close()

	creados, errores, err := h.Svc.ImportCSV(r.Context(), file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// HTMX: retornar resumen HTML
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("X-Toast", fmt.Sprintf("✓ %d empleados importados, %d errores", creados, len(errores)))
		tmpl, err := template.New("").Funcs(templateFuncs).ParseFiles("web/templates/empleados/_import_result.html")
		if err != nil {
			http.Error(w, "Error al cargar template", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, map[string]interface{}{
			"Creados": creados,
			"Errores": errores,
		})
		return
	}

	writeJSON(w, map[string]interface{}{
		"creados": creados,
		"errores": errores,
	})
}

// CSVTemplatesAPI retorna un CSV de ejemplo para importar empleados.
func (h *Handler) CSVTemplatesAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=plantilla_empleados.csv")
	io.WriteString(w, "nombre,email,cargo,departamento\n")
}

// =============================================================================
// API: Incentivos (extendida)
// =============================================================================

// GetIncentivoAPI retorna un incentivo por ID en JSON.
func (h *Handler) GetIncentivoAPI(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}
	inc, err := h.Svc.GetIncentivo(r.Context(), id)
	if err != nil || inc == nil {
		http.Error(w, "Incentivo no encontrado", http.StatusNotFound)
		return
	}
	writeJSON(w, inc)
}

// UpdateIncentivoAPI actualiza un incentivo.
func (h *Handler) UpdateIncentivoAPI(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	inc, err := h.Svc.GetIncentivo(r.Context(), id)
	if err != nil || inc == nil {
		http.Error(w, "Incentivo no encontrado", http.StatusNotFound)
		return
	}

	var input struct {
		Nombre         string  `json:"nombre"`
		Descripcion    string  `json:"descripcion"`
		Tipo           string  `json:"tipo"`
		Intensidad     float64 `json:"intensidad"`
		Costo          float64 `json:"costo"`
		Disponibilidad string  `json:"disponibilidad"`
		Cupos          *int    `json:"cupos"`
		Activo         *bool   `json:"activo"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	if input.Nombre != "" {
		inc.Nombre = input.Nombre
	}
	if input.Descripcion != "" {
		inc.Descripcion = input.Descripcion
	}
	if input.Tipo != "" {
		inc.Tipo = domain.TipoIncentivo(input.Tipo)
	}
	inc.Intensidad = input.Intensidad
	inc.Costo = input.Costo
	if input.Disponibilidad != "" {
		inc.Disponibilidad = domain.Disponibilidad(input.Disponibilidad)
	}
	if input.Cupos != nil {
		inc.Cupos = *input.Cupos
	}
	if input.Activo != nil {
		inc.Activo = *input.Activo
	}

	if err := h.Svc.UpdateIncentivo(r.Context(), inc); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, inc)
}

// DeleteIncentivoAPI elimina un incentivo.
func (h *Handler) DeleteIncentivoAPI(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}
	if err := h.Svc.DeleteIncentivo(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"status": "ok"})
}

// IncentivosElegiblesAPI retorna los incentivos para los que un empleado es elegible.
func (h *Handler) IncentivosElegiblesAPI(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("empleadoId")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}
	elegibles, err := h.Svc.GetIncentivosElegibles(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, elegibles)
}

// AddElegibilidadAPI agrega un criterio de elegibilidad a un incentivo.
func (h *Handler) AddElegibilidadAPI(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	incentivoID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	var input struct {
		Campo    string `json:"campo"`
		Operador string `json:"operador"`
		Valor    string `json:"valor"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	e, err := h.Svc.AddElegibilidad(r.Context(), incentivoID, input.Campo, input.Operador, input.Valor)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, e)
}

// DeleteElegibilidadAPI elimina un criterio de elegibilidad.
func (h *Handler) DeleteElegibilidadAPI(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}
	if err := h.Svc.DeleteElegibilidad(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"status": "ok"})
}

// =============================================================================
// API: Nudges (extendida)
// =============================================================================

// GetNudgeAPI retorna un nudge por ID en JSON.
func (h *Handler) GetNudgeAPI(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}
	n, err := h.Svc.GetNudge(r.Context(), id)
	if err != nil || n == nil {
		http.Error(w, "Nudge no encontrado", http.StatusNotFound)
		return
	}
	writeJSON(w, n)
}

// UpdateNudgeAPI actualiza un nudge.
func (h *Handler) UpdateNudgeAPI(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	n, err := h.Svc.GetNudge(r.Context(), id)
	if err != nil || n == nil {
		http.Error(w, "Nudge no encontrado", http.StatusNotFound)
		return
	}

	var input struct {
		Nombre      string `json:"nombre"`
		Descripcion string `json:"descripcion"`
		Tipo        string `json:"tipo"`
		Ambito      string `json:"ambito"`
		TargetID    *int64 `json:"target_id"`
		Activo      *bool  `json:"activo"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	if input.Nombre != "" {
		n.Nombre = input.Nombre
	}
	if input.Descripcion != "" {
		n.Descripcion = input.Descripcion
	}
	if input.Tipo != "" {
		n.Tipo = domain.TipoNudge(input.Tipo)
	}
	if input.Ambito != "" {
		n.Ambito = domain.AmbitoNudge(input.Ambito)
	}
	if input.TargetID != nil {
		n.TargetID = *input.TargetID
	}
	if input.Activo != nil {
		n.Activo = *input.Activo
	}

	if err := h.Svc.UpdateNudge(r.Context(), n); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, n)
}

// =============================================================================
// API: Estímulos (extendida)
// =============================================================================

// ListEstimulosAPI lista estímulos con filtro por estado.
func (h *Handler) ListEstimulosAPI(w http.ResponseWriter, r *http.Request) {
	estado := r.URL.Query().Get("estado")
	if estado == "" {
		estado = "todos"
	}
	estimulos, err := h.Svc.ListEstimulos(r.Context(), estado)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, estimulos)
}

// GetEstimuloAPI retorna un estímulo por ID en JSON.
func (h *Handler) GetEstimuloAPI(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}
	e, err := h.Svc.GetEstimulo(r.Context(), id)
	if err != nil || e == nil {
		http.Error(w, "Estímulo no encontrado", http.StatusNotFound)
		return
	}
	writeJSON(w, e)
}

// EstimuloApplyFormAPI retorna un formulario HTML parcial para aplicar un estímulo.
func (h *Handler) EstimuloApplyFormAPI(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	estimulo, err := h.Svc.GetEstimulo(r.Context(), id)
	if err != nil || estimulo == nil {
		http.Error(w, "Estímulo no encontrado", http.StatusNotFound)
		return
	}

	tmpl, err := template.New("").Funcs(templateFuncs).ParseFiles("web/templates/estimulos/_apply_form.html")
	if err != nil {
		http.Error(w, "Formulario no disponible", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, estimulo)
}

// =============================================================================
// API: Existente (preservada)
// =============================================================================

// ListEmpleadosAPI retorna todos los empleados en JSON.
func (h *Handler) ListEmpleadosAPI(w http.ResponseWriter, r *http.Request) {
	empleados, err := h.Svc.ListEmpleados(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, empleados)
}

// CreateEmpleadoAPI crea un nuevo empleado con perfil MAP y umbral.
// Si es HTMX, retorna una fila HTML para insertar en la tabla.
func (h *Handler) CreateEmpleadoAPI(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Nombre       string `json:"nombre"`
		Email        string `json:"email"`
		Cargo        string `json:"cargo"`
		Departamento string `json:"departamento"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	e, err := h.Svc.CreateEmpleado(r.Context(), input.Nombre, input.Email, input.Cargo, input.Departamento)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// HTMX: retornar fila HTML para insertar en la tabla
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("X-Toast", "✓ Empleado creado")
		tmpl, err := template.New("").Funcs(templateFuncs).ParseFiles("web/templates/empleados/_row.html")
		if err != nil {
			http.Error(w, "Error al cargar template", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, e)
		return
	}

	writeJSON(w, e)
}

// AnalisisAPI retorna el análisis descriptivo del capital humano.
func (h *Handler) AnalisisAPI(w http.ResponseWriter, r *http.Request) {
	result, err := h.Svc.Analizar(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, result)
}

// RiesgosAPI retorna la zona de riesgo de empleados.
func (h *Handler) RiesgosAPI(w http.ResponseWriter, r *http.Request) {
	result, err := h.Svc.Analizar(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, result.ZonaRiesgo)
}

// RecomendarAPI genera recomendaciones MAP para un empleado.
// Si la petición viene de HTMX (HX-Request: true), retorna HTML.
// Si no, retorna JSON (API programática).
func (h *Handler) RecomendarAPI(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}
	result, err := h.Svc.Recomendar(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// HTMX: retornar HTML parcial
	if r.Header.Get("HX-Request") == "true" {
		tmpl, err := template.New("").Funcs(templateFuncs).ParseFiles("web/templates/_recomendacion.html")
		if err != nil {
			http.Error(w, "Error al cargar template de recomendación", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, result)
		return
	}

	writeJSON(w, result)
}

// ListIncentivosAPI retorna todos los incentivos activos en JSON.
func (h *Handler) ListIncentivosAPI(w http.ResponseWriter, r *http.Request) {
	incentivos, err := h.Svc.ListIncentivos(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, incentivos)
}

// CreateIncentivoAPI crea un nuevo incentivo.
// Si es HTMX, redirige a la página de incentivos.
func (h *Handler) CreateIncentivoAPI(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Nombre         string  `json:"nombre"`
		Descripcion    string  `json:"descripcion"`
		Tipo           string  `json:"tipo"`
		Intensidad     float64 `json:"intensidad"`
		Costo          float64 `json:"costo"`
		Disponibilidad string  `json:"disponibilidad"`
		Cupos          int     `json:"cupos"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	inc, err := h.Svc.CreateIncentivo(r.Context(),
		input.Nombre, input.Descripcion,
		domain.TipoIncentivo(input.Tipo),
		input.Intensidad, input.Costo,
		domain.Disponibilidad(input.Disponibilidad), input.Cupos)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// HTMX: redirigir a la lista de incentivos
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/incentivos")
		w.WriteHeader(http.StatusOK)
		return
	}

	writeJSON(w, inc)
}

// ListNudgesAPI retorna todos los nudges activos en JSON.
func (h *Handler) ListNudgesAPI(w http.ResponseWriter, r *http.Request) {
	nudges, err := h.Svc.ListNudges(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, nudges)
}

// CreateNudgeAPI crea un nuevo nudge.
// Si es HTMX, redirige a la página de nudges.
func (h *Handler) CreateNudgeAPI(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Nombre      string `json:"nombre"`
		Descripcion string `json:"descripcion"`
		Tipo        string `json:"tipo"`
		Ambito      string `json:"ambito"`
		TargetID    int64  `json:"target_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	nudge, err := h.Svc.CreateNudge(r.Context(),
		input.Nombre, input.Descripcion,
		domain.TipoNudge(input.Tipo), domain.AmbitoNudge(input.Ambito), input.TargetID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// HTMX: redirigir a la lista de nudges
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/nudges")
		w.WriteHeader(http.StatusOK)
		return
	}

	writeJSON(w, nudge)
}

// ApplyEstimuloAPI aplica un estímulo y recalibra el umbral del empleado.
func (h *Handler) ApplyEstimuloAPI(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	var input struct {
		Respuesta float64 `json:"respuesta"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	if err := h.Svc.ApplyEstimulo(r.Context(), id, input.Respuesta); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("X-Toast", "Estímulo aplicado y umbral recalibrado")
	writeJSON(w, map[string]string{"status": "ok"})
}

// =============================================================================
// Dashboard partials (HTMX)
// =============================================================================

// DashboardStatsAPI retorna el partial HTML de KPIs del dashboard.
func (h *Handler) DashboardStatsAPI(w http.ResponseWriter, r *http.Request) {
	result, _ := h.Svc.Analizar(r.Context())
	incentivos, _ := h.Svc.ListIncentivos(r.Context())
	nudges, _ := h.Svc.ListNudges(r.Context())
	data := map[string]interface{}{
		"TotalEmpleados":    result.TotalEmpleados,
		"EmpleadosEnRiesgo": result.EmpleadosEnRiesgo,
		"IncentivosActivos": len(incentivos),
		"NudgesActivos":     len(nudges),
	}
	tmpl, err := template.New("").Funcs(templateFuncs).ParseFiles("web/templates/_stats.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, data)
}

// DashboardRiesgosAPI retorna el partial HTML de zona de riesgo.
func (h *Handler) DashboardRiesgosAPI(w http.ResponseWriter, r *http.Request) {
	result, _ := h.Svc.Analizar(r.Context())
	tmpl, err := template.New("").Funcs(templateFuncs).ParseFiles("web/templates/_riesgos.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, result)
}

// DashboardDistribucionAPI retorna el partial HTML de distribución MAP.
func (h *Handler) DashboardDistribucionAPI(w http.ResponseWriter, r *http.Request) {
	result, _ := h.Svc.Analizar(r.Context())
	tmpl, err := template.New("").Funcs(templateFuncs).ParseFiles("web/templates/_distribucion.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, result)
}

// DashboardEfectividadAPI retorna el partial HTML de efectividad.
func (h *Handler) DashboardEfectividadAPI(w http.ResponseWriter, r *http.Request) {
	result, _ := h.Svc.Analizar(r.Context())
	tmpl, err := template.New("").Funcs(templateFuncs).ParseFiles("web/templates/_efectividad.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, result)
}

// ImportFormAPI retorna el partial HTML del formulario de importación CSV.
func (h *Handler) ImportFormAPI(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("").Funcs(templateFuncs).ParseFiles("web/templates/empleados/_import_form.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

// IncentivoFormAPI retorna el partial HTML del formulario de creación de incentivo.
func (h *Handler) IncentivoFormAPI(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("").Funcs(templateFuncs).ParseFiles("web/templates/incentivos/_form.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

// NudgeFormAPI retorna el partial HTML del formulario de creación de nudge.
func (h *Handler) NudgeFormAPI(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("").Funcs(templateFuncs).ParseFiles("web/templates/nudges/_form.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

// ToggleNudgeAPI invierte el estado activo/inactivo de un nudge.
func (h *Handler) ToggleNudgeAPI(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}
	nudge, err := h.Svc.ToggleNudge(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Devolver la card actualizada
	tmpl, err := template.New("").Funcs(templateFuncs).ParseFiles("web/templates/nudges/_card.html")
	if err != nil {
		writeJSON(w, nudge)
		return
	}
	tmpl.ExecuteTemplate(w, "nudge-card", nudge)
}

// DeleteEmpleadoAPI elimina un empleado.
func (h *Handler) DeleteEmpleadoAPI(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}
	if err := h.Svc.DeleteEmpleado(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("X-Toast", "Empleado eliminado")
	w.WriteHeader(http.StatusOK)
}

// IncentivoDetailPage renderiza la página de detalle de un incentivo.
func (h *Handler) IncentivoDetailPage(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	detail, err := h.Svc.GetIncentivoDetail(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	tmpl, err := template.New("").Funcs(templateFuncs).ParseFiles("web/templates/base.html", "web/templates/incentivos/detail.html")
	if err != nil {
		http.Error(w, "Error al cargar templates: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.ExecuteTemplate(w, "base", detail)
}

// NudgeDetailPage renderiza la página de detalle de un nudge.
func (h *Handler) NudgeDetailPage(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	nudge, err := h.Svc.GetNudge(r.Context(), id)
	if err != nil || nudge == nil {
		http.Error(w, "Nudge no encontrado", http.StatusNotFound)
		return
	}

	data := map[string]interface{}{"Nudge": nudge}
	tmpl, err := template.New("").Funcs(templateFuncs).ParseFiles("web/templates/base.html", "web/templates/nudges/detail.html")
	if err != nil {
		http.Error(w, "Error al cargar templates: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.ExecuteTemplate(w, "base", data)
}

// =============================================================================
// Utilidades
// =============================================================================

// writeJSON escribe una respuesta JSON con el header adecuado.
func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, "Error al serializar respuesta", http.StatusInternalServerError)
	}
}

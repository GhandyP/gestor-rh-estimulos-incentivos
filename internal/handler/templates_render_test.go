package handler

import (
	"bytes"
	"html/template"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"estimulos-incentivos/internal/domain"
)

// fixtureEmpleados produce datos deterministas para el render de list.html:
// IDs y nombres fijos (sin time.Now() ni random) para que la salida sea
// reproducible byte a byte.
func fixtureEmpleados() []domain.Empleado {
	return []domain.Empleado{
		{ID: 1, Nombre: "María García", Cargo: "Senior Developer", Departamento: domain.Departamento("Ingeniería"), Email: "maria@empresa.com"},
		{ID: 2, Nombre: "Juan Pérez", Cargo: "Junior Developer", Departamento: domain.Departamento("Ingeniería"), Email: "juan@empresa.com"},
		{ID: 3, Nombre: "Ana López", Cargo: "Tech Lead", Departamento: domain.Departamento("Producto"), Email: "ana@empresa.com"},
	}
}

// renderListFixture parsea base.html + empleados/list.html con el mismo set de
// funcs que usa el handler y ejecuta el template "base". Resuelve los paths
// desde la raíz del repo (runtime.Caller), independiente del cwd.
func renderListFixture(t *testing.T) string {
	t.Helper()
	root := repoRoot(t)
	tmpl, err := template.New("").Funcs(templateFuncs).ParseFiles(
		filepath.Join(root, "web", "templates", "base.html"),
		filepath.Join(root, "web", "templates", "empleados", "list.html"),
	)
	if err != nil {
		t.Fatalf("parse base+list: %v", err)
	}
	data := map[string]interface{}{
		"Empleados": fixtureEmpleados(),
		"CSRFToken": "test-token-0123456789abcdef",
	}
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "base", data); err != nil {
		t.Fatalf("render list: %v", err)
	}
	return buf.String()
}

// Spec: "Template rendering is stable" — el template parsea, el render
// incluye los datos del fixture y renderizar dos veces produce bytes
// idénticos (reproducible).
func TestEmpleadosListParseAndDeterministicRender(t *testing.T) {
	out1 := renderListFixture(t)
	out2 := renderListFixture(t)

	if out1 != out2 {
		t.Fatal("dos renders del mismo fixture difieren — salida no reproducible")
	}
	if !strings.Contains(out1, `name="csrf-token" content="test-token-0123456789abcdef"`) {
		t.Error("meta csrf-token no renderiza el token del fixture")
	}
	for _, e := range fixtureEmpleados() {
		if !strings.Contains(out1, e.Nombre) {
			t.Errorf("render no contiene el nombre del empleado %q", e.Nombre)
		}
	}
}

// Spec: "Employee list has one action" — cada fila expone EXACTAMENTE una
// acción de borrado válida (hx-delete hacia /api/empleados/{id}).
// RED: el list.html actual duplica el botón de borrado.
func TestEmpleadosListExactlyOneDeleteActionPerRow(t *testing.T) {
	out := renderListFixture(t)

	for _, e := range fixtureEmpleados() {
		marker := `hx-delete="/api/empleados/` + strconv.FormatInt(e.ID, 10) + `"`
		if got := strings.Count(out, marker); got != 1 {
			t.Errorf("empleado %d: %d acción(es) de borrado, want 1", e.ID, got)
		}
	}
}

// Spec: "no malformed cell structure" — cada fila de datos tiene 6 celdas
// abiertas y 6 cerradas (mismo conteo que el encabezado <th>).
// RED: el list.html actual tiene un </td> huérfano tras el botón duplicado.
func TestEmpleadosListBalancedCellsPerRow(t *testing.T) {
	out := renderListFixture(t)

	rows := strings.Split(out, "<tr")
	dataRows := 0
	for _, frag := range rows {
		if !strings.Contains(frag, "hx-delete=") {
			continue // encabezado, filas de recomendación ocultas, etc.
		}
		dataRows++
		opens := strings.Count(frag, "<td")
		closes := strings.Count(frag, "</td")
		if opens != 6 || closes != 6 {
			t.Errorf("fila de datos: %d <td abiertos, %d </td cerrados; want 6/6", opens, closes)
		}
	}
	if dataRows != len(fixtureEmpleados()) {
		t.Errorf("filas de datos detectadas = %d, want %d", dataRows, len(fixtureEmpleados()))
	}
}

// Spec: "Template rendering is stable" — las páginas clave parsean y se
// renderizan de forma reproducible (mismo fixture → mismos bytes).
func TestKeyPagesParseAndDeterministicRender(t *testing.T) {
	root := repoRoot(t)

	pages := []struct {
		name    string
		files   []string
		execute string // nombre de template a ejecutar; "" = por basename
		data    map[string]interface{}
	}{
		{
			name:    "dashboard",
			files:   []string{"base.html", "dashboard.html"},
			execute: "base",
			data:    map[string]interface{}{"CSRFToken": "tok-dashboard"},
		},
		{
			name:    "incentivos",
			files:   []string{"base.html", "incentivos/list.html"},
			execute: "base",
			data: map[string]interface{}{
				"Incentivos": []domain.Incentivo{},
				"CSRFToken":  "tok-incentivos",
			},
		},
		{
			name:    "estimulos",
			files:   []string{"base.html", "estimulos/list.html"},
			execute: "base",
			data: map[string]interface{}{
				"Estimulos": []domain.Estimulo{},
				"Estado":    "todos",
				"CSRFToken": "tok-estimulos",
			},
		},
		{
			name:  "login",
			files: []string{"login.html"},
			data:  map[string]interface{}{"Error": ""},
		},
	}

	for _, pg := range pages {
		t.Run(pg.name, func(t *testing.T) {
			paths := make([]string, 0, len(pg.files))
			for _, f := range pg.files {
				paths = append(paths, filepath.Join(root, "web", "templates", f))
			}
			tmpl, err := template.New("").Funcs(templateFuncs).ParseFiles(paths...)
			if err != nil {
				t.Fatalf("parse %v: %v", pg.files, err)
			}
			execute := pg.execute
			if execute == "" {
				execute = filepath.Base(pg.files[len(pg.files)-1])
			}
			render := func() string {
				var buf bytes.Buffer
				if err := tmpl.ExecuteTemplate(&buf, execute, pg.data); err != nil {
					t.Fatalf("render %v: %v", pg.files, err)
				}
				return buf.String()
			}
			first, second := render(), render()
			if first != second {
				t.Fatal("dos renders del mismo fixture difieren — salida no reproducible")
			}
			if strings.TrimSpace(first) == "" {
				t.Fatal("render vacío")
			}
		})
	}
}

func fixtureNudges() []domain.Nudge {
	return []domain.Nudge{{
		ID:          7,
		Nombre:      "Ahorro energético",
		Descripcion: "Mostrar el consumo actual antes de elegir",
		Tipo:        domain.NudgeDefaults,
		Ambito:      domain.AmbitoGlobal,
		Activo:      true,
	}}
}

func TestNudgesListRendersCard(t *testing.T) {
	ts := mustParseTemplates(t)
	nudge := fixtureNudges()[0]
	var buf bytes.Buffer
	data := map[string]interface{}{
		"Nudges":    fixtureNudges(),
		"CSRFToken": "tok-nudges",
	}
	if err := ts.execute(&buf, "nudges-list", "base", data); err != nil {
		t.Fatalf("render nudges list: %v", err)
	}

	out := buf.String()
	for _, want := range []string{
		nudge.Nombre,
		nudge.Descripcion,
		`href="/nudges/` + strconv.FormatInt(nudge.ID, 10) + `"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("nudges list does not render %q: %q", want, out)
		}
	}
}

func TestNudgeDetailRendersNudgeData(t *testing.T) {
	ts := mustParseTemplates(t)
	nudge := fixtureNudges()[0]
	var buf bytes.Buffer
	data := map[string]interface{}{
		"Nudge":     nudge,
		"CSRFToken": "tok-nudges",
	}
	if err := ts.execute(&buf, "nudges-detail", "base", data); err != nil {
		t.Fatalf("render nudge detail: %v", err)
	}

	out := buf.String()
	for _, want := range []string{nudge.Nombre, nudge.Descripcion, string(nudge.Tipo), string(nudge.Ambito)} {
		if !strings.Contains(out, want) {
			t.Errorf("nudge detail does not render %q: %q", want, out)
		}
	}
}

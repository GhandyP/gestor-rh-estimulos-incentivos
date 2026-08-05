package handler

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Spec (operable-delivery): "Templates MUST load deterministically independent
// of the caller working directory" y "startup validates template availability".
// ParseTemplates carga UNA VEZ todos los sets de páginas/partials desde una
// raíz explícita; cualquier archivo faltante o error de parseo es un fallo de
// arranque. Los métodos de este archivo referencian el tipo TemplateSet y la
// función ParseTemplates que todavía no existen (RED: error de compilación).

// mustParseTemplates es el helper compartido: parsea desde la raíz real del repo.
func mustParseTemplates(t *testing.T) *TemplateSet {
	t.Helper()
	ts, err := ParseTemplates(repoRoot(t))
	if err != nil {
		t.Fatalf("ParseTemplates(%s): %v", repoRoot(t), err)
	}
	return ts
}

// Cada página registrada en pageDefs debe quedar disponible en el set parseado.
func TestParseTemplatesLoadsEveryRegisteredPage(t *testing.T) {
	ts := mustParseTemplates(t)
	for name := range pageDefs {
		if _, ok := ts.pages[name]; !ok {
			t.Errorf("page %q no fue parseada en el set", name)
		}
	}
	if len(ts.pages) != len(pageDefs) {
		t.Errorf("pages parseadas = %d, pageDefs = %d", len(ts.pages), len(pageDefs))
	}
}

// Spec: "Missing startup asset fails safely" — si un template requerido falta,
// la carga falla con error (el arranque debe detenerse antes de servir tráfico).
func TestParseTemplatesFailsWhenAssetMissing(t *testing.T) {
	realRoot := repoRoot(t)
	src := filepath.Join(realRoot, "web", "templates")

	// Réplica del árbol de templates en un directorio temporal, omitiendo
	// deliberadamente estimulos/_table.html (asset requerido por pageDefs).
	fakeRoot := t.TempDir()
	dst := filepath.Join(fakeRoot, "web", "templates")
	err := filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		if rel == filepath.Join("estimulos", "_table.html") {
			return nil
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		out := filepath.Join(dst, rel)
		if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
			return err
		}
		return os.WriteFile(out, b, 0o644)
	})
	if err != nil {
		t.Fatalf("replica del árbol de templates: %v", err)
	}

	if _, err := ParseTemplates(fakeRoot); err == nil {
		t.Fatal("la carga debe fallar cuando un template requerido falta")
	}
}

// Una raíz sin el directorio web/templates tampoco puede cargar templates.
func TestParseTemplatesRejectsRootWithoutTemplatesDir(t *testing.T) {
	if _, err := ParseTemplates(t.TempDir()); err == nil {
		t.Fatal("raíz sin web/templates debe fallar la carga")
	}
}

// Spec: "Template rendering is stable" — renderizar dos veces produce bytes
// idénticos (reproducible) y la página incluye los datos del fixture.
func TestTemplateSetExecutePageIsDeterministic(t *testing.T) {
	ts := mustParseTemplates(t)
	render := func() string {
		var buf bytes.Buffer
		if err := ts.execute(&buf, "dashboard", "base", map[string]interface{}{"CSRFToken": "tok-det"}); err != nil {
			t.Fatalf("render dashboard: %v", err)
		}
		return buf.String()
	}
	first, second := render(), render()
	if first != second {
		t.Fatal("dos renders del mismo fixture difieren — salida no reproducible")
	}
	if !strings.Contains(first, "tok-det") {
		t.Error("el render no incluye el token CSRF del fixture")
	}
	if strings.TrimSpace(first) == "" {
		t.Fatal("render vacío")
	}
}

// Los partials (stats, forms, rows) se ejecutan por su raíz: con el patrón
// previo (template.New("").ParseFiles + Execute) el root vacío hacía fallar
// todos los partials HTMX. El TemplateSet nombra la raíz como el archivo
// (semántica de ParseFiles de nivel superior) para que Execute renderice.
func TestTemplateSetExecutePartialRendersRoot(t *testing.T) {
	ts := mustParseTemplates(t)
	var buf bytes.Buffer
	if err := ts.execute(&buf, "stats", "", map[string]interface{}{"TotalEmpleados": 42, "EmpleadosEnRiesgo": 3}); err != nil {
		t.Fatalf("render stats: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, ">42<") {
		t.Errorf("stats no renderiza el valor del fixture: %q", out)
	}
	if strings.TrimSpace(out) == "" {
		t.Fatal("partial stats renderizado vacío")
	}
}

// Una página desconocida debe producir error (no silencio).
func TestTemplateSetExecuteUnknownPageErrors(t *testing.T) {
	ts := mustParseTemplates(t)
	var buf bytes.Buffer
	if err := ts.execute(&buf, "no-such-page", "", nil); err == nil {
		t.Fatal("página desconocida debe dar error")
	}
}

// Cargar y renderizar desde una raíz absoluta funciona con cualquier cwd.
func TestTemplateSetIndependentOfCwd(t *testing.T) {
	t.Chdir(t.TempDir()) // cwd arbitraria, ajena al repo
	ts := mustParseTemplates(t)
	var buf bytes.Buffer
	if err := ts.execute(&buf, "login", "", map[string]interface{}{"Error": ""}); err != nil {
		t.Fatalf("render login con cwd arbitraria: %v", err)
	}
	if strings.TrimSpace(buf.String()) == "" {
		t.Fatal("render vacío con cwd arbitraria")
	}
}

// Sin SetTemplates el render falla; con el set inyectado, renderiza.
func TestHandlerSetTemplatesEnablesRender(t *testing.T) {
	h := &Handler{}

	var buf bytes.Buffer
	if err := h.execute(&buf, "stats", "", map[string]interface{}{"TotalEmpleados": 1}); err == nil {
		t.Fatal("render sin templates inyectados debe fallar")
	}

	h.SetTemplates(mustParseTemplates(t))
	buf.Reset()
	if err := h.execute(&buf, "stats", "", map[string]interface{}{"TotalEmpleados": 1, "EmpleadosEnRiesgo": 0}); err != nil {
		t.Fatalf("render con templates inyectados: %v", err)
	}
	if !strings.Contains(buf.String(), ">1<") {
		t.Errorf("render no incluye el dato: %q", buf.String())
	}
}

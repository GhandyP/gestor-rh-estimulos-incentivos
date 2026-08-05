package handler

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// repoRoot resuelve la raíz del proyecto desde la ubicación de este archivo de
// test, independiente del directorio de trabajo (los templates se cargan hoy
// con rutas relativas al cwd; el slice 4 hace el arranque cwd-independiente).
func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("no se pudo localizar el archivo de test")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func readTemplate(t *testing.T, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(repoRoot(t), "web", "templates", rel))
	if err != nil {
		t.Fatalf("leer web/templates/%s: %v", rel, err)
	}
	return string(b)
}

// base.html debe exponer el token CSRF ligado a la sesión (meta tag) e inyectar
// el header X-CSRF-Token en TODA request HTMX: esto cubre también los botones y
// formularios de mutación sin hidden input (hx-delete, hx-put, json-enc).
func TestBaseTemplateExposesCSRFToken(t *testing.T) {
	base := readTemplate(t, "base.html")

	for _, want := range []string{
		`name="csrf-token"`,
		`{{.CSRFToken}}`,
		"X-CSRF-Token",
		"htmx:configRequest",
		`action="/logout"`,
	} {
		if !strings.Contains(base, want) {
			t.Errorf("base.html no contiene %q", want)
		}
	}
}

// login.html debe existir y postear las credenciales del operador al login.
func TestLoginTemplatePostsCredentials(t *testing.T) {
	login := readTemplate(t, "login.html")

	for _, want := range []string{
		`action="/login"`,
		`name="username"`,
		`name="password"`,
		`method="post"`,
	} {
		if !strings.Contains(login, want) {
			t.Errorf("login.html no contiene %q", want)
		}
	}
}

// Todo formulario de mutación dedicado debe incluir el token CSRF como hidden
// input (la inyección por header de base.html es el respaldo universal).
func TestMutationFormsCarryCSRFToken(t *testing.T) {
	forms := []string{
		"empleados/_form.html",
		"empleados/_import_form.html",
		"empleados/_perfil_form.html",
		"incentivos/_form.html",
		"nudges/_form.html",
		"estimulos/_apply_form.html",
	}
	for _, rel := range forms {
		content := readTemplate(t, rel)
		if !strings.Contains(content, `name="_csrf"`) {
			t.Errorf("%s no incluye hidden _csrf", rel)
		}
		if !strings.Contains(content, `{{.CSRFToken}}`) {
			t.Errorf("%s no renderiza el token CSRF", rel)
		}
	}
}

package main

import (
	"os"
	"path/filepath"
	"testing"
)

// Spec (operable-delivery): "Templates MUST load deterministically independent
// of the caller working directory" — resolveAppRoot localiza la raíz del repo
// (go.mod) sin depender del cwd; APP_ROOT permite sobreescribirla (contenedores).

func TestResolveAppRootFindsRepoRoot(t *testing.T) {
	root, err := resolveAppRoot()
	if err != nil {
		t.Fatalf("resolveAppRoot: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("raíz %q no contiene go.mod: %v", root, err)
	}
	if _, err := os.Stat(filepath.Join(root, "web", "templates")); err != nil {
		t.Fatalf("raíz %q no contiene web/templates: %v", root, err)
	}
}

// Cambiar el directorio de trabajo a un lugar arbitrario (incluso inexistente
// dentro del repo) no debe afectar la resolución de la raíz.
func TestResolveAppRootIndependentOfCwd(t *testing.T) {
	t.Setenv("APP_ROOT", "")
	t.Chdir(t.TempDir())

	root, err := resolveAppRoot()
	if err != nil {
		t.Fatalf("resolveAppRoot con cwd arbitraria: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("raíz %q no contiene go.mod: %v", root, err)
	}
}

func TestResolveAppRootUsesEnvOverride(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("APP_ROOT", tmp)

	root, err := resolveAppRoot()
	if err != nil {
		t.Fatalf("resolveAppRoot: %v", err)
	}
	if root != tmp {
		t.Errorf("root = %q, want %q (APP_ROOT)", root, tmp)
	}
}

// Un APP_ROOT relativo se convierte en absoluto (determinismo del root).
func TestResolveAppRootEnvOverrideNormalizesRelative(t *testing.T) {
	t.Setenv("APP_ROOT", "relative-dir")
	root, err := resolveAppRoot()
	if err != nil {
		t.Fatalf("resolveAppRoot: %v", err)
	}
	if !filepath.IsAbs(root) {
		t.Errorf("root = %q, want ruta absoluta", root)
	}
}

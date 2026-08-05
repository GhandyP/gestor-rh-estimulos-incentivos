package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Spec (operable-delivery): "Local/container delivery and CI MUST reproduce
// go test/build/vet/gofmt; Docker and Compose MAY provide the single-instance
// demonstration path." — checks estructurales sobre los artefactos de entrega
// (sin browser E2E). RED: los archivos no existen todavía.

// readRepoFile lee un archivo del repo resolviendo la raíz sin depender del cwd.
func readRepoFile(t *testing.T, rel string) string {
	t.Helper()
	root, err := resolveAppRoot()
	if err != nil {
		t.Fatalf("root: %v", err)
	}
	b, err := os.ReadFile(filepath.Join(root, rel))
	if err != nil {
		t.Fatalf("leer %s: %v", rel, err)
	}
	return string(b)
}

// Dockerfile: imagen de build y runtime fijadas (reproducible), binario
// estático CGO-free, templates dentro de la imagen, APP_ROOT/DB_PATH de
// arranque seguro y HEALTHCHECK.
func TestDockerfileReproducibleAndOperable(t *testing.T) {
	df := readRepoFile(t, "Dockerfile")

	for _, want := range []string{
		"FROM golang:1.26.2-alpine", // build fijado (go.mod: 1.26.2)
		"CGO_ENABLED=0",             // sin CGO, sin libc en runtime
		"FROM alpine:3.21",          // runtime fijado
		"COPY web/",                 // templates dentro de la imagen
		"APP_ROOT=/app",
		"HEALTHCHECK",
		"EXPOSE 8080",
		"DB_PATH=/data/estimulos.db",
	} {
		if !strings.Contains(df, want) {
			t.Errorf("Dockerfile no contiene %q", want)
		}
	}
}

// compose.yaml: UNA instancia única, volumen de SQLite persistente y
// healthcheck; sin orquestación innecesaria (sin deploy/replicas).
func TestComposeSingleInstanceWithVolumeAndHealthcheck(t *testing.T) {
	c := readRepoFile(t, "compose.yaml")

	for _, want := range []string{
		"services:",
		"8080:8080",
		"estimulos-data:/data",
		"healthcheck:",
		"DB_PATH: /data/estimulos.db",
		"APP_ROOT: /app",
		"volumes:",
	} {
		if !strings.Contains(c, want) {
			t.Errorf("compose.yaml no contiene %q", want)
		}
	}

	// Exactamente un servicio y una definición de volumen: single instance.
	if got := strings.Count(c, "container_name:"); got != 1 {
		t.Errorf("container_name = %d, want 1 (instancia única)", got)
	}
	if got := strings.Count(c, "image:"); got != 1 {
		t.Errorf("image = %d, want 1 (instancia única)", got)
	}
	if strings.Contains(c, "deploy:") || strings.Contains(c, "replicas:") {
		t.Error("compose.yaml no debe incluir orquestación (deploy/replicas)")
	}
}

// .dockerignore mantiene el build context reproducible (sin .git, openspec,
// bases locales ni binarios).
func TestDockerignoreExcludesRepositoryNoise(t *testing.T) {
	d := readRepoFile(t, ".dockerignore")
	for _, want := range []string{".git", "openspec", "*.db", "/server"} {
		if !strings.Contains(d, want) {
			t.Errorf(".dockerignore no excluye %q", want)
		}
	}
}

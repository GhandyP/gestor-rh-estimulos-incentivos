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
		`"${HOST_PORT:-8080}:8080"`,
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

func TestComposeDemoConfigurationAndSmokeScript(t *testing.T) {
	compose := readRepoFile(t, "compose.yaml")
	for _, want := range []string{
		`"${HOST_PORT:-8080}:8080"`,
		`COOKIE_SECURE: "${COOKIE_SECURE:-false}"`,
		`PORT: "8080"`,
	} {
		if !strings.Contains(compose, want) {
			t.Errorf("compose.yaml no contiene %q", want)
		}
	}

	envExample := readRepoFile(t, ".env.example")
	for _, want := range []string{
		"COOKIE_SECURE=false",
		"OPERATOR_USER=admin",
		"OPERATOR_PASSWORD=dev-admin-password-change-me",
		"SESSION_SECRET=dev-only-session-secret-change-me-0123456789abcdef",
		"HOST_PORT=8080",
	} {
		if !strings.Contains(envExample, want) {
			t.Errorf(".env.example no contiene %q", want)
		}
	}
	if !strings.Contains(strings.ToLower(envExample), "development only") {
		t.Error(".env.example debe marcar las credenciales como development only")
	}

	gitignore := readRepoFile(t, ".gitignore")
	for _, want := range []string{".atl/", ".codegraph/", ".env", "!.env.example", "*.db", "*.db-wal", "*.db-shm", "/server"} {
		if !strings.Contains(gitignore, want) {
			t.Errorf(".gitignore no contiene %q", want)
		}
	}

	dockerignore := readRepoFile(t, ".dockerignore")
	for _, want := range []string{".env", ".env.*", ".codegraph/", "*.db-wal", "*.db-shm"} {
		if !strings.Contains(dockerignore, want) {
			t.Errorf(".dockerignore no excluye %q", want)
		}
	}

	smoke := readRepoFile(t, "scripts/compose-smoke.sh")
	for _, want := range []string{
		"#!/usr/bin/env bash",
		"docker compose",
		"up -d --build",
		"trap",
		"HOST_PORT",
		"/readyz",
		"/healthz",
		"/login",
		"--volumes",
	} {
		if !strings.Contains(smoke, want) {
			t.Errorf("scripts/compose-smoke.sh no contiene %q", want)
		}
	}
	root, err := resolveAppRoot()
	if err != nil {
		t.Fatalf("root: %v", err)
	}
	info, err := os.Stat(filepath.Join(root, "scripts/compose-smoke.sh"))
	if err != nil {
		t.Fatalf("stat smoke script: %v", err)
	}
	if info.Mode().Perm()&0111 == 0 {
		t.Error("scripts/compose-smoke.sh debe ser ejecutable")
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

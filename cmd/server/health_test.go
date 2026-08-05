package main

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"estimulos-incentivos/internal/handler"
	"estimulos-incentivos/internal/store/sqlite"
)

// Spec (operable-delivery): "expose meaningful health and readiness behavior"
// — /healthz reporta liveness del proceso; /readyz reporta readiness real
// (la base responde y los templates están cargados). Ambos son públicos.

func newHealthTestEnv(t *testing.T) (*sqlite.Store, *handler.TemplateSet) {
	t.Helper()
	store, err := sqlite.New(filepath.Join(t.TempDir(), "health.db"))
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	root, err := resolveAppRoot()
	if err != nil {
		t.Fatalf("root: %v", err)
	}
	tmpls, err := handler.ParseTemplates(root)
	if err != nil {
		t.Fatalf("templates: %v", err)
	}
	return store, tmpls
}

func getHealth(t *testing.T, mux *http.ServeMux, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	return rr
}

func TestHealthzReportsOK(t *testing.T) {
	store, tmpls := newHealthTestEnv(t)
	defer store.Close()
	mux := newHealthMux(store, tmpls, discardLogger())

	rr := getHealth(t, mux, "/healthz")
	if rr.Code != http.StatusOK {
		t.Fatalf("healthz: status %d, want 200", rr.Code)
	}
	if rr.Body.String() != "ok\n" {
		t.Errorf("healthz body = %q, want %q", rr.Body.String(), "ok\n")
	}
}

func TestReadyzReportsReady(t *testing.T) {
	store, tmpls := newHealthTestEnv(t)
	defer store.Close()
	mux := newHealthMux(store, tmpls, discardLogger())

	rr := getHealth(t, mux, "/readyz")
	if rr.Code != http.StatusOK {
		t.Fatalf("readyz: status %d, want 200", rr.Code)
	}
	if rr.Body.String() != "ready\n" {
		t.Errorf("readyz body = %q, want %q", rr.Body.String(), "ready\n")
	}
}

// Readiness debe reflejar el estado real: con la base cerrada, 503.
func TestReadyzReportsNotReadyWhenDatabaseDown(t *testing.T) {
	store, tmpls := newHealthTestEnv(t)
	store.Close() // base no disponible

	mux := newHealthMux(store, tmpls, discardLogger())
	rr := getHealth(t, mux, "/readyz")
	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("readyz con db caída: status %d, want 503", rr.Code)
	}
	if rr.Body.String() != "not ready\n" {
		t.Errorf("readyz body = %q, want %q", rr.Body.String(), "not ready\n")
	}
}

// Readiness también reporta 503 si los templates no están cargados.
func TestReadyzReportsNotReadyWhenTemplatesMissing(t *testing.T) {
	store, _ := newHealthTestEnv(t)
	defer store.Close()

	mux := newHealthMux(store, nil, discardLogger())
	rr := getHealth(t, mux, "/readyz")
	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("readyz sin templates: status %d, want 503", rr.Code)
	}
}

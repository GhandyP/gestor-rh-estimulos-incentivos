package main

import (
	"context"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Spec (operable-delivery): "startup fails before traffic", "shutdown drains
// safely (stops accepting work, closes resources, completes within the bound)"
// y "structured slog startup/request/error events".

func testRunConfig(t *testing.T) Config {
	t.Helper()
	return Config{
		DBPath:          filepath.Join(t.TempDir(), "run.db"),
		Addr:            freeAddr(t),
		ShutdownTimeout: 5 * time.Second,
		ReadTimeout:     5 * time.Second,
		WriteTimeout:    5 * time.Second,
		IdleTimeout:     10 * time.Second,
	}
}

// El ciclo completo: arranca con base real + templates reales + auth de env,
// sirve health y login (render a través del set de templates), y un cancel
// del contexto dispara el apagado acotado que deja de aceptar conexiones.
// El log estructurado contiene eventos de arranque, request y cierre.
func TestRunServesHealthAndShutsDownWithinBound(t *testing.T) {
	setAuthEnv(t)
	t.Setenv("APP_ROOT", "")

	var logBuf syncBuffer
	logger := newTextLogger(&logBuf)
	cfg := testRunConfig(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- run(ctx, cfg, logger) }()

	addr := waitForListening(t, &logBuf)
	base := "http://" + addr
	waitForStatus(t, base+"/readyz", http.StatusOK)

	if code := waitForStatus(t, base+"/healthz", http.StatusOK); code != http.StatusOK {
		t.Fatalf("healthz: %d", code)
	}
	// login se renderiza con el set de templates pre-parsado (cwd-independiente).
	if code := waitForStatus(t, base+"/login", http.StatusOK); code != http.StatusOK {
		t.Fatalf("login: %d", code)
	}

	start := time.Now()
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("run tras cancel: %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("run no retornó dentro del plazo tras cancelar")
	}
	if elapsed := time.Since(start); elapsed >= 10*time.Second {
		t.Errorf("apagado tardó %v", elapsed)
	}

	// Después del shutdown el listener está cerrado: no acepta más trabajo.
	if err := dialAddr(addr); err == nil {
		t.Fatal("el servidor sigue aceptando conexiones tras el shutdown")
	}

	// Eventos estructurados: arranque, listener, request, cierre.
	for _, want := range []string{"starting server", "server listening", "msg=request", "shutdown complete"} {
		if !strings.Contains(logBuf.String(), want) {
			t.Errorf("el log estructurado no contiene %q", want)
		}
	}
}

// Spec: "Missing startup asset fails safely" — sin templates, el arranque
// falla con error y con un evento estructurado de error; jamás llega a escuchar.
func TestRunStartupFailsOnMissingTemplates(t *testing.T) {
	setAuthEnv(t)

	var logBuf syncBuffer
	logger := newTextLogger(&logBuf)
	cfg := testRunConfig(t)
	cfg.AppRoot = t.TempDir() // sin web/templates

	err := run(context.Background(), cfg, logger)
	if err == nil {
		t.Fatal("run debe fallar sin templates")
	}
	if !strings.Contains(err.Error(), "templates") {
		t.Errorf("error = %v, want mención a templates", err)
	}
	if strings.Contains(logBuf.String(), "server listening") {
		t.Error("el servidor escuchó a pesar del fallo de templates")
	}
	if !strings.Contains(logBuf.String(), "level=ERROR") {
		t.Error("no se emitió un evento estructurado de error")
	}
}

// Configuración inválida: una base inabrible (directorio padre inexistente)
// detiene el arranque antes de servir tráfico.
func TestRunStartupFailsOnUnopenableDatabase(t *testing.T) {
	setAuthEnv(t)

	var logBuf syncBuffer
	logger := newTextLogger(&logBuf)
	cfg := testRunConfig(t)
	cfg.DBPath = filepath.Join(t.TempDir(), "missing-dir", "x.db")

	err := run(context.Background(), cfg, logger)
	if err == nil {
		t.Fatal("run debe fallar con base inabrible")
	}
	if !strings.Contains(err.Error(), "database") {
		t.Errorf("error = %v, want mención a database", err)
	}
	if strings.Contains(logBuf.String(), "server listening") {
		t.Error("el servidor escuchó a pesar del fallo de base")
	}
}

// Apagado acotado con request en curso: shutdownServer drena la request
// (completa con 200), no excede el bound y cierra el listener.
func TestShutdownDrainsInFlightRequestWithinBound(t *testing.T) {
	started := make(chan struct{})
	srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		time.Sleep(150 * time.Millisecond) // request lenta en curso
		_, _ = io.WriteString(w, "done")
	})}
	ln := mustListen(t)
	addr := ln.Addr().String()
	go srv.Serve(ln)

	respCh := make(chan int, 1)
	go func() {
		resp, err := http.Get("http://" + addr + "/")
		if err != nil {
			respCh <- -1
			return
		}
		_, _ = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		respCh <- resp.StatusCode
	}()

	<-started // la request YA está en el handler
	start := time.Now()
	if err := shutdownServer(srv, discardLogger(), 3*time.Second); err != nil {
		t.Fatalf("shutdownServer: %v", err)
	}
	elapsed := time.Since(start)

	if elapsed < 100*time.Millisecond {
		t.Errorf("shutdown retornó antes de drenar la request en curso (%v)", elapsed)
	}
	if elapsed >= 3*time.Second {
		t.Errorf("shutdown excedió el bound (%v)", elapsed)
	}
	if code := <-respCh; code != http.StatusOK {
		t.Errorf("request en curso: status %d, want 200", code)
	}
	if err := dialAddr(addr); err == nil {
		t.Fatal("el listener sigue aceptando conexiones tras el shutdown")
	}
}

// Si el bound se agota, shutdownServer fuerza el cierre y reporta el error:
// el proceso termina dentro del plazo, nunca cuelga.
func TestShutdownForceClosesWhenBoundExceeded(t *testing.T) {
	started := make(chan struct{})
	srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		time.Sleep(10 * time.Second) // nunca termina dentro del bound
	})}
	ln := mustListen(t)
	addr := ln.Addr().String()
	go srv.Serve(ln)

	go func() {
		resp, err := http.Get("http://" + addr + "/")
		if err == nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	}()
	<-started

	start := time.Now()
	err := shutdownServer(srv, discardLogger(), 300*time.Millisecond)
	if err == nil {
		t.Fatal("shutdownServer debe reportar error cuando el bound se agota")
	}
	if elapsed := time.Since(start); elapsed >= 2*time.Second {
		t.Errorf("el cierre forzado tardó %v", elapsed)
	}
}

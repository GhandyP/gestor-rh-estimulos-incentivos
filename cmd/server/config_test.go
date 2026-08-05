package main

import (
	"testing"
	"time"
)

// Spec (operable-delivery): "startup validates configuration" — ConfigFromEnv
// arma la configuración desde el entorno y rechaza valores inválidos antes de
// que el servidor acepte tráfico.

func TestConfigFromEnvDefaults(t *testing.T) {
	t.Setenv("DB_PATH", "")
	t.Setenv("PORT", "")
	t.Setenv("SHUTDOWN_TIMEOUT", "")
	t.Setenv("APP_ROOT", "")

	cfg, err := ConfigFromEnv()
	if err != nil {
		t.Fatalf("defaults: %v", err)
	}
	if cfg.DBPath != "estimulos.db" {
		t.Errorf("DBPath = %q, want estimulos.db", cfg.DBPath)
	}
	if cfg.Port != "8080" || cfg.Addr != ":8080" {
		t.Errorf("Port/Addr = %q/%q, want 8080/:8080", cfg.Port, cfg.Addr)
	}
	if cfg.ShutdownTimeout != 15*time.Second {
		t.Errorf("ShutdownTimeout = %v, want 15s", cfg.ShutdownTimeout)
	}
	if cfg.AppRoot != "" {
		t.Errorf("AppRoot = %q, want vacío (resolución por source)", cfg.AppRoot)
	}
}

func TestConfigFromEnvCustomValues(t *testing.T) {
	t.Setenv("DB_PATH", "/tmp/custom.db")
	t.Setenv("PORT", "9090")
	t.Setenv("SHUTDOWN_TIMEOUT", "7s")
	t.Setenv("APP_ROOT", "/opt/app")

	cfg, err := ConfigFromEnv()
	if err != nil {
		t.Fatalf("custom: %v", err)
	}
	if cfg.DBPath != "/tmp/custom.db" {
		t.Errorf("DBPath = %q", cfg.DBPath)
	}
	if cfg.Port != "9090" || cfg.Addr != ":9090" {
		t.Errorf("Port/Addr = %q/%q", cfg.Port, cfg.Addr)
	}
	if cfg.ShutdownTimeout != 7*time.Second {
		t.Errorf("ShutdownTimeout = %v, want 7s", cfg.ShutdownTimeout)
	}
	if cfg.AppRoot != "/opt/app" {
		t.Errorf("AppRoot = %q", cfg.AppRoot)
	}
}

func TestConfigFromEnvRejectsInvalidPort(t *testing.T) {
	for _, bad := range []string{"abc", "0", "65536", "-1", "8080x", "12 34"} {
		t.Run("port="+bad, func(t *testing.T) {
			t.Setenv("PORT", bad)
			if _, err := ConfigFromEnv(); err == nil {
				t.Errorf("PORT=%q debe ser rechazado", bad)
			}
		})
	}
}

func TestConfigFromEnvRejectsInvalidShutdownTimeout(t *testing.T) {
	for _, bad := range []string{"abc", "0s", "-5s", "s"} {
		t.Run("timeout="+bad, func(t *testing.T) {
			t.Setenv("SHUTDOWN_TIMEOUT", bad)
			if _, err := ConfigFromEnv(); err == nil {
				t.Errorf("SHUTDOWN_TIMEOUT=%q debe ser rechazado", bad)
			}
		})
	}
}

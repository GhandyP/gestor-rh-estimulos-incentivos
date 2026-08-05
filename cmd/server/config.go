package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config concentra la configuración del servidor leída del entorno.
type Config struct {
	DBPath          string
	Port            string
	Addr            string
	AppRoot         string
	ShutdownTimeout time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
}

// envString devuelve la variable de entorno o un default si está vacía.
func envString(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// ConfigFromEnv valida y arma la configuración desde variables de entorno.
// Cualquier valor inválido es un error: el arranque falla ANTES de servir
// tráfico. Los timeouts de lectura/escritura conservan los valores actuales.
func ConfigFromEnv() (Config, error) {
	cfg := Config{
		DBPath:          envString("DB_PATH", "estimulos.db"),
		Port:            envString("PORT", "8080"),
		AppRoot:         os.Getenv("APP_ROOT"),
		ShutdownTimeout: 15 * time.Second,
		ReadTimeout:     15 * time.Second,
		WriteTimeout:    15 * time.Second,
		IdleTimeout:     60 * time.Second,
	}

	if n, err := strconv.Atoi(cfg.Port); err != nil || n < 1 || n > 65535 {
		return Config{}, fmt.Errorf("PORT debe ser un número entre 1 y 65535 (recibido %q)", cfg.Port)
	}
	cfg.Addr = ":" + cfg.Port

	if s := os.Getenv("SHUTDOWN_TIMEOUT"); s != "" {
		d, err := time.ParseDuration(s)
		if err != nil || d <= 0 {
			return Config{}, fmt.Errorf("SHUTDOWN_TIMEOUT debe ser una duración positiva (recibido %q)", s)
		}
		cfg.ShutdownTimeout = d
	}

	return cfg, nil
}

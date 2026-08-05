package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

// main valida la configuración, prepara el logger estructurado y delega el
// ciclo de vida completo en run(). Los fallos de arranque se reportan con un
// evento de error estructurado y un exit code no cero, antes de servir tráfico.
func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	cfg, err := ConfigFromEnv()
	if err != nil {
		logger.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, cfg, logger); err != nil {
		logger.Error("server exited with error", "error", err)
		os.Exit(1)
	}
}

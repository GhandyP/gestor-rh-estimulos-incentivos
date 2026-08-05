package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"estimulos-incentivos/internal/handler"
	"estimulos-incentivos/internal/service"
	"estimulos-incentivos/internal/store/sqlite"
)

// statusRecorder captura el status HTTP para el log estructurado de requests.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// requestLogger emite un evento estructurado slog por cada request servida.
func requestLogger(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		logger.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"duration_ms", time.Since(start).Milliseconds(),
			"remote", r.RemoteAddr,
		)
	})
}

// shutdownServer detiene el servidor de forma acotada: cierra el listener
// (deja de aceptar trabajo), drena las requests en curso y devuelve dentro del
// bound. Si el bound se agota, fuerza el cierre y reporta el error.
func shutdownServer(srv *http.Server, logger *slog.Logger, bound time.Duration) error {
	logger.Info("shutting down, draining active requests", "timeout", bound.String())
	shutdownCtx, cancel := context.WithTimeout(context.Background(), bound)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown exceeded bound, forcing close", "error", err)
		_ = srv.Close()
		return fmt.Errorf("graceful shutdown: %w", err)
	}
	logger.Info("shutdown complete")
	return nil
}

// run orquesta el ciclo de vida completo. Todo fallo de configuración, base,
// seed, auth o templates se reporta ANTES de escuchar tráfico. El servidor
// queda escuchando hasta que el contexto se cancele (señal en main) y entonces
// ejecuta el apagado acotado.
func run(ctx context.Context, cfg Config, logger *slog.Logger) error {
	logger.Info("starting server",
		"db_path", cfg.DBPath,
		"addr", cfg.Addr,
		"app_root", cfg.AppRoot,
		"shutdown_timeout", cfg.ShutdownTimeout.String(),
	)

	store, err := sqlite.New(cfg.DBPath)
	if err != nil {
		logger.Error("database initialization failed", "error", err)
		return fmt.Errorf("database: %w", err)
	}
	defer store.Close()

	if err := store.DB().PingContext(ctx); err != nil {
		logger.Error("database ping failed", "error", err)
		return fmt.Errorf("database ping: %w", err)
	}

	svc := service.New(store)
	seedCtx, cancelSeed := context.WithTimeout(ctx, 30*time.Second)
	err = svc.Seed(seedCtx)
	cancelSeed()
	if err != nil {
		logger.Error("seed failed", "error", err)
		return fmt.Errorf("seed: %w", err)
	}

	authCfg, err := handler.AuthConfigFromEnv()
	if err != nil {
		logger.Error("invalid auth configuration", "error", err)
		return fmt.Errorf("auth config: %w", err)
	}
	auth, err := handler.NewAuthenticator(authCfg)
	if err != nil {
		logger.Error("invalid auth configuration", "error", err)
		return fmt.Errorf("auth: %w", err)
	}
	if authCfg.DevDefaults {
		logger.Warn("using development-only operator credentials and session secret; set OPERATOR_USER, OPERATOR_PASSWORD and SESSION_SECRET before production use")
	}

	root := cfg.AppRoot
	if root == "" {
		root, err = resolveAppRoot()
		if err != nil {
			logger.Error("app root resolution failed", "error", err)
			return fmt.Errorf("app root: %w", err)
		}
	}
	tmpls, err := handler.ParseTemplates(root)
	if err != nil {
		logger.Error("template loading failed", "root", root, "error", err)
		return fmt.Errorf("templates: %w", err)
	}

	h := handler.New(svc, auth)
	h.SetTemplates(tmpls)

	appMux := http.NewServeMux()
	h.RegisterRoutes(appMux)
	appHandler := requestLogger(logger, h.Middleware(appMux))

	topMux := http.NewServeMux()
	health := newHealthMux(store, tmpls, logger)
	topMux.Handle("/healthz", health)
	topMux.Handle("/readyz", health)
	topMux.Handle("/", appHandler)

	srv := &http.Server{
		Addr:         cfg.Addr,
		Handler:      topMux,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	ln, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		logger.Error("listen failed", "addr", cfg.Addr, "error", err)
		return fmt.Errorf("listen %s: %w", cfg.Addr, err)
	}
	logger.Info("server listening", "addr", ln.Addr().String())

	errCh := make(chan error, 1)
	go func() { errCh <- srv.Serve(ln) }()

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server failed", "error", err)
			return fmt.Errorf("server: %w", err)
		}
		return nil
	case <-ctx.Done():
		return shutdownServer(srv, logger, cfg.ShutdownTimeout)
	}
}

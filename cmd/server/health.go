package main

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"time"

	"estimulos-incentivos/internal/handler"
	"estimulos-incentivos/internal/store/sqlite"
)

// newHealthMux expone endpoints públicos de salud, fuera de la cadena de
// autenticación, para healthchecks de orquestadores y contenedores:
//   - GET /healthz — liveness: el proceso está vivo.
//   - GET /readyz  — readiness real: los templates están cargados y la base
//     responde a un ping acotado (503 si algo falla).
func newHealthMux(store *sqlite.Store, tmpls *handler.TemplateSet, logger *slog.Logger) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = io.WriteString(w, "ok\n")
	})

	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		if tmpls == nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = io.WriteString(w, "not ready\n")
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := store.DB().PingContext(ctx); err != nil {
			logger.Error("readiness check failed", "error", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = io.WriteString(w, "not ready\n")
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = io.WriteString(w, "ready\n")
	})

	return mux
}

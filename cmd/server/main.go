package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"estimulos-incentivos/internal/handler"
	"estimulos-incentivos/internal/service"
	"estimulos-incentivos/internal/store/sqlite"
)

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "estimulos.db"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	store, err := sqlite.New(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer store.Close()

	svc := service.New(store)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := svc.Seed(ctx); err != nil {
		log.Fatalf("Failed to seed data: %v", err)
	}
	log.Println("Data seeded successfully")

	authCfg, err := handler.AuthConfigFromEnv()
	if err != nil {
		log.Fatalf("Invalid auth configuration: %v", err)
	}
	auth, err := handler.NewAuthenticator(authCfg)
	if err != nil {
		log.Fatalf("Invalid auth configuration: %v", err)
	}
	if authCfg.DevDefaults {
		log.Println("WARNING: using development-only operator credentials and session secret. Set OPERATOR_USER, OPERATOR_PASSWORD and SESSION_SECRET before production use.")
	}

	h := handler.New(svc, auth)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      h.Middleware(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	fmt.Printf("RRHH · Estímulos e Incentivos\n")
	fmt.Printf("Server running on http://localhost:%s\n", port)
	fmt.Printf("Dashboard: http://localhost:%s/\n", port)
	fmt.Printf("API:       http://localhost:%s/api/analisis\n", port)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}
}

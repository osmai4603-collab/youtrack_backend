package main

import (
	"log"
	"net/http"
	"time"

	"youtrack_backend/internal/api"
	"youtrack_backend/internal/app"
	"youtrack_backend/internal/config"
	"youtrack_backend/internal/store/sqlstore"
)

func main() {
	cfg := config.Load()

	store, err := sqlstore.New(cfg.DSN())
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer store.Close()

	a := app.New(store)
	router := api.NewRouter(a, cfg.JWTSecret)

	server := &http.Server{
		Addr:         cfg.Port(),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Server listening on %s (env: %s)", cfg.Port(), cfg.AppEnv)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}

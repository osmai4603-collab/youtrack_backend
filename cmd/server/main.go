package main

import (
	"log"
	"net/http"
	"time"

	"youtrack_backend/channels/api"
	"youtrack_backend/channels/app"
	"youtrack_backend/channels/config"
	"youtrack_backend/channels/model"
	"youtrack_backend/channels/store/sqlstore"
)

func main() {
	cfg := config.Load()

	store, err := sqlstore.New(cfg.DSN())
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer store.Close()

	srv, err := app.NewServer()
	if err != nil {
		model.NewInternalError("init server", "", err)
		return
	}
	router := api.NewServerRouter(srv)

	httpServer := &http.Server{
		Addr:         cfg.Port(),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Server listening on %s (env: %s)", cfg.Port(), cfg.AppEnv)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}

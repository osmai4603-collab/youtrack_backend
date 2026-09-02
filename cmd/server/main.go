package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"youtrack_backend/channels/api"
	"youtrack_backend/channels/app"
	"youtrack_backend/channels/config"
)

func main() {
	cfg := config.Load()

	srv, err := app.NewServer(
		app.WithStore(cfg.DSN()),
		app.WithConfig(cfg),
		app.WithJWTSecret(cfg.JWTSecret),
	)
	if err != nil {
		log.Fatalf("failed to initialize server: %v", err)
	}
	if err := srv.Start(); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}

	router := api.NewServerRouter(srv)
	srv.Server = &http.Server{
		Addr:         cfg.Port(),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("server shutdown error: %v", err)
		}
	}()

	log.Printf("Server listening on %s (env: %s)", cfg.Port(), cfg.AppEnv)
	if err := srv.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server failed: %v", err)
	}
}

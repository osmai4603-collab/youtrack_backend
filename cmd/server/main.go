package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"youtrack_backend/channels/api"
	"youtrack_backend/channels/app"
	"youtrack_backend/channels/config"
	"youtrack_backend/channels/model"
)

type bootstrapDependencies struct {
	newServer func(*model.ServerConfig) (*app.YouTrackServer, error)
	newRouter func(*app.YouTrackServer) http.Handler
}

func defaultBootstrapDependencies() bootstrapDependencies {
	return bootstrapDependencies{
		newServer: func(cfg *model.ServerConfig) (*app.YouTrackServer, error) {
			return app.NewServer(
				app.WithStore(cfg.DSN()),
				app.WithConfig(cfg),
				app.WithJWTSecret(cfg.JWTSecret),
			)
		},
		newRouter: api.NewServerRouter,
	}
}

func run(ctx context.Context, cfg *model.ServerConfig, deps bootstrapDependencies) error {
	if cfg == nil {
		return errors.New("server config is required")
	}
	if deps.newServer == nil || deps.newRouter == nil {
		return errors.New("bootstrap dependencies are required")
	}
	srv, err := deps.newServer(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize server: %w", err)
	}

	router := deps.newRouter(srv)
	if err := srv.SetHTTPServer(&http.Server{
		Addr:         cfg.Port(),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}); err != nil {
		return fmt.Errorf("failed to configure HTTP server: %w", err)
	}

	log.Printf("Server listening on %s (env: %s)", cfg.Port(), cfg.AppEnv)
	shutdownInitiated := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			log.Println("Shutting down server...")
		case <-shutdownInitiated:
		}
	}()
	defer close(shutdownInitiated)

	if err := srv.Run(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server failed: %w", err)
	}
	log.Println("Server stopped successfully")
	return nil
}

func main() {
	cfg := config.Load()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, cfg, defaultBootstrapDependencies()); err != nil {
		log.Fatal(err)
	}
}

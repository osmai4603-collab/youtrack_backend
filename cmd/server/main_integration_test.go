//go:build integration

package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"youtrack_backend/channels/api"
	"youtrack_backend/channels/app"
	"youtrack_backend/channels/model"
)

func TestProductionBootstrapWithPostgres(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("TEST_DATABASE_DSN is required")
	}

	cfg := &model.ServerConfig{
		ServerPort: "0",
		AppEnv:     "test",
		JWTSecret:  "integration-test-secret",
	}
	listenerReady := make(chan net.Listener, 1)
	deps := bootstrapDependencies{
		newServer: func(cfg *model.ServerConfig) (*app.YouTrackServer, error) {
			return app.NewServer(
				app.WithStore(dsn),
				app.WithConfig(cfg),
				app.WithJWTSecret(cfg.JWTSecret),
				app.WithListenerFactory(func(network, address string) (net.Listener, error) {
					listener, err := net.Listen(network, address)
					if err == nil {
						listenerReady <- listener
					}
					return listener, err
				}),
			)
		},
		newRouter: api.NewServerRouter,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	runErr := make(chan error, 1)
	go func() {
		runErr <- run(ctx, cfg, deps)
	}()

	var listener net.Listener
	select {
	case listener = <-listenerReady:
	case <-time.After(10 * time.Second):
		t.Fatal("server did not bind")
	}
	client := &http.Client{Timeout: time.Second}
	deadline := time.Now().Add(10 * time.Second)
	for {
		response, err := client.Get("http://" + listener.Addr().String() + "/ready")
		if err == nil {
			response.Body.Close()
			if response.StatusCode == http.StatusOK {
				break
			}
		}
		if time.Now().After(deadline) {
			t.Fatal("server did not become ready")
		}
		time.Sleep(20 * time.Millisecond)
	}

	cancel()
	select {
	case err := <-runErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			t.Fatalf("run returned error: %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("run did not stop after cancellation")
	}
}

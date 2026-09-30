package main

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"youtrack_backend/channels/api"
	"youtrack_backend/channels/app"
	"youtrack_backend/channels/app/platform"
	"youtrack_backend/channels/model"
)

func TestRunUsesInjectedBootstrapDependencies(t *testing.T) {
	ps, err := platform.New(
		platform.ServiceOptionConfig(&model.ServerConfig{ServerPort: "0", AppEnv: "test"}),
		platform.ServiceOptionJWTSecret("test-secret"),
	)
	if err != nil {
		t.Fatalf("new platform: %v", err)
	}
	routerBuilt := make(chan struct{})
	deps := bootstrapDependencies{
		newServer: func(*model.ServerConfig) (*app.YouTrackServer, error) {
			server, err := app.NewServerWithOptions(ps)
			if err != nil {
				return nil, err
			}
			return server, nil
		},
		newRouter: func(srv *app.YouTrackServer) http.Handler {
			close(routerBuilt)
			return api.NewServerRouter(srv)
		},
	}
	ctx, cancel := context.WithCancel(context.Background())
	runErr := make(chan error, 1)
	go func() {
		runErr <- run(ctx, &model.ServerConfig{ServerPort: "0", AppEnv: "test"}, deps)
	}()

	select {
	case <-routerBuilt:
	case <-time.After(time.Second):
		t.Fatal("router was not constructed")
	}
	cancel()
	select {
	case err := <-runErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			t.Fatalf("run returned error: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("run did not stop after context cancellation")
	}
}

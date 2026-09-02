package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"youtrack_backend/channels/app"
	"youtrack_backend/channels/app/platform"
	"youtrack_backend/channels/model"
)

func TestVersionedAPI(t *testing.T) {
	handler := VersionedAPI(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/health" {
			t.Errorf("expected rewritten path /api/health, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	request := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", response.Code)
	}
	if response.Header().Get("X-API-Version") != "v1" {
		t.Fatalf("expected X-API-Version v1, got %q", response.Header().Get("X-API-Version"))
	}
	if request.URL.Path != "/api/v1/health" {
		t.Fatalf("expected request path restoration, got %s", request.URL.Path)
	}
}

func TestHealthEndpointReflectsReadiness(t *testing.T) {
	srv, err := app.NewServer(
		app.WithPlatformService(mustPlatform()),
	)
	if err != nil {
		t.Fatalf("new server: %v", err)
	}
	if err := srv.Start(); err != nil {
		t.Fatalf("server start: %v", err)
	}
	if srv.Router == nil {
		srv.Router = srv.RootRouter
	}
	if srv.RootRouter == nil {
		srv.RootRouter = srv.Router
	}
	Init(srv)

	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	response := httptest.NewRecorder()
	srv.Router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200 when server is ready, got %d", response.Code)
	}

	if err := srv.Shutdown(context.Background()); err != nil {
		t.Fatalf("server shutdown: %v", err)
	}

	response = httptest.NewRecorder()
	srv.Router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health", nil))
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 when server is not ready, got %d", response.Code)
	}
}

func mustPlatform() *platform.YouTrackPlatformService {
	ps, err := platform.New(
		platform.ServiceOptionConfig(&model.ServerConfig{ServerPort: "8080", AppEnv: "test"}),
		platform.ServiceOptionJWTSecret("test-secret"),
	)
	if err != nil {
		panic(err)
	}
	return ps
}

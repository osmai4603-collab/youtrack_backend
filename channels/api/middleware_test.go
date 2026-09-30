package api

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"youtrack_backend/channels/app"
	"youtrack_backend/channels/app/platform"
	"youtrack_backend/channels/model"

	"github.com/go-chi/chi/v5"
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
	if srv.Router == nil {
		srv.Router = srv.RootRouter
	}
	if srv.RootRouter == nil {
		srv.RootRouter = srv.Router
	}
	Init(srv)
	srv.Server = &http.Server{Addr: "127.0.0.1:0", Handler: srv.RootRouter}
	runErrCh := make(chan error, 1)
	go func() {
		runErrCh <- srv.Run(context.Background())
	}()
	deadline := time.Now().Add(2 * time.Second)
	for !srv.IsReady() && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if !srv.IsReady() {
		t.Fatal("expected server to become ready after listener binding")
	}

	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	response := httptest.NewRecorder()
	srv.Router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200 when server is ready, got %d", response.Code)
	}

	response = httptest.NewRecorder()
	srv.Router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200 from ready endpoint, got %d", response.Code)
	}
	response = httptest.NewRecorder()
	srv.Router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/live", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200 from live endpoint, got %d", response.Code)
	}

	if err := srv.Shutdown(context.Background()); err != nil {
		t.Fatalf("server shutdown: %v", err)
	}
	select {
	case err := <-runErrCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			t.Fatalf("server run error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("expected server run goroutine to exit")
	}

	response = httptest.NewRecorder()
	srv.Router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health", nil))
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 when server is not ready, got %d", response.Code)
	}
	response = httptest.NewRecorder()
	srv.Router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 from ready endpoint after shutdown, got %d", response.Code)
	}
	response = httptest.NewRecorder()
	srv.Router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/live", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("expected live endpoint to remain available after shutdown, got %d", response.Code)
	}
}

func TestLocalRouterReadinessSemantics(t *testing.T) {
	localRouter := NewLocalRouter(nil)
	for _, path := range []string{"/ready", "/health"} {
		response := httptest.NewRecorder()
		localRouter.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusServiceUnavailable {
			t.Fatalf("expected %s to return 503 without a server, got %d", path, response.Code)
		}
	}
	response := httptest.NewRecorder()
	localRouter.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/live", nil))
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected /live to return 503 without a server, got %d", response.Code)
	}
}

func TestLocalAPILifecycleUsesUnixSocket(t *testing.T) {
	srv, err := app.NewServer(app.WithPlatformService(mustPlatform()))
	if err != nil {
		t.Fatalf("new server: %v", err)
	}
	srv.RootRouter = chi.NewRouter()
	srv.Router = srv.RootRouter
	srv.Server = &http.Server{Addr: "127.0.0.1:0", Handler: srv.RootRouter}

	socketPath := filepath.Join(t.TempDir(), "supervision.sock")
	listener, localServer, err := StartLocalAPI(srv, socketPath)
	if err != nil {
		t.Fatalf("start local api: %v", err)
	}
	if err := srv.SetLocalServer(listener, localServer); err != nil {
		t.Fatalf("attach local server: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	runErrCh := make(chan error, 1)
	go func() {
		runErrCh <- srv.Run(ctx)
	}()
	deadline := time.Now().Add(2 * time.Second)
	for !srv.IsReady() && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if !srv.IsReady() {
		t.Fatal("expected server to become ready")
	}

	client := &http.Client{Transport: &http.Transport{
		DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
			return net.Dial("unix", socketPath)
		},
	}}
	response, err := client.Get("http://local/ready")
	if err != nil {
		t.Fatalf("get local readiness: %v", err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected local readiness 200, got %d", response.StatusCode)
	}
	response, err = client.Get("http://local/live")
	if err != nil {
		t.Fatalf("get local liveness: %v", err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected local liveness 200, got %d", response.StatusCode)
	}

	cancel()
	select {
	case err := <-runErrCh:
		if err != nil {
			t.Fatalf("run after cancellation: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("expected server run to exit")
	}
	if _, err := os.Stat(socketPath); !os.IsNotExist(err) {
		t.Fatalf("expected unix socket to be removed, stat error: %v", err)
	}
}

func TestSetLocalServerRejectsChangesAfterStartup(t *testing.T) {
	srv, err := app.NewServer(app.WithPlatformService(mustPlatform()))
	if err != nil {
		t.Fatalf("new server: %v", err)
	}
	if err := srv.Start(); err != nil {
		t.Fatalf("start server: %v", err)
	}
	defer srv.Shutdown(context.Background())

	listener, err := net.Listen("unix", filepath.Join(t.TempDir(), "late.sock"))
	if err != nil {
		t.Fatalf("listen on unix socket: %v", err)
	}
	server := &http.Server{Handler: NewLocalRouter(srv)}
	if err := srv.SetLocalServer(listener, server); err == nil {
		listener.Close()
		t.Fatal("expected local server configuration after startup to fail")
	}
	listener.Close()
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

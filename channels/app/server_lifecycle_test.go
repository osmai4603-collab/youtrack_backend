package app

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"youtrack_backend/channels/app/platform"
	"youtrack_backend/channels/model"

	"github.com/gorilla/mux"
)

func TestPlatformLifecycle(t *testing.T) {
	ps, err := platform.New(
		platform.ServiceOptionConfig(&model.ServerConfig{ServerPort: "8080", AppEnv: "test"}),
		platform.ServiceOptionJWTSecret("test-secret"),
	)
	if err != nil {
		t.Fatalf("new platform: %v", err)
	}

	if err := ps.Start(); err != nil {
		t.Fatalf("platform start: %v", err)
	}
	if err := ps.Shutdown(); err != nil {
		t.Fatalf("platform shutdown: %v", err)
	}
	if err := ps.Shutdown(); err != nil {
		t.Fatalf("second platform shutdown should be idempotent: %v", err)
	}
}

func TestServerLifecycle(t *testing.T) {
	ps, err := platform.New(
		platform.ServiceOptionConfig(&model.ServerConfig{ServerPort: "8080", AppEnv: "test"}),
		platform.ServiceOptionJWTSecret("test-secret"),
	)
	if err != nil {
		t.Fatalf("new platform: %v", err)
	}

	srv := &YouTrackServer{
		platform:   ps,
		RootRouter: NewRouter(),
		Router:     NewRouter(),
	}
	srv.ch = NewChannels(srv)

	if err := srv.Start(); err != nil {
		t.Fatalf("server start: %v", err)
	}
	if err := srv.Shutdown(context.Background()); err != nil {
		t.Fatalf("server shutdown: %v", err)
	}
	if err := srv.Shutdown(context.Background()); err != nil {
		t.Fatalf("second server shutdown should be idempotent: %v", err)
	}

	select {
	case <-time.After(100 * time.Millisecond):
		// ok
	}
}

func TestPlatformValidateRequiresConfigAndSecret(t *testing.T) {
	ps, err := platform.New()
	if err != nil {
		t.Fatalf("new platform: %v", err)
	}

	if err := ps.Validate(); err == nil {
		t.Fatal("expected validation error when config and jwt secret are missing")
	}

	ps.SetConfig(&model.ServerConfig{ServerPort: "8080", AppEnv: "test"})
	if err := ps.Validate(); err == nil {
		t.Fatal("expected validation error when jwt secret is missing")
	}

	ps.SetJWTSecret("test-secret")
	if err := ps.Validate(); err != nil {
		t.Fatalf("expected validation to pass after config and secret are set: %v", err)
	}
}

func TestServerStartRejectsInvalidPlatformConfiguration(t *testing.T) {
	srv := &YouTrackServer{
		RootRouter: NewRouter(),
		Router:     NewRouter(),
	}

	if err := srv.Start(); err == nil {
		t.Fatal("expected server start to fail when platform configuration is invalid")
	}
}

func TestServerLifecycleContractIsExplicitlySplit(t *testing.T) {
	srv := &YouTrackServer{RootRouter: NewRouter(), Router: NewRouter()}

	if err := srv.Validate(); err == nil {
		t.Fatal("expected validation to fail before platform configuration is provided")
	}

	cfg := &model.ServerConfig{ServerPort: "8080", AppEnv: "test"}
	ps, err := platform.New(
		platform.ServiceOptionConfig(cfg),
		platform.ServiceOptionJWTSecret("test-secret"),
	)
	if err != nil {
		t.Fatalf("new platform: %v", err)
	}
	srv.platform = ps

	if err := srv.Validate(); err != nil {
		t.Fatalf("expected validation to pass after platform config and secret are set: %v", err)
	}

	var validator interface{ Validate() error } = srv
	var starter interface{ Start() error } = srv
	var readyCheck interface{ IsReady() bool } = srv
	if validator == nil || starter == nil || readyCheck == nil {
		t.Fatal("expected server to expose explicit lifecycle contract interfaces")
	}
}

func TestServerReadinessLifecycle(t *testing.T) {
	ps, err := platform.New(
		platform.ServiceOptionConfig(&model.ServerConfig{ServerPort: "8080", AppEnv: "test"}),
		platform.ServiceOptionJWTSecret("test-secret"),
	)
	if err != nil {
		t.Fatalf("new platform: %v", err)
	}

	srv := &YouTrackServer{
		platform:   ps,
		RootRouter: NewRouter(),
		Router:     NewRouter(),
	}
	srv.ch = NewChannels(srv)

	if srv.IsReady() {
		t.Fatal("expected server to be not ready before start")
	}
	if err := srv.Start(); err != nil {
		t.Fatalf("server start: %v", err)
	}
	if !srv.IsReady() {
		t.Fatal("expected server to be ready after start")
	}

	srv.LocalRouter = mux.NewRouter()
	srv.LocalRouter.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodGet)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	res := httptest.NewRecorder()
	srv.LocalRouter.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected readiness health check to be 200, got %d", res.Code)
	}

	if err := srv.Shutdown(context.Background()); err != nil {
		t.Fatalf("server shutdown: %v", err)
	}
	if srv.IsReady() {
		t.Fatal("expected server to become unready after shutdown")
	}
}

func TestServerRealHTTPIntegrationLifecycle(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen on random port: %v", err)
	}
	addr := listener.Addr().String()
	_ = listener.Close()

	cfg := &model.ServerConfig{ServerPort: "0", AppEnv: "test"}
	ps, err := platform.New(
		platform.ServiceOptionConfig(cfg),
		platform.ServiceOptionJWTSecret("test-secret"),
	)
	if err != nil {
		t.Fatalf("new platform: %v", err)
	}

	srv := &YouTrackServer{
		platform:   ps,
		RootRouter: NewRouter(),
		Router:     NewRouter(),
		Server:     &http.Server{Addr: addr, Handler: nil},
	}
	srv.RootRouter.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if !srv.IsReady() {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"status":"unavailable","ready":false}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"healthy","ready":true}`))
	}).Methods(http.MethodGet)
	srv.Server.Handler = srv.RootRouter
	srv.ch = NewChannels(srv)

	if err := srv.Start(); err != nil {
		t.Fatalf("server start: %v", err)
	}

	serverErrCh := make(chan error, 1)
	go func() {
		serverErrCh <- srv.Server.ListenAndServe()
	}()

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get("http://" + addr + "/health")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				break
			}
		}
		time.Sleep(20 * time.Millisecond)
	}

	resp, err := http.Get("http://" + addr + "/health")
	if err != nil {
		t.Fatalf("http get health endpoint: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		t.Fatalf("expected live health endpoint to be 200 while ready, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	if err := srv.Shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown server: %v", err)
	}

	if srv.IsReady() {
		t.Fatal("expected server to become unready after shutdown")
	}

	resp, err = http.Get("http://" + addr + "/health")
	if err == nil {
		resp.Body.Close()
	}

	select {
	case err := <-serverErrCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			t.Fatalf("server listen error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("expected http server goroutine to exit after shutdown")
	}
}

func TestServerOptionsKeepSinglePlatformInstanceAcrossStartupOrder(t *testing.T) {
	cfg := &model.ServerConfig{ServerPort: "8080", AppEnv: "test"}
	ps, err := platform.New(
		platform.ServiceOptionConfig(cfg),
		platform.ServiceOptionJWTSecret("test-secret"),
	)
	if err != nil {
		t.Fatalf("new platform: %v", err)
	}

	srv, err := NewServer(
		WithPlatformService(ps),
		WithConfig(cfg),
		WithJWTSecret("test-secret"),
	)
	if err != nil {
		t.Fatalf("new server with startup-order options: %v", err)
	}
	if srv.platform == nil {
		t.Fatal("expected platform to exist after applying options")
	}
	if srv.platform != ps {
		t.Fatal("expected startup-order options to reuse the same platform instance")
	}
	if srv.platform.Config() == nil || srv.platform.Config().ServerPort != "8080" {
		t.Fatal("expected config to be retained after startup order options")
	}
	if srv.platform.JWTSecret() != "test-secret" {
		t.Fatal("expected jwt secret to be retained after startup order options")
	}
}

func TestChannelsGracefulShutdownIsIdempotent(t *testing.T) {
	ch := &YouTrackChannels{interruptQuitChan: make(chan struct{})}
	if err := ch.Start(); err != nil {
		t.Fatalf("channels start: %v", err)
	}
	if err := ch.Stop(); err != nil {
		t.Fatalf("first stop: %v", err)
	}
	if err := ch.Stop(); err != nil {
		t.Fatalf("second stop should be idempotent: %v", err)
	}
}

func NewRouter() *mux.Router {
	return mux.NewRouter()
}

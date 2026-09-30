package app

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"youtrack_backend/channels/app/platform"
	"youtrack_backend/channels/model"

	"github.com/go-chi/chi/v5"
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
	if srv.LifecycleState() != LifecycleCreated {
		t.Fatalf("expected created state, got %s", srv.LifecycleState())
	}

	if err := srv.Start(); err != nil {
		t.Fatalf("server start: %v", err)
	}
	if srv.LifecycleState() != LifecycleStarting {
		t.Fatalf("expected starting state after Start, got %s", srv.LifecycleState())
	}
	if err := srv.Shutdown(context.Background()); err != nil {
		t.Fatalf("server shutdown: %v", err)
	}
	if srv.LifecycleState() != LifecycleStopped {
		t.Fatalf("expected stopped state after shutdown, got %s", srv.LifecycleState())
	}
	if err := srv.Shutdown(context.Background()); err != nil {
		t.Fatalf("second server shutdown should be idempotent: %v", err)
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
	if srv.LifecycleState() != LifecycleFailed {
		t.Fatalf("expected failed state after invalid startup, got %s", srv.LifecycleState())
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
	if srv.IsReady() {
		t.Fatal("expected server to remain not ready until a listener is bound")
	}
	srv.Server = &http.Server{Addr: "127.0.0.1:0", Handler: srv.RootRouter}
	serverErrCh := make(chan error, 1)
	go func() {
		serverErrCh <- srv.Run(context.Background())
	}()
	deadline := time.Now().Add(2 * time.Second)
	for !srv.IsReady() && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if !srv.IsReady() {
		t.Fatal("expected server to be ready after listener binding")
	}
	if srv.LifecycleState() != LifecycleRunning {
		t.Fatalf("expected running state after listener binding, got %s", srv.LifecycleState())
	}

	srv.LocalRouter = chi.NewRouter()
	srv.LocalRouter.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

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
	if srv.listener != nil {
		t.Fatal("expected main listener ownership to be released after shutdown")
	}
	if srv.LifecycleState() != LifecycleStopped {
		t.Fatalf("expected stopped state after shutdown, got %s", srv.LifecycleState())
	}
	select {
	case err := <-serverErrCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			t.Fatalf("server run error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("expected server run goroutine to exit after shutdown")
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
	srv.RootRouter.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		if !srv.IsReady() {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"status":"unavailable","ready":false}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"healthy","ready":true}`))
	})
	srv.Server.Handler = srv.RootRouter
	srv.ch = NewChannels(srv)

	serverErrCh := make(chan error, 1)
	go func() {
		serverErrCh <- srv.Run(context.Background())
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

func TestServerRunFailsBeforeReadinessWhenListenerCannotBind(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen on random port: %v", err)
	}
	defer listener.Close()

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
		Server:     &http.Server{Addr: listener.Addr().String(), Handler: NewRouter()},
	}
	srv.ch = NewChannels(srv)

	if err := srv.Run(context.Background()); err == nil {
		t.Fatal("expected bind failure")
	}
	if srv.IsReady() {
		t.Fatal("expected server to remain unready after bind failure")
	}
}

func TestServerUsesInjectedListenerFactory(t *testing.T) {
	ps, err := platform.New(
		platform.ServiceOptionConfig(&model.ServerConfig{ServerPort: "8080", AppEnv: "test"}),
		platform.ServiceOptionJWTSecret("test-secret"),
	)
	if err != nil {
		t.Fatalf("new platform: %v", err)
	}
	expectedErr := errors.New("injected bind failure")
	srv, err := NewServer(
		WithPlatformService(ps),
		WithListenerFactory(func(network, address string) (net.Listener, error) {
			if network != "tcp" || address != ":8080" {
				t.Fatalf("unexpected listener arguments: %s %s", network, address)
			}
			return nil, expectedErr
		}),
	)
	if err != nil {
		t.Fatalf("new server: %v", err)
	}

	if err := srv.Run(context.Background()); err == nil || !strings.Contains(err.Error(), expectedErr.Error()) {
		t.Fatalf("expected injected bind error, got %v", err)
	}
	if srv.IsReady() {
		t.Fatal("expected server to remain unready after injected bind failure")
	}
}

func TestServerBindFailureRollsBackPlatformWorkers(t *testing.T) {
	ps, err := platform.New(
		platform.ServiceOptionConfig(&model.ServerConfig{ServerPort: "8080", AppEnv: "test"}),
		platform.ServiceOptionJWTSecret("test-secret"),
	)
	if err != nil {
		t.Fatalf("new platform: %v", err)
	}
	expectedErr := errors.New("injected bind failure")
	workerFinished := make(chan struct{})
	var srv *YouTrackServer
	srv, err = NewServer(
		WithPlatformService(ps),
		WithListenerFactory(func(network, address string) (net.Listener, error) {
			if !srv.Channels().StartWorker(func(ctx context.Context) {
				<-ctx.Done()
				close(workerFinished)
			}) {
				t.Fatal("expected rollback worker to start")
			}
			return nil, expectedErr
		}),
	)
	if err != nil {
		t.Fatalf("new server: %v", err)
	}

	err = srv.Run(context.Background())
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected bind error to be preserved, got %v", err)
	}
	select {
	case <-workerFinished:
	case <-time.After(time.Second):
		t.Fatal("expected startup rollback to cancel platform workers")
	}
}

func TestServerRunRejectsConcurrentRun(t *testing.T) {
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
		Server:     &http.Server{Addr: "127.0.0.1:0", Handler: NewRouter()},
	}
	srv.ch = NewChannels(srv)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	firstErr := make(chan error, 1)
	go func() { firstErr <- srv.Run(ctx) }()

	deadline := time.Now().Add(2 * time.Second)
	for !srv.IsReady() && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if !srv.IsReady() {
		t.Fatal("expected first run to become ready")
	}
	if err := srv.Run(context.Background()); err == nil || !strings.Contains(err.Error(), "run has already started") {
		t.Fatalf("expected concurrent run to be rejected, got %v", err)
	}
	cancel()
	if err := <-firstErr; err != nil {
		t.Fatalf("first run after cancellation: %v", err)
	}
}

func TestSetHTTPServerRequiresCreatedState(t *testing.T) {
	ps, err := platform.New(
		platform.ServiceOptionConfig(&model.ServerConfig{ServerPort: "8080", AppEnv: "test"}),
		platform.ServiceOptionJWTSecret("test-secret"),
	)
	if err != nil {
		t.Fatalf("new platform: %v", err)
	}
	srv := &YouTrackServer{platform: ps, RootRouter: NewRouter(), Router: NewRouter()}
	srv.ch = NewChannels(srv)
	if err := srv.SetHTTPServer(&http.Server{Handler: srv.RootRouter}); err != nil {
		t.Fatalf("set HTTP server: %v", err)
	}
	if err := srv.Start(); err != nil {
		t.Fatalf("server start: %v", err)
	}
	if err := srv.SetHTTPServer(&http.Server{}); err == nil {
		t.Fatal("expected HTTP server replacement after startup to fail")
	}
	if err := srv.Shutdown(context.Background()); err != nil {
		t.Fatalf("server shutdown: %v", err)
	}
}

func TestServerRunStopsWhenContextIsCanceled(t *testing.T) {
	ps, err := platform.New(
		platform.ServiceOptionConfig(&model.ServerConfig{ServerPort: "8080", AppEnv: "test"}),
		platform.ServiceOptionJWTSecret("test-secret"),
	)
	if err != nil {
		t.Fatalf("new platform: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	srv := &YouTrackServer{
		platform:   ps,
		RootRouter: NewRouter(),
		Router:     NewRouter(),
		Server:     &http.Server{Addr: "127.0.0.1:0", Handler: NewRouter()},
	}
	srv.ch = NewChannels(srv)
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
	cancel()

	select {
	case err := <-runErrCh:
		if err != nil {
			t.Fatalf("run after context cancellation: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("expected Run to stop after context cancellation")
	}
	if srv.IsReady() {
		t.Fatal("expected server to become unready after context cancellation")
	}
}

func TestServerShutdownIsSafeWhenCalledConcurrently(t *testing.T) {
	ps, err := platform.New(
		platform.ServiceOptionConfig(&model.ServerConfig{ServerPort: "8080", AppEnv: "test"}),
		platform.ServiceOptionJWTSecret("test-secret"),
	)
	if err != nil {
		t.Fatalf("new platform: %v", err)
	}
	srv := &YouTrackServer{platform: ps, RootRouter: NewRouter(), Router: NewRouter()}
	srv.ch = NewChannels(srv)
	if err := srv.Start(); err != nil {
		t.Fatalf("server start: %v", err)
	}

	errs := make(chan error, 8)
	for range 8 {
		go func() {
			errs <- srv.Shutdown(context.Background())
		}()
	}
	for range 8 {
		if err := <-errs; err != nil {
			t.Fatalf("concurrent shutdown: %v", err)
		}
	}
	if srv.LifecycleState() != LifecycleStopped {
		t.Fatalf("expected stopped state, got %s", srv.LifecycleState())
	}
}

func TestServerStartIsIdempotentAndRejectsRestart(t *testing.T) {
	ps, err := platform.New(
		platform.ServiceOptionConfig(&model.ServerConfig{ServerPort: "8080", AppEnv: "test"}),
		platform.ServiceOptionJWTSecret("test-secret"),
	)
	if err != nil {
		t.Fatalf("new platform: %v", err)
	}
	srv := &YouTrackServer{platform: ps, RootRouter: NewRouter(), Router: NewRouter()}
	srv.ch = NewChannels(srv)

	if err := srv.Start(); err != nil {
		t.Fatalf("first server start: %v", err)
	}
	if err := srv.Start(); err != nil {
		t.Fatalf("repeated server start should be idempotent: %v", err)
	}
	if err := srv.Shutdown(context.Background()); err != nil {
		t.Fatalf("server shutdown: %v", err)
	}
	if err := srv.Start(); err == nil {
		t.Fatal("expected restart after shutdown to fail")
	}
}

func TestServerStartAndShutdownAreSerialized(t *testing.T) {
	ps, err := platform.New(
		platform.ServiceOptionConfig(&model.ServerConfig{ServerPort: "8080", AppEnv: "test"}),
		platform.ServiceOptionJWTSecret("test-secret"),
	)
	if err != nil {
		t.Fatalf("new platform: %v", err)
	}
	srv := &YouTrackServer{platform: ps, RootRouter: NewRouter(), Router: NewRouter()}
	srv.ch = NewChannels(srv)

	startErr := make(chan error, 1)
	shutdownErr := make(chan error, 1)
	go func() { startErr <- srv.Start() }()
	go func() { shutdownErr <- srv.Shutdown(context.Background()) }()

	startResult := <-startErr
	shutdownResult := <-shutdownErr
	if startResult != nil && !strings.Contains(startResult.Error(), "server cannot start from stopped state") {
		t.Fatalf("unexpected server start error: %v", startResult)
	}
	if shutdownResult != nil {
		t.Fatalf("server shutdown: %v", shutdownResult)
	}
	if srv.LifecycleState() != LifecycleStopped {
		t.Fatalf("expected stopped state after serialized lifecycle operations, got %s", srv.LifecycleState())
	}
}

func TestChannelsStartWorkerUsesPlatformLifecycle(t *testing.T) {
	ps, err := platform.New(
		platform.ServiceOptionConfig(&model.ServerConfig{ServerPort: "8080", AppEnv: "test"}),
		platform.ServiceOptionJWTSecret("test-secret"),
	)
	if err != nil {
		t.Fatalf("new platform: %v", err)
	}
	srv := &YouTrackServer{platform: ps, RootRouter: NewRouter(), Router: NewRouter()}
	srv.ch = NewChannels(srv)
	if err := srv.Start(); err != nil {
		t.Fatalf("server start: %v", err)
	}

	started := make(chan struct{})
	finished := make(chan struct{})
	if !srv.Channels().StartWorker(func(ctx context.Context) {
		close(started)
		<-ctx.Done()
		if srv.IsReady() {
			t.Error("expected readiness to be false before worker cancellation")
		}
		close(finished)
	}) {
		t.Fatal("expected channels worker to start")
	}
	<-started
	if err := srv.Shutdown(context.Background()); err != nil {
		t.Fatalf("server shutdown: %v", err)
	}
	select {
	case <-finished:
	default:
		t.Fatal("expected channels worker to finish before shutdown returns")
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

func TestChannelsStopBeforeStartIsSafe(t *testing.T) {
	ch := NewChannels(nil)
	if err := ch.Stop(); err != nil {
		t.Fatalf("stop before start: %v", err)
	}
	if err := ch.Stop(); err != nil {
		t.Fatalf("second stop before start: %v", err)
	}
}

func TestChannelsStopCancelsRegisteredScheduledTasks(t *testing.T) {
	ch := NewChannels(nil)
	started := make(chan struct{})
	release := make(chan struct{})
	task := model.CreateRecurringTask("test-worker", func() {
		select {
		case <-started:
		default:
			close(started)
		}
		<-release
	}, time.Millisecond)
	if !ch.RegisterScheduledTask(task) {
		t.Fatal("expected scheduled task registration to succeed")
	}
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("scheduled task did not start")
	}

	stopDone := make(chan struct{})
	go func() {
		_ = ch.Stop()
		close(stopDone)
	}()
	select {
	case <-stopDone:
		t.Fatal("expected Stop to wait for scheduled task cancellation")
	case <-time.After(10 * time.Millisecond):
	}
	close(release)
	select {
	case <-stopDone:
	case <-time.After(time.Second):
		t.Fatal("expected Stop to finish after scheduled task cancellation")
	}
	if ch.RegisterScheduledTask(model.CreateTask("late-worker", func() {}, time.Millisecond)) {
		t.Fatal("expected task registration after Stop to fail")
	}
}

func TestChannelsStartScheduledTaskStartsAfterOwnership(t *testing.T) {
	ch := NewChannels(nil)
	started := make(chan struct{})
	task := model.NewRecurringTask("owned-worker", func() {
		select {
		case <-started:
		default:
			close(started)
		}
	}, time.Millisecond)
	if !ch.StartScheduledTask(task) {
		t.Fatal("expected scheduled task to start")
	}
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("owned scheduled task did not start")
	}
	if err := ch.Stop(); err != nil {
		t.Fatalf("stop channels: %v", err)
	}
}

func NewRouter() *chi.Mux {
	return chi.NewRouter()
}

package platform

import (
	"context"
	"errors"
	"testing"
	"time"

	"youtrack_backend/channels/model"
)

type mockTestStore struct{}

func (m *mockTestStore) Close() error { return nil }

func TestPlatformServiceLifecycle(t *testing.T) {
	mockStore := &mockTestStore{}
	cfg := &model.ServerConfig{
		ServerPort: "8080",
		AppEnv:     "test",
	}

	ps, err := New(
		ServiceOptionConfig(cfg),
		ServiceOptionJWTSecret("secret-123"),
	)
	if err != nil {
		t.Fatalf("unexpected error creating platform service: %v", err)
	}

	if ps.Config().ServerPort != "8080" {
		t.Errorf("expected ServerPort 8080, got %s", ps.Config().ServerPort)
	}
	if ps.JWTSecret() != "secret-123" {
		t.Errorf("expected secret-123, got %s", ps.JWTSecret())
	}

	ps.SetJWTSecret("new-secret")
	if ps.JWTSecret() != "new-secret" {
		t.Errorf("expected new-secret, got %s", ps.JWTSecret())
	}

	if err := ps.Start(); err != nil {
		t.Errorf("failed to start platform service: %v", err)
	}

	if err := ps.Shutdown(); err != nil {
		t.Errorf("failed to shutdown platform service: %v", err)
	}

	_ = mockStore
}

func TestPlatformWorkersAreCanceledAndDrainedBeforeShutdownReturns(t *testing.T) {
	ps, err := New(
		ServiceOptionConfig(&model.ServerConfig{ServerPort: "8080", AppEnv: "test"}),
		ServiceOptionJWTSecret("secret-123"),
	)
	if err != nil {
		t.Fatalf("new platform service: %v", err)
	}
	if err := ps.Start(); err != nil {
		t.Fatalf("start platform service: %v", err)
	}

	started := make(chan struct{})
	finished := make(chan struct{})
	if !ps.Go(func() {
		close(started)
		<-ps.Context().Done()
		close(finished)
	}) {
		t.Fatal("expected platform worker to start")
	}
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("worker did not start")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := ps.ShutdownContext(shutdownCtx); err != nil {
		t.Fatalf("shutdown platform service: %v", err)
	}
	select {
	case <-finished:
	default:
		t.Fatal("expected shutdown to drain platform worker")
	}
	if ps.Go(func() {}) {
		t.Fatal("expected workers to be rejected after shutdown")
	}
}

func TestPlatformGoContextProvidesCancellationContext(t *testing.T) {
	ps, err := New(
		ServiceOptionConfig(&model.ServerConfig{ServerPort: "8080", AppEnv: "test"}),
		ServiceOptionJWTSecret("secret-123"),
	)
	if err != nil {
		t.Fatalf("new platform service: %v", err)
	}
	if err := ps.Start(); err != nil {
		t.Fatalf("start platform service: %v", err)
	}

	started := make(chan struct{})
	finished := make(chan struct{})
	if !ps.GoContext(func(ctx context.Context) {
		close(started)
		<-ctx.Done()
		close(finished)
	}) {
		t.Fatal("expected context-aware worker to start")
	}
	<-started
	if err := ps.ShutdownContext(context.Background()); err != nil {
		t.Fatalf("shutdown platform service: %v", err)
	}
	select {
	case <-finished:
	default:
		t.Fatal("expected context-aware worker to finish before shutdown returns")
	}
}

func TestPlatformShutdownReportsWorkerTimeout(t *testing.T) {
	ps, err := New(
		ServiceOptionConfig(&model.ServerConfig{ServerPort: "8080", AppEnv: "test"}),
		ServiceOptionJWTSecret("secret-123"),
	)
	if err != nil {
		t.Fatalf("new platform service: %v", err)
	}
	if err := ps.Start(); err != nil {
		t.Fatalf("start platform service: %v", err)
	}
	workerStarted := make(chan struct{})
	workerRelease := make(chan struct{})
	if !ps.Go(func() {
		close(workerStarted)
		<-workerRelease
	}) {
		t.Fatal("expected platform worker to start")
	}
	<-workerStarted

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	err = ps.ShutdownContext(shutdownCtx)
	if err == nil || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected worker timeout error, got %v", err)
	}
	close(workerRelease)
}

package platform

import (
	"testing"

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

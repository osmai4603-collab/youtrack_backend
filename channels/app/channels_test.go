package app

import (
	"testing"

	"youtrack_backend/channels/app/platform"
	"youtrack_backend/channels/model"
)

func TestChannelsCoordination(t *testing.T) {
	cfg := &model.ServerConfig{
		ServerPort: "9000",
	}
	ps, err := platform.New(
		platform.ServiceOptionConfig(cfg),
		platform.ServiceOptionJWTSecret("jwt-token-key"),
	)
	if err != nil {
		t.Fatalf("unexpected error creating platform service: %v", err)
	}

	server := &YouTrackServer{platform: ps}
	server.ch = NewChannels(server)

	if server.Platform() != ps {
		t.Errorf("expected platform pointer match")
	}
	if server.Config().ServerPort != "9000" {
		t.Errorf("expected ServerPort 9000, got %s", server.Config().ServerPort)
	}
	if server.JWTSecret() != "jwt-token-key" {
		t.Errorf("expected jwt-token-key, got %s", server.JWTSecret())
	}
}

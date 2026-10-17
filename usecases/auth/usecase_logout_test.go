package authusecase

import (
	"context"
	"testing"
	"time"

	"youtrack_backend/internal/domain"
	authstorage "youtrack_backend/internal/storage/auth"
)

func TestAuthLogoutUseCase_DeletesSession(t *testing.T) {
	ctx := context.Background()
	sessions := authstorage.NewInMemorySessionRepository()
	uc := NewLogout(sessions)

	now := time.Now().UTC()
	sess := domain.UserSession{
		ID:        "sess-1",
		SessionID: "sess-1",
		UserID:    "user-1",
		CreatedAt: now,
		UpdatedAt: now,
		ExpiresAt: now.Add(1 * time.Hour),
	}
	_ = sessions.Create(ctx, sess)

	if err := uc.Logout(ctx, "sess-1", "user-1"); err != nil {
		t.Fatalf("logout failed: %v", err)
	}

	_, err := sessions.GetSession(ctx, "sess-1")
	if err != authstorage.ErrSessionNotFound {
		t.Fatalf("expected ErrSessionNotFound after logout, got %v", err)
	}
}

func TestAuthLogoutUseCase_RequiresSessionID(t *testing.T) {
	uc := NewLogout(authstorage.NewInMemorySessionRepository())

	if err := uc.Logout(context.Background(), "", "user-1"); err == nil {
		t.Fatal("expected error when session ID is empty")
	}
}

func TestAuthLogoutUseCase_NilStoreIsNoop(t *testing.T) {
	uc := NewLogout(nil)

	if err := uc.Logout(context.Background(), "sess-1", "user-1"); err != nil {
		t.Fatalf("expected no-op without store, got %v", err)
	}
}

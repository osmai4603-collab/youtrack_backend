package authusecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"youtrack_backend/internal/domain"
	authstorage "youtrack_backend/internal/storage/auth"
	storageusers "youtrack_backend/internal/storage/users"
)

func TestAuthRefreshUseCase_ExtendsSession(t *testing.T) {
	ctx := context.Background()
	users := storageusers.NewInMemoryRepository()
	user := testUser("alice", "password123")
	users.Upsert(user)

	sessions := authstorage.NewInMemorySessionRepository()
	uc := NewRefresh(nil, users, sessions, nil, 24*time.Hour)

	now := time.Now().UTC()
	sess := domain.UserSession{
		ID:        "sess-alice",
		SessionID: "sess-alice",
		UserID:    user.ID,
		CreatedAt: now,
		UpdatedAt: now,
		ExpiresAt: now.Add(1 * time.Hour),
	}
	if err := sessions.Create(ctx, sess); err != nil {
		t.Fatalf("create session: %v", err)
	}

	result, err := uc.Refresh(ctx, &RefreshParams{RefreshToken: "sess-alice"})
	if err != nil {
		t.Fatalf("refresh failed: %v", err)
	}
	if result.Session == nil || result.Session.SessionID != "sess-alice" {
		t.Fatal("expected session-alice to be returned")
	}
	if !result.Session.ExpiresAt.After(now.Add(12 * time.Hour)) {
		t.Error("expected session expiry to be extended")
	}
}

func TestAuthRefreshUseCase_RejectsGarbageToken(t *testing.T) {
	sessions := authstorage.NewInMemorySessionRepository()
	uc := NewRefresh(nil, storageusers.NewInMemoryRepository(), sessions, nil, 24*time.Hour)

	_, err := uc.Refresh(context.Background(), &RefreshParams{RefreshToken: "invalid-session-id"})
	if !errors.Is(err, ErrInvalidRefreshToken) {
		t.Errorf("expected ErrInvalidRefreshToken, got %v", err)
	}
}

func TestAuthRefreshUseCase_RejectsDeletedUser(t *testing.T) {
	ctx := context.Background()
	sessions := authstorage.NewInMemorySessionRepository()

	now := time.Now().UTC()
	sess := domain.UserSession{
		ID:        "sess-ghost",
		SessionID: "sess-ghost",
		UserID:    "user-ghost",
		CreatedAt: now,
		UpdatedAt: now,
		ExpiresAt: now.Add(1 * time.Hour),
	}
	_ = sessions.Create(ctx, sess)

	uc := NewRefresh(nil, storageusers.NewInMemoryRepository(), sessions, nil, 24*time.Hour)
	_, err := uc.Refresh(ctx, &RefreshParams{RefreshToken: "sess-ghost"})
	if !errors.Is(err, ErrInvalidRefreshToken) {
		t.Errorf("expected ErrInvalidRefreshToken for deleted user, got %v", err)
	}
}

func TestAuthRefreshUseCase_RejectsBannedUser(t *testing.T) {
	ctx := context.Background()
	users := storageusers.NewInMemoryRepository()
	user := testUser("mallory", "password123")
	user.Banned = true
	users.Upsert(user)

	sessions := authstorage.NewInMemorySessionRepository()
	now := time.Now().UTC()
	sess := domain.UserSession{
		ID:        "sess-mallory",
		SessionID: "sess-mallory",
		UserID:    user.ID,
		CreatedAt: now,
		UpdatedAt: now,
		ExpiresAt: now.Add(1 * time.Hour),
	}
	_ = sessions.Create(ctx, sess)

	uc := NewRefresh(nil, users, sessions, nil, 24*time.Hour)
	_, err := uc.Refresh(ctx, &RefreshParams{RefreshToken: "sess-mallory"})
	if !errors.Is(err, ErrInvalidRefreshToken) {
		t.Errorf("expected ErrInvalidRefreshToken for banned user, got %v", err)
	}
}

func TestAuthRefreshUseCase_RequiresToken(t *testing.T) {
	sessions := authstorage.NewInMemorySessionRepository()
	uc := NewRefresh(nil, storageusers.NewInMemoryRepository(), sessions, nil, 24*time.Hour)

	if _, err := uc.Refresh(context.Background(), nil); !errors.Is(err, ErrInvalidRefreshToken) {
		t.Errorf("expected ErrInvalidRefreshToken for nil params, got %v", err)
	}
	if _, err := uc.Refresh(context.Background(), &RefreshParams{}); !errors.Is(err, ErrInvalidRefreshToken) {
		t.Errorf("expected ErrInvalidRefreshToken for empty token, got %v", err)
	}
}

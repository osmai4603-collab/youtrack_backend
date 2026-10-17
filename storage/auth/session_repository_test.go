package authstorage_test

import (
	"context"
	"testing"
	"time"

	"youtrack_backend/internal/domain"
	authstorage "youtrack_backend/internal/storage/auth"
)

func TestInMemorySessionRepository(t *testing.T) {
	t.Parallel()

	repo := authstorage.NewInMemorySessionRepository()
	ctx := context.Background()

	now := time.Now().UTC()
	session := domain.UserSession{
		ID:        "sess-1",
		SessionID: "sess-1",
		UserID:    "user-100",
		CreatedAt: now,
		UpdatedAt: now,
		ExpiresAt: now.Add(1 * time.Hour),
	}

	// 1. Create
	if err := repo.Create(ctx, session); err != nil {
		t.Fatalf("unexpected error creating session: %v", err)
	}

	// 2. Get
	fetched, err := repo.GetSession(ctx, "sess-1")
	if err != nil {
		t.Fatalf("unexpected error fetching session: %v", err)
	}
	if fetched.UserID != "user-100" {
		t.Errorf("expected user-100, got %s", fetched.UserID)
	}

	// 3. Delete
	if err := repo.Delete(ctx, "sess-1"); err != nil {
		t.Fatalf("unexpected error deleting session: %v", err)
	}
	_, err = repo.GetSession(ctx, "sess-1")
	if err != authstorage.ErrSessionNotFound {
		t.Errorf("expected ErrSessionNotFound, got %v", err)
	}

	// 4. Expiry
	expiredSession := domain.UserSession{
		ID:        "sess-expired",
		SessionID: "sess-expired",
		UserID:    "user-100",
		CreatedAt: now.Add(-2 * time.Hour),
		UpdatedAt: now.Add(-2 * time.Hour),
		ExpiresAt: now.Add(-1 * time.Hour),
	}
	if err := repo.Create(ctx, expiredSession); err != nil {
		t.Fatalf("unexpected error creating expired session: %v", err)
	}
	_, err = repo.GetSession(ctx, "sess-expired")
	if err != authstorage.ErrSessionNotFound {
		t.Errorf("expected ErrSessionNotFound for expired session, got %v", err)
	}

	// 5. PurgeExpired
	if err := repo.PurgeExpired(ctx); err != nil {
		t.Fatalf("unexpected error purging expired sessions: %v", err)
	}
}

package users

import (
	"context"
	"errors"
	"testing"
	"time"

	"youtrack_backend/internal/domain"
)

func TestInMemoryRepository_UpsertAndFind(t *testing.T) {
	t.Parallel()

	repo := NewInMemoryRepository()
	ctx := context.Background()

	if _, err := repo.FindByLogin(ctx, "alice"); !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
	if _, err := repo.FindByID(ctx, "user-1"); !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}

	hash := "hashed-password"
	user := &domain.User{
		ID:              "user-1",
		Login:           "alice",
		Email:           "alice@example.com",
		FullName:        "Alice Liddell",
		Name:            "Alice",
		AvatarURL:       "https://example.com/alice.png",
		UserTypeID:      "standard",
		IsEmailVerified: true,
		Guest:           false,
		Online:          true,
		CanReadProfile:  true,
		RingID:          "ring-alice",
		PasswordHash:    &hash,
	}
	repo.Upsert(user)

	byLogin, err := repo.FindByLogin(ctx, "alice")
	if err != nil {
		t.Fatalf("FindByLogin failed: %v", err)
	}
	if byLogin.ID != "user-1" || byLogin.Login != "alice" {
		t.Errorf("unexpected user: %+v", byLogin)
	}
	if byLogin.FullName != "Alice Liddell" {
		t.Errorf("expected FullName %q, got %q", "Alice Liddell", byLogin.FullName)
	}
	if byLogin.AvatarURL != "https://example.com/alice.png" {
		t.Errorf("expected AvatarURL %q, got %q", "https://example.com/alice.png", byLogin.AvatarURL)
	}
	if byLogin.RingID != "ring-alice" {
		t.Errorf("expected RingID %q, got %q", "ring-alice", byLogin.RingID)
	}
	if !byLogin.IsEmailVerified {
		t.Error("expected IsEmailVerified to be true")
	}

	byID, err := repo.FindByID(ctx, "user-1")
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if byID.Email != "alice@example.com" {
		t.Errorf("expected email to round-trip, got %q", byID.Email)
	}
}

func TestInMemoryRepository_UpdateLoginState(t *testing.T) {
	t.Parallel()

	repo := NewInMemoryRepository()
	ctx := context.Background()
	repo.Upsert(&domain.User{ID: "user-1", Login: "alice"})

	now := time.Now().UTC()
	state := domain.LoginState{
		FailedLoginCount:  3,
		LastFailedLoginAt: &now,
		IsLocked:          true,
	}
	if err := repo.UpdateLoginState(ctx, "user-1", state); err != nil {
		t.Fatalf("UpdateLoginState failed: %v", err)
	}

	user, err := repo.FindByID(ctx, "user-1")
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if user.FailedLoginCount != 3 {
		t.Errorf("expected 3 failed logins, got %d", user.FailedLoginCount)
	}
	if user.LastFailedLoginAt == nil || !user.LastFailedLoginAt.Equal(now) {
		t.Errorf("expected LastFailedLoginAt to be set, got %v", user.LastFailedLoginAt)
	}
	if !user.IsLocked {
		t.Error("expected account to be locked")
	}
}

func TestInMemoryRepository_UpdateLoginStateLiftsLock(t *testing.T) {
	t.Parallel()

	repo := NewInMemoryRepository()
	ctx := context.Background()
	repo.Upsert(&domain.User{ID: "user-1", Login: "alice", IsLocked: true})

	now := time.Now().UTC()
	if err := repo.UpdateLoginState(ctx, "user-1", domain.LoginState{LastLoginAt: &now, FailedLoginCount: 0}); err != nil {
		t.Fatalf("UpdateLoginState failed: %v", err)
	}

	user, _ := repo.FindByID(ctx, "user-1")
	if user.IsLocked {
		t.Error("expected lock to be lifted")
	}
	if user.FailedLoginCount != 0 {
		t.Errorf("expected failed logins reset to 0, got %d", user.FailedLoginCount)
	}
}

func TestFormatOptionalTime(t *testing.T) {
	t.Parallel()

	if got := formatOptionalTime(nil); got != "" {
		t.Fatalf("expected empty string for nil time, got %q", got)
	}

	now := time.Date(2025, time.January, 2, 3, 4, 5, 0, time.UTC)
	if got := formatOptionalTime(&now); got != "2025-01-02 03:04:05" {
		t.Fatalf("expected formatted time, got %q", got)
	}
}

func TestInMemoryRepository_UpdateLoginStateUnknownUser(t *testing.T) {
	t.Parallel()

	repo := NewInMemoryRepository()
	if err := repo.UpdateLoginState(context.Background(), "missing", domain.LoginState{}); !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}
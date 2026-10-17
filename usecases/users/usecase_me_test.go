package usersusecase

import (
	"context"
	"errors"
	"testing"

	"youtrack_backend/internal/domain"
	storageusers "youtrack_backend/internal/storage/users"
)

func meTestUser() *domain.User {
	return &domain.User{
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
		IsLocked:        false,
	}
}

func TestMeUseCase_MeSuccess(t *testing.T) {
	repo := storageusers.NewInMemoryRepository()
	repo.Upsert(meTestUser())

	uc := NewMe(nil, repo)

	result, err := uc.Me(context.Background(), &MeParams{UserID: "user-1"})
	if err != nil {
		t.Fatalf("me failed: %v", err)
	}
	if result.User == nil {
		t.Fatal("expected a user in the result")
	}
	if result.User.Login != "alice" {
		t.Errorf("expected alice, got %q", result.User.Login)
	}
	if result.User.FullName != "Alice Liddell" {
		t.Errorf("expected FullName round-trip, got %q", result.User.FullName)
	}
}

func TestMeUseCase_MeEmptyParams(t *testing.T) {
	uc := NewMe(nil, storageusers.NewInMemoryRepository())

	if _, err := uc.Me(context.Background(), nil); !errors.Is(err, ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound for nil params, got %v", err)
	}
	if _, err := uc.Me(context.Background(), &MeParams{}); !errors.Is(err, ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound for empty params, got %v", err)
	}
}

func TestMeUseCase_MeUserNotFound(t *testing.T) {
	uc := NewMe(nil, storageusers.NewInMemoryRepository())

	if _, err := uc.Me(context.Background(), &MeParams{UserID: "ghost"}); !errors.Is(err, ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestMeUseCase_MeBanned(t *testing.T) {
	repo := storageusers.NewInMemoryRepository()
	user := meTestUser()
	user.Banned = true
	repo.Upsert(user)

	uc := NewMe(nil, repo)

	if _, err := uc.Me(context.Background(), &MeParams{UserID: "user-1"}); !errors.Is(err, ErrUserUnavailable) {
		t.Errorf("expected ErrUserUnavailable for banned user, got %v", err)
	}
}

func TestMeUseCase_MeLocked(t *testing.T) {
	repo := storageusers.NewInMemoryRepository()
	user := meTestUser()
	user.IsLocked = true
	repo.Upsert(user)

	uc := NewMe(nil, repo)

	if _, err := uc.Me(context.Background(), &MeParams{UserID: "user-1"}); !errors.Is(err, ErrUserUnavailable) {
		t.Errorf("expected ErrUserUnavailable for locked user, got %v", err)
	}
}

func TestMeUseCase_MeContextCanceled(t *testing.T) {
	repo := storageusers.NewInMemoryRepository()
	repo.Upsert(meTestUser())

	uc := NewMe(nil, repo)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := uc.Me(ctx, &MeParams{UserID: "user-1"})
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}
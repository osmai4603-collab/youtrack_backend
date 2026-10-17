package authusecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"youtrack_backend/internal/domain"
	authstorage "youtrack_backend/internal/storage/auth"
	storageusers "youtrack_backend/internal/storage/users"

	"golang.org/x/crypto/bcrypt"
)

func testUser(login, password string) *domain.User {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	hashStr := string(hash)
	return &domain.User{
		ID:           "user-" + login,
		Login:        login,
		Email:        login + "@example.com",
		PasswordHash: &hashStr,
	}
}

func TestAuthLoginUseCase_LoginSuccess(t *testing.T) {
	ctx := context.Background()
	users := storageusers.NewInMemoryRepository()
	users.Upsert(testUser("alice", "password123"))

	sessions := authstorage.NewInMemorySessionRepository()
	uc := NewLogin(nil, users, sessions, nil, 24*time.Hour)

	result, err := uc.TryLogin(ctx, &LoginParams{Username: "alice", Password: "password123"})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if result.User.Login != "alice" {
		t.Errorf("expected alice, got %q", result.User.Login)
	}
	if result.Session == nil || result.Session.SessionID == "" {
		t.Fatal("expected a populated session")
	}
	if result.Session.UserID != "user-alice" {
		t.Errorf("expected user-alice, got %q", result.Session.UserID)
	}
	if len(result.Roles) == 0 {
		t.Fatal("expected at least the default role")
	}

	fetched, err := sessions.GetSession(ctx, result.Session.SessionID)
	if err != nil {
		t.Fatalf("fetch session from repo: %v", err)
	}
	if fetched.UserID != "user-alice" {
		t.Errorf("expected session user-alice, got %q", fetched.UserID)
	}
}

func TestAuthLoginUseCase_LoginWrongPassword(t *testing.T) {
	ctx := context.Background()
	users := storageusers.NewInMemoryRepository()
	users.Upsert(testUser("alice", "password123"))

	sessions := authstorage.NewInMemorySessionRepository()
	uc := NewLogin(nil, users, sessions, nil, 24*time.Hour)

	for range maxFailedLogins {
		_, err := uc.TryLogin(ctx, &LoginParams{Username: "alice", Password: "wrong-password"})
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("expected ErrInvalidCredentials, got %v", err)
		}
	}

	alice, err := users.FindByLogin(ctx, "alice")
	if err != nil {
		t.Fatalf("reload user: %v", err)
	}
	if alice.FailedLoginCount != maxFailedLogins {
		t.Errorf("expected %d failed logins, got %d", maxFailedLogins, alice.FailedLoginCount)
	}
	if !alice.IsLocked {
		t.Error("expected account to be locked after successive failures")
	}

	// A correct password now must be rejected while the account is locked.
	_, err = uc.TryLogin(ctx, &LoginParams{Username: "alice", Password: "password123"})
	if !errors.Is(err, ErrAccountInactive) {
		t.Errorf("expected ErrAccountInactive for locked account, got %v", err)
	}
}

func TestAuthLoginUseCase_LoginLockResetsAfterSuccess(t *testing.T) {
	ctx := context.Background()
	users := storageusers.NewInMemoryRepository()
	user := testUser("bob", "password123")
	user.IsLocked = true
	lockedUntil := time.Now().Add(-1 * time.Hour)
	user.LockedUntil = &lockedUntil
	users.Upsert(user)

	sessions := authstorage.NewInMemorySessionRepository()
	uc := NewLogin(nil, users, sessions, nil, 24*time.Hour)

	result, err := uc.TryLogin(ctx, &LoginParams{Username: "bob", Password: "password123"})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if result.User.IsLocked {
		t.Error("expected lock to be lifted after successful login")
	}
}

func TestAuthLoginUseCase_LoginBanned(t *testing.T) {
	ctx := context.Background()
	users := storageusers.NewInMemoryRepository()
	user := testUser("mallory", "password123")
	user.Banned = true
	users.Upsert(user)

	sessions := authstorage.NewInMemorySessionRepository()
	uc := NewLogin(nil, users, sessions, nil, 24*time.Hour)

	_, err := uc.TryLogin(ctx, &LoginParams{Username: "mallory", Password: "password123"})
	if !errors.Is(err, ErrAccountLocked) {
		t.Errorf("expected ErrAccountLocked for banned user, got %v", err)
	}
}

func TestAuthLoginUseCase_LoginUnknownUser(t *testing.T) {
	sessions := authstorage.NewInMemorySessionRepository()
	uc := NewLogin(nil, storageusers.NewInMemoryRepository(), sessions, nil, 24*time.Hour)

	_, err := uc.TryLogin(context.Background(), &LoginParams{Username: "ghost", Password: "password123"})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthLoginUseCase_LoginMissingParams(t *testing.T) {
	sessions := authstorage.NewInMemorySessionRepository()
	uc := NewLogin(nil, storageusers.NewInMemoryRepository(), sessions, nil, 24*time.Hour)

	if _, err := uc.TryLogin(context.Background(), nil); !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials for nil params, got %v", err)
	}
	if _, err := uc.TryLogin(context.Background(), &LoginParams{}); !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials for empty params, got %v", err)
	}
}

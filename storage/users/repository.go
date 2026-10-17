// Package users provides the persistence layer for the authenticated user
// records consumed by the authentication use cases and the JWT middleware.
package users

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"youtrack_backend/internal/domain"
	"youtrack_backend/internal/infrastructure/postgres"
	requestlogger "youtrack_backend/internal/platform/logger"
)

var (
	// ErrUserNotFound is returned when a login or ID does not match any user.
	ErrUserNotFound = errors.New("user not found")
	// ErrUserIDRequired is returned when an empty user ID is used.
	ErrUserIDRequired = errors.New("user ID is required")
)

// Repository defines the read/write contract for authentication-relevant user data.
type Repository interface {
	FindByLogin(ctx context.Context, login string) (*domain.User, error)
	FindByID(ctx context.Context, id string) (*domain.User, error)
	UpdateLoginState(ctx context.Context, userID string, state domain.LoginState) error
}

// ---------------------------------------------------------------------------
// In-memory implementation (tests, memory driver)
// ---------------------------------------------------------------------------

// InMemoryRepository is a simple, thread-safe user store for unit tests and the
// in-memory development driver. Seed users with Upsert before exercising login.
type InMemoryRepository struct {
	mu    sync.RWMutex
	users map[string]*domain.User
	byID  map[string]string // id -> login
}

// NewInMemoryRepository creates an empty in-memory user store.
func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		users: make(map[string]*domain.User),
		byID:  make(map[string]string),
	}
}

// Upsert creates or replaces a user record (test seeding helper).
func (r *InMemoryRepository) Upsert(user *domain.User) {
	r.mu.Lock()
	defer r.mu.Unlock()
	copyUser := *user
	r.users[user.Login] = &copyUser
	r.byID[user.ID] = user.Login
}

// ---------------------------------------------------------------------------
// PostgreSQL implementation
// ---------------------------------------------------------------------------

// PostgresRepository implements Repository against the `users` table.
type PostgresRepository struct {
	db postgres.DBTX
}

// NewRepository returns a Postgres-backed repository, or an in-memory one when
// no database handle is available.
func NewRepository(db postgres.DBTX) Repository {
	if db == nil {
		return NewInMemoryRepository()
	}
	return &PostgresRepository{db: db}
}

const userSelectColumns = `
	id,
	login,
	COALESCE(email, ''),
	COALESCE(full_name, ''),
	COALESCE(name, ''),
	COALESCE(avatar_url, ''),
	COALESCE(type::text, ''),
	false,
	COALESCE(guest, false),
	COALESCE(online, false),
	COALESCE(banned, false),
	'',
	true,
	COALESCE(is_locked, false),
	COALESCE(ring_id, ''),
	COALESCE(hash_password, ''),
	COALESCE(failed_login_count, 0),
	last_failed_login_at,
	last_login_at,
	locked_until,
	COALESCE(created_at, creation_time)
`

func scanUser(row interface{ Scan(...any) error }) (*domain.User, error) {
	var (
		u            domain.User
		passwordHash string
		lastFailed   *time.Time
		lastLogin    *time.Time
		lockedUntil  *time.Time
		createdAt    *time.Time
	)
	if err := row.Scan(
		&u.ID,
		&u.Login,
		&u.Email,
		&u.FullName,
		&u.Name,
		&u.AvatarURL,
		&u.UserTypeID,
		&u.IsEmailVerified,
		&u.Guest,
		&u.Online,
		&u.Banned,
		&u.BanBadge,
		&u.CanReadProfile,
		&u.IsLocked,
		&u.RingID,
		&passwordHash,
		&u.FailedLoginCount,
		&lastFailed,
		&lastLogin,
		&lockedUntil,
		&createdAt,
	); err != nil {
		return nil, err
	}

	u.LastFailedLoginAt = lastFailed
	u.LastLoginAt = lastLogin
	u.LockedUntil = lockedUntil
	if createdAt != nil {
		u.CreatedAt = *createdAt
	}
	if passwordHash != "" {
		u.PasswordHash = &passwordHash
	}
	if u.Name == "" {
		u.Name = u.FullName
	}
	if u.FullName == "" {
		u.FullName = u.Name
	}
	return &u, nil
}

func logUsers(ctx context.Context, op string, users []*domain.User, args ...slog.Attr) {
	if len(users) == 0 {
		requestlogger.FromContext(ctx).LogAttrs(ctx, slog.LevelDebug, op,
			//slog.String("issues", "[]"),
			args...,
		)
		return
	}

	requestlogger.FromContext(ctx).LogAttrs(ctx, slog.LevelDebug, op, args...)
	for index, user := range users {
		requestlogger.FromContext(ctx).LogAttrs(ctx, slog.LevelDebug, "  "+strconv.Itoa((index+1)),
			slog.String("id", user.ID),
			slog.String("name", user.Name),
			slog.String("email", user.Email),
			slog.Bool("guest", user.Guest),
			slog.Bool("isLocked", user.IsLocked),
			slog.Bool("isInactive", user.IsInactive()),
		)
	}
}

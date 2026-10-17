package usersusecase

import (
	"context"
	"errors"
	"log/slog"

	"youtrack_backend/internal/domain"
	storageusers "youtrack_backend/internal/storage/users"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrUserUnavailable = errors.New("user account unavailable")
)

// UserReader abstracts reading the authenticated user record needed to answer /me.
// It is satisfied structurally by storage/users.Repository (FindByID).
type UserReader interface {
	FindByID(ctx context.Context, id string) (*domain.User, error)
}

type MeUseCase struct {
	users  UserReader
	logger *slog.Logger
}

// NewMe builds the "me" use case.
func NewMe(logger *slog.Logger, users UserReader) *MeUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	return &MeUseCase{
		users:  users,
		logger: logger,
	}
}

// MeParams carries the authenticated identity extracted from the request context.
type MeParams struct {
	UserID string
	Roles  []string
}

// MeResult is returned by Me and carries the live user record plus its roles.
type MeResult struct {
	User  *domain.User
	Roles []string
}

// Me loads the current user by ID and validates the account state.
func (uc *MeUseCase) Me(ctx context.Context, params *MeParams) (*MeResult, error) {
	if params == nil || params.UserID == "" {
		return nil, ErrUserNotFound
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	user, err := uc.users.FindByID(ctx, params.UserID)
	if err != nil {
		if errors.Is(err, storageusers.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	if user == nil || user.Banned || user.IsLocked {
		return nil, ErrUserUnavailable
	}

	return &MeResult{User: user, Roles: params.Roles}, nil
}

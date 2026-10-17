package authusecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"youtrack_backend/internal/domain"
	authstorage "youtrack_backend/internal/storage/auth"
)

// AuthRefreshUseCase extends or verifies an active session.
type AuthRefreshUseCase struct {
	users      AuthDependencies
	sessions   authstorage.SessionRepository
	roles      RoleFinder
	logger     *slog.Logger
	sessionTTL time.Duration
}

// NewRefresh builds the refresh use case.
func NewRefresh(logger *slog.Logger, users AuthDependencies, sessions authstorage.SessionRepository, roles RoleFinder, sessionTTL time.Duration) *AuthRefreshUseCase {
	if logger == nil {
		logger = slog.Default()
	}
	if sessionTTL <= 0 {
		sessionTTL = defaultSessionTTL
	}
	return &AuthRefreshUseCase{
		users:      users,
		sessions:   sessions,
		roles:      roles,
		logger:     logger,
		sessionTTL: sessionTTL,
	}
}

// RefreshParams carries the session ID / refresh token supplied to the refresh endpoint.
type RefreshParams struct {
	RefreshToken string
}

// Refresh extends the expiry of the active session and returns updated account state.
func (uc *AuthRefreshUseCase) Refresh(ctx context.Context, params *RefreshParams) (*LoginResult, error) {
	if params == nil || params.RefreshToken == "" {
		return nil, ErrInvalidRefreshToken
	}

	if uc.sessions == nil {
		return nil, ErrInvalidRefreshToken
	}

	session, err := uc.sessions.GetSession(ctx, params.RefreshToken)
	if err != nil || session == nil {
		return nil, ErrInvalidRefreshToken
	}

	var user = &domain.User{ID: session.UserID}
	if uc.users != nil {
		u, err := uc.users.FindByID(ctx, session.UserID)
		if err != nil || u == nil || u.Banned || u.IsLocked {
			return nil, ErrInvalidRefreshToken
		}
		user = u
	}

	roles := resolveRoles(ctx, uc.roles, uc.logger, session.UserID)

	now := time.Now().UTC()
	session.UpdatedAt = now
	session.ExpiresAt = now.Add(uc.sessionTTL)
	if err := uc.sessions.Create(ctx, *session); err != nil {
		return nil, fmt.Errorf("extend session: %w", err)
	}

	return &LoginResult{
		User:    user,
		Session: session,
		Roles:   roles,
	}, nil
}

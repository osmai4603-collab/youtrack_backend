package authusecase

import (
	"context"
	"fmt"

	authstorage "youtrack_backend/internal/storage/auth"
)

// AuthLogoutUseCase deletes active sessions.
type AuthLogoutUseCase struct {
	sessions authstorage.SessionRepository
}

// NewLogout builds the logout use case.
func NewLogout(sessions authstorage.SessionRepository) *AuthLogoutUseCase {
	return &AuthLogoutUseCase{
		sessions: sessions,
	}
}

// Logout deletes the session identified by sessionID.
func (uc *AuthLogoutUseCase) Logout(ctx context.Context, sessionID, userID string) error {
	if sessionID == "" {
		return fmt.Errorf("%w: session id missing", ErrLogoutFailed)
	}
	if uc.sessions == nil {
		return nil
	}
	return uc.sessions.Delete(ctx, sessionID)
}

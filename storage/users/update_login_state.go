package users

import (
	"context"
	"log/slog"
	"time"

	"youtrack_backend/internal/domain"
	"youtrack_backend/internal/infrastructure/postgres"
	requestlogger "youtrack_backend/internal/platform/logger"
)

func (r *InMemoryRepository) UpdateLoginState(ctx context.Context, userID string, state domain.LoginState) error {
	r.mu.Lock()
	login, ok := r.byID[userID]
	if !ok {
		r.mu.Unlock()
		return ErrUserNotFound
	}
	user := r.users[login]
	copyState := state
	if copyState.LastFailedLoginAt != nil {
		t := *copyState.LastFailedLoginAt
		user.LastFailedLoginAt = &t
	}
	if copyState.LastLoginAt != nil {
		t := *copyState.LastLoginAt
		user.LastLoginAt = &t
	}
	if copyState.LockedUntil != nil {
		t := *copyState.LockedUntil
		user.LockedUntil = &t
	}
	user.FailedLoginCount = copyState.FailedLoginCount
	user.IsLocked = copyState.IsLocked
	r.mu.Unlock()
	requestlogger.LogStorageResult(ctx, "users.UpdateLoginState", map[string]any{"updated": true})
	return nil
}

func formatOptionalTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.UTC().Format("2006-01-02 15:04:05")
}

func (r *PostgresRepository) UpdateLoginState(ctx context.Context, userID string, state domain.LoginState) error {
	if userID == "" {
		return ErrUserIDRequired
	}
	query := `
		UPDATE users SET
			failed_login_count  = $2,
			last_failed_login_at = $3,
			last_login_at        = $4,
			is_locked            = $5,
			locked_until         = $6
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query,
		userID,
		state.FailedLoginCount,
		state.LastFailedLoginAt,
		state.LastLoginAt,
		state.IsLocked,
		state.LockedUntil,
	)
	if err != nil {
		return postgres.TranslateError("users.UpdateLoginState", err)
	}
	requestlogger.FromContext(ctx).LogAttrs(ctx, slog.LevelInfo, "UpdateLoginState",
		slog.String("userID", userID),
		slog.Int("failedLoginCount", state.FailedLoginCount),
		slog.String("lastLoginAt", formatOptionalTime(state.LastLoginAt)),
		slog.Bool("isLocked", state.IsLocked),
		slog.String("lockedUntil", formatOptionalTime(state.LockedUntil)),
	)
	return nil
}

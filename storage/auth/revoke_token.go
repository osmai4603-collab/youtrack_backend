package authstorage

import (
	"context"
	"time"

	"youtrack_backend/internal/infrastructure/postgres"
	requestlogger "youtrack_backend/internal/platform/logger"
)

const revokeQuery = `
	INSERT INTO revoked_tokens (token_id, user_id, expires_at)
	VALUES ($1, $2, $3)
	ON CONFLICT (token_id) DO UPDATE SET
		user_id = excluded.user_id,
		expires_at = excluded.expires_at
`

func (r *PostgresTokenDenylist) Revoke(ctx context.Context, tokenID, userID string, expiresAt time.Time) error {
	const op = "authstorage.Revoke"

	if r.db == nil {
		return errNoDatabase
	}
	_, err := r.db.Exec(ctx, revokeQuery, tokenID, userID, expiresAt)
	if err != nil {
		return postgres.TranslateError(op, err)
	}
	requestlogger.LogStorageResult(ctx, op, map[string]any{"revoked": true})
	return nil
}

const revokeAllQuery = `
	INSERT INTO revoked_tokens (token_id, user_id, expires_at)
	VALUES ($1, $2, $3)
	ON CONFLICT (token_id) DO UPDATE SET
		user_id = excluded.user_id,
		expires_at = excluded.expires_at
`

func (r *PostgresTokenDenylist) RevokeAllForUser(ctx context.Context, userID string, expiresAt time.Time) error {
	const op = "authstorage.RevokeAllForUser"

	if r.db == nil {
		return errNoDatabase
	}
	_, err := r.db.Exec(ctx, revokeAllQuery, userMarker(userID), userID, expiresAt)
	if err != nil {
		return postgres.TranslateError(op, err)
	}
	requestlogger.LogStorageResult(ctx, op, map[string]any{"revoked": true})
	return nil
}

package authstorage

import (
	"context"

	"youtrack_backend/internal/infrastructure/postgres"
	requestlogger "youtrack_backend/internal/platform/logger"
)

const purgeQuery = `DELETE FROM revoked_tokens WHERE expires_at <= now()`

func (r *PostgresTokenDenylist) PurgeExpired(ctx context.Context) error {
	const op = "authstorage.PurgeExpired"

	if r.db == nil {
		return errNoDatabase
	}
	_, err := r.db.Exec(ctx, purgeQuery)
	if err != nil {
		return postgres.TranslateError(op, err)
	}
	requestlogger.LogStorageResult(ctx, op, map[string]any{"purged": true})
	return nil
}

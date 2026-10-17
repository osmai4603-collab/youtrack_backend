package authstorage

import (
	"context"

	"youtrack_backend/internal/infrastructure/postgres"
	requestlogger "youtrack_backend/internal/platform/logger"
)

const purgeSessionsQuery = `DELETE FROM sessions WHERE expires_at <= now()`

func (r *PostgresSessionRepository) PurgeExpired(ctx context.Context) error {
	const op = "authstorage.PurgeExpiredSessions"
	if r.db == nil {
		return errNoSessionDatabase
	}
	_, err := r.db.Exec(ctx, purgeSessionsQuery)
	if err != nil {
		return postgres.TranslateError(op, err)
	}
	requestlogger.LogStorageResult(ctx, op, map[string]any{"purged": true})
	return nil
}

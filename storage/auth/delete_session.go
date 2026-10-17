package authstorage

import (
	"context"

	"youtrack_backend/internal/infrastructure/postgres"
	requestlogger "youtrack_backend/internal/platform/logger"
)

const deleteSessionQuery = `DELETE FROM sessions WHERE session_id = $1`

func (r *PostgresSessionRepository) Delete(ctx context.Context, sessionID string) error {
	const op = "authstorage.DeleteSession"
	if r.db == nil {
		return errNoSessionDatabase
	}
	_, err := r.db.Exec(ctx, deleteSessionQuery, sessionID)
	if err != nil {
		return postgres.TranslateError(op, err)
	}
	requestlogger.LogStorageResult(ctx, op+"(sessionID: "+sessionID+")", map[string]any{"deleted": true})
	return nil
}

const deleteSessionsByUserIDQuery = `DELETE FROM sessions WHERE user_id = $1`

func (r *PostgresSessionRepository) DeleteByUserID(ctx context.Context, userID string) error {
	const op = "authstorage.DeleteSessionsByUserID"
	if r.db == nil {
		return errNoSessionDatabase
	}
	_, err := r.db.Exec(ctx, deleteSessionsByUserIDQuery, userID)
	if err != nil {
		return postgres.TranslateError(op, err)
	}
	requestlogger.LogStorageResult(ctx, op+"(userID: "+userID+")", map[string]any{"deleted": true})
	return nil
}

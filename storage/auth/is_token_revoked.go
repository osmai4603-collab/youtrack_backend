package authstorage

import (
	"context"

	"youtrack_backend/internal/infrastructure/postgres"
	requestlogger "youtrack_backend/internal/platform/logger"
)

const isRevokedQuery = `
	SELECT EXISTS (
		SELECT 1 FROM revoked_tokens
		WHERE (token_id = $1 OR token_id = $2)
		  AND expires_at > now()
	)
`

func (r *PostgresTokenDenylist) IsRevoked(ctx context.Context, tokenID, userID string) (bool, error) {
	const op = "authstorage.IsRevoked"

	if r.db == nil {
		return false, errNoDatabase
	}
	var revoked bool
	marker := ""
	if userID != "" {
		marker = userMarker(userID)
	}
	err := r.db.QueryRow(ctx, isRevokedQuery, tokenID, marker).Scan(&revoked)
	if err != nil {
		return false, postgres.TranslateError(op, err)
	}
	requestlogger.LogStorageResult(ctx, op, revoked)
	return revoked, nil
}

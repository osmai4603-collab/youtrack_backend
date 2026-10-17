package authstorage

import (
	"context"

	"youtrack_backend/internal/domain"
	"youtrack_backend/internal/infrastructure/postgres"
)

const createSessionQuery = `
	INSERT INTO sessions (id, session_id, data, user_id, created_at, updated_at, expires_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	ON CONFLICT (session_id) DO UPDATE SET
		data = excluded.data,
		updated_at = excluded.updated_at,
		expires_at = excluded.expires_at
`

func (r *PostgresSessionRepository) Create(ctx context.Context, session domain.UserSession) error {
	const op = "authstorage.CreateSession"
	if r.db == nil {
		return errNoSessionDatabase
	}
	id := session.ID
	if id == "" {
		id = session.SessionID
	}
	_, err := r.db.Exec(ctx, createSessionQuery,
		id, session.SessionID, session.Data, session.UserID,
		session.CreatedAt, session.UpdatedAt, session.ExpiresAt,
	)
	if err != nil {
		return postgres.TranslateError(op, err)
	}
	logSessions(ctx, op, []*domain.UserSession{&session})
	return nil
}

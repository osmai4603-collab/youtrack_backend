package authstorage

import (
	"context"
	"errors"
	"log/slog"

	"youtrack_backend/internal/domain"
	"youtrack_backend/internal/infrastructure/postgres"

	"github.com/jackc/pgx/v5"
)

const getSessionQuery = `
	SELECT id, session_id, COALESCE(data, ''), user_id, created_at, updated_at, expires_at
	FROM sessions
	WHERE session_id = $1 AND expires_at > now()
`

func (r *PostgresSessionRepository) GetSession(ctx context.Context, sessionID string) (*domain.UserSession, error) {
	const op = "GetSession"
	if r.db == nil {
		return nil, errNoSessionDatabase
	}
	var s domain.UserSession
	err := r.db.QueryRow(ctx, getSessionQuery, sessionID).Scan(
		&s.ID, &s.SessionID, &s.Data, &s.UserID,
		&s.CreatedAt, &s.UpdatedAt, &s.ExpiresAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSessionNotFound
		}
		return nil, postgres.TranslateError(op, err)
	}
	logSessions(ctx, op, []*domain.UserSession{&s}, slog.String("sessionID", sessionID))
	return &s, nil
}

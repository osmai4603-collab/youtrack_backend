package authstorage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"youtrack_backend/internal/infrastructure/postgres"
	requestlogger "youtrack_backend/internal/platform/logger"
	"youtrack_backend/internal/services/rbac"

	"github.com/jackc/pgx/v5"
)

// PostgresRBACSessionAdapter implements rbac.RBACSessionRepository backed by PostgreSQL.
type PostgresRBACSessionAdapter struct {
	db postgres.DBTX
}

// NewPostgresRBACSessionAdapter constructs a new session repository adapter for the RBAC engine.
func NewPostgresRBACSessionAdapter(db postgres.DBTX) *PostgresRBACSessionAdapter {
	return &PostgresRBACSessionAdapter{db: db}
}

// Ensure interface compliance.
var _ rbac.RBACSessionRepository = (*PostgresRBACSessionAdapter)(nil)

type rbacSessionPayload struct {
	ActiveRoles map[string]time.Time `json:"active_roles"`
	Metadata    map[string]string    `json:"metadata,omitempty"`
}

const saveRBACSessionQuery = `
	INSERT INTO sessions (id, session_id, data, user_id, created_at, updated_at, expires_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	ON CONFLICT (session_id) DO UPDATE SET
		data = excluded.data,
		updated_at = excluded.updated_at,
		expires_at = excluded.expires_at
`

// SaveSession persists an RBAC session (insert or update).
func (a *PostgresRBACSessionAdapter) SaveSession(ctx context.Context, session *rbac.Session) error {
	const op = "rbac.PostgresSessionAdapter.SaveSession"
	if session == nil || session.ID == "" || session.UserID == "" {
		return fmt.Errorf("%w: session ID and user ID are required", rbac.ErrInvalidInput)
	}
	if a.db == nil {
		return fmt.Errorf("%w: db is nil", rbac.ErrInvalidInput)
	}

	payload := rbacSessionPayload{
		ActiveRoles: session.ActiveRoles,
		Metadata:    session.Metadata,
	}
	dataBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal session data: %w", err)
	}

	now := time.Now().UTC()
	createdAt := session.CreatedAt
	if createdAt.IsZero() {
		createdAt = now
	}
	expiresAt := session.ExpiresAt
	if expiresAt.IsZero() {
		expiresAt = now.Add(24 * time.Hour)
	}

	_, err = a.db.Exec(ctx, saveRBACSessionQuery,
		session.ID, session.ID, string(dataBytes), session.UserID,
		createdAt, now, expiresAt,
	)
	if err != nil {
		return postgres.TranslateError(op, err)
	}

	requestlogger.LogStorageResult(ctx, op+"(sessionID: "+session.ID+")", "{userID: "+session.UserID+"}")
	return nil
}

const getRBACSessionQuery = `
	SELECT session_id, COALESCE(data, ''), user_id, created_at, expires_at
	FROM sessions
	WHERE session_id = $1 AND expires_at > now()
`

// GetSession retrieves an active session by ID.
func (a *PostgresRBACSessionAdapter) GetSession(ctx context.Context, sessionID string) (*rbac.Session, error) {
	const op = "rbac.PostgresSessionAdapter.GetSession"
	if sessionID == "" {
		return nil, fmt.Errorf("%w: session ID is required", rbac.ErrInvalidInput)
	}
	if a.db == nil {
		return nil, fmt.Errorf("%w: db is nil", rbac.ErrInvalidInput)
	}

	var sess rbac.Session
	var rawData string
	err := a.db.QueryRow(ctx, getRBACSessionQuery, sessionID).Scan(
		&sess.ID, &rawData, &sess.UserID, &sess.CreatedAt, &sess.ExpiresAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: %s", rbac.ErrSessionNotFound, sessionID)
		}
		return nil, postgres.TranslateError(op, err)
	}

	sess.ActiveRoles = make(map[string]time.Time)
	sess.Metadata = make(map[string]string)

	if rawData != "" {
		var payload rbacSessionPayload
		if err := json.Unmarshal([]byte(rawData), &payload); err == nil {
			if payload.ActiveRoles != nil {
				sess.ActiveRoles = payload.ActiveRoles
			}
			if payload.Metadata != nil {
				sess.Metadata = payload.Metadata
			}
		}
	}

	return &sess, nil
}

// DeleteSession removes a session by ID.
func (a *PostgresRBACSessionAdapter) DeleteSession(ctx context.Context, sessionID string) error {
	const op = "rbac.PostgresSessionAdapter.DeleteSession"
	if sessionID == "" {
		return fmt.Errorf("%w: session ID is required", rbac.ErrInvalidInput)
	}
	if a.db == nil {
		return fmt.Errorf("%w: db is nil", rbac.ErrInvalidInput)
	}

	_, err := a.db.Exec(ctx, "DELETE FROM sessions WHERE session_id = $1", sessionID)
	if err != nil {
		return postgres.TranslateError(op, err)
	}
	return nil
}

const listRBACSessionsByUserQuery = `
	SELECT session_id, COALESCE(data, ''), user_id, created_at, expires_at
	FROM sessions
	WHERE user_id = $1 AND expires_at > now()
	ORDER BY created_at DESC
`

// ListSessionsByUser lists all active sessions for a user.
func (a *PostgresRBACSessionAdapter) ListSessionsByUser(ctx context.Context, userID string) ([]*rbac.Session, error) {
	const op = "rbac.PostgresSessionAdapter.ListSessionsByUser"
	if userID == "" {
		return nil, fmt.Errorf("%w: user ID is required", rbac.ErrInvalidInput)
	}
	if a.db == nil {
		return nil, fmt.Errorf("%w: db is nil", rbac.ErrInvalidInput)
	}

	rows, err := a.db.Query(ctx, listRBACSessionsByUserQuery, userID)
	if err != nil {
		return nil, postgres.TranslateError(op, err)
	}
	defer rows.Close()

	var sessions []*rbac.Session
	for rows.Next() {
		var sess rbac.Session
		var rawData string
		if err := rows.Scan(&sess.ID, &rawData, &sess.UserID, &sess.CreatedAt, &sess.ExpiresAt); err != nil {
			return nil, postgres.TranslateError(op, err)
		}
		sess.ActiveRoles = make(map[string]time.Time)
		sess.Metadata = make(map[string]string)
		if rawData != "" {
			var payload rbacSessionPayload
			if err := json.Unmarshal([]byte(rawData), &payload); err == nil {
				if payload.ActiveRoles != nil {
					sess.ActiveRoles = payload.ActiveRoles
				}
				if payload.Metadata != nil {
					sess.Metadata = payload.Metadata
				}
			}
		}
		sessions = append(sessions, &sess)
	}
	if err := rows.Err(); err != nil {
		return nil, postgres.TranslateError(op, err)
	}

	return sessions, nil
}

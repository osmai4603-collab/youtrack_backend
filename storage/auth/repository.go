package authstorage

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"time"

	"youtrack_backend/internal/domain"
	"youtrack_backend/internal/infrastructure/postgres"
	requestlogger "youtrack_backend/internal/platform/logger"
)

// ErrSessionNotFound is returned when a requested session is missing or expired.
var ErrSessionNotFound = errors.New("session not found")

// SessionRepository defines storage operations for server-side user sessions.
type SessionRepository interface {
	Create(ctx context.Context, session domain.UserSession) error
	GetSession(ctx context.Context, sessionID string) (*domain.UserSession, error)
	Delete(ctx context.Context, sessionID string) error
	DeleteByUserID(ctx context.Context, userID string) error
	PurgeExpired(ctx context.Context) error
}

type PostgresSessionRepository struct {
	db postgres.DBTX
}

func NewSessionRepository(db postgres.DBTX) SessionRepository {
	if db == nil {
		return NewInMemorySessionRepository()
	}
	return &PostgresSessionRepository{db: db}
}

var errNoSessionDatabase = errors.New("session repository has no database handle")

func logSessions(ctx context.Context, op string, sessions []*domain.UserSession, args ...slog.Attr) {
	if len(sessions) == 0 {
		requestlogger.FromContext(ctx).LogAttrs(ctx, slog.LevelDebug, op,
			//slog.String("issues", "[]"),
			args...,
		)
		return
	}

	requestlogger.FromContext(ctx).LogAttrs(ctx, slog.LevelDebug, op, args...)
	for index, ss := range sessions {
		requestlogger.FromContext(ctx).LogAttrs(ctx, slog.LevelDebug, "  "+strconv.Itoa((index+1)),
			slog.String("id", ss.ID),
			slog.String("userID", ss.UserID),
			slog.String("expireAT", formatOptionalTime(&ss.ExpiresAt)),
		)
	}
}

func formatOptionalTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.UTC().Format("2006-01-02 15:04:05")
}

package authstorage

import (
	"context"
	"errors"
	"sync"
	"time"
	"youtrack_backend/internal/domain"
	requestlogger "youtrack_backend/internal/platform/logger"
)

type InMemorySessionRepository struct {
	mu       sync.RWMutex
	sessions map[string]domain.UserSession
}

// NewInMemorySessionRepository creates an in-memory session store.
func NewInMemorySessionRepository() *InMemorySessionRepository {
	return &InMemorySessionRepository{
		sessions: make(map[string]domain.UserSession),
	}
}

func (r *InMemorySessionRepository) Create(ctx context.Context, session domain.UserSession) error {
	if session.SessionID == "" {
		return errors.New("session ID is required")
	}
	r.mu.Lock()
	r.sessions[session.SessionID] = session
	r.mu.Unlock()
	requestlogger.LogStorageResult(ctx, "authstorage.CreateSession", session)
	return nil
}

func (r *InMemorySessionRepository) GetSession(ctx context.Context, sessionID string) (*domain.UserSession, error) {
	if sessionID == "" {
		return nil, ErrSessionNotFound
	}
	r.mu.RLock()
	s, ok := r.sessions[sessionID]
	r.mu.RUnlock()
	if !ok {
		return nil, ErrSessionNotFound
	}
	if !s.ExpiresAt.IsZero() && !s.ExpiresAt.After(time.Now().UTC()) {
		return nil, ErrSessionNotFound
	}
	requestlogger.LogStorageResult(ctx, "authstorage.GetSession", s)
	return &s, nil
}

func (r *InMemorySessionRepository) Delete(ctx context.Context, sessionID string) error {
	r.mu.Lock()
	delete(r.sessions, sessionID)
	r.mu.Unlock()
	requestlogger.LogStorageResult(ctx, "authstorage.DeleteSession", map[string]any{"deleted": true})
	return nil
}

func (r *InMemorySessionRepository) DeleteByUserID(ctx context.Context, userID string) error {
	r.mu.Lock()
	for id, s := range r.sessions {
		if s.UserID == userID {
			delete(r.sessions, id)
		}
	}
	r.mu.Unlock()
	requestlogger.LogStorageResult(ctx, "authstorage.DeleteSessionsByUserID", map[string]any{"deleted": true})
	return nil
}

func (r *InMemorySessionRepository) PurgeExpired(ctx context.Context) error {
	now := time.Now().UTC()
	r.mu.Lock()
	for id, s := range r.sessions {
		if !s.ExpiresAt.IsZero() && !s.ExpiresAt.After(now) {
			delete(r.sessions, id)
		}
	}
	r.mu.Unlock()
	requestlogger.LogStorageResult(ctx, "authstorage.PurgeExpiredSessions", map[string]any{"purged": true})
	return nil
}

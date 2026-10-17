package rbac

import (
	"context"
	"sync"
)

// inMemoryRBACSessionRepository is a thread-safe in-memory implementation
// of RBACSessionRepository. Used as the default when no external repository is provided.
type inMemoryRBACSessionRepository struct {
	mu       sync.RWMutex
	sessions map[string]*Session
}

func newInMemoryRBACSessionRepository() *inMemoryRBACSessionRepository {
	return &inMemoryRBACSessionRepository{
		sessions: make(map[string]*Session),
	}
}

func (r *inMemoryRBACSessionRepository) SaveSession(_ context.Context, session *Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	// Store a deep copy to prevent external mutation from affecting the store.
	r.sessions[session.ID] = cloneSession(session)
	return nil
}

func (r *inMemoryRBACSessionRepository) GetSession(_ context.Context, sessionID string) (*Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	sess, ok := r.sessions[sessionID]
	if !ok {
		return nil, ErrSessionNotFound
	}
	// Return a deep copy so callers can safely mutate before calling SaveSession.
	return cloneSession(sess), nil
}

func (r *inMemoryRBACSessionRepository) DeleteSession(_ context.Context, sessionID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sessions, sessionID)
	return nil
}

func (r *inMemoryRBACSessionRepository) ListSessionsByUser(_ context.Context, userID string) ([]*Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*Session
	for _, s := range r.sessions {
		if s.UserID == userID {
			result = append(result, cloneSession(s))
		}
	}
	return result, nil
}

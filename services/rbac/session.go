package rbac

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sort"
	"time"
)

// RBACSessionRepository abstracts persistent storage for RBAC sessions.
// Implementations may use PostgreSQL, Redis, or any other backing store.
type RBACSessionRepository interface {
	// SaveSession persists a session (insert or update).
	SaveSession(ctx context.Context, session *Session) error

	// GetSession retrieves a session by its ID. Returns ErrSessionNotFound if absent.
	GetSession(ctx context.Context, sessionID string) (*Session, error)

	// DeleteSession removes a session by its ID.
	DeleteSession(ctx context.Context, sessionID string) error

	// ListSessionsByUser returns all stored sessions for the given user.
	ListSessionsByUser(ctx context.Context, userID string) ([]*Session, error)
}

// SessionManager oversees runtime security contexts, dynamic role activation,
// JIT expiration enforcement, and Dynamic Separation of Duties (DSD).
type SessionManager struct {
	repo   RBACSessionRepository
	graph  *RoleGraph
	sodMgr *SoDManager
}

// NewSessionManager creates a new SessionManager.
// If repo is nil, an in-memory store is used automatically.
func NewSessionManager(repo RBACSessionRepository) *SessionManager {
	if repo == nil {
		repo = newInMemoryRBACSessionRepository()
	}
	return &SessionManager{
		repo: repo,
	}
}

// SetGraph binds a RoleGraph for permission resolution in active sessions.
func (sm *SessionManager) SetGraph(graph *RoleGraph) {
	sm.graph = graph
}

// SetSoDManager binds a SoDManager for Dynamic Separation of Duties enforcement.
func (sm *SessionManager) SetSoDManager(sodMgr *SoDManager) {
	sm.sodMgr = sodMgr
}

// CreateSession initializes a new session for the specified user with an optional initial role set and TTL.
func (sm *SessionManager) CreateSession(userID string, initialRoles []RoleID, ttl time.Duration) (*Session, error) {
	if userID == "" {
		return nil, fmt.Errorf("%w: user ID must not be empty", ErrInvalidInput)
	}
	if ttl <= 0 {
		ttl = 24 * time.Hour // Default 24h session window
	}

	now := time.Now()
	expiresAt := now.Add(ttl)

	// Pre-validate initial roles against DSD
	if sm.sodMgr != nil && len(initialRoles) > 1 {
		var validated []RoleID
		for _, r := range initialRoles {
			if err := sm.sodMgr.ValidateSessionActivation(validated, r); err != nil {
				return nil, fmt.Errorf("%w: initial roles breach DSD: %v", err, initialRoles)
			}
			validated = append(validated, r)
		}
	}

	sessionID := generateSessionID()

	activeRoles := make(map[RoleID]time.Time, len(initialRoles))
	for _, r := range initialRoles {
		activeRoles[r] = expiresAt
	}

	sess := &Session{
		ID:          sessionID,
		UserID:      userID,
		ActiveRoles: activeRoles,
		CreatedAt:   now,
		ExpiresAt:   expiresAt,
		Metadata:    make(map[string]string),
	}

	ctx := context.Background()
	if err := sm.repo.SaveSession(ctx, sess); err != nil {
		return nil, fmt.Errorf("rbac: failed to persist session: %w", err)
	}

	return cloneSession(sess), nil
}

// GetSession retrieves the session by ID, verifying validity and purging expired JIT roles.
func (sm *SessionManager) GetSession(sessionID string) (*Session, error) {
	ctx := context.Background()

	sess, err := sm.repo.GetSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrSessionNotFound, sessionID)
	}

	now := time.Now()
	if !sess.IsActive(now) {
		_ = sm.repo.DeleteSession(ctx, sessionID)
		return nil, fmt.Errorf("%w: session %s has expired", ErrSessionExpired, sessionID)
	}

	// Purge expired active roles
	if sm.purgeExpiredRoles(sess, now) {
		_ = sm.repo.SaveSession(ctx, sess)
	}

	return cloneSession(sess), nil
}

// ActivateRole activates a role inside an active session.
// If ttl > 0, the role is activated with a temporary JIT lease.
// It strictly validates DSD constraints before activation.
func (sm *SessionManager) ActivateRole(sessionID string, role RoleID, ttl time.Duration) error {
	if role == "" {
		return fmt.Errorf("%w: role ID must not be empty", ErrInvalidInput)
	}

	ctx := context.Background()

	sess, err := sm.repo.GetSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrSessionNotFound, sessionID)
	}

	now := time.Now()
	if !sess.IsActive(now) {
		_ = sm.repo.DeleteSession(ctx, sessionID)
		return fmt.Errorf("%w: session %s has expired", ErrSessionExpired, sessionID)
	}

	// Purge expired roles first to reflect current active state
	sm.purgeExpiredRoles(sess, now)

	// Collect currently active roles
	activeList := make([]RoleID, 0, len(sess.ActiveRoles))
	for r := range sess.ActiveRoles {
		activeList = append(activeList, r)
	}

	// Dynamic Separation of Duties (DSD) verification
	if sm.sodMgr != nil {
		if err := sm.sodMgr.ValidateSessionActivation(activeList, role); err != nil {
			return err
		}
	}

	// Determine role expiry
	roleExpiry := sess.ExpiresAt
	if ttl > 0 {
		candidate := now.Add(ttl)
		if candidate.Before(sess.ExpiresAt) {
			roleExpiry = candidate
		}
	}

	sess.ActiveRoles[role] = roleExpiry

	if err := sm.repo.SaveSession(ctx, sess); err != nil {
		return fmt.Errorf("rbac: failed to persist session after role activation: %w", err)
	}
	return nil
}

// DeactivateRole removes a role from the active roles of a session.
func (sm *SessionManager) DeactivateRole(sessionID string, role RoleID) error {
	ctx := context.Background()

	sess, err := sm.repo.GetSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrSessionNotFound, sessionID)
	}

	delete(sess.ActiveRoles, role)

	if err := sm.repo.SaveSession(ctx, sess); err != nil {
		return fmt.Errorf("rbac: failed to persist session after role deactivation: %w", err)
	}
	return nil
}

// TerminateSession destroys an active session.
func (sm *SessionManager) TerminateSession(sessionID string) error {
	ctx := context.Background()

	// Verify session exists before deleting
	if _, err := sm.repo.GetSession(ctx, sessionID); err != nil {
		return fmt.Errorf("%w: %s", ErrSessionNotFound, sessionID)
	}

	if err := sm.repo.DeleteSession(ctx, sessionID); err != nil {
		return fmt.Errorf("rbac: failed to delete session: %w", err)
	}
	return nil
}

// GetActiveRoles returns the list of non-expired roles active in the session.
func (sm *SessionManager) GetActiveRoles(sessionID string) ([]RoleID, error) {
	ctx := context.Background()

	sess, err := sm.repo.GetSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrSessionNotFound, sessionID)
	}

	now := time.Now()
	if !sess.IsActive(now) {
		_ = sm.repo.DeleteSession(ctx, sessionID)
		return nil, fmt.Errorf("%w: session %s has expired", ErrSessionExpired, sessionID)
	}

	if sm.purgeExpiredRoles(sess, now) {
		_ = sm.repo.SaveSession(ctx, sess)
	}

	roles := make([]RoleID, 0, len(sess.ActiveRoles))
	for r := range sess.ActiveRoles {
		roles = append(roles, r)
	}
	sort.Strings(roles)
	return roles, nil
}

// GetActivePermissions aggregates the explicitly granted permissions for all
// non-expired roles currently active in the session.
func (sm *SessionManager) GetActivePermissions(sessionID string) ([]string, error) {
	activeRoles, err := sm.GetActiveRoles(sessionID)
	if err != nil {
		return nil, err
	}

	permSet := make(map[string]struct{})
	for _, r := range activeRoles {
		if sm.graph != nil {
			effective, err := sm.graph.GetEffectivePermissions(r)
			if err == nil {
				for _, p := range effective {
					permSet[p] = struct{}{}
				}
			}
		}
	}

	result := make([]string, 0, len(permSet))
	for p := range permSet {
		result = append(result, p)
	}
	sort.Strings(result)
	return result, nil
}

// ListUserSessions returns all active sessions for a given user.
func (sm *SessionManager) ListUserSessions(userID string) []*Session {
	ctx := context.Background()

	sessions, err := sm.repo.ListSessionsByUser(ctx, userID)
	if err != nil {
		return nil
	}

	now := time.Now()
	var res []*Session
	for _, s := range sessions {
		if s.IsActive(now) {
			sm.purgeExpiredRoles(s, now)
			res = append(res, cloneSession(s))
		} else {
			_ = sm.repo.DeleteSession(ctx, s.ID)
		}
	}
	return res
}

// purgeExpiredRoles removes roles whose activation TTL has passed.
// Returns true if any roles were purged (indicating the session was modified).
func (sm *SessionManager) purgeExpiredRoles(s *Session, now time.Time) bool {
	changed := false
	for r, expiry := range s.ActiveRoles {
		if !expiry.IsZero() && now.After(expiry) {
			delete(s.ActiveRoles, r)
			changed = true
		}
	}
	return changed
}

// generateSessionID generates a cryptographically random session token.
func generateSessionID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return "sess_" + hex.EncodeToString(b)
}

// cloneSession creates a deep copy of a Session.
func cloneSession(s *Session) *Session {
	if s == nil {
		return nil
	}
	cp := *s
	if s.ActiveRoles != nil {
		cp.ActiveRoles = make(map[RoleID]time.Time, len(s.ActiveRoles))
		for k, v := range s.ActiveRoles {
			cp.ActiveRoles[k] = v
		}
	}
	if s.Metadata != nil {
		cp.Metadata = make(map[string]string, len(s.Metadata))
		for k, v := range s.Metadata {
			cp.Metadata[k] = v
		}
	}
	return &cp
}

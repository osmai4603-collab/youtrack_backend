package authstorage

import (
	"context"
	"errors"
	"sync"
	"time"

	"youtrack_backend/internal/infrastructure/postgres"
	requestlogger "youtrack_backend/internal/platform/logger"
)

// TokenDenylist provides JWT revocation support without server-side sessions.
// Individual token pairs are revoked by their shared TokenID; a user-wide
// revocation marker invalidates every outstanding pair of a user at once.
//
// Implementations are fail-closed: an unreachable backing store must surface an
// error rather than answer "not revoked", because a permissive answer re-admits
// every revoked token.
type TokenDenylist interface {
	// Revoke invalidates a single token pair identified by tokenID until expiresAt.
	Revoke(ctx context.Context, tokenID, userID string, expiresAt time.Time) error
	// RevokeAllForUser invalidates every outstanding token pair of a user until expiresAt.
	RevokeAllForUser(ctx context.Context, userID string, expiresAt time.Time) error
	// IsRevoked reports whether the token pair (or the user-wide marker) is currently revoked.
	IsRevoked(ctx context.Context, tokenID, userID string) (bool, error)
	// PurgeExpired removes expired denylist entries.
	PurgeExpired(ctx context.Context) error
}

const userMarkerPrefix = "user:"

func userMarker(userID string) string {
	return userMarkerPrefix + userID
}

// ---------------------------------------------------------------------------
// In-memory implementation (tests, memory driver)
// ---------------------------------------------------------------------------

type inMemoryEntry struct {
	expiresAt time.Time
}

// InMemoryTokenDenylist is a process-local, mutex-guarded revocation list.
type InMemoryTokenDenylist struct {
	mu     sync.RWMutex
	tokens map[string]inMemoryEntry
}

// NewInMemoryTokenDenylist creates an empty, thread-safe in-memory denylist.
func NewInMemoryTokenDenylist() *InMemoryTokenDenylist {
	return &InMemoryTokenDenylist{tokens: make(map[string]inMemoryEntry)}
}

func (r *InMemoryTokenDenylist) Revoke(ctx context.Context, tokenID, _ string, expiresAt time.Time) error {
	if tokenID == "" {
		return errors.New("token ID is required for revocation")
	}
	r.mu.Lock()
	r.tokens[tokenID] = inMemoryEntry{expiresAt: expiresAt}
	r.mu.Unlock()
	requestlogger.LogStorageResult(ctx, "authstorage.Revoke", map[string]any{"revoked": true})
	return nil
}

func (r *InMemoryTokenDenylist) RevokeAllForUser(ctx context.Context, userID string, expiresAt time.Time) error {
	if userID == "" {
		return errors.New("user ID is required for user-wide revocation")
	}
	r.mu.Lock()
	r.tokens[userMarker(userID)] = inMemoryEntry{expiresAt: expiresAt}
	r.mu.Unlock()
	requestlogger.LogStorageResult(ctx, "authstorage.RevokeAllForUser", map[string]any{"revoked": true})
	return nil
}

func (r *InMemoryTokenDenylist) IsRevoked(ctx context.Context, tokenID, userID string) (bool, error) {
	now := time.Now().UTC()
	r.mu.Lock()

	if entry, ok := r.tokens[tokenID]; ok {
		if entry.expiresAt.After(now) {
			r.mu.Unlock()
			requestlogger.LogStorageResult(ctx, "authstorage.IsRevoked", true)
			return true, nil
		}
		delete(r.tokens, tokenID)
	}

	if userID != "" {
		marker := userMarker(userID)
		if entry, ok := r.tokens[marker]; ok {
			if entry.expiresAt.After(now) {
				r.mu.Unlock()
				requestlogger.LogStorageResult(ctx, "authstorage.IsRevoked", true)
				return true, nil
			}
			delete(r.tokens, marker)
		}
	}

	r.mu.Unlock()
	requestlogger.LogStorageResult(ctx, "authstorage.IsRevoked", false)
	return false, nil
}

func (r *InMemoryTokenDenylist) PurgeExpired(ctx context.Context) error {
	now := time.Now().UTC()
	r.mu.Lock()

	for id, entry := range r.tokens {
		if !entry.expiresAt.After(now) {
			delete(r.tokens, id)
		}
	}
	r.mu.Unlock()
	requestlogger.LogStorageResult(ctx, "authstorage.PurgeExpired", map[string]any{"purged": true})
	return nil
}

// ---------------------------------------------------------------------------
// PostgreSQL-backed implementation
// ---------------------------------------------------------------------------

// PostgresTokenDenylist persists revoked tokens in the `revoked_tokens` table.
//
// The implementation is strictly fail-closed: a database failure is returned to
// the caller and is never downgraded to an in-memory decision. Reporting a token
// as "not revoked" because the store was unreachable would silently re-admit
// every revoked token for the duration of the outage.
type PostgresTokenDenylist struct {
	db postgres.DBTX
}

// NewTokenDenylist returns a Postgres-backed denylist, or a pure in-memory one
// when no database handle is available (memory driver, unit tests).
func NewTokenDenylist(db postgres.DBTX) TokenDenylist {
	if db == nil {
		return NewInMemoryTokenDenylist()
	}
	return &PostgresTokenDenylist{db: db}
}

// errNoDatabase guards against a PostgresTokenDenylist built without a handle.
// NewTokenDenylist never produces one, but failing loudly is safer than a nil
// interface panic and far safer than a permissive answer.
var errNoDatabase = errors.New("token denylist has no database handle")

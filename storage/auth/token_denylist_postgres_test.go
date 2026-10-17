package authstorage

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	platErrors "youtrack_backend/internal/platform/errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var errDBDown = errors.New("connection refused")

// failingDB is a postgres.DBTX that fails every operation, standing in for an
// unreachable database.
type failingDB struct{ err error }

func (f failingDB) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, f.err
}

func (f failingDB) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, f.err
}

func (f failingDB) QueryRow(context.Context, string, ...any) pgx.Row {
	return failingRow{err: f.err}
}

type failingRow struct{ err error }

func (r failingRow) Scan(...any) error { return r.err }

// scriptedDB records the statements it is given and replays canned results.
type scriptedDB struct {
	execTag pgconn.CommandTag
	execErr error
	scanErr error
	scanVal bool

	statements []string
}

func (s *scriptedDB) Exec(_ context.Context, sql string, _ ...any) (pgconn.CommandTag, error) {
	s.statements = append(s.statements, sql)
	return s.execTag, s.execErr
}

func (s *scriptedDB) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, s.execErr
}

func (s *scriptedDB) QueryRow(_ context.Context, sql string, _ ...any) pgx.Row {
	s.statements = append(s.statements, sql)
	return scriptedRow{err: s.scanErr, val: s.scanVal}
}

type scriptedRow struct {
	err error
	val bool
}

func (r scriptedRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	if len(dest) > 0 {
		if p, ok := dest[0].(*bool); ok {
			*p = r.val
		}
	}
	return nil
}

// IsRevoked is the security-critical read: a database outage must not be
// reported as "not revoked", because that silently re-admits every revoked
// token for the duration of the outage. This is the F-02 regression guard.
func TestPostgresTokenDenylist_IsRevokedPropagatesDBError(t *testing.T) {
	t.Parallel()

	list := NewTokenDenylist(failingDB{err: errDBDown})

	revoked, err := list.IsRevoked(context.Background(), "token-1", "user-1")
	if err == nil {
		t.Fatal("expected an error when the denylist store is unreachable, got nil")
	}
	if revoked {
		t.Fatal("expected revoked=false alongside the error, got true")
	}
	if !errors.Is(err, errDBDown) {
		t.Fatalf("expected the driver error to be wrapped, got %v", err)
	}
}

// The same must hold with no user ID, where the marker argument is empty.
func TestPostgresTokenDenylist_IsRevokedPropagatesDBErrorWithoutUser(t *testing.T) {
	t.Parallel()

	list := NewTokenDenylist(failingDB{err: errDBDown})

	if _, err := list.IsRevoked(context.Background(), "token-1", ""); err == nil {
		t.Fatal("expected an error when the denylist store is unreachable, got nil")
	}
}

func TestPostgresTokenDenylist_RevokePropagatesDBError(t *testing.T) {
	t.Parallel()

	list := NewTokenDenylist(failingDB{err: errDBDown})

	err := list.Revoke(context.Background(), "token-1", "user-1", time.Now().Add(time.Hour))
	if err == nil {
		t.Fatal("expected Revoke to report the store failure, got nil")
	}
	if !errors.Is(err, errDBDown) {
		t.Fatalf("expected the driver error to be wrapped, got %v", err)
	}
}

// A Revoke that silently succeeded would let logout report success while the
// token stays valid, so the failure has to reach the caller.
func TestPostgresTokenDenylist_RevokeAllForUserPropagatesDBError(t *testing.T) {
	t.Parallel()

	list := NewTokenDenylist(failingDB{err: errDBDown})

	if err := list.RevokeAllForUser(context.Background(), "user-1", time.Now().Add(time.Hour)); err == nil {
		t.Fatal("expected RevokeAllForUser to report the store failure, got nil")
	}
}

func TestPostgresTokenDenylist_PurgeExpiredPropagatesDBError(t *testing.T) {
	t.Parallel()

	list := NewTokenDenylist(failingDB{err: errDBDown})

	if err := list.PurgeExpired(context.Background()); err == nil {
		t.Fatal("expected PurgeExpired to report the store failure, got nil")
	}
}

func TestPostgresTokenDenylist_RevokeSucceedsWithoutError(t *testing.T) {
	t.Parallel()

	db := &scriptedDB{execTag: pgconn.CommandTag{}}
	list := NewTokenDenylist(db)

	if err := list.Revoke(context.Background(), "token-1", "user-1", time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("expected no error on a healthy store, got %v", err)
	}
	if len(db.statements) != 1 {
		t.Fatalf("expected exactly one statement, got %d", len(db.statements))
	}
	if want := "INSERT INTO revoked_tokens"; !strings.Contains(db.statements[0], want) {
		t.Fatalf("expected an insert into revoked_tokens, got %q", db.statements[0])
	}
}

// A user-wide revocation must be stored under the shared marker so a single row
// invalidates every outstanding pair of that user.
func TestPostgresTokenDenylist_RevokeAllForUserUsesUserMarker(t *testing.T) {
	t.Parallel()

	db := &scriptedDB{execTag: pgconn.CommandTag{}}
	record := &capturingDB{scriptedDB: db, args: make([][]any, 0)}

	list := NewTokenDenylist(record)
	if err := list.RevokeAllForUser(context.Background(), "user-9", time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(record.args) != 1 || len(record.args[0]) < 1 {
		t.Fatalf("expected one call with arguments, got %+v", record.args)
	}
	if got := record.args[0][0]; got != userMarker("user-9") {
		t.Fatalf("expected the user marker as token id, got %v", got)
	}
}

type capturingDB struct {
	*scriptedDB
	args [][]any
}

func (c *capturingDB) Exec(_ context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	c.statements = append(c.statements, sql)
	c.args = append(c.args, args)
	return c.execTag, c.execErr
}

func TestPostgresTokenDenylist_IsRevokedReportsTrueFromStore(t *testing.T) {
	t.Parallel()

	db := &scriptedDB{scanVal: true}
	list := NewTokenDenylist(db)

	revoked, err := list.IsRevoked(context.Background(), "token-1", "user-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !revoked {
		t.Fatal("expected revoked=true when the store reports a hit")
	}
}

func TestPostgresTokenDenylist_IsRevokedReportsFalseFromStore(t *testing.T) {
	t.Parallel()

	db := &scriptedDB{scanVal: false}
	list := NewTokenDenylist(db)

	revoked, err := list.IsRevoked(context.Background(), "token-1", "user-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if revoked {
		t.Fatal("expected revoked=false when the store reports no hit")
	}
}

// A nil handle must not panic and must not fabricate a permissive answer.
func TestPostgresTokenDenylist_WithoutDBHandleFailsClosed(t *testing.T) {
	t.Parallel()

	list := &PostgresTokenDenylist{}
	ctx := context.Background()

	if _, err := list.IsRevoked(ctx, "token-1", "user-1"); !errors.Is(err, errNoDatabase) {
		t.Fatalf("expected errNoDatabase from IsRevoked, got %v", err)
	}
	if err := list.Revoke(ctx, "token-1", "user-1", time.Now()); !errors.Is(err, errNoDatabase) {
		t.Fatalf("expected errNoDatabase from Revoke, got %v", err)
	}
	if err := list.RevokeAllForUser(ctx, "user-1", time.Now()); !errors.Is(err, errNoDatabase) {
		t.Fatalf("expected errNoDatabase from RevokeAllForUser, got %v", err)
	}
	if err := list.PurgeExpired(ctx); !errors.Is(err, errNoDatabase) {
		t.Fatalf("expected errNoDatabase from PurgeExpired, got %v", err)
	}
}

// The memory driver keeps working untouched: NewTokenDenylist(nil) must not
// regress into an error-returning store.
func TestNewTokenDenylist_NilDBReturnsInMemory(t *testing.T) {
	t.Parallel()

	list := NewTokenDenylist(nil)
	ctx := context.Background()

	if _, ok := list.(*InMemoryTokenDenylist); !ok {
		t.Fatalf("expected an in-memory denylist, got %T", list)
	}
	if err := list.Revoke(ctx, "token-1", "user-1", time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("expected the in-memory revoke to succeed, got %v", err)
	}
	revoked, err := list.IsRevoked(ctx, "token-1", "user-1")
	if err != nil || !revoked {
		t.Fatalf("expected the in-memory revocation to stick, got revoked=%v err=%v", revoked, err)
	}
}

// TranslateError must classify the outage as a typed platform error rather than
// letting the raw driver error escape, so callers can tell a blip from a bug.
func TestPostgresTokenDenylist_DatabaseErrorsAreTranslated(t *testing.T) {
	t.Parallel()

	list := NewTokenDenylist(failingDB{err: errDBDown})

	_, err := list.IsRevoked(context.Background(), "token-1", "user-1")
	if err == nil {
		t.Fatal("expected an error")
	}
	if got := platErrors.ErrorCode(err); got == "" {
		t.Fatalf("expected a translated platform error code, got %q", got)
	}
}

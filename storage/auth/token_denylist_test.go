package authstorage

import (
	"context"
	"testing"
	"time"
)

func TestInMemoryTokenDenylist_RevokeAndCheck(t *testing.T) {
	t.Parallel()

	list := NewInMemoryTokenDenylist()
	ctx := context.Background()

	if revoked, err := list.IsRevoked(ctx, "token-1", "user-1"); err != nil || revoked {
		t.Fatalf("expected not revoked, got revoked=%v err=%v", revoked, err)
	}

	if err := list.Revoke(ctx, "token-1", "user-1", time.Now().Add(1*time.Hour)); err != nil {
		t.Fatalf("failed to revoke: %v", err)
	}

	if revoked, _ := list.IsRevoked(ctx, "token-1", "user-1"); !revoked {
		t.Fatal("expected token to be revoked")
	}
	// A different token of the same user must stay valid.
	if revoked, _ := list.IsRevoked(ctx, "token-2", "user-1"); revoked {
		t.Fatal("expected unrelated token to remain valid")
	}
}

func TestInMemoryTokenDenylist_ExpiredEntriesAreIgnored(t *testing.T) {
	t.Parallel()

	list := NewInMemoryTokenDenylist()
	ctx := context.Background()

	if err := list.Revoke(ctx, "token-expired", "user-1", time.Now().Add(-1*time.Second)); err != nil {
		t.Fatalf("failed to revoke: %v", err)
	}

	if revoked, _ := list.IsRevoked(ctx, "token-expired", "user-1"); revoked {
		t.Fatal("expected expired revocation entry to be ignored")
	}
}

func TestInMemoryTokenDenylist_RevokeAllForUser(t *testing.T) {
	t.Parallel()

	list := NewInMemoryTokenDenylist()
	ctx := context.Background()

	if err := list.RevokeAllForUser(ctx, "user-1", time.Now().Add(1*time.Hour)); err != nil {
		t.Fatalf("failed to revoke all: %v", err)
	}

	if revoked, _ := list.IsRevoked(ctx, "any-old-token", "user-1"); !revoked {
		t.Fatal("expected user-wide marker to revoke all tokens")
	}
	if revoked, _ := list.IsRevoked(ctx, "any-old-token", "user-2"); revoked {
		t.Fatal("expected other user's tokens to remain valid")
	}
}

func TestInMemoryTokenDenylist_UserMarkerExpiry(t *testing.T) {
	t.Parallel()

	list := NewInMemoryTokenDenylist()
	ctx := context.Background()

	if err := list.RevokeAllForUser(ctx, "user-1", time.Now().Add(-1*time.Second)); err != nil {
		t.Fatalf("failed to revoke all: %v", err)
	}

	if revoked, _ := list.IsRevoked(ctx, "any-old-token", "user-1"); revoked {
		t.Fatal("expected expired user-wide marker to be ignored")
	}
}

func TestInMemoryTokenDenylist_PurgeExpired(t *testing.T) {
	t.Parallel()

	list := NewInMemoryTokenDenylist()
	ctx := context.Background()

	_ = list.Revoke(ctx, "token-old", "user-1", time.Now().Add(-1*time.Second))
	_ = list.Revoke(ctx, "token-new", "user-1", time.Now().Add(1*time.Hour))

	if err := list.PurgeExpired(ctx); err != nil {
		t.Fatalf("failed to purge: %v", err)
	}

	if revoked, _ := list.IsRevoked(ctx, "token-old", "user-1"); revoked {
		t.Fatal("expected purged token to be gone")
	}
	if revoked, _ := list.IsRevoked(ctx, "token-new", "user-1"); !revoked {
		t.Fatal("expected active token to remain revoked")
	}
}

func TestInMemoryTokenDenylist_RequiresTokenID(t *testing.T) {
	t.Parallel()

	list := NewInMemoryTokenDenylist()
	if err := list.Revoke(context.Background(), "", "", time.Now()); err == nil {
		t.Fatal("expected error when token ID is empty")
	}
	if err := list.RevokeAllForUser(context.Background(), "", time.Now()); err == nil {
		t.Fatal("expected error when user ID is empty")
	}
}
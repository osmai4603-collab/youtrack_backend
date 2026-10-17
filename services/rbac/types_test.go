package rbac_test

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"youtrack_backend/internal/services/rbac"
)

func TestRoleSerialization(t *testing.T) {
	role := rbac.Role{
		ID:          "SECURITY_LEAD",
		Name:        "Security Lead",
		Description: "Lead of information security",
		Permissions: []string{"READ_LOGS", "UPDATE_POLICIES"},
		IsImmutable: false,
		Metadata: map[string]string{
			"department": "SecOps",
		},
	}

	data, err := json.Marshal(role)
	if err != nil {
		t.Fatalf("failed to marshal role: %v", err)
	}

	var unmarshaled rbac.Role
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal role: %v", err)
	}

	if unmarshaled.ID != role.ID {
		t.Errorf("expected ID %s, got %s", role.ID, unmarshaled.ID)
	}
	if len(unmarshaled.Permissions) != 2 || unmarshaled.Permissions[0] != "READ_LOGS" {
		t.Errorf("expected explicitly granted permissions, got %v", unmarshaled.Permissions)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		t.Fatalf("failed to inspect serialized role: %v", err)
	}
	if _, exists := fields["parents"]; exists {
		t.Error("role serialization must not expose a parent relationship")
	}
}

func TestUserAssignment_IsExpired(t *testing.T) {
	now := time.Now()

	// Indefinite assignment (ExpiresAt is nil)
	indefinite := rbac.UserAssignment{
		UserID:    "user_1",
		RoleIDs:   []string{"DEVELOPER"},
		ExpiresAt: nil,
	}
	if indefinite.IsExpired(now.Add(24 * time.Hour)) {
		t.Errorf("indefinite assignment should never expire")
	}

	// Active JIT assignment
	future := now.Add(1 * time.Hour)
	activeJIT := rbac.UserAssignment{
		UserID:    "user_2",
		RoleIDs:   []string{"TEMP_ADMIN"},
		ExpiresAt: &future,
	}
	if activeJIT.IsExpired(now) {
		t.Errorf("active JIT assignment should not be expired before expiry time")
	}

	// Expired JIT assignment
	past := now.Add(-10 * time.Minute)
	expiredJIT := rbac.UserAssignment{
		UserID:    "user_3",
		RoleIDs:   []string{"TEMP_ADMIN"},
		ExpiresAt: &past,
	}
	if !expiredJIT.IsExpired(now) {
		t.Errorf("expired JIT assignment should be detected as expired")
	}
}

func TestSession_IsActive(t *testing.T) {
	now := time.Now()

	activeSession := rbac.Session{
		ID:        "sess_123",
		UserID:    "alice",
		CreatedAt: now.Add(-1 * time.Hour),
		ExpiresAt: now.Add(1 * time.Hour),
	}
	if !activeSession.IsActive(now) {
		t.Errorf("session should be active")
	}

	expiredSession := rbac.Session{
		ID:        "sess_456",
		UserID:    "bob",
		CreatedAt: now.Add(-2 * time.Hour),
		ExpiresAt: now.Add(-1 * time.Hour),
	}
	if expiredSession.IsActive(now) {
		t.Errorf("session should be expired")
	}
}

func TestSentinelErrors(t *testing.T) {
	errList := []error{
		rbac.ErrRoleNotFound,
		rbac.ErrRoleAlreadyExists,
		rbac.ErrRoleImmutable,
		rbac.ErrSSDViolation,
		rbac.ErrDSDViolation,
		rbac.ErrSessionNotFound,
		rbac.ErrSessionExpired,
		rbac.ErrRoleExpired,
		rbac.ErrAccessDenied,
		rbac.ErrInvalidInput,
		rbac.ErrConstraintNotFound,
		rbac.ErrConstraintAlreadyExists,
	}

	for _, err := range errList {
		if err == nil || err.Error() == "" {
			t.Errorf("expected non-empty error message for sentinel error")
		}
		if !errors.Is(err, err) {
			t.Errorf("expected errors.Is to match sentinel error %v", err)
		}
	}
}

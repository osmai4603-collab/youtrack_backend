package domain_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"youtrack_backend/internal/domain"
)

func TestUser_DisplayName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		user     domain.User
		expected string
	}{
		{
			name: "returns FullName when set",
			user: domain.User{
				Login:    "jsmith",
				Name:     "John",
				FullName: "John Smith",
			},
			expected: "John Smith",
		},
		{
			name: "returns Name when FullName is empty",
			user: domain.User{
				Login:    "jsmith",
				Name:     "John",
				FullName: "",
			},
			expected: "John",
		},
		{
			name: "returns Login when Name and FullName are empty",
			user: domain.User{
				Login:    "jsmith",
				Name:     "",
				FullName: "",
			},
			expected: "jsmith",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.user.DisplayName(); got != tt.expected {
				t.Errorf("DisplayName() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestUser_IsInactive(t *testing.T) {
	t.Parallel()

	activeUser := domain.User{Banned: false}
	if activeUser.IsInactive() {
		t.Error("expected activeUser.IsInactive() to be false")
	}

	bannedUser := domain.User{Banned: true}
	if !bannedUser.IsInactive() {
		t.Error("expected bannedUser.IsInactive() to be true")
	}
}

func TestUser_JSONOmitSensitiveFields(t *testing.T) {
	t.Parallel()

	pwd := "super-secret-hash"
	now := time.Now().UTC()
	user := domain.User{
		ID:                "user-1",
		Login:             "alice",
		Email:             "alice@example.com",
		FullName:          "Alice Wonderland",
		Name:              "Alice",
		AvatarURL:         "https://example.com/avatar.png",
		UserTypeID:        "standard",
		IsEmailVerified:   true,
		Guest:             false,
		Online:            true,
		Banned:            false,
		BanBadge:          "",
		CanReadProfile:    true,
		IsLocked:          false,
		RingID:            "ring-abc-123",
		PasswordHash:      &pwd,
		FailedLoginCount:  5,
		LastFailedLoginAt: &now,
		LastLoginAt:       &now,
		LockedUntil:       &now,
		CreatedAt:         now,
	}

	bytes, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("failed to marshal user: %v", err)
	}

	jsonStr := string(bytes)
	if strings.Contains(jsonStr, "password_hash") || strings.Contains(jsonStr, pwd) {
		t.Errorf("PasswordHash must not be exposed in JSON: %s", jsonStr)
	}
	if strings.Contains(jsonStr, "failed_login_count") {
		t.Errorf("FailedLoginCount must not be exposed in JSON: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"full_name":"Alice Wonderland"`) {
		t.Errorf("expected full_name in JSON: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"ring_id":"ring-abc-123"`) {
		t.Errorf("expected ring_id in JSON: %s", jsonStr)
	}
	if !strings.Contains(jsonStr, `"is_email_verified":true`) {
		t.Errorf("expected is_email_verified in JSON: %s", jsonStr)
	}
}

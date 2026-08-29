package model

import (
	"testing"
)

func TestDefaultUserProfiles(t *testing.T) {
	p := DefaultUserProfiles()
	if p == nil {
		t.Fatalf("expected non-nil default user profiles")
	}

	if p.Type != "UserProfiles" {
		t.Errorf("expected Type 'UserProfiles', got %s", p.Type)
	}

	// General
	if p.General == nil || p.General.Type != "GeneralUserProfile" {
		t.Errorf("expected GeneralUserProfile")
	}
	if p.General.Locale.ID != "en_US" {
		t.Errorf("expected locale en_US, got %s", p.General.Locale.ID)
	}

	// Appearance
	if p.Appearance == nil || p.Appearance.Type != "AppearanceUserProfile" {
		t.Errorf("expected AppearanceUserProfile")
	}
	if p.Appearance.FirstDayOfWeek != 0 {
		t.Errorf("expected FirstDayOfWeek 0, got %d", p.Appearance.FirstDayOfWeek)
	}

	// Tips
	if p.Tips == nil || p.Tips.Type != "TipsUserProfile" {
		t.Errorf("expected TipsUserProfile")
	}
	if p.Tips.OnboardingTourState != "idle" {
		t.Errorf("expected OnboardingTourState idle, got %s", p.Tips.OnboardingTourState)
	}

	// AI
	if p.AI == nil || p.AI.Type != "AiUserProfile" {
		t.Errorf("expected AiUserProfile")
	}
	if p.AI.ChatMode != "docked" {
		t.Errorf("expected ChatMode docked, got %s", p.AI.ChatMode)
	}

	// Notifications
	if p.Notifications == nil || p.Notifications.Type != "NotificationsUserProfile" {
		t.Errorf("expected NotificationsUserProfile")
	}
	if !p.Notifications.EmailNotificationsEnabled {
		t.Errorf("expected EmailNotificationsEnabled true")
	}
}

func TestUserNormalizeMe(t *testing.T) {
	u := &User{
		ID:    "1-1",
		Login: "admin",
	}
	u.NormalizeMe()

	if u.Type != "Me" {
		t.Errorf("expected Type 'Me', got %s", u.Type)
	}
}

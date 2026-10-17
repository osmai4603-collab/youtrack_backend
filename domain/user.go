package domain

import "time"

// User represents a user entity mapped directly from the `users` table in the database.
type User struct {
	ID                string     `json:"id" db:"id"`
	Login             string     `json:"login" db:"login"`
	Email             string     `json:"email,omitempty" db:"email"`
	FullName          string     `json:"full_name,omitempty" db:"full_name"`
	Name              string     `json:"name,omitempty" db:"name"`
	AvatarURL         string     `json:"avatar_url,omitempty" db:"avatar_url"`
	UserTypeID        string     `json:"user_type_id,omitempty" db:"user_type_id"`
	IsEmailVerified   bool       `json:"is_email_verified" db:"is_email_verified"`
	Guest             bool       `json:"guest" db:"guest"`
	Online            bool       `json:"online" db:"online"`
	Banned            bool       `json:"banned" db:"banned"`
	BanBadge          string     `json:"ban_badge,omitempty" db:"ban_badge"`
	CanReadProfile    bool       `json:"can_read_profile" db:"can_read_profile"`
	IsLocked          bool       `json:"is_locked" db:"is_locked"`
	RingID            string     `json:"ring_id,omitempty" db:"ring_id"`
	PasswordHash      *string    `json:"-" db:"password_hash"`
	FailedLoginCount  int        `json:"-" db:"failed_login_count"`
	LastFailedLoginAt *time.Time `json:"last_failed_login_at,omitempty" db:"last_failed_login_at"`
	LastLoginAt       *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
	LockedUntil       *time.Time `json:"locked_until,omitempty" db:"locked_until"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
}

// DisplayName returns FullName if available, otherwise Name if available, otherwise Login.
func (u *User) DisplayName() string {
	if u.FullName != "" {
		return u.FullName
	}
	if u.Name != "" {
		return u.Name
	}
	return u.Login
}

func (u *User) IsInactive() bool {
	return u.Banned
}

type LoginState struct {
	FailedLoginCount  int
	LastFailedLoginAt *time.Time
	LastLoginAt       *time.Time
	LockedUntil       *time.Time
	IsLocked          bool
}

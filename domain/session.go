package domain

import "time"

// UserSession represents an active user login session.
type UserSession struct {
	ID        string    `json:"id"`
	SessionID string    `json:"session_id"`
	Data      string    `json:"data,omitempty"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

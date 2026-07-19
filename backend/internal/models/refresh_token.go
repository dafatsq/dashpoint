package models

import (
	"time"

	"github.com/google/uuid"
)

// RefreshToken represents a refresh token in the database.
type RefreshToken struct {
	ID            uuid.UUID  `json:"id"`
	UserID        uuid.UUID  `json:"user_id"`
	TokenHash     string     `json:"-"`
	ExpiresAt     time.Time  `json:"expires_at"`
	CreatedAt     time.Time  `json:"created_at"`
	RevokedAt     *time.Time `json:"revoked_at,omitempty"`
	RevokedReason *string    `json:"revoked_reason,omitempty"`
}

// IsRevoked returns true if the refresh token has been revoked.
func (rt *RefreshToken) IsRevoked() bool {
	return rt.RevokedAt != nil
}

// IsExpired returns true if the refresh token has expired.
func (rt *RefreshToken) IsExpired() bool {
	return time.Now().After(rt.ExpiresAt)
}

// IsValid returns true if the token is not revoked and not expired.
func (rt *RefreshToken) IsValid() bool {
	return !rt.IsRevoked() && !rt.IsExpired()
}

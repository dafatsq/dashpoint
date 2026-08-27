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
	// FamilyID groups all tokens minted from the same login (and their
	// rotation descendants) so replaying a revoked token can revoke the
	// whole family. Server-side only; never serialized to clients.
	FamilyID uuid.UUID `json:"-"`
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

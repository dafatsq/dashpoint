package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID           uuid.UUID  `json:"id"`
	Email        *string    `json:"email,omitempty"`
	Name         string     `json:"name"`
	PasswordHash *string    `json:"-"`
	PINHash      *string    `json:"-"`
	RoleID       uuid.UUID  `json:"role_id"`
	IsActive     bool       `json:"is_active"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`

	// Joined fields
	Role *Role `json:"role,omitempty"`
}

// Role represents a role in the system
type Role struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Permission represents a permission in the system
type Permission struct {
	ID          uuid.UUID `json:"id"`
	Key         string    `json:"key"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	Category    string    `json:"category"`
	CreatedAt   time.Time `json:"created_at"`
}

// UserPermission represents a per-user permission override
type UserPermission struct {
	UserID       uuid.UUID  `json:"user_id"`
	PermissionID uuid.UUID  `json:"permission_id"`
	Allowed      bool       `json:"allowed"`
	GrantedBy    *uuid.UUID `json:"granted_by,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`

	// Joined fields
	Permission *Permission `json:"permission,omitempty"`
}

// RefreshToken represents a refresh token in the database
type RefreshToken struct {
	ID            uuid.UUID  `json:"id"`
	UserID        uuid.UUID  `json:"user_id"`
	TokenHash     string     `json:"-"`
	ExpiresAt     time.Time  `json:"expires_at"`
	CreatedAt     time.Time  `json:"created_at"`
	RevokedAt     *time.Time `json:"revoked_at,omitempty"`
	RevokedReason *string    `json:"revoked_reason,omitempty"`
}

// IsRevoked returns true if the refresh token has been revoked
func (rt *RefreshToken) IsRevoked() bool {
	return rt.RevokedAt != nil
}

// IsExpired returns true if the refresh token has expired
func (rt *RefreshToken) IsExpired() bool {
	return time.Now().After(rt.ExpiresAt)
}

// IsValid returns true if the token is not revoked and not expired
func (rt *RefreshToken) IsValid() bool {
	return !rt.IsRevoked() && !rt.IsExpired()
}

// UserWithPermissions represents a user with their effective permissions
type UserWithPermissions struct {
	User        *User    `json:"user"`
	Permissions []string `json:"permissions"` // List of permission keys
}

// HasPermission checks if the user has a specific permission
func (uwp *UserWithPermissions) HasPermission(permissionKey string) bool {
	for _, p := range uwp.Permissions {
		if p == permissionKey {
			return true
		}
	}
	return false
}

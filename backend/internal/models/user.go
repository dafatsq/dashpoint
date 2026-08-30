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
	TokenVersion int        `json:"-"`
	RoleID       uuid.UUID  `json:"role_id"`
	IsActive     bool       `json:"is_active"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`

	// Joined fields
	Role *Role `json:"role,omitempty"`
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

package models

import (
	"time"

	"github.com/google/uuid"
)

// Role represents a role in the system.
type Role struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Permission represents a permission in the system.
type Permission struct {
	ID          uuid.UUID `json:"id"`
	Key         string    `json:"key"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	Category    string    `json:"category"`
	CreatedAt   time.Time `json:"created_at"`
}

// UserPermission represents a per-user permission override.
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

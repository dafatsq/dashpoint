package models

import (
	"time"

	"github.com/google/uuid"
)

// AuditLog represents an audit log entry.
type AuditLog struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`

	UserID   *uuid.UUID `json:"user_id,omitempty"`
	UserName *string    `json:"user_name,omitempty"`
	UserRole *string    `json:"user_role,omitempty"`

	Action      AuditAction     `json:"action"`
	EntityType  AuditEntityType `json:"entity_type"`
	EntityID    *string         `json:"entity_id,omitempty"`
	Description *string         `json:"description,omitempty"`

	OldValues map[string]interface{} `json:"old_values,omitempty"`
	NewValues map[string]interface{} `json:"new_values,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`

	Status AuditStatus `json:"status"`
}

// AuditLogEntry is a builder for creating audit log entries.
type AuditLogEntry struct {
	UserID      *uuid.UUID
	UserName    string
	UserRole    string
	Action      AuditAction
	EntityType  AuditEntityType
	EntityID    string
	Description string
	OldValues   map[string]interface{}
	NewValues   map[string]interface{}
	Metadata    map[string]interface{}
	Status      AuditStatus
}

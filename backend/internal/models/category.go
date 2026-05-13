package models

import (
	"time"

	"github.com/google/uuid"
)

// Category represents a product category
type Category struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
	SortOrder   int        `json:"sort_order"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	Parent   *Category   `json:"parent,omitempty"`
	Children []*Category `json:"children,omitempty"`
}

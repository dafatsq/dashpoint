package models

import (
	"time"

	"github.com/google/uuid"
)

// ExpenseCategory represents a category for classifying expenses.
type ExpenseCategory struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

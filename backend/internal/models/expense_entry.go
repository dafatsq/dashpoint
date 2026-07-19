package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Expense represents an operating expense.
type Expense struct {
	ID               uuid.UUID        `json:"id"`
	CategoryID       *uuid.UUID       `json:"category_id,omitempty"`
	CategoryName     *string          `json:"category_name,omitempty"`
	ProductID        *uuid.UUID       `json:"product_id,omitempty"`
	ProductName      *string          `json:"product_name,omitempty"`
	Quantity         *decimal.Decimal `json:"quantity,omitempty"`
	AppliesInventory bool             `json:"applies_inventory"`
	Amount           decimal.Decimal  `json:"amount"`
	Description      string           `json:"description"`
	ExpenseDate      time.Time        `json:"expense_date"`
	Vendor           *string          `json:"vendor,omitempty"`
	ReferenceNumber  *string          `json:"reference_number,omitempty"`
	Notes            *string          `json:"notes,omitempty"`
	CreatedBy        uuid.UUID        `json:"created_by"`
	CreatedByName    *string          `json:"created_by_name,omitempty"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
}

package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ExpenseCategory represents a category for classifying expenses
type ExpenseCategory struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Expense represents an operating expense
type Expense struct {
	ID              uuid.UUID        `json:"id"`
	CategoryID      *uuid.UUID       `json:"category_id,omitempty"`
	CategoryName    *string          `json:"category_name,omitempty"`
	ProductID       *uuid.UUID       `json:"product_id,omitempty"`
	ProductName     *string          `json:"product_name,omitempty"`
	Quantity        *decimal.Decimal `json:"quantity,omitempty"`
	Amount          decimal.Decimal  `json:"amount"`
	Description     string           `json:"description"`
	ExpenseDate     time.Time        `json:"expense_date"`
	Vendor          *string          `json:"vendor,omitempty"`
	ReferenceNumber *string          `json:"reference_number,omitempty"`
	Notes           *string          `json:"notes,omitempty"`
	CreatedBy       uuid.UUID        `json:"created_by"`
	CreatedByName   *string          `json:"created_by_name,omitempty"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
}

// ExpenseSummary represents aggregated expense data
type ExpenseSummary struct {
	TotalAmount  decimal.Decimal          `json:"total_amount"`
	ExpenseCount int                      `json:"expense_count"`
	ByCategory   []CategoryExpenseSummary `json:"by_category"`
}

// CategoryExpenseSummary represents expense totals by category
type CategoryExpenseSummary struct {
	CategoryID   *uuid.UUID      `json:"category_id,omitempty"`
	CategoryName string          `json:"category_name"`
	TotalAmount  decimal.Decimal `json:"total_amount"`
	Count        int             `json:"count"`
}

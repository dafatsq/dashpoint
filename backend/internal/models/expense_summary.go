package models

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ExpenseSummary represents aggregated expense data.
type ExpenseSummary struct {
	TotalAmount  decimal.Decimal          `json:"total_amount"`
	ExpenseCount int                      `json:"expense_count"`
	ByCategory   []CategoryExpenseSummary `json:"by_category"`
}

// CategoryExpenseSummary represents expense totals by category.
type CategoryExpenseSummary struct {
	CategoryID   *uuid.UUID      `json:"category_id,omitempty"`
	CategoryName string          `json:"category_name"`
	TotalAmount  decimal.Decimal `json:"total_amount"`
	Count        int             `json:"count"`
}

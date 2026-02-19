package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// CashDrawerOpType represents the type of cash drawer operation
type CashDrawerOpType string

const (
	CashDrawerOpPayIn  CashDrawerOpType = "pay_in"
	CashDrawerOpPayOut CashDrawerOpType = "pay_out"
)

// CashDrawerOperation represents a pay-in or pay-out operation on a shift's cash drawer
type CashDrawerOperation struct {
	ID          uuid.UUID        `json:"id"`
	ShiftID     uuid.UUID        `json:"shift_id"`
	Type        CashDrawerOpType `json:"type"`
	Amount      decimal.Decimal  `json:"amount"`
	Reason      string           `json:"reason"`
	PerformedBy uuid.UUID        `json:"performed_by"`
	CreatedAt   time.Time        `json:"created_at"`

	// Joined fields
	PerformedByName *string `json:"performed_by_name,omitempty"`
}

package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ShiftStatus represents the status of a shift.
type ShiftStatus string

const (
	ShiftStatusOpen      ShiftStatus = "open"
	ShiftStatusClosed    ShiftStatus = "closed"
	ShiftStatusSuspended ShiftStatus = "suspended"
)

// Shift represents a cashier shift.
type Shift struct {
	ID        uuid.UUID   `json:"id"`
	OpenedBy  uuid.UUID   `json:"opened_by"`
	StartedAt time.Time   `json:"started_at"`
	EndedAt   *time.Time  `json:"ended_at,omitempty"`
	Status    ShiftStatus `json:"status"`

	OpeningCash    decimal.Decimal  `json:"opening_cash"`
	ClosingCash    *decimal.Decimal `json:"closing_cash,omitempty"`
	ExpectedCash   *decimal.Decimal `json:"expected_cash,omitempty"`
	CashDifference *decimal.Decimal `json:"cash_difference,omitempty"`

	TotalSales       decimal.Decimal `json:"total_sales"`
	TotalCashSales   decimal.Decimal `json:"total_cash_sales"`
	TotalVoided      decimal.Decimal `json:"total_voided"`
	TransactionCount int             `json:"transaction_count"`
	VoidCount        int             `json:"void_count"`

	Notes *string `json:"notes,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	OpenedByName *string               `json:"opened_by_name,omitempty"`
	ClosedBy     *uuid.UUID            `json:"closed_by,omitempty"`
	ClosedByName *string               `json:"closed_by_name,omitempty"`
	Operations   []CashDrawerOperation `json:"operations,omitempty"`
}

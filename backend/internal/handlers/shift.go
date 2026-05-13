package handlers

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"dashpoint/backend/internal/models"
	"dashpoint/backend/internal/repository"
)

type shiftStore interface {
	Create(context.Context, *models.Shift) error
	GetByID(context.Context, uuid.UUID) (*models.Shift, error)
	GetOpenShiftByEmployee(context.Context, uuid.UUID) (*models.Shift, error)
	CloseShift(context.Context, uuid.UUID, decimal.Decimal, *string, uuid.UUID) error
	List(context.Context, *repository.ShiftFilter) ([]models.Shift, int, error)
}

// ShiftHandler handles shift endpoints.
type ShiftHandler struct {
	shiftRepo shiftStore
}

// NewShiftHandler creates a new shift handler.
func NewShiftHandler(shiftRepo shiftStore) *ShiftHandler {
	return &ShiftHandler{shiftRepo: shiftRepo}
}

// StartShiftRequest represents the request to start a shift.
type StartShiftRequest struct {
	OpeningCash string  `json:"opening_cash"`
	Notes       *string `json:"notes"`
}

// CloseShiftRequest represents the request to close a shift.
type CloseShiftRequest struct {
	ClosingCash string  `json:"closing_cash"`
	Notes       *string `json:"notes"`
}

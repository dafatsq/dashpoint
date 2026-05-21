package handlers

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"

	"dashpoint/backend/internal/audit"
	"dashpoint/backend/internal/middleware"
	"dashpoint/backend/internal/models"
)

type cashDrawerStore interface {
	Create(context.Context, *models.CashDrawerOperation) error
	ListByShift(context.Context, uuid.UUID) ([]models.CashDrawerOperation, error)
	GetTotalsByShift(context.Context, uuid.UUID) (decimal.Decimal, decimal.Decimal, error)
}

// CashDrawerHandler handles cash drawer operation endpoints.
type CashDrawerHandler struct {
	cashDrawerRepo cashDrawerStore
	shiftRepo      shiftLookupStore
}

// NewCashDrawerHandler creates a new cash drawer handler.
func NewCashDrawerHandler(cashDrawerRepo cashDrawerStore, shiftRepo shiftLookupStore) *CashDrawerHandler {
	return &CashDrawerHandler{
		cashDrawerRepo: cashDrawerRepo,
		shiftRepo:      shiftRepo,
	}
}

// CashDrawerRequest represents the request for pay-in/pay-out.
type CashDrawerRequest struct {
	Amount string `json:"amount"`
	Reason string `json:"reason"`
}

// PayIn handles POST /api/v1/shifts/pay-in.
func (h *CashDrawerHandler) PayIn(c *fiber.Ctx) error {
	return h.recordCashDrawerOperation(c, models.CashDrawerOpPayIn, models.AuditActionCashPayIn, "Pay-in recorded successfully")
}

// PayOut handles POST /api/v1/shifts/pay-out.
func (h *CashDrawerHandler) PayOut(c *fiber.Ctx) error {
	return h.recordCashDrawerOperation(c, models.CashDrawerOpPayOut, models.AuditActionCashPayOut, "Pay-out recorded successfully")
}

// ListOperations handles GET /api/v1/shifts/:id/operations.
func (h *CashDrawerHandler) ListOperations(c *fiber.Ctx) error {
	shiftID, err := shiftParamUUID(c, "id", "INVALID_ID", "Invalid shift ID format")
	if err != nil {
		return respondAPIError(c, err)
	}

	ops, err := h.cashDrawerRepo.ListByShift(c.Context(), shiftID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list cash drawer operations")
		return middleware.JSONError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve operations")
	}

	payInTotal, payOutTotal, err := h.cashDrawerRepo.GetTotalsByShift(c.Context(), shiftID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get operation totals")
	}

	return c.JSON(fiber.Map{
		"operations":    ops,
		"pay_in_total":  payInTotal.String(),
		"pay_out_total": payOutTotal.String(),
	})
}

func (h *CashDrawerHandler) recordCashDrawerOperation(c *fiber.Ctx, opType models.CashDrawerOpType, auditAction models.AuditAction, successMessage string) error {
	userID := middleware.GetUserID(c)
	shift, err := h.shiftRepo.GetOpenShiftByEmployee(c.Context(), userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get open shift")
		return middleware.JSONError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get open shift")
	}
	if shift == nil {
		return middleware.JSONError(c, fiber.StatusNotFound, "NO_OPEN_SHIFT", "No open shift found")
	}

	req, amount, err := parseCashDrawerRequest(c)
	if err != nil {
		return respondAPIError(c, err)
	}

	op := &models.CashDrawerOperation{
		ShiftID:     shift.ID,
		Type:        opType,
		Amount:      amount,
		Reason:      req.Reason,
		PerformedBy: userID,
	}
	if err := h.cashDrawerRepo.Create(c.Context(), op); err != nil {
		log.Error().Err(err).Msg("Failed to create cash drawer operation")
		return middleware.JSONError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "Failed to record cash drawer operation")
	}

	label := cashDrawerOperationLabel(opType)
	newValues := map[string]interface{}{
		"type":   label,
		"amount": amount.String(),
		"reason": req.Reason,
	}
	audit.LogWithValues(c, auditAction, models.AuditEntityShift, shift.ID.String(), fmt.Sprintf("%s: %s - %s", label, amount.String(), req.Reason), nil, newValues)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":   successMessage,
		"operation": op,
	})
}

func cashDrawerOperationLabel(opType models.CashDrawerOpType) string {
	if opType == models.CashDrawerOpPayOut {
		return "Pay-out"
	}
	return "Pay-in"
}

func parseCashDrawerRequest(c *fiber.Ctx) (*CashDrawerRequest, decimal.Decimal, error) {
	var req CashDrawerRequest
	if err := c.BodyParser(&req); err != nil {
		return nil, decimal.Zero, &apiError{status: fiber.StatusBadRequest, code: "INVALID_REQUEST", message: "Invalid request body"}
	}
	if req.Reason == "" {
		return nil, decimal.Zero, &apiError{status: fiber.StatusBadRequest, code: "REASON_REQUIRED", message: "Reason is required for cash drawer operations"}
	}
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil || amount.LessThanOrEqual(decimal.Zero) {
		return nil, decimal.Zero, &apiError{status: fiber.StatusBadRequest, code: "INVALID_AMOUNT", message: "Amount must be a positive number"}
	}
	return &req, amount, nil
}

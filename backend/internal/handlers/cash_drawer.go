package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"

	"dashpoint/backend/internal/audit"
	"dashpoint/backend/internal/middleware"
	"dashpoint/backend/internal/models"
	"dashpoint/backend/internal/repository"
)

// CashDrawerHandler handles cash drawer operation endpoints
type CashDrawerHandler struct {
	cashDrawerRepo *repository.CashDrawerRepository
	shiftRepo      *repository.ShiftRepository
}

// NewCashDrawerHandler creates a new cash drawer handler
func NewCashDrawerHandler(cashDrawerRepo *repository.CashDrawerRepository, shiftRepo *repository.ShiftRepository) *CashDrawerHandler {
	return &CashDrawerHandler{
		cashDrawerRepo: cashDrawerRepo,
		shiftRepo:      shiftRepo,
	}
}

// CashDrawerRequest represents the request for pay-in/pay-out
type CashDrawerRequest struct {
	Amount string `json:"amount"`
	Reason string `json:"reason"`
}

// PayIn handles POST /api/v1/shifts/pay-in
func (h *CashDrawerHandler) PayIn(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	// Get current open shift
	shift, err := h.shiftRepo.GetOpenShiftByEmployee(c.Context(), userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get open shift")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to get open shift",
		})
	}

	if shift == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"code":    "NO_OPEN_SHIFT",
			"message": "No open shift found",
		})
	}

	var req CashDrawerRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request body",
		})
	}

	if req.Reason == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "REASON_REQUIRED",
			"message": "Reason is required for cash drawer operations",
		})
	}

	amount, err := decimal.NewFromString(req.Amount)
	if err != nil || amount.LessThanOrEqual(decimal.Zero) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_AMOUNT",
			"message": "Amount must be a positive number",
		})
	}

	op := &models.CashDrawerOperation{
		ShiftID:     shift.ID,
		Type:        models.CashDrawerOpPayIn,
		Amount:      amount,
		Reason:      req.Reason,
		PerformedBy: userID,
	}

	if err := h.cashDrawerRepo.Create(c.Context(), op); err != nil {
		log.Error().Err(err).Msg("Failed to create pay-in operation")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to record pay-in",
		})
	}

	// Audit log
	audit.LogFromFiber(c, models.AuditActionCashPayIn, models.AuditEntityShift, shift.ID.String(),
		fmt.Sprintf("Pay-in: %s - %s", amount.String(), req.Reason))

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":   "Pay-in recorded successfully",
		"operation": op,
	})
}

// PayOut handles POST /api/v1/shifts/pay-out
func (h *CashDrawerHandler) PayOut(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	// Get current open shift
	shift, err := h.shiftRepo.GetOpenShiftByEmployee(c.Context(), userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get open shift")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to get open shift",
		})
	}

	if shift == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"code":    "NO_OPEN_SHIFT",
			"message": "No open shift found",
		})
	}

	var req CashDrawerRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request body",
		})
	}

	if req.Reason == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "REASON_REQUIRED",
			"message": "Reason is required for cash drawer operations",
		})
	}

	amount, err := decimal.NewFromString(req.Amount)
	if err != nil || amount.LessThanOrEqual(decimal.Zero) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_AMOUNT",
			"message": "Amount must be a positive number",
		})
	}

	op := &models.CashDrawerOperation{
		ShiftID:     shift.ID,
		Type:        models.CashDrawerOpPayOut,
		Amount:      amount,
		Reason:      req.Reason,
		PerformedBy: userID,
	}

	if err := h.cashDrawerRepo.Create(c.Context(), op); err != nil {
		log.Error().Err(err).Msg("Failed to create pay-out operation")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to record pay-out",
		})
	}

	// Audit log
	audit.LogFromFiber(c, models.AuditActionCashPayOut, models.AuditEntityShift, shift.ID.String(),
		fmt.Sprintf("Pay-out: %s - %s", amount.String(), req.Reason))

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":   "Pay-out recorded successfully",
		"operation": op,
	})
}

// ListOperations handles GET /api/v1/shifts/:id/operations
func (h *CashDrawerHandler) ListOperations(c *fiber.Ctx) error {
	idStr := c.Params("id")
	shiftID, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_ID",
			"message": "Invalid shift ID format",
		})
	}

	ops, err := h.cashDrawerRepo.ListByShift(c.Context(), shiftID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list cash drawer operations")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to retrieve operations",
		})
	}

	// Get totals
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

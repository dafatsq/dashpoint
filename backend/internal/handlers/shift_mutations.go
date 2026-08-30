package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"

	"dashpoint/backend/internal/audit"
	"dashpoint/backend/internal/middleware"
	"dashpoint/backend/internal/models"
	"dashpoint/backend/internal/repository"
)

const (
	shiftStartSuccessMessage = "Shift started successfully"
	shiftCloseSuccessMessage = "Shift closed successfully"
)

// StartShift handles POST /api/v1/shifts/start.
func (h *ShiftHandler) StartShift(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	existingShift, err := h.shiftRepo.GetCurrentOpenShift(c.Context())
	if err != nil {
		log.Error().Err(err).Msg("Failed to check existing shift")
		return shiftInternalError(c, "Failed to check existing shift")
	}
	if existingShift != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"code":    "SHIFT_EXISTS",
			"message": "There is already an open shift",
			"shift":   existingShift,
		})
	}

	req, openingCash, err := parseStartShiftRequest(c)
	if err != nil {
		return shiftInvalidRequest(c)
	}

	shift := &models.Shift{
		OpenedBy:    userID,
		OpeningCash: openingCash,
		Notes:       req.Notes,
	}
	if err := h.shiftRepo.Create(c.Context(), shift); err != nil {
		if errors.Is(err, repository.ErrShiftAlreadyOpen) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"code":    "SHIFT_EXISTS",
				"message": "There is already an open shift",
			})
		}
		log.Error().Err(err).Msg("Failed to create shift")
		return shiftInternalError(c, "Failed to start shift")
	}

	if created, _ := h.shiftRepo.GetByID(c.Context(), shift.ID); created != nil {
		shift = created
	}

	newValues := map[string]interface{}{}
	if req.OpeningCash != "" {
		newValues["opening_cash"] = req.OpeningCash
	}
	if req.Notes != nil {
		newValues["notes"] = *req.Notes
	}
	audit.LogWithValues(c, models.AuditActionShiftStart, models.AuditEntityShift, shift.ID.String(), "Started shift", nil, newValues)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": shiftStartSuccessMessage,
		"shift":   shift,
	})
}

// CloseShift handles POST /api/v1/shifts/close.
func (h *ShiftHandler) CloseShift(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	shift, err := h.shiftRepo.GetCurrentOpenShift(c.Context())
	if err != nil {
		log.Error().Err(err).Msg("Failed to get open shift")
		return shiftInternalError(c, "Failed to get open shift")
	}
	if shift == nil {
		return middleware.JSONError(c, fiber.StatusNotFound, "NO_OPEN_SHIFT", "No open shift found")
	}

	req, closingCash, err := parseCloseShiftRequest(c)
	if err != nil {
		return err
	}
	expectedShiftID, err := parseExpectedUUID(req.ShiftID)
	if err != nil {
		return middleware.JSONError(c, fiber.StatusBadRequest, "INVALID_SHIFT_ID", "Invalid shift_id")
	}
	if expectedShiftID != nil && *expectedShiftID != shift.ID {
		return middleware.JSONError(c, fiber.StatusConflict, "STALE_SUBMIT", staleShiftMessage)
	}

	if err := h.shiftRepo.CloseShift(c.Context(), shift.ID, closingCash, req.Notes, userID); err != nil {
		log.Error().Err(err).Msg("Failed to close shift")
		return shiftInternalError(c, "Failed to close shift")
	}

	closed, _ := h.shiftRepo.GetByID(c.Context(), shift.ID)
	oldValues := map[string]interface{}{}
	if closed.ExpectedCash != nil {
		oldValues["expected_cash"] = closed.ExpectedCash.String()
	}
	newValues := map[string]interface{}{}
	if closed.ClosingCash != nil {
		newValues["closing_cash"] = closed.ClosingCash.String()
	}
	if closed.CashDifference != nil {
		newValues["difference"] = closed.CashDifference.String()
	}
	if req.Notes != nil {
		newValues["notes"] = *req.Notes
	}
	audit.LogWithValues(c, models.AuditActionShiftClose, models.AuditEntityShift, shift.ID.String(), "Closed shift", oldValues, newValues)

	return c.JSON(fiber.Map{
		"message": shiftCloseSuccessMessage,
		"shift":   closed,
	})
}

func parseStartShiftRequest(c *fiber.Ctx) (*StartShiftRequest, decimal.Decimal, error) {
	var req StartShiftRequest
	if err := c.BodyParser(&req); err != nil {
		return nil, decimal.Zero, err
	}

	openingCash := decimal.Zero
	if req.OpeningCash != "" {
		cash, err := decimal.NewFromString(req.OpeningCash)
		if err != nil {
			return nil, decimal.Zero, middleware.JSONError(c, fiber.StatusBadRequest, "INVALID_AMOUNT", "Invalid opening cash amount")
		}
		if cash.IsNegative() {
			return nil, decimal.Zero, middleware.JSONError(c, fiber.StatusBadRequest, "INVALID_AMOUNT", "Opening cash amount cannot be negative")
		}
		openingCash = cash
	}

	return &req, openingCash, nil
}

func parseCloseShiftRequest(c *fiber.Ctx) (*CloseShiftRequest, decimal.Decimal, error) {
	var req CloseShiftRequest
	if err := c.BodyParser(&req); err != nil {
		return nil, decimal.Zero, shiftInvalidRequest(c)
	}

	closingCash, err := decimal.NewFromString(req.ClosingCash)
	if err != nil {
		return nil, decimal.Zero, middleware.JSONError(c, fiber.StatusBadRequest, "INVALID_AMOUNT", "Invalid closing cash amount")
	}
	if closingCash.IsNegative() {
		return nil, decimal.Zero, middleware.JSONError(c, fiber.StatusBadRequest, "INVALID_AMOUNT", "Closing cash amount cannot be negative")
	}

	return &req, closingCash, nil
}

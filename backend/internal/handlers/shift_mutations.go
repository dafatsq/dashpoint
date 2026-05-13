package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"

	"dashpoint/backend/internal/audit"
	"dashpoint/backend/internal/middleware"
	"dashpoint/backend/internal/models"
)

// StartShift handles POST /api/v1/shifts/start.
func (h *ShiftHandler) StartShift(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	existingShift, err := h.shiftRepo.GetOpenShiftByEmployee(c.Context(), userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check existing shift")
		return shiftInternalError(c, "Failed to check existing shift")
	}
	if existingShift != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"code":    "SHIFT_EXISTS",
			"message": "You already have an open shift",
			"shift":   existingShift,
		})
	}

	var req StartShiftRequest
	if err := c.BodyParser(&req); err != nil {
		return shiftInvalidRequest(c)
	}

	openingCash := decimal.Zero
	if req.OpeningCash != "" {
		cash, err := decimal.NewFromString(req.OpeningCash)
		if err != nil {
			return middleware.JSONError(c, fiber.StatusBadRequest, "INVALID_AMOUNT", "Invalid opening cash amount")
		}
		openingCash = cash
	}

	shift := &models.Shift{
		EmployeeID:  userID,
		OpeningCash: openingCash,
		Notes:       req.Notes,
	}
	if err := h.shiftRepo.Create(c.Context(), shift); err != nil {
		log.Error().Err(err).Msg("Failed to create shift")
		return shiftInternalError(c, "Failed to start shift")
	}

	if created, _ := h.shiftRepo.GetByID(c.Context(), shift.ID); created != nil {
		shift = created
	}

	audit.LogFromFiber(c, models.AuditActionShiftStart, models.AuditEntityShift, shift.ID.String(), "Started shift")

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Shift started successfully",
		"shift":   shift,
	})
}

// CloseShift handles POST /api/v1/shifts/close.
func (h *ShiftHandler) CloseShift(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	shift, err := h.shiftRepo.GetOpenShiftByEmployee(c.Context(), userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get open shift")
		return shiftInternalError(c, "Failed to get open shift")
	}
	if shift == nil {
		return middleware.JSONError(c, fiber.StatusNotFound, "NO_OPEN_SHIFT", "No open shift found")
	}

	var req CloseShiftRequest
	if err := c.BodyParser(&req); err != nil {
		return shiftInvalidRequest(c)
	}

	closingCash, err := decimal.NewFromString(req.ClosingCash)
	if err != nil {
		return middleware.JSONError(c, fiber.StatusBadRequest, "INVALID_AMOUNT", "Invalid closing cash amount")
	}

	if err := h.shiftRepo.CloseShift(c.Context(), shift.ID, closingCash, req.Notes, userID); err != nil {
		log.Error().Err(err).Msg("Failed to close shift")
		return shiftInternalError(c, "Failed to close shift")
	}

	closed, _ := h.shiftRepo.GetByID(c.Context(), shift.ID)
	audit.LogFromFiber(c, models.AuditActionShiftClose, models.AuditEntityShift, shift.ID.String(), "Closed shift")

	return c.JSON(fiber.Map{
		"message": "Shift closed successfully",
		"shift":   closed,
	})
}

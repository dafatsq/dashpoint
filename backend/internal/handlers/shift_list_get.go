package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"

	"dashpoint/backend/internal/middleware"
)

// GetCurrentShift handles GET /api/v1/shifts/current.
func (h *ShiftHandler) GetCurrentShift(c *fiber.Ctx) error {
	shift, err := h.shiftRepo.GetCurrentOpenShift(c.Context())
	if err != nil {
		log.Error().Err(err).Msg("Failed to get current shift")
		return shiftInternalError(c, "Failed to get current shift")
	}
	if shift == nil {
		return c.JSON(fiber.Map{"shift": nil, "message": "No open shift"})
	}

	if c.Query("blind") == "true" {
		shift = blindShiftView(shift)
	}

	return c.JSON(fiber.Map{"shift": shift})
}

// GetShift handles GET /api/v1/shifts/:id.
func (h *ShiftHandler) GetShift(c *fiber.Ctx) error {
	id, err := shiftParamUUID(c, "id", "INVALID_ID", "Invalid shift ID format")
	if err != nil {
		return respondAPIError(c, err)
	}

	shift, err := h.shiftRepo.GetByID(c.Context(), id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get shift")
		return shiftInternalError(c, "Failed to retrieve shift")
	}
	if shift == nil {
		return middleware.JSONError(c, fiber.StatusNotFound, "NOT_FOUND", "Shift not found")
	}

	return c.JSON(fiber.Map{"shift": shift})
}

// ListShifts handles GET /api/v1/shifts.
func (h *ShiftHandler) ListShifts(c *fiber.Ctx) error {
	filter, err := parseShiftListFilter(c)
	if err != nil {
		return respondAPIError(c, err)
	}

	shifts, total, err := h.shiftRepo.List(c.Context(), filter)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list shifts")
		return shiftInternalError(c, "Failed to retrieve shifts")
	}

	return c.JSON(fiber.Map{
		"shifts": shifts,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

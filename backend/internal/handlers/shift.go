package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"

	"dashpoint/backend/internal/audit"
	"dashpoint/backend/internal/middleware"
	"dashpoint/backend/internal/models"
	"dashpoint/backend/internal/repository"
)

// ShiftHandler handles shift endpoints
type ShiftHandler struct {
	shiftRepo *repository.ShiftRepository
}

// NewShiftHandler creates a new shift handler
func NewShiftHandler(shiftRepo *repository.ShiftRepository) *ShiftHandler {
	return &ShiftHandler{shiftRepo: shiftRepo}
}

// StartShiftRequest represents the request to start a shift
type StartShiftRequest struct {
	OpeningCash string  `json:"opening_cash"`
	Notes       *string `json:"notes"`
}

// CloseShiftRequest represents the request to close a shift
type CloseShiftRequest struct {
	ClosingCash string  `json:"closing_cash"`
	Notes       *string `json:"notes"`
}

// StartShift handles POST /api/v1/shifts/start
func (h *ShiftHandler) StartShift(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	// Check if user already has an open shift
	existingShift, err := h.shiftRepo.GetOpenShiftByEmployee(c.Context(), userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check existing shift")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to check existing shift",
		})
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request body",
		})
	}

	openingCash := decimal.Zero
	if req.OpeningCash != "" {
		cash, err := decimal.NewFromString(req.OpeningCash)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code":    "INVALID_AMOUNT",
				"message": "Invalid opening cash amount",
			})
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
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to start shift",
		})
	}

	// Fetch the created shift with employee name
	created, _ := h.shiftRepo.GetByID(c.Context(), shift.ID)
	if created != nil {
		shift = created
	}

	// Audit log
	audit.LogFromFiber(c, models.AuditActionShiftStart, models.AuditEntityShift, shift.ID.String(), "Started shift")

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Shift started successfully",
		"shift":   shift,
	})
}

// CloseShift handles POST /api/v1/shifts/close
func (h *ShiftHandler) CloseShift(c *fiber.Ctx) error {
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

	var req CloseShiftRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request body",
		})
	}

	closingCash, err := decimal.NewFromString(req.ClosingCash)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_AMOUNT",
			"message": "Invalid closing cash amount",
		})
	}

	if err := h.shiftRepo.CloseShift(c.Context(), shift.ID, closingCash, req.Notes); err != nil {
		log.Error().Err(err).Msg("Failed to close shift")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to close shift",
		})
	}

	// Fetch the closed shift
	closed, _ := h.shiftRepo.GetByID(c.Context(), shift.ID)

	// Audit log
	audit.LogFromFiber(c, models.AuditActionShiftClose, models.AuditEntityShift, shift.ID.String(), "Closed shift")

	return c.JSON(fiber.Map{
		"message": "Shift closed successfully",
		"shift":   closed,
	})
}

// GetCurrentShift handles GET /api/v1/shifts/current
// Supports ?blind=true to hide financial details for blind close
func (h *ShiftHandler) GetCurrentShift(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	shift, err := h.shiftRepo.GetOpenShiftByEmployee(c.Context(), userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get current shift")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to get current shift",
		})
	}

	if shift == nil {
		return c.JSON(fiber.Map{
			"shift":   nil,
			"message": "No open shift",
		})
	}

	// Blind close mode: hide financial details so cashier can't match their count
	blind := c.Query("blind") == "true"
	if blind {
		shift.TotalSales = decimal.Zero
		shift.TotalRefunds = decimal.Zero
		shift.ExpectedCash = nil
		shift.TransactionCount = 0
		shift.RefundCount = 0
	}

	return c.JSON(fiber.Map{
		"shift": shift,
	})
}

// GetShift handles GET /api/v1/shifts/:id
func (h *ShiftHandler) GetShift(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_ID",
			"message": "Invalid shift ID format",
		})
	}

	shift, err := h.shiftRepo.GetByID(c.Context(), id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get shift")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to retrieve shift",
		})
	}

	if shift == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"code":    "NOT_FOUND",
			"message": "Shift not found",
		})
	}

	return c.JSON(fiber.Map{
		"shift": shift,
	})
}

// ListShifts handles GET /api/v1/shifts
func (h *ShiftHandler) ListShifts(c *fiber.Ctx) error {
	// Parse query parameters
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	var employeeID *uuid.UUID
	if empIDStr := c.Query("employee_id"); empIDStr != "" {
		id, err := uuid.Parse(empIDStr)
		if err == nil {
			employeeID = &id
		}
	}

	// For non-owner/manager, only show their own shifts
	roleName := middleware.GetRoleName(c)
	if roleName == "cashier" {
		userID := middleware.GetUserID(c)
		employeeID = &userID
	}

	shifts, total, err := h.shiftRepo.List(c.Context(), employeeID, limit, offset)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list shifts")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to retrieve shifts",
		})
	}

	return c.JSON(fiber.Map{
		"shifts": shifts,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

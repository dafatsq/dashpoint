package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"dashpoint/backend/internal/middleware"
	"dashpoint/backend/internal/models"
	"dashpoint/backend/internal/repository"
)

func shiftInvalidRequest(c *fiber.Ctx) error {
	return middleware.JSONError(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
}

func shiftInternalError(c *fiber.Ctx, message string) error {
	return middleware.JSONError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", message)
}

func shiftParamUUID(c *fiber.Ctx, name, code, message string) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Params(name))
	if err != nil {
		return uuid.Nil, &apiError{status: fiber.StatusBadRequest, code: code, message: message}
	}
	return id, nil
}

func parseShiftListFilter(c *fiber.Ctx) (*repository.ShiftFilter, error) {
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	filter := &repository.ShiftFilter{Limit: limit, Offset: offset}

	openedByStr := c.Query("opened_by_id")
	if openedByStr == "" {
		openedByStr = c.Query("user_id")
	}
	if openedByStr != "" {
		id, err := uuid.Parse(openedByStr)
		if err != nil {
			return nil, &apiError{status: fiber.StatusBadRequest, code: "INVALID_OPENED_BY_ID", message: "Invalid opened_by ID format"}
		}
		filter.OpenedByID = &id
	}

	if startStr := c.Query("from"); startStr != "" {
		t, err := parseReportDay(startStr, "from")
		if err != nil {
			return nil, &apiError{status: fiber.StatusBadRequest, code: "INVALID_START_DATE", message: "Invalid from date format. Use YYYY-MM-DD"}
		}
		t = reportDayStart(t)
		filter.StartDate = &t
	}

	if endStr := c.Query("to"); endStr != "" {
		t, err := parseReportDay(endStr, "to")
		if err != nil {
			return nil, &apiError{status: fiber.StatusBadRequest, code: "INVALID_END_DATE", message: "Invalid to date format. Use YYYY-MM-DD"}
		}
		exclusiveEnd := reportDayStart(t).Add(24 * time.Hour)
		filter.EndDate = &exclusiveEnd
	}

	return filter, nil
}

func blindShiftView(shift *models.Shift) *models.Shift {
	if shift == nil {
		return nil
	}
	copy := *shift
	copy.TotalSales = decimal.Zero
	copy.TotalVoided = decimal.Zero
	copy.ExpectedCash = nil
	copy.TransactionCount = 0
	copy.VoidCount = 0
	return &copy
}

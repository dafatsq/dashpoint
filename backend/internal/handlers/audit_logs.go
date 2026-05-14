package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// List handles GET /api/v1/logs
func (h *AuditHandler) List(c *fiber.Ctx) error {
	filter, err := parseAuditFilter(c)
	if err != nil {
		return err
	}

	logs, total, err := h.auditRepo.List(c.Context(), filter)
	if err != nil {
		return auditInternalError(c, err, "Failed to list audit logs", "Failed to retrieve audit logs")
	}

	return c.JSON(fiber.Map{
		"logs":   logs,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

// Get handles GET /api/v1/logs/:id
func (h *AuditHandler) Get(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return auditError(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid audit log ID format")
	}

	auditLog, err := h.auditRepo.GetByID(c.Context(), id)
	if err != nil {
		return auditInternalError(c, err, "Failed to get audit log", "Failed to retrieve audit log")
	}

	return c.JSON(fiber.Map{"log": auditLog})
}

// GetEntityHistory handles GET /api/v1/logs/entity/:type/:id
func (h *AuditHandler) GetEntityHistory(c *fiber.Ctx) error {
	entityType := c.Params("type")
	entityID := c.Params("id")
	limit := c.QueryInt("limit", 20)

	if entityType == "" || entityID == "" {
		return auditError(c, fiber.StatusBadRequest, "INVALID_PARAMS", "Entity type and ID are required")
	}

	logs, err := h.auditRepo.GetEntityHistory(c.Context(), entityType, entityID, limit)
	if err != nil {
		return auditInternalError(c, err, "Failed to get entity history", "Failed to retrieve entity history")
	}

	return c.JSON(fiber.Map{
		"entity_type": entityType,
		"entity_id":   entityID,
		"logs":        logs,
		"count":       len(logs),
	})
}

// GetUserActivity handles GET /api/v1/logs/user/:id
func (h *AuditHandler) GetUserActivity(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return auditError(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid user ID format")
	}

	limit := c.QueryInt("limit", 20)
	logs, err := h.auditRepo.GetUserActivity(c.Context(), userID, limit)
	if err != nil {
		return auditInternalError(c, err, "Failed to get user activity", "Failed to retrieve user activity")
	}

	return c.JSON(fiber.Map{
		"user_id": userID.String(),
		"logs":    logs,
		"count":   len(logs),
	})
}

// GetSummary handles GET /api/v1/logs/summary
func (h *AuditHandler) GetSummary(c *fiber.Ctx) error {
	dateRange, err := parseReportRangeResponse(c, 7, false, 0)
	if err != nil {
		return err
	}

	summary, err := h.auditRepo.GetActionSummary(c.Context(), dateRange.start, dateRange.end.Add(24*time.Hour))
	if err != nil {
		return auditInternalError(c, err, "Failed to get audit summary", "Failed to retrieve audit summary")
	}

	return c.JSON(fiber.Map{
		"start_date": dateRange.start.Format(reportDateLayout),
		"end_date":   dateRange.end.Format(reportDateLayout),
		"actions":    summary,
	})
}

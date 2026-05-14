package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"dashpoint/backend/internal/repository"
)

func auditError(c *fiber.Ctx, status int, code, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"code":    code,
		"message": message,
	})
}

func parseAuditFilter(c *fiber.Ctx) (repository.AuditFilter, error) {
	limit, err := strconv.Atoi(c.Query("limit", "50"))
	if err != nil {
		return repository.AuditFilter{}, auditError(c, fiber.StatusBadRequest, "INVALID_LIMIT", "Invalid limit")
	}

	offset, err := strconv.Atoi(c.Query("offset", "0"))
	if err != nil {
		return repository.AuditFilter{}, auditError(c, fiber.StatusBadRequest, "INVALID_OFFSET", "Invalid offset")
	}

	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		return repository.AuditFilter{}, auditError(c, fiber.StatusBadRequest, "INVALID_OFFSET", "Invalid offset")
	}

	filter := repository.AuditFilter{
		Limit:  limit,
		Offset: offset,
	}

	if userIDStr := c.Query("user_id"); userIDStr != "" {
		userID, parseErr := uuid.Parse(userIDStr)
		if parseErr != nil {
			return repository.AuditFilter{}, auditError(c, fiber.StatusBadRequest, "INVALID_USER_ID", "Invalid user_id format")
		}
		filter.UserID = &userID
	}

	if action := c.Query("action"); action != "" {
		filter.Action = &action
	}
	if entityType := c.Query("entity_type"); entityType != "" {
		filter.EntityType = &entityType
	}
	if entityID := c.Query("entity_id"); entityID != "" {
		filter.EntityID = &entityID
	}
	if status := c.Query("status"); status != "" {
		filter.Status = &status
	}
	if search := c.Query("search"); search != "" {
		filter.Search = &search
	}

	if startStr := c.Query("start_date"); startStr != "" {
		startDate, parseErr := time.Parse(reportDateLayout, startStr)
		if parseErr != nil {
			return repository.AuditFilter{}, auditError(c, fiber.StatusBadRequest, "INVALID_DATE", "Invalid start_date format")
		}
		filter.StartDate = &startDate
	}

	if endStr := c.Query("end_date"); endStr != "" {
		endDate, parseErr := time.Parse(reportDateLayout, endStr)
		if parseErr != nil {
			return repository.AuditFilter{}, auditError(c, fiber.StatusBadRequest, "INVALID_DATE", "Invalid end_date format")
		}
		endOfDay := endDate.Add(24*time.Hour - time.Second)
		filter.EndDate = &endOfDay
	}

	return filter, nil
}

func auditInternalError(c *fiber.Ctx, err error, logMessage, responseMessage string) error {
	log.Error().Err(err).Msg(logMessage)
	return auditError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", responseMessage)
}

package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"dashpoint/backend/internal/repository"
)

// AuditHandler handles audit log endpoints
type AuditHandler struct {
	auditRepo *repository.AuditRepository
}

// NewAuditHandler creates a new audit handler
func NewAuditHandler(auditRepo *repository.AuditRepository) *AuditHandler {
	return &AuditHandler{auditRepo: auditRepo}
}

// List handles GET /api/v1/logs
func (h *AuditHandler) List(c *fiber.Ctx) error {
	// Parse query parameters
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	if limit <= 0 || limit > 100 {
		limit = 50
	}

	filter := repository.AuditFilter{
		Limit:  limit,
		Offset: offset,
	}

	// User filter
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		id, err := uuid.Parse(userIDStr)
		if err == nil {
			filter.UserID = &id
		}
	}

	// Action filter
	if action := c.Query("action"); action != "" {
		filter.Action = &action
	}

	// Entity type filter
	if entityType := c.Query("entity_type"); entityType != "" {
		filter.EntityType = &entityType
	}

	// Entity ID filter
	if entityID := c.Query("entity_id"); entityID != "" {
		filter.EntityID = &entityID
	}

	// Status filter
	if status := c.Query("status"); status != "" {
		filter.Status = &status
	}

	// Date filters
	if startStr := c.Query("start_date"); startStr != "" {
		t, err := time.Parse("2006-01-02", startStr)
		if err == nil {
			filter.StartDate = &t
		}
	}

	if endStr := c.Query("end_date"); endStr != "" {
		t, err := time.Parse("2006-01-02", endStr)
		if err == nil {
			endOfDay := t.Add(24*time.Hour - time.Second)
			filter.EndDate = &endOfDay
		}
	}

	// Search filter
	if search := c.Query("search"); search != "" {
		filter.Search = &search
	}

	logs, total, err := h.auditRepo.List(c.Context(), filter)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list audit logs")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to retrieve audit logs",
		})
	}

	return c.JSON(fiber.Map{
		"logs":   logs,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// Get handles GET /api/v1/logs/:id
func (h *AuditHandler) Get(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_ID",
			"message": "Invalid audit log ID format",
		})
	}

	auditLog, err := h.auditRepo.GetByID(c.Context(), id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get audit log")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to retrieve audit log",
		})
	}

	return c.JSON(fiber.Map{
		"log": auditLog,
	})
}

// GetEntityHistory handles GET /api/v1/logs/entity/:type/:id
func (h *AuditHandler) GetEntityHistory(c *fiber.Ctx) error {
	entityType := c.Params("type")
	entityID := c.Params("id")
	limit := c.QueryInt("limit", 20)

	if entityType == "" || entityID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_PARAMS",
			"message": "Entity type and ID are required",
		})
	}

	logs, err := h.auditRepo.GetEntityHistory(c.Context(), entityType, entityID, limit)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get entity history")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to retrieve entity history",
		})
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
	idStr := c.Params("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_ID",
			"message": "Invalid user ID format",
		})
	}

	limit := c.QueryInt("limit", 20)

	logs, err := h.auditRepo.GetUserActivity(c.Context(), userID, limit)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user activity")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to retrieve user activity",
		})
	}

	return c.JSON(fiber.Map{
		"user_id": userID.String(),
		"logs":    logs,
		"count":   len(logs),
	})
}

// GetSummary handles GET /api/v1/logs/summary
func (h *AuditHandler) GetSummary(c *fiber.Ctx) error {
	startStr := c.Query("start_date")
	endStr := c.Query("end_date")

	var startDate, endDate time.Time
	var err error

	if startStr == "" || endStr == "" {
		// Default to last 7 days
		endDate = time.Now()
		startDate = endDate.AddDate(0, 0, -7)
	} else {
		startDate, err = time.Parse("2006-01-02", startStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code":    "INVALID_DATE",
				"message": "Invalid start_date format",
			})
		}
		endDate, err = time.Parse("2006-01-02", endStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code":    "INVALID_DATE",
				"message": "Invalid end_date format",
			})
		}
	}

	summary, err := h.auditRepo.GetActionSummary(c.Context(), startDate, endDate.Add(24*time.Hour))
	if err != nil {
		log.Error().Err(err).Msg("Failed to get audit summary")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to retrieve audit summary",
		})
	}

	return c.JSON(fiber.Map{
		"start_date": startDate.Format("2006-01-02"),
		"end_date":   endDate.Format("2006-01-02"),
		"actions":    summary,
	})
}

// GetActions handles GET /api/v1/logs/actions - returns available action types
func (h *AuditHandler) GetActions(c *fiber.Ctx) error {
	actions := []map[string]string{
		// Auth
		{"action": "auth.login", "description": "User login"},
		{"action": "auth.login_failed", "description": "Failed login attempt"},
		{"action": "auth.logout", "description": "User logout"},
		{"action": "auth.pin_login", "description": "PIN login"},
		// User
		{"action": "user.create", "description": "User created"},
		{"action": "user.update", "description": "User updated"},
		{"action": "user.delete", "description": "User deleted"},
		{"action": "user.deactivate", "description": "User deactivated"},
		{"action": "user.password_change", "description": "Password changed"},
		{"action": "user.pin_change", "description": "PIN changed"},
		{"action": "user.permission_change", "description": "Permissions changed"},
		// Product
		{"action": "product.create", "description": "Product created"},
		{"action": "product.update", "description": "Product updated"},
		{"action": "product.delete", "description": "Product deleted"},
		// Inventory
		{"action": "inventory.adjust", "description": "Stock adjusted"},
		{"action": "inventory.count", "description": "Stock count"},
		// Category
		{"action": "category.create", "description": "Category created"},
		{"action": "category.update", "description": "Category updated"},
		{"action": "category.delete", "description": "Category deleted"},
		// Sale
		{"action": "sale.create", "description": "Sale created"},
		{"action": "sale.void", "description": "Sale voided"},
		{"action": "sale.refund", "description": "Refund processed"},
		// Shift
		{"action": "shift.start", "description": "Shift started"},
		{"action": "shift.close", "description": "Shift closed"},
		// Report
		{"action": "report.export", "description": "Report exported"},
	}

	entityTypes := []map[string]string{
		{"type": "user", "description": "User accounts"},
		{"type": "product", "description": "Products"},
		{"type": "category", "description": "Categories"},
		{"type": "inventory", "description": "Inventory"},
		{"type": "sale", "description": "Sales"},
		{"type": "shift", "description": "Shifts"},
		{"type": "payment", "description": "Payments"},
		{"type": "report", "description": "Reports"},
		{"type": "auth", "description": "Authentication"},
		{"type": "system", "description": "System"},
	}

	return c.JSON(fiber.Map{
		"actions":      actions,
		"entity_types": entityTypes,
	})
}

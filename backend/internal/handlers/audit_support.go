package handlers

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"dashpoint/backend/internal/models"
	"dashpoint/backend/internal/repository"
)

const (
	auditMaxSearchLength   = 200
	auditMaxFilterLength   = 80
	auditMaxEntityIDLength = 128
	auditMaxOffset         = 10000
)

var validAuditActions = map[string]struct{}{
	string(models.AuditActionLogin):            {},
	string(models.AuditActionLoginFailed):      {},
	string(models.AuditActionLogout):           {},
	string(models.AuditActionPINLogin):         {},
	string(models.AuditActionUserCreate):       {},
	string(models.AuditActionUserUpdate):       {},
	string(models.AuditActionUserDelete):       {},
	string(models.AuditActionUserArchive):      {},
	string(models.AuditActionUserRestore):      {},
	string(models.AuditActionPasswordChange):   {},
	string(models.AuditActionPINChange):        {},
	string(models.AuditActionPermissionChange): {},
	string(models.AuditActionProductCreate):    {},
	string(models.AuditActionProductUpdate):    {},
	string(models.AuditActionProductDelete):    {},
	string(models.AuditActionProductArchive):   {},
	string(models.AuditActionProductRestore):   {},
	string(models.AuditActionStockAdjust):      {},
	string(models.AuditActionStockCount):       {},
	string(models.AuditActionThresholdUpdate):  {},
	string(models.AuditActionCategoryCreate):   {},
	string(models.AuditActionCategoryUpdate):   {},
	string(models.AuditActionCategoryDelete):   {},
	string(models.AuditActionCategoryArchive):  {},
	string(models.AuditActionCategoryRestore):  {},
	string(models.AuditActionSaleCreate):       {},
	string(models.AuditActionSaleVoid):         {},
	string(models.AuditActionShiftStart):       {},
	string(models.AuditActionShiftClose):       {},
	string(models.AuditActionCashPayIn):        {},
	string(models.AuditActionCashPayOut):       {},
	string(models.AuditActionExpenseCreate):    {},
	string(models.AuditActionExpenseUpdate):    {},
	string(models.AuditActionExpenseDelete):    {},
	string(models.AuditActionExpenseArchive):   {},
	string(models.AuditActionExpenseRestore):   {},
	string(models.AuditActionReportExport):     {},
	string(models.AuditActionSettingChange):    {},
	"user.deactivate":                          {},
}

var validAuditEntityTypes = map[string]struct{}{
	string(models.AuditEntityUser):      {},
	string(models.AuditEntityRole):      {},
	string(models.AuditEntityProduct):   {},
	string(models.AuditEntityCategory):  {},
	string(models.AuditEntityInventory): {},
	string(models.AuditEntitySale):      {},
	string(models.AuditEntityShift):     {},
	string(models.AuditEntityExpense):   {},
	string(models.AuditEntityPayment):   {},
	string(models.AuditEntityReport):    {},
	string(models.AuditEntitySystem):    {},
	string(models.AuditEntityAuth):      {},
}

var validAuditStatuses = map[string]struct{}{
	string(models.AuditStatusSuccess): {},
	string(models.AuditStatusFailure): {},
	string(models.AuditStatusWarning): {},
}

func auditError(c *fiber.Ctx, status int, code, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"code":    code,
		"message": message,
	})
}

func parseAuditFilter(c *fiber.Ctx) (repository.AuditFilter, error) {
	limit, err := parseBoundedIntQuery(c, "limit", 50, 100)
	if err != nil {
		return repository.AuditFilter{}, auditError(c, fiber.StatusBadRequest, "INVALID_LIMIT", "Invalid limit")
	}

	offset, err := parseNonNegativeIntQuery(c, "offset", 0, auditMaxOffset)
	if err != nil {
		return repository.AuditFilter{}, auditError(c, fiber.StatusBadRequest, "INVALID_OFFSET", "Invalid offset")
	}

	filter := repository.AuditFilter{
		Limit:  limit,
		Offset: offset,
	}

	if userIDStr := strings.TrimSpace(c.Query("user_id")); userIDStr != "" {
		userID, parseErr := uuid.Parse(userIDStr)
		if parseErr != nil {
			return repository.AuditFilter{}, auditError(c, fiber.StatusBadRequest, "INVALID_USER_ID", "Invalid user_id format")
		}
		filter.UserID = &userID
	}

	if action, parseErr := parseAuditEnumFilter(c, "action", validAuditActions, "INVALID_ACTION", "Invalid action filter"); parseErr != nil {
		return repository.AuditFilter{}, parseErr
	} else if action != "" {
		filter.Action = &action
	}
	if entityType, parseErr := parseAuditEnumFilter(c, "entity_type", validAuditEntityTypes, "INVALID_ENTITY_TYPE", "Invalid entity_type filter"); parseErr != nil {
		return repository.AuditFilter{}, parseErr
	} else if entityType != "" {
		filter.EntityType = &entityType
	}
	if entityID, parseErr := parseAuditEntityIDFilter(c.Query("entity_id")); parseErr != nil {
		return repository.AuditFilter{}, auditError(c, fiber.StatusBadRequest, "INVALID_ENTITY_ID", "Invalid entity_id filter")
	} else if entityID != "" {
		filter.EntityID = &entityID
	}
	if status, parseErr := parseAuditEnumFilter(c, "status", validAuditStatuses, "INVALID_STATUS", "Invalid status filter"); parseErr != nil {
		return repository.AuditFilter{}, parseErr
	} else if status != "" {
		filter.Status = &status
	}
	if search := strings.TrimSpace(c.Query("search")); search != "" {
		if len(search) > auditMaxSearchLength {
			return repository.AuditFilter{}, auditError(c, fiber.StatusBadRequest, "INVALID_SEARCH", "Invalid search filter")
		}
		filter.Search = &search
	}

	if startStr := strings.TrimSpace(c.Query("start_date")); startStr != "" {
		startDate, parseErr := parseReportDay(startStr, "start_date")
		if parseErr != nil {
			return repository.AuditFilter{}, auditError(c, fiber.StatusBadRequest, "INVALID_DATE", "Invalid start_date format")
		}
		filter.StartDate = &startDate
	}

	if endStr := strings.TrimSpace(c.Query("end_date")); endStr != "" {
		endDate, parseErr := parseReportDay(endStr, "end_date")
		if parseErr != nil {
			return repository.AuditFilter{}, auditError(c, fiber.StatusBadRequest, "INVALID_DATE", "Invalid end_date format")
		}
		exclusiveEndDate := reportDayStart(endDate).Add(24 * time.Hour)
		filter.EndDate = &exclusiveEndDate
	}

	return filter, nil
}

func parseAuditEnumFilter(c *fiber.Ctx, key string, allowed map[string]struct{}, code, message string) (string, error) {
	value := strings.TrimSpace(c.Query(key))
	if value == "" {
		return "", nil
	}
	if len(value) > auditMaxFilterLength {
		return "", auditError(c, fiber.StatusBadRequest, code, message)
	}
	if _, ok := allowed[value]; !ok {
		return "", auditError(c, fiber.StatusBadRequest, code, message)
	}
	return value, nil
}

func parseAuditEntityIDFilter(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", nil
	}
	if len(value) > auditMaxEntityIDLength {
		return "", errors.New("too long")
	}
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			continue
		}
		switch r {
		case '-', '_', '.', ':':
			continue
		default:
			return "", errors.New("invalid")
		}
	}
	return value, nil
}

func parseAuditHistoryEntityType(c *fiber.Ctx) (string, error) {
	entityType := strings.TrimSpace(c.Params("type"))
	if entityType == "" {
		return "", auditError(c, fiber.StatusBadRequest, "INVALID_PARAMS", "Entity type and ID are required")
	}
	if _, ok := validAuditEntityTypes[entityType]; !ok {
		return "", auditError(c, fiber.StatusBadRequest, "INVALID_ENTITY_TYPE", "Invalid entity type")
	}
	return entityType, nil
}

func parseAuditHistoryEntityID(c *fiber.Ctx) (string, error) {
	entityID, err := parseAuditEntityIDFilter(c.Params("id"))
	if err != nil || entityID == "" {
		return "", auditError(c, fiber.StatusBadRequest, "INVALID_ENTITY_ID", "Invalid entity ID")
	}
	return entityID, nil
}

func parseAuditLimitQuery(c *fiber.Ctx) (int, error) {
	limit, err := parseBoundedIntQuery(c, "limit", 20, 100)
	if err != nil {
		return 0, auditError(c, fiber.StatusBadRequest, "INVALID_LIMIT", "Invalid limit")
	}
	return limit, nil
}

func auditInternalError(c *fiber.Ctx, err error, logMessage, responseMessage string) error {
	log.Error().Err(err).Msg(logMessage)
	return auditError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", responseMessage)
}

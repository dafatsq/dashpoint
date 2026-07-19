package handlers

import (
	"context"
	"sort"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"dashpoint/backend/internal/models"
)

type roleEndpointReader interface {
	List(ctx context.Context) ([]*models.Role, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Role, error)
	ListActiveUserIDs(ctx context.Context, roleID uuid.UUID) ([]uuid.UUID, error)
	UpdatePermissions(ctx context.Context, id uuid.UUID, permissionKeys []string) error
}

const rolePermissionMaxJSONBodyBytes = 8192

func parseStrictRoleJSON(c *fiber.Ctx, dest interface{}) error {
	return parseStrictJSONBody(c, dest, rolePermissionMaxJSONBodyBytes)
}

func parseRoleID(c *fiber.Ctx) (uuid.UUID, error) {
	return uuid.Parse(c.Params("id"))
}

func roleInvalidIDError(c *fiber.Ctx) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"code":    "INVALID_ID",
		"message": "Invalid role ID format",
	})
}

func roleInternalError(c *fiber.Ctx, err error, message string) error {
	log.Error().Err(err).Msg(message)
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"code":    "INTERNAL_ERROR",
		"message": message,
	})
}

func mapRoleResponses(roles []*models.Role) []RoleResponse {
	response := make([]RoleResponse, len(roles))
	for i, role := range roles {
		response[i] = RoleResponse{
			ID:          role.ID.String(),
			Name:        role.Name,
			Description: role.Description,
			Permissions: role.Permissions,
		}
	}
	return response
}

var managePermissionAccessParents = map[string]string{
	"manage_pos_page":        "access_pos_page",
	"manage_products_page":   "access_products_page",
	"manage_inventory_page":  "access_inventory_page",
	"manage_sales_page":      "access_sales_page",
	"manage_reports_page":    "access_reports_page",
	"manage_expenses_page":   "access_expenses_page",
	"manage_categories_page": "access_categories_page",
	"manage_users_page":      "access_users_page",
	"manage_shifts_page":     "access_shifts_page",
}

var pagePermissionKeys = map[string]bool{
	"access_pos_page":        true,
	"manage_pos_page":        true,
	"access_products_page":   true,
	"manage_products_page":   true,
	"access_inventory_page":  true,
	"manage_inventory_page":  true,
	"access_sales_page":      true,
	"manage_sales_page":      true,
	"access_reports_page":    true,
	"manage_reports_page":    true,
	"access_expenses_page":   true,
	"manage_expenses_page":   true,
	"access_categories_page": true,
	"manage_categories_page": true,
	"access_users_page":      true,
	"manage_users_page":      true,
	"access_shifts_page":     true,
	"manage_shifts_page":     true,
	"access_changes_page":    true,
	"access_audit_page":      true,
}

var deprecatedRolePermissionKeys = map[string]bool{
	"access_settings_page": true,
}

func normalizeRolePermissionKeys(keys []string) ([]string, bool) {
	if len(keys) > 64 {
		return nil, false
	}

	permissionSet := make(map[string]bool, len(keys))
	for _, key := range keys {
		if deprecatedRolePermissionKeys[key] {
			continue
		}
		if !pagePermissionKeys[key] {
			return nil, false
		}
		permissionSet[key] = true
	}

	for manageKey, accessKey := range managePermissionAccessParents {
		if permissionSet[manageKey] {
			permissionSet[accessKey] = true
		}
	}

	normalized := make([]string, 0, len(permissionSet))
	for key := range permissionSet {
		normalized = append(normalized, key)
	}
	sort.Strings(normalized)

	return normalized, true
}

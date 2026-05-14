package handlers

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"dashpoint/backend/internal/models"
)

type roleEndpointReader interface {
	List(ctx context.Context) ([]*models.Role, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Role, error)
	GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*models.Permission, error)
}

type rolePermissionEndpointReader interface {
	List(ctx context.Context) ([]*models.Permission, error)
	ListByCategory(ctx context.Context) (map[string][]*models.Permission, error)
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
		}
	}
	return response
}

func mapPermissionResponses(permissions []*models.Permission) []PermissionResponse {
	response := make([]PermissionResponse, len(permissions))
	for i, permission := range permissions {
		response[i] = PermissionResponse{
			ID:          permission.ID.String(),
			Key:         permission.Key,
			Name:        permission.Name,
			Description: permission.Description,
			Category:    permission.Category,
		}
	}
	return response
}

func mapGroupedPermissionResponses(grouped map[string][]*models.Permission) map[string][]PermissionResponse {
	response := make(map[string][]PermissionResponse, len(grouped))
	for category, permissions := range grouped {
		response[category] = mapPermissionResponses(permissions)
	}
	return response
}

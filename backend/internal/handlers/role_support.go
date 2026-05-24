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

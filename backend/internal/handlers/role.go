package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"dashpoint/backend/internal/repository"
)

// RoleHandler handles role and permission endpoints
type RoleHandler struct {
	roleRepo       *repository.RoleRepository
	permissionRepo *repository.PermissionRepository
}

// NewRoleHandler creates a new role handler
func NewRoleHandler(
	roleRepo *repository.RoleRepository,
	permissionRepo *repository.PermissionRepository,
) *RoleHandler {
	return &RoleHandler{
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
	}
}

// PermissionResponse represents a permission in API responses
type PermissionResponse struct {
	ID          string  `json:"id"`
	Key         string  `json:"key"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Category    string  `json:"category"`
}

// RoleWithPermissionsResponse represents a role with its permissions
type RoleWithPermissionsResponse struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Description *string              `json:"description,omitempty"`
	Permissions []PermissionResponse `json:"permissions"`
}

// ListRoles handles GET /api/v1/roles
func (h *RoleHandler) ListRoles(c *fiber.Ctx) error {
	roles, err := h.roleRepo.List(c.Context())
	if err != nil {
		log.Error().Err(err).Msg("Failed to list roles")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to retrieve roles",
		})
	}

	response := make([]RoleResponse, len(roles))
	for i, role := range roles {
		response[i] = RoleResponse{
			ID:          role.ID.String(),
			Name:        role.Name,
			Description: role.Description,
		}
	}

	return c.JSON(fiber.Map{
		"roles": response,
	})
}

// GetRole handles GET /api/v1/roles/:id
func (h *RoleHandler) GetRole(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_ID",
			"message": "Invalid role ID format",
		})
	}

	role, err := h.roleRepo.GetByID(c.Context(), id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get role")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to retrieve role",
		})
	}

	if role == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"code":    "NOT_FOUND",
			"message": "Role not found",
		})
	}

	// Get role permissions
	permissions, err := h.roleRepo.GetRolePermissions(c.Context(), id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get role permissions")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to retrieve role permissions",
		})
	}

	permResponses := make([]PermissionResponse, len(permissions))
	for i, p := range permissions {
		permResponses[i] = PermissionResponse{
			ID:          p.ID.String(),
			Key:         p.Key,
			Name:        p.Name,
			Description: p.Description,
			Category:    p.Category,
		}
	}

	return c.JSON(RoleWithPermissionsResponse{
		ID:          role.ID.String(),
		Name:        role.Name,
		Description: role.Description,
		Permissions: permResponses,
	})
}

// ListPermissions handles GET /api/v1/permissions
func (h *RoleHandler) ListPermissions(c *fiber.Ctx) error {
	grouped := c.Query("grouped", "false") == "true"

	if grouped {
		permsByCategory, err := h.permissionRepo.ListByCategory(c.Context())
		if err != nil {
			log.Error().Err(err).Msg("Failed to list permissions")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to retrieve permissions",
			})
		}

		response := make(map[string][]PermissionResponse)
		for category, perms := range permsByCategory {
			response[category] = make([]PermissionResponse, len(perms))
			for i, p := range perms {
				response[category][i] = PermissionResponse{
					ID:          p.ID.String(),
					Key:         p.Key,
					Name:        p.Name,
					Description: p.Description,
					Category:    p.Category,
				}
			}
		}

		return c.JSON(fiber.Map{
			"permissions": response,
		})
	}

	permissions, err := h.permissionRepo.List(c.Context())
	if err != nil {
		log.Error().Err(err).Msg("Failed to list permissions")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to retrieve permissions",
		})
	}

	response := make([]PermissionResponse, len(permissions))
	for i, p := range permissions {
		response[i] = PermissionResponse{
			ID:          p.ID.String(),
			Key:         p.Key,
			Name:        p.Name,
			Description: p.Description,
			Category:    p.Category,
		}
	}

	return c.JSON(fiber.Map{
		"permissions": response,
	})
}

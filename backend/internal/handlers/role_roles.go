package handlers

import "github.com/gofiber/fiber/v2"

// GetRole handles GET /api/v1/roles/:id
func (h *RoleHandler) GetRole(c *fiber.Ctx) error {
	id, err := parseRoleID(c)
	if err != nil {
		return roleInvalidIDError(c)
	}

	role, err := h.roleRepo.GetByID(c.Context(), id)
	if err != nil {
		return roleInternalError(c, err, "Failed to retrieve role")
	}
	if role == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"code":    "NOT_FOUND",
			"message": "Role not found",
		})
	}

	permissions, err := h.roleRepo.GetRolePermissions(c.Context(), id)
	if err != nil {
		return roleInternalError(c, err, "Failed to retrieve role permissions")
	}

	return c.JSON(RoleWithPermissionsResponse{
		ID:          role.ID.String(),
		Name:        role.Name,
		Description: role.Description,
		Permissions: mapPermissionResponses(permissions),
	})
}

// ListRoles handles GET /api/v1/roles
func (h *RoleHandler) ListRoles(c *fiber.Ctx) error {
	roles, err := h.roleRepo.List(c.Context())
	if err != nil {
		return roleInternalError(c, err, "Failed to retrieve roles")
	}

	return c.JSON(fiber.Map{
		"roles": mapRoleResponses(roles),
	})
}

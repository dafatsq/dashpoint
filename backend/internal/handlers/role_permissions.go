package handlers

import "github.com/gofiber/fiber/v2"

// ListPermissions handles GET /api/v1/permissions
func (h *RoleHandler) ListPermissions(c *fiber.Ctx) error {
	grouped := c.Query("grouped", "false") == "true"

	if grouped {
		permissionsByCategory, err := h.permissionRepo.ListByCategory(c.Context())
		if err != nil {
			return roleInternalError(c, err, "Failed to retrieve permissions")
		}

		return c.JSON(fiber.Map{
			"permissions": mapGroupedPermissionResponses(permissionsByCategory),
		})
	}

	permissions, err := h.permissionRepo.List(c.Context())
	if err != nil {
		return roleInternalError(c, err, "Failed to retrieve permissions")
	}

	return c.JSON(fiber.Map{
		"permissions": mapPermissionResponses(permissions),
	})
}

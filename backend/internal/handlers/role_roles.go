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

	return c.JSON(RoleDetailResponse{
		ID:          role.ID.String(),
		Name:        role.Name,
		Description: role.Description,
		Permissions: role.Permissions,
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

// UpdateRolePermissions handles PATCH /api/v1/roles/:id/permissions.
func (h *RoleHandler) UpdateRolePermissions(c *fiber.Ctx) error {
	id, err := parseRoleID(c)
	if err != nil {
		return roleInvalidIDError(c)
	}

	var req UpdateRolePermissionsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request body",
		})
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
	if role.Name == roleOwner {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"code":    "OWNER_ROLE_LOCKED",
			"message": "Owner permissions cannot be changed",
		})
	}
	if req.ExpectedPermissions != nil {
		currentPermissions, ok := normalizeRolePermissionKeys(role.Permissions)
		if !ok {
			return roleInternalError(c, nil, "Failed to validate current permissions")
		}
		expectedPermissions, ok := normalizeRolePermissionKeys(req.ExpectedPermissions)
		if !ok {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code":    "INVALID_PERMISSION",
				"message": "Invalid permission key",
			})
		}
		if !sameStringSet(currentPermissions, expectedPermissions) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"code":    "STALE_SUBMIT",
				"message": staleSubmitMessage,
			})
		}
	}

	permissions, ok := normalizeRolePermissionKeys(req.Permissions)
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_PERMISSION",
			"message": "Invalid permission key",
		})
	}

	if err := h.roleRepo.UpdatePermissions(c.Context(), id, permissions); err != nil {
		return roleInternalError(c, err, "Failed to update role permissions")
	}

	updatedRole, err := h.roleRepo.GetByID(c.Context(), id)
	if err != nil {
		return roleInternalError(c, err, "Failed to retrieve role")
	}

	return c.JSON(fiber.Map{
		"message": "Role permissions updated successfully",
		"role": RoleDetailResponse{
			ID:          updatedRole.ID.String(),
			Name:        updatedRole.Name,
			Description: updatedRole.Description,
			Permissions: updatedRole.Permissions,
		},
	})
}

func sameStringSet(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

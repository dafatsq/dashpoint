package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"dashpoint/backend/internal/audit"
	"dashpoint/backend/internal/middleware"
	"dashpoint/backend/internal/models"
)

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
	if err := parseStrictRoleJSON(c, &req); err != nil {
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
	if req.ExpectedPermissions == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "EXPECTED_PERMISSIONS_REQUIRED",
			"message": "expected_permissions is required",
		})
	}

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

	audit.LogWithValues(c, models.AuditActionPermissionChange, models.AuditEntityRole, id.String(), "Updated role permissions: "+role.Name,
		map[string]interface{}{"affected_role": role.Name, "permissions": currentPermissions},
		map[string]interface{}{"affected_role": role.Name, "permissions": permissions})
	h.broadcastRolePermissionsChanged(c, id, role.Name)

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

func (h *RoleHandler) broadcastRolePermissionsChanged(c *fiber.Ctx, roleID uuid.UUID, roleName string) {
	if h.eventsHandler == nil {
		return
	}

	userIDs, err := h.roleRepo.ListActiveUserIDs(c.Context(), roleID)
	if err != nil {
		log.Error().Err(err).Str("role_id", roleID.String()).Msg("Failed to notify role users")
		return
	}

	changedBy := middleware.GetUserID(c)
	eventTime := time.Now()
	for _, userID := range userIDs {
		h.eventsHandler.BroadcastToUser(userID, UserEvent{
			Type:      EventPermissionsChanged,
			UserID:    userID.String(),
			ChangedBy: changedBy.String(),
			Timestamp: eventTime,
			Details: map[string]interface{}{
				"role_id":   roleID.String(),
				"role_name": roleName,
			},
		})
	}
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

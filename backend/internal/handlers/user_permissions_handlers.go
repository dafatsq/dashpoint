package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"dashpoint/backend/internal/audit"
	"dashpoint/backend/internal/middleware"
	"dashpoint/backend/internal/models"
)

// SetPermissions handles PATCH /api/v1/users/:id/permissions.
func (h *UserHandler) SetPermissions(c *fiber.Ctx) error {
	id, err := parseUserIDParam(c)
	if err != nil {
		return err
	}

	user, err := h.userRepo.GetByID(c.Context(), id)
	if err != nil || user == nil {
		return userNotFound(c)
	}

	currentUserID := middleware.GetUserID(c)
	if currentUserID == id {
		return middleware.JSONError(c, fiber.StatusForbidden, "CANNOT_MODIFY_SELF", "You cannot modify your own permissions")
	}

	targetRoleName := roleNameOfUser(user)
	if !h.enforceTargetUserAction(c, targetRoleName, userActionManagePermissions) {
		return nil
	}

	var req SetPermissionsRequest
	if err := c.BodyParser(&req); err != nil {
		return badUserRequest(c, "INVALID_REQUEST", "Invalid request body")
	}

	grantedBy := middleware.GetUserID(c)
	currentUserPermSet, err := h.permissionSetForPermissionManagement(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch current user permissions")
		return userInternalError(c, "Failed to validate permissions")
	}

	oldOverrides, _ := h.userRepo.GetUserPermissionOverrides(c.Context(), id)
	oldOverrideMap := make(map[string]string)
	for _, override := range oldOverrides {
		if override.Permission != nil {
			status := "denied"
			if override.Allowed {
				status = "granted"
			}
			oldOverrideMap[override.PermissionID.String()] = status
		}
	}

	oldValues := make(map[string]interface{})
	newValues := make(map[string]interface{})
	for _, perm := range req.Permissions {
		permID, err := uuid.Parse(perm.PermissionID)
		if err != nil {
			return badUserRequest(c, "INVALID_PERMISSION_ID", "Invalid permission ID: "+perm.PermissionID)
		}

		permission, err := h.permissionRepo.GetByID(c.Context(), permID)
		if err != nil || permission == nil {
			return badUserRequest(c, "INVALID_PERMISSION", "Permission not found: "+perm.PermissionID)
		}

		if currentUserPermSet != nil && !currentUserPermSet[permission.Key] {
			action := "grant"
			if !perm.Allowed {
				action = "deny"
			}
			return userForbidden(c, "You cannot "+action+" the '"+permission.Name+"' permission because you do not have it yourself")
		}

		if err := h.userRepo.SetUserPermission(c.Context(), id, permID, perm.Allowed, &grantedBy); err != nil {
			log.Error().Err(err).Msg("Failed to set user permission")
			return userInternalError(c, "Failed to set permissions")
		}

		if oldStatus, exists := oldOverrideMap[perm.PermissionID]; exists {
			oldValues[permission.Name] = oldStatus
		} else {
			oldValues[permission.Name] = "from role"
		}

		newStatus := "denied"
		if perm.Allowed {
			newStatus = "granted"
		}
		newValues[permission.Name] = newStatus
	}

	permissions, _ := h.userRepo.GetUserPermissions(c.Context(), id)
	overrides, _ := h.userRepo.GetUserPermissionOverrides(c.Context(), id)
	oldValues["affected_user"] = user.Name
	newValues["affected_user"] = user.Name

	audit.LogWithValues(c, models.AuditActionUserUpdate, models.AuditEntityUser, id.String(), "Updated permissions for: "+user.Name, oldValues, newValues)
	h.broadcastUserEvent(id, EventPermissionsChanged, grantedBy, map[string]interface{}{"permissions": permissions})

	return c.JSON(fiber.Map{
		"message":               "Permissions updated successfully",
		"effective_permissions": permissions,
		"overrides":             len(overrides),
	})
}

func (h *UserHandler) permissionSetForPermissionManagement(c *fiber.Ctx) (map[string]bool, error) {
	if isRoleName(middleware.GetRoleName(c), roleOwner) {
		return nil, nil
	}

	return h.currentUserPermissionSet(c)
}

// GetPermissions handles GET /api/v1/users/:id/permissions.
func (h *UserHandler) GetPermissions(c *fiber.Ctx) error {
	id, err := parseUserIDParam(c)
	if err != nil {
		return err
	}

	permissions, err := h.userRepo.GetUserPermissions(c.Context(), id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user permissions")
		return userInternalError(c, "Failed to retrieve permissions")
	}

	overrides, err := h.userRepo.GetUserPermissionOverrides(c.Context(), id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get permission overrides")
	}

	return c.JSON(fiber.Map{
		"effective_permissions": permissions,
		"overrides":             overrides,
	})
}

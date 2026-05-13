package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"

	"dashpoint/backend/internal/audit"
	"dashpoint/backend/internal/middleware"
	"dashpoint/backend/internal/models"
)

// Delete handles DELETE /api/v1/users/:id.
func (h *UserHandler) Delete(c *fiber.Ctx) error {
	id, err := parseUserIDParam(c)
	if err != nil {
		return err
	}

	currentUserID := middleware.GetUserID(c)
	if currentUserID == id {
		return badUserRequest(c, "CANNOT_DELETE_SELF", "You cannot deactivate your own account")
	}

	user, err := h.userRepo.GetByID(c.Context(), id)
	if err != nil || user == nil {
		return userNotFound(c)
	}

	targetRoleName := ""
	if user.Role != nil {
		targetRoleName = user.Role.Name
	}
	if err := h.enforceTargetUserAction(c, targetRoleName, userActionDelete); err != nil {
		return err
	}

	if err := h.userRepo.Deactivate(c.Context(), id); err != nil {
		log.Error().Err(err).Msg("Failed to deactivate user")
		return userInternalError(c, "Failed to deactivate user")
	}

	roleName := ""
	if user.Role != nil {
		roleName = user.Role.Name
	}
	audit.LogWithValues(c, models.AuditActionUserArchive, models.AuditEntityUser, id.String(), "Archived user: "+user.Name,
		map[string]interface{}{"affected_user": user.Name, "name": user.Name, "role": roleName, "status": "active"}, nil)

	h.broadcastUserEvent(id, EventUserDeactivated, currentUserID, nil)
	if h.eventsHandler != nil {
		h.eventsHandler.DisconnectUser(id)
	}

	return c.JSON(fiber.Map{"message": "User deactivated successfully"})
}

// PermanentDelete handles DELETE /api/v1/users/:id/permanent.
func (h *UserHandler) PermanentDelete(c *fiber.Ctx) error {
	id, err := parseUserIDParam(c)
	if err != nil {
		return err
	}

	currentUserID := middleware.GetUserID(c)
	if currentUserID == id {
		return badUserRequest(c, "CANNOT_DELETE_SELF", "You cannot delete your own account")
	}

	user, err := h.userRepo.GetByID(c.Context(), id)
	if err != nil || user == nil {
		return userNotFound(c)
	}

	targetRoleName := ""
	if user.Role != nil {
		targetRoleName = user.Role.Name
	}
	if err := h.enforceTargetUserAction(c, targetRoleName, userActionDelete); err != nil {
		return err
	}

	hasSales, err := h.userRepo.HasSalesHistory(c.Context(), id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check sales history")
		return userInternalError(c, "Failed to check user sales history")
	}
	if hasSales {
		return userConflict(c, "HAS_SALES_HISTORY", "Cannot permanently delete user with sales history. The user must remain archived.")
	}

	if err := h.userRepo.PermanentDelete(c.Context(), id); err != nil {
		log.Error().Err(err).Msg("Failed to permanently delete user")
		return userInternalError(c, "Failed to permanently delete user")
	}

	delRoleName := ""
	if user.Role != nil {
		delRoleName = user.Role.Name
	}
	delEmail := ""
	if user.Email != nil {
		delEmail = *user.Email
	}
	audit.LogWithValues(c, models.AuditActionUserDelete, models.AuditEntityUser, id.String(), "Permanently deleted user: "+user.Name,
		map[string]interface{}{"affected_user": user.Name, "name": user.Name, "role": delRoleName, "email": delEmail, "status": "archived"}, nil)

	h.broadcastUserEvent(id, EventUserDeleted, currentUserID, nil)
	if h.eventsHandler != nil {
		h.eventsHandler.DisconnectUser(id)
	}

	return c.JSON(fiber.Map{"message": "User permanently deleted"})
}

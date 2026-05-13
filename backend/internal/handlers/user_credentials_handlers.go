package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"

	"dashpoint/backend/internal/audit"
	"dashpoint/backend/internal/auth"
	"dashpoint/backend/internal/models"
)

// UpdatePassword handles PATCH /api/v1/users/:id/password.
func (h *UserHandler) UpdatePassword(c *fiber.Ctx) error {
	id, err := parseUserIDParam(c)
	if err != nil {
		return err
	}

	var req UpdatePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return badUserRequest(c, "INVALID_REQUEST", "Invalid request body")
	}

	user, err := h.userRepo.GetByID(c.Context(), id)
	if err != nil || user == nil {
		return userNotFound(c)
	}

	targetRoleName := ""
	if user.Role != nil {
		targetRoleName = user.Role.Name
	}
	if err := h.enforceTargetUserAction(c, targetRoleName, userActionEdit); err != nil {
		return err
	}

	if req.Password == "" {
		return badUserRequest(c, "VALIDATION_ERROR", "Password is required")
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		log.Error().Err(err).Msg("Failed to hash password")
		return userInternalError(c, "Failed to process password")
	}
	if err := h.userRepo.UpdatePassword(c.Context(), id, hash); err != nil {
		log.Error().Err(err).Msg("Failed to update password")
		return userInternalError(c, "Failed to update password")
	}

	user, _ = h.userRepo.GetByID(c.Context(), id)
	userName := "Unknown"
	if user != nil {
		userName = user.Name
	}
	audit.LogWithValues(c, models.AuditActionUserUpdate, models.AuditEntityUser, id.String(), "Updated password for: "+userName,
		map[string]interface{}{"affected_user": userName, "password": "[set]"},
		map[string]interface{}{"affected_user": userName, "password": "[changed]"})

	return c.JSON(fiber.Map{"message": "Password updated successfully"})
}

// UpdatePIN handles PATCH /api/v1/users/:id/pin.
func (h *UserHandler) UpdatePIN(c *fiber.Ctx) error {
	id, err := parseUserIDParam(c)
	if err != nil {
		return err
	}

	var req UpdatePINRequest
	if err := c.BodyParser(&req); err != nil {
		return badUserRequest(c, "INVALID_REQUEST", "Invalid request body")
	}

	user, err := h.userRepo.GetByID(c.Context(), id)
	if err != nil || user == nil {
		return userNotFound(c)
	}

	targetRoleName := ""
	if user.Role != nil {
		targetRoleName = user.Role.Name
	}
	if err := h.enforceTargetUserAction(c, targetRoleName, userActionEdit); err != nil {
		return err
	}

	var pinHash *string
	if req.PIN != nil && *req.PIN != "" {
		hash, err := auth.HashPIN(*req.PIN)
		if err != nil {
			log.Error().Err(err).Msg("Failed to hash PIN")
			return userInternalError(c, "Failed to process PIN")
		}
		pinHash = &hash
	}

	if err := h.userRepo.UpdatePIN(c.Context(), id, pinHash); err != nil {
		log.Error().Err(err).Msg("Failed to update PIN")
		return userInternalError(c, "Failed to update PIN")
	}

	user, _ = h.userRepo.GetByID(c.Context(), id)
	userName := "Unknown"
	if user != nil {
		userName = user.Name
	}
	newPinStatus := "[changed]"
	if pinHash == nil {
		newPinStatus = "[removed]"
	}
	audit.LogWithValues(c, models.AuditActionUserUpdate, models.AuditEntityUser, id.String(), "Updated PIN for: "+userName,
		map[string]interface{}{"affected_user": userName, "pin": "[set]"},
		map[string]interface{}{"affected_user": userName, "pin": newPinStatus})

	message := "PIN updated successfully"
	if pinHash == nil {
		message = "PIN removed successfully"
	}
	return c.JSON(fiber.Map{"message": message})
}

package handlers

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"

	"dashpoint/backend/internal/audit"
	"dashpoint/backend/internal/auth"
	"dashpoint/backend/internal/middleware"
	"dashpoint/backend/internal/models"
)

// UpdatePassword handles PATCH /api/v1/users/:id/password.
func (h *UserHandler) UpdatePassword(c *fiber.Ctx) error {
	id, err := parseUserIDParam(c)
	if err != nil {
		return err
	}

	var req UpdatePasswordRequest
	if err := parseStrictUserJSON(c, &req); err != nil {
		return badUserRequest(c, "INVALID_REQUEST", "Invalid request body")
	}
	if message := validateUserPassword(&req.Password, true); message != "" {
		return badUserRequest(c, "VALIDATION_ERROR", message)
	}

	user, err := h.userRepo.GetByID(c.Context(), id)
	if err != nil || user == nil {
		return userNotFound(c)
	}
	if !user.IsActive {
		return userArchivedConflict(c, "Archived users cannot be changed")
	}
	if ok, err := requireExpectedUpdatedAt(c, req.ExpectedUpdatedAt, user.UpdatedAt); !ok {
		return err
	}

	if middleware.GetUserID(c) != id {
		targetRoleName := roleNameOfUser(user)
		if !h.enforceTargetUserAction(c, targetRoleName, userActionEdit) {
			return nil
		}
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
	if err := h.revokeUserRefreshTokens(c, id, "user_password_changed"); err != nil {
		return err
	}

	user, _ = h.userRepo.GetByID(c.Context(), id)
	userName := userNameOrUnknown(user)
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
	if err := parseStrictUserJSON(c, &req); err != nil {
		return badUserRequest(c, "INVALID_REQUEST", "Invalid request body")
	}
	if req.PIN == nil {
		return badUserRequest(c, "VALIDATION_ERROR", "PIN is required")
	}
	if req.PIN != nil {
		trimmed := strings.TrimSpace(*req.PIN)
		req.PIN = &trimmed
		if message := validateUserPIN(req.PIN, false); message != "" {
			return badUserRequest(c, "VALIDATION_ERROR", message)
		}
	}

	user, err := h.userRepo.GetByID(c.Context(), id)
	if err != nil || user == nil {
		return userNotFound(c)
	}
	if !user.IsActive {
		return userArchivedConflict(c, "Archived users cannot be changed")
	}
	if ok, err := requireExpectedUpdatedAt(c, req.ExpectedUpdatedAt, user.UpdatedAt); !ok {
		return err
	}

	if middleware.GetUserID(c) != id {
		targetRoleName := roleNameOfUser(user)
		if !h.enforceTargetUserAction(c, targetRoleName, userActionEdit) {
			return nil
		}
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
	if err := h.revokeUserRefreshTokens(c, id, "user_pin_changed"); err != nil {
		return err
	}

	user, _ = h.userRepo.GetByID(c.Context(), id)
	userName := userNameOrUnknown(user)
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

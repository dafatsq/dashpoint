package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"dashpoint/backend/internal/audit"
	"dashpoint/backend/internal/auth"
	"dashpoint/backend/internal/middleware"
	"dashpoint/backend/internal/models"
)

const (
	setupMaxJSONBodyBytes = 4096
	setupMinPasswordLen   = 8
	setupOwnerRoleName    = "owner"
)

type setupUserStore interface {
	HasActiveUser(context.Context) (bool, error)
	CreateInitialOwner(context.Context, *models.User) (bool, error)
	EmailExists(context.Context, string, *uuid.UUID) (bool, error)
}

type setupRoleStore interface {
	GetByName(context.Context, string) (*models.Role, error)
}

// SetupHandler provisions the first owner account while the database has no
// active users. Once any active user exists, both endpoints report completed /
// refuse creation.
type SetupHandler struct {
	userRepo setupUserStore
	roleRepo setupRoleStore
}

// NewSetupHandler creates a new setup handler.
func NewSetupHandler(userRepo setupUserStore, roleRepo setupRoleStore) *SetupHandler {
	return &SetupHandler{userRepo: userRepo, roleRepo: roleRepo}
}

// Status handles GET /api/v1/setup/status.
func (h *SetupHandler) Status(c *fiber.Ctx) error {
	hasActiveUser, err := h.userRepo.HasActiveUser(c.Context())
	if err != nil {
		log.Error().Err(err).Msg("Failed to determine initial setup status")
		return middleware.JSONError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "Failed to determine setup status")
	}

	return c.JSON(fiber.Map{"setup_required": !hasActiveUser})
}

// CreateOwnerRequest represents the initial setup request body.
type CreateOwnerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	PIN      string `json:"pin"`
}

// CreateOwner handles POST /api/v1/setup/owner.
func (h *SetupHandler) CreateOwner(c *fiber.Ctx) error {
	var req CreateOwnerRequest
	if err := parseStrictJSONBody(c, &req, setupMaxJSONBodyBytes); err != nil {
		return badUserRequest(c, "INVALID_REQUEST", "Invalid request body")
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.PIN = strings.TrimSpace(req.PIN)

	if message := validateUserName(req.Name, true); message != "" {
		return badUserRequest(c, "VALIDATION_ERROR", message)
	}
	if message := validateSetupEmail(req.Email); message != "" {
		return badUserRequest(c, "VALIDATION_ERROR", message)
	}
	if message := validateSetupPassword(req.Password); message != "" {
		return badUserRequest(c, "VALIDATION_ERROR", message)
	}
	if message := validateUserPIN(&req.PIN, true); message != "" {
		return badUserRequest(c, "VALIDATION_ERROR", message)
	}

	role, err := h.roleRepo.GetByName(c.Context(), setupOwnerRoleName)
	if err != nil {
		log.Error().Err(err).Msg("Failed to look up owner role during initial setup")
		return userInternalError(c, "Failed to complete initial setup")
	}
	if role == nil {
		log.Error().Msg("Owner role not found during initial setup")
		return userInternalError(c, "Failed to complete initial setup")
	}

	emailExists, err := h.userRepo.EmailExists(c.Context(), req.Email, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check email during initial setup")
		return userInternalError(c, "Failed to complete initial setup")
	}
	if emailExists {
		return badUserRequest(c, "EMAIL_EXISTS", "Email is already in use")
	}

	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		log.Error().Err(err).Msg("Failed to hash initial owner password")
		return userInternalError(c, "Failed to complete initial setup")
	}
	pinHash, err := auth.HashPIN(req.PIN)
	if err != nil {
		log.Error().Err(err).Msg("Failed to hash initial owner PIN")
		return userInternalError(c, "Failed to complete initial setup")
	}

	user := &models.User{
		Email:        &req.Email,
		Name:         req.Name,
		PasswordHash: &passwordHash,
		PINHash:      &pinHash,
		RoleID:       role.ID,
		IsActive:     true,
		Role:         role,
	}

	created, err := h.userRepo.CreateInitialOwner(c.Context(), user)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create initial owner account")
		return userInternalError(c, "Failed to complete initial setup")
	}
	if !created {
		return userConflict(c, "SETUP_ALREADY_COMPLETED", "This instance has already been set up. Please sign in.")
	}

	newValues := map[string]interface{}{
		"name":  user.Name,
		"email": req.Email,
		"role":  role.Name,
	}
	audit.LogWithValues(c, models.AuditActionUserCreate, models.AuditEntityUser, user.ID.String(),
		"Created initial owner account: "+user.Name, nil, newValues)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Initial owner account created successfully",
		"user": fiber.Map{
			"id":    user.ID.String(),
			"name":  user.Name,
			"email": req.Email,
		},
	})
}

func validateSetupEmail(email string) string {
	return validateUserEmail(&email, true)
}

func validateSetupPassword(password string) string {
	if message := validateUserPassword(&password, true); message != "" {
		return message
	}
	if len(password) < setupMinPasswordLen {
		return fmt.Sprintf("Password must be at least %d characters long", setupMinPasswordLen)
	}
	return ""
}

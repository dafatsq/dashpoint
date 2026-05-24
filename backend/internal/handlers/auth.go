package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"dashpoint/backend/internal/authz"
	"dashpoint/backend/internal/middleware"
	"dashpoint/backend/internal/models"
	"dashpoint/backend/internal/repository"
)

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	userRepo         authUserReader
	refreshTokenRepo authRefreshTokenStore
	jwtManager       authTokenManager
	workflow         *authWorkflow
}

// NewAuthHandler creates a new auth handler.
func NewAuthHandler(
	userRepo *repository.UserRepository,
	refreshTokenRepo *repository.RefreshTokenRepository,
	jwtManager authTokenManager,
) *AuthHandler {
	return &AuthHandler{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtManager:       jwtManager,
		workflow:         newAuthWorkflow(userRepo, refreshTokenRepo, jwtManager),
	}
}

// Login handles POST /api/v1/auth/login.
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return authInvalidRequest(c)
	}

	if req.Email == "" || req.Password == "" {
		return authValidationError(c, "Email and password are required")
	}

	user, err := h.userRepo.GetByEmail(c.Context(), normalizeLoginEmail(req.Email))
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user by email")
		return authInternalError(c, "An error occurred during login")
	}
	if user == nil {
		logAuthFailure(c, authFailureContext{
			action: models.AuditActionLoginFailed,
			email:  normalizeLoginEmail(req.Email),
			reason: "user_not_found",
		})
		return authUnauthorized(c, "INVALID_CREDENTIALS", "Invalid email or password")
	}
	if !user.IsActive {
		logAuthFailure(c, authFailureContext{
			action:   models.AuditActionLoginFailed,
			userID:   &user.ID,
			email:    normalizeLoginEmail(req.Email),
			userName: user.Name,
			roleName: user.Role.Name,
			reason:   "account_disabled",
		})
		return authUnauthorized(c, "ACCOUNT_DISABLED", "Your account has been disabled")
	}
	if user.PasswordHash == nil || !checkPassword(req.Password, *user.PasswordHash) {
		logAuthFailure(c, authFailureContext{
			action:   models.AuditActionLoginFailed,
			userID:   &user.ID,
			email:    normalizeLoginEmail(req.Email),
			userName: user.Name,
			roleName: user.Role.Name,
			reason:   "invalid_password",
		})
		return authUnauthorized(c, "INVALID_CREDENTIALS", "Invalid email or password")
	}

	return h.workflow.issueAuthResponse(c, user, false)
}

// PINLogin handles POST /api/v1/auth/pin-login.
func (h *AuthHandler) PINLogin(c *fiber.Ctx) error {
	var req PINLoginRequest
	if err := c.BodyParser(&req); err != nil {
		return authInvalidRequest(c)
	}
	if req.UserID == "" {
		return authValidationError(c, "User ID is required")
	}
	if req.PIN == "" {
		return authValidationError(c, "PIN is required")
	}

	userID, err := parseAuthUserID(req.UserID)
	if err != nil {
		log.Error().Err(err).Str("user_id", req.UserID).Msg("Invalid user ID format")
		return authValidationError(c, "Invalid user ID format")
	}

	user, err := h.userRepo.GetByID(c.Context(), userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", req.UserID).Msg("Failed to get user by ID")
		return authUnauthorized(c, "INVALID_CREDENTIALS", "Invalid credentials")
	}
	if user == nil {
		logAuthFailure(c, authFailureContext{
			action: models.AuditActionLoginFailed,
			email:  req.UserID,
			reason: "user_not_found",
		})
		return authUnauthorized(c, "INVALID_CREDENTIALS", "Invalid credentials")
	}
	if !user.IsActive {
		logAuthFailure(c, authFailureContext{
			action:   models.AuditActionLoginFailed,
			userID:   &user.ID,
			email:    req.UserID,
			userName: user.Name,
			roleName: user.Role.Name,
			reason:   "account_disabled",
		})
		return authUnauthorized(c, "ACCOUNT_INACTIVE", "Account is inactive")
	}
	if user.PINHash == nil {
		logAuthFailure(c, authFailureContext{
			action:   models.AuditActionLoginFailed,
			userID:   &user.ID,
			email:    req.UserID,
			userName: user.Name,
			roleName: user.Role.Name,
			reason:   "pin_not_set",
		})
		return authUnauthorized(c, "INVALID_CREDENTIALS", "Invalid credentials")
	}
	if !checkPIN(req.PIN, *user.PINHash) {
		logAuthFailure(c, authFailureContext{
			action:   models.AuditActionLoginFailed,
			userID:   &user.ID,
			email:    req.UserID,
			userName: user.Name,
			roleName: user.Role.Name,
			reason:   "invalid_pin",
		})
		return authUnauthorized(c, "INVALID_CREDENTIALS", "Invalid PIN")
	}

	return h.workflow.issueAuthResponse(c, user, false)
}

// Refresh handles POST /api/v1/auth/refresh.
func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	var req RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return authInvalidRequest(c)
	}
	if req.RefreshToken == "" {
		return authValidationError(c, "Refresh token is required")
	}

	claims, err := h.jwtManager.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return authUnauthorized(c, "INVALID_TOKEN", "Invalid or expired refresh token")
	}

	tokenHash := hashToken(req.RefreshToken)
	storedToken, err := h.refreshTokenRepo.GetByTokenHash(c.Context(), tokenHash)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get refresh token from database")
		return authInternalError(c, "An error occurred during token refresh")
	}
	if storedToken == nil || !storedToken.IsValid() {
		return authUnauthorized(c, "INVALID_TOKEN", "Refresh token has been revoked or expired")
	}

	user, err := h.userRepo.GetByID(c.Context(), claims.UserID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user")
		return authInternalError(c, "An error occurred during token refresh")
	}
	if user == nil || !user.IsActive {
		return authUnauthorized(c, "ACCOUNT_DISABLED", "Your account has been disabled")
	}

	return h.workflow.rotateRefreshToken(c, tokenHash, user)
}

// Logout handles POST /api/v1/auth/logout.
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	var req RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return authInvalidRequest(c)
	}

	if req.RefreshToken != "" {
		if err := h.refreshTokenRepo.Revoke(c.Context(), hashToken(req.RefreshToken), "user_logout"); err != nil {
			log.Error().Err(err).Msg("Failed to revoke refresh token")
		}
	}

	return c.JSON(fiber.Map{
		"message": "Logged out successfully",
	})
}

// Me handles GET /api/v1/me.
func (h *AuthHandler) Me(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return authUnauthorized(c, "UNAUTHORIZED", "Authentication required")
	}

	user, err := h.userRepo.GetByID(c.Context(), userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user")
		return authInternalError(c, "Failed to retrieve user profile")
	}
	if user == nil {
		return middleware.JSONError(c, fiber.StatusNotFound, "USER_NOT_FOUND", "User not found")
	}

	return c.JSON(fiber.Map{
		"user": authUserResponse(user, authz.PermissionsForRole(roleNameOfUser(user)), true),
	})
}

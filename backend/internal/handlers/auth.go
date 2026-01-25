package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"dashpoint/backend/internal/audit"
	"dashpoint/backend/internal/auth"
	"dashpoint/backend/internal/models"
	"dashpoint/backend/internal/repository"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	userRepo         *repository.UserRepository
	refreshTokenRepo *repository.RefreshTokenRepository
	jwtManager       *auth.JWTManager
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(
	userRepo *repository.UserRepository,
	refreshTokenRepo *repository.RefreshTokenRepository,
	jwtManager *auth.JWTManager,
) *AuthHandler {
	return &AuthHandler{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtManager:       jwtManager,
	}
}

// LoginRequest represents the login request body
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// PINLoginRequest represents the PIN login request body
type PINLoginRequest struct {
	UserID string `json:"user_id"`
	PIN    string `json:"pin"`
}

// RefreshRequest represents the refresh token request body
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresAt    time.Time    `json:"expires_at"`
	User         UserResponse `json:"user"`
}

// UserResponse represents the user in auth responses
type UserResponse struct {
	ID          string   `json:"id"`
	Email       *string  `json:"email,omitempty"`
	Name        string   `json:"name"`
	RoleID      string   `json:"role_id"`
	RoleName    string   `json:"role_name"`
	IsActive    bool     `json:"is_active"`
	HasPIN      bool     `json:"has_pin"`
	Permissions []string `json:"permissions"`
	CreatedAt   string   `json:"created_at,omitempty"`
	UpdatedAt   string   `json:"updated_at,omitempty"`
}

// Login handles POST /api/v1/auth/login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request body",
		})
	}

	// Validate input
	if req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "VALIDATION_ERROR",
			"message": "Email and password are required",
		})
	}

	// Find user by email
	user, err := h.userRepo.GetByEmail(c.Context(), req.Email)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user by email")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "An error occurred during login",
		})
	}

	if user == nil {
		// Log failed login attempt
		audit.LogAuth(c, models.AuditActionLoginFailed, nil, req.Email, false, map[string]interface{}{
			"reason": "user_not_found",
		})
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    "INVALID_CREDENTIALS",
			"message": "Invalid email or password",
		})
	}

	// Check if user is active
	if !user.IsActive {
		// Log failed login - account disabled
		audit.LogAuth(c, models.AuditActionLoginFailed, &user.ID, req.Email, false, map[string]interface{}{
			"reason": "account_disabled",
		})
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    "ACCOUNT_DISABLED",
			"message": "Your account has been disabled",
		})
	}

	// Check password
	if user.PasswordHash == nil || !auth.CheckPassword(req.Password, *user.PasswordHash) {
		// Log failed login - wrong password
		audit.LogAuth(c, models.AuditActionLoginFailed, &user.ID, req.Email, false, map[string]interface{}{
			"reason": "invalid_password",
		})
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    "INVALID_CREDENTIALS",
			"message": "Invalid email or password",
		})
	}

	// Generate tokens
	return h.generateAuthResponse(c, user)
}

// PINLogin handles POST /api/v1/auth/pin-login
func (h *AuthHandler) PINLogin(c *fiber.Ctx) error {
	var req PINLoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request body",
		})
	}

	// Validate input
	if req.UserID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "VALIDATION_ERROR",
			"message": "User ID is required",
		})
	}

	if req.PIN == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "VALIDATION_ERROR",
			"message": "PIN is required",
		})
	}

	// Parse user ID
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		log.Error().Err(err).Str("user_id", req.UserID).Msg("Invalid user ID format")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "VALIDATION_ERROR",
			"message": "Invalid user ID format",
		})
	}

	// Get the specific user by ID
	user, err := h.userRepo.GetByID(c.Context(), userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", req.UserID).Msg("Failed to get user by ID")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    "INVALID_CREDENTIALS",
			"message": "Invalid credentials",
		})
	}

	// Check if user is active
	if !user.IsActive {
		log.Warn().Str("user_id", req.UserID).Msg("Attempted PIN login for inactive user")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    "ACCOUNT_INACTIVE",
			"message": "Account is inactive",
		})
	}

	// Verify the PIN belongs to this specific user
	if user.PINHash == nil {
		log.Warn().Str("user_id", req.UserID).Msg("Attempted PIN login for user without PIN")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    "INVALID_CREDENTIALS",
			"message": "Invalid credentials",
		})
	}

	if !auth.CheckPIN(req.PIN, *user.PINHash) {
		log.Warn().Str("user_id", req.UserID).Msg("Invalid PIN provided")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    "INVALID_CREDENTIALS",
			"message": "Invalid PIN",
		})
	}

	// Generate tokens
	return h.generateAuthResponse(c, user)
}

// Refresh handles POST /api/v1/auth/refresh
func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	var req RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request body",
		})
	}

	if req.RefreshToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "VALIDATION_ERROR",
			"message": "Refresh token is required",
		})
	}

	// Validate the refresh token JWT
	claims, err := h.jwtManager.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    "INVALID_TOKEN",
			"message": "Invalid or expired refresh token",
		})
	}

	// Check if the refresh token is in the database and not revoked
	tokenHash := auth.HashToken(req.RefreshToken)
	storedToken, err := h.refreshTokenRepo.GetByTokenHash(c.Context(), tokenHash)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get refresh token from database")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "An error occurred during token refresh",
		})
	}

	if storedToken == nil || !storedToken.IsValid() {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    "INVALID_TOKEN",
			"message": "Refresh token has been revoked or expired",
		})
	}

	// Get the user
	user, err := h.userRepo.GetByID(c.Context(), claims.UserID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "An error occurred during token refresh",
		})
	}

	if user == nil || !user.IsActive {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    "ACCOUNT_DISABLED",
			"message": "Your account has been disabled",
		})
	}

	// Revoke the old refresh token
	if err := h.refreshTokenRepo.Revoke(c.Context(), tokenHash, "token_refresh"); err != nil {
		log.Error().Err(err).Msg("Failed to revoke old refresh token")
	}

	// Generate new tokens
	return h.generateAuthResponse(c, user)
}

// Logout handles POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	var req RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request body",
		})
	}

	if req.RefreshToken != "" {
		// Revoke the specific refresh token
		tokenHash := auth.HashToken(req.RefreshToken)
		if err := h.refreshTokenRepo.Revoke(c.Context(), tokenHash, "user_logout"); err != nil {
			log.Error().Err(err).Msg("Failed to revoke refresh token")
		}
	}

	return c.JSON(fiber.Map{
		"message": "Logged out successfully",
	})
}

// Me handles GET /api/v1/me
// Returns the currently authenticated user's full profile
func (h *AuthHandler) Me(c *fiber.Ctx) error {
	// Get user ID from claims (set by auth middleware)
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code":    "UNAUTHORIZED",
			"message": "Authentication required",
		})
	}

	// Fetch full user details
	user, err := h.userRepo.GetByID(c.Context(), userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to retrieve user profile",
		})
	}

	if user == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"code":    "USER_NOT_FOUND",
			"message": "User not found",
		})
	}

	// Get user permissions
	permissions, err := h.userRepo.GetUserPermissions(c.Context(), user.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user permissions")
		permissions = []string{}
	}

	return c.JSON(fiber.Map{
		"user": UserResponse{
			ID:          user.ID.String(),
			Email:       user.Email,
			Name:        user.Name,
			RoleID:      user.RoleID.String(),
			RoleName:    user.Role.Name,
			IsActive:    user.IsActive,
			HasPIN:      user.PINHash != nil,
			Permissions: permissions,
			CreatedAt:   user.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   user.UpdatedAt.Format(time.RFC3339),
		},
	})
}

// generateAuthResponse generates tokens and returns the auth response
func (h *AuthHandler) generateAuthResponse(c *fiber.Ctx, user *models.User) error {
	// Generate token pair
	tokenPair, err := h.jwtManager.GenerateTokenPair(
		user.ID,
		func() string {
			if user.Email != nil {
				return *user.Email
			}
			return ""
		}(),
		user.RoleID,
		user.Role.Name,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate tokens")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "An error occurred during login",
		})
	}

	// Store the refresh token hash in the database
	refreshTokenHash := auth.HashToken(tokenPair.RefreshToken)
	refreshTokenRecord := &models.RefreshToken{
		UserID:    user.ID,
		TokenHash: refreshTokenHash,
		ExpiresAt: tokenPair.RefreshTokenExpiresAt,
	}

	if err := h.refreshTokenRepo.Create(c.Context(), refreshTokenRecord); err != nil {
		log.Error().Err(err).Msg("Failed to store refresh token")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "An error occurred during login",
		})
	}

	// Update last login
	if err := h.userRepo.UpdateLastLogin(c.Context(), user.ID); err != nil {
		log.Error().Err(err).Msg("Failed to update last login")
		// Don't fail the request for this
	}

	// Get user permissions
	permissions, err := h.userRepo.GetUserPermissions(c.Context(), user.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user permissions")
		// Don't fail login, just use empty permissions
		permissions = []string{}
	}

	// Log successful login
	email := ""
	if user.Email != nil {
		email = *user.Email
	}
	audit.LogAuth(c, models.AuditActionLogin, &user.ID, email, true, map[string]interface{}{
		"role": user.Role.Name,
	})

	return c.JSON(AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.AccessTokenExpiresAt,
		User: UserResponse{
			ID:          user.ID.String(),
			Email:       user.Email,
			Name:        user.Name,
			RoleID:      user.RoleID.String(),
			RoleName:    user.Role.Name,
			IsActive:    user.IsActive,
			HasPIN:      user.PINHash != nil,
			Permissions: permissions,
		},
	})
}

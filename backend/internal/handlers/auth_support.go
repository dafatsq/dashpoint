package handlers

import (
	"context"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"dashpoint/backend/internal/auth"
	"dashpoint/backend/internal/middleware"
	"dashpoint/backend/internal/models"
)

type authUserReader interface {
	GetByID(context.Context, uuid.UUID) (*models.User, error)
	GetByEmail(context.Context, string) (*models.User, error)
	UpdateLastLogin(context.Context, uuid.UUID) error
}

type authRefreshTokenStore interface {
	Create(context.Context, *models.RefreshToken) error
	GetByTokenHash(context.Context, string) (*models.RefreshToken, error)
	Revoke(context.Context, string, string) error
	Rotate(context.Context, string, string, *models.RefreshToken) error
}

type authTokenManager interface {
	GenerateTokenPair(uuid.UUID, string, string, uuid.UUID, string) (*auth.TokenPair, error)
	ValidateAccessToken(string) (*auth.Claims, error)
	ValidateRefreshToken(string) (*auth.Claims, error)
}

// LoginRequest represents the login request body.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// PINLoginRequest represents the PIN login request body.
type PINLoginRequest struct {
	UserID string `json:"user_id"`
	PIN    string `json:"pin"`
}

// RefreshRequest represents the refresh token request body.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// AuthResponse represents the authentication response.
type AuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresAt    time.Time    `json:"expires_at"`
	User         UserResponse `json:"user"`
}

// UserResponse represents the user in auth responses.
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

type authFailureContext struct {
	action   models.AuditAction
	userID   *uuid.UUID
	email    string
	userName string
	roleName string
	reason   string
}

func authInvalidRequest(c *fiber.Ctx) error {
	return middleware.JSONError(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
}

func authValidationError(c *fiber.Ctx, message string) error {
	return middleware.JSONError(c, fiber.StatusBadRequest, "VALIDATION_ERROR", message)
}

func authUnauthorized(c *fiber.Ctx, code, message string) error {
	return middleware.JSONError(c, fiber.StatusUnauthorized, code, message)
}

func authInternalError(c *fiber.Ctx, message string) error {
	return middleware.JSONError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", message)
}

func normalizeLoginEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func authUserResponse(user *models.User, permissions []string, includeTimestamps bool) UserResponse {
	response := UserResponse{
		ID:          user.ID.String(),
		Email:       user.Email,
		Name:        user.Name,
		RoleID:      user.RoleID.String(),
		IsActive:    user.IsActive,
		HasPIN:      user.PINHash != nil,
		Permissions: permissions,
	}

	if user.Role != nil {
		response.RoleName = user.Role.Name
	}

	if includeTimestamps {
		response.CreatedAt = user.CreatedAt.Format(time.RFC3339)
		response.UpdatedAt = user.UpdatedAt.Format(time.RFC3339)
	}

	return response
}

func checkPassword(password, hash string) bool {
	return auth.CheckPassword(password, hash)
}

func checkPIN(pin, hash string) bool {
	return auth.CheckPIN(pin, hash)
}

func hashToken(token string) string {
	return auth.HashToken(token)
}

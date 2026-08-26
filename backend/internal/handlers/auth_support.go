package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"dashpoint/backend/internal/auth"
	"dashpoint/backend/internal/middleware"
	"dashpoint/backend/internal/models"
	"dashpoint/backend/internal/repository"
)

type authUserReader interface {
	GetByID(context.Context, uuid.UUID) (*models.User, error)
	GetByEmail(context.Context, string) (*models.User, error)
	GetUserPermissions(context.Context, uuid.UUID) ([]string, error)
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

const (
	authMaxJSONBodyBytes = 4096
	refreshTokenCookie   = "refresh_token"
	refreshTokenPath     = "/api/v1/auth"
	wailsCustomOrigin    = "wails://wails"
	wailsWindowsOrigin   = "http://wails.localhost"
)

var errEmptyAuthBody = errors.New("empty auth request body")

// LoginRequest represents the login request body.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	// RememberMe opts the browser session into a persistent refresh cookie.
	// When explicitly false the cookie is session-scoped and disappears when
	// the browser closes. Absent keeps the historical persistent default so
	// existing clients (desktop, older web bundles) are unaffected.
	RememberMe *bool `json:"remember_me"`
}

// PINLoginRequest represents the PIN login request body.
type PINLoginRequest struct {
	UserID string `json:"user_id"`
	PIN    string `json:"pin"`
}

// AuthResponse represents the authentication response.
type AuthResponse struct {
	AccessToken string       `json:"access_token"`
	ExpiresAt   time.Time    `json:"expires_at"`
	User        UserResponse `json:"user"`
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

func parseStrictAuthJSON(c *fiber.Ctx, dest interface{}, allowEmpty bool) error {
	body := bytes.TrimSpace(c.Body())
	if len(body) == 0 {
		if allowEmpty {
			return nil
		}
		return errEmptyAuthBody
	}
	if len(body) > authMaxJSONBodyBytes {
		return errors.New("auth request body is too large")
	}

	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dest); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return errors.New("auth request body must contain a single JSON object")
	}
	return nil
}

func refreshTokenFromCookie(c *fiber.Ctx) string {
	return strings.TrimSpace(c.Cookies(refreshTokenCookie))
}

func refreshCookieSameSite(c *fiber.Ctx) string {
	origin := strings.TrimSpace(c.Get("Origin"))
	if strings.EqualFold(origin, wailsCustomOrigin) || strings.EqualFold(origin, wailsWindowsOrigin) {
		return fiber.CookieSameSiteNoneMode
	}

	return fiber.CookieSameSiteStrictMode
}

func setRefreshTokenCookie(c *fiber.Ctx, token string, expiresAt time.Time, sessionScoped bool) {
	cookie := &fiber.Cookie{
		Name:     refreshTokenCookie,
		Value:    token,
		Path:     refreshTokenPath,
		Expires:  expiresAt,
		HTTPOnly: true,
		Secure:   isSecureRequest(c),
		SameSite: refreshCookieSameSite(c),
	}
	if sessionScoped {
		// No Expires/MaxAge attributes: the browser drops the cookie when the
		// session ends — "remember me off" semantics for memory-only tokens.
		cookie.Expires = time.Time{}
	}
	c.Cookie(cookie)
}

func clearRefreshTokenCookie(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     refreshTokenCookie,
		Value:    "",
		Path:     refreshTokenPath,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HTTPOnly: true,
		Secure:   isSecureRequest(c),
		SameSite: refreshCookieSameSite(c),
	})
}

func isSecureRequest(c *fiber.Ctx) bool {
	return c.Protocol() == "https"
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

// dummyPasswordHash is a precomputed cost-12 bcrypt hash used to equalize
// login timing for unknown emails: without it, missing accounts skip the
// bcrypt round-trip and respond measurably faster than existing ones.
var dummyPasswordHash = func() string {
	hash, err := auth.HashPassword("dashpoint-dummy-password-equalizer")
	if err != nil {
		panic("failed to seed dummy password hash: " + err.Error())
	}
	return hash
}()

func checkPIN(pin, hash string) bool {
	return auth.CheckPIN(pin, hash)
}

func hashToken(token string) string {
	return auth.HashToken(token)
}

// asSanitizedSaleError lets domain-validation messages (insufficient stock,
// price mismatches, payment mismatches) reach the client verbatim, while
// infrastructure failures are collapsed to ok=false so their internals stay
// server-side.
func asSanitizedSaleError(err error) (string, bool) {
	var internal *repository.InternalError
	if errors.As(err, &internal) {
		return "", false
	}
	return err.Error(), true
}

// asSanitizedExpenseError mirrors asSanitizedSaleError for the expense
// domain: user-facing validation text passes through, infrastructure
// internals stay server-side.
func asSanitizedExpenseError(err error) (string, bool) {
	var internal *repository.InternalError
	if errors.As(err, &internal) {
		return "", false
	}
	return err.Error(), true
}

package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"dashpoint/backend/internal/auth"
	"dashpoint/backend/internal/repository"
)

// AuthMiddleware creates an authentication middleware
func AuthMiddleware(jwtManager *auth.JWTManager, userRepo *repository.UserRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get the Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return JSONError(c, fiber.StatusUnauthorized, "MISSING_TOKEN", "Authorization header is required")
		}

		// Check for Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return JSONError(c, fiber.StatusUnauthorized, "INVALID_TOKEN_FORMAT", "Authorization header must be in format: Bearer <token>")
		}

		tokenString := parts[1]

		// Validate the token
		claims, err := jwtManager.ValidateAccessToken(tokenString)
		if err != nil {
			return JSONError(c, fiber.StatusUnauthorized, "INVALID_TOKEN", "Invalid or expired access token")
		}

		// Verify user is still active
		user, err := userRepo.GetByID(c.Context(), claims.UserID)
		if err != nil {
			log.Error().Err(err).Msg("Failed to verify user status")
			return JSONError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "Failed to verify authentication")
		}

		if user == nil || !user.IsActive {
			return JSONError(c, fiber.StatusUnauthorized, "ACCOUNT_INACTIVE", "Your account has been deactivated")
		}

		roleName := claims.RoleName
		if user.Role != nil {
			roleName = user.Role.Name
		}

		// Store current database-backed identity in context.
		c.Locals("user_id", claims.UserID)
		c.Locals("email", user.Email)
		c.Locals("role_id", user.RoleID)
		c.Locals("role_name", roleName)
		c.Locals("claims", claims)

		return c.Next()
	}
}

// GetUserID retrieves the user ID from the context
func GetUserID(c *fiber.Ctx) uuid.UUID {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return uuid.Nil
	}
	return userID
}

// GetRoleName retrieves the role name from the context
func GetRoleName(c *fiber.Ctx) string {
	roleName, ok := c.Locals("role_name").(string)
	if !ok {
		return ""
	}
	return roleName
}

// GetClaims retrieves the full claims from the context
func GetClaims(c *fiber.Ctx) *auth.Claims {
	claims, ok := c.Locals("claims").(*auth.Claims)
	if !ok {
		return nil
	}
	return claims
}

// RequireRole creates a middleware that requires a specific role
func RequireRole(roles ...string) fiber.Handler {
	roleSet := make(map[string]bool)
	for _, role := range roles {
		roleSet[role] = true
	}

	return func(c *fiber.Ctx) error {
		roleName := GetRoleName(c)
		if roleName == "" {
			return JSONError(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		}

		if !roleSet[roleName] {
			return JSONError(c, fiber.StatusForbidden, "FORBIDDEN", "You do not have permission to access this resource")
		}

		return c.Next()
	}
}

// RequirePermission creates a middleware factory that checks for specific permissions
// Note: This requires fetching permissions from the database, so it takes a permission checker function
type PermissionChecker func(c *fiber.Ctx, userID uuid.UUID, permission string) (bool, error)

func RequirePermission(checker PermissionChecker, permissions ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := GetUserID(c)
		if userID == uuid.Nil {
			return JSONError(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		}

		// Check each required permission
		for _, perm := range permissions {
			hasPermission, err := checker(c, userID, perm)
			if err != nil {
				return JSONError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "Failed to check permissions")
			}

			if !hasPermission {
				return JSONError(c, fiber.StatusForbidden, "FORBIDDEN", "You do not have the required permission: "+perm)
			}
		}

		return c.Next()
	}
}

// RequirePermissionOrSelfParam allows users to operate on their own record while
// preserving permission checks for every other target record.
func RequirePermissionOrSelfParam(checker PermissionChecker, paramName string, permissions ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := GetUserID(c)
		if userID == uuid.Nil {
			return JSONError(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		}

		targetID, err := uuid.Parse(c.Params(paramName))
		if err == nil && targetID == userID {
			return c.Next()
		}

		for _, perm := range permissions {
			hasPermission, err := checker(c, userID, perm)
			if err != nil {
				return JSONError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "Failed to check permissions")
			}

			if !hasPermission {
				return JSONError(c, fiber.StatusForbidden, "FORBIDDEN", "You do not have the required permission: "+perm)
			}
		}

		return c.Next()
	}
}

// RequireAnyPermission creates a middleware factory that checks if any of the specified permissions match
func RequireAnyPermission(checker PermissionChecker, permissions ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := GetUserID(c)
		if userID == uuid.Nil {
			return JSONError(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		}

		// Check if any of the permissions are granted
		for _, perm := range permissions {
			hasPermission, err := checker(c, userID, perm)
			if err != nil {
				return JSONError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "Failed to check permissions")
			}

			if hasPermission {
				return c.Next()
			}
		}

		return JSONError(c, fiber.StatusForbidden, "FORBIDDEN", "You do not have any of the required permissions")
	}
}

package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"dashpoint/backend/internal/audit"
	"dashpoint/backend/internal/auth"
	"dashpoint/backend/internal/authz"
	"dashpoint/backend/internal/models"
)

type authWorkflow struct {
	userRepo   authUserReader
	tokenStore authRefreshTokenStore
	jwtManager authTokenManager
}

func newAuthWorkflow(userRepo authUserReader, tokenStore authRefreshTokenStore, jwtManager authTokenManager) *authWorkflow {
	return &authWorkflow{
		userRepo:   userRepo,
		tokenStore: tokenStore,
		jwtManager: jwtManager,
	}
}

func (w *authWorkflow) issueAuthResponse(c *fiber.Ctx, user *models.User, isRefresh bool) error {
	tokenPair, err := w.jwtManager.GenerateTokenPair(
		user.ID,
		func() string {
			if user.Email != nil {
				return *user.Email
			}
			return ""
		}(),
		user.Name,
		user.RoleID,
		user.Role.Name,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate tokens")
		return authInternalError(c, "An error occurred during login")
	}

	refreshTokenRecord := &models.RefreshToken{
		UserID:    user.ID,
		TokenHash: auth.HashToken(tokenPair.RefreshToken),
		ExpiresAt: tokenPair.RefreshTokenExpiresAt,
	}

	if err := w.tokenStore.Create(c.Context(), refreshTokenRecord); err != nil {
		log.Error().Err(err).Msg("Failed to store refresh token")
		return authInternalError(c, "An error occurred during login")
	}

	return w.finishAuthResponse(c, user, tokenPair, isRefresh)
}

func (w *authWorkflow) rotateRefreshToken(c *fiber.Ctx, currentHash string, user *models.User) error {
	tokenPair, err := w.jwtManager.GenerateTokenPair(
		user.ID,
		func() string {
			if user.Email != nil {
				return *user.Email
			}
			return ""
		}(),
		user.Name,
		user.RoleID,
		user.Role.Name,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate tokens")
		return authInternalError(c, "An error occurred during login")
	}

	refreshTokenRecord := &models.RefreshToken{
		UserID:    user.ID,
		TokenHash: auth.HashToken(tokenPair.RefreshToken),
		ExpiresAt: tokenPair.RefreshTokenExpiresAt,
	}

	if err := w.tokenStore.Rotate(c.Context(), currentHash, "token_refresh", refreshTokenRecord); err != nil {
		log.Error().Err(err).Msg("Failed to rotate refresh token")
		return authInternalError(c, "An error occurred during token refresh")
	}

	return w.finishAuthResponse(c, user, tokenPair, true)
}

func (w *authWorkflow) finishAuthResponse(c *fiber.Ctx, user *models.User, tokenPair *auth.TokenPair, isRefresh bool) error {
	if !isRefresh {
		if err := w.userRepo.UpdateLastLogin(c.Context(), user.ID); err != nil {
			log.Error().Err(err).Msg("Failed to update last login")
		}
	}

	if !isRefresh {
		email := ""
		if user.Email != nil {
			email = *user.Email
		}
		audit.LogAuth(c, models.AuditActionLogin, &user.ID, email, user.Name, user.Role.Name, true, map[string]interface{}{
			"role": user.Role.Name,
		})
	}

	return c.JSON(AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.AccessTokenExpiresAt,
		User:         authUserResponse(user, authz.PermissionsForRole(roleNameOfUser(user)), false),
	})
}

func logAuthFailure(c *fiber.Ctx, ctx authFailureContext) {
	metadata := map[string]interface{}{
		"reason": ctx.reason,
	}
	audit.LogAuth(c, ctx.action, ctx.userID, ctx.email, ctx.userName, ctx.roleName, false, metadata)
}

func parseAuthUserID(id string) (uuid.UUID, error) {
	userID, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user ID format: %w", err)
	}
	return userID, nil
}

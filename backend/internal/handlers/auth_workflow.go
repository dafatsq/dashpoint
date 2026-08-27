package handlers

import (
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"dashpoint/backend/internal/audit"
	"dashpoint/backend/internal/auth"
	"dashpoint/backend/internal/models"
	"dashpoint/backend/internal/repository"
)

type authWorkflow struct {
	userRepo   authUserReader
	tokenStore authRefreshTokenStore
	jwtManager authTokenManager
}

const (
	// refreshReuseGraceWindow is how long a just-rotated refresh token may
	// still be exchanged when a sibling tab presents it mid-rotation. Without
	// this, reloading one tab logs out its sibling (memory-only access tokens
	// make every page load refresh).
	refreshReuseGraceWindow = 30 * time.Second
	// tokenRefreshRevocationReason must match the reason the repository uses
	// during normal rotation; only rotation revocations are grace-eligible.
	tokenRefreshRevocationReason = "token_refresh"
	// tokenFamilyReuseRevocationReason marks tokens killed because a revoked
	// sibling in their family was replayed beyond the rotation grace window.
	tokenFamilyReuseRevocationReason = "token_family_reuse"
)

func newAuthWorkflow(userRepo authUserReader, tokenStore authRefreshTokenStore, jwtManager authTokenManager) *authWorkflow {
	return &authWorkflow{
		userRepo:   userRepo,
		tokenStore: tokenStore,
		jwtManager: jwtManager,
	}
}

// issueAuthResponse mints a fresh token pair. A zero familyID starts a new
// refresh-token family (login paths); an existing familyID continues one
// (grace-window sibling continuation).
func (w *authWorkflow) issueAuthResponse(c *fiber.Ctx, user *models.User, isRefresh bool, rememberMe *bool, familyID uuid.UUID) error {
	if familyID == uuid.Nil {
		familyID = uuid.New()
	}

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
		FamilyID:  familyID,
	}

	if err := w.tokenStore.Create(c.Context(), refreshTokenRecord); err != nil {
		log.Error().Err(err).Msg("Failed to store refresh token")
		return authInternalError(c, "An error occurred during login")
	}

	return w.finishAuthResponse(c, user, tokenPair, isRefresh, rememberMe)
}

// isRecentlyRotatedToken reports whether a no-longer-active refresh token was
// revoked by a normal rotation inside the reuse-grace window, meaning a
// sibling tab likely rotated it moments ago.
func isRecentlyRotatedToken(token *models.RefreshToken) bool {
	if token == nil || token.RevokedAt == nil || token.RevokedReason == nil {
		return false
	}
	return *token.RevokedReason == tokenRefreshRevocationReason &&
		time.Since(*token.RevokedAt) <= refreshReuseGraceWindow
}

// rotateRefreshToken replaces the presented token with a new pair in the same
// refresh-token family.
func (w *authWorkflow) rotateRefreshToken(c *fiber.Ctx, currentHash string, familyID uuid.UUID, user *models.User, rememberMe *bool) error {
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
		FamilyID:  familyID,
	}

	if err := w.tokenStore.Rotate(c.Context(), currentHash, tokenRefreshRevocationReason, refreshTokenRecord); err != nil {
		if errors.Is(err, repository.ErrRefreshTokenNotActive) {
			return authUnauthorized(c, "INVALID_TOKEN", "Refresh token has been revoked or expired")
		}
		log.Error().Err(err).Msg("Failed to rotate refresh token")
		return authInternalError(c, "An error occurred during token refresh")
	}

	return w.finishAuthResponse(c, user, tokenPair, true, rememberMe)
}

func (w *authWorkflow) finishAuthResponse(c *fiber.Ctx, user *models.User, tokenPair *auth.TokenPair, isRefresh bool, rememberMe *bool) error {
	responseUser := user
	if !isRefresh {
		if err := w.userRepo.UpdateLastLogin(c.Context(), user.ID); err != nil {
			log.Error().Err(err).Msg("Failed to update last login")
		} else if updatedUser, err := w.userRepo.GetByID(c.Context(), user.ID); err != nil {
			log.Error().Err(err).Msg("Failed to refresh user after login")
		} else if updatedUser != nil {
			responseUser = updatedUser
		}
	}

	if !isRefresh {
		audit.LogAuth(c, models.AuditActionLogin, &user.ID, user.Name, user.Role.Name, true, map[string]interface{}{
			"role": user.Role.Name,
		})
	}

	permissions, err := w.userRepo.GetUserPermissions(c.Context(), user.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user permissions")
		return authInternalError(c, "Failed to retrieve user permissions")
	}

	sessionScoped := rememberMe != nil && !*rememberMe
	setRefreshTokenCookie(c, tokenPair.RefreshToken, tokenPair.RefreshTokenExpiresAt, sessionScoped)

	return c.JSON(AuthResponse{
		AccessToken: tokenPair.AccessToken,
		ExpiresAt:   tokenPair.AccessTokenExpiresAt,
		User:        authUserResponse(responseUser, permissions, true),
	})
}

func logAuthFailure(c *fiber.Ctx, ctx authFailureContext) {
	metadata := map[string]interface{}{
		"reason": ctx.reason,
	}
	audit.LogAuth(c, ctx.action, ctx.userID, ctx.userName, ctx.roleName, false, metadata)
}

func parseAuthUserID(id string) (uuid.UUID, error) {
	userID, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user ID format: %w", err)
	}
	return userID, nil
}

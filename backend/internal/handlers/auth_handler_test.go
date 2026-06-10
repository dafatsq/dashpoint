package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	authpkg "dashpoint/backend/internal/auth"
	"dashpoint/backend/internal/models"
)

type fakeAuthUserRepo struct {
	userByEmail          *models.User
	userByID             *models.User
	permissions          []string
	getByEmailErr        error
	getByIDErr           error
	updateLastLoginErr   error
	updateLastLoginCalls int
}

func (f *fakeAuthUserRepo) GetByID(context.Context, uuid.UUID) (*models.User, error) {
	return f.userByID, f.getByIDErr
}

func (f *fakeAuthUserRepo) GetByEmail(context.Context, string) (*models.User, error) {
	return f.userByEmail, f.getByEmailErr
}

func (f *fakeAuthUserRepo) GetUserPermissions(context.Context, uuid.UUID) ([]string, error) {
	if f.permissions != nil {
		return f.permissions, nil
	}
	return []string{"access_pos_page"}, nil
}

func (f *fakeAuthUserRepo) UpdateLastLogin(context.Context, uuid.UUID) error {
	f.updateLastLoginCalls++
	return f.updateLastLoginErr
}

type fakeRefreshTokenStore struct {
	stored          *models.RefreshToken
	getErr          error
	createErr       error
	revokeErr       error
	rotateErr       error
	created         *models.RefreshToken
	revokedHash     string
	revokeReason    string
	rotatedOldHash  string
	rotatedReason   string
	rotatedNewToken *models.RefreshToken
}

func (f *fakeRefreshTokenStore) Create(_ context.Context, token *models.RefreshToken) error {
	f.created = token
	return f.createErr
}

func (f *fakeRefreshTokenStore) GetByTokenHash(context.Context, string) (*models.RefreshToken, error) {
	return f.stored, f.getErr
}

func (f *fakeRefreshTokenStore) Revoke(_ context.Context, tokenHash string, reason string) error {
	f.revokedHash = tokenHash
	f.revokeReason = reason
	return f.revokeErr
}

func (f *fakeRefreshTokenStore) Rotate(_ context.Context, oldHash, reason string, replacement *models.RefreshToken) error {
	f.rotatedOldHash = oldHash
	f.rotatedReason = reason
	f.rotatedNewToken = replacement
	return f.rotateErr
}

type fakeJWTManager struct {
	validateAccessClaims  *authpkg.Claims
	validateRefreshClaims *authpkg.Claims
	validateAccessErr     error
	validateRefreshErr    error
	tokenPair             *authpkg.TokenPair
	generateErr           error
}

func (f *fakeJWTManager) GenerateTokenPair(uuid.UUID, string, string, uuid.UUID, string) (*authpkg.TokenPair, error) {
	return f.tokenPair, f.generateErr
}

func (f *fakeJWTManager) ValidateAccessToken(string) (*authpkg.Claims, error) {
	return f.validateAccessClaims, f.validateAccessErr
}

func (f *fakeJWTManager) ValidateRefreshToken(string) (*authpkg.Claims, error) {
	return f.validateRefreshClaims, f.validateRefreshErr
}

func TestAuthHandlerLoginSuccess(t *testing.T) {
	roleID := uuid.New()
	userID := uuid.New()
	passwordHash, err := authpkg.HashPassword("secret123")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	email := "owner@example.com"
	postLoginUser := &models.User{
		ID:           userID,
		Email:        &email,
		Name:         "Owner",
		PasswordHash: &passwordHash,
		RoleID:       roleID,
		IsActive:     true,
		CreatedAt:    time.Unix(100, 0),
		UpdatedAt:    time.Unix(250, 0),
		Role:         &models.Role{ID: roleID, Name: "owner"},
	}
	userRepo := &fakeAuthUserRepo{
		userByEmail: &models.User{
			ID:           userID,
			Email:        &email,
			Name:         "Owner",
			PasswordHash: &passwordHash,
			RoleID:       roleID,
			IsActive:     true,
			CreatedAt:    time.Unix(100, 0),
			UpdatedAt:    time.Unix(200, 0),
			Role:         &models.Role{ID: roleID, Name: "owner"},
		},
		userByID: postLoginUser,
	}

	tokenStore := &fakeRefreshTokenStore{}
	jwtManager := &fakeJWTManager{
		tokenPair: &authpkg.TokenPair{
			AccessToken:           "access-token",
			RefreshToken:          "refresh-token",
			AccessTokenExpiresAt:  time.Unix(300, 0),
			RefreshTokenExpiresAt: time.Unix(400, 0),
		},
	}

	handler := &AuthHandler{
		userRepo:         userRepo,
		refreshTokenRepo: tokenStore,
		jwtManager:       jwtManager,
		workflow:         newAuthWorkflow(userRepo, tokenStore, jwtManager),
	}

	app := fiber.New()
	app.Post("/login", handler.Login)

	req := httptest.NewRequest("POST", "/login", bytes.NewBufferString(`{"email":"OWNER@example.com","password":"secret123"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	if userRepo.updateLastLoginCalls != 1 {
		t.Fatalf("expected UpdateLastLogin to be called once, got %d", userRepo.updateLastLoginCalls)
	}
	if tokenStore.created == nil || tokenStore.created.TokenHash == "" {
		t.Fatalf("expected refresh token to be stored")
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll returned error: %v", err)
	}

	var body AuthResponse
	if err := json.Unmarshal(bodyBytes, &body); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	if body.AccessToken != "access-token" {
		t.Fatalf("expected access token in response")
	}
	if body.User.RoleName != "owner" {
		t.Fatalf("expected role name owner, got %q", body.User.RoleName)
	}
	if len(body.User.Permissions) == 0 {
		t.Fatalf("expected role-derived permissions in response")
	}
	if body.User.UpdatedAt != postLoginUser.UpdatedAt.Format(time.RFC3339) {
		t.Fatalf("expected post-login updated_at %q, got %q", postLoginUser.UpdatedAt.Format(time.RFC3339), body.User.UpdatedAt)
	}

	var rawBody map[string]any
	if err := json.Unmarshal(bodyBytes, &rawBody); err != nil {
		t.Fatalf("Unmarshal raw body returned error: %v", err)
	}
	if _, exists := rawBody["refresh_token"]; exists {
		t.Fatalf("did not expect refresh_token in auth response")
	}
	if cookie := findCookie(resp.Cookies(), refreshTokenCookie); cookie == nil || cookie.Value == "" || !cookie.HttpOnly {
		t.Fatalf("expected HttpOnly refresh token cookie")
	}
}

func TestAuthHandlerLoginRejectsUnknownFields(t *testing.T) {
	handler := &AuthHandler{}
	app := fiber.New()
	app.Post("/login", handler.Login)

	req := httptest.NewRequest("POST", "/login", bytes.NewBufferString(`{"email":"owner@example.com","password":"secret123","refresh_token":"bad"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestAuthHandlerPINLoginRejectsInvalidUserID(t *testing.T) {
	handler := &AuthHandler{}
	app := fiber.New()
	app.Post("/pin-login", handler.PINLogin)

	req := httptest.NewRequest("POST", "/pin-login", bytes.NewBufferString(`{"user_id":"not-a-uuid","pin":"1234"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestAuthHandlerRefreshRotatesToken(t *testing.T) {
	roleID := uuid.New()
	userID := uuid.New()
	email := "manager@example.com"
	currentToken := "refresh-token"
	currentHash := authpkg.HashToken(currentToken)

	userRepo := &fakeAuthUserRepo{
		userByID: &models.User{
			ID:        userID,
			Email:     &email,
			Name:      "Manager",
			RoleID:    roleID,
			IsActive:  true,
			CreatedAt: time.Unix(100, 0),
			UpdatedAt: time.Unix(200, 0),
			Role:      &models.Role{ID: roleID, Name: "manager"},
		},
	}
	tokenStore := &fakeRefreshTokenStore{
		stored: &models.RefreshToken{
			UserID:    userID,
			TokenHash: currentHash,
			ExpiresAt: time.Now().Add(time.Hour),
		},
	}
	jwtManager := &fakeJWTManager{
		validateRefreshClaims: &authpkg.Claims{UserID: userID},
		tokenPair: &authpkg.TokenPair{
			AccessToken:           "new-access",
			RefreshToken:          "new-refresh",
			AccessTokenExpiresAt:  time.Unix(300, 0),
			RefreshTokenExpiresAt: time.Unix(400, 0),
		},
	}
	handler := &AuthHandler{
		userRepo:         userRepo,
		refreshTokenRepo: tokenStore,
		jwtManager:       jwtManager,
		workflow:         newAuthWorkflow(userRepo, tokenStore, jwtManager),
	}

	app := fiber.New()
	app.Post("/refresh", handler.Refresh)

	req := httptest.NewRequest("POST", "/refresh", nil)
	req.AddCookie(&http.Cookie{Name: refreshTokenCookie, Value: currentToken})
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	if tokenStore.rotatedOldHash != currentHash {
		t.Fatalf("expected old token hash %q, got %q", currentHash, tokenStore.rotatedOldHash)
	}
	if tokenStore.rotatedReason != "token_refresh" {
		t.Fatalf("expected token_refresh reason, got %q", tokenStore.rotatedReason)
	}
	if cookie := findCookie(resp.Cookies(), refreshTokenCookie); cookie == nil || cookie.Value == "" || !cookie.HttpOnly {
		t.Fatalf("expected rotated HttpOnly refresh token cookie")
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll returned error: %v", err)
	}
	var body AuthResponse
	if err := json.Unmarshal(bodyBytes, &body); err != nil {
		t.Fatalf("Unmarshal returned error: %v", err)
	}
	expectedUpdatedAt := userRepo.userByID.UpdatedAt.Format(time.RFC3339)
	if body.User.UpdatedAt != expectedUpdatedAt {
		t.Fatalf("expected refresh updated_at %q, got %q", expectedUpdatedAt, body.User.UpdatedAt)
	}
}

func TestAuthHandlerRefreshRejectsTokenInBody(t *testing.T) {
	handler := &AuthHandler{}
	app := fiber.New()
	app.Post("/refresh", handler.Refresh)

	req := httptest.NewRequest("POST", "/refresh", bytes.NewBufferString(`{"refresh_token":"refresh-token"}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: refreshTokenCookie, Value: "refresh-token"})
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestAuthHandlerRefreshRejectsInactiveStoredToken(t *testing.T) {
	roleID := uuid.New()
	userID := uuid.New()
	email := "manager@example.com"
	currentToken := "refresh-token"
	currentHash := authpkg.HashToken(currentToken)
	revokedAt := time.Now()

	userRepo := &fakeAuthUserRepo{
		userByID: &models.User{
			ID:       userID,
			Email:    &email,
			Name:     "Manager",
			RoleID:   roleID,
			IsActive: true,
			Role:     &models.Role{ID: roleID, Name: "manager"},
		},
	}
	tokenStore := &fakeRefreshTokenStore{
		stored: &models.RefreshToken{
			UserID:    userID,
			TokenHash: currentHash,
			ExpiresAt: time.Now().Add(time.Hour),
			RevokedAt: &revokedAt,
		},
	}
	jwtManager := &fakeJWTManager{validateRefreshClaims: &authpkg.Claims{UserID: userID}}
	handler := &AuthHandler{
		userRepo:         userRepo,
		refreshTokenRepo: tokenStore,
		jwtManager:       jwtManager,
		workflow:         newAuthWorkflow(userRepo, tokenStore, jwtManager),
	}

	app := fiber.New()
	app.Post("/refresh", handler.Refresh)

	req := httptest.NewRequest("POST", "/refresh", nil)
	req.AddCookie(&http.Cookie{Name: refreshTokenCookie, Value: currentToken})
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", resp.StatusCode)
	}
}

func TestAuthHandlerLogoutRevokesCookieTokenAndClearsCookie(t *testing.T) {
	token := "refresh-token"
	tokenStore := &fakeRefreshTokenStore{}
	handler := &AuthHandler{refreshTokenRepo: tokenStore}

	app := fiber.New()
	app.Post("/logout", handler.Logout)

	req := httptest.NewRequest("POST", "/logout", nil)
	req.AddCookie(&http.Cookie{Name: refreshTokenCookie, Value: token})
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	if tokenStore.revokedHash != authpkg.HashToken(token) {
		t.Fatalf("expected cookie token hash to be revoked")
	}
	if tokenStore.revokeReason != "user_logout" {
		t.Fatalf("expected logout revoke reason, got %q", tokenStore.revokeReason)
	}
	if cookie := findCookie(resp.Cookies(), refreshTokenCookie); cookie == nil || cookie.Value != "" || cookie.Expires.After(time.Now()) {
		t.Fatalf("expected refresh token cookie to be cleared")
	}
}

func TestAuthHandlerMeRequiresAuthentication(t *testing.T) {
	handler := &AuthHandler{}
	app := fiber.New()
	app.Get("/me", handler.Me)

	req := httptest.NewRequest("GET", "/me", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", resp.StatusCode)
	}
}

func findCookie(cookies []*http.Cookie, name string) *http.Cookie {
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie
		}
	}
	return nil
}

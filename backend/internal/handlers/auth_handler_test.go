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
	stored             *models.RefreshToken
	getErr             error
	createErr          error
	revokeErr          error
	rotateErr          error
	familyRevokeErr    error
	created            *models.RefreshToken
	revokedHash        string
	revokeReason       string
	rotatedOldHash     string
	rotatedReason      string
	rotatedNewToken    *models.RefreshToken
	familyRevokeCalls  int
	revokedFamilyID    uuid.UUID
	familyRevokeReason string
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

func (f *fakeRefreshTokenStore) RevokeFamily(_ context.Context, familyID uuid.UUID, reason string) error {
	f.familyRevokeCalls++
	f.revokedFamilyID = familyID
	f.familyRevokeReason = reason
	return f.familyRevokeErr
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
	} else if cookie.SameSite != http.SameSiteStrictMode {
		t.Fatalf("expected website cookie SameSite=Strict, got %v", cookie.SameSite)
	}
}

func TestAuthHandlerLoginUsesWailsCompatibleRefreshCookie(t *testing.T) {
	roleID := uuid.New()
	userID := uuid.New()
	passwordHash, err := authpkg.HashPassword("secret123")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	email := "owner@example.com"
	user := &models.User{
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
	userRepo := &fakeAuthUserRepo{userByEmail: user, userByID: user}
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

	req := httptest.NewRequest("POST", "/login", bytes.NewBufferString(`{"email":"owner@example.com","password":"secret123"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "http://wails.localhost")
	req.Header.Set("X-Forwarded-Proto", "https")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	cookie := findCookie(resp.Cookies(), refreshTokenCookie)
	if cookie == nil {
		t.Fatalf("expected refresh token cookie")
	}
	if cookie.SameSite != http.SameSiteNoneMode {
		t.Fatalf("expected Wails cookie SameSite=None, got %v", cookie.SameSite)
	}
	if !cookie.Secure {
		t.Fatalf("expected Wails refresh cookie to be Secure")
	}
}

func TestRefreshCookieSameSiteSupportsWailsCustomScheme(t *testing.T) {
	app := fiber.New()
	app.Get("/cookie", func(c *fiber.Ctx) error {
		setRefreshTokenCookie(c, "refresh-token", time.Unix(400, 0), false)
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/cookie", nil)
	req.Header.Set("Origin", "wails://wails")
	req.Header.Set("X-Forwarded-Proto", "https")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}

	cookie := findCookie(resp.Cookies(), refreshTokenCookie)
	if cookie == nil {
		t.Fatalf("expected refresh token cookie")
	}
	if cookie.SameSite != http.SameSiteNoneMode {
		t.Fatalf("expected custom-scheme Wails cookie SameSite=None, got %v", cookie.SameSite)
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

func newRememberMeLoginApp(t *testing.T) *fiber.App {
	t.Helper()
	roleID := uuid.New()
	userID := uuid.New()
	passwordHash, err := authpkg.HashPassword("secret123")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	email := "remember@example.com"
	user := &models.User{
		ID:           userID,
		Email:        &email,
		Name:         "Owner",
		PasswordHash: &passwordHash,
		RoleID:       roleID,
		IsActive:     true,
		Role:         &models.Role{ID: roleID, Name: "owner"},
	}
	userRepo := &fakeAuthUserRepo{userByEmail: user, userByID: user}
	tokenStore := &fakeRefreshTokenStore{}
	jwtManager := &fakeJWTManager{tokenPair: &authpkg.TokenPair{
		AccessToken:           "access-token",
		RefreshToken:          "refresh-token",
		AccessTokenExpiresAt:  time.Unix(300, 0),
		RefreshTokenExpiresAt: time.Unix(400, 0),
	}}
	handler := &AuthHandler{
		userRepo:         userRepo,
		refreshTokenRepo: tokenStore,
		jwtManager:       jwtManager,
		workflow:         newAuthWorkflow(userRepo, tokenStore, jwtManager),
	}
	app := fiber.New()
	app.Post("/login", handler.Login)
	return app
}

func loginWithRememberMe(t *testing.T, app *fiber.App, body string) *http.Cookie {
	t.Helper()
	req := httptest.NewRequest("POST", "/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	cookie := findCookie(resp.Cookies(), refreshTokenCookie)
	if cookie == nil {
		t.Fatalf("expected refresh cookie in response")
	}
	return cookie
}

func TestLoginRememberMeFalseIssuesSessionScopedCookie(t *testing.T) {
	app := newRememberMeLoginApp(t)
	cookie := loginWithRememberMe(t, app, `{"email":"remember@example.com","password":"secret123","remember_me":false}`)
	if !cookie.Expires.IsZero() {
		t.Fatalf("session-scoped cookie must not carry Expires, got %v", cookie.Expires)
	}
	if cookie.MaxAge != 0 {
		t.Fatalf("session-scoped cookie must not carry MaxAge, got %d", cookie.MaxAge)
	}
}

func TestLoginRememberMeTrueKeepsPersistentCookie(t *testing.T) {
	app := newRememberMeLoginApp(t)
	cookie := loginWithRememberMe(t, app, `{"email":"remember@example.com","password":"secret123","remember_me":true}`)
	if cookie.Expires.IsZero() || cookie.Expires.Unix() != 400 {
		t.Fatalf("persistent cookie should keep configured expiry, got %v", cookie.Expires)
	}
}

func TestLoginRememberMeAbsentKeepsPersistentDefault(t *testing.T) {
	app := newRememberMeLoginApp(t)
	cookie := loginWithRememberMe(t, app, `{"email":"remember@example.com","password":"secret123"}`)
	if cookie.Expires.IsZero() || cookie.Expires.Unix() != 400 {
		t.Fatalf("absent remember_me should keep legacy persistent cookie, got %v", cookie.Expires)
	}
}

func newRefreshRaceTestApp(t *testing.T, revokedReason string, revokedAgo time.Duration) *fiber.App {
	t.Helper()
	app, _ := newRefreshTestAppWithToken(t, refreshRaceStoredToken(revokedReason, revokedAgo))
	return app
}

// refreshRaceStoredToken builds a stored refresh token revoked revokedAgo ago
// with the given reason, belonging to a fresh token family.
func refreshRaceStoredToken(revokedReason string, revokedAgo time.Duration) *models.RefreshToken {
	userID := uuid.New()
	hash := authpkg.HashToken("stale-token")
	revokedAt := time.Now().Add(-revokedAgo)
	return &models.RefreshToken{
		UserID:        userID,
		TokenHash:     hash,
		ExpiresAt:     time.Now().Add(time.Hour),
		RevokedAt:     &revokedAt,
		RevokedReason: &revokedReason,
		FamilyID:      uuid.New(),
	}
}

// newRefreshTestAppWithToken wires a refresh handler around the given stored
// token and returns the app plus the fake store for call assertions.
func newRefreshTestAppWithToken(t *testing.T, stored *models.RefreshToken) (*fiber.App, *fakeRefreshTokenStore) {
	t.Helper()
	email := "race@example.com"
	user := &models.User{
		ID:       stored.UserID,
		Email:    &email,
		Name:     "Manager",
		RoleID:   uuid.New(),
		IsActive: true,
		Role:     &models.Role{ID: uuid.New(), Name: "manager"},
	}
	userRepo := &fakeAuthUserRepo{userByID: user}
	tokenStore := &fakeRefreshTokenStore{stored: stored}
	jwtManager := &fakeJWTManager{
		validateRefreshClaims: &authpkg.Claims{UserID: stored.UserID},
		tokenPair: &authpkg.TokenPair{
			AccessToken:           "grace-access",
			RefreshToken:          "grace-refresh",
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
	return app, tokenStore
}

func postRefresh(t *testing.T, app *fiber.App) *http.Response {
	t.Helper()
	req := httptest.NewRequest("POST", "/refresh", nil)
	req.AddCookie(&http.Cookie{Name: refreshTokenCookie, Value: "stale-token"})
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	return resp
}

func TestRefreshGraceWindowHonorsSiblingRotationRace(t *testing.T) {
	app := newRefreshRaceTestApp(t, "token_refresh", 5*time.Second)
	resp := postRefresh(t, app)
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected grace window to honor sibling rotation (200), got %d", resp.StatusCode)
	}
	if findCookie(resp.Cookies(), refreshTokenCookie) == nil {
		t.Fatalf("expected fresh refresh cookie so the lagging tab converges")
	}
}

func TestRefreshGraceWindowRejectsBeyondWindow(t *testing.T) {
	app := newRefreshRaceTestApp(t, "token_refresh", 45*time.Second)
	resp := postRefresh(t, app)
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected beyond-grace replay to be rejected (401), got %d", resp.StatusCode)
	}
}

func TestRefreshGraceWindowRejectsNonRotationRevocations(t *testing.T) {
	app := newRefreshRaceTestApp(t, "user_logout", 5*time.Second)
	resp := postRefresh(t, app)
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected non-rotation revocation to stay rejected (401), got %d", resp.StatusCode)
	}
}

func TestRefreshReuseBeyondGraceRevokesFamily(t *testing.T) {
	stored := refreshRaceStoredToken("token_refresh", 45*time.Second)
	app, store := newRefreshTestAppWithToken(t, stored)

	resp := postRefresh(t, app)

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected beyond-grace replay to be rejected (401), got %d", resp.StatusCode)
	}
	if store.familyRevokeCalls != 1 {
		t.Fatalf("expected one family revocation, got %d", store.familyRevokeCalls)
	}
	if store.revokedFamilyID != stored.FamilyID {
		t.Fatalf("expected family %s to be revoked, got %s", stored.FamilyID, store.revokedFamilyID)
	}
	if store.familyRevokeReason != tokenFamilyReuseRevocationReason {
		t.Fatalf("expected reason %s, got %s", tokenFamilyReuseRevocationReason, store.familyRevokeReason)
	}
}

func TestRefreshGraceWindowContinuationJoinsSameFamily(t *testing.T) {
	stored := refreshRaceStoredToken("token_refresh", 5*time.Second)
	app, store := newRefreshTestAppWithToken(t, stored)

	resp := postRefresh(t, app)

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected grace window continuation to succeed (200), got %d", resp.StatusCode)
	}
	if store.familyRevokeCalls != 0 {
		t.Fatalf("grace-window sibling race must not revoke the family, got %d revocations", store.familyRevokeCalls)
	}
	if store.created == nil || store.created.FamilyID != stored.FamilyID {
		t.Fatalf("continuation token must join the presented token's family")
	}
}

func TestRefreshReuseOfLoggedOutTokenRevokesFamily(t *testing.T) {
	stored := refreshRaceStoredToken("user_logout", 5*time.Second)
	app, store := newRefreshTestAppWithToken(t, stored)

	resp := postRefresh(t, app)

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected replay after logout to be rejected (401), got %d", resp.StatusCode)
	}
	if store.familyRevokeCalls != 1 || store.revokedFamilyID != stored.FamilyID {
		t.Fatalf("replay of a logout-revoked token must contain the family")
	}
}

func TestRefreshExpiredTokenDoesNotRevokeFamily(t *testing.T) {
	// An expired-but-never-revoked token is an idle session, not theft: the
	// request is rejected without family containment.
	stored := refreshRaceStoredToken("", 0)
	stored.RevokedAt = nil
	stored.RevokedReason = nil
	stored.ExpiresAt = time.Now().Add(-time.Hour)
	app, store := newRefreshTestAppWithToken(t, stored)

	resp := postRefresh(t, app)

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected expired token to be rejected (401), got %d", resp.StatusCode)
	}
	if store.familyRevokeCalls != 0 {
		t.Fatalf("expired token must not trigger family revocation, got %d", store.familyRevokeCalls)
	}
}

func TestLoginAntiEnumerationUniformResponses(t *testing.T) {
	roleID := uuid.New()
	userID := uuid.New()
	passwordHash, err := authpkg.HashPassword("secret123")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	email := "disabled@example.com"
	disabledUser := &models.User{
		ID:           userID,
		Email:        &email,
		Name:         "Disabled",
		PasswordHash: &passwordHash,
		RoleID:       roleID,
		IsActive:     false,
		Role:         &models.Role{ID: roleID, Name: "cashier"},
	}
	userRepo := &fakeAuthUserRepo{userByEmail: disabledUser, userByID: disabledUser}
	tokenStore := &fakeRefreshTokenStore{}
	jwtManager := &fakeJWTManager{tokenPair: &authpkg.TokenPair{
		AccessToken:           "access-token",
		RefreshToken:          "refresh-token",
		AccessTokenExpiresAt:  time.Unix(300, 0),
		RefreshTokenExpiresAt: time.Unix(400, 0),
	}}
	handler := &AuthHandler{
		userRepo:         userRepo,
		refreshTokenRepo: tokenStore,
		jwtManager:       jwtManager,
		workflow:         newAuthWorkflow(userRepo, tokenStore, jwtManager),
	}
	app := fiber.New()
	app.Post("/login", handler.Login)

	cases := []struct {
		name     string
		body     string
		wantCode string
		wantMsg  string
	}{
		{
			name:     "unknown email",
			body:     `{"email":"nobody@example.com","password":"whatever"}`,
			wantCode: "INVALID_CREDENTIALS",
			wantMsg:  "Invalid email or password",
		},
		{
			name:     "disabled account with wrong password",
			body:     `{"email":"disabled@example.com","password":"wrong"}`,
			wantCode: "INVALID_CREDENTIALS",
			wantMsg:  "Invalid email or password",
		},
		{
			name:     "disabled account with correct password",
			body:     `{"email":"disabled@example.com","password":"secret123"}`,
			wantCode: "ACCOUNT_DISABLED",
			wantMsg:  "Your account has been disabled",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/login", bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", "application/json")
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("app.Test returned error: %v", err)
			}
			if resp.StatusCode != fiber.StatusUnauthorized {
				t.Fatalf("%s: expected 401, got %d", tc.name, resp.StatusCode)
			}
			var body struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			}
			raw, rerr := io.ReadAll(resp.Body)
			if rerr != nil {
				t.Fatalf("read body: %v", rerr)
			}
			if err := json.Unmarshal(raw, &body); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if body.Code != tc.wantCode || body.Message != tc.wantMsg {
				t.Fatalf("%s: got %s/%q want %s/%q", tc.name, body.Code, body.Message, tc.wantCode, tc.wantMsg)
			}
		})
	}
}

func TestPINLoginAntiEnumerationOrder(t *testing.T) {
	roleID := uuid.New()
	userID := uuid.New()
	pinHash, err := authpkg.HashPIN("9911")
	if err != nil {
		t.Fatalf("HashPIN returned error: %v", err)
	}
	inactiveUser := &models.User{
		ID:       userID,
		Name:     "Disabled",
		PINHash:  &pinHash,
		RoleID:   roleID,
		IsActive: false,
		Role:     &models.Role{ID: roleID, Name: "cashier"},
	}
	userRepo := &fakeAuthUserRepo{userByID: inactiveUser}
	tokenStore := &fakeRefreshTokenStore{}
	jwtManager := &fakeJWTManager{tokenPair: &authpkg.TokenPair{
		AccessToken:           "access-token",
		RefreshToken:          "refresh-token",
		AccessTokenExpiresAt:  time.Unix(300, 0),
		RefreshTokenExpiresAt: time.Unix(400, 0),
	}}
	handler := &AuthHandler{
		userRepo:         userRepo,
		refreshTokenRepo: tokenStore,
		jwtManager:       jwtManager,
		workflow:         newAuthWorkflow(userRepo, tokenStore, jwtManager),
	}
	app := fiber.New()
	app.Post("/pin-login", handler.PINLogin)

	// wrong PIN on an inactive account must read as generic invalid
	// credentials, not disclose that the account is inactive
	req := httptest.NewRequest("POST", "/pin-login", bytes.NewBufferString(
		`{"user_id":"`+userID.String()+`","pin":"1111"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	var body struct {
		Code string `json:"code"`
	}
	raw, rerr := io.ReadAll(resp.Body)
	if rerr != nil {
		t.Fatalf("read body: %v", rerr)
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized || body.Code != "INVALID_CREDENTIALS" {
		t.Fatalf("expected generic INVALID_CREDENTIALS for wrong pin, got %d/%s", resp.StatusCode, body.Code)
	}
}

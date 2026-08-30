package middleware

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"dashpoint/backend/internal/auth"
	"dashpoint/backend/internal/models"
)

type stubAuthUserRepo struct {
	user *models.User
}

func (s *stubAuthUserRepo) GetByID(context.Context, uuid.UUID) (*models.User, error) {
	return s.user, nil
}

func requestWithToken(t *testing.T, app *fiber.App, jwtManager *auth.JWTManager, user *models.User, tokenVersion int) int {
	t.Helper()
	pair, err := jwtManager.GenerateTokenPair(user.ID, "u@example.com", user.Name, user.RoleID, user.Role.Name, tokenVersion)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	return resp.StatusCode
}

func TestAuthMiddlewareAcceptsCurrentTokenVersion(t *testing.T) {
	user := &models.User{ID: uuid.New(), Name: "Cashier", RoleID: uuid.New(), IsActive: true, TokenVersion: 3, Role: &models.Role{Name: "cashier"}}
	jwtManager := auth.NewJWTManager("test-secret", 15, 168)
	app := fiber.New()
	app.Use(AuthMiddleware(jwtManager, &stubAuthUserRepo{user: user}))
	app.Get("/", func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) })

	if status := requestWithToken(t, app, jwtManager, user, 3); status != fiber.StatusOK {
		t.Fatalf("expected current-version token to pass (200), got %d", status)
	}
}

func TestAuthMiddlewareRejectsStaleTokenVersion(t *testing.T) {
	user := &models.User{ID: uuid.New(), Name: "Cashier", RoleID: uuid.New(), IsActive: true, TokenVersion: 4, Role: &models.Role{Name: "cashier"}}
	jwtManager := auth.NewJWTManager("test-secret", 15, 168)
	app := fiber.New()
	app.Use(AuthMiddleware(jwtManager, &stubAuthUserRepo{user: user}))
	app.Get("/", func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) })

	// A token minted before the credential change (version 3) must die
	// instantly once the stored version is 4.
	if status := requestWithToken(t, app, jwtManager, user, 3); status != fiber.StatusUnauthorized {
		t.Fatalf("expected stale-version token to be rejected (401), got %d", status)
	}
}

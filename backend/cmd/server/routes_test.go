package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"dashpoint/backend/internal/handlers"
	"dashpoint/backend/internal/models"
)

type routeTestCategoryStore struct{}

func (routeTestCategoryStore) List(context.Context, string) ([]*models.Category, error) {
	return []*models.Category{{
		ID:        uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Name:      "Beverages",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}}, nil
}

func (routeTestCategoryStore) GetByID(context.Context, uuid.UUID) (*models.Category, error) {
	return &models.Category{
		ID:        uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Name:      "Beverages",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (routeTestCategoryStore) Create(context.Context, *models.Category) error { return nil }
func (routeTestCategoryStore) Update(context.Context, *models.Category) error { return nil }
func (routeTestCategoryStore) Delete(context.Context, uuid.UUID) error        { return nil }
func (routeTestCategoryStore) PermanentDelete(context.Context, uuid.UUID) error {
	return nil
}
func (routeTestCategoryStore) GetProductCount(context.Context, uuid.UUID) (int, error) {
	return 0, nil
}
func (routeTestCategoryStore) GetProductCounts(context.Context, []uuid.UUID) (map[uuid.UUID]int, error) {
	return map[uuid.UUID]int{}, nil
}
func (routeTestCategoryStore) DuplicateSiblingExists(context.Context, string, *uuid.UUID, *uuid.UUID) (bool, error) {
	return false, nil
}

func TestCatalogCategoryMetadataRoutesAllowProductViewers(t *testing.T) {
	app := fiber.New()
	protected := app.Group("", func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.MustParse("22222222-2222-2222-2222-222222222222"))
		return c.Next()
	})

	deps := &serverDependencies{
		permissionChecker: func(_ *fiber.Ctx, _ uuid.UUID, permission string) (bool, error) {
			return permission == "can_view_products", nil
		},
		categoryHandler: handlers.NewCategoryHandler(routeTestCategoryStore{}),
		productHandler:  &handlers.ProductHandler{},
	}

	registerCatalogRoutes(protected, deps)

	req := httptest.NewRequest(http.MethodGet, "/categories?status=active", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestAdjustmentPermissionNormalizesInput(t *testing.T) {
	permission, ok := adjustmentPermission(" Purchase ")
	if !ok {
		t.Fatal("expected permission to resolve")
	}
	if permission != "can_add_stock" {
		t.Fatalf("expected can_add_stock, got %q", permission)
	}
}

func TestPublicAuthRoutesRateLimitLogin(t *testing.T) {
	app := fiber.New()
	deps := &serverDependencies{
		authHandler: &handlers.AuthHandler{},
	}

	registerPublicRoutes(app.Group("/api/v1"), deps)

	for attempt := 0; attempt < 10; attempt++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{`))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("attempt %d returned error: %v", attempt+1, err)
		}
		if resp.StatusCode != fiber.StatusBadRequest {
			t.Fatalf("expected attempt %d to hit handler and return 400, got %d", attempt+1, resp.StatusCode)
		}
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("limited request returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusTooManyRequests {
		t.Fatalf("expected 429 after auth limit, got %d", resp.StatusCode)
	}
}

package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
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
func (routeTestCategoryStore) DuplicateSiblingExists(context.Context, string, *uuid.UUID) (bool, error) {
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
			return permission == "access_products_page", nil
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
	if permission != "manage_inventory_page" {
		t.Fatalf("expected manage_inventory_page, got %q", permission)
	}
}

func TestSubmitRoutesDenyWhenManagePermissionIsRevoked(t *testing.T) {
	userID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	var checkedPermissions []string

	app := fiber.New()
	protected := app.Group("", func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		c.Locals("role_name", "manager")
		return c.Next()
	})

	deps := &serverDependencies{
		permissionChecker: func(_ *fiber.Ctx, _ uuid.UUID, permission string) (bool, error) {
			checkedPermissions = append(checkedPermissions, permission)
			return false, nil
		},
	}

	registerUserRoutes(protected, deps)
	registerCatalogRoutes(protected, deps)
	registerOperationsRoutes(protected, deps)
	registerReportsRoutes(protected, deps)
	registerExpenseRoutes(protected, deps)
	registerUploadRoutes(protected, deps)

	type routeCase struct {
		name        string
		method      string
		path        string
		body        string
		permissions []string
	}

	id := "11111111-1111-1111-1111-111111111111"
	tests := []routeCase{
		{name: "create user", method: http.MethodPost, path: "/users", permissions: []string{"manage_users_page"}},
		{name: "update user", method: http.MethodPatch, path: "/users/" + id, permissions: []string{"manage_users_page"}},
		{name: "update user password", method: http.MethodPatch, path: "/users/" + id + "/password", permissions: []string{"manage_users_page"}},
		{name: "update user pin", method: http.MethodPatch, path: "/users/" + id + "/pin", permissions: []string{"manage_users_page"}},
		{name: "archive user", method: http.MethodDelete, path: "/users/" + id, permissions: []string{"manage_users_page"}},
		{name: "permanently delete user", method: http.MethodDelete, path: "/users/" + id + "/permanent", permissions: []string{"manage_users_page"}},
		{name: "create category", method: http.MethodPost, path: "/categories", permissions: []string{"manage_categories_page"}},
		{name: "update category", method: http.MethodPatch, path: "/categories/" + id, permissions: []string{"manage_categories_page"}},
		{name: "archive category", method: http.MethodDelete, path: "/categories/" + id, permissions: []string{"manage_categories_page"}},
		{name: "permanently delete category", method: http.MethodDelete, path: "/categories/" + id + "/permanent", permissions: []string{"manage_categories_page"}},
		{name: "create product", method: http.MethodPost, path: "/products", permissions: []string{"manage_products_page"}},
		{name: "update product", method: http.MethodPatch, path: "/products/" + id, permissions: []string{"manage_products_page"}},
		{name: "archive product", method: http.MethodDelete, path: "/products/" + id, permissions: []string{"manage_products_page"}},
		{name: "permanently delete product", method: http.MethodDelete, path: "/products/" + id + "/permanent", permissions: []string{"manage_products_page"}},
		{name: "update inventory threshold", method: http.MethodPatch, path: "/products/" + id + "/inventory", permissions: []string{"manage_inventory_page"}},
		{name: "adjust inventory", method: http.MethodPost, path: "/inventory/adjust", body: `{"adjustment_type":"purchase"}`, permissions: []string{"manage_inventory_page"}},
		{name: "start shift", method: http.MethodPost, path: "/shifts/start", permissions: []string{"manage_shifts_page"}},
		{name: "close shift", method: http.MethodPost, path: "/shifts/close", permissions: []string{"manage_shifts_page"}},
		{name: "cash pay in", method: http.MethodPost, path: "/shifts/pay-in", permissions: []string{"manage_shifts_page"}},
		{name: "cash pay out", method: http.MethodPost, path: "/shifts/pay-out", permissions: []string{"manage_shifts_page"}},
		{name: "create sale", method: http.MethodPost, path: "/sales", permissions: []string{"manage_pos_page"}},
		{name: "validate sale cart", method: http.MethodPost, path: "/sales/validate", permissions: []string{"manage_pos_page"}},
		{name: "void sale", method: http.MethodPost, path: "/sales/" + id + "/void", permissions: []string{"manage_sales_page"}},
		{name: "export sales report", method: http.MethodGet, path: "/reports/export/sales", permissions: []string{"manage_reports_page"}},
		{name: "export inventory report", method: http.MethodGet, path: "/reports/export/inventory", permissions: []string{"manage_reports_page"}},
		{name: "export top sellers report", method: http.MethodGet, path: "/reports/export/top-sellers", permissions: []string{"manage_reports_page"}},
		{name: "export comprehensive report", method: http.MethodGet, path: "/reports/export/comprehensive", permissions: []string{"manage_reports_page"}},
		{name: "create expense category", method: http.MethodPost, path: "/expenses/categories", permissions: []string{"manage_categories_page"}},
		{name: "update expense category", method: http.MethodPatch, path: "/expenses/categories/" + id, permissions: []string{"manage_categories_page"}},
		{name: "archive expense category", method: http.MethodDelete, path: "/expenses/categories/" + id, permissions: []string{"manage_categories_page"}},
		{name: "permanently delete expense category", method: http.MethodDelete, path: "/expenses/categories/" + id + "/permanent", permissions: []string{"manage_categories_page"}},
		{name: "create expense", method: http.MethodPost, path: "/expenses", permissions: []string{"manage_expenses_page"}},
		{name: "update expense", method: http.MethodPatch, path: "/expenses/" + id, permissions: []string{"manage_expenses_page"}},
		{name: "archive expense", method: http.MethodDelete, path: "/expenses/" + id, permissions: []string{"manage_expenses_page"}},
		{name: "upload image", method: http.MethodPost, path: "/upload/image", permissions: []string{"manage_products_page", "manage_expenses_page"}},
		{name: "delete image", method: http.MethodDelete, path: "/upload/image/product.png", permissions: []string{"manage_products_page", "manage_expenses_page"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkedPermissions = nil
			body := tt.body
			if body == "" {
				body = `{}`
			}
			req := httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("app.Test returned error: %v", err)
			}
			if resp.StatusCode != fiber.StatusForbidden {
				t.Fatalf("expected status 403, got %d", resp.StatusCode)
			}
			if !reflect.DeepEqual(checkedPermissions, tt.permissions) {
				t.Fatalf("expected permission checks %v, got %v", tt.permissions, checkedPermissions)
			}
		})
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

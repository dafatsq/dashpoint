package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"dashpoint/backend/internal/audit"
	"dashpoint/backend/internal/models"
	"dashpoint/backend/internal/repository"
)

type stubAuditRepository struct {
	calls int
	last  *models.AuditLogEntry
}

func (s *stubAuditRepository) Create(_ context.Context, entry *models.AuditLogEntry) error {
	s.calls++
	s.last = entry
	return nil
}

func TestParseCreateProductInputRejectsInvalidOptionalFields(t *testing.T) {
	invalidCategory := "not-a-uuid"
	invalidCost := "abc"

	_, err := parseCreateProductInput(CreateProductRequest{
		Name:       "Test Product",
		Price:      "10.50",
		Cost:       &invalidCost,
		CategoryID: &invalidCategory,
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "cost") {
		t.Fatalf("expected cost validation error, got %v", err)
	}
}

func TestParseCreateProductInputRejectsInvalidCategoryID(t *testing.T) {
	invalidCategory := "not-a-uuid"

	_, err := parseCreateProductInput(CreateProductRequest{
		Name:       "Test Product",
		Price:      "10.50",
		CategoryID: &invalidCategory,
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "category_id") {
		t.Fatalf("expected category_id validation error, got %v", err)
	}
}

func TestApplyUpdateProductRequestRejectsInvalidTaxRate(t *testing.T) {
	product := testProduct()
	invalidTaxRate := "bad-value"

	err := applyUpdateProductRequest(product, UpdateProductRequest{TaxRate: &invalidTaxRate})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "tax_rate") {
		t.Fatalf("expected tax_rate validation error, got %v", err)
	}
}

func TestParseStockAdjustmentRequestRejectsInvalidQuantity(t *testing.T) {
	app := fiber.New()
	app.Post("/", func(c *fiber.Ctx) error {
		_, err := parseStockAdjustmentRequest(c)
		return err
	})

	req := strings.NewReader(`{"product_id":"00000000-0000-0000-0000-000000000001","adjustment_type":"purchase","quantity":"oops"}`)
	httpReq := httptest.NewRequest("POST", "/", req)
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(httpReq)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestParseStockAdjustmentRequestAllowsDestructiveAdjustmentsWithoutReason(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{
			name: "damage without reason",
			body: `{"product_id":"00000000-0000-0000-0000-000000000001","adjustment_type":"damage","quantity":"2"}`,
		},
		{
			name: "loss without reason",
			body: `{"product_id":"00000000-0000-0000-0000-000000000001","adjustment_type":"loss","quantity":"1"}`,
		},
		{
			name: "negative correction without reason",
			body: `{"product_id":"00000000-0000-0000-0000-000000000001","adjustment_type":"adjustment","quantity":"-3"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Post("/", func(c *fiber.Ctx) error {
				_, err := parseStockAdjustmentRequest(c)
				return err
			})

			httpReq := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
			httpReq.Header.Set("Content-Type", "application/json")
			resp, err := app.Test(httpReq)
			if err != nil {
				t.Fatalf("app.Test returned error: %v", err)
			}
			if resp.StatusCode != fiber.StatusOK {
				t.Fatalf("expected 200, got %d", resp.StatusCode)
			}
		})
	}
}

func TestGetInventoryUsesRequestedPagination(t *testing.T) {
	productID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	var capturedLimit int
	var capturedOffset int
	var capturedAdjustmentType *models.AdjustmentType
	var capturedAdjustedBy *uuid.UUID

	handler := NewProductHandler(&fakeProductStore{}, &fakeInventoryStore{
		getByProductIDFunc: func(context.Context, uuid.UUID) (*models.InventoryItem, error) {
			return &models.InventoryItem{
				ProductID:         productID,
				Quantity:          decimal.RequireFromString("5"),
				LowStockThreshold: decimal.RequireFromString("2"),
				UpdatedAt:         time.Now(),
			}, nil
		},
		getAdjustmentHistoryFunc: func(_ context.Context, id uuid.UUID, limit int, offset int, adjustmentType *models.AdjustmentType, adjustedBy *uuid.UUID, startDate, endDate *time.Time) ([]*models.StockAdjustment, int, error) {
			if id != productID {
				t.Fatalf("expected product id %s, got %s", productID, id)
			}
			capturedLimit = limit
			capturedOffset = offset
			capturedAdjustmentType = adjustmentType
			capturedAdjustedBy = adjustedBy
			return nil, 0, nil
		},
	}, &fakeCategoryStore{}, "")

	app := fiber.New()
	app.Get("/products/:id/inventory", handler.GetInventory)

	req := httptest.NewRequest(http.MethodGet, "/products/00000000-0000-0000-0000-000000000001/inventory?limit=25&offset=50", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if capturedLimit != 25 || capturedOffset != 50 {
		t.Fatalf("expected limit=25 and offset=50, got limit=%d offset=%d", capturedLimit, capturedOffset)
	}
	if capturedAdjustmentType != nil {
		t.Fatalf("expected nil adjustment type filter, got %q", *capturedAdjustmentType)
	}
	if capturedAdjustedBy != nil {
		t.Fatalf("expected nil user filter, got %s", capturedAdjustedBy.String())
	}
}

func TestGetInventoryAcceptsAdjustmentTypeFilter(t *testing.T) {
	productID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	var capturedAdjustmentType *models.AdjustmentType

	handler := NewProductHandler(&fakeProductStore{}, &fakeInventoryStore{
		getByProductIDFunc: func(context.Context, uuid.UUID) (*models.InventoryItem, error) {
			return &models.InventoryItem{
				ProductID:         productID,
				Quantity:          decimal.RequireFromString("5"),
				LowStockThreshold: decimal.RequireFromString("2"),
				UpdatedAt:         time.Now(),
			}, nil
		},
		getAdjustmentHistoryFunc: func(_ context.Context, id uuid.UUID, limit int, offset int, adjustmentType *models.AdjustmentType, adjustedBy *uuid.UUID, startDate, endDate *time.Time) ([]*models.StockAdjustment, int, error) {
			if id != productID {
				t.Fatalf("expected product id %s, got %s", productID, id)
			}
			capturedAdjustmentType = adjustmentType
			if adjustedBy != nil {
				t.Fatalf("expected nil user filter, got %s", adjustedBy.String())
			}
			return nil, 0, nil
		},
	}, &fakeCategoryStore{}, "")

	app := fiber.New()
	app.Get("/products/:id/inventory", handler.GetInventory)

	req := httptest.NewRequest(http.MethodGet, "/products/00000000-0000-0000-0000-000000000001/inventory?adjustment_type=sale", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if capturedAdjustmentType == nil || *capturedAdjustmentType != models.AdjustmentSale {
		t.Fatalf("expected adjustment type %q, got %v", models.AdjustmentSale, capturedAdjustmentType)
	}
}

func TestGetInventoryAcceptsUserFilter(t *testing.T) {
	productID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000010")
	var capturedAdjustedBy *uuid.UUID

	handler := NewProductHandler(&fakeProductStore{}, &fakeInventoryStore{
		getByProductIDFunc: func(context.Context, uuid.UUID) (*models.InventoryItem, error) {
			return &models.InventoryItem{
				ProductID:         productID,
				Quantity:          decimal.RequireFromString("5"),
				LowStockThreshold: decimal.RequireFromString("2"),
				UpdatedAt:         time.Now(),
			}, nil
		},
		getAdjustmentHistoryFunc: func(_ context.Context, id uuid.UUID, limit int, offset int, adjustmentType *models.AdjustmentType, adjustedBy *uuid.UUID, startDate, endDate *time.Time) ([]*models.StockAdjustment, int, error) {
			if id != productID {
				t.Fatalf("expected product id %s, got %s", productID, id)
			}
			capturedAdjustedBy = adjustedBy
			if adjustmentType != nil {
				t.Fatalf("expected nil adjustment type filter, got %q", *adjustmentType)
			}
			return nil, 0, nil
		},
	}, &fakeCategoryStore{}, "")

	app := fiber.New()
	app.Get("/products/:id/inventory", handler.GetInventory)

	req := httptest.NewRequest(http.MethodGet, "/products/00000000-0000-0000-0000-000000000001/inventory?user_id=00000000-0000-0000-0000-000000000010", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if capturedAdjustedBy == nil || *capturedAdjustedBy != userID {
		t.Fatalf("expected user filter %s, got %v", userID, capturedAdjustedBy)
	}
}

func TestGetInventoryAcceptsDateFilters(t *testing.T) {
	productID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	var capturedStart *time.Time
	var capturedEnd *time.Time

	handler := NewProductHandler(&fakeProductStore{}, &fakeInventoryStore{
		getByProductIDFunc: func(context.Context, uuid.UUID) (*models.InventoryItem, error) {
			return &models.InventoryItem{
				ProductID:         productID,
				Quantity:          decimal.RequireFromString("5"),
				LowStockThreshold: decimal.RequireFromString("2"),
				UpdatedAt:         time.Now(),
			}, nil
		},
		getAdjustmentHistoryFunc: func(_ context.Context, id uuid.UUID, limit int, offset int, adjustmentType *models.AdjustmentType, adjustedBy *uuid.UUID, startDate, endDate *time.Time) ([]*models.StockAdjustment, int, error) {
			if id != productID {
				t.Fatalf("expected product id %s, got %s", productID, id)
			}
			capturedStart = startDate
			capturedEnd = endDate
			return nil, 0, nil
		},
	}, &fakeCategoryStore{}, "")

	app := fiber.New()
	app.Get("/products/:id/inventory", handler.GetInventory)

	req := httptest.NewRequest(http.MethodGet, "/products/00000000-0000-0000-0000-000000000001/inventory?from=2026-05-01&to=2026-05-14", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if capturedStart == nil || capturedStart.Format("2006-01-02") != "2026-05-01" {
		t.Fatalf("expected start date 2026-05-01, got %v", capturedStart)
	}
	if capturedEnd == nil || capturedEnd.Format("2006-01-02") != "2026-05-15" {
		t.Fatalf("expected end date 2026-05-15, got %v", capturedEnd)
	}
}

func TestGetInventoryRejectsInvalidAdjustmentTypeFilter(t *testing.T) {
	handler := NewProductHandler(&fakeProductStore{}, &fakeInventoryStore{}, &fakeCategoryStore{}, "")

	app := fiber.New()
	app.Get("/products/:id/inventory", handler.GetInventory)

	req := httptest.NewRequest(http.MethodGet, "/products/00000000-0000-0000-0000-000000000001/inventory?adjustment_type=bad", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestGetInventoryRejectsInvalidUserFilter(t *testing.T) {
	handler := NewProductHandler(&fakeProductStore{}, &fakeInventoryStore{}, &fakeCategoryStore{}, "")

	app := fiber.New()
	app.Get("/products/:id/inventory", handler.GetInventory)

	req := httptest.NewRequest(http.MethodGet, "/products/00000000-0000-0000-0000-000000000001/inventory?user_id=bad", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestUpdateInventoryThreshold(t *testing.T) {
	productID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	updateCalls := 0
	auditRepo := &stubAuditRepository{}
	audit.Init(auditRepo)

	handler := NewProductHandler(&fakeProductStore{
		getByIDFunc: func(_ context.Context, id uuid.UUID) (*models.Product, error) {
			if id != productID {
				t.Fatalf("expected product id %s, got %s", productID, id)
			}
			product := testProduct()
			product.ID = id
			return product, nil
		},
	}, &fakeInventoryStore{
		getByProductIDFunc: func(_ context.Context, id uuid.UUID) (*models.InventoryItem, error) {
			if id != productID {
				t.Fatalf("expected product id %s, got %s", productID, id)
			}
			threshold := "2"
			if updateCalls > 0 {
				threshold = "7"
			}
			return &models.InventoryItem{
				ProductID:         productID,
				Quantity:          decimal.RequireFromString("5"),
				LowStockThreshold: decimal.RequireFromString(threshold),
				UpdatedAt:         time.Now(),
			}, nil
		},
		updateThresholdsFunc: func(_ context.Context, id uuid.UUID, threshold decimal.Decimal) error {
			if id != productID {
				t.Fatalf("expected product id %s, got %s", productID, id)
			}
			if !threshold.Equal(decimal.RequireFromString("7")) {
				t.Fatalf("expected threshold 7, got %s", threshold.String())
			}
			updateCalls++
			return nil
		},
	}, &fakeCategoryStore{}, "")

	app := fiber.New()
	app.Patch("/products/:id/inventory", handler.UpdateInventoryThreshold)

	req := httptest.NewRequest(http.MethodPatch, "/products/00000000-0000-0000-0000-000000000001/inventory", strings.NewReader(`{"low_stock_threshold":"7"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if updateCalls != 1 {
		t.Fatalf("expected one update call, got %d", updateCalls)
	}
	if auditRepo.calls != 1 {
		t.Fatalf("expected one audit call, got %d", auditRepo.calls)
	}
	if auditRepo.last == nil {
		t.Fatal("expected captured audit entry")
	}
	if auditRepo.last.Action != models.AuditActionThresholdUpdate {
		t.Fatalf("expected audit action %q, got %q", models.AuditActionThresholdUpdate, auditRepo.last.Action)
	}
	if auditRepo.last.EntityType != models.AuditEntityInventory {
		t.Fatalf("expected inventory audit entity, got %q", auditRepo.last.EntityType)
	}

	var payload struct {
		Message   string `json:"message"`
		Inventory struct {
			LowStockThreshold string `json:"low_stock_threshold"`
		} `json:"inventory"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if payload.Inventory.LowStockThreshold != "7" {
		t.Fatalf("expected updated threshold 7, got %q", payload.Inventory.LowStockThreshold)
	}
}

func TestUpdateInventoryThresholdRejectsInvalidThreshold(t *testing.T) {
	handler := NewProductHandler(&fakeProductStore{}, &fakeInventoryStore{}, &fakeCategoryStore{}, "")

	app := fiber.New()
	app.Patch("/products/:id/inventory", handler.UpdateInventoryThreshold)

	req := httptest.NewRequest(http.MethodPatch, "/products/00000000-0000-0000-0000-000000000001/inventory", strings.NewReader(`{"low_stock_threshold":"bad"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestUpdateInventoryThresholdRequiresThreshold(t *testing.T) {
	handler := NewProductHandler(&fakeProductStore{}, &fakeInventoryStore{}, &fakeCategoryStore{}, "")

	app := fiber.New()
	app.Patch("/products/:id/inventory", handler.UpdateInventoryThreshold)

	req := httptest.NewRequest(http.MethodPatch, "/products/00000000-0000-0000-0000-000000000001/inventory", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestUpdateInventoryThresholdRejectsNegativeThreshold(t *testing.T) {
	handler := NewProductHandler(&fakeProductStore{}, &fakeInventoryStore{}, &fakeCategoryStore{}, "")

	app := fiber.New()
	app.Patch("/products/:id/inventory", handler.UpdateInventoryThreshold)

	req := httptest.NewRequest(http.MethodPatch, "/products/00000000-0000-0000-0000-000000000001/inventory", strings.NewReader(`{"low_stock_threshold":"-1"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestUpdateInventoryThresholdRejectsArchivedProduct(t *testing.T) {
	productID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	updateCalled := false
	handler := NewProductHandler(&fakeProductStore{
		getByIDFunc: func(_ context.Context, id uuid.UUID) (*models.Product, error) {
			product := testProduct()
			product.ID = id
			product.IsActive = false
			return product, nil
		},
	}, &fakeInventoryStore{
		updateThresholdsFunc: func(context.Context, uuid.UUID, decimal.Decimal) error {
			updateCalled = true
			return nil
		},
	}, &fakeCategoryStore{}, "")

	app := fiber.New()
	app.Patch("/products/:id/inventory", handler.UpdateInventoryThreshold)

	req := httptest.NewRequest(http.MethodPatch, "/products/"+productID.String()+"/inventory", strings.NewReader(`{"low_stock_threshold":"7"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusConflict {
		t.Fatalf("expected 409, got %d", resp.StatusCode)
	}
	if updateCalled {
		t.Fatalf("expected threshold update to be blocked for archived product")
	}
}

func TestAdjustStockRejectsArchivedProduct(t *testing.T) {
	adjustCalled := false
	handler := NewProductHandler(&fakeProductStore{
		getByIDFunc: func(_ context.Context, id uuid.UUID) (*models.Product, error) {
			product := testProduct()
			product.ID = id
			product.IsActive = false
			return product, nil
		},
	}, &fakeInventoryStore{
		adjustStockFunc: func(context.Context, uuid.UUID, models.AdjustmentType, decimal.Decimal, *string, *string, *uuid.UUID, uuid.UUID) (*models.StockAdjustment, error) {
			adjustCalled = true
			return nil, nil
		},
	}, &fakeCategoryStore{}, "")

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New())
		return c.Next()
	})
	app.Post("/inventory/adjust", handler.AdjustStock)

	req := httptest.NewRequest(http.MethodPost, "/inventory/adjust", strings.NewReader(`{"product_id":"00000000-0000-0000-0000-000000000001","adjustment_type":"purchase","quantity":"2"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusConflict {
		t.Fatalf("expected 409, got %d", resp.StatusCode)
	}
	if adjustCalled {
		t.Fatalf("expected stock adjustment to be blocked for archived product")
	}
}

func TestAdjustStockAllowsRemovalWithoutReason(t *testing.T) {
	productID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	var capturedType models.AdjustmentType
	var capturedQuantity decimal.Decimal
	var capturedReason *string
	handler := NewProductHandler(&fakeProductStore{
		getByIDFunc: func(_ context.Context, id uuid.UUID) (*models.Product, error) {
			product := testProduct()
			product.ID = id
			return product, nil
		},
	}, &fakeInventoryStore{
		getByProductIDFunc: func(_ context.Context, id uuid.UUID) (*models.InventoryItem, error) {
			return &models.InventoryItem{
				ProductID:         id,
				Quantity:          decimal.RequireFromString("100"),
				LowStockThreshold: decimal.Zero,
				UpdatedAt:         time.Now(),
			}, nil
		},
		adjustStockFunc: func(_ context.Context, _ uuid.UUID, adjustmentType models.AdjustmentType, quantity decimal.Decimal, reason *string, _ *string, _ *uuid.UUID, _ uuid.UUID) (*models.StockAdjustment, error) {
			capturedType = adjustmentType
			capturedQuantity = quantity
			capturedReason = reason
			return &models.StockAdjustment{ID: uuid.New()}, nil
		},
	}, &fakeCategoryStore{}, "")

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New())
		return c.Next()
	})
	app.Post("/inventory/adjust", handler.AdjustStock)

	req := httptest.NewRequest(http.MethodPost, "/inventory/adjust", strings.NewReader(`{"product_id":"`+productID.String()+`","adjustment_type":"adjustment","quantity":"-10"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if capturedType != models.AdjustmentAdjustment {
		t.Fatalf("expected adjustment type, got %q", capturedType)
	}
	if !capturedQuantity.Equal(decimal.RequireFromString("-10")) {
		t.Fatalf("expected quantity -10, got %s", capturedQuantity)
	}
	if capturedReason != nil {
		t.Fatalf("expected nil reason, got %v", capturedReason)
	}
}

func TestAdjustStockConvertsDamageRemovalToNegativeDelta(t *testing.T) {
	productID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	var capturedQuantity decimal.Decimal
	handler := NewProductHandler(&fakeProductStore{
		getByIDFunc: func(_ context.Context, id uuid.UUID) (*models.Product, error) {
			product := testProduct()
			product.ID = id
			return product, nil
		},
	}, &fakeInventoryStore{
		getByProductIDFunc: func(_ context.Context, id uuid.UUID) (*models.InventoryItem, error) {
			return &models.InventoryItem{
				ProductID:         id,
				Quantity:          decimal.RequireFromString("100"),
				LowStockThreshold: decimal.Zero,
				UpdatedAt:         time.Now(),
			}, nil
		},
		adjustStockFunc: func(_ context.Context, _ uuid.UUID, adjustmentType models.AdjustmentType, quantity decimal.Decimal, _ *string, _ *string, _ *uuid.UUID, _ uuid.UUID) (*models.StockAdjustment, error) {
			if adjustmentType != models.AdjustmentDamage {
				t.Fatalf("expected damage adjustment, got %q", adjustmentType)
			}
			capturedQuantity = quantity
			return &models.StockAdjustment{ID: uuid.New()}, nil
		},
	}, &fakeCategoryStore{}, "")

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New())
		return c.Next()
	})
	app.Post("/inventory/adjust", handler.AdjustStock)

	req := httptest.NewRequest(http.MethodPost, "/inventory/adjust", strings.NewReader(`{"product_id":"`+productID.String()+`","adjustment_type":"damage","quantity":"10"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if !capturedQuantity.Equal(decimal.RequireFromString("-10")) {
		t.Fatalf("expected quantity -10, got %s", capturedQuantity)
	}
}

func TestAdjustStockRejectsStaleInventoryVersion(t *testing.T) {
	productID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	inventoryUpdatedAt := time.Now()
	adjustCalled := false
	handler := NewProductHandler(&fakeProductStore{
		getByIDFunc: func(_ context.Context, id uuid.UUID) (*models.Product, error) {
			product := testProduct()
			product.ID = id
			return product, nil
		},
	}, &fakeInventoryStore{
		getByProductIDFunc: func(_ context.Context, id uuid.UUID) (*models.InventoryItem, error) {
			return &models.InventoryItem{
				ProductID:         id,
				Quantity:          decimal.RequireFromString("100"),
				LowStockThreshold: decimal.Zero,
				UpdatedAt:         inventoryUpdatedAt,
			}, nil
		},
		adjustStockFunc: func(context.Context, uuid.UUID, models.AdjustmentType, decimal.Decimal, *string, *string, *uuid.UUID, uuid.UUID) (*models.StockAdjustment, error) {
			adjustCalled = true
			return nil, nil
		},
	}, &fakeCategoryStore{}, "")

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New())
		return c.Next()
	})
	app.Post("/inventory/adjust", handler.AdjustStock)

	staleVersion := inventoryUpdatedAt.Add(-time.Minute).Format(time.RFC3339Nano)
	body := `{"product_id":"` + productID.String() + `","adjustment_type":"adjustment","quantity":"1","expected_updated_at":"` + staleVersion + `"}`
	req := httptest.NewRequest(http.MethodPost, "/inventory/adjust", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusConflict {
		t.Fatalf("expected 409, got %d", resp.StatusCode)
	}
	if adjustCalled {
		t.Fatalf("expected stale inventory submit to be blocked")
	}
}

func testProduct() *models.Product {
	return &models.Product{
		ID:       uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		Name:     "Test Product",
		Price:    decimal.RequireFromString("10.00"),
		Cost:     decimal.RequireFromString("5.00"),
		TaxRate:  decimal.RequireFromString("11.00"),
		IsActive: true,
	}
}

func TestCreateSucceedsWhenThresholdUpdateFails(t *testing.T) {
	productRepo := &fakeProductStore{
		createFunc: func(_ context.Context, product *models.Product, _ *decimal.Decimal) error {
			if product.ID == uuid.Nil {
				product.ID = uuid.MustParse("00000000-0000-0000-0000-000000000010")
			}
			now := time.Now()
			product.CreatedAt = now
			product.UpdatedAt = now
			return nil
		},
		getByIDFunc: func(_ context.Context, id uuid.UUID) (*models.Product, error) {
			product := testProduct()
			product.ID = id
			return product, nil
		},
	}
	inventoryRepo := &fakeInventoryStore{
		updateThresholdsFunc: func(context.Context, uuid.UUID, decimal.Decimal) error {
			return context.DeadlineExceeded
		},
	}
	handler := NewProductHandler(productRepo, inventoryRepo, &fakeCategoryStore{}, "")

	app := fiber.New()
	app.Post("/products", handler.Create)

	body := `{"name":"Milk","price":"12.50","low_stock_threshold":"3"}`
	req := httptest.NewRequest(http.MethodPost, "/products", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}
}

func TestValidateActiveProductCategoryRejectsArchivedCategory(t *testing.T) {
	categoryID := uuid.New()
	handler := NewProductHandler(&fakeProductStore{}, &fakeInventoryStore{}, &fakeCategoryStore{
		getByIDFunc: func(context.Context, uuid.UUID) (*models.Category, error) {
			return &models.Category{ID: categoryID, Name: "Archived", IsActive: false}, nil
		},
	}, "")

	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		handler.validateActiveProductCategory(c, &categoryID)
		return nil
	})

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/", nil))
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusConflict {
		t.Fatalf("expected 409, got %d", resp.StatusCode)
	}
}

func TestCreateRejectsMissingProductCategory(t *testing.T) {
	categoryID := uuid.New()
	var createdCategoryID *uuid.UUID
	productRepo := &fakeProductStore{
		createFunc: func(_ context.Context, product *models.Product, _ *decimal.Decimal) error {
			createdCategoryID = product.CategoryID
			return nil
		},
	}
	categoryLookupCalled := false
	handler := NewProductHandler(productRepo, &fakeInventoryStore{}, &fakeCategoryStore{
		getByIDFunc: func(context.Context, uuid.UUID) (*models.Category, error) {
			categoryLookupCalled = true
			return nil, nil
		},
	}, "")

	app := fiber.New()
	app.Post("/products", handler.Create)

	body := `{"name":"Milk","price":"12.50","category_id":"` + categoryID.String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/products", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d; category lookup called=%v, created category id=%v", resp.StatusCode, categoryLookupCalled, createdCategoryID)
	}
}

func TestCreateRejectsArchivedProductCategory(t *testing.T) {
	categoryID := uuid.New()
	var createCalled bool
	productRepo := &fakeProductStore{
		createFunc: func(context.Context, *models.Product, *decimal.Decimal) error {
			createCalled = true
			return nil
		},
	}
	handler := NewProductHandler(productRepo, &fakeInventoryStore{}, &fakeCategoryStore{
		getByIDFunc: func(context.Context, uuid.UUID) (*models.Category, error) {
			return &models.Category{ID: categoryID, Name: "Archived", IsActive: false}, nil
		},
	}, "")

	app := fiber.New()
	app.Post("/products", handler.Create)

	body := `{"name":"Milk","price":"12.50","category_id":"` + categoryID.String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/products", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusConflict {
		t.Fatalf("expected 409, got %d; create called=%v", resp.StatusCode, createCalled)
	}
}

func TestUpdateAllowsClearingSKU(t *testing.T) {
	current := testProduct()
	sku := "SKU-123"
	current.SKU = &sku
	now := time.Now()
	current.CreatedAt = now
	current.UpdatedAt = now

	productRepo := &fakeProductStore{
		getByIDFunc: func(_ context.Context, id uuid.UUID) (*models.Product, error) {
			product := *current
			product.ID = id
			return &product, nil
		},
		updateFunc: func(_ context.Context, product *models.Product) error {
			copy := *product
			current = &copy
			return nil
		},
	}
	handler := NewProductHandler(productRepo, &fakeInventoryStore{}, &fakeCategoryStore{}, "")

	app := fiber.New()
	app.Patch("/products/:id", handler.Update)

	req := httptest.NewRequest(http.MethodPatch, "/products/00000000-0000-0000-0000-000000000001", strings.NewReader(`{"sku":""}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var payload struct {
		Product struct {
			SKU *string `json:"sku"`
		} `json:"product"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if payload.Product.SKU == nil {
		t.Fatal("expected sku field to be present")
	}
	if *payload.Product.SKU != "" {
		t.Fatalf("expected sku to be cleared to empty string, got %q", *payload.Product.SKU)
	}
}

func TestUpdateRejectsMissingProductCategory(t *testing.T) {
	categoryID := uuid.New()
	current := testProduct()
	current.CategoryID = nil
	var updatedCategoryID *uuid.UUID

	productRepo := &fakeProductStore{
		getByIDFunc: func(_ context.Context, id uuid.UUID) (*models.Product, error) {
			product := *current
			product.ID = id
			return &product, nil
		},
		updateFunc: func(_ context.Context, product *models.Product) error {
			updatedCategoryID = product.CategoryID
			return nil
		},
	}
	categoryLookupCalled := false
	handler := NewProductHandler(productRepo, &fakeInventoryStore{}, &fakeCategoryStore{
		getByIDFunc: func(context.Context, uuid.UUID) (*models.Category, error) {
			categoryLookupCalled = true
			return nil, nil
		},
	}, "")

	app := fiber.New()
	app.Patch("/products/:id", handler.Update)

	body := `{"category_id":"` + categoryID.String() + `"}`
	req := httptest.NewRequest(http.MethodPatch, "/products/00000000-0000-0000-0000-000000000001", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d; category lookup called=%v, updated category id=%v", resp.StatusCode, categoryLookupCalled, updatedCategoryID)
	}
}

func TestUpdateRejectsArchivedProductCategory(t *testing.T) {
	categoryID := uuid.New()
	current := testProduct()
	current.CategoryID = nil
	var updateCalled bool

	productRepo := &fakeProductStore{
		getByIDFunc: func(_ context.Context, id uuid.UUID) (*models.Product, error) {
			product := *current
			product.ID = id
			return &product, nil
		},
		updateFunc: func(context.Context, *models.Product) error {
			updateCalled = true
			return nil
		},
	}
	handler := NewProductHandler(productRepo, &fakeInventoryStore{}, &fakeCategoryStore{
		getByIDFunc: func(context.Context, uuid.UUID) (*models.Category, error) {
			return &models.Category{ID: categoryID, Name: "Archived", IsActive: false}, nil
		},
	}, "")

	app := fiber.New()
	app.Patch("/products/:id", handler.Update)

	body := `{"category_id":"` + categoryID.String() + `"}`
	req := httptest.NewRequest(http.MethodPatch, "/products/00000000-0000-0000-0000-000000000001", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusConflict {
		t.Fatalf("expected 409, got %d; update called=%v", resp.StatusCode, updateCalled)
	}
}

func TestUpdateAllowsUnchangedArchivedProductCategory(t *testing.T) {
	categoryID := uuid.New()
	current := testProduct()
	current.CategoryID = &categoryID
	updateCalled := false

	productRepo := &fakeProductStore{
		getByIDFunc: func(_ context.Context, id uuid.UUID) (*models.Product, error) {
			product := *current
			product.ID = id
			return &product, nil
		},
		updateFunc: func(_ context.Context, product *models.Product) error {
			updateCalled = true
			if product.CategoryID == nil || *product.CategoryID != categoryID {
				t.Fatalf("expected category id to stay %s, got %v", categoryID, product.CategoryID)
			}
			if product.Name != "Renamed" {
				t.Fatalf("expected name update, got %q", product.Name)
			}
			return nil
		},
	}
	handler := NewProductHandler(productRepo, &fakeInventoryStore{}, &fakeCategoryStore{
		getByIDFunc: func(context.Context, uuid.UUID) (*models.Category, error) {
			t.Fatalf("expected unchanged archived category not to be revalidated")
			return nil, nil
		},
	}, "")

	app := fiber.New()
	app.Patch("/products/:id", handler.Update)

	body := `{"name":"Renamed","category_id":"` + categoryID.String() + `"}`
	req := httptest.NewRequest(http.MethodPatch, "/products/00000000-0000-0000-0000-000000000001", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if !updateCalled {
		t.Fatalf("expected product update to be called")
	}
}

func TestUpdateRejectsArchivedProductWithoutRestore(t *testing.T) {
	current := testProduct()
	current.IsActive = false

	productRepo := &fakeProductStore{
		getByIDFunc: func(_ context.Context, id uuid.UUID) (*models.Product, error) {
			product := *current
			product.ID = id
			return &product, nil
		},
		updateFunc: func(_ context.Context, product *models.Product) error {
			t.Fatalf("expected update not to be called for archived product")
			return nil
		},
	}
	handler := NewProductHandler(productRepo, &fakeInventoryStore{}, &fakeCategoryStore{}, "")

	app := fiber.New()
	app.Patch("/products/:id", handler.Update)

	req := httptest.NewRequest(http.MethodPatch, "/products/00000000-0000-0000-0000-000000000001", strings.NewReader(`{"name":"Renamed"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusConflict {
		t.Fatalf("expected 409, got %d", resp.StatusCode)
	}
}

type fakeProductStore struct {
	listFunc                    func(context.Context, repository.ProductFilter) ([]*models.Product, int, error)
	getByIDFunc                 func(context.Context, uuid.UUID) (*models.Product, error)
	lookupFunc                  func(context.Context, string) (*models.Product, error)
	createFunc                  func(context.Context, *models.Product, *decimal.Decimal) error
	updateFunc                  func(context.Context, *models.Product) error
	deleteFunc                  func(context.Context, uuid.UUID) error
	hasSalesHistoryFunc         func(context.Context, uuid.UUID) (bool, error)
	permanentDeleteFunc         func(context.Context, uuid.UUID) error
	getBySKUIncludingInactiveFn func(context.Context, string) (*models.Product, error)
	getByBarcodeInactiveFn      func(context.Context, string) (*models.Product, error)
	skuExistsFunc               func(context.Context, string, *uuid.UUID) (bool, error)
	barcodeExistsFunc           func(context.Context, string, *uuid.UUID) (bool, error)
}

func (f *fakeProductStore) List(ctx context.Context, filter repository.ProductFilter) ([]*models.Product, int, error) {
	if f.listFunc != nil {
		return f.listFunc(ctx, filter)
	}
	return nil, 0, nil
}
func (f *fakeProductStore) GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	if f.getByIDFunc != nil {
		return f.getByIDFunc(ctx, id)
	}
	return nil, nil
}
func (f *fakeProductStore) Lookup(ctx context.Context, code string) (*models.Product, error) {
	if f.lookupFunc != nil {
		return f.lookupFunc(ctx, code)
	}
	return nil, nil
}
func (f *fakeProductStore) Create(ctx context.Context, product *models.Product, qty *decimal.Decimal) error {
	if f.createFunc != nil {
		return f.createFunc(ctx, product, qty)
	}
	return nil
}
func (f *fakeProductStore) Update(ctx context.Context, product *models.Product) error {
	if f.updateFunc != nil {
		return f.updateFunc(ctx, product)
	}
	return nil
}
func (f *fakeProductStore) Delete(ctx context.Context, id uuid.UUID) error {
	if f.deleteFunc != nil {
		return f.deleteFunc(ctx, id)
	}
	return nil
}
func (f *fakeProductStore) HasSalesHistory(ctx context.Context, id uuid.UUID) (bool, error) {
	if f.hasSalesHistoryFunc != nil {
		return f.hasSalesHistoryFunc(ctx, id)
	}
	return false, nil
}
func (f *fakeProductStore) PermanentDelete(ctx context.Context, id uuid.UUID) error {
	if f.permanentDeleteFunc != nil {
		return f.permanentDeleteFunc(ctx, id)
	}
	return nil
}
func (f *fakeProductStore) GetBySKUIncludingInactive(ctx context.Context, sku string) (*models.Product, error) {
	if f.getBySKUIncludingInactiveFn != nil {
		return f.getBySKUIncludingInactiveFn(ctx, sku)
	}
	return nil, nil
}
func (f *fakeProductStore) GetByBarcodeIncludingInactive(ctx context.Context, barcode string) (*models.Product, error) {
	if f.getByBarcodeInactiveFn != nil {
		return f.getByBarcodeInactiveFn(ctx, barcode)
	}
	return nil, nil
}
func (f *fakeProductStore) SKUExists(ctx context.Context, sku string, excludeID *uuid.UUID) (bool, error) {
	if f.skuExistsFunc != nil {
		return f.skuExistsFunc(ctx, sku, excludeID)
	}
	return false, nil
}
func (f *fakeProductStore) BarcodeExists(ctx context.Context, barcode string, excludeID *uuid.UUID) (bool, error) {
	if f.barcodeExistsFunc != nil {
		return f.barcodeExistsFunc(ctx, barcode, excludeID)
	}
	return false, nil
}

func (f *fakeProductStore) NameExists(ctx context.Context, name string, excludeID *uuid.UUID) (bool, error) {
	return false, nil
}

type fakeInventoryStore struct {
	getByProductIDFunc       func(context.Context, uuid.UUID) (*models.InventoryItem, error)
	adjustStockFunc          func(context.Context, uuid.UUID, models.AdjustmentType, decimal.Decimal, *string, *string, *uuid.UUID, uuid.UUID) (*models.StockAdjustment, error)
	setQuantityFunc          func(context.Context, uuid.UUID, decimal.Decimal, *string, uuid.UUID) (*models.StockAdjustment, error)
	updateThresholdsFunc     func(context.Context, uuid.UUID, decimal.Decimal) error
	getLowStockProductsFunc  func(context.Context) ([]*models.ProductWithInventory, error)
	getAdjustmentHistoryFunc func(context.Context, uuid.UUID, int, int, *models.AdjustmentType, *uuid.UUID, *time.Time, *time.Time) ([]*models.StockAdjustment, int, error)
}

func (f *fakeInventoryStore) GetByProductID(ctx context.Context, id uuid.UUID) (*models.InventoryItem, error) {
	if f.getByProductIDFunc != nil {
		return f.getByProductIDFunc(ctx, id)
	}
	return nil, nil
}
func (f *fakeInventoryStore) AdjustStock(ctx context.Context, productID uuid.UUID, adjustmentType models.AdjustmentType, quantity decimal.Decimal, reason *string, referenceType *string, referenceID *uuid.UUID, adjustedBy uuid.UUID) (*models.StockAdjustment, error) {
	if f.adjustStockFunc != nil {
		return f.adjustStockFunc(ctx, productID, adjustmentType, quantity, reason, referenceType, referenceID, adjustedBy)
	}
	return nil, nil
}
func (f *fakeInventoryStore) SetQuantity(ctx context.Context, productID uuid.UUID, quantity decimal.Decimal, reason *string, adjustedBy uuid.UUID) (*models.StockAdjustment, error) {
	if f.setQuantityFunc != nil {
		return f.setQuantityFunc(ctx, productID, quantity, reason, adjustedBy)
	}
	return nil, nil
}
func (f *fakeInventoryStore) UpdateThresholds(ctx context.Context, productID uuid.UUID, lowStockThreshold decimal.Decimal) error {
	if f.updateThresholdsFunc != nil {
		return f.updateThresholdsFunc(ctx, productID, lowStockThreshold)
	}
	return nil
}
func (f *fakeInventoryStore) GetLowStockProducts(ctx context.Context) ([]*models.ProductWithInventory, error) {
	if f.getLowStockProductsFunc != nil {
		return f.getLowStockProductsFunc(ctx)
	}
	return nil, nil
}
func (f *fakeInventoryStore) GetAdjustmentHistory(ctx context.Context, productID uuid.UUID, limit int, offset int, adjustmentType *models.AdjustmentType, adjustedBy *uuid.UUID, startDate, endDate *time.Time) ([]*models.StockAdjustment, int, error) {
	if f.getAdjustmentHistoryFunc != nil {
		return f.getAdjustmentHistoryFunc(ctx, productID, limit, offset, adjustmentType, adjustedBy, startDate, endDate)
	}
	return nil, 0, nil
}

type fakeCategoryStore struct {
	getByIDFunc func(context.Context, uuid.UUID) (*models.Category, error)
}

func (f *fakeCategoryStore) GetByID(ctx context.Context, id uuid.UUID) (*models.Category, error) {
	if f.getByIDFunc != nil {
		return f.getByIDFunc(ctx, id)
	}
	return nil, nil
}

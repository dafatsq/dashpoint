package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"dashpoint/backend/internal/models"
)

type fakeExpenseStore struct {
	getCategoryByIDFunc func(context.Context, uuid.UUID) (*models.ExpenseCategory, error)
	deleteCategoryFunc  func(context.Context, uuid.UUID) error
}

func (f *fakeExpenseStore) ListCategories(context.Context, string) ([]models.ExpenseCategory, error) {
	return nil, nil
}

func (f *fakeExpenseStore) CategoryNameExists(context.Context, string, *uuid.UUID) (bool, error) {
	return false, nil
}

func (f *fakeExpenseStore) CreateCategory(context.Context, string, *string) (*models.ExpenseCategory, error) {
	return nil, nil
}
func (f *fakeExpenseStore) GetCategoryByID(ctx context.Context, id uuid.UUID) (*models.ExpenseCategory, error) {
	if f.getCategoryByIDFunc != nil {
		return f.getCategoryByIDFunc(ctx, id)
	}
	return nil, nil
}
func (f *fakeExpenseStore) UpdateCategory(context.Context, *models.ExpenseCategory) (*models.ExpenseCategory, error) {
	return nil, nil
}
func (f *fakeExpenseStore) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	if f.deleteCategoryFunc != nil {
		return f.deleteCategoryFunc(ctx, id)
	}
	return nil
}
func (f *fakeExpenseStore) PermanentDeleteCategory(context.Context, uuid.UUID) error {
	return nil
}
func (f *fakeExpenseStore) Create(context.Context, *models.Expense) (*models.Expense, error) {
	return nil, nil
}
func (f *fakeExpenseStore) BeginTx(context.Context) (pgx.Tx, error) {
	return nil, nil
}
func (f *fakeExpenseStore) CreateWithTx(context.Context, pgx.Tx, *models.Expense) (*models.Expense, error) {
	return nil, nil
}
func (f *fakeExpenseStore) GetByIDWithTx(context.Context, pgx.Tx, uuid.UUID) (*models.Expense, error) {
	return nil, nil
}
func (f *fakeExpenseStore) GetByID(context.Context, uuid.UUID) (*models.Expense, error) {
	return nil, nil
}
func (f *fakeExpenseStore) List(context.Context, *uuid.UUID, *time.Time, *time.Time, int, int) ([]models.Expense, int, error) {
	return nil, 0, nil
}
func (f *fakeExpenseStore) Update(context.Context, *models.Expense) (*models.Expense, error) {
	return nil, nil
}
func (f *fakeExpenseStore) UpdateWithTx(context.Context, pgx.Tx, *models.Expense) (*models.Expense, error) {
	return nil, nil
}
func (f *fakeExpenseStore) Delete(context.Context, uuid.UUID) error {
	return nil
}
func (f *fakeExpenseStore) DeleteWithTx(context.Context, pgx.Tx, uuid.UUID) error {
	return nil
}
func (f *fakeExpenseStore) GetSummary(context.Context, time.Time, time.Time) (*models.ExpenseSummary, error) {
	return nil, nil
}
func (f *fakeExpenseStore) GetMonthlyTotals(context.Context, int) ([]map[string]interface{}, error) {
	return nil, nil
}

type fakeInventoryAdjuster struct {
	adjustStockWithTxFunc func(context.Context, pgx.Tx, uuid.UUID, models.AdjustmentType, decimal.Decimal, *string, *string, *uuid.UUID, uuid.UUID) (*models.StockAdjustment, error)
}

func (f *fakeInventoryAdjuster) AdjustStockWithTx(ctx context.Context, tx pgx.Tx, productID uuid.UUID, adjustmentType models.AdjustmentType, quantity decimal.Decimal, reason *string, referenceType *string, referenceID *uuid.UUID, adjustedBy uuid.UUID) (*models.StockAdjustment, error) {
	if f.adjustStockWithTxFunc != nil {
		return f.adjustStockWithTxFunc(ctx, tx, productID, adjustmentType, quantity, reason, referenceType, referenceID, adjustedBy)
	}
	return nil, nil
}

func TestListExpensesRejectsInvalidCategoryID(t *testing.T) {
	handler := NewExpenseHandler(&fakeExpenseStore{}, &fakeInventoryAdjuster{}, nil)
	app := fiber.New()
	app.Get("/expenses", handler.List)

	req := httptest.NewRequest(http.MethodGet, "/expenses?category_id=bad-uuid", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestCreateExpenseRejectsInvalidCategoryID(t *testing.T) {
	handler := NewExpenseHandler(&fakeExpenseStore{}, &fakeInventoryAdjuster{}, nil)
	app := expenseHandlerTestApp(handler)

	resp := performExpenseJSONRequest(t, app, http.MethodPost, "/expenses", `{"amount":"10.00","description":"Taxi","expense_date":"2026-05-14","category_id":"bad-uuid"}`)
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestCreateExpenseRejectsInvalidProductID(t *testing.T) {
	handler := NewExpenseHandler(&fakeExpenseStore{}, &fakeInventoryAdjuster{}, nil)
	app := expenseHandlerTestApp(handler)

	resp := performExpenseJSONRequest(t, app, http.MethodPost, "/expenses", `{"amount":"10.00","description":"Taxi","expense_date":"2026-05-14","product_id":"bad-uuid"}`)
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestCreateExpenseRejectsInvalidQuantity(t *testing.T) {
	handler := NewExpenseHandler(&fakeExpenseStore{}, &fakeInventoryAdjuster{}, nil)
	app := expenseHandlerTestApp(handler)

	resp := performExpenseJSONRequest(t, app, http.MethodPost, "/expenses", `{"amount":"10.00","description":"Taxi","expense_date":"2026-05-14","quantity":"oops"}`)
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestCreateExpenseModelRequiresProductAndQuantityForInventoryPurchase(t *testing.T) {
	categoryID := uuid.New()
	handler := NewExpenseHandler(&fakeExpenseStore{
		getCategoryByIDFunc: func(context.Context, uuid.UUID) (*models.ExpenseCategory, error) {
			systemKey := inventoryPurchaseCategorySystemKey
			return &models.ExpenseCategory{ID: categoryID, Name: "Inventory Purchase", SystemKey: &systemKey}, nil
		},
	}, &fakeInventoryAdjuster{}, nil)

	_, err := handler.createExpenseModel(context.Background(), CreateExpenseRequest{
		CategoryID:  stringPtr(categoryID.String()),
		Amount:      "10.00",
		Description: "Restock",
		ExpenseDate: "2026-05-14",
	}, uuid.New())
	if err == nil || err.Error() != "Product is required for Inventory Purchase" {
		t.Fatalf("expected product required error, got %v", err)
	}

	productID := uuid.New().String()
	_, err = handler.createExpenseModel(context.Background(), CreateExpenseRequest{
		CategoryID:  stringPtr(categoryID.String()),
		ProductID:   &productID,
		Amount:      "10.00",
		Description: "Restock",
		ExpenseDate: "2026-05-14",
	}, uuid.New())
	if err == nil || err.Error() != "Quantity is required for Inventory Purchase" {
		t.Fatalf("expected quantity required error, got %v", err)
	}
}

func TestCreateExpenseModelSetsAppliesInventoryOnlyWhenRequested(t *testing.T) {
	categoryID := uuid.New()
	productID := uuid.New().String()
	quantity := "2"
	handler := NewExpenseHandler(&fakeExpenseStore{
		getCategoryByIDFunc: func(context.Context, uuid.UUID) (*models.ExpenseCategory, error) {
			systemKey := inventoryPurchaseCategorySystemKey
			return &models.ExpenseCategory{ID: categoryID, Name: "Inventory Purchase", SystemKey: &systemKey}, nil
		},
	}, &fakeInventoryAdjuster{}, nil)

	expense, err := handler.createExpenseModel(context.Background(), CreateExpenseRequest{
		CategoryID:       stringPtr(categoryID.String()),
		ProductID:        &productID,
		Quantity:         &quantity,
		AppliesInventory: false,
		Amount:           "10.00",
		Description:      "Restock",
		ExpenseDate:      "2026-05-14",
	}, uuid.New())
	if err != nil {
		t.Fatalf("createExpenseModel returned error: %v", err)
	}
	if expense.AppliesInventory {
		t.Fatalf("expected applies_inventory false")
	}
	if expense.ProductID == nil || expense.Quantity == nil {
		t.Fatalf("expected product and quantity to be preserved")
	}
}

func TestDeleteCategoryAllowsArchivingInventoryPurchaseCategory(t *testing.T) {
	categoryID := uuid.New()
	deleteCalled := false
	store := &fakeExpenseStore{
		getCategoryByIDFunc: func(context.Context, uuid.UUID) (*models.ExpenseCategory, error) {
			systemKey := inventoryPurchaseCategorySystemKey
			return &models.ExpenseCategory{ID: categoryID, Name: "Renamed Inventory Purchase", SystemKey: &systemKey, IsActive: true}, nil
		},
		deleteCategoryFunc: func(_ context.Context, id uuid.UUID) error {
			deleteCalled = true
			if id != categoryID {
				t.Fatalf("expected delete category id %s, got %s", categoryID, id)
			}
			return nil
		},
	}
	handler := NewExpenseHandler(store, &fakeInventoryAdjuster{}, nil)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New())
		return c.Next()
	})
	app.Delete("/expenses/categories/:id", func(c *fiber.Ctx) error {
		return handler.DeleteCategory(c)
	})

	req := httptest.NewRequest(http.MethodDelete, "/expenses/categories/"+categoryID.String(), nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if !deleteCalled {
		t.Fatalf("expected delete repository method to be called")
	}
}

func expenseHandlerTestApp(handler *ExpenseHandler) *fiber.App {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New())
		return c.Next()
	})
	app.Get("/expenses", handler.List)
	app.Post("/expenses", handler.Create)
	return app
}

func performExpenseJSONRequest(t *testing.T, app *fiber.App, method, path, body string) *http.Response {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	return resp
}

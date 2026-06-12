package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shopspring/decimal"

	"dashpoint/backend/internal/models"
)

type fakeExpenseStore struct {
	listCategoriesFunc          func(context.Context, string) ([]models.ExpenseCategory, error)
	getCategoryByIDFunc         func(context.Context, uuid.UUID) (*models.ExpenseCategory, error)
	deleteCategoryFunc          func(context.Context, uuid.UUID) error
	permanentDeleteCategoryFunc func(context.Context, uuid.UUID) error
	updateCategoryFunc          func(context.Context, *models.ExpenseCategory) (*models.ExpenseCategory, error)
	beginTxFunc                 func(context.Context) (pgx.Tx, error)
	getByIDFunc                 func(context.Context, uuid.UUID) (*models.Expense, error)
	getByIDWithTxFunc           func(context.Context, pgx.Tx, uuid.UUID) (*models.Expense, error)
	updateWithTxFunc            func(context.Context, pgx.Tx, *models.Expense) (*models.Expense, error)
	deleteWithTxFunc            func(context.Context, pgx.Tx, uuid.UUID) error
}

func (f *fakeExpenseStore) ListCategories(ctx context.Context, status string) ([]models.ExpenseCategory, error) {
	if f.listCategoriesFunc != nil {
		return f.listCategoriesFunc(ctx, status)
	}
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
func (f *fakeExpenseStore) UpdateCategory(ctx context.Context, category *models.ExpenseCategory) (*models.ExpenseCategory, error) {
	if f.updateCategoryFunc != nil {
		return f.updateCategoryFunc(ctx, category)
	}
	return nil, nil
}
func (f *fakeExpenseStore) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	if f.deleteCategoryFunc != nil {
		return f.deleteCategoryFunc(ctx, id)
	}
	return nil
}
func (f *fakeExpenseStore) PermanentDeleteCategory(ctx context.Context, id uuid.UUID) error {
	if f.permanentDeleteCategoryFunc != nil {
		return f.permanentDeleteCategoryFunc(ctx, id)
	}
	return nil
}
func (f *fakeExpenseStore) Create(context.Context, *models.Expense) (*models.Expense, error) {
	return nil, nil
}
func (f *fakeExpenseStore) BeginTx(ctx context.Context) (pgx.Tx, error) {
	if f.beginTxFunc != nil {
		return f.beginTxFunc(ctx)
	}
	return nil, nil
}
func (f *fakeExpenseStore) CreateWithTx(context.Context, pgx.Tx, *models.Expense) (*models.Expense, error) {
	return nil, nil
}
func (f *fakeExpenseStore) GetByIDWithTx(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*models.Expense, error) {
	if f.getByIDWithTxFunc != nil {
		return f.getByIDWithTxFunc(ctx, tx, id)
	}
	return nil, nil
}
func (f *fakeExpenseStore) GetByID(ctx context.Context, id uuid.UUID) (*models.Expense, error) {
	if f.getByIDFunc != nil {
		return f.getByIDFunc(ctx, id)
	}
	return nil, nil
}
func (f *fakeExpenseStore) List(context.Context, *uuid.UUID, *time.Time, *time.Time, int, int) ([]models.Expense, int, error) {
	return nil, 0, nil
}
func (f *fakeExpenseStore) Update(context.Context, *models.Expense) (*models.Expense, error) {
	return nil, nil
}
func (f *fakeExpenseStore) UpdateWithTx(ctx context.Context, tx pgx.Tx, expense *models.Expense) (*models.Expense, error) {
	if f.updateWithTxFunc != nil {
		return f.updateWithTxFunc(ctx, tx, expense)
	}
	return nil, nil
}
func (f *fakeExpenseStore) Delete(context.Context, uuid.UUID) error {
	return nil
}
func (f *fakeExpenseStore) DeleteWithTx(ctx context.Context, tx pgx.Tx, id uuid.UUID) error {
	if f.deleteWithTxFunc != nil {
		return f.deleteWithTxFunc(ctx, tx, id)
	}
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

type fakeExpenseProductStore struct {
	getByIDFunc func(context.Context, uuid.UUID) (*models.Product, error)
}

type fakeExpenseTx struct {
	committed bool
}

func (f *fakeExpenseTx) Begin(context.Context) (pgx.Tx, error) { return f, nil }
func (f *fakeExpenseTx) Commit(context.Context) error {
	f.committed = true
	return nil
}
func (f *fakeExpenseTx) Rollback(context.Context) error { return nil }
func (f *fakeExpenseTx) CopyFrom(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) (int64, error) {
	return 0, nil
}
func (f *fakeExpenseTx) SendBatch(context.Context, *pgx.Batch) pgx.BatchResults { return nil }
func (f *fakeExpenseTx) LargeObjects() pgx.LargeObjects                         { return pgx.LargeObjects{} }
func (f *fakeExpenseTx) Prepare(context.Context, string, string) (*pgconn.StatementDescription, error) {
	return nil, nil
}
func (f *fakeExpenseTx) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}
func (f *fakeExpenseTx) Query(context.Context, string, ...any) (pgx.Rows, error) { return nil, nil }
func (f *fakeExpenseTx) QueryRow(context.Context, string, ...any) pgx.Row        { return nil }
func (f *fakeExpenseTx) Conn() *pgx.Conn                                         { return nil }

func (f *fakeExpenseProductStore) GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	if f.getByIDFunc != nil {
		return f.getByIDFunc(ctx, id)
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

func TestListExpensesRejectsInvalidPagination(t *testing.T) {
	handler := NewExpenseHandler(&fakeExpenseStore{}, &fakeInventoryAdjuster{}, nil)
	app := fiber.New()
	app.Get("/expenses", handler.List)

	for _, path := range []string{"/expenses?limit=101", "/expenses?offset=-1", "/expenses?limit=abc"} {
		resp, err := app.Test(httptest.NewRequest(http.MethodGet, path, nil))
		if err != nil {
			t.Fatalf("app.Test returned error: %v", err)
		}
		if resp.StatusCode != fiber.StatusBadRequest {
			t.Fatalf("expected 400 for %s, got %d", path, resp.StatusCode)
		}
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

func TestCreateExpenseRejectsUnknownField(t *testing.T) {
	handler := NewExpenseHandler(&fakeExpenseStore{}, &fakeInventoryAdjuster{}, nil)
	app := expenseHandlerTestApp(handler)

	resp := performExpenseJSONRequest(t, app, http.MethodPost, "/expenses", `{"amount":"10.00","description":"Taxi","expense_date":"2026-05-14","unexpected":true}`)
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestCreateExpenseRejectsOversizedVendor(t *testing.T) {
	handler := NewExpenseHandler(&fakeExpenseStore{}, &fakeInventoryAdjuster{}, nil)
	app := expenseHandlerTestApp(handler)

	body := `{"amount":"10.00","description":"Taxi","expense_date":"2026-05-14","vendor":"` + strings.Repeat("x", expenseVendorMaxLength+1) + `"}`
	resp := performExpenseJSONRequest(t, app, http.MethodPost, "/expenses", body)
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestCreateExpenseModelRejectsArchivedExpenseCategory(t *testing.T) {
	categoryID := uuid.New()
	handler := NewExpenseHandler(&fakeExpenseStore{
		getCategoryByIDFunc: func(context.Context, uuid.UUID) (*models.ExpenseCategory, error) {
			return &models.ExpenseCategory{ID: categoryID, Name: "Archived", IsActive: false}, nil
		},
	}, &fakeInventoryAdjuster{}, nil)

	_, err := handler.createExpenseModel(context.Background(), CreateExpenseRequest{
		CategoryID:  stringPtr(categoryID.String()),
		Amount:      "10.00",
		Description: "Taxi",
		ExpenseDate: "2026-05-14",
	}, uuid.New())
	if err == nil || err.Error() != "Archived expense categories cannot be used" {
		t.Fatalf("expected archived category error, got %v", err)
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

func TestCreateInventoryPurchaseRequiresExpectedProductUpdatedAt(t *testing.T) {
	categoryID := uuid.New()
	productID := uuid.New()
	systemKey := inventoryPurchaseCategorySystemKey
	store := &fakeExpenseStore{
		getCategoryByIDFunc: func(context.Context, uuid.UUID) (*models.ExpenseCategory, error) {
			return &models.ExpenseCategory{ID: categoryID, Name: "Inventory Purchase", SystemKey: &systemKey, IsActive: true}, nil
		},
	}
	handler := NewExpenseHandler(store, &fakeInventoryAdjuster{}, &fakeExpenseProductStore{
		getByIDFunc: func(context.Context, uuid.UUID) (*models.Product, error) {
			product := testProduct()
			product.ID = productID
			product.UpdatedAt = time.Date(2026, 5, 14, 10, 0, 0, 0, time.UTC)
			return product, nil
		},
	})
	app := expenseHandlerTestApp(handler)

	body := `{"category_id":"` + categoryID.String() + `","product_id":"` + productID.String() + `","quantity":"2","applies_inventory":true,"amount":"10.00","description":"Restock","expense_date":"2026-05-14"}`
	resp := performExpenseJSONRequest(t, app, http.MethodPost, "/expenses", body)
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestCreateInventoryPurchaseRejectsStaleProduct(t *testing.T) {
	categoryID := uuid.New()
	productID := uuid.New()
	oldUpdatedAt := time.Date(2026, 5, 14, 10, 0, 0, 0, time.UTC)
	newUpdatedAt := oldUpdatedAt.Add(time.Minute)
	systemKey := inventoryPurchaseCategorySystemKey
	store := &fakeExpenseStore{
		getCategoryByIDFunc: func(context.Context, uuid.UUID) (*models.ExpenseCategory, error) {
			return &models.ExpenseCategory{ID: categoryID, Name: "Inventory Purchase", SystemKey: &systemKey, IsActive: true}, nil
		},
	}
	handler := NewExpenseHandler(store, &fakeInventoryAdjuster{}, &fakeExpenseProductStore{
		getByIDFunc: func(context.Context, uuid.UUID) (*models.Product, error) {
			product := testProduct()
			product.ID = productID
			product.UpdatedAt = newUpdatedAt
			return product, nil
		},
	})
	app := expenseHandlerTestApp(handler)

	body := `{"category_id":"` + categoryID.String() + `","product_id":"` + productID.String() + `","quantity":"2","applies_inventory":true,"amount":"10.00","description":"Restock","expense_date":"2026-05-14","expected_product_updated_at":"` + oldUpdatedAt.Format(time.RFC3339Nano) + `"}`
	resp := performExpenseJSONRequest(t, app, http.MethodPost, "/expenses", body)
	if resp.StatusCode != fiber.StatusConflict {
		t.Fatalf("expected 409, got %d", resp.StatusCode)
	}
}

func TestUpdateInventoryPurchaseRejectsStaleProduct(t *testing.T) {
	expenseID := uuid.New()
	categoryID := uuid.New()
	productID := uuid.New()
	userID := uuid.New()
	expenseUpdatedAt := time.Date(2026, 5, 14, 10, 0, 0, 0, time.UTC)
	oldProductUpdatedAt := time.Date(2026, 5, 14, 9, 0, 0, 0, time.UTC)
	newProductUpdatedAt := oldProductUpdatedAt.Add(time.Minute)
	systemKey := inventoryPurchaseCategorySystemKey
	store := &fakeExpenseStore{
		beginTxFunc: func(context.Context) (pgx.Tx, error) {
			return &fakeExpenseTx{}, nil
		},
		getByIDWithTxFunc: func(context.Context, pgx.Tx, uuid.UUID) (*models.Expense, error) {
			return &models.Expense{
				ID:               expenseID,
				CategoryID:       &categoryID,
				ProductID:        &productID,
				Quantity:         decimalPtr("2"),
				AppliesInventory: true,
				Amount:           decimal.RequireFromString("10"),
				Description:      "Old restock",
				ExpenseDate:      time.Date(2026, 5, 14, 0, 0, 0, 0, time.UTC),
				CreatedBy:        userID,
				UpdatedAt:        expenseUpdatedAt,
			}, nil
		},
		getCategoryByIDFunc: func(context.Context, uuid.UUID) (*models.ExpenseCategory, error) {
			return &models.ExpenseCategory{ID: categoryID, Name: "Inventory Purchase", SystemKey: &systemKey, IsActive: true}, nil
		},
		updateWithTxFunc: func(context.Context, pgx.Tx, *models.Expense) (*models.Expense, error) {
			t.Fatal("expected update not to be called with stale product")
			return nil, nil
		},
	}
	handler := NewExpenseHandler(store, &fakeInventoryAdjuster{}, &fakeExpenseProductStore{
		getByIDFunc: func(context.Context, uuid.UUID) (*models.Product, error) {
			product := testProduct()
			product.ID = productID
			product.UpdatedAt = newProductUpdatedAt
			return product, nil
		},
	})
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		return c.Next()
	})
	app.Patch("/expenses/:id", handler.Update)

	body := `{"description":"Updated restock","expected_updated_at":"` + expenseUpdatedAt.Format(time.RFC3339Nano) + `","expected_product_updated_at":"` + oldProductUpdatedAt.Format(time.RFC3339Nano) + `"}`
	req := httptest.NewRequest(http.MethodPatch, "/expenses/"+expenseID.String(), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusConflict {
		t.Fatalf("expected 409, got %d", resp.StatusCode)
	}
}

func TestUpdateExpenseRequiresExpectedUpdatedAt(t *testing.T) {
	expenseID := uuid.New()
	userID := uuid.New()
	updatedAt := time.Date(2026, 5, 14, 10, 0, 0, 0, time.UTC)
	store := &fakeExpenseStore{
		beginTxFunc: func(context.Context) (pgx.Tx, error) {
			return &fakeExpenseTx{}, nil
		},
		getByIDWithTxFunc: func(context.Context, pgx.Tx, uuid.UUID) (*models.Expense, error) {
			return &models.Expense{
				ID:          expenseID,
				Amount:      decimal.RequireFromString("10"),
				Description: "Old description",
				ExpenseDate: time.Date(2026, 5, 14, 0, 0, 0, 0, time.UTC),
				CreatedBy:   userID,
				UpdatedAt:   updatedAt,
			}, nil
		},
		updateWithTxFunc: func(context.Context, pgx.Tx, *models.Expense) (*models.Expense, error) {
			t.Fatal("expected update not to be called without expected_updated_at")
			return nil, nil
		},
	}
	handler := NewExpenseHandler(store, &fakeInventoryAdjuster{}, nil)
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		return c.Next()
	})
	app.Patch("/expenses/:id", handler.Update)

	req := httptest.NewRequest(http.MethodPatch, "/expenses/"+expenseID.String(), strings.NewReader(`{"description":"New description"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestDeleteExpenseRequiresExpectedUpdatedAt(t *testing.T) {
	expenseID := uuid.New()
	userID := uuid.New()
	updatedAt := time.Date(2026, 5, 14, 10, 0, 0, 0, time.UTC)
	store := &fakeExpenseStore{
		beginTxFunc: func(context.Context) (pgx.Tx, error) {
			return &fakeExpenseTx{}, nil
		},
		getByIDWithTxFunc: func(context.Context, pgx.Tx, uuid.UUID) (*models.Expense, error) {
			return &models.Expense{
				ID:          expenseID,
				Amount:      decimal.RequireFromString("10"),
				Description: "Old description",
				ExpenseDate: time.Date(2026, 5, 14, 0, 0, 0, 0, time.UTC),
				CreatedBy:   userID,
				UpdatedAt:   updatedAt,
			}, nil
		},
		deleteWithTxFunc: func(context.Context, pgx.Tx, uuid.UUID) error {
			t.Fatal("expected delete not to be called without expected_updated_at")
			return nil
		},
	}
	handler := NewExpenseHandler(store, &fakeInventoryAdjuster{}, nil)
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		return c.Next()
	})
	app.Delete("/expenses/:id", handler.Delete)

	resp, err := app.Test(httptest.NewRequest(http.MethodDelete, "/expenses/"+expenseID.String(), nil))
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestCreateExpenseModelRequiresProductAndQuantityForInventoryPurchase(t *testing.T) {
	categoryID := uuid.New()
	handler := NewExpenseHandler(&fakeExpenseStore{
		getCategoryByIDFunc: func(context.Context, uuid.UUID) (*models.ExpenseCategory, error) {
			systemKey := inventoryPurchaseCategorySystemKey
			return &models.ExpenseCategory{ID: categoryID, Name: "Inventory Purchase", SystemKey: &systemKey, IsActive: true}, nil
		},
	}, &fakeInventoryAdjuster{}, &fakeExpenseProductStore{
		getByIDFunc: func(context.Context, uuid.UUID) (*models.Product, error) {
			return testProduct(), nil
		},
	})

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
			return &models.ExpenseCategory{ID: categoryID, Name: "Inventory Purchase", SystemKey: &systemKey, IsActive: true}, nil
		},
	}, &fakeInventoryAdjuster{}, &fakeExpenseProductStore{
		getByIDFunc: func(context.Context, uuid.UUID) (*models.Product, error) {
			return testProduct(), nil
		},
	})

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

func TestCreateExpenseModelRejectsArchivedProduct(t *testing.T) {
	categoryID := uuid.New()
	productID := uuid.New().String()
	quantity := "2"
	handler := NewExpenseHandler(&fakeExpenseStore{
		getCategoryByIDFunc: func(context.Context, uuid.UUID) (*models.ExpenseCategory, error) {
			systemKey := inventoryPurchaseCategorySystemKey
			return &models.ExpenseCategory{ID: categoryID, Name: "Inventory Purchase", SystemKey: &systemKey, IsActive: true}, nil
		},
	}, &fakeInventoryAdjuster{}, &fakeExpenseProductStore{
		getByIDFunc: func(context.Context, uuid.UUID) (*models.Product, error) {
			product := testProduct()
			product.IsActive = false
			return product, nil
		},
	})

	_, err := handler.createExpenseModel(context.Background(), CreateExpenseRequest{
		CategoryID:       stringPtr(categoryID.String()),
		ProductID:        &productID,
		Quantity:         &quantity,
		AppliesInventory: true,
		Amount:           "10.00",
		Description:      "Restock",
		ExpenseDate:      "2026-05-14",
	}, uuid.New())
	if err == nil || err.Error() != "Archived products cannot be changed" {
		t.Fatalf("expected archived product error, got %v", err)
	}
}

func TestExpenseInventoryReason(t *testing.T) {
	expenseID := uuid.MustParse("00000000-0000-0000-0000-000000000321")

	if got := expenseInventoryReason("purchase", expenseID); got == nil || *got != "Expense inventory purchase 00000000-0000-0000-0000-000000000321" {
		t.Fatalf("unexpected purchase reason: %v", got)
	}
	if got := expenseInventoryReason("delete revert", expenseID); got == nil || *got != "Expense inventory delete revert 00000000-0000-0000-0000-000000000321" {
		t.Fatalf("unexpected delete revert reason: %v", got)
	}
}

func TestDeleteCategoryAllowsArchivingInventoryPurchaseCategory(t *testing.T) {
	categoryID := uuid.New()
	updatedAt := time.Date(2026, 5, 14, 10, 0, 0, 0, time.UTC)
	deleteCalled := false
	store := &fakeExpenseStore{
		getCategoryByIDFunc: func(context.Context, uuid.UUID) (*models.ExpenseCategory, error) {
			systemKey := inventoryPurchaseCategorySystemKey
			return &models.ExpenseCategory{ID: categoryID, Name: "Renamed Inventory Purchase", SystemKey: &systemKey, IsActive: true, UpdatedAt: updatedAt}, nil
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

	req := httptest.NewRequest(http.MethodDelete, "/expenses/categories/"+categoryID.String()+"?expected_updated_at="+url.QueryEscape(updatedAt.Format(time.RFC3339Nano)), nil)
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

func TestListExpenseCategoriesRejectsInvalidStatus(t *testing.T) {
	store := &fakeExpenseStore{
		listCategoriesFunc: func(context.Context, string) ([]models.ExpenseCategory, error) {
			t.Fatal("expected list categories not to be called for invalid status")
			return nil, nil
		},
	}
	handler := NewExpenseHandler(store, &fakeInventoryAdjuster{}, nil)
	app := fiber.New()
	app.Get("/expenses/categories", handler.ListCategories)

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/expenses/categories?status=unknown", nil))
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestSyncExpenseInventoryRejectsArchivedProduct(t *testing.T) {
	expenseID := uuid.New()
	productID := uuid.New()
	handler := NewExpenseHandler(&fakeExpenseStore{}, &fakeInventoryAdjuster{
		adjustStockWithTxFunc: func(context.Context, pgx.Tx, uuid.UUID, models.AdjustmentType, decimal.Decimal, *string, *string, *uuid.UUID, uuid.UUID) (*models.StockAdjustment, error) {
			t.Fatal("expected inventory adjustment not to run for archived product")
			return nil, nil
		},
	}, &fakeExpenseProductStore{
		getByIDFunc: func(context.Context, uuid.UUID) (*models.Product, error) {
			product := testProduct()
			product.ID = productID
			product.IsActive = false
			return product, nil
		},
	})

	err := handler.syncExpenseInventory(context.Background(), nil, expenseID, uuid.New(), &models.Expense{
		ID:               expenseID,
		ProductID:        &productID,
		Quantity:         decimalPtr("2"),
		AppliesInventory: true,
	}, false, nil, nil)
	if err == nil || err.Error() != "Archived products cannot be changed" {
		t.Fatalf("expected archived product error, got %v", err)
	}
}

func TestUpdateExpenseAllowsUnchangedArchivedExpenseCategory(t *testing.T) {
	expenseID := uuid.New()
	categoryID := uuid.New()
	userID := uuid.New()
	updatedAt := time.Date(2026, 5, 14, 10, 0, 0, 0, time.UTC)
	tx := &fakeExpenseTx{}
	updateCalled := false

	store := &fakeExpenseStore{
		beginTxFunc: func(context.Context) (pgx.Tx, error) {
			return tx, nil
		},
		getByIDWithTxFunc: func(context.Context, pgx.Tx, uuid.UUID) (*models.Expense, error) {
			return &models.Expense{
				ID:          expenseID,
				CategoryID:  &categoryID,
				Amount:      decimal.RequireFromString("10"),
				Description: "Old description",
				ExpenseDate: time.Date(2026, 5, 14, 0, 0, 0, 0, time.UTC),
				CreatedBy:   userID,
				UpdatedAt:   updatedAt,
			}, nil
		},
		getCategoryByIDFunc: func(context.Context, uuid.UUID) (*models.ExpenseCategory, error) {
			return &models.ExpenseCategory{ID: categoryID, Name: "Archived Supplies", IsActive: false}, nil
		},
		updateWithTxFunc: func(_ context.Context, _ pgx.Tx, expense *models.Expense) (*models.Expense, error) {
			updateCalled = true
			if expense.CategoryID == nil || *expense.CategoryID != categoryID {
				t.Fatalf("expected category id to stay %s, got %v", categoryID, expense.CategoryID)
			}
			if expense.Description != "New description" {
				t.Fatalf("expected description update, got %q", expense.Description)
			}
			return expense, nil
		},
	}
	handler := NewExpenseHandler(store, &fakeInventoryAdjuster{}, nil)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		return c.Next()
	})
	app.Patch("/expenses/:id", handler.Update)

	body := `{"description":"New description","category_id":"` + categoryID.String() + `","expected_updated_at":"` + updatedAt.Format(time.RFC3339Nano) + `"}`
	req := httptest.NewRequest(http.MethodPatch, "/expenses/"+expenseID.String(), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if !updateCalled {
		t.Fatalf("expected expense update to be called")
	}
	if !tx.committed {
		t.Fatalf("expected transaction to commit")
	}
}

func TestUpdateExpenseRejectsChangedArchivedExpenseCategory(t *testing.T) {
	expenseID := uuid.New()
	currentCategoryID := uuid.New()
	archivedCategoryID := uuid.New()
	userID := uuid.New()
	updatedAt := time.Date(2026, 5, 14, 10, 0, 0, 0, time.UTC)

	store := &fakeExpenseStore{
		beginTxFunc: func(context.Context) (pgx.Tx, error) {
			return &fakeExpenseTx{}, nil
		},
		getByIDWithTxFunc: func(context.Context, pgx.Tx, uuid.UUID) (*models.Expense, error) {
			return &models.Expense{
				ID:          expenseID,
				CategoryID:  &currentCategoryID,
				Amount:      decimal.RequireFromString("10"),
				Description: "Old description",
				ExpenseDate: time.Date(2026, 5, 14, 0, 0, 0, 0, time.UTC),
				CreatedBy:   userID,
				UpdatedAt:   updatedAt,
			}, nil
		},
		getCategoryByIDFunc: func(context.Context, uuid.UUID) (*models.ExpenseCategory, error) {
			return &models.ExpenseCategory{ID: archivedCategoryID, Name: "Archived Supplies", IsActive: false}, nil
		},
		updateWithTxFunc: func(context.Context, pgx.Tx, *models.Expense) (*models.Expense, error) {
			t.Fatal("expected update not to be called with changed archived category")
			return nil, nil
		},
	}
	handler := NewExpenseHandler(store, &fakeInventoryAdjuster{}, nil)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		return c.Next()
	})
	app.Patch("/expenses/:id", handler.Update)

	body := `{"description":"New description","category_id":"` + archivedCategoryID.String() + `","expected_updated_at":"` + updatedAt.Format(time.RFC3339Nano) + `"}`
	req := httptest.NewRequest(http.MethodPatch, "/expenses/"+expenseID.String(), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestUpdateExpenseCategoryRejectsArchivedCategoryWithoutRestore(t *testing.T) {
	categoryID := uuid.New()
	updatedAt := time.Date(2026, 5, 14, 10, 0, 0, 0, time.UTC)
	store := &fakeExpenseStore{
		getCategoryByIDFunc: func(context.Context, uuid.UUID) (*models.ExpenseCategory, error) {
			return &models.ExpenseCategory{ID: categoryID, Name: "Supplies", IsActive: false, UpdatedAt: updatedAt}, nil
		},
		updateCategoryFunc: func(context.Context, *models.ExpenseCategory) (*models.ExpenseCategory, error) {
			t.Fatal("expected update not to be called for archived expense category")
			return nil, nil
		},
	}
	handler := NewExpenseHandler(store, &fakeInventoryAdjuster{}, nil)

	app := fiber.New()
	app.Patch("/expenses/categories/:id", func(c *fiber.Ctx) error {
		return handler.UpdateCategory(c)
	})

	req := httptest.NewRequest(http.MethodPatch, "/expenses/categories/"+categoryID.String(), strings.NewReader(`{"name":"Office Supplies","expected_updated_at":"`+updatedAt.Format(time.RFC3339Nano)+`"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusConflict {
		t.Fatalf("expected 409, got %d", resp.StatusCode)
	}
}

func TestUpdateExpenseCategoryRequiresExpectedUpdatedAt(t *testing.T) {
	categoryID := uuid.New()
	store := &fakeExpenseStore{
		getCategoryByIDFunc: func(context.Context, uuid.UUID) (*models.ExpenseCategory, error) {
			return &models.ExpenseCategory{ID: categoryID, Name: "Supplies", IsActive: true, UpdatedAt: time.Date(2026, 5, 14, 10, 0, 0, 0, time.UTC)}, nil
		},
		updateCategoryFunc: func(context.Context, *models.ExpenseCategory) (*models.ExpenseCategory, error) {
			t.Fatal("expected update not to be called without expected_updated_at")
			return nil, nil
		},
	}
	handler := NewExpenseHandler(store, &fakeInventoryAdjuster{}, nil)

	app := fiber.New()
	app.Patch("/expenses/categories/:id", handler.UpdateCategory)

	req := httptest.NewRequest(http.MethodPatch, "/expenses/categories/"+categoryID.String(), strings.NewReader(`{"name":"Office Supplies"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestDeleteExpenseCategoryRejectsAlreadyArchivedCategory(t *testing.T) {
	categoryID := uuid.New()
	updatedAt := time.Date(2026, 5, 14, 10, 0, 0, 0, time.UTC)
	store := &fakeExpenseStore{
		getCategoryByIDFunc: func(context.Context, uuid.UUID) (*models.ExpenseCategory, error) {
			return &models.ExpenseCategory{ID: categoryID, Name: "Supplies", IsActive: false, UpdatedAt: updatedAt}, nil
		},
		deleteCategoryFunc: func(context.Context, uuid.UUID) error {
			t.Fatal("expected delete not to be called for archived expense category")
			return nil
		},
	}
	handler := NewExpenseHandler(store, &fakeInventoryAdjuster{}, nil)

	app := fiber.New()
	app.Delete("/expenses/categories/:id", func(c *fiber.Ctx) error {
		return handler.DeleteCategory(c)
	})

	resp, err := app.Test(httptest.NewRequest(http.MethodDelete, "/expenses/categories/"+categoryID.String()+"?expected_updated_at="+url.QueryEscape(updatedAt.Format(time.RFC3339Nano)), nil))
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusConflict {
		t.Fatalf("expected 409, got %d", resp.StatusCode)
	}
}

func TestDeleteExpenseCategoryRequiresExpectedUpdatedAt(t *testing.T) {
	categoryID := uuid.New()
	store := &fakeExpenseStore{
		getCategoryByIDFunc: func(context.Context, uuid.UUID) (*models.ExpenseCategory, error) {
			return &models.ExpenseCategory{ID: categoryID, Name: "Supplies", IsActive: true, UpdatedAt: time.Date(2026, 5, 14, 10, 0, 0, 0, time.UTC)}, nil
		},
		deleteCategoryFunc: func(context.Context, uuid.UUID) error {
			t.Fatal("expected delete not to be called without expected_updated_at")
			return nil
		},
	}
	handler := NewExpenseHandler(store, &fakeInventoryAdjuster{}, nil)

	app := fiber.New()
	app.Delete("/expenses/categories/:id", handler.DeleteCategory)

	resp, err := app.Test(httptest.NewRequest(http.MethodDelete, "/expenses/categories/"+categoryID.String(), nil))
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestPermanentDeleteExpenseCategoryRejectsActiveCategory(t *testing.T) {
	categoryID := uuid.New()
	updatedAt := time.Date(2026, 5, 14, 10, 0, 0, 0, time.UTC)
	store := &fakeExpenseStore{
		getCategoryByIDFunc: func(context.Context, uuid.UUID) (*models.ExpenseCategory, error) {
			return &models.ExpenseCategory{ID: categoryID, Name: "Supplies", IsActive: true, UpdatedAt: updatedAt}, nil
		},
		permanentDeleteCategoryFunc: func(context.Context, uuid.UUID) error {
			t.Fatal("expected permanent delete not to be called for active category")
			return nil
		},
	}
	handler := NewExpenseHandler(store, &fakeInventoryAdjuster{}, nil)

	app := fiber.New()
	app.Delete("/expenses/categories/:id/permanent", handler.PermanentDeleteCategory)

	resp, err := app.Test(httptest.NewRequest(http.MethodDelete, "/expenses/categories/"+categoryID.String()+"/permanent?expected_updated_at="+url.QueryEscape(updatedAt.Format(time.RFC3339Nano)), nil))
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusConflict {
		t.Fatalf("expected 409, got %d", resp.StatusCode)
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

func decimalPtr(value string) *decimal.Decimal {
	parsed := decimal.RequireFromString(value)
	return &parsed
}

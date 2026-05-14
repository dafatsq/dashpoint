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

type fakeExpenseStore struct{}

func (f *fakeExpenseStore) ListCategories(context.Context, string) ([]models.ExpenseCategory, error) {
	return nil, nil
}
func (f *fakeExpenseStore) CreateCategory(context.Context, string, *string) (*models.ExpenseCategory, error) {
	return nil, nil
}
func (f *fakeExpenseStore) GetCategoryByID(context.Context, uuid.UUID) (*models.ExpenseCategory, error) {
	return nil, nil
}
func (f *fakeExpenseStore) UpdateCategory(context.Context, *models.ExpenseCategory) (*models.ExpenseCategory, error) {
	return nil, nil
}
func (f *fakeExpenseStore) DeleteCategory(context.Context, uuid.UUID) error {
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

type fakeInventoryAdjuster struct{}

func (f *fakeInventoryAdjuster) AdjustStockWithTx(context.Context, pgx.Tx, uuid.UUID, models.AdjustmentType, decimal.Decimal, *string, *string, *uuid.UUID, uuid.UUID) (*models.StockAdjustment, error) {
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

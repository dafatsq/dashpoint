package handlers

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"dashpoint/backend/internal/models"
)

type expenseStore interface {
	ListCategories(context.Context, string) ([]models.ExpenseCategory, error)
	CreateCategory(context.Context, string, *string) (*models.ExpenseCategory, error)
	GetCategoryByID(context.Context, uuid.UUID) (*models.ExpenseCategory, error)
	UpdateCategory(context.Context, *models.ExpenseCategory) (*models.ExpenseCategory, error)
	DeleteCategory(context.Context, uuid.UUID) error
	PermanentDeleteCategory(context.Context, uuid.UUID) error
	Create(context.Context, *models.Expense) (*models.Expense, error)
	BeginTx(context.Context) (pgx.Tx, error)
	CreateWithTx(context.Context, pgx.Tx, *models.Expense) (*models.Expense, error)
	GetByIDWithTx(context.Context, pgx.Tx, uuid.UUID) (*models.Expense, error)
	GetByID(context.Context, uuid.UUID) (*models.Expense, error)
	List(context.Context, *uuid.UUID, *time.Time, *time.Time, int, int) ([]models.Expense, int, error)
	Update(context.Context, *models.Expense) (*models.Expense, error)
	UpdateWithTx(context.Context, pgx.Tx, *models.Expense) (*models.Expense, error)
	Delete(context.Context, uuid.UUID) error
	DeleteWithTx(context.Context, pgx.Tx, uuid.UUID) error
	GetSummary(context.Context, time.Time, time.Time) (*models.ExpenseSummary, error)
	GetMonthlyTotals(context.Context, int) ([]map[string]interface{}, error)
}

type expenseInventoryStore interface {
	AdjustStockWithTx(context.Context, pgx.Tx, uuid.UUID, models.AdjustmentType, decimal.Decimal, *string, *string, *uuid.UUID, uuid.UUID) (*models.StockAdjustment, error)
}

type expenseProductStore interface{}

// ExpenseHandler handles expense-related HTTP requests.
type ExpenseHandler struct {
	repo          expenseStore
	inventoryRepo expenseInventoryStore
}

// NewExpenseHandler creates a new expense handler.
func NewExpenseHandler(repo expenseStore, inventoryRepo expenseInventoryStore, _ expenseProductStore) *ExpenseHandler {
	return &ExpenseHandler{
		repo:          repo,
		inventoryRepo: inventoryRepo,
	}
}

type CreateExpenseRequest struct {
	CategoryID      *string `json:"category_id"`
	ProductID       *string `json:"product_id"`
	Quantity        *string `json:"quantity"`
	Amount          string  `json:"amount"`
	Description     string  `json:"description"`
	ExpenseDate     string  `json:"expense_date"`
	Vendor          *string `json:"vendor"`
	ReferenceNumber *string `json:"reference_number"`
	Notes           *string `json:"notes"`
}

type UpdateExpenseRequest struct {
	CategoryID      *string `json:"category_id"`
	ProductID       *string `json:"product_id"`
	Quantity        *string `json:"quantity"`
	Amount          *string `json:"amount"`
	Description     *string `json:"description"`
	ExpenseDate     *string `json:"expense_date"`
	Vendor          *string `json:"vendor"`
	ReferenceNumber *string `json:"reference_number"`
	Notes           *string `json:"notes"`
}

type ExpenseResponse struct {
	ID              string  `json:"id"`
	CategoryID      *string `json:"category_id,omitempty"`
	CategoryName    *string `json:"category_name,omitempty"`
	ProductID       *string `json:"product_id,omitempty"`
	ProductName     *string `json:"product_name,omitempty"`
	Quantity        *string `json:"quantity,omitempty"`
	Amount          string  `json:"amount"`
	Description     string  `json:"description"`
	ExpenseDate     string  `json:"expense_date"`
	Vendor          *string `json:"vendor,omitempty"`
	ReferenceNumber *string `json:"reference_number,omitempty"`
	Notes           *string `json:"notes,omitempty"`
	CreatedBy       string  `json:"created_by"`
	CreatedByName   *string `json:"created_by_name,omitempty"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

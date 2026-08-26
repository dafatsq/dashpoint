package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"dashpoint/backend/internal/models"
)

// Create creates a new expense.
func (r *ExpenseRepository) Create(ctx context.Context, expense *models.Expense) (*models.Expense, error) {
	expense.ID = uuid.New()
	now := time.Now()
	expense.CreatedAt = now
	expense.UpdatedAt = now

	query := `
		INSERT INTO expenses (id, category_id, product_id, quantity, applies_inventory, amount, description, expense_date, vendor, reference_number, notes, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $13)
	`

	_, err := r.pool.Exec(ctx, query,
		expense.ID, expense.CategoryID, expense.ProductID, expense.Quantity, expense.AppliesInventory, expense.Amount, expense.Description,
		expense.ExpenseDate, expense.Vendor, expense.ReferenceNumber, expense.Notes, expense.CreatedBy, now,
	)
	if err != nil {
		return nil, NewInternalError(fmt.Errorf("failed to create expense: %w", err))
	}
	return r.GetByID(ctx, expense.ID)
}

// BeginTx starts a new database transaction.
func (r *ExpenseRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.pool.Begin(ctx)
}

// CreateWithTx creates a new expense within a transaction.
func (r *ExpenseRepository) CreateWithTx(ctx context.Context, tx pgx.Tx, expense *models.Expense) (*models.Expense, error) {
	expense.ID = uuid.New()
	now := time.Now()
	expense.CreatedAt = now
	expense.UpdatedAt = now

	query := `
		INSERT INTO expenses (id, category_id, product_id, quantity, applies_inventory, amount, description, expense_date, vendor, reference_number, notes, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $13)
	`

	_, err := tx.Exec(ctx, query,
		expense.ID, expense.CategoryID, expense.ProductID, expense.Quantity, expense.AppliesInventory, expense.Amount, expense.Description,
		expense.ExpenseDate, expense.Vendor, expense.ReferenceNumber, expense.Notes, expense.CreatedBy, now,
	)
	if err != nil {
		return nil, NewInternalError(fmt.Errorf("failed to create expense: %w", err))
	}
	return r.GetByIDWithTx(ctx, tx, expense.ID)
}

// Update updates an expense.
func (r *ExpenseRepository) Update(ctx context.Context, expense *models.Expense) (*models.Expense, error) {
	expense.UpdatedAt = time.Now()
	query := `
		UPDATE expenses
		SET category_id = $2, product_id = $3, quantity = $4, applies_inventory = $5, amount = $6, description = $7, expense_date = $8,
		    vendor = $9, reference_number = $10, notes = $11, updated_at = $12
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query,
		expense.ID, expense.CategoryID, expense.ProductID, expense.Quantity, expense.AppliesInventory, expense.Amount, expense.Description,
		expense.ExpenseDate, expense.Vendor, expense.ReferenceNumber, expense.Notes, expense.UpdatedAt,
	)
	if err != nil {
		return nil, NewInternalError(fmt.Errorf("failed to update expense: %w", err))
	}
	if result.RowsAffected() == 0 {
		return nil, fmt.Errorf("expense not found")
	}
	return r.GetByID(ctx, expense.ID)
}

// UpdateWithTx updates an expense within a transaction.
func (r *ExpenseRepository) UpdateWithTx(ctx context.Context, tx pgx.Tx, expense *models.Expense) (*models.Expense, error) {
	expense.UpdatedAt = time.Now()
	query := `
		UPDATE expenses
		SET category_id = $2, product_id = $3, quantity = $4, applies_inventory = $5, amount = $6, description = $7, expense_date = $8,
		    vendor = $9, reference_number = $10, notes = $11, updated_at = $12
		WHERE id = $1
	`

	result, err := tx.Exec(ctx, query,
		expense.ID, expense.CategoryID, expense.ProductID, expense.Quantity, expense.AppliesInventory, expense.Amount, expense.Description,
		expense.ExpenseDate, expense.Vendor, expense.ReferenceNumber, expense.Notes, expense.UpdatedAt,
	)
	if err != nil {
		return nil, NewInternalError(fmt.Errorf("failed to update expense: %w", err))
	}
	if result.RowsAffected() == 0 {
		return nil, fmt.Errorf("expense not found")
	}
	return r.GetByIDWithTx(ctx, tx, expense.ID)
}

// Delete deletes an expense.
func (r *ExpenseRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM expenses WHERE id = $1`, id)
	if err != nil {
		return NewInternalError(fmt.Errorf("failed to delete expense: %w", err))
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("expense not found")
	}
	return nil
}

// DeleteWithTx deletes an expense within a transaction.
func (r *ExpenseRepository) DeleteWithTx(ctx context.Context, tx pgx.Tx, id uuid.UUID) error {
	result, err := tx.Exec(ctx, `DELETE FROM expenses WHERE id = $1`, id)
	if err != nil {
		return NewInternalError(fmt.Errorf("failed to delete expense: %w", err))
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("expense not found")
	}
	return nil
}

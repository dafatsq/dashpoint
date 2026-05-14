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
		INSERT INTO expenses (id, category_id, product_id, quantity, amount, description, expense_date, vendor, reference_number, notes, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $12)
	`

	_, err := r.pool.Exec(ctx, query,
		expense.ID, expense.CategoryID, expense.ProductID, expense.Quantity, expense.Amount, expense.Description,
		expense.ExpenseDate, expense.Vendor, expense.ReferenceNumber, expense.Notes, expense.CreatedBy, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create expense: %w", err)
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
		INSERT INTO expenses (id, category_id, product_id, quantity, amount, description, expense_date, vendor, reference_number, notes, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $12)
	`

	_, err := tx.Exec(ctx, query,
		expense.ID, expense.CategoryID, expense.ProductID, expense.Quantity, expense.Amount, expense.Description,
		expense.ExpenseDate, expense.Vendor, expense.ReferenceNumber, expense.Notes, expense.CreatedBy, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create expense: %w", err)
	}
	return r.GetByIDWithTx(ctx, tx, expense.ID)
}

// Update updates an expense.
func (r *ExpenseRepository) Update(ctx context.Context, expense *models.Expense) (*models.Expense, error) {
	expense.UpdatedAt = time.Now()
	query := `
		UPDATE expenses
		SET category_id = $2, product_id = $3, quantity = $4, amount = $5, description = $6, expense_date = $7,
		    vendor = $8, reference_number = $9, notes = $10, updated_at = $11
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query,
		expense.ID, expense.CategoryID, expense.ProductID, expense.Quantity, expense.Amount, expense.Description,
		expense.ExpenseDate, expense.Vendor, expense.ReferenceNumber, expense.Notes, expense.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update expense: %w", err)
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
		SET category_id = $2, product_id = $3, quantity = $4, amount = $5, description = $6, expense_date = $7,
		    vendor = $8, reference_number = $9, notes = $10, updated_at = $11
		WHERE id = $1
	`

	result, err := tx.Exec(ctx, query,
		expense.ID, expense.CategoryID, expense.ProductID, expense.Quantity, expense.Amount, expense.Description,
		expense.ExpenseDate, expense.Vendor, expense.ReferenceNumber, expense.Notes, expense.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update expense: %w", err)
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
		return fmt.Errorf("failed to delete expense: %w", err)
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
		return fmt.Errorf("failed to delete expense: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("expense not found")
	}
	return nil
}

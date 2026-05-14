package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"dashpoint/backend/internal/models"
)

// GetByIDWithTx retrieves an expense by ID within a transaction.
func (r *ExpenseRepository) GetByIDWithTx(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*models.Expense, error) {
	return scanExpense(tx.QueryRow(ctx, expenseSelectColumns+" WHERE e.id = $1", id))
}

// GetByID retrieves an expense by ID.
func (r *ExpenseRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Expense, error) {
	return scanExpense(r.pool.QueryRow(ctx, expenseSelectColumns+" WHERE e.id = $1", id))
}

// List retrieves expenses with optional filtering.
func (r *ExpenseRepository) List(ctx context.Context, categoryID *uuid.UUID, startDate, endDate *time.Time, limit, offset int) ([]models.Expense, int, error) {
	baseQuery, args, argNum := buildExpenseListBaseQuery(categoryID, startDate, endDate)

	var total int
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) "+baseQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count expenses: %w", err)
	}

	query := expenseSelectColumns + baseQuery + fmt.Sprintf(" ORDER BY e.expense_date DESC, e.created_at DESC LIMIT $%d OFFSET $%d", argNum, argNum+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list expenses: %w", err)
	}
	defer rows.Close()

	expenses := make([]models.Expense, 0)
	for rows.Next() {
		expense, scanErr := scanExpense(rows)
		if scanErr != nil {
			return nil, 0, scanErr
		}
		expenses = append(expenses, *expense)
	}

	return expenses, total, nil
}

// GetSummary gets expense summary for a date range.
func (r *ExpenseRepository) GetSummary(ctx context.Context, startDate, endDate time.Time) (*models.ExpenseSummary, error) {
	totalQuery := `
		SELECT COALESCE(SUM(amount), 0), COUNT(*)
		FROM expenses
		WHERE expense_date >= $1 AND expense_date <= $2
	`

	var summary models.ExpenseSummary
	if err := r.pool.QueryRow(ctx, totalQuery, startDate, endDate).Scan(&summary.TotalAmount, &summary.ExpenseCount); err != nil {
		return nil, fmt.Errorf("failed to get expense summary: %w", err)
	}

	categoryQuery := `
		SELECT e.category_id, COALESCE(ec.name, 'Uncategorized') as category_name,
		       SUM(e.amount) as total_amount, COUNT(*) as count
		FROM expenses e
		LEFT JOIN expense_categories ec ON e.category_id = ec.id
		WHERE e.expense_date >= $1 AND e.expense_date <= $2
		GROUP BY e.category_id, ec.name
		ORDER BY total_amount DESC
	`

	rows, err := r.pool.Query(ctx, categoryQuery, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get category summary: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var categorySummary models.CategoryExpenseSummary
		if err := rows.Scan(&categorySummary.CategoryID, &categorySummary.CategoryName, &categorySummary.TotalAmount, &categorySummary.Count); err != nil {
			return nil, fmt.Errorf("failed to scan category summary: %w", err)
		}
		summary.ByCategory = append(summary.ByCategory, categorySummary)
	}

	return &summary, nil
}

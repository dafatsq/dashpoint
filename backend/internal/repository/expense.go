package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"dashpoint/backend/internal/models"
)

// ExpenseRepository handles expense database operations
type ExpenseRepository struct {
	pool *pgxpool.Pool
}

// NewExpenseRepository creates a new expense repository
func NewExpenseRepository(pool *pgxpool.Pool) *ExpenseRepository {
	return &ExpenseRepository{pool: pool}
}

// --- Expense Categories ---

// ListCategories retrieves all expense categories
func (r *ExpenseRepository) ListCategories(ctx context.Context, status string) ([]models.ExpenseCategory, error) {
	query := `
		SELECT id, name, description, is_active, created_at, updated_at
		FROM expense_categories
	`
	switch status {
	case "active":
		query += ` WHERE is_active = true`
	case "archived":
		query += ` WHERE is_active = false`
	}
	query += ` ORDER BY name`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list expense categories: %w", err)
	}
	defer rows.Close()

	var categories []models.ExpenseCategory
	for rows.Next() {
		var cat models.ExpenseCategory
		err := rows.Scan(
			&cat.ID,
			&cat.Name,
			&cat.Description,
			&cat.IsActive,
			&cat.CreatedAt,
			&cat.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan expense category: %w", err)
		}
		categories = append(categories, cat)
	}

	return categories, nil
}

// CreateCategory creates a new expense category
func (r *ExpenseRepository) CreateCategory(ctx context.Context, name string, description *string) (*models.ExpenseCategory, error) {
	id := uuid.New()
	now := time.Now()

	query := `
		INSERT INTO expense_categories (id, name, description, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, true, $4, $4)
		RETURNING id, name, description, is_active, created_at, updated_at
	`

	var cat models.ExpenseCategory
	err := r.pool.QueryRow(ctx, query, id, name, description, now).Scan(
		&cat.ID,
		&cat.Name,
		&cat.Description,
		&cat.IsActive,
		&cat.CreatedAt,
		&cat.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create expense category: %w", err)
	}

	return &cat, nil
}

// GetCategoryByID retrieves an expense category by ID
func (r *ExpenseRepository) GetCategoryByID(ctx context.Context, id uuid.UUID) (*models.ExpenseCategory, error) {
	query := `
		SELECT id, name, description, is_active, created_at, updated_at
		FROM expense_categories
		WHERE id = $1
	`

	var cat models.ExpenseCategory
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&cat.ID,
		&cat.Name,
		&cat.Description,
		&cat.IsActive,
		&cat.CreatedAt,
		&cat.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get expense category: %w", err)
	}

	return &cat, nil
}

// UpdateCategory updates an expense category
func (r *ExpenseRepository) UpdateCategory(ctx context.Context, cat *models.ExpenseCategory) (*models.ExpenseCategory, error) {
	cat.UpdatedAt = time.Now()

	query := `
		UPDATE expense_categories
		SET name = $2, description = $3, is_active = $4, updated_at = $5
		WHERE id = $1
		RETURNING id, name, description, is_active, created_at, updated_at
	`

	var updated models.ExpenseCategory
	err := r.pool.QueryRow(ctx, query, cat.ID, cat.Name, cat.Description, cat.IsActive, cat.UpdatedAt).Scan(
		&updated.ID,
		&updated.Name,
		&updated.Description,
		&updated.IsActive,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update expense category: %w", err)
	}

	return &updated, nil
}

// DeleteCategory soft-deletes an expense category
func (r *ExpenseRepository) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE expense_categories SET is_active = false, updated_at = $1 WHERE id = $2`
	result, err := r.pool.Exec(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to delete expense category: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("expense category not found")
	}
	return nil
}

// PermanentDeleteCategory permanently deletes an expense category
func (r *ExpenseRepository) PermanentDeleteCategory(ctx context.Context, id uuid.UUID) error {
	// First check if category is in use
	var count int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM expenses WHERE category_id = $1", id).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check category usage: %w", err)
	}

	if count > 0 {
		return fmt.Errorf("cannot delete category as it is currently in use by %d expenses", count)
	}

	result, err := r.pool.Exec(ctx, "DELETE FROM expense_categories WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to permanently delete expense category: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("expense category not found")
	}

	return nil
}

// --- Expenses ---

// Create creates a new expense
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
		expense.ID,
		expense.CategoryID,
		expense.ProductID,
		expense.Quantity,
		expense.Amount,
		expense.Description,
		expense.ExpenseDate,
		expense.Vendor,
		expense.ReferenceNumber,
		expense.Notes,
		expense.CreatedBy,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create expense: %w", err)
	}

	// Fetch with category name
	return r.GetByID(ctx, expense.ID)
}

// BeginTx starts a new database transaction
func (r *ExpenseRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.pool.Begin(ctx)
}

// CreateWithTx creates a new expense within a transaction
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
		expense.ID,
		expense.CategoryID,
		expense.ProductID,
		expense.Quantity,
		expense.Amount,
		expense.Description,
		expense.ExpenseDate,
		expense.Vendor,
		expense.ReferenceNumber,
		expense.Notes,
		expense.CreatedBy,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create expense: %w", err)
	}

	// Fetch with category name
	return r.GetByIDWithTx(ctx, tx, expense.ID)
}

// GetByIDWithTx retrieves an expense by ID within a transaction
func (r *ExpenseRepository) GetByIDWithTx(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*models.Expense, error) {
	query := `
		SELECT e.id, e.category_id, ec.name as category_name, e.product_id, p.name as product_name, 
		       e.quantity, e.amount, e.description, 
		       e.expense_date, e.vendor, e.reference_number, e.notes, 
		       e.created_by, u.name as created_by_name, e.created_at, e.updated_at
		FROM expenses e
		LEFT JOIN expense_categories ec ON e.category_id = ec.id
		LEFT JOIN products p ON e.product_id = p.id
		LEFT JOIN users u ON e.created_by = u.id
		WHERE e.id = $1
	`

	var expense models.Expense
	err := tx.QueryRow(ctx, query, id).Scan(
		&expense.ID,
		&expense.CategoryID,
		&expense.CategoryName,
		&expense.ProductID,
		&expense.ProductName,
		&expense.Quantity,
		&expense.Amount,
		&expense.Description,
		&expense.ExpenseDate,
		&expense.Vendor,
		&expense.ReferenceNumber,
		&expense.Notes,
		&expense.CreatedBy,
		&expense.CreatedByName,
		&expense.CreatedAt,
		&expense.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get expense: %w", err)
	}

	return &expense, nil
}

// GetByID retrieves an expense by ID
func (r *ExpenseRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Expense, error) {
	query := `
		SELECT e.id, e.category_id, ec.name as category_name, e.product_id, p.name as product_name, 
		       e.quantity, e.amount, e.description, 
		       e.expense_date, e.vendor, e.reference_number, e.notes, 
		       e.created_by, u.name as created_by_name, e.created_at, e.updated_at
		FROM expenses e
		LEFT JOIN expense_categories ec ON e.category_id = ec.id
		LEFT JOIN products p ON e.product_id = p.id
		LEFT JOIN users u ON e.created_by = u.id
		WHERE e.id = $1
	`

	var expense models.Expense
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&expense.ID,
		&expense.CategoryID,
		&expense.CategoryName,
		&expense.ProductID,
		&expense.ProductName,
		&expense.Quantity,
		&expense.Amount,
		&expense.Description,
		&expense.ExpenseDate,
		&expense.Vendor,
		&expense.ReferenceNumber,
		&expense.Notes,
		&expense.CreatedBy,
		&expense.CreatedByName,
		&expense.CreatedAt,
		&expense.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get expense: %w", err)
	}

	return &expense, nil
}

// List retrieves expenses with optional filtering
func (r *ExpenseRepository) List(ctx context.Context, categoryID *uuid.UUID, startDate, endDate *time.Time, limit, offset int) ([]models.Expense, int, error) {
	baseQuery := `
		FROM expenses e
		LEFT JOIN expense_categories ec ON e.category_id = ec.id
		LEFT JOIN products p ON e.product_id = p.id
		LEFT JOIN users u ON e.created_by = u.id
		WHERE 1=1
	`
	args := []interface{}{}
	argNum := 1

	if categoryID != nil {
		baseQuery += fmt.Sprintf(" AND e.category_id = $%d", argNum)
		args = append(args, *categoryID)
		argNum++
	}

	if startDate != nil {
		baseQuery += fmt.Sprintf(" AND e.expense_date >= $%d", argNum)
		args = append(args, *startDate)
		argNum++
	}

	if endDate != nil {
		baseQuery += fmt.Sprintf(" AND e.expense_date <= $%d", argNum)
		args = append(args, *endDate)
		argNum++
	}

	// Get total count
	countQuery := "SELECT COUNT(*) " + baseQuery
	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count expenses: %w", err)
	}

	// Get expenses
	selectQuery := `
		SELECT e.id, e.category_id, ec.name as category_name, e.product_id, p.name as product_name, 
		       e.quantity, e.amount, e.description, 
		       e.expense_date, e.vendor, e.reference_number, e.notes, 
		       e.created_by, u.name as created_by_name, e.created_at, e.updated_at
	` + baseQuery + fmt.Sprintf(" ORDER BY e.expense_date DESC, e.created_at DESC LIMIT $%d OFFSET $%d", argNum, argNum+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list expenses: %w", err)
	}
	defer rows.Close()

	var expenses []models.Expense
	for rows.Next() {
		var expense models.Expense
		err := rows.Scan(
			&expense.ID,
			&expense.CategoryID,
			&expense.CategoryName,
			&expense.ProductID,
			&expense.ProductName,
			&expense.Quantity,
			&expense.Amount,
			&expense.Description,
			&expense.ExpenseDate,
			&expense.Vendor,
			&expense.ReferenceNumber,
			&expense.Notes,
			&expense.CreatedBy,
			&expense.CreatedByName,
			&expense.CreatedAt,
			&expense.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan expense: %w", err)
		}
		expenses = append(expenses, expense)
	}

	return expenses, total, nil
}

// Update updates an expense
func (r *ExpenseRepository) Update(ctx context.Context, expense *models.Expense) (*models.Expense, error) {
	expense.UpdatedAt = time.Now()

	query := `
		UPDATE expenses 
		SET category_id = $2, product_id = $3, quantity = $4, amount = $5, description = $6, expense_date = $7, 
		    vendor = $8, reference_number = $9, notes = $10, updated_at = $11
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query,
		expense.ID,
		expense.CategoryID,
		expense.ProductID,
		expense.Quantity,
		expense.Amount,
		expense.Description,
		expense.ExpenseDate,
		expense.Vendor,
		expense.ReferenceNumber,
		expense.Notes,
		expense.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update expense: %w", err)
	}

	if result.RowsAffected() == 0 {
		return nil, fmt.Errorf("expense not found")
	}

	return r.GetByID(ctx, expense.ID)
}

// UpdateWithTx updates an expense within a transaction
func (r *ExpenseRepository) UpdateWithTx(ctx context.Context, tx pgx.Tx, expense *models.Expense) (*models.Expense, error) {
	expense.UpdatedAt = time.Now()

	query := `
		UPDATE expenses 
		SET category_id = $2, product_id = $3, quantity = $4, amount = $5, description = $6, expense_date = $7, 
		    vendor = $8, reference_number = $9, notes = $10, updated_at = $11
		WHERE id = $1
	`

	result, err := tx.Exec(ctx, query,
		expense.ID,
		expense.CategoryID,
		expense.ProductID,
		expense.Quantity,
		expense.Amount,
		expense.Description,
		expense.ExpenseDate,
		expense.Vendor,
		expense.ReferenceNumber,
		expense.Notes,
		expense.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update expense: %w", err)
	}

	if result.RowsAffected() == 0 {
		return nil, fmt.Errorf("expense not found")
	}

	return r.GetByIDWithTx(ctx, tx, expense.ID)
}

// Delete deletes an expense
func (r *ExpenseRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM expenses WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete expense: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("expense not found")
	}

	return nil
}

// DeleteWithTx deletes an expense within a transaction
func (r *ExpenseRepository) DeleteWithTx(ctx context.Context, tx pgx.Tx, id uuid.UUID) error {
	query := `DELETE FROM expenses WHERE id = $1`

	result, err := tx.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete expense: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("expense not found")
	}

	return nil
}

// GetSummary gets expense summary for a date range
func (r *ExpenseRepository) GetSummary(ctx context.Context, startDate, endDate time.Time) (*models.ExpenseSummary, error) {
	// Get total
	totalQuery := `
		SELECT COALESCE(SUM(amount), 0), COUNT(*)
		FROM expenses
		WHERE expense_date >= $1 AND expense_date <= $2
	`

	var summary models.ExpenseSummary
	err := r.pool.QueryRow(ctx, totalQuery, startDate, endDate).Scan(&summary.TotalAmount, &summary.ExpenseCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get expense summary: %w", err)
	}

	// Get by category
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
		var cs models.CategoryExpenseSummary
		err := rows.Scan(&cs.CategoryID, &cs.CategoryName, &cs.TotalAmount, &cs.Count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category summary: %w", err)
		}
		summary.ByCategory = append(summary.ByCategory, cs)
	}

	return &summary, nil
}

// GetMonthlyTotals gets monthly expense totals
func (r *ExpenseRepository) GetMonthlyTotals(ctx context.Context, months int) ([]map[string]interface{}, error) {
	query := `
		SELECT DATE_TRUNC('month', expense_date) as month, SUM(amount) as total
		FROM expenses
		WHERE expense_date >= DATE_TRUNC('month', CURRENT_DATE) - INTERVAL '%d months'
		GROUP BY DATE_TRUNC('month', expense_date)
		ORDER BY month
	`

	rows, err := r.pool.Query(ctx, fmt.Sprintf(query, months-1))
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly totals: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var month time.Time
		var total decimal.Decimal
		err := rows.Scan(&month, &total)
		if err != nil {
			return nil, fmt.Errorf("failed to scan monthly total: %w", err)
		}
		results = append(results, map[string]interface{}{
			"month": month.Format("2006-01"),
			"total": total.String(),
		})
	}

	return results, nil
}

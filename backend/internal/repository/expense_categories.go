package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"dashpoint/backend/internal/models"
)

// ListCategories retrieves all expense categories.
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

	categories := make([]models.ExpenseCategory, 0)
	for rows.Next() {
		var category models.ExpenseCategory
		if err := rows.Scan(&category.ID, &category.Name, &category.Description, &category.IsActive, &category.CreatedAt, &category.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan expense category: %w", err)
		}
		categories = append(categories, category)
	}

	return categories, nil
}

// CreateCategory creates a new expense category.
func (r *ExpenseRepository) CreateCategory(ctx context.Context, name string, description *string) (*models.ExpenseCategory, error) {
	query := `
		INSERT INTO expense_categories (id, name, description, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, true, $4, $4)
		RETURNING id, name, description, is_active, created_at, updated_at
	`

	var category models.ExpenseCategory
	now := time.Now()
	err := r.pool.QueryRow(ctx, query, uuid.New(), name, description, now).Scan(
		&category.ID,
		&category.Name,
		&category.Description,
		&category.IsActive,
		&category.CreatedAt,
		&category.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create expense category: %w", err)
	}
	return &category, nil
}

// GetCategoryByID retrieves an expense category by ID.
func (r *ExpenseRepository) GetCategoryByID(ctx context.Context, id uuid.UUID) (*models.ExpenseCategory, error) {
	query := `
		SELECT id, name, description, is_active, created_at, updated_at
		FROM expense_categories
		WHERE id = $1
	`

	var category models.ExpenseCategory
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&category.ID,
		&category.Name,
		&category.Description,
		&category.IsActive,
		&category.CreatedAt,
		&category.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get expense category: %w", err)
	}
	return &category, nil
}

// UpdateCategory updates an expense category.
func (r *ExpenseRepository) UpdateCategory(ctx context.Context, category *models.ExpenseCategory) (*models.ExpenseCategory, error) {
	category.UpdatedAt = time.Now()
	query := `
		UPDATE expense_categories
		SET name = $2, description = $3, is_active = $4, updated_at = $5
		WHERE id = $1
		RETURNING id, name, description, is_active, created_at, updated_at
	`

	var updated models.ExpenseCategory
	err := r.pool.QueryRow(ctx, query, category.ID, category.Name, category.Description, category.IsActive, category.UpdatedAt).Scan(
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

// DeleteCategory soft-deletes an expense category.
func (r *ExpenseRepository) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	result, err := r.pool.Exec(ctx, `UPDATE expense_categories SET is_active = false, updated_at = $1 WHERE id = $2`, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to delete expense category: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("expense category not found")
	}
	return nil
}

// PermanentDeleteCategory permanently deletes an expense category.
func (r *ExpenseRepository) PermanentDeleteCategory(ctx context.Context, id uuid.UUID) error {
	var count int
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM expenses WHERE category_id = $1", id).Scan(&count); err != nil {
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

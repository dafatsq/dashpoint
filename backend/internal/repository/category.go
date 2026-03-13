package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"dashpoint/backend/internal/models"
)

// CategoryRepository handles category database operations
type CategoryRepository struct {
	pool *pgxpool.Pool
}

// NewCategoryRepository creates a new category repository
func NewCategoryRepository(pool *pgxpool.Pool) *CategoryRepository {
	return &CategoryRepository{pool: pool}
}

// Create creates a new category
func (r *CategoryRepository) Create(ctx context.Context, category *models.Category) error {
	now := time.Now()
	category.ID = uuid.New()
	category.CreatedAt = now
	category.UpdatedAt = now

	query := `
		INSERT INTO categories (id, name, description, parent_id, sort_order, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.pool.Exec(ctx, query,
		category.ID,
		category.Name,
		category.Description,
		category.ParentID,
		category.SortOrder,
		category.IsActive,
		category.CreatedAt,
		category.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create category: %w", err)
	}

	return nil
}

// GetByID retrieves a category by ID
func (r *CategoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Category, error) {
	query := `
		SELECT id, name, description, parent_id, sort_order, is_active, created_at, updated_at
		FROM categories
		WHERE id = $1
	`

	category := &models.Category{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&category.ID,
		&category.Name,
		&category.Description,
		&category.ParentID,
		&category.SortOrder,
		&category.IsActive,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	return category, nil
}

// List retrieves categories based on status
func (r *CategoryRepository) List(ctx context.Context, status string) ([]*models.Category, error) {
	query := `
		SELECT id, name, description, parent_id, sort_order, is_active, created_at, updated_at
		FROM categories
	`
	
	switch status {
	case "active":
		query += ` WHERE is_active = true`
	case "archived":
		query += ` WHERE is_active = false`
	}
	
	query += ` ORDER BY sort_order ASC, name ASC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query categories: %w", err)
	}
	defer rows.Close()

	var categories []*models.Category
	for rows.Next() {
		category := &models.Category{}
		err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.Description,
			&category.ParentID,
			&category.SortOrder,
			&category.IsActive,
			&category.CreatedAt,
			&category.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, category)
	}

	return categories, nil
}

// Update updates a category
func (r *CategoryRepository) Update(ctx context.Context, category *models.Category) error {
	category.UpdatedAt = time.Now()

	query := `
		UPDATE categories
		SET name = $1, description = $2, parent_id = $3, sort_order = $4, is_active = $5, updated_at = $6
		WHERE id = $7
	`

	result, err := r.pool.Exec(ctx, query,
		category.Name,
		category.Description,
		category.ParentID,
		category.SortOrder,
		category.IsActive,
		category.UpdatedAt,
		category.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update category: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("category not found")
	}

	return nil
}

// Delete soft-deletes a category
func (r *CategoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE categories SET is_active = false, updated_at = $1 WHERE id = $2`
	result, err := r.pool.Exec(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("category not found")
	}
	return nil
}

// PermanentDelete permanently deletes a category
func (r *CategoryRepository) PermanentDelete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM categories WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to permanently delete category: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("category not found")
	}
	return nil
}

// GetProductCount returns the number of products in a category
func (r *CategoryRepository) GetProductCount(ctx context.Context, categoryID uuid.UUID) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM products WHERE category_id = $1 AND is_active = true`,
		categoryID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count products: %w", err)
	}
	return count, nil
}

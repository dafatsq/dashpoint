package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"dashpoint/backend/internal/models"
)

func (r *CategoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Category, error) {
	query := `
		SELECT id, name, description, parent_id, sort_order, is_active, created_at, updated_at
		FROM categories
		WHERE id = $1
	`

	category, err := scanCategoryRow(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if isCategoryNotFoundError(err) {
			return nil, fmt.Errorf("category not found")
		}
		return nil, fmt.Errorf("failed to get category: %w", err)
	}
	return category, nil
}

func (r *CategoryRepository) List(ctx context.Context, status string) ([]*models.Category, error) {
	status = normalizeCategoryStatus(status)

	query := `
		SELECT id, name, description, parent_id, sort_order, is_active, created_at, updated_at
		FROM categories
	`
	args := []interface{}{}
	switch status {
	case "active":
		query += ` WHERE is_active = true`
	case "archived":
		query += ` WHERE is_active = false`
	}
	query += ` ORDER BY sort_order ASC, name ASC`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}
	defer rows.Close()

	categories := []*models.Category{}
	for rows.Next() {
		category, scanErr := scanCategory(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("failed to scan category: %w", scanErr)
		}
		categories = append(categories, category)
	}

	return categories, rows.Err()
}

func (r *CategoryRepository) GetProductCount(ctx context.Context, id uuid.UUID) (int, error) {
	var count int
	err := r.pool.QueryRow(
		ctx,
		`SELECT COUNT(*) FROM products WHERE category_id = $1 AND is_active = true`,
		id,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get product count: %w", err)
	}
	return count, nil
}

func (r *CategoryRepository) GetProductCounts(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]int, error) {
	counts := make(map[uuid.UUID]int, len(ids))
	if len(ids) == 0 {
		return counts, nil
	}

	query := `
		SELECT category_id, COUNT(*)::int
		FROM products
		WHERE category_id = ANY($1) AND is_active = true
		GROUP BY category_id
	`

	rows, err := r.pool.Query(ctx, query, collectCategoryIDs(ids))
	if err != nil {
		return nil, fmt.Errorf("failed to get product counts: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var categoryID uuid.UUID
		var count int
		if scanErr := rows.Scan(&categoryID, &count); scanErr != nil {
			return nil, fmt.Errorf("failed to scan product count: %w", scanErr)
		}
		counts[categoryID] = count
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate product counts: %w", err)
	}

	for _, id := range ids {
		if _, ok := counts[id]; !ok {
			counts[id] = 0
		}
	}

	return counts, nil
}

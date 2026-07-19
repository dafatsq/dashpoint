package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"dashpoint/backend/internal/models"
)

const categorySelectColumns = `
	SELECT id, name, description, is_active, created_at, updated_at
	FROM categories
`

func (r *CategoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Category, error) {
	query := categorySelectColumns + ` WHERE id = $1`

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

	query := categorySelectColumns
	switch status {
	case "active":
		query += ` WHERE is_active = true`
	case "archived":
		query += ` WHERE is_active = false`
	}
	query += ` ORDER BY name ASC`

	rows, err := r.pool.Query(ctx, query)
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

// DuplicateSiblingExists checks if an active category with the same name already exists (case-insensitive).
func (r *CategoryRepository) DuplicateSiblingExists(ctx context.Context, name string, excludeID *uuid.UUID) (bool, error) {
	query := `SELECT COUNT(*) FROM categories WHERE name ILIKE $1 AND is_active = true`
	args := []interface{}{name}

	if excludeID != nil {
		query += fmt.Sprintf(` AND id != $%d`, len(args)+1)
		args = append(args, *excludeID)
	}

	var count int
	if err := r.pool.QueryRow(ctx, query, args...).Scan(&count); err != nil {
		return false, fmt.Errorf("failed to check duplicate category: %w", err)
	}
	return count > 0, nil
}

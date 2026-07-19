package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"dashpoint/backend/internal/models"
)

func (r *CategoryRepository) Create(ctx context.Context, category *models.Category) error {
	query := `
		INSERT INTO categories (name, description, is_active)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`

	return r.pool.QueryRow(
		ctx,
		query,
		category.Name,
		category.Description,
		category.IsActive,
	).Scan(&category.ID, &category.CreatedAt, &category.UpdatedAt)
}

func (r *CategoryRepository) Update(ctx context.Context, category *models.Category) error {
	query := `
		UPDATE categories
		SET name = $2, description = $3, is_active = $4, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		category.ID,
		category.Name,
		category.Description,
		category.IsActive,
	).Scan(&category.UpdatedAt)
	if err != nil {
		return fmt.Errorf("category not found")
	}

	return nil
}

func (r *CategoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.pool.Exec(ctx, `
		UPDATE categories
		SET is_active = false, updated_at = NOW()
		WHERE id = $1
	`, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("category not found")
	}

	return nil
}

func (r *CategoryRepository) PermanentDelete(ctx context.Context, id uuid.UUID) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM categories WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("category not found")
	}
	return nil
}

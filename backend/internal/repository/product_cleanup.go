package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// HasSalesHistory checks if a product has any sales history
func (r *ProductRepository) HasSalesHistory(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM sale_items WHERE product_id = $1`, id).Scan(&count); err != nil {
		return false, fmt.Errorf("failed to check sales history: %w", err)
	}
	return count > 0, nil
}

// PermanentDelete permanently deletes a product and its related data
func (r *ProductRepository) PermanentDelete(ctx context.Context, id uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM inventory_items WHERE product_id = $1`, id); err != nil {
		return fmt.Errorf("failed to delete inventory during permanent delete: %w", err)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM stock_adjustments WHERE product_id = $1`, id); err != nil {
		return fmt.Errorf("failed to delete stock adjustments during permanent delete: %w", err)
	}
	result, err := tx.Exec(ctx, `DELETE FROM products WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete product during permanent delete: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("product not found")
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit permanent delete: %w", err)
	}
	return nil
}

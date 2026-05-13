package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// SKUExists checks if a SKU is already in use
func (r *ProductRepository) SKUExists(ctx context.Context, sku string, excludeID *uuid.UUID) (bool, error) {
	query := `SELECT COUNT(*) FROM products WHERE sku = $1`
	args := []interface{}{sku}
	if excludeID != nil {
		query += ` AND id != $2`
		args = append(args, *excludeID)
	}
	var count int
	if err := r.pool.QueryRow(ctx, query, args...).Scan(&count); err != nil {
		return false, fmt.Errorf("failed to check SKU: %w", err)
	}
	return count > 0, nil
}

// BarcodeExists checks if a barcode is already in use
func (r *ProductRepository) BarcodeExists(ctx context.Context, barcode string, excludeID *uuid.UUID) (bool, error) {
	query := `SELECT COUNT(*) FROM products WHERE barcode = $1`
	args := []interface{}{barcode}
	if excludeID != nil {
		query += ` AND id != $2`
		args = append(args, *excludeID)
	}
	var count int
	if err := r.pool.QueryRow(ctx, query, args...).Scan(&count); err != nil {
		return false, fmt.Errorf("failed to check barcode: %w", err)
	}
	return count > 0, nil
}

package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (r *ProductRepository) countMatchingProducts(ctx context.Context, query string, args ...interface{}) (int, error) {
	var count int
	if err := r.pool.QueryRow(ctx, query, args...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// SKUExists checks if a SKU is already in use
func (r *ProductRepository) SKUExists(ctx context.Context, sku string, excludeID *uuid.UUID) (bool, error) {
	query := `SELECT COUNT(*) FROM products WHERE sku = $1`
	args := []interface{}{sku}
	if excludeID != nil {
		query += ` AND id != $2`
		args = append(args, *excludeID)
	}
	count, err := r.countMatchingProducts(ctx, query, args...)
	if err != nil {
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
	count, err := r.countMatchingProducts(ctx, query, args...)
	if err != nil {
		return false, fmt.Errorf("failed to check barcode: %w", err)
	}
	return count > 0, nil
}

// NameExists checks if a product name is already in use (case-insensitive)
func (r *ProductRepository) NameExists(ctx context.Context, name string, excludeID *uuid.UUID) (bool, error) {
	query := `SELECT COUNT(*) FROM products WHERE name ILIKE $1`
	args := []interface{}{name}
	if excludeID != nil {
		query += ` AND id != $2`
		args = append(args, *excludeID)
	}
	count, err := r.countMatchingProducts(ctx, query, args...)
	if err != nil {
		return false, fmt.Errorf("failed to check product name: %w", err)
	}
	return count > 0, nil
}

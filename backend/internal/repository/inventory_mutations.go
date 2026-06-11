package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"dashpoint/backend/internal/models"
)

// AdjustStock adjusts the stock quantity and records the adjustment
func (r *InventoryRepository) AdjustStock(
	ctx context.Context,
	productID uuid.UUID,
	adjustmentType models.AdjustmentType,
	quantityChange decimal.Decimal,
	reason *string,
	referenceType *string,
	referenceID *uuid.UUID,
	adjustedBy uuid.UUID,
) (*models.StockAdjustment, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	adjustment, err := r.AdjustStockWithTx(ctx, tx, productID, adjustmentType, quantityChange, reason, referenceType, referenceID, adjustedBy)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return adjustment, nil
}

// AdjustStockWithTx adjusts the stock quantity and records the adjustment within an existing transaction
func (r *InventoryRepository) AdjustStockWithTx(
	ctx context.Context,
	tx pgx.Tx,
	productID uuid.UUID,
	adjustmentType models.AdjustmentType,
	quantityChange decimal.Decimal,
	reason *string,
	referenceType *string,
	referenceID *uuid.UUID,
	adjustedBy uuid.UUID,
) (*models.StockAdjustment, error) {
	now := time.Now()
	var currentQty decimal.Decimal
	err := tx.QueryRow(ctx, `SELECT quantity FROM inventory_items WHERE product_id = $1 FOR UPDATE`, productID).Scan(&currentQty)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("inventory not found for product")
		}
		return nil, fmt.Errorf("failed to get current quantity: %w", err)
	}

	newQty := currentQty.Add(quantityChange)
	if newQty.LessThan(decimal.Zero) {
		return nil, fmt.Errorf("insufficient stock: available %s, requested %s", currentQty.String(), quantityChange.Abs().String())
	}

	if _, err := tx.Exec(ctx, `UPDATE inventory_items SET quantity = $1, updated_at = $2 WHERE product_id = $3`, newQty, now, productID); err != nil {
		return nil, fmt.Errorf("failed to update inventory: %w", err)
	}

	adjustment := &models.StockAdjustment{
		ID:             uuid.New(),
		ProductID:      productID,
		AdjustmentType: adjustmentType,
		QuantityBefore: currentQty,
		QuantityChange: quantityChange,
		QuantityAfter:  newQty,
		Reason:         reason,
		ReferenceType:  referenceType,
		ReferenceID:    referenceID,
		AdjustedBy:     adjustedBy,
		CreatedAt:      now,
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO stock_adjustments (id, product_id, adjustment_type, quantity_before, quantity_change, quantity_after, reason, reference_type, reference_id, adjusted_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, adjustment.ID, adjustment.ProductID, adjustment.AdjustmentType, adjustment.QuantityBefore, adjustment.QuantityChange, adjustment.QuantityAfter, adjustment.Reason, adjustment.ReferenceType, adjustment.ReferenceID, adjustment.AdjustedBy, adjustment.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to record adjustment: %w", err)
	}

	return adjustment, nil
}

// SetQuantity sets the inventory quantity directly (for stock count)
func (r *InventoryRepository) SetQuantity(ctx context.Context, productID uuid.UUID, newQuantity decimal.Decimal, reason *string, adjustedBy uuid.UUID) (*models.StockAdjustment, error) {
	if newQuantity.LessThan(decimal.Zero) {
		return nil, fmt.Errorf("quantity cannot be negative")
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	now := time.Now()
	currentQty, err := getInventoryQuantityForUpdateTx(ctx, tx, productID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("inventory not found for product")
		}
		return nil, fmt.Errorf("failed to get current quantity: %w", err)
	}
	quantityChange := newQuantity.Sub(currentQty)

	if err := setInventoryQuantityTx(ctx, tx, productID, newQuantity, now); err != nil {
		return nil, fmt.Errorf("failed to update inventory: %w", err)
	}

	adjustment := &models.StockAdjustment{
		ID:             uuid.New(),
		ProductID:      productID,
		AdjustmentType: models.AdjustmentCount,
		QuantityBefore: currentQty,
		QuantityChange: quantityChange,
		QuantityAfter:  newQuantity,
		Reason:         reason,
		AdjustedBy:     adjustedBy,
		CreatedAt:      now,
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO stock_adjustments (id, product_id, adjustment_type, quantity_before, quantity_change, quantity_after, reason, reference_type, reference_id, adjusted_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NULL, NULL, $8, $9)
	`, adjustment.ID, adjustment.ProductID, adjustment.AdjustmentType, adjustment.QuantityBefore, adjustment.QuantityChange, adjustment.QuantityAfter, adjustment.Reason, adjustment.AdjustedBy, adjustment.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to record adjustment: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return adjustment, nil
}

// UpdateThresholds updates the low stock threshold.
func (r *InventoryRepository) UpdateThresholds(ctx context.Context, productID uuid.UUID, lowStockThreshold decimal.Decimal) error {
	result, err := r.pool.Exec(ctx, `
		UPDATE inventory_items 
		SET low_stock_threshold = $1, updated_at = $2
		WHERE product_id = $3
	`, lowStockThreshold, time.Now(), productID)
	if err != nil {
		return fmt.Errorf("failed to update thresholds: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("inventory not found")
	}
	return nil
}

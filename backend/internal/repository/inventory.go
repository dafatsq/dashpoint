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

// InventoryRepository handles inventory database operations
type InventoryRepository struct {
	pool *pgxpool.Pool
}

// NewInventoryRepository creates a new inventory repository
func NewInventoryRepository(pool *pgxpool.Pool) *InventoryRepository {
	return &InventoryRepository{pool: pool}
}

// GetByProductID retrieves inventory for a product
func (r *InventoryRepository) GetByProductID(ctx context.Context, productID uuid.UUID) (*models.InventoryItem, error) {
	query := `
		SELECT product_id, quantity, reserved_quantity, low_stock_threshold, reorder_quantity,
		       last_counted_at, last_restocked_at, created_at, updated_at
		FROM inventory_items
		WHERE product_id = $1
	`

	item := &models.InventoryItem{}
	err := r.pool.QueryRow(ctx, query, productID).Scan(
		&item.ProductID,
		&item.Quantity,
		&item.ReservedQuantity,
		&item.LowStockThreshold,
		&item.ReorderQuantity,
		&item.LastCountedAt,
		&item.LastRestockedAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get inventory: %w", err)
	}

	return item, nil
}

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

	// Get current quantity with lock
	var currentQty decimal.Decimal
	err := tx.QueryRow(ctx,
		`SELECT quantity FROM inventory_items WHERE product_id = $1 FOR UPDATE`,
		productID,
	).Scan(&currentQty)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("inventory not found for product")
		}
		return nil, fmt.Errorf("failed to get current quantity: %w", err)
	}

	// Calculate new quantity
	newQty := currentQty.Add(quantityChange)

	// Check if we need to validate negative stock
	if newQty.LessThan(decimal.Zero) {
		// Check if product allows negative stock
		var allowNegative bool
		err = tx.QueryRow(ctx,
			`SELECT allow_negative_stock FROM products WHERE id = $1`,
			productID,
		).Scan(&allowNegative)
		if err != nil {
			return nil, fmt.Errorf("failed to check product settings: %w", err)
		}
		if !allowNegative {
			return nil, fmt.Errorf("insufficient stock: available %s, requested %s", currentQty.String(), quantityChange.Abs().String())
		}
	}

	// Update inventory
	updateQuery := `
		UPDATE inventory_items 
		SET quantity = $1, updated_at = $2
		WHERE product_id = $3
	`
	_, err = tx.Exec(ctx, updateQuery, newQty, now, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to update inventory: %w", err)
	}

	// Update last_restocked_at for positive adjustments
	if quantityChange.GreaterThan(decimal.Zero) &&
		(adjustmentType == models.AdjustmentPurchase || adjustmentType == models.AdjustmentReturn) {
		_, err = tx.Exec(ctx,
			`UPDATE inventory_items SET last_restocked_at = $1 WHERE product_id = $2`,
			now, productID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to update last_restocked_at: %w", err)
		}
	}

	// Record adjustment
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

	adjQuery := `
		INSERT INTO stock_adjustments (id, product_id, adjustment_type, quantity_before, quantity_change, quantity_after, reason, reference_type, reference_id, adjusted_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err = tx.Exec(ctx, adjQuery,
		adjustment.ID,
		adjustment.ProductID,
		adjustment.AdjustmentType,
		adjustment.QuantityBefore,
		adjustment.QuantityChange,
		adjustment.QuantityAfter,
		adjustment.Reason,
		adjustment.ReferenceType,
		adjustment.ReferenceID,
		adjustment.AdjustedBy,
		adjustment.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to record adjustment: %w", err)
	}

	return adjustment, nil
}

// SetQuantity sets the inventory quantity directly (for stock count)
func (r *InventoryRepository) SetQuantity(
	ctx context.Context,
	productID uuid.UUID,
	newQuantity decimal.Decimal,
	reason *string,
	adjustedBy uuid.UUID,
) (*models.StockAdjustment, error) {
	// Get current quantity
	item, err := r.GetByProductID(ctx, productID)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, fmt.Errorf("inventory not found for product")
	}

	// Calculate the change
	quantityChange := newQuantity.Sub(item.Quantity)

	// Use AdjustStock with the count type
	return r.AdjustStock(ctx, productID, models.AdjustmentCount, quantityChange, reason, nil, nil, adjustedBy)
}

// UpdateThresholds updates the low stock and reorder thresholds
func (r *InventoryRepository) UpdateThresholds(
	ctx context.Context,
	productID uuid.UUID,
	lowStockThreshold, reorderQuantity decimal.Decimal,
) error {
	query := `
		UPDATE inventory_items 
		SET low_stock_threshold = $1, reorder_quantity = $2, updated_at = $3
		WHERE product_id = $4
	`
	result, err := r.pool.Exec(ctx, query, lowStockThreshold, reorderQuantity, time.Now(), productID)
	if err != nil {
		return fmt.Errorf("failed to update thresholds: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("inventory not found")
	}
	return nil
}

// GetLowStockProducts retrieves products that are at or below their low stock threshold
func (r *InventoryRepository) GetLowStockProducts(ctx context.Context) ([]*models.ProductWithInventory, error) {
	query := `
		SELECT p.id, p.sku, p.barcode, p.name, p.category_id, p.price, p.unit, p.is_active,
		       i.quantity, i.reserved_quantity, i.low_stock_threshold
		FROM products p
		JOIN inventory_items i ON p.id = i.product_id
		WHERE p.is_active = true 
		  AND p.track_inventory = true
		  AND i.quantity <= i.low_stock_threshold
		ORDER BY (i.quantity - i.low_stock_threshold) ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query low stock products: %w", err)
	}
	defer rows.Close()

	var products []*models.ProductWithInventory
	for rows.Next() {
		p := &models.Product{}
		var qty, reserved, threshold decimal.Decimal

		err := rows.Scan(
			&p.ID,
			&p.SKU,
			&p.Barcode,
			&p.Name,
			&p.CategoryID,
			&p.Price,
			&p.Unit,
			&p.IsActive,
			&qty,
			&reserved,
			&threshold,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}

		products = append(products, &models.ProductWithInventory{
			Product:           p,
			Quantity:          qty,
			AvailableQuantity: qty.Sub(reserved),
			IsLowStock:        true,
		})
	}

	return products, nil
}

// GetAdjustmentHistory retrieves stock adjustment history for a product
func (r *InventoryRepository) GetAdjustmentHistory(
	ctx context.Context,
	productID uuid.UUID,
	limit, offset int,
) ([]*models.StockAdjustment, int, error) {
	// Count total
	var total int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM stock_adjustments WHERE product_id = $1`,
		productID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count adjustments: %w", err)
	}

	// Get adjustments
	query := `
		SELECT sa.id, sa.product_id, sa.adjustment_type, sa.quantity_before, sa.quantity_change, sa.quantity_after,
		       sa.reason, sa.reference_type, sa.reference_id, sa.adjusted_by, sa.created_at,
		       u.name
		FROM stock_adjustments sa
		JOIN users u ON sa.adjusted_by = u.id
		WHERE sa.product_id = $1
		ORDER BY sa.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, productID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query adjustments: %w", err)
	}
	defer rows.Close()

	var adjustments []*models.StockAdjustment
	for rows.Next() {
		adj := &models.StockAdjustment{AdjustedByUser: &models.User{}}
		err := rows.Scan(
			&adj.ID,
			&adj.ProductID,
			&adj.AdjustmentType,
			&adj.QuantityBefore,
			&adj.QuantityChange,
			&adj.QuantityAfter,
			&adj.Reason,
			&adj.ReferenceType,
			&adj.ReferenceID,
			&adj.AdjustedBy,
			&adj.CreatedAt,
			&adj.AdjustedByUser.Name,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan adjustment: %w", err)
		}
		adjustments = append(adjustments, adj)
	}

	return adjustments, total, nil
}

// RecordLastCounted updates the last counted timestamp
func (r *InventoryRepository) RecordLastCounted(ctx context.Context, productID uuid.UUID) error {
	query := `UPDATE inventory_items SET last_counted_at = $1, updated_at = $1 WHERE product_id = $2`
	_, err := r.pool.Exec(ctx, query, time.Now(), productID)
	return err
}

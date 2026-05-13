package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"dashpoint/backend/internal/models"
)

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
		&item.ProductID, &item.Quantity, &item.ReservedQuantity, &item.LowStockThreshold,
		&item.ReorderQuantity, &item.LastCountedAt, &item.LastRestockedAt, &item.CreatedAt, &item.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get inventory: %w", err)
	}
	return item, nil
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
		if err := rows.Scan(&p.ID, &p.SKU, &p.Barcode, &p.Name, &p.CategoryID, &p.Price, &p.Unit, &p.IsActive, &qty, &reserved, &threshold); err != nil {
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
func (r *InventoryRepository) GetAdjustmentHistory(ctx context.Context, productID uuid.UUID, limit, offset int) ([]*models.StockAdjustment, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM stock_adjustments WHERE product_id = $1`, productID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count adjustments: %w", err)
	}

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
		if err := rows.Scan(
			&adj.ID, &adj.ProductID, &adj.AdjustmentType, &adj.QuantityBefore, &adj.QuantityChange, &adj.QuantityAfter,
			&adj.Reason, &adj.ReferenceType, &adj.ReferenceID, &adj.AdjustedBy, &adj.CreatedAt, &adj.AdjustedByUser.Name,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan adjustment: %w", err)
		}
		adjustments = append(adjustments, adj)
	}

	return adjustments, total, nil
}

package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"dashpoint/backend/internal/models"
)

// GetByProductID retrieves inventory for a product
func (r *InventoryRepository) GetByProductID(ctx context.Context, productID uuid.UUID) (*models.InventoryItem, error) {
	query := `
		SELECT product_id, quantity, low_stock_threshold, updated_at
		FROM inventory_items
		WHERE product_id = $1
	`
	item := &models.InventoryItem{}
	err := r.pool.QueryRow(ctx, query, productID).Scan(
		&item.ProductID, &item.Quantity, &item.LowStockThreshold, &item.UpdatedAt,
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
		SELECT p.id, p.sku, p.barcode, p.name, p.category_id, p.price, p.is_active,
		       i.quantity, i.low_stock_threshold
		FROM products p
		JOIN inventory_items i ON p.id = i.product_id
		WHERE p.is_active = true 
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
		var qty, threshold decimal.Decimal
		if err := rows.Scan(&p.ID, &p.SKU, &p.Barcode, &p.Name, &p.CategoryID, &p.Price, &p.IsActive, &qty, &threshold); err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, &models.ProductWithInventory{
			Product:           p,
			Quantity:          qty,
			AvailableQuantity: qty,
			IsLowStock:        true,
		})
	}
	return products, nil
}

// GetAdjustmentHistory retrieves stock adjustment history for a product
func (r *InventoryRepository) GetAdjustmentHistory(ctx context.Context, productID uuid.UUID, limit, offset int, adjustmentType *models.AdjustmentType, adjustedBy *uuid.UUID, startDate, endDate *time.Time) ([]*models.StockAdjustment, int, error) {
	countArgs := []interface{}{productID}
	countConditions := []string{"product_id = $1"}
	if adjustmentType != nil {
		countArgs = append(countArgs, *adjustmentType)
		countConditions = append(countConditions, fmt.Sprintf("adjustment_type = $%d", len(countArgs)))
	}
	if adjustedBy != nil {
		countArgs = append(countArgs, *adjustedBy)
		countConditions = append(countConditions, fmt.Sprintf("adjusted_by = $%d", len(countArgs)))
	}
	if startDate != nil {
		countArgs = append(countArgs, *startDate)
		countConditions = append(countConditions, fmt.Sprintf("created_at >= $%d", len(countArgs)))
	}
	if endDate != nil {
		countArgs = append(countArgs, *endDate)
		countConditions = append(countConditions, fmt.Sprintf("created_at < $%d", len(countArgs)))
	}
	countQuery := `SELECT COUNT(*) FROM stock_adjustments WHERE ` + strings.Join(countConditions, " AND ")

	var total int
	if err := r.pool.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count adjustments: %w", err)
	}

	query := `
		SELECT sa.id, sa.product_id, sa.adjustment_type, sa.quantity_before, sa.quantity_change, sa.quantity_after,
		       sa.reason, sa.reference_type, sa.reference_id,
		       COALESCE(sa.adjusted_by, '00000000-0000-0000-0000-000000000000'::uuid),
		       sa.created_at, COALESCE(u.name, 'Former user')
		FROM stock_adjustments sa
		LEFT JOIN users u ON sa.adjusted_by = u.id
		WHERE sa.product_id = $1
	`
	queryArgs := []interface{}{productID}
	if adjustmentType != nil {
		queryArgs = append(queryArgs, *adjustmentType)
		query += fmt.Sprintf(` AND sa.adjustment_type = $%d`, len(queryArgs))
	}
	if adjustedBy != nil {
		queryArgs = append(queryArgs, *adjustedBy)
		query += fmt.Sprintf(` AND sa.adjusted_by = $%d`, len(queryArgs))
	}
	if startDate != nil {
		queryArgs = append(queryArgs, *startDate)
		query += fmt.Sprintf(` AND sa.created_at >= $%d`, len(queryArgs))
	}
	if endDate != nil {
		queryArgs = append(queryArgs, *endDate)
		query += fmt.Sprintf(` AND sa.created_at < $%d`, len(queryArgs))
	}
	query += `
		ORDER BY sa.created_at DESC
		LIMIT $` + fmt.Sprint(len(queryArgs)+1) + ` OFFSET $` + fmt.Sprint(len(queryArgs)+2)
	queryArgs = append(queryArgs, limit, offset)

	rows, err := r.pool.Query(ctx, query, queryArgs...)
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

package repository

import (
	"context"

	"github.com/google/uuid"
)

// GetInventoryValuation gets current inventory valuation.
func (r *ReportRepository) GetInventoryValuation(ctx context.Context, categoryID *uuid.UUID, includeItems bool) (*InventoryValuation, error) {
	valuation := &InventoryValuation{}
	whereClause, args := inventoryWhereClause(categoryID)

	query := `
		SELECT
			COUNT(DISTINCT p.id),
			COALESCE(SUM(i.quantity), 0),
			COALESCE(SUM(0), 0) as total_cost_value,
			COALESCE(SUM(i.quantity * p.price), 0)
		FROM products p
		LEFT JOIN inventory_items i ON p.id = i.product_id
		` + whereClause

	if err := r.pool.QueryRow(ctx, query, args...).Scan(
		&valuation.TotalProducts,
		&valuation.TotalQuantity,
		&valuation.TotalCostValue,
		&valuation.TotalRetailValue,
	); err != nil {
		return nil, err
	}

	valuation.PotentialProfit = valuation.TotalRetailValue.Sub(valuation.TotalCostValue)

	if !includeItems {
		return valuation, nil
	}

	itemQuery := `
		SELECT
			p.id,
			p.name,
			p.sku,
			c.name as category_name,
			COALESCE(i.quantity, 0),
			0 as cost,
			p.price,
			0 as cost_value,
			COALESCE(i.quantity * p.price, 0)
		FROM products p
		LEFT JOIN inventory_items i ON p.id = i.product_id
		LEFT JOIN categories c ON p.category_id = c.id
		` + whereClause + `
		ORDER BY i.quantity * p.price DESC
	`

	rows, err := r.pool.Query(ctx, itemQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item InventoryItem
		if err := rows.Scan(
			&item.ProductID,
			&item.ProductName,
			&item.ProductSKU,
			&item.CategoryName,
			&item.Quantity,
			&item.CostPrice,
			&item.SellPrice,
			&item.CostValue,
			&item.RetailValue,
		); err != nil {
			return nil, err
		}
		valuation.Items = append(valuation.Items, item)
	}

	return valuation, nil
}

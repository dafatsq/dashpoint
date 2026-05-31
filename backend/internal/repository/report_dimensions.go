package repository

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

// GetTopSellers gets top selling products.
func (r *ReportRepository) GetTopSellers(ctx context.Context, startDate, endDate time.Time, limit int) ([]TopSellerItem, error) {
	if limit <= 0 {
		limit = 10
	}

	query := `
		SELECT
			si.product_id,
			si.product_name,
			si.product_sku,
			c.name as category_name,
			SUM(si.quantity) as quantity_sold,
			SUM(si.total) as total_revenue,
			SUM(si.total) as total_profit
		FROM sale_items si
		JOIN sales s ON si.sale_id = s.id
		LEFT JOIN products p ON si.product_id = p.id
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE s.created_at >= $1 AND s.created_at < $2 AND s.status = 'completed'
		GROUP BY si.product_id, si.product_name, si.product_sku, c.name
		ORDER BY quantity_sold DESC
		LIMIT $3
	`

	rows, err := r.pool.Query(ctx, query, startOfReportDay(startDate), exclusiveEndDate(endDate), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]TopSellerItem, 0)
	for rows.Next() {
		var item TopSellerItem
		if err := rows.Scan(
			&item.ProductID,
			&item.ProductName,
			&item.ProductSKU,
			&item.CategoryName,
			&item.QuantitySold,
			&item.TotalRevenue,
			&item.TotalProfit,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

// GetEmployeeSalesReport gets sales by employee.
func (r *ReportRepository) GetEmployeeSalesReport(ctx context.Context, startDate, endDate time.Time) ([]map[string]interface{}, error) {
	query := `
		SELECT
			u.id,
			u.name,
			COUNT(s.id) as transaction_count,
			COALESCE(SUM(s.item_count), 0) as item_count,
			COALESCE(SUM(s.total_amount), 0) as total_sales,
			COALESCE(AVG(s.total_amount), 0) as avg_transaction
		FROM users u
		LEFT JOIN sales s ON u.id = s.employee_id
			AND s.created_at >= $1 AND s.created_at < $2
			AND s.status = 'completed'
		WHERE u.is_active = true
		GROUP BY u.id, u.name
		HAVING COUNT(s.id) > 0
		ORDER BY total_sales DESC
	`

	rows, err := r.pool.Query(ctx, query, startOfReportDay(startDate), exclusiveEndDate(endDate))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, name string
		var txCount, itemCount int
		var totalSales, avgTx decimal.Decimal

		if err := rows.Scan(&id, &name, &txCount, &itemCount, &totalSales, &avgTx); err != nil {
			return nil, err
		}

		results = append(results, map[string]interface{}{
			"employee_id":       id,
			"employee_name":     name,
			"transaction_count": txCount,
			"item_count":        itemCount,
			"total_sales":       totalSales.String(),
			"avg_transaction":   avgTx.Round(2).String(),
		})
	}

	return results, nil
}

// GetCategorySalesReport gets sales by category.
func (r *ReportRepository) GetCategorySalesReport(ctx context.Context, startDate, endDate time.Time) ([]map[string]interface{}, error) {
	query := `
		SELECT
			COALESCE(c.id::text, 'uncategorized') as category_id,
			COALESCE(c.name, 'Uncategorized') as category_name,
			COUNT(DISTINCT si.id) as items_sold,
			COALESCE(SUM(si.quantity), 0) as quantity_sold,
			COALESCE(SUM(si.total), 0) as total_revenue
		FROM sale_items si
		JOIN sales s ON si.sale_id = s.id
		LEFT JOIN products p ON si.product_id = p.id
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE s.created_at >= $1 AND s.created_at < $2 AND s.status = 'completed'
		GROUP BY c.id, c.name
		ORDER BY total_revenue DESC
	`

	rows, err := r.pool.Query(ctx, query, startOfReportDay(startDate), exclusiveEndDate(endDate))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]map[string]interface{}, 0)
	for rows.Next() {
		var categoryID, categoryName string
		var itemsSold int
		var qtySold, revenue decimal.Decimal

		if err := rows.Scan(&categoryID, &categoryName, &itemsSold, &qtySold, &revenue); err != nil {
			return nil, err
		}

		results = append(results, map[string]interface{}{
			"category_id":   categoryID,
			"category_name": categoryName,
			"items_sold":    itemsSold,
			"quantity_sold": qtySold.String(),
			"total_revenue": revenue.String(),
		})
	}

	return results, nil
}

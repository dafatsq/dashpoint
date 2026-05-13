package repository

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

// GetDailySummary gets sales summary for a date.
func (r *SaleRepository) GetDailySummary(ctx context.Context, date time.Time) (map[string]interface{}, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	var totalSales, totalTax, totalDiscount, totalAmount decimal.Decimal
	var transactionCount, itemCount int

	err := r.pool.QueryRow(ctx, `
		SELECT
			COALESCE(SUM(subtotal), 0),
			COALESCE(SUM(tax_amount), 0),
			COALESCE(SUM(discount_amount), 0),
			COALESCE(SUM(total_amount), 0),
			COUNT(*),
			COALESCE(SUM(item_count), 0)
		FROM sales
		WHERE created_at >= $1 AND created_at < $2 AND status = 'completed'
	`, startOfDay, endOfDay).Scan(&totalSales, &totalTax, &totalDiscount, &totalAmount, &transactionCount, &itemCount)
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT p.payment_method, COALESCE(SUM(p.amount), 0)
		FROM payments p
		JOIN sales s ON p.sale_id = s.id
		WHERE s.created_at >= $1 AND s.created_at < $2 AND s.status = 'completed'
		GROUP BY p.payment_method
	`, startOfDay, endOfDay)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	paymentBreakdown := make(map[string]string)
	for rows.Next() {
		var method string
		var amount decimal.Decimal
		if err := rows.Scan(&method, &amount); err != nil {
			return nil, err
		}
		paymentBreakdown[method] = amount.String()
	}

	return map[string]interface{}{
		"date":              date.Format("2006-01-02"),
		"total_sales":       totalSales.String(),
		"total_tax":         totalTax.String(),
		"total_discount":    totalDiscount.String(),
		"total_amount":      totalAmount.String(),
		"transaction_count": transactionCount,
		"item_count":        itemCount,
		"payment_breakdown": paymentBreakdown,
	}, nil
}

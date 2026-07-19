package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// GetMonthlyTotals gets monthly expense totals.
func (r *ExpenseRepository) GetMonthlyTotals(ctx context.Context, months int) ([]map[string]interface{}, error) {
	query := `
		SELECT DATE_TRUNC('month', expense_date) as month, SUM(amount) as total
		FROM expenses
		WHERE expense_date >= DATE_TRUNC('month', CURRENT_DATE) - ($1::int - 1) * INTERVAL '1 month'
		GROUP BY DATE_TRUNC('month', expense_date)
		ORDER BY month
	`

	rows, err := r.pool.Query(ctx, query, months)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly totals: %w", err)
	}
	defer rows.Close()

	results := make([]map[string]interface{}, 0)
	for rows.Next() {
		var month time.Time
		var total decimal.Decimal
		if err := rows.Scan(&month, &total); err != nil {
			return nil, fmt.Errorf("failed to scan monthly total: %w", err)
		}
		results = append(results, map[string]interface{}{
			"month": month.Format("2006-01"),
			"total": total.String(),
		})
	}

	return results, nil
}

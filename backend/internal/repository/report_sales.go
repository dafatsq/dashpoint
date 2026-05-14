package repository

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

// GetDailySalesReport gets sales report for a specific date.
func (r *ReportRepository) GetDailySalesReport(ctx context.Context, date time.Time) (*DailySalesReport, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	report := &DailySalesReport{
		Date:             date.Format("2006-01-02"),
		PaymentBreakdown: make(map[string]string),
	}

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
	`, startOfDay, endOfDay).Scan(
		&report.TotalSales,
		&report.TotalTax,
		&report.TotalDiscount,
		&report.TotalAmount,
		&report.TransactionCount,
		&report.ItemCount,
	)
	if err != nil {
		return nil, err
	}

	err = r.pool.QueryRow(ctx, `
		SELECT COUNT(*), COALESCE(SUM(total_amount), 0)
		FROM sales
		WHERE created_at >= $1 AND created_at < $2 AND status = 'voided'
	`, startOfDay, endOfDay).Scan(&report.VoidedCount, &report.VoidedAmount)
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT p.payment_method, COALESCE(SUM(p.amount), 0)
		FROM payments p
		JOIN sales s ON p.sale_id = s.id
		WHERE s.created_at >= $1 AND s.created_at < $2 AND s.status = 'completed'
		GROUP BY p.payment_method
		ORDER BY SUM(p.amount) DESC
	`, startOfDay, endOfDay)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var method string
		var amount decimal.Decimal
		if err := rows.Scan(&method, &amount); err != nil {
			return nil, err
		}
		report.PaymentBreakdown[method] = amount.String()
	}

	hourlyRows, err := r.pool.Query(ctx, `
		SELECT
			EXTRACT(HOUR FROM created_at)::int as hour,
			COALESCE(SUM(total_amount), 0),
			COUNT(*)
		FROM sales
		WHERE created_at >= $1 AND created_at < $2 AND status = 'completed'
		GROUP BY EXTRACT(HOUR FROM created_at)
		ORDER BY hour
	`, startOfDay, endOfDay)
	if err != nil {
		return nil, err
	}
	defer hourlyRows.Close()

	for hourlyRows.Next() {
		var hourly HourlySales
		if err := hourlyRows.Scan(&hourly.Hour, &hourly.Sales, &hourly.Transactions); err != nil {
			return nil, err
		}
		report.HourlySales = append(report.HourlySales, hourly)
	}

	return report, nil
}

// GetSalesRangeReport gets sales report for a date range.
func (r *ReportRepository) GetSalesRangeReport(ctx context.Context, startDate, endDate time.Time) ([]DailySalesReport, error) {
	reports := make([]DailySalesReport, 0)
	for current := startDate; !current.After(endDate); current = current.AddDate(0, 0, 1) {
		report, err := r.GetDailySalesReport(ctx, current)
		if err != nil {
			return nil, err
		}
		reports = append(reports, *report)
	}
	return reports, nil
}

// GetSalesRangeSummary gets aggregated summary for a date range.
func (r *ReportRepository) GetSalesRangeSummary(ctx context.Context, startDate, endDate time.Time) (*SalesRangeSummary, error) {
	summary := &SalesRangeSummary{}
	err := r.pool.QueryRow(ctx, `
		SELECT
			COALESCE(COUNT(*), 0),
			COALESCE(SUM(item_count), 0),
			COALESCE(SUM(total_amount), 0),
			COALESCE(SUM(tax_amount), 0),
			COALESCE(SUM(discount_amount), 0)
		FROM sales
		WHERE created_at >= $1 AND created_at < $2 AND status = 'completed'
	`, startDate, inclusiveEndDate(endDate)).Scan(
		&summary.TotalTransactions,
		&summary.TotalItems,
		&summary.TotalAmount,
		&summary.TotalTax,
		&summary.TotalDiscount,
	)
	return summary, err
}

// GetSalesForExport gets sales data formatted for CSV export.
func (r *ReportRepository) GetSalesForExport(ctx context.Context, startDate, endDate time.Time) ([]SalesReportItem, error) {
	query := `
		SELECT
			s.invoice_no,
			TO_CHAR(s.created_at, 'YYYY-MM-DD') as date,
			TO_CHAR(s.created_at, 'HH24:MI:SS') as time,
			COALESCE(u.name, 'Unknown') as employee_name,
			s.customer_name,
			s.item_count,
			s.subtotal,
			s.tax_amount,
			s.discount_amount,
			s.total_amount,
			COALESCE((
				SELECT STRING_AGG(DISTINCT p.payment_method, ', ')
				FROM payments p WHERE p.sale_id = s.id
			), 'N/A') as payment_methods,
			s.status
		FROM sales s
		LEFT JOIN users u ON s.employee_id = u.id
		WHERE s.created_at >= $1 AND s.created_at < $2
		ORDER BY s.created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, startDate, inclusiveEndDate(endDate))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]SalesReportItem, 0)
	for rows.Next() {
		var item SalesReportItem
		if err := rows.Scan(
			&item.InvoiceNo,
			&item.Date,
			&item.Time,
			&item.EmployeeName,
			&item.CustomerName,
			&item.ItemCount,
			&item.Subtotal,
			&item.Tax,
			&item.Discount,
			&item.Total,
			&item.PaymentMethod,
			&item.Status,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

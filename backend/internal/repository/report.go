package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

// ReportRepository handles report database operations
type ReportRepository struct {
	pool *pgxpool.Pool
}

// NewReportRepository creates a new report repository
func NewReportRepository(pool *pgxpool.Pool) *ReportRepository {
	return &ReportRepository{pool: pool}
}

// DailySalesReport represents daily sales summary
type DailySalesReport struct {
	Date             string            `json:"date"`
	TotalSales       decimal.Decimal   `json:"total_sales"`
	TotalTax         decimal.Decimal   `json:"total_tax"`
	TotalDiscount    decimal.Decimal   `json:"total_discount"`
	TotalAmount      decimal.Decimal   `json:"total_amount"`
	TransactionCount int               `json:"transaction_count"`
	ItemCount        int               `json:"item_count"`
	VoidedCount      int               `json:"voided_count"`
	VoidedAmount     decimal.Decimal   `json:"voided_amount"`
	PaymentBreakdown map[string]string `json:"payment_breakdown"`
	HourlySales      []HourlySales     `json:"hourly_sales,omitempty"`
}

// HourlySales represents sales for a specific hour
type HourlySales struct {
	Hour         int             `json:"hour"`
	Sales        decimal.Decimal `json:"sales"`
	Transactions int             `json:"transactions"`
}

// TopSellerItem represents a top-selling product
type TopSellerItem struct {
	ProductID    string          `json:"product_id"`
	ProductName  string          `json:"product_name"`
	ProductSKU   *string         `json:"product_sku,omitempty"`
	CategoryName *string         `json:"category_name,omitempty"`
	QuantitySold decimal.Decimal `json:"quantity_sold"`
	TotalRevenue decimal.Decimal `json:"total_revenue"`
	TotalProfit  decimal.Decimal `json:"total_profit"`
}

// InventoryValuation represents inventory valuation
type InventoryValuation struct {
	TotalProducts    int             `json:"total_products"`
	TotalQuantity    decimal.Decimal `json:"total_quantity"`
	TotalCostValue   decimal.Decimal `json:"total_cost_value"`
	TotalRetailValue decimal.Decimal `json:"total_retail_value"`
	PotentialProfit  decimal.Decimal `json:"potential_profit"`
	Items            []InventoryItem `json:"items,omitempty"`
}

// InventoryItem represents a single inventory item for valuation
type InventoryItem struct {
	ProductID    string          `json:"product_id"`
	ProductName  string          `json:"product_name"`
	ProductSKU   *string         `json:"product_sku,omitempty"`
	CategoryName *string         `json:"category_name,omitempty"`
	Quantity     decimal.Decimal `json:"quantity"`
	CostPrice    decimal.Decimal `json:"cost_price"`
	SellPrice    decimal.Decimal `json:"sell_price"`
	CostValue    decimal.Decimal `json:"cost_value"`
	RetailValue  decimal.Decimal `json:"retail_value"`
}

// SalesReportItem represents a sale in export format
type SalesReportItem struct {
	InvoiceNo     string          `json:"invoice_no"`
	Date          string          `json:"date"`
	Time          string          `json:"time"`
	EmployeeName  string          `json:"employee_name"`
	CustomerName  *string         `json:"customer_name,omitempty"`
	ItemCount     int             `json:"item_count"`
	Subtotal      decimal.Decimal `json:"subtotal"`
	Tax           decimal.Decimal `json:"tax"`
	Discount      decimal.Decimal `json:"discount"`
	Total         decimal.Decimal `json:"total"`
	PaymentMethod string          `json:"payment_method"`
	Status        string          `json:"status"`
}

// CashReport represents cash flow report
type CashReport struct {
	Date         string          `json:"date"`
	OpeningCash  decimal.Decimal `json:"opening_cash"`
	CashSales    decimal.Decimal `json:"cash_sales"`
	CashRefunds  decimal.Decimal `json:"cash_refunds"`
	PayInTotal   decimal.Decimal `json:"pay_in_total"`
	PayOutTotal  decimal.Decimal `json:"pay_out_total"`
	ExpectedCash decimal.Decimal `json:"expected_cash"`
	ActualCash   decimal.Decimal `json:"actual_cash"`
	Difference   decimal.Decimal `json:"difference"`
	ShiftCount   int             `json:"shift_count"`
}

// SalesRangeSummary represents aggregated summary for a date range
type SalesRangeSummary struct {
	TotalTransactions int             `json:"total_transactions"`
	TotalItems        int             `json:"total_items"`
	TotalAmount       decimal.Decimal `json:"total_amount"`
	TotalTax          decimal.Decimal `json:"total_tax"`
	TotalDiscount     decimal.Decimal `json:"total_discount"`
}

// GetDailySalesReport gets sales report for a specific date
func (r *ReportRepository) GetDailySalesReport(ctx context.Context, date time.Time) (*DailySalesReport, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	report := &DailySalesReport{
		Date:             date.Format("2006-01-02"),
		PaymentBreakdown: make(map[string]string),
	}

	// Get main totals for completed sales
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

	// Get voided sales
	err = r.pool.QueryRow(ctx, `
		SELECT COUNT(*), COALESCE(SUM(total_amount), 0)
		FROM sales 
		WHERE created_at >= $1 AND created_at < $2 AND status = 'voided'
	`, startOfDay, endOfDay).Scan(&report.VoidedCount, &report.VoidedAmount)
	if err != nil {
		return nil, err
	}

	// Get payment breakdown
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

	// Get hourly breakdown
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
		var hs HourlySales
		if err := hourlyRows.Scan(&hs.Hour, &hs.Sales, &hs.Transactions); err != nil {
			return nil, err
		}
		report.HourlySales = append(report.HourlySales, hs)
	}

	return report, nil
}

// GetSalesRangeReport gets sales report for a date range
func (r *ReportRepository) GetSalesRangeReport(ctx context.Context, startDate, endDate time.Time) ([]DailySalesReport, error) {
	var reports []DailySalesReport

	current := startDate
	for !current.After(endDate) {
		report, err := r.GetDailySalesReport(ctx, current)
		if err != nil {
			return nil, err
		}
		reports = append(reports, *report)
		current = current.AddDate(0, 0, 1)
	}

	return reports, nil
}

// GetSalesRangeSummary gets aggregated summary for a date range
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
	`, startDate, endDate.Add(24*time.Hour)).Scan(
		&summary.TotalTransactions,
		&summary.TotalItems,
		&summary.TotalAmount,
		&summary.TotalTax,
		&summary.TotalDiscount,
	)

	return summary, err
}

// GetTopSellers gets top selling products
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
			SUM(si.total) - SUM(si.cost_price * si.quantity) as total_profit
		FROM sale_items si
		JOIN sales s ON si.sale_id = s.id
		LEFT JOIN products p ON si.product_id = p.id
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE s.created_at >= $1 AND s.created_at < $2 AND s.status = 'completed'
		GROUP BY si.product_id, si.product_name, si.product_sku, c.name
		ORDER BY quantity_sold DESC
		LIMIT $3
	`

	rows, err := r.pool.Query(ctx, query, startDate, endDate.Add(24*time.Hour), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []TopSellerItem
	for rows.Next() {
		var item TopSellerItem
		err := rows.Scan(
			&item.ProductID,
			&item.ProductName,
			&item.ProductSKU,
			&item.CategoryName,
			&item.QuantitySold,
			&item.TotalRevenue,
			&item.TotalProfit,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

// GetInventoryValuation gets current inventory valuation
func (r *ReportRepository) GetInventoryValuation(ctx context.Context, categoryID *uuid.UUID, includeItems bool) (*InventoryValuation, error) {
	valuation := &InventoryValuation{}

	whereClause := "WHERE p.is_active = true AND p.track_inventory = true"
	args := []interface{}{}
	argIndex := 1

	if categoryID != nil {
		whereClause += " AND p.category_id = $" + string(rune('0'+argIndex))
		args = append(args, *categoryID)
		argIndex++
	}

	// Get totals
	query := `
		SELECT 
			COUNT(DISTINCT p.id),
			COALESCE(SUM(i.quantity), 0),
			COALESCE(SUM(i.quantity * p.cost), 0),
			COALESCE(SUM(i.quantity * p.price), 0)
		FROM products p
		LEFT JOIN inventory_items i ON p.id = i.product_id
		` + whereClause

	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&valuation.TotalProducts,
		&valuation.TotalQuantity,
		&valuation.TotalCostValue,
		&valuation.TotalRetailValue,
	)
	if err != nil {
		return nil, err
	}

	valuation.PotentialProfit = valuation.TotalRetailValue.Sub(valuation.TotalCostValue)

	// Get individual items if requested
	if includeItems {
		itemQuery := `
			SELECT 
				p.id,
				p.name,
				p.sku,
				c.name as category_name,
				COALESCE(i.quantity, 0),
				p.cost,
				p.price,
				COALESCE(i.quantity * p.cost, 0),
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
			err := rows.Scan(
				&item.ProductID,
				&item.ProductName,
				&item.ProductSKU,
				&item.CategoryName,
				&item.Quantity,
				&item.CostPrice,
				&item.SellPrice,
				&item.CostValue,
				&item.RetailValue,
			)
			if err != nil {
				return nil, err
			}
			valuation.Items = append(valuation.Items, item)
		}
	}

	return valuation, nil
}

// GetSalesForExport gets sales data formatted for CSV export
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

	rows, err := r.pool.Query(ctx, query, startDate, endDate.Add(24*time.Hour))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []SalesReportItem
	for rows.Next() {
		var item SalesReportItem
		err := rows.Scan(
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
		)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

// GetCashReport gets cash report for a date range
func (r *ReportRepository) GetCashReport(ctx context.Context, startDate, endDate time.Time) (*CashReport, error) {
	report := &CashReport{
		Date: fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
	}

	// Get shift data
	err := r.pool.QueryRow(ctx, `
		SELECT 
			COUNT(*),
			COALESCE(SUM(opening_cash), 0),
			COALESCE(SUM(closing_cash), 0)
		FROM shifts
		WHERE started_at >= $1 AND started_at < $2 AND status = 'closed'
	`, startDate, endDate).Scan(&report.ShiftCount, &report.OpeningCash, &report.ActualCash)
	if err != nil {
		return nil, err
	}

	// Get cash sales
	err = r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(p.amount), 0)
		FROM payments p
		JOIN sales s ON p.sale_id = s.id
		WHERE s.created_at >= $1 AND s.created_at < $2 
		AND s.status = 'completed' AND p.payment_method = 'cash'
	`, startDate, endDate).Scan(&report.CashSales)
	if err != nil {
		return nil, err
	}

	// Get cash refunds (voided sales that were paid in cash)
	err = r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(p.amount), 0)
		FROM payments p
		JOIN sales s ON p.sale_id = s.id
		WHERE s.created_at >= $1 AND s.created_at < $2 
		AND s.status = 'voided' AND p.payment_method = 'cash'
	`, startDate, endDate).Scan(&report.CashRefunds)
	if err != nil {
		return nil, err
	}

	// Get pay-in and pay-out totals from cash drawer operations
	// (only for shifts that fall within the date range)
	err = r.pool.QueryRow(ctx, `
		SELECT 
			COALESCE(SUM(CASE WHEN cdo.type = 'pay_in' THEN cdo.amount ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN cdo.type = 'pay_out' THEN cdo.amount ELSE 0 END), 0)
		FROM cash_drawer_operations cdo
		JOIN shifts sh ON cdo.shift_id = sh.id
		WHERE sh.started_at >= $1 AND sh.started_at < $2
	`, startDate, endDate).Scan(&report.PayInTotal, &report.PayOutTotal)
	if err != nil {
		return nil, err
	}

	report.ExpectedCash = report.OpeningCash.
		Add(report.CashSales).
		Sub(report.CashRefunds).
		Add(report.PayInTotal).
		Sub(report.PayOutTotal)
	report.Difference = report.ActualCash.Sub(report.ExpectedCash)

	return report, nil
}


// GetEmployeeSalesReport gets sales by employee
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

	rows, err := r.pool.Query(ctx, query, startDate, endDate.Add(24*time.Hour))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
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

// GetCategorySalesReport gets sales by category
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

	rows, err := r.pool.Query(ctx, query, startDate, endDate.Add(24*time.Hour))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
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

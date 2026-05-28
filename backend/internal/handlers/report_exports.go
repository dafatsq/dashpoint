package handlers

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/shopspring/decimal"
)

// ExportSalesCSV handles GET /api/v1/reports/export/sales
func (h *ReportHandler) ExportSalesCSV(c *fiber.Ctx) error {
	dateRange, err := parseReportRangeResponse(c, 30, false, 0)
	if err != nil {
		return err
	}

	items, err := h.reportRepo.GetSalesForExport(c.Context(), dateRange.start, dateRange.end)
	if err != nil {
		return reportExportInternalError(c, err, "Failed to get sales for export")
	}

	summary, err := h.reportRepo.GetSalesRangeSummary(c.Context(), dateRange.start, dateRange.end)
	if err != nil {
		return reportExportInternalError(c, err, "Failed to get sales summary")
	}

	rows := [][]string{
		{"SALES REPORT SUMMARY"},
		{"Period:", fmt.Sprintf("%s to %s", dateRange.start.Format(reportDateLayout), dateRange.end.Format(reportDateLayout))},
		{""},
		{"Total Transactions:", fmt.Sprintf("%d", summary.TotalTransactions)},
		{"Total Items Sold:", fmt.Sprintf("%d", summary.TotalItems)},
		{"Total Revenue:", summary.TotalAmount.String()},
		{"Total Tax Collected:", summary.TotalTax.String()},
		{"Total Discounts:", summary.TotalDiscount.String()},
		{""},
		{""},
		{"Invoice No", "Date", "Time", "Employee", "Items", "Subtotal", "Tax", "Discount", "Total", "Payment Method", "Status"},
	}

	for _, item := range items {
		rows = append(rows, []string{
			item.InvoiceNo,
			item.Date,
			item.Time,
			item.EmployeeName,
			fmt.Sprintf("%d", item.ItemCount),
			item.Subtotal.String(),
			item.Tax.String(),
			item.Discount.String(),
			item.Total.String(),
			item.PaymentMethod,
			item.Status,
		})
	}

	data, err := writeCSV(rows)
	if err != nil {
		return reportExportInternalError(c, err, "Failed to write sales export")
	}

	filename := fmt.Sprintf("sales_%s_to_%s.csv", dateRange.start.Format("20060102"), dateRange.end.Format("20060102"))
	return sendCSV(c, filename, data)
}

// ExportInventoryCSV handles GET /api/v1/reports/export/inventory
func (h *ReportHandler) ExportInventoryCSV(c *fiber.Ctx) error {
	valuation, err := h.reportRepo.GetInventoryValuation(c.Context(), nil, true)
	if err != nil {
		return reportExportInternalError(c, err, "Failed to get inventory for export")
	}

	rows := [][]string{{
		"Product ID", "Product Name", "SKU", "Category", "Quantity", "Sell Price", "Retail Value",
	}}

	for _, item := range valuation.Items {
		sku := ""
		if item.ProductSKU != nil {
			sku = *item.ProductSKU
		}
		category := ""
		if item.CategoryName != nil {
			category = *item.CategoryName
		}
		rows = append(rows, []string{
			item.ProductID,
			item.ProductName,
			sku,
			category,
			item.Quantity.String(),
			item.SellPrice.String(),
			item.RetailValue.String(),
		})
	}

	rows = append(rows, []string{})
	rows = append(rows, []string{
		"TOTAL", "", "", "", valuation.TotalQuantity.String(), "", valuation.TotalRetailValue.String(),
	})

	data, err := writeCSV(rows)
	if err != nil {
		return reportExportInternalError(c, err, "Failed to write inventory export")
	}

	filename := fmt.Sprintf("inventory_%s.csv", time.Now().Format("20060102"))
	return sendCSV(c, filename, data)
}

// ExportTopSellersCSV handles GET /api/v1/reports/export/top-sellers
func (h *ReportHandler) ExportTopSellersCSV(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 50)
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	dateRange, err := parseReportRangeResponse(c, 30, false, 0)
	if err != nil {
		return err
	}

	items, err := h.reportRepo.GetTopSellers(c.Context(), dateRange.start, dateRange.end, limit)
	if err != nil {
		return reportExportInternalError(c, err, "Failed to get top sellers for export")
	}

	rows := [][]string{{
		"Rank", "Product ID", "Product Name", "SKU", "Category", "Quantity Sold", "Total Revenue",
	}}

	for i, item := range items {
		sku := ""
		if item.ProductSKU != nil {
			sku = *item.ProductSKU
		}
		category := ""
		if item.CategoryName != nil {
			category = *item.CategoryName
		}
		rows = append(rows, []string{
			fmt.Sprintf("%d", i+1),
			item.ProductID,
			item.ProductName,
			sku,
			category,
			item.QuantitySold.String(),
			item.TotalRevenue.String(),
		})
	}

	data, err := writeCSV(rows)
	if err != nil {
		return reportExportInternalError(c, err, "Failed to write top sellers export")
	}

	filename := fmt.Sprintf("top_sellers_%s_to_%s.csv", dateRange.start.Format("20060102"), dateRange.end.Format("20060102"))
	return sendCSV(c, filename, data)
}

// ExportComprehensiveReportCSV exports all analytics and statistics in one CSV
func (h *ReportHandler) ExportComprehensiveReportCSV(c *fiber.Ctx) error {
	dateRange, err := parseReportRangeResponse(c, 30, false, 0)
	if err != nil {
		return err
	}

	summary, err := h.reportRepo.GetSalesRangeSummary(c.Context(), dateRange.start, dateRange.end)
	if err != nil {
		return reportExportInternalError(c, err, "Failed to get sales summary")
	}
	dailyReports, err := h.reportRepo.GetSalesRangeReport(c.Context(), dateRange.start, dateRange.end)
	if err != nil {
		return reportExportInternalError(c, err, "Failed to get daily reports")
	}
	topSellers, err := h.reportRepo.GetTopSellers(c.Context(), dateRange.start, dateRange.end, 20)
	if err != nil {
		return reportExportInternalError(c, err, "Failed to get top sellers")
	}
	employeeSales, err := h.reportRepo.GetEmployeeSalesReport(c.Context(), dateRange.start, dateRange.end)
	if err != nil {
		return reportExportInternalError(c, err, "Failed to get employee sales")
	}
	categorySales, err := h.reportRepo.GetCategorySalesReport(c.Context(), dateRange.start, dateRange.end)
	if err != nil {
		return reportExportInternalError(c, err, "Failed to get category sales")
	}

	rows := [][]string{
		{"COMPREHENSIVE SALES REPORT"},
		{"Generated:", time.Now().Format("2006-01-02 15:04:05")},
		{"Period:", fmt.Sprintf("%s to %s", dateRange.start.Format(reportDateLayout), dateRange.end.Format(reportDateLayout))},
		{""},
		{"=== SUMMARY STATISTICS ==="},
		{"Total Transactions:", fmt.Sprintf("%d", summary.TotalTransactions)},
		{"Total Items Sold:", fmt.Sprintf("%d", summary.TotalItems)},
		{"Total Revenue:", summary.TotalAmount.String()},
		{"Total Tax:", summary.TotalTax.String()},
		{"Total Discounts:", summary.TotalDiscount.String()},
	}

	avgPerTransaction := decimal.Zero
	if summary.TotalTransactions > 0 {
		avgPerTransaction = summary.TotalAmount.Div(decimal.NewFromInt(int64(summary.TotalTransactions)))
	}
	rows = append(rows, []string{"Average per Transaction:", avgPerTransaction.String()})
	rows = append(rows, []string{}, []string{})

	rows = append(rows, []string{"=== DAILY SALES BREAKDOWN ==="})
	rows = append(rows, []string{"Date", "Transactions", "Items Sold", "Revenue", "Tax", "Discounts"})
	for _, day := range dailyReports {
		rows = append(rows, []string{
			day.Date,
			fmt.Sprintf("%d", day.TransactionCount),
			fmt.Sprintf("%d", day.ItemCount),
			day.TotalAmount.String(),
			day.TotalTax.String(),
			day.TotalDiscount.String(),
		})
	}
	rows = append(rows, []string{}, []string{})

	rows = append(rows, []string{"=== TOP 20 SELLING PRODUCTS ==="})
	rows = append(rows, []string{"Rank", "Product Name", "SKU", "Category", "Qty Sold", "Revenue"})
	for i, item := range topSellers {
		sku := ""
		if item.ProductSKU != nil {
			sku = *item.ProductSKU
		}
		category := ""
		if item.CategoryName != nil {
			category = *item.CategoryName
		}
		rows = append(rows, []string{
			fmt.Sprintf("%d", i+1),
			item.ProductName,
			sku,
			category,
			item.QuantitySold.String(),
			item.TotalRevenue.String(),
		})
	}
	rows = append(rows, []string{}, []string{})

	if len(employeeSales) > 0 {
		rows = append(rows, []string{"=== EMPLOYEE SALES PERFORMANCE ==="})
		rows = append(rows, []string{"Employee", "Transactions", "Items Sold", "Total Sales", "Avg per Transaction"})
		for _, emp := range employeeSales {
			rows = append(rows, []string{
				fmt.Sprintf("%v", emp["employee_name"]),
				fmt.Sprintf("%v", emp["transaction_count"]),
				fmt.Sprintf("%v", emp["item_count"]),
				fmt.Sprintf("%v", emp["total_sales"]),
				fmt.Sprintf("%v", emp["avg_transaction"]),
			})
		}
		rows = append(rows, []string{}, []string{})
	}

	if len(categorySales) > 0 {
		rows = append(rows, []string{"=== SALES BY CATEGORY ==="})
		rows = append(rows, []string{"Category", "Items Sold (Line Items)", "Total Quantity", "Revenue"})
		for _, cat := range categorySales {
			rows = append(rows, []string{
				fmt.Sprintf("%v", cat["category_name"]),
				fmt.Sprintf("%v", cat["items_sold"]),
				fmt.Sprintf("%v", cat["quantity_sold"]),
				fmt.Sprintf("%v", cat["total_revenue"]),
			})
		}
	}

	data, err := writeCSV(rows)
	if err != nil {
		return reportExportInternalError(c, err, "Failed to write comprehensive export")
	}

	filename := fmt.Sprintf("comprehensive_report_%s_to_%s.csv", dateRange.start.Format("20060102"), dateRange.end.Format("20060102"))
	return sendCSV(c, filename, data)
}

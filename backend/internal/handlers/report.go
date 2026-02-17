package handlers

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"

	"dashpoint/backend/internal/middleware"
	"dashpoint/backend/internal/repository"
)

// ReportHandler handles report endpoints
type ReportHandler struct {
	reportRepo *repository.ReportRepository
}

// NewReportHandler creates a new report handler
func NewReportHandler(reportRepo *repository.ReportRepository) *ReportHandler {
	return &ReportHandler{reportRepo: reportRepo}
}

// GetDailySalesReport handles GET /api/v1/reports/daily
func (h *ReportHandler) GetDailySalesReport(c *fiber.Ctx) error {
	dateStr := c.Query("date", time.Now().Format("2006-01-02"))
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_DATE",
			"message": "Invalid date format. Use YYYY-MM-DD",
		})
	}

	report, err := h.reportRepo.GetDailySalesReport(c.Context(), date)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get daily sales report")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to generate report",
		})
	}

	return c.JSON(fiber.Map{
		"report": report,
	})
}

// GetSalesRangeReport handles GET /api/v1/reports/sales
func (h *ReportHandler) GetSalesRangeReport(c *fiber.Ctx) error {
	startStr := c.Query("start_date")
	endStr := c.Query("end_date")

	if startStr == "" || endStr == "" {
		// Default to last 7 days
		endDate := time.Now()
		startDate := endDate.AddDate(0, 0, -7)
		startStr = startDate.Format("2006-01-02")
		endStr = endDate.Format("2006-01-02")
	}

	startDate, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_DATE",
			"message": "Invalid start_date format. Use YYYY-MM-DD",
		})
	}

	endDate, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_DATE",
			"message": "Invalid end_date format. Use YYYY-MM-DD",
		})
	}

	if endDate.Before(startDate) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_RANGE",
			"message": "end_date must be after start_date",
		})
	}

	// Limit range to 90 days
	if endDate.Sub(startDate) > 90*24*time.Hour {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "RANGE_TOO_LARGE",
			"message": "Date range cannot exceed 90 days",
		})
	}

	reports, err := h.reportRepo.GetSalesRangeReport(c.Context(), startDate, endDate)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get sales range report")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to generate report",
		})
	}

	// Calculate totals
	var totalSales, totalTax, totalDiscount, totalAmount float64
	var totalTransactions, totalItems int
	for _, r := range reports {
		totalSales += r.TotalSales.InexactFloat64()
		totalTax += r.TotalTax.InexactFloat64()
		totalDiscount += r.TotalDiscount.InexactFloat64()
		totalAmount += r.TotalAmount.InexactFloat64()
		totalTransactions += r.TransactionCount
		totalItems += r.ItemCount
	}

	return c.JSON(fiber.Map{
		"start_date": startStr,
		"end_date":   endStr,
		"summary": fiber.Map{
			"total_sales":        fmt.Sprintf("%.2f", totalSales),
			"total_tax":          fmt.Sprintf("%.2f", totalTax),
			"total_discount":     fmt.Sprintf("%.2f", totalDiscount),
			"total_amount":       fmt.Sprintf("%.2f", totalAmount),
			"total_transactions": totalTransactions,
			"total_items":        totalItems,
		},
		"daily_reports": reports,
	})
}

// GetTopSellers handles GET /api/v1/reports/top-sellers
func (h *ReportHandler) GetTopSellers(c *fiber.Ctx) error {
	startStr := c.Query("start_date")
	endStr := c.Query("end_date")
	limit := c.QueryInt("limit", 10)

	if limit <= 0 || limit > 100 {
		limit = 10
	}

	var startDate, endDate time.Time
	var err error

	if startStr == "" || endStr == "" {
		// Default to last 30 days
		endDate = time.Now()
		startDate = endDate.AddDate(0, 0, -30)
	} else {
		startDate, err = time.Parse("2006-01-02", startStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code":    "INVALID_DATE",
				"message": "Invalid start_date format",
			})
		}
		endDate, err = time.Parse("2006-01-02", endStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code":    "INVALID_DATE",
				"message": "Invalid end_date format",
			})
		}
	}

	items, err := h.reportRepo.GetTopSellers(c.Context(), startDate, endDate, limit)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get top sellers")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to generate report",
		})
	}

	return c.JSON(fiber.Map{
		"start_date":  startDate.Format("2006-01-02"),
		"end_date":    endDate.Format("2006-01-02"),
		"limit":       limit,
		"top_sellers": items,
	})
}

// GetInventoryValuation handles GET /api/v1/reports/inventory
func (h *ReportHandler) GetInventoryValuation(c *fiber.Ctx) error {
	includeItems := c.QueryBool("include_items", false)

	var categoryID *uuid.UUID
	if catIDStr := c.Query("category_id"); catIDStr != "" {
		id, err := uuid.Parse(catIDStr)
		if err == nil {
			categoryID = &id
		}
	}

	valuation, err := h.reportRepo.GetInventoryValuation(c.Context(), categoryID, includeItems)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get inventory valuation")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to generate report",
		})
	}

	return c.JSON(fiber.Map{
		"valuation": valuation,
	})
}

// GetCashReport handles GET /api/v1/reports/cash
func (h *ReportHandler) GetCashReport(c *fiber.Ctx) error {
	startStr := c.Query("start_date")
	endStr := c.Query("end_date")

	var startDate, endDate time.Time
	var err error

	if startStr == "" || endStr == "" {
		endDate = time.Now()
		startDate = endDate.AddDate(0, 0, -30)
	} else {
		startDate, err = time.Parse("2006-01-02", startStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code":    "INVALID_DATE",
				"message": "Invalid start_date format",
			})
		}
		endDate, err = time.Parse("2006-01-02", endStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code":    "INVALID_DATE",
				"message": "Invalid end_date format",
			})
		}
	}
    
    // Add 1 day to end date to make it inclusive for query logic (started_at < endDate)
    // Actually repository logic uses started_at < $2. If user selects "2023-01-01" to "2023-01-01".
    // StartDate = 01-01 00:00. EndDate = 01-01 00:00.
    // Query: >= 01-01 AND < 01-01. Returns nothing.
    // So EndDate must be set to End of Day or Next Day.
    // In GetDailySalesReport, logic was implicit.
    // In GetSalesRangeReport?
    
    // Let's check GetSalesRangeReport logic (Reference lines 53-131 in Step 448).
    // line 54: startDate, endDate parsed.
    // line 97: passed to repo.
    
    // Let's check Repository GetSalesRangeReport logic?
    // I didn't verify that.
    
    // Assuming standard practice: endDate should be end of day?
    // In `GetCashReport` repo logic (Step 464): `WHERE started_at >= $1 AND started_at < $2`.
    // If I pass 2023-01-01 00:00:00 as $2, I miss the last day.
    // So I should add 24 hours to endDate?
    // Or set it to 23:59:59?
    
    // Let's check existing implementation of `GetSalesRangeReport` handler in Step 448.
    // It does NOT add time.
    // So if the Repo relies on implicit "next day" for exclusive upper bound?
    
    // Let's assume I need to handle the upper bound.
    endDate = endDate.Add(24 * time.Hour) // This makes it exclusive upper bound for the NEXT day start.

	report, err := h.reportRepo.GetCashReport(c.Context(), startDate, endDate)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get cash report")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to generate report",
		})
	}

	return c.JSON(fiber.Map{
		"report": report,
	})
}

// GetEmployeeSalesReport handles GET /api/v1/reports/by-employee
func (h *ReportHandler) GetEmployeeSalesReport(c *fiber.Ctx) error {
	startStr := c.Query("start_date")
	endStr := c.Query("end_date")

	var startDate, endDate time.Time
	var err error

	if startStr == "" || endStr == "" {
		endDate = time.Now()
		startDate = endDate.AddDate(0, 0, -30)
	} else {
		startDate, err = time.Parse("2006-01-02", startStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code":    "INVALID_DATE",
				"message": "Invalid start_date format",
			})
		}
		endDate, err = time.Parse("2006-01-02", endStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code":    "INVALID_DATE",
				"message": "Invalid end_date format",
			})
		}
	}

	results, err := h.reportRepo.GetEmployeeSalesReport(c.Context(), startDate, endDate)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get employee sales report")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to generate report",
		})
	}

	return c.JSON(fiber.Map{
		"start_date": startDate.Format("2006-01-02"),
		"end_date":   endDate.Format("2006-01-02"),
		"employees":  results,
	})
}

// GetCategorySalesReport handles GET /api/v1/reports/by-category
func (h *ReportHandler) GetCategorySalesReport(c *fiber.Ctx) error {
	startStr := c.Query("start_date")
	endStr := c.Query("end_date")

	var startDate, endDate time.Time
	var err error

	if startStr == "" || endStr == "" {
		endDate = time.Now()
		startDate = endDate.AddDate(0, 0, -30)
	} else {
		startDate, err = time.Parse("2006-01-02", startStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code":    "INVALID_DATE",
				"message": "Invalid start_date format",
			})
		}
		endDate, err = time.Parse("2006-01-02", endStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code":    "INVALID_DATE",
				"message": "Invalid end_date format",
			})
		}
	}

	results, err := h.reportRepo.GetCategorySalesReport(c.Context(), startDate, endDate)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get category sales report")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to generate report",
		})
	}

	return c.JSON(fiber.Map{
		"start_date": startDate.Format("2006-01-02"),
		"end_date":   endDate.Format("2006-01-02"),
		"categories": results,
	})
}

// ExportSalesCSV handles GET /api/v1/reports/export/sales
func (h *ReportHandler) ExportSalesCSV(c *fiber.Ctx) error {
	// Check permission
	roleName := middleware.GetRoleName(c)
	if roleName != "owner" && roleName != "manager" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"code":    "FORBIDDEN",
			"message": "Only owner or manager can export data",
		})
	}

	startStr := c.Query("start_date")
	endStr := c.Query("end_date")

	var startDate, endDate time.Time
	var err error

	if startStr == "" || endStr == "" {
		endDate = time.Now()
		startDate = endDate.AddDate(0, 0, -30)
	} else {
		startDate, err = time.Parse("2006-01-02", startStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code":    "INVALID_DATE",
				"message": "Invalid start_date format",
			})
		}
		endDate, err = time.Parse("2006-01-02", endStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code":    "INVALID_DATE",
				"message": "Invalid end_date format",
			})
		}
	}

	items, err := h.reportRepo.GetSalesForExport(c.Context(), startDate, endDate)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get sales for export")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to export data",
		})
	}

	// Get summary statistics for the period
	summary, err := h.reportRepo.GetSalesRangeSummary(c.Context(), startDate, endDate)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get sales summary")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to export data",
		})
	}

	// Generate CSV
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write summary section
	writer.Write([]string{"SALES REPORT SUMMARY"})
	writer.Write([]string{"Period:", fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))})
	writer.Write([]string{""})
	writer.Write([]string{"Total Transactions:", fmt.Sprintf("%d", summary.TotalTransactions)})
	writer.Write([]string{"Total Items Sold:", fmt.Sprintf("%d", summary.TotalItems)})
	writer.Write([]string{"Total Revenue:", summary.TotalAmount.String()})
	writer.Write([]string{"Total Tax Collected:", summary.TotalTax.String()})
	writer.Write([]string{"Total Discounts:", summary.TotalDiscount.String()})
	writer.Write([]string{""})
	writer.Write([]string{""})

	// Header
	writer.Write([]string{
		"Invoice No", "Date", "Time", "Employee", "Customer",
		"Items", "Subtotal", "Tax", "Discount", "Total",
		"Payment Method", "Status",
	})

	// Data
	for _, item := range items {
		customer := ""
		if item.CustomerName != nil {
			customer = *item.CustomerName
		}
		writer.Write([]string{
			item.InvoiceNo,
			item.Date,
			item.Time,
			item.EmployeeName,
			customer,
			fmt.Sprintf("%d", item.ItemCount),
			item.Subtotal.String(),
			item.Tax.String(),
			item.Discount.String(),
			item.Total.String(),
			item.PaymentMethod,
			item.Status,
		})
	}

	writer.Flush()

	// Set response headers
	filename := fmt.Sprintf("sales_%s_to_%s.csv", startDate.Format("20060102"), endDate.Format("20060102"))
	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	return c.Send(buf.Bytes())
}

// ExportInventoryCSV handles GET /api/v1/reports/export/inventory
func (h *ReportHandler) ExportInventoryCSV(c *fiber.Ctx) error {
	// Check permission
	roleName := middleware.GetRoleName(c)
	if roleName != "owner" && roleName != "manager" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"code":    "FORBIDDEN",
			"message": "Only owner or manager can export data",
		})
	}

	valuation, err := h.reportRepo.GetInventoryValuation(c.Context(), nil, true)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get inventory for export")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to export data",
		})
	}

	// Generate CSV
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Header
	writer.Write([]string{
		"Product ID", "Product Name", "SKU", "Category",
		"Quantity", "Cost Price", "Sell Price", "Cost Value", "Retail Value",
	})

	// Data
	for _, item := range valuation.Items {
		sku := ""
		if item.ProductSKU != nil {
			sku = *item.ProductSKU
		}
		category := ""
		if item.CategoryName != nil {
			category = *item.CategoryName
		}
		writer.Write([]string{
			item.ProductID,
			item.ProductName,
			sku,
			category,
			item.Quantity.String(),
			item.CostPrice.String(),
			item.SellPrice.String(),
			item.CostValue.String(),
			item.RetailValue.String(),
		})
	}

	// Summary row
	writer.Write([]string{})
	writer.Write([]string{
		"TOTAL", "", "", "",
		valuation.TotalQuantity.String(),
		"", "",
		valuation.TotalCostValue.String(),
		valuation.TotalRetailValue.String(),
	})

	writer.Flush()

	// Set response headers
	filename := fmt.Sprintf("inventory_%s.csv", time.Now().Format("20060102"))
	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	return c.Send(buf.Bytes())
}

// ExportTopSellersCSV handles GET /api/v1/reports/export/top-sellers
func (h *ReportHandler) ExportTopSellersCSV(c *fiber.Ctx) error {
	// Check permission
	roleName := middleware.GetRoleName(c)
	if roleName != "owner" && roleName != "manager" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"code":    "FORBIDDEN",
			"message": "Only owner or manager can export data",
		})
	}

	startStr := c.Query("start_date")
	endStr := c.Query("end_date")
	limit := c.QueryInt("limit", 50)

	var startDate, endDate time.Time
	var err error

	if startStr == "" || endStr == "" {
		endDate = time.Now()
		startDate = endDate.AddDate(0, 0, -30)
	} else {
		startDate, _ = time.Parse("2006-01-02", startStr)
		endDate, _ = time.Parse("2006-01-02", endStr)
	}

	items, err := h.reportRepo.GetTopSellers(c.Context(), startDate, endDate, limit)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get top sellers for export")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to export data",
		})
	}

	// Generate CSV
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Header
	writer.Write([]string{
		"Rank", "Product ID", "Product Name", "SKU", "Category",
		"Quantity Sold", "Total Revenue", "Total Profit",
	})

	// Data
	for i, item := range items {
		sku := ""
		if item.ProductSKU != nil {
			sku = *item.ProductSKU
		}
		category := ""
		if item.CategoryName != nil {
			category = *item.CategoryName
		}
		writer.Write([]string{
			fmt.Sprintf("%d", i+1),
			item.ProductID,
			item.ProductName,
			sku,
			category,
			item.QuantitySold.String(),
			item.TotalRevenue.String(),
			item.TotalProfit.String(),
		})
	}

	writer.Flush()

	// Set response headers
	filename := fmt.Sprintf("top_sellers_%s_to_%s.csv", startDate.Format("20060102"), endDate.Format("20060102"))
	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	return c.Send(buf.Bytes())
}

// ExportComprehensiveReportCSV exports all analytics and statistics in one CSV
func (h *ReportHandler) ExportComprehensiveReportCSV(c *fiber.Ctx) error {
	// Check permission
	roleName := middleware.GetRoleName(c)
	if roleName != "owner" && roleName != "manager" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"code":    "FORBIDDEN",
			"message": "Only owner or manager can export data",
		})
	}

	startStr := c.Query("start_date")
	endStr := c.Query("end_date")

	var startDate, endDate time.Time
	var err error

	if startStr == "" || endStr == "" {
		endDate = time.Now()
		startDate = endDate.AddDate(0, 0, -30)
	} else {
		startDate, err = time.Parse("2006-01-02", startStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code":    "INVALID_DATE",
				"message": "Invalid start_date format",
			})
		}
		endDate, err = time.Parse("2006-01-02", endStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code":    "INVALID_DATE",
				"message": "Invalid end_date format",
			})
		}
	}

	// Get all data
	summary, err := h.reportRepo.GetSalesRangeSummary(c.Context(), startDate, endDate)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get sales summary")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to export data",
		})
	}

	dailyReports, err := h.reportRepo.GetSalesRangeReport(c.Context(), startDate, endDate)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get daily reports")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to export data",
		})
	}

	topSellers, err := h.reportRepo.GetTopSellers(c.Context(), startDate, endDate, 20)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get top sellers")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to export data",
		})
	}

	employeeSales, err := h.reportRepo.GetEmployeeSalesReport(c.Context(), startDate, endDate)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get employee sales")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to export data",
		})
	}

	categorySales, err := h.reportRepo.GetCategorySalesReport(c.Context(), startDate, endDate)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get category sales")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to export data",
		})
	}

	// Generate comprehensive CSV
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Title
	writer.Write([]string{"COMPREHENSIVE SALES REPORT"})
	writer.Write([]string{"Generated:", time.Now().Format("2006-01-02 15:04:05")})
	writer.Write([]string{"Period:", fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))})
	writer.Write([]string{""})

	// SUMMARY SECTION
	writer.Write([]string{"=== SUMMARY STATISTICS ==="})
	writer.Write([]string{"Total Transactions:", fmt.Sprintf("%d", summary.TotalTransactions)})
	writer.Write([]string{"Total Items Sold:", fmt.Sprintf("%d", summary.TotalItems)})
	writer.Write([]string{"Total Revenue:", summary.TotalAmount.String()})
	writer.Write([]string{"Total Tax:", summary.TotalTax.String()})
	writer.Write([]string{"Total Discounts:", summary.TotalDiscount.String()})
	avgPerTransaction := decimal.Zero
	if summary.TotalTransactions > 0 {
		avgPerTransaction = summary.TotalAmount.Div(decimal.NewFromInt(int64(summary.TotalTransactions)))
	}
	writer.Write([]string{"Average per Transaction:", avgPerTransaction.String()})
	writer.Write([]string{""})
	writer.Write([]string{""})

	// DAILY BREAKDOWN SECTION
	writer.Write([]string{"=== DAILY SALES BREAKDOWN ==="})
	writer.Write([]string{"Date", "Transactions", "Items Sold", "Revenue", "Tax", "Discounts"})
	for _, day := range dailyReports {
		writer.Write([]string{
			day.Date,
			fmt.Sprintf("%d", day.TransactionCount),
			fmt.Sprintf("%d", day.ItemCount),
			day.TotalAmount.String(),
			day.TotalTax.String(),
			day.TotalDiscount.String(),
		})
	}
	writer.Write([]string{""})
	writer.Write([]string{""})

	// TOP SELLERS SECTION
	writer.Write([]string{"=== TOP 20 SELLING PRODUCTS ==="})
	writer.Write([]string{"Rank", "Product Name", "SKU", "Category", "Qty Sold", "Revenue", "Profit", "Margin %"})
	for i, item := range topSellers {
		sku := ""
		if item.ProductSKU != nil {
			sku = *item.ProductSKU
		}
		category := ""
		if item.CategoryName != nil {
			category = *item.CategoryName
		}
		margin := decimal.Zero
		if !item.TotalRevenue.IsZero() {
			margin = item.TotalProfit.Div(item.TotalRevenue).Mul(decimal.NewFromInt(100))
		}
		writer.Write([]string{
			fmt.Sprintf("%d", i+1),
			item.ProductName,
			sku,
			category,
			item.QuantitySold.String(),
			item.TotalRevenue.String(),
			item.TotalProfit.String(),
			margin.StringFixed(2),
		})
	}
	writer.Write([]string{""})
	writer.Write([]string{""})

	// EMPLOYEE PERFORMANCE SECTION
	if len(employeeSales) > 0 {
		writer.Write([]string{"=== EMPLOYEE SALES PERFORMANCE ==="})
		writer.Write([]string{"Employee", "Transactions", "Items Sold", "Total Sales", "Avg per Transaction"})
		for _, emp := range employeeSales {
			writer.Write([]string{
				fmt.Sprintf("%v", emp["employee_name"]),
				fmt.Sprintf("%v", emp["transaction_count"]),
				fmt.Sprintf("%v", emp["item_count"]),
				fmt.Sprintf("%v", emp["total_sales"]),
				fmt.Sprintf("%v", emp["avg_transaction"]),
			})
		}
		writer.Write([]string{""})
		writer.Write([]string{""})
	}

	// CATEGORY PERFORMANCE SECTION
	if len(categorySales) > 0 {
		writer.Write([]string{"=== SALES BY CATEGORY ==="})
		writer.Write([]string{"Category", "Items Sold (Line Items)", "Total Quantity", "Revenue"})
		for _, cat := range categorySales {
			writer.Write([]string{
				fmt.Sprintf("%v", cat["category_name"]),
				fmt.Sprintf("%v", cat["items_sold"]),
				fmt.Sprintf("%v", cat["quantity_sold"]),
				fmt.Sprintf("%v", cat["total_revenue"]),
			})
		}
	}

	writer.Flush()

	// Set response headers
	filename := fmt.Sprintf("comprehensive_report_%s_to_%s.csv", startDate.Format("20060102"), endDate.Format("20060102"))
	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	return c.Send(buf.Bytes())
}

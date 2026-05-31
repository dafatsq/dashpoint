package handlers

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
)

// GetDailySalesReport handles GET /api/v1/reports/daily
func (h *ReportHandler) GetDailySalesReport(c *fiber.Ctx) error {
	dateStr := c.Query("date", time.Now().In(reportBusinessLocation).Format(reportDateLayout))
	date, err := parseReportDay(dateStr, "date")
	if err != nil {
		return reportError(c, fiber.StatusBadRequest, "INVALID_DATE", "Invalid date format. Use YYYY-MM-DD")
	}

	report, err := h.reportRepo.GetDailySalesReport(c.Context(), date)
	if err != nil {
		return reportInternalError(c, err, "Failed to get daily sales report")
	}

	return c.JSON(fiber.Map{"report": report})
}

// GetSalesRangeReport handles GET /api/v1/reports/sales
func (h *ReportHandler) GetSalesRangeReport(c *fiber.Ctx) error {
	dateRange, err := parseReportRangeResponse(c, 7, false, 90)
	if err != nil {
		return err
	}

	reports, err := h.reportRepo.GetSalesRangeReport(c.Context(), dateRange.start, dateRange.end)
	if err != nil {
		return reportInternalError(c, err, "Failed to get sales range report")
	}

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
		"start_date": dateRange.startStr,
		"end_date":   dateRange.endStr,
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

// GetCashReport handles GET /api/v1/reports/cash
func (h *ReportHandler) GetCashReport(c *fiber.Ctx) error {
	dateRange, err := parseReportRangeResponse(c, 30, false, 0)
	if err != nil {
		return err
	}

	report, err := h.reportRepo.GetCashReport(c.Context(), dateRange.start, dateRange.end)
	if err != nil {
		return reportInternalError(c, err, "Failed to get cash report")
	}

	return c.JSON(fiber.Map{"report": report})
}

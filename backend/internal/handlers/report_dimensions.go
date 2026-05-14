package handlers

import (
	"github.com/gofiber/fiber/v2"
)

// GetTopSellers handles GET /api/v1/reports/top-sellers
func (h *ReportHandler) GetTopSellers(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 10)
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	dateRange, err := parseReportRangeResponse(c, 30, false, 0)
	if err != nil {
		return err
	}

	items, err := h.reportRepo.GetTopSellers(c.Context(), dateRange.start, dateRange.end, limit)
	if err != nil {
		return reportInternalError(c, err, "Failed to get top sellers")
	}

	return c.JSON(fiber.Map{
		"start_date":  dateRange.start.Format(reportDateLayout),
		"end_date":    dateRange.end.Format(reportDateLayout),
		"limit":       limit,
		"top_sellers": items,
	})
}

// GetInventoryValuation handles GET /api/v1/reports/inventory
func (h *ReportHandler) GetInventoryValuation(c *fiber.Ctx) error {
	includeItems := c.QueryBool("include_items", false)

	categoryID, err := parseReportCategoryID(c)
	if err != nil {
		return err
	}

	valuation, repoErr := h.reportRepo.GetInventoryValuation(c.Context(), categoryID, includeItems)
	if repoErr != nil {
		return reportInternalError(c, repoErr, "Failed to get inventory valuation")
	}

	return c.JSON(fiber.Map{"valuation": valuation})
}

// GetEmployeeSalesReport handles GET /api/v1/reports/by-employee
func (h *ReportHandler) GetEmployeeSalesReport(c *fiber.Ctx) error {
	dateRange, err := parseReportRangeResponse(c, 30, false, 0)
	if err != nil {
		return err
	}

	results, repoErr := h.reportRepo.GetEmployeeSalesReport(c.Context(), dateRange.start, dateRange.end)
	if repoErr != nil {
		return reportInternalError(c, repoErr, "Failed to get employee sales report")
	}

	return c.JSON(fiber.Map{
		"start_date": dateRange.start.Format(reportDateLayout),
		"end_date":   dateRange.end.Format(reportDateLayout),
		"employees":  results,
	})
}

// GetCategorySalesReport handles GET /api/v1/reports/by-category
func (h *ReportHandler) GetCategorySalesReport(c *fiber.Ctx) error {
	dateRange, err := parseReportRangeResponse(c, 30, false, 0)
	if err != nil {
		return err
	}

	results, repoErr := h.reportRepo.GetCategorySalesReport(c.Context(), dateRange.start, dateRange.end)
	if repoErr != nil {
		return reportInternalError(c, repoErr, "Failed to get category sales report")
	}

	return c.JSON(fiber.Map{
		"start_date": dateRange.start.Format(reportDateLayout),
		"end_date":   dateRange.end.Format(reportDateLayout),
		"categories": results,
	})
}

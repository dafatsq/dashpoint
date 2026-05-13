package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"

	"dashpoint/backend/internal/middleware"
)

// GetSale handles GET /api/v1/sales/:id.
func (h *SaleHandler) GetSale(c *fiber.Ctx) error {
	id, err := saleParamUUID(c, "id", "INVALID_ID", "Invalid sale ID format")
	if err != nil {
		return respondAPIError(c, err)
	}

	sale, err := h.saleRepo.GetByID(c.Context(), id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get sale")
		return saleInternalError(c, "Failed to retrieve sale")
	}
	if sale == nil {
		return middleware.JSONError(c, fiber.StatusNotFound, "NOT_FOUND", "Sale not found")
	}

	return c.JSON(fiber.Map{"sale": h.toSaleResponse(sale)})
}

// GetSaleByInvoice handles GET /api/v1/sales/invoice/:invoiceNo.
func (h *SaleHandler) GetSaleByInvoice(c *fiber.Ctx) error {
	sale, err := h.saleRepo.GetByInvoiceNo(c.Context(), c.Params("invoiceNo"))
	if err != nil {
		log.Error().Err(err).Msg("Failed to get sale by invoice")
		return saleInternalError(c, "Failed to retrieve sale")
	}
	if sale == nil {
		return middleware.JSONError(c, fiber.StatusNotFound, "NOT_FOUND", "Sale not found")
	}
	return c.JSON(fiber.Map{"sale": h.toSaleResponse(sale)})
}

// ListSales handles GET /api/v1/sales.
func (h *SaleHandler) ListSales(c *fiber.Ctx) error {
	filter, err := parseSaleFilter(c)
	if err != nil {
		return respondAPIError(c, err)
	}

	sales, total, err := h.saleRepo.List(c.Context(), filter)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list sales")
		return saleInternalError(c, "Failed to retrieve sales")
	}

	salesResponse := make([]fiber.Map, 0, len(sales))
	for i := range sales {
		salesResponse = append(salesResponse, h.toSaleListResponse(&sales[i]))
	}

	return c.JSON(fiber.Map{
		"sales":  salesResponse,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

// GetDailySummary handles GET /api/v1/sales/summary/daily.
func (h *SaleHandler) GetDailySummary(c *fiber.Ctx) error {
	dateStr := c.Query("date", time.Now().Format("2006-01-02"))
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return middleware.JSONError(c, fiber.StatusBadRequest, "INVALID_DATE", "Invalid date format. Use YYYY-MM-DD")
	}

	summary, err := h.saleRepo.GetDailySummary(c.Context(), date)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get daily summary")
		return saleInternalError(c, "Failed to retrieve daily summary")
	}

	return c.JSON(fiber.Map{"summary": summary})
}

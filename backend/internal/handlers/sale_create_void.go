package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"dashpoint/backend/internal/audit"
	"dashpoint/backend/internal/middleware"
	"dashpoint/backend/internal/models"
	"dashpoint/backend/internal/repository"
)

// CreateSale handles POST /api/v1/sales.
func (h *SaleHandler) CreateSale(c *fiber.Ctx) error {
	var req CreateSaleRequest
	if err := c.BodyParser(&req); err != nil {
		return saleInvalidRequest(c)
	}

	input, err := parseCreateSaleInput(req)
	if err != nil {
		return respondAPIError(c, err)
	}

	userID := middleware.GetUserID(c)

	var shiftID *uuid.UUID
	shift, err := h.shiftRepo.GetOpenShiftByEmployee(c.Context(), userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get open shift")
		return saleInternalError(c, "Failed to get open shift")
	}
	if shift != nil {
		shiftID = &shift.ID
	} else if middleware.GetRoleName(c) == "cashier" {
		return middleware.JSONError(c, fiber.StatusConflict, "NO_OPEN_SHIFT", "You must start a shift before processing sales")
	}

	createReq := &repository.CreateSaleRequest{
		Items:          input.items,
		Payments:       input.payments,
		EmployeeID:     userID,
		ShiftID:        shiftID,
		DiscountType:   input.discountType,
		DiscountValue:  input.discountValue,
		DiscountReason: input.discountReason,
		Notes:          input.notes,
	}

	sale, err := h.saleRepo.Create(c.Context(), createReq)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create sale")
		return middleware.JSONError(c, fiber.StatusBadRequest, "SALE_FAILED", err.Error())
	}

	newVals := map[string]interface{}{
		"invoice_no":    sale.InvoiceNo,
		"items_summary": saleItemsSummary(sale.Items),
		"total":         sale.TotalAmount.String(),
	}
	audit.LogWithValues(c, models.AuditActionSaleCreate, models.AuditEntitySale, sale.ID.String(), "Created sale", nil, newVals)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Sale completed successfully",
		"sale":    h.toSaleResponse(sale),
	})
}

// VoidSale handles POST /api/v1/sales/:id/void.
func (h *SaleHandler) VoidSale(c *fiber.Ctx) error {
	id, err := saleParamUUID(c, "id", "INVALID_ID", "Invalid sale ID format")
	if err != nil {
		return respondAPIError(c, err)
	}

	var req VoidSaleRequest
	if err := c.BodyParser(&req); err != nil {
		return saleInvalidRequest(c)
	}
	if req.Reason == "" {
		return middleware.JSONError(c, fiber.StatusBadRequest, "REASON_REQUIRED", "Reason is required to void a sale")
	}

	userID := middleware.GetUserID(c)
	if err := h.saleRepo.VoidSale(c.Context(), id, userID, req.Reason); err != nil {
		log.Error().Err(err).Msg("Failed to void sale")
		return middleware.JSONError(c, fiber.StatusBadRequest, "VOID_FAILED", err.Error())
	}

	sale, _ := h.saleRepo.GetByID(c.Context(), id)
	if sale != nil {
		newVals := map[string]interface{}{
			"invoice_no": sale.InvoiceNo,
			"reason":     req.Reason,
		}
		audit.LogWithValues(c, models.AuditActionSaleVoid, models.AuditEntitySale, id.String(), "Voided sale: "+req.Reason, nil, newVals)
	}

	return c.JSON(fiber.Map{
		"message": "Sale voided successfully",
		"sale":    h.toSaleResponse(sale),
	})
}

func saleItemsSummary(items []models.SaleItem) string {
	summary := ""
	for i, item := range items {
		if i > 0 {
			summary += ", "
		}
		summary += fmt.Sprintf("%sx %s", item.Quantity.StringFixed(0), item.ProductName)
	}
	return summary
}

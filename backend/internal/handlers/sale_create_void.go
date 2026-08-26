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

	shiftID, err := h.validateSaleShift(c, input.shiftID)
	if err != nil {
		return err
	}
	if shiftID == nil {
		return nil
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
		if saleErr, ok := asSanitizedSaleError(err); ok {
			return middleware.JSONError(c, fiber.StatusBadRequest, "SALE_FAILED", saleErr)
		}
		return middleware.JSONError(c, fiber.StatusInternalServerError, "SALE_FAILED", "Could not complete the sale. Please try again.")
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

// ValidateCart handles POST /api/v1/sales/validate.
func (h *SaleHandler) ValidateCart(c *fiber.Ctx) error {
	var req ValidateSaleCartRequest
	if err := c.BodyParser(&req); err != nil {
		return saleInvalidRequest(c)
	}

	input, err := parseSaleCartValidationInput(req)
	if err != nil {
		return respondAPIError(c, err)
	}

	if _, err := h.validateSaleShift(c, input.shiftID); err != nil {
		return err
	}
	if c.Response().StatusCode() >= fiber.StatusBadRequest {
		return nil
	}

	if err := h.saleRepo.ValidateCart(c.Context(), &repository.ValidateSaleCartRequest{Items: input.items}); err != nil {
		if saleErr, ok := asSanitizedSaleError(err); ok {
			return middleware.JSONError(c, fiber.StatusBadRequest, "SALE_VALIDATION_FAILED", saleErr)
		}
		log.Error().Err(err).Msg("Cart validation failed with internal error")
		return middleware.JSONError(c, fiber.StatusInternalServerError, "SALE_VALIDATION_FAILED", "Could not validate the cart. Please try again.")
	}

	return c.JSON(fiber.Map{
		"message": "Cart validated successfully",
	})
}

func (h *SaleHandler) validateSaleShift(c *fiber.Ctx, expectedShiftID *uuid.UUID) (*uuid.UUID, error) {
	shift, err := h.shiftRepo.GetCurrentOpenShift(c.Context())
	if err != nil {
		log.Error().Err(err).Msg("Failed to get open shift")
		return nil, saleInternalError(c, "Failed to get open shift")
	}
	if shift == nil {
		return nil, middleware.JSONError(c, fiber.StatusConflict, "NO_OPEN_SHIFT", "You must start a shift before processing sales")
	}
	if expectedShiftID != nil && *expectedShiftID != shift.ID {
		return nil, middleware.JSONError(c, fiber.StatusConflict, "STALE_SUBMIT", staleShiftMessage)
	}
	return &shift.ID, nil
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
	sale, repoErr := h.saleRepo.GetByID(c.Context(), id)
	if repoErr != nil {
		return saleInternalError(c, "Failed to retrieve sale")
	}
	if sale == nil {
		return middleware.JSONError(c, fiber.StatusNotFound, "NOT_FOUND", "Sale not found")
	}
	stale, staleErr := isStaleSubmit(req.ExpectedUpdatedAt, sale.UpdatedAt)
	if staleErr != nil {
		return middleware.JSONError(c, fiber.StatusBadRequest, "INVALID_EXPECTED_UPDATED_AT", "Invalid expected_updated_at")
	}
	if stale {
		return middleware.JSONError(c, fiber.StatusConflict, "STALE_SUBMIT", staleSubmitMessage)
	}

	if err := h.saleRepo.VoidSale(c.Context(), id, userID, req.Reason); err != nil {
		log.Error().Err(err).Msg("Failed to void sale")
		if saleErr, ok := asSanitizedSaleError(err); ok {
			return middleware.JSONError(c, fiber.StatusBadRequest, "VOID_FAILED", saleErr)
		}
		return middleware.JSONError(c, fiber.StatusInternalServerError, "VOID_FAILED", "Could not void the sale. Please try again.")
	}

	sale, _ = h.saleRepo.GetByID(c.Context(), id)
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

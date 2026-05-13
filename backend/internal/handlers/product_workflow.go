package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/shopspring/decimal"

	"dashpoint/backend/internal/models"
)

func (h *ProductHandler) reactivateProduct(c *fiber.Ctx, existingProduct *models.Product, req CreateProductRequest, price decimal.Decimal) (*models.Product, error) {
	existingProduct.IsActive = true
	existingProduct.Name = req.Name
	existingProduct.Description = req.Description
	existingProduct.Barcode = req.Barcode
	existingProduct.Price = price

	if req.Cost != nil {
		cost, err := parseDecimalField(*req.Cost, "cost", false)
		if err != nil {
			return nil, productJSONError(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		}
		existingProduct.Cost = cost
	}
	if req.TaxRate != nil {
		taxRate, err := parseDecimalField(*req.TaxRate, "tax_rate", false)
		if err != nil {
			return nil, productJSONError(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		}
		existingProduct.TaxRate = taxRate
	}
	if req.Unit != nil {
		existingProduct.Unit = *req.Unit
	}
	categoryID, err := parseOptionalUUIDField(req.CategoryID, "category_id")
	if err != nil {
		return nil, productJSONError(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}
	existingProduct.CategoryID = categoryID

	if err := h.productRepo.Update(c.Context(), existingProduct); err != nil {
		return nil, productInternalError(c, err, "Failed to reactivate product", "Failed to reactivate product")
	}
	reactivated, err := h.productRepo.GetByID(c.Context(), existingProduct.ID)
	if err != nil {
		return nil, productInternalError(c, err, "Failed to fetch reactivated product", "Failed to reactivate product")
	}
	if reactivated != nil {
		existingProduct = reactivated
	}
	return existingProduct, nil
}

package handlers

import (
	"github.com/gofiber/fiber/v2"

	"dashpoint/backend/internal/models"
)

// Delete handles DELETE /api/v1/products/:id
func (h *ProductHandler) Delete(c *fiber.Ctx) error {
	id, err := parseUUIDParam(c, "id", "INVALID_ID", "Invalid product ID format")
	if err != nil {
		return err
	}

	productToDelete, repoErr := h.productRepo.GetByID(c.Context(), id)
	if repoErr != nil {
		return productInternalError(c, repoErr, "Failed to get product", "Failed to retrieve product")
	}
	if productToDelete == nil {
		return productJSONError(c, fiber.StatusNotFound, "NOT_FOUND", "Product not found")
	}
	if !productToDelete.IsActive {
		return productJSONError(c, fiber.StatusConflict, "PRODUCT_INACTIVE", "Product is already archived")
	}
	stale, staleErr := isStaleSubmit(expectedUpdatedAtFromQuery(c), productToDelete.UpdatedAt)
	if staleErr != nil {
		return productJSONError(c, fiber.StatusBadRequest, "INVALID_EXPECTED_UPDATED_AT", "Invalid expected_updated_at")
	}
	if stale {
		return productJSONError(c, fiber.StatusConflict, "STALE_SUBMIT", staleSubmitMessage)
	}

	if err := h.productRepo.Delete(c.Context(), id); err != nil {
		return productInternalError(c, err, "Failed to delete product", "Failed to delete product")
	}

	oldValues := map[string]interface{}{
		"affected_product": productToDelete.Name,
		"name":             productToDelete.Name,
	}
	if productToDelete.SKU != nil {
		oldValues["sku"] = *productToDelete.SKU
	}
	logProductAudit(c, models.AuditActionProductArchive, id.String(), "Archived product: "+productToDelete.Name, oldValues, nil)

	return c.JSON(fiber.Map{"message": "Product deleted successfully"})
}

// PermanentDelete handles DELETE /api/v1/products/:id/permanent
func (h *ProductHandler) PermanentDelete(c *fiber.Ctx) error {
	id, err := parseUUIDParam(c, "id", "INVALID_ID", "Invalid product ID format")
	if err != nil {
		return err
	}

	productToDelete, repoErr := h.productRepo.GetByID(c.Context(), id)
	if repoErr != nil {
		return productInternalError(c, repoErr, "Failed to get product", "Failed to retrieve product")
	}
	if productToDelete == nil {
		return productJSONError(c, fiber.StatusNotFound, "NOT_FOUND", "Product not found")
	}
	stale, staleErr := isStaleSubmit(expectedUpdatedAtFromQuery(c), productToDelete.UpdatedAt)
	if staleErr != nil {
		return productJSONError(c, fiber.StatusBadRequest, "INVALID_EXPECTED_UPDATED_AT", "Invalid expected_updated_at")
	}
	if stale {
		return productJSONError(c, fiber.StatusConflict, "STALE_SUBMIT", staleSubmitMessage)
	}

	hasSales, repoErr := h.productRepo.HasSalesHistory(c.Context(), id)
	if repoErr != nil {
		return productInternalError(c, repoErr, "Failed to check sales history", "Failed to check product sales history")
	}
	if hasSales {
		return productJSONError(c, fiber.StatusConflict, "HAS_SALES_HISTORY", "Cannot permanently delete product with sales history. The product must remain archived.")
	}

	if err := h.productRepo.PermanentDelete(c.Context(), id); err != nil {
		return productInternalError(c, err, "Failed to permanently delete product", "Failed to permanently delete product")
	}

	if productToDelete.ImageURL != nil {
		h.deleteImageFile(*productToDelete.ImageURL)
	}

	oldValues := map[string]interface{}{
		"affected_product": productToDelete.Name,
		"name":             productToDelete.Name,
	}
	if productToDelete.SKU != nil {
		oldValues["sku"] = *productToDelete.SKU
	}
	logProductAudit(c, models.AuditActionProductDelete, id.String(), "Permanently deleted product: "+productToDelete.Name, oldValues, nil)

	return c.JSON(fiber.Map{"message": "Product permanently deleted"})
}

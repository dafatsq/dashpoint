package handlers

import (
	"github.com/gofiber/fiber/v2"

	"dashpoint/backend/internal/audit"
	"dashpoint/backend/internal/middleware"
	"dashpoint/backend/internal/models"
)

// GetInventory handles GET /api/v1/products/:id/inventory
func (h *ProductHandler) GetInventory(c *fiber.Ctx) error {
	id, err := parseUUIDParam(c, "id", "INVALID_ID", "Invalid product ID format")
	if err != nil {
		return err
	}

	inventory, repoErr := h.inventoryRepo.GetByProductID(c.Context(), id)
	if repoErr != nil {
		return productInternalError(c, repoErr, "Failed to get inventory", "Failed to retrieve inventory")
	}
	if inventory == nil {
		return productJSONError(c, fiber.StatusNotFound, "NOT_FOUND", "Inventory not found")
	}

	adjustments, total, _ := h.inventoryRepo.GetAdjustmentHistory(c.Context(), id, 10, 0)

	return c.JSON(fiber.Map{
		"inventory": fiber.Map{
			"product_id":          inventory.ProductID.String(),
			"quantity":            inventory.Quantity.String(),
			"available_quantity":  inventory.AvailableQuantity().String(),
			"reserved_quantity":   inventory.ReservedQuantity.String(),
			"low_stock_threshold": inventory.LowStockThreshold.String(),
			"reorder_quantity":    inventory.ReorderQuantity.String(),
			"is_low_stock":        inventory.IsLowStock(),
			"last_counted_at":     inventory.LastCountedAt,
			"last_restocked_at":   inventory.LastRestockedAt,
		},
		"recent_adjustments": adjustments,
		"total_adjustments":  total,
	})
}

// AdjustStock handles POST /api/v1/inventory/adjust
func (h *ProductHandler) AdjustStock(c *fiber.Ctx) error {
	req, err := parseStockAdjustmentRequest(c)
	if err != nil {
		return err
	}

	userID := middleware.GetUserID(c)
	quantity := req.Quantity
	var adjustment *models.StockAdjustment

	if req.AdjustmentType == models.AdjustmentCount {
		adjustment, err = h.inventoryRepo.SetQuantity(c.Context(), req.ProductID, quantity, req.Reason, userID)
	} else {
		if req.AdjustmentType == models.AdjustmentDamage || req.AdjustmentType == models.AdjustmentLoss {
			quantity = quantity.Neg()
		}
		adjustment, err = h.inventoryRepo.AdjustStock(c.Context(), req.ProductID, req.AdjustmentType, quantity, req.Reason, nil, nil, userID)
	}
	if err != nil {
		return productJSONError(c, fiber.StatusBadRequest, "ADJUSTMENT_FAILED", err.Error())
	}

	product, _ := h.productRepo.GetByID(c.Context(), req.ProductID)
	productName := ""
	if product != nil {
		productName = product.Name
	}
	auditAction := models.AuditActionStockAdjust
	if req.AdjustmentType == models.AdjustmentCount {
		auditAction = models.AuditActionStockCount
	}
	auditValues := map[string]interface{}{
		"affected_product": productName,
		"product_name":     productName,
		"adjustment_type":  string(req.AdjustmentType),
		"quantity":         quantity.String(),
	}
	if req.Reason != nil {
		auditValues["reason"] = *req.Reason
	}
	if adjustment != nil {
		auditValues["new_quantity"] = adjustment.QuantityAfter.String()
	}
	audit.LogWithValues(c, auditAction, models.AuditEntityInventory, req.ProductID.String(), "Stock adjusted: "+productName, nil, auditValues)

	return c.JSON(fiber.Map{"message": "Stock adjusted successfully", "adjustment": adjustment})
}

// GetLowStock handles GET /api/v1/inventory/low-stock
func (h *ProductHandler) GetLowStock(c *fiber.Ctx) error {
	products, err := h.inventoryRepo.GetLowStockProducts(c.Context())
	if err != nil {
		return productInternalError(c, err, "Failed to get low stock products", "Failed to retrieve low stock products")
	}

	return c.JSON(fiber.Map{"products": products, "count": len(products)})
}

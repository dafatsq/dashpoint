package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

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

	limit := c.QueryInt("limit", 10)
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := c.QueryInt("offset", 0)
	if offset < 0 {
		offset = 0
	}
	adjustmentType, err := parseInventoryAdjustmentTypeFilter(c.Query("adjustment_type"))
	if err != nil {
		return productJSONError(c, fiber.StatusBadRequest, "INVALID_ADJUSTMENT_TYPE", err.Error())
	}
	adjustedBy, err := parseInventoryAdjustedByFilter(c.Query("user_id"))
	if err != nil {
		return productJSONError(c, fiber.StatusBadRequest, "INVALID_USER_ID", err.Error())
	}

	var startDate, endDate *time.Time
	if startStr := c.Query("from"); startStr != "" {
		parsed, err := parseReportDay(startStr, "from")
		if err != nil {
			return productJSONError(c, fiber.StatusBadRequest, "INVALID_START_DATE", "Invalid from format. Use YYYY-MM-DD")
		}
		parsed = reportDayStart(parsed)
		startDate = &parsed
	}
	if endStr := c.Query("to"); endStr != "" {
		parsed, err := parseReportDay(endStr, "to")
		if err != nil {
			return productJSONError(c, fiber.StatusBadRequest, "INVALID_END_DATE", "Invalid to format. Use YYYY-MM-DD")
		}
		exclusiveEnd := reportDayStart(parsed).Add(24 * time.Hour)
		endDate = &exclusiveEnd
	}

	inventory, repoErr := h.inventoryRepo.GetByProductID(c.Context(), id)
	if repoErr != nil {
		return productInternalError(c, repoErr, "Failed to get inventory", "Failed to retrieve inventory")
	}
	if inventory == nil {
		return productJSONError(c, fiber.StatusNotFound, "NOT_FOUND", "Inventory not found")
	}

	adjustments, total, _ := h.inventoryRepo.GetAdjustmentHistory(c.Context(), id, limit, offset, adjustmentType, adjustedBy, startDate, endDate)

	return c.JSON(fiber.Map{
		"inventory":          inventoryResponse(inventory),
		"recent_adjustments": adjustments,
		"total_adjustments":  total,
	})
}

func parseInventoryAdjustedByFilter(value string) (*uuid.UUID, error) {
	if value == "" || value == "all" {
		return nil, nil
	}

	userID, err := uuid.Parse(value)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, "Invalid user_id format")
	}

	return &userID, nil
}

func parseInventoryAdjustmentTypeFilter(value string) (*models.AdjustmentType, error) {
	if value == "" || value == "all" {
		return nil, nil
	}

	adjustmentType := models.AdjustmentType(value)
	switch adjustmentType {
	case models.AdjustmentInitial,
		models.AdjustmentPurchase,
		models.AdjustmentSale,
		models.AdjustmentReturn,
		models.AdjustmentAdjustment,
		models.AdjustmentDamage,
		models.AdjustmentLoss,
		models.AdjustmentTransfer,
		models.AdjustmentCount:
		return &adjustmentType, nil
	default:
		return nil, fiber.NewError(fiber.StatusBadRequest, "Invalid adjustment_type value")
	}
}

// UpdateInventoryThreshold handles PATCH /api/v1/products/:id/inventory
func (h *ProductHandler) UpdateInventoryThreshold(c *fiber.Ctx) error {
	id, err := parseUUIDParam(c, "id", "INVALID_ID", "Invalid product ID format")
	if err != nil {
		return err
	}

	req, err := parseInventoryThresholdUpdateRequest(c)
	if err != nil {
		return err
	}
	if req == nil {
		return nil
	}

	product, repoErr := h.productRepo.GetByID(c.Context(), id)
	if repoErr != nil {
		return productInternalError(c, repoErr, "Failed to get product", "Failed to retrieve product")
	}
	if product == nil {
		return productJSONError(c, fiber.StatusNotFound, "NOT_FOUND", "Product not found")
	}
	if !product.IsActive {
		return productJSONError(c, fiber.StatusConflict, "PRODUCT_INACTIVE", "Archived products cannot be changed")
	}

	inventory, repoErr := h.inventoryRepo.GetByProductID(c.Context(), id)
	if repoErr != nil {
		return productInternalError(c, repoErr, "Failed to get inventory", "Failed to retrieve inventory")
	}
	if inventory == nil {
		return productJSONError(c, fiber.StatusNotFound, "NOT_FOUND", "Inventory not found")
	}
	stale, staleErr := isStaleSubmit(req.ExpectedUpdatedAt, inventory.UpdatedAt)
	if staleErr != nil {
		return productJSONError(c, fiber.StatusBadRequest, "INVALID_EXPECTED_UPDATED_AT", "Invalid expected_updated_at")
	}
	if stale {
		return productJSONError(c, fiber.StatusConflict, "STALE_SUBMIT", staleSubmitMessage)
	}

	oldThreshold := inventory.LowStockThreshold
	if err := h.inventoryRepo.UpdateThresholds(c.Context(), id, req.LowStockThreshold); err != nil {
		return productInternalError(c, err, "Failed to update threshold", "Failed to update low stock threshold")
	}

	updatedInventory, repoErr := h.inventoryRepo.GetByProductID(c.Context(), id)
	if repoErr != nil {
		return productInternalError(c, repoErr, "Failed to get updated inventory", "Failed to retrieve inventory")
	}
	if updatedInventory == nil {
		return productJSONError(c, fiber.StatusNotFound, "NOT_FOUND", "Inventory not found")
	}

	audit.LogWithValues(
		c,
		models.AuditActionThresholdUpdate,
		models.AuditEntityInventory,
		id.String(),
		"Updated low stock threshold: "+product.Name,
		map[string]interface{}{
			"low_stock_threshold": oldThreshold.String(),
		},
		map[string]interface{}{
			"affected_product":    product.Name,
			"product_name":        product.Name,
			"low_stock_threshold": updatedInventory.LowStockThreshold.String(),
		},
	)

	return c.JSON(fiber.Map{
		"message":   "Low stock threshold updated successfully",
		"inventory": inventoryResponse(updatedInventory),
	})
}

func inventoryResponse(inventory *models.InventoryItem) fiber.Map {
	return fiber.Map{
		"product_id":          inventory.ProductID.String(),
		"quantity":            inventory.Quantity.String(),
		"available_quantity":  inventory.AvailableQuantity().String(),
		"low_stock_threshold": inventory.LowStockThreshold.String(),
		"is_low_stock":        inventory.IsLowStock(),
		"updated_at":          inventory.UpdatedAt.Format(time.RFC3339Nano),
	}
}

// AdjustStock handles POST /api/v1/inventory/adjust
func (h *ProductHandler) AdjustStock(c *fiber.Ctx) error {
	req, err := parseStockAdjustmentRequest(c)
	if err != nil {
		return err
	}

	product, repoErr := h.productRepo.GetByID(c.Context(), req.ProductID)
	if repoErr != nil {
		return productInternalError(c, repoErr, "Failed to get product", "Failed to retrieve product")
	}
	if product == nil {
		return productJSONError(c, fiber.StatusNotFound, "NOT_FOUND", "Product not found")
	}
	if !product.IsActive {
		return productJSONError(c, fiber.StatusConflict, "PRODUCT_INACTIVE", "Archived products cannot be changed")
	}
	inventory, repoErr := h.inventoryRepo.GetByProductID(c.Context(), req.ProductID)
	if repoErr != nil {
		return productInternalError(c, repoErr, "Failed to get inventory", "Failed to retrieve inventory")
	}
	if inventory == nil {
		return productJSONError(c, fiber.StatusNotFound, "NOT_FOUND", "Inventory not found")
	}
	stale, staleErr := isStaleSubmit(req.ExpectedUpdatedAt, inventory.UpdatedAt)
	if staleErr != nil {
		return productJSONError(c, fiber.StatusBadRequest, "INVALID_EXPECTED_UPDATED_AT", "Invalid expected_updated_at")
	}
	if stale {
		return productJSONError(c, fiber.StatusConflict, "STALE_SUBMIT", staleSubmitMessage)
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

	productName := product.Name
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

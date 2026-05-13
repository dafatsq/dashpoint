package handlers

import (
	"github.com/gofiber/fiber/v2"

	"dashpoint/backend/internal/audit"
	"dashpoint/backend/internal/models"
)

func buildProductAuditValues(product *models.Product) map[string]interface{} {
	values := map[string]interface{}{
		"affected_product":     product.Name,
		"name":                 product.Name,
		"price":                product.Price.String(),
		"cost":                 product.Cost.String(),
		"tax_rate":             product.TaxRate.String(),
		"unit":                 product.Unit,
		"is_active":            product.IsActive,
		"track_inventory":      product.TrackInventory,
		"allow_negative_stock": product.AllowNegativeStock,
	}
	if product.SKU != nil {
		values["sku"] = *product.SKU
	}
	if product.Barcode != nil {
		values["barcode"] = *product.Barcode
	}
	if product.Description != nil {
		values["description"] = *product.Description
	}
	if product.CategoryID != nil {
		values["category_id"] = product.CategoryID.String()
	}
	if product.ImageURL != nil {
		values["image_url"] = *product.ImageURL
	}
	return values
}

func logProductAudit(c *fiber.Ctx, action models.AuditAction, entityID string, description string, oldValues, newValues map[string]interface{}) {
	audit.LogWithValues(c, action, models.AuditEntityProduct, entityID, description, oldValues, newValues)
}

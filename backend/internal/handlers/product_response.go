package handlers

import "dashpoint/backend/internal/models"

func (h *ProductHandler) toProductResponse(p *models.Product) ProductResponse {
	response := ProductResponse{
		ID:                 p.ID.String(),
		SKU:                p.SKU,
		Barcode:            p.Barcode,
		Name:               p.Name,
		Description:        p.Description,
		Price:              p.Price.String(),
		Cost:               p.Cost.String(),
		TaxRate:            p.TaxRate.String(),
		IsActive:           p.IsActive,
		ImageURL:           p.ImageURL,
		CreatedAt:          p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:          p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if p.CategoryID != nil {
		catIDStr := p.CategoryID.String()
		response.CategoryID = &catIDStr
	}
	if p.Category != nil {
		response.CategoryName = &p.Category.Name
	}
	if p.Inventory != nil {
		response.Inventory = &InventoryResponse{
			Quantity:          p.Inventory.Quantity.String(),
			AvailableQuantity: p.Inventory.AvailableQuantity().String(),
			LowStockThreshold: p.Inventory.LowStockThreshold.String(),
			IsLowStock:        p.Inventory.IsLowStock(),
			UpdatedAt:         p.Inventory.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}
	return response
}

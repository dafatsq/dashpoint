package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// InventoryItem represents stock levels for a product
type InventoryItem struct {
	ProductID         uuid.UUID       `json:"product_id"`
	Quantity          decimal.Decimal `json:"quantity"`
	LowStockThreshold decimal.Decimal `json:"low_stock_threshold"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

func (i *InventoryItem) AvailableQuantity() decimal.Decimal {
	return i.Quantity
}

func (i *InventoryItem) IsLowStock() bool {
	return i.Quantity.LessThanOrEqual(i.LowStockThreshold)
}

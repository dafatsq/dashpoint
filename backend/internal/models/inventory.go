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
	ReservedQuantity  decimal.Decimal `json:"reserved_quantity"`
	LowStockThreshold decimal.Decimal `json:"low_stock_threshold"`
	ReorderQuantity   decimal.Decimal `json:"reorder_quantity"`
	LastCountedAt     *time.Time      `json:"last_counted_at,omitempty"`
	LastRestockedAt   *time.Time      `json:"last_restocked_at,omitempty"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

func (i *InventoryItem) AvailableQuantity() decimal.Decimal {
	return i.Quantity.Sub(i.ReservedQuantity)
}

func (i *InventoryItem) IsLowStock() bool {
	return i.Quantity.LessThanOrEqual(i.LowStockThreshold)
}

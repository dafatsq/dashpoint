package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Product represents a product in the catalog
type Product struct {
	ID          uuid.UUID       `json:"id"`
	SKU         *string         `json:"sku,omitempty"`
	Barcode     *string         `json:"barcode,omitempty"`
	Name        string          `json:"name"`
	Description *string         `json:"description,omitempty"`
	CategoryID  *uuid.UUID      `json:"category_id,omitempty"`
	Price       decimal.Decimal `json:"price"`
	Cost        decimal.Decimal `json:"cost"`
	TaxRate     decimal.Decimal `json:"tax_rate"`
	IsActive    bool            `json:"is_active"`
	ImageURL    *string         `json:"image_url,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`

	Category  *Category      `json:"category,omitempty"`
	Inventory *InventoryItem `json:"inventory,omitempty"`
}

func (p *Product) CalculateTax(quantity decimal.Decimal) decimal.Decimal {
	subtotal := p.Price.Mul(quantity)
	taxRate := p.TaxRate.Div(decimal.NewFromInt(100))
	return subtotal.Mul(taxRate).Round(2)
}

func (p *Product) CalculateTotal(quantity decimal.Decimal) decimal.Decimal {
	subtotal := p.Price.Mul(quantity)
	tax := p.CalculateTax(quantity)
	return subtotal.Add(tax).Round(2)
}

// ProductWithInventory is a convenience struct for products with inventory
type ProductWithInventory struct {
	*Product
	Quantity          decimal.Decimal `json:"quantity"`
	AvailableQuantity decimal.Decimal `json:"available_quantity"`
	IsLowStock        bool            `json:"is_low_stock"`
}

package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Category represents a product category
type Category struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
	SortOrder   int        `json:"sort_order"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	// Joined fields
	Parent   *Category   `json:"parent,omitempty"`
	Children []*Category `json:"children,omitempty"`
}

// Product represents a product in the catalog
type Product struct {
	ID                 uuid.UUID       `json:"id"`
	SKU                *string         `json:"sku,omitempty"`
	Barcode            *string         `json:"barcode,omitempty"`
	Name               string          `json:"name"`
	Description        *string         `json:"description,omitempty"`
	CategoryID         *uuid.UUID      `json:"category_id,omitempty"`
	Price              decimal.Decimal `json:"price"`
	Cost               decimal.Decimal `json:"cost"`
	TaxRate            decimal.Decimal `json:"tax_rate"`
	Unit               string          `json:"unit"`
	IsActive           bool            `json:"is_active"`
	TrackInventory     bool            `json:"track_inventory"`
	AllowNegativeStock bool            `json:"allow_negative_stock"`
	ImageURL           *string         `json:"image_url,omitempty"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`

	// Joined fields
	Category  *Category      `json:"category,omitempty"`
	Inventory *InventoryItem `json:"inventory,omitempty"`
}

// CalculateTax calculates the tax amount for a given quantity
func (p *Product) CalculateTax(quantity decimal.Decimal) decimal.Decimal {
	subtotal := p.Price.Mul(quantity)
	taxRate := p.TaxRate.Div(decimal.NewFromInt(100))
	return subtotal.Mul(taxRate).Round(2)
}

// CalculateTotal calculates the total including tax for a given quantity
func (p *Product) CalculateTotal(quantity decimal.Decimal) decimal.Decimal {
	subtotal := p.Price.Mul(quantity)
	tax := p.CalculateTax(quantity)
	return subtotal.Add(tax).Round(2)
}

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

// AvailableQuantity returns the quantity available for sale
func (i *InventoryItem) AvailableQuantity() decimal.Decimal {
	return i.Quantity.Sub(i.ReservedQuantity)
}

// IsLowStock returns true if quantity is at or below the low stock threshold
func (i *InventoryItem) IsLowStock() bool {
	return i.Quantity.LessThanOrEqual(i.LowStockThreshold)
}

// AdjustmentType represents the type of stock adjustment
type AdjustmentType string

const (
	AdjustmentInitial    AdjustmentType = "initial"
	AdjustmentPurchase   AdjustmentType = "purchase"
	AdjustmentSale       AdjustmentType = "sale"
	AdjustmentReturn     AdjustmentType = "return"
	AdjustmentAdjustment AdjustmentType = "adjustment"
	AdjustmentDamage     AdjustmentType = "damage"
	AdjustmentLoss       AdjustmentType = "loss"
	AdjustmentTransfer   AdjustmentType = "transfer"
	AdjustmentCount      AdjustmentType = "count"
)

// StockAdjustment represents a stock adjustment record
type StockAdjustment struct {
	ID             uuid.UUID       `json:"id"`
	ProductID      uuid.UUID       `json:"product_id"`
	AdjustmentType AdjustmentType  `json:"adjustment_type"`
	QuantityBefore decimal.Decimal `json:"quantity_before"`
	QuantityChange decimal.Decimal `json:"quantity_change"`
	QuantityAfter  decimal.Decimal `json:"quantity_after"`
	Reason         *string         `json:"reason,omitempty"`
	ReferenceType  *string         `json:"reference_type,omitempty"`
	ReferenceID    *uuid.UUID      `json:"reference_id,omitempty"`
	AdjustedBy     uuid.UUID       `json:"adjusted_by"`
	CreatedAt      time.Time       `json:"created_at"`

	// Joined fields
	Product        *Product `json:"product,omitempty"`
	AdjustedByUser *User    `json:"adjusted_by_user,omitempty"`
}

// ProductWithInventory is a convenience struct for products with inventory
type ProductWithInventory struct {
	*Product
	Quantity          decimal.Decimal `json:"quantity"`
	AvailableQuantity decimal.Decimal `json:"available_quantity"`
	IsLowStock        bool            `json:"is_low_stock"`
}

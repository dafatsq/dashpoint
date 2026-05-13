package handlers

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"dashpoint/backend/internal/models"
	"dashpoint/backend/internal/repository"
)

type productStore interface {
	List(context.Context, repository.ProductFilter) ([]*models.Product, int, error)
	GetByID(context.Context, uuid.UUID) (*models.Product, error)
	Lookup(context.Context, string) (*models.Product, error)
	Create(context.Context, *models.Product, *decimal.Decimal) error
	Update(context.Context, *models.Product) error
	Delete(context.Context, uuid.UUID) error
	HasSalesHistory(context.Context, uuid.UUID) (bool, error)
	PermanentDelete(context.Context, uuid.UUID) error
	GetBySKUIncludingInactive(context.Context, string) (*models.Product, error)
	GetByBarcodeIncludingInactive(context.Context, string) (*models.Product, error)
	SKUExists(context.Context, string, *uuid.UUID) (bool, error)
	BarcodeExists(context.Context, string, *uuid.UUID) (bool, error)
}

type inventoryStore interface {
	GetByProductID(context.Context, uuid.UUID) (*models.InventoryItem, error)
	AdjustStock(context.Context, uuid.UUID, models.AdjustmentType, decimal.Decimal, *string, *string, *uuid.UUID, uuid.UUID) (*models.StockAdjustment, error)
	SetQuantity(context.Context, uuid.UUID, decimal.Decimal, *string, uuid.UUID) (*models.StockAdjustment, error)
	UpdateThresholds(context.Context, uuid.UUID, decimal.Decimal, decimal.Decimal) error
	GetLowStockProducts(context.Context) ([]*models.ProductWithInventory, error)
	GetAdjustmentHistory(context.Context, uuid.UUID, int, int) ([]*models.StockAdjustment, int, error)
}

type categoryStore interface {
	GetByID(context.Context, uuid.UUID) (*models.Category, error)
}

// ProductHandler handles product endpoints
type ProductHandler struct {
	productRepo   productStore
	inventoryRepo inventoryStore
	categoryRepo  categoryStore
	uploadDir     string
}

// NewProductHandler creates a new product handler
func NewProductHandler(
	productRepo productStore,
	inventoryRepo inventoryStore,
	categoryRepo categoryStore,
	uploadDir string,
) *ProductHandler {
	return &ProductHandler{
		productRepo:   productRepo,
		inventoryRepo: inventoryRepo,
		categoryRepo:  categoryRepo,
		uploadDir:     uploadDir,
	}
}

// ProductResponse represents a product in API responses
type ProductResponse struct {
	ID                 string             `json:"id"`
	SKU                *string            `json:"sku,omitempty"`
	Barcode            *string            `json:"barcode,omitempty"`
	Name               string             `json:"name"`
	Description        *string            `json:"description,omitempty"`
	CategoryID         *string            `json:"category_id,omitempty"`
	CategoryName       *string            `json:"category_name,omitempty"`
	Price              string             `json:"price"`
	Cost               string             `json:"cost"`
	TaxRate            string             `json:"tax_rate"`
	Unit               string             `json:"unit"`
	IsActive           bool               `json:"is_active"`
	TrackInventory     bool               `json:"track_inventory"`
	AllowNegativeStock bool               `json:"allow_negative_stock"`
	ImageURL           *string            `json:"image_url,omitempty"`
	Inventory          *InventoryResponse `json:"inventory,omitempty"`
	CreatedAt          string             `json:"created_at"`
	UpdatedAt          string             `json:"updated_at"`
}

type InventoryResponse struct {
	Quantity          string `json:"quantity"`
	AvailableQuantity string `json:"available_quantity"`
	LowStockThreshold string `json:"low_stock_threshold"`
	IsLowStock        bool   `json:"is_low_stock"`
}

type CreateProductRequest struct {
	SKU                *string `json:"sku"`
	Barcode            *string `json:"barcode"`
	Name               string  `json:"name"`
	Description        *string `json:"description"`
	CategoryID         *string `json:"category_id"`
	Price              string  `json:"price"`
	Cost               *string `json:"cost"`
	TaxRate            *string `json:"tax_rate"`
	Unit               *string `json:"unit"`
	TrackInventory     *bool   `json:"track_inventory"`
	AllowNegativeStock *bool   `json:"allow_negative_stock"`
	InitialQuantity    *string `json:"initial_quantity"`
	LowStockThreshold  *string `json:"low_stock_threshold"`
	ImageURL           *string `json:"image_url"`
}

type UpdateProductRequest struct {
	SKU                *string `json:"sku"`
	Barcode            *string `json:"barcode"`
	Name               *string `json:"name"`
	Description        *string `json:"description"`
	CategoryID         *string `json:"category_id"`
	Price              *string `json:"price"`
	Cost               *string `json:"cost"`
	TaxRate            *string `json:"tax_rate"`
	Unit               *string `json:"unit"`
	IsActive           *bool   `json:"is_active"`
	TrackInventory     *bool   `json:"track_inventory"`
	AllowNegativeStock *bool   `json:"allow_negative_stock"`
	ImageURL           *string `json:"image_url"`
}

type productCreateInput struct {
	product           *models.Product
	initialQuantity   *decimal.Decimal
	lowStockThreshold *decimal.Decimal
}

type stockAdjustmentRequest struct {
	ProductID      uuid.UUID
	AdjustmentType models.AdjustmentType
	Quantity       decimal.Decimal
	Reason         *string
}

package repository

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"dashpoint/backend/internal/models"
)

// ProductRepository handles product database operations
type ProductRepository struct {
	pool *pgxpool.Pool
}

// NewProductRepository creates a new product repository
func NewProductRepository(pool *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{pool: pool}
}

// ProductFilter represents filters for product queries
type ProductFilter struct {
	Search     string
	CategoryID *uuid.UUID
	IsActive   *bool
	LowStock   bool
	Limit      int
	Offset     int
}

func hydrateProductRelations(product *models.Product, catID, catName, catDesc *string, invQty, invReserved, invThreshold *decimal.Decimal) {
	if catID != nil && catName != nil {
		catUUID, _ := uuid.Parse(*catID)
		product.Category = &models.Category{
			ID:          catUUID,
			Name:        *catName,
			Description: catDesc,
		}
	}

	if invQty != nil {
		product.Inventory = &models.InventoryItem{
			ProductID:         product.ID,
			Quantity:          *invQty,
			ReservedQuantity:  decimal.Zero,
			LowStockThreshold: decimal.Zero,
		}
		if invReserved != nil {
			product.Inventory.ReservedQuantity = *invReserved
		}
		if invThreshold != nil {
			product.Inventory.LowStockThreshold = *invThreshold
		}
	}
}

func buildProductWhereClause(filter ProductFilter) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argNum := 1

	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf(
			"(p.name ILIKE $%d OR p.sku ILIKE $%d OR p.barcode ILIKE $%d OR p.description ILIKE $%d)",
			argNum, argNum, argNum, argNum,
		))
		args = append(args, "%"+filter.Search+"%")
		argNum++
	}

	if filter.CategoryID != nil {
		conditions = append(conditions, fmt.Sprintf("p.category_id = $%d", argNum))
		args = append(args, *filter.CategoryID)
		argNum++
	}

	if filter.IsActive != nil {
		conditions = append(conditions, fmt.Sprintf("p.is_active = $%d", argNum))
		args = append(args, *filter.IsActive)
		argNum++
	}

	if filter.LowStock {
		conditions = append(conditions, "i.quantity <= i.low_stock_threshold")
	}

	return strings.Join(conditions, " AND "), args
}

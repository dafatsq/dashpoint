package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

// Create creates a new product with optional inventory
func (r *ProductRepository) Create(ctx context.Context, product *models.Product, initialQuantity *decimal.Decimal) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	now := time.Now()
	product.ID = uuid.New()
	product.CreatedAt = now
	product.UpdatedAt = now

	query := `
		INSERT INTO products (id, sku, barcode, name, description, category_id, price, cost, tax_rate, unit, is_active, track_inventory, allow_negative_stock, image_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`

	_, err = tx.Exec(ctx, query,
		product.ID,
		product.SKU,
		product.Barcode,
		product.Name,
		product.Description,
		product.CategoryID,
		product.Price,
		product.Cost,
		product.TaxRate,
		product.Unit,
		product.IsActive,
		product.TrackInventory,
		product.AllowNegativeStock,
		product.ImageURL,
		product.CreatedAt,
		product.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}

	// Create inventory record if tracking inventory
	if product.TrackInventory {
		qty := decimal.Zero
		if initialQuantity != nil {
			qty = *initialQuantity
		}

		invQuery := `
			INSERT INTO inventory_items (product_id, quantity, created_at, updated_at)
			VALUES ($1, $2, $3, $3)
		`
		_, err = tx.Exec(ctx, invQuery, product.ID, qty, now)
		if err != nil {
			return fmt.Errorf("failed to create inventory record: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetByID retrieves a product by ID with optional inventory
func (r *ProductRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	query := `
		SELECT p.id, p.sku, p.barcode, p.name, p.description, p.category_id, p.price, p.cost, p.tax_rate, p.unit, 
		       p.is_active, p.track_inventory, p.allow_negative_stock, p.image_url, p.created_at, p.updated_at,
		       c.id, c.name, c.description,
		       i.quantity, i.reserved_quantity, i.low_stock_threshold
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		LEFT JOIN inventory_items i ON p.id = i.product_id
		WHERE p.id = $1
	`

	product := &models.Product{}
	var catID, catName, catDesc *string
	var invQty, invReserved, invThreshold *decimal.Decimal

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&product.ID,
		&product.SKU,
		&product.Barcode,
		&product.Name,
		&product.Description,
		&product.CategoryID,
		&product.Price,
		&product.Cost,
		&product.TaxRate,
		&product.Unit,
		&product.IsActive,
		&product.TrackInventory,
		&product.AllowNegativeStock,
		&product.ImageURL,
		&product.CreatedAt,
		&product.UpdatedAt,
		&catID,
		&catName,
		&catDesc,
		&invQty,
		&invReserved,
		&invThreshold,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	// Set category if exists
	if catID != nil && catName != nil {
		catUUID, _ := uuid.Parse(*catID)
		product.Category = &models.Category{
			ID:          catUUID,
			Name:        *catName,
			Description: catDesc,
		}
	}

	// Set inventory if exists
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

	return product, nil
}

// GetByBarcode retrieves a product by barcode
func (r *ProductRepository) GetByBarcode(ctx context.Context, barcode string) (*models.Product, error) {
	query := `SELECT id FROM products WHERE barcode = $1 AND is_active = true`
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, query, barcode).Scan(&id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get product by barcode: %w", err)
	}
	return r.GetByID(ctx, id)
}

// GetByBarcodeIncludingInactive retrieves a product by barcode including inactive products
func (r *ProductRepository) GetByBarcodeIncludingInactive(ctx context.Context, barcode string) (*models.Product, error) {
	query := `SELECT id FROM products WHERE barcode = $1`
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, query, barcode).Scan(&id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get product by barcode: %w", err)
	}
	return r.GetByID(ctx, id)
}

// GetBySKU retrieves a product by SKU
func (r *ProductRepository) GetBySKU(ctx context.Context, sku string) (*models.Product, error) {
	query := `SELECT id FROM products WHERE sku = $1 AND is_active = true`
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, query, sku).Scan(&id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get product by SKU: %w", err)
	}
	return r.GetByID(ctx, id)
}

// GetBySKUIncludingInactive retrieves a product by SKU including inactive products
func (r *ProductRepository) GetBySKUIncludingInactive(ctx context.Context, sku string) (*models.Product, error) {
	query := `SELECT id FROM products WHERE sku = $1`
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, query, sku).Scan(&id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get product by SKU: %w", err)
	}
	return r.GetByID(ctx, id)
}

// Lookup searches for a product by barcode or SKU (for POS quick lookup)
func (r *ProductRepository) Lookup(ctx context.Context, code string) (*models.Product, error) {
	query := `SELECT id FROM products WHERE (barcode = $1 OR sku = $1) AND is_active = true LIMIT 1`
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, query, code).Scan(&id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to lookup product: %w", err)
	}
	return r.GetByID(ctx, id)
}

// List retrieves products with filtering and pagination
func (r *ProductRepository) List(ctx context.Context, filter ProductFilter) ([]*models.Product, int, error) {
	// Build count query
	countQuery := `SELECT COUNT(*) FROM products p`
	whereClause, args := r.buildWhereClause(filter)
	if whereClause != "" {
		countQuery += " WHERE " + whereClause
	}

	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}

	// Build list query
	query := `
		SELECT p.id, p.sku, p.barcode, p.name, p.description, p.category_id, p.price, p.cost, p.tax_rate, p.unit,
		       p.is_active, p.track_inventory, p.allow_negative_stock, p.image_url, p.created_at, p.updated_at,
		       c.id, c.name,
		       i.quantity, i.reserved_quantity, i.low_stock_threshold
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		LEFT JOIN inventory_items i ON p.id = i.product_id
	`

	if whereClause != "" {
		query += " WHERE " + whereClause
	}

	query += " ORDER BY p.name ASC"
	query += fmt.Sprintf(" LIMIT %d OFFSET %d", filter.Limit, filter.Offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query products: %w", err)
	}
	defer rows.Close()

	var products []*models.Product
	for rows.Next() {
		product := &models.Product{}
		var catID, catName *string
		var invQty, invReserved, invThreshold *decimal.Decimal

		err := rows.Scan(
			&product.ID,
			&product.SKU,
			&product.Barcode,
			&product.Name,
			&product.Description,
			&product.CategoryID,
			&product.Price,
			&product.Cost,
			&product.TaxRate,
			&product.Unit,
			&product.IsActive,
			&product.TrackInventory,
			&product.AllowNegativeStock,
			&product.ImageURL,
			&product.CreatedAt,
			&product.UpdatedAt,
			&catID,
			&catName,
			&invQty,
			&invReserved,
			&invThreshold,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan product: %w", err)
		}

		if catID != nil && catName != nil {
			catUUID, _ := uuid.Parse(*catID)
			product.Category = &models.Category{
				ID:   catUUID,
				Name: *catName,
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

		products = append(products, product)
	}

	return products, total, nil
}

// Update updates a product
func (r *ProductRepository) Update(ctx context.Context, product *models.Product) error {
	product.UpdatedAt = time.Now()

	query := `
		UPDATE products 
		SET sku = $1, barcode = $2, name = $3, description = $4, category_id = $5, 
		    price = $6, cost = $7, tax_rate = $8, unit = $9, is_active = $10,
		    track_inventory = $11, allow_negative_stock = $12, image_url = $13, updated_at = $14
		WHERE id = $15
	`

	result, err := r.pool.Exec(ctx, query,
		product.SKU,
		product.Barcode,
		product.Name,
		product.Description,
		product.CategoryID,
		product.Price,
		product.Cost,
		product.TaxRate,
		product.Unit,
		product.IsActive,
		product.TrackInventory,
		product.AllowNegativeStock,
		product.ImageURL,
		product.UpdatedAt,
		product.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("product not found")
	}

	return nil
}

// Delete soft-deletes a product
func (r *ProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE products SET is_active = false, updated_at = $1 WHERE id = $2`
	result, err := r.pool.Exec(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("product not found")
	}
	return nil
}

// HasSalesHistory checks if a product has any sales history
func (r *ProductRepository) HasSalesHistory(ctx context.Context, id uuid.UUID) (bool, error) {
	query := `SELECT COUNT(*) FROM sale_items WHERE product_id = $1`
	var count int
	err := r.pool.QueryRow(ctx, query, id).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check sales history: %w", err)
	}
	return count > 0, nil
}

// PermanentDelete permanently deletes a product and its related data
func (r *ProductRepository) PermanentDelete(ctx context.Context, id uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Delete inventory items (cascades from foreign key, but explicit is clearer)
	_, err = tx.Exec(ctx, `DELETE FROM inventory_items WHERE product_id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete inventory: %w", err)
	}

	// Delete stock adjustments
	_, err = tx.Exec(ctx, `DELETE FROM stock_adjustments WHERE product_id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete stock adjustments: %w", err)
	}

	// Delete the product
	result, err := tx.Exec(ctx, `DELETE FROM products WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("product not found")
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// SKUExists checks if a SKU is already in use
func (r *ProductRepository) SKUExists(ctx context.Context, sku string, excludeID *uuid.UUID) (bool, error) {
	query := `SELECT COUNT(*) FROM products WHERE sku = $1`
	args := []interface{}{sku}
	if excludeID != nil {
		query += ` AND id != $2`
		args = append(args, *excludeID)
	}

	var count int
	err := r.pool.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check SKU: %w", err)
	}
	return count > 0, nil
}

// BarcodeExists checks if a barcode is already in use
func (r *ProductRepository) BarcodeExists(ctx context.Context, barcode string, excludeID *uuid.UUID) (bool, error) {
	query := `SELECT COUNT(*) FROM products WHERE barcode = $1`
	args := []interface{}{barcode}
	if excludeID != nil {
		query += ` AND id != $2`
		args = append(args, *excludeID)
	}

	var count int
	err := r.pool.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check barcode: %w", err)
	}
	return count > 0, nil
}

// buildWhereClause builds the WHERE clause for product queries
func (r *ProductRepository) buildWhereClause(filter ProductFilter) (string, []interface{}) {
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

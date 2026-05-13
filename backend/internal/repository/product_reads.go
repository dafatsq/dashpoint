package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"dashpoint/backend/internal/models"
)

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
		&product.ID, &product.SKU, &product.Barcode, &product.Name, &product.Description,
		&product.CategoryID, &product.Price, &product.Cost, &product.TaxRate, &product.Unit,
		&product.IsActive, &product.TrackInventory, &product.AllowNegativeStock, &product.ImageURL,
		&product.CreatedAt, &product.UpdatedAt,
		&catID, &catName, &catDesc, &invQty, &invReserved, &invThreshold,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	hydrateProductRelations(product, catID, catName, catDesc, invQty, invReserved, invThreshold)
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

// Lookup searches for a product by barcode or SKU.
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
	countQuery := `SELECT COUNT(*) FROM products p`
	whereClause, args := buildProductWhereClause(filter)
	if whereClause != "" {
		countQuery += " WHERE " + whereClause
	}

	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}

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
	query += fmt.Sprintf(" ORDER BY p.name ASC LIMIT %d OFFSET %d", filter.Limit, filter.Offset)

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
		if err := rows.Scan(
			&product.ID, &product.SKU, &product.Barcode, &product.Name, &product.Description,
			&product.CategoryID, &product.Price, &product.Cost, &product.TaxRate, &product.Unit,
			&product.IsActive, &product.TrackInventory, &product.AllowNegativeStock, &product.ImageURL,
			&product.CreatedAt, &product.UpdatedAt, &catID, &catName, &invQty, &invReserved, &invThreshold,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan product: %w", err)
		}
		hydrateProductRelations(product, catID, catName, nil, invQty, invReserved, invThreshold)
		products = append(products, product)
	}

	return products, total, nil
}

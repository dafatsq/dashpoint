package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"dashpoint/backend/internal/models"
)

// Create creates a new product with optional inventory
func (r *ProductRepository) Create(ctx context.Context, product *models.Product, initialQuantity *decimal.Decimal) error {
	if err := r.ensureActiveProductCategory(ctx, product.CategoryID); err != nil {
		return err
	}

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
		INSERT INTO products (id, sku, barcode, name, description, category_id, price, cost, tax_rate, is_active, image_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err = tx.Exec(ctx, query,
		product.ID, product.SKU, product.Barcode, product.Name, product.Description, product.CategoryID,
		product.Price, product.Cost, product.TaxRate, product.IsActive,
		product.ImageURL, product.CreatedAt, product.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}

	qty := decimal.Zero
	if initialQuantity != nil {
		qty = *initialQuantity
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO inventory_items (product_id, quantity, low_stock_threshold, updated_at)
		VALUES ($1, $2, $3, $4)
	`, product.ID, qty, decimal.Zero, now)
	if err != nil {
		return fmt.Errorf("failed to create inventory record: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

// Update updates a product
func (r *ProductRepository) Update(ctx context.Context, product *models.Product) error {
	if err := r.ensureProductUpdateCategoryAllowed(ctx, product.ID, product.CategoryID); err != nil {
		return err
	}

	product.UpdatedAt = time.Now()
	query := `
		UPDATE products 
		SET sku = $1, barcode = $2, name = $3, description = $4, category_id = $5, 
		    price = $6, cost = $7, tax_rate = $8, is_active = $9,
		    image_url = $10, updated_at = $11
		WHERE id = $12
	`
	result, err := r.pool.Exec(ctx, query,
		product.SKU, product.Barcode, product.Name, product.Description, product.CategoryID,
		product.Price, product.Cost, product.TaxRate, product.IsActive,
		product.ImageURL, product.UpdatedAt, product.ID,
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
	result, err := r.pool.Exec(ctx, `UPDATE products SET is_active = false, updated_at = $1 WHERE id = $2`, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("product not found")
	}
	return nil
}

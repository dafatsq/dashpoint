package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var (
	ErrProductCategoryNotFound = errors.New("product category not found")
	ErrProductCategoryInactive = errors.New("product category inactive")
)

func (r *ProductRepository) ensureActiveProductCategory(ctx context.Context, categoryID *uuid.UUID) error {
	if categoryID == nil {
		return nil
	}

	var isActive bool
	if err := r.pool.QueryRow(ctx, `SELECT is_active FROM categories WHERE id = $1`, *categoryID).Scan(&isActive); err != nil {
		if err == pgx.ErrNoRows {
			return ErrProductCategoryNotFound
		}
		return fmt.Errorf("failed to validate product category: %w", err)
	}
	if !isActive {
		return ErrProductCategoryInactive
	}
	return nil
}

func (r *ProductRepository) ensureProductUpdateCategoryAllowed(ctx context.Context, productID uuid.UUID, categoryID *uuid.UUID) error {
	if categoryID == nil {
		return nil
	}

	var isActive bool
	if err := r.pool.QueryRow(ctx, `SELECT is_active FROM categories WHERE id = $1`, *categoryID).Scan(&isActive); err != nil {
		if err == pgx.ErrNoRows {
			return ErrProductCategoryNotFound
		}
		return fmt.Errorf("failed to validate product category: %w", err)
	}
	if isActive {
		return nil
	}

	var currentCategoryID *uuid.UUID
	if err := r.pool.QueryRow(ctx, `SELECT category_id FROM products WHERE id = $1`, productID).Scan(&currentCategoryID); err != nil {
		return fmt.Errorf("failed to validate current product category: %w", err)
	}
	if currentCategoryID != nil && *currentCategoryID == *categoryID {
		return nil
	}
	return ErrProductCategoryInactive
}

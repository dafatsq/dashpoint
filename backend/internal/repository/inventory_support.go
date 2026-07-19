package repository

import "github.com/jackc/pgx/v5/pgxpool"

// InventoryRepository handles inventory database operations
type InventoryRepository struct {
	pool *pgxpool.Pool
}

// NewInventoryRepository creates a new inventory repository
func NewInventoryRepository(pool *pgxpool.Pool) *InventoryRepository {
	return &InventoryRepository{pool: pool}
}

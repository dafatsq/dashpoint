package repository

import "github.com/jackc/pgx/v5/pgxpool"

// RoleRepository handles role database operations
type RoleRepository struct {
	pool *pgxpool.Pool
}

// NewRoleRepository creates a new role repository
func NewRoleRepository(pool *pgxpool.Pool) *RoleRepository {
	return &RoleRepository{pool: pool}
}

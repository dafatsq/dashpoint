package repository

import "github.com/jackc/pgx/v5/pgxpool"

// UserRepository handles user database operations.
type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository creates a new user repository.
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

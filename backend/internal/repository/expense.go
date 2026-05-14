package repository

import "github.com/jackc/pgx/v5/pgxpool"

// ExpenseRepository handles expense database operations.
type ExpenseRepository struct {
	pool *pgxpool.Pool
}

// NewExpenseRepository creates a new expense repository.
func NewExpenseRepository(pool *pgxpool.Pool) *ExpenseRepository {
	return &ExpenseRepository{pool: pool}
}

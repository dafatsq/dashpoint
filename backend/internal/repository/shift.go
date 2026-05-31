package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ShiftRepository handles shift database operations.
type ShiftRepository struct {
	pool *pgxpool.Pool
}

// ShiftFilter contains shift list filters.
type ShiftFilter struct {
	OpenedByID *uuid.UUID
	StartDate  *time.Time
	EndDate    *time.Time
	Limit      int
	Offset     int
}

// NewShiftRepository creates a new shift repository.
func NewShiftRepository(pool *pgxpool.Pool) *ShiftRepository {
	return &ShiftRepository{pool: pool}
}

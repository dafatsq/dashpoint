package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AuditRepository handles audit log database operations.
type AuditRepository struct {
	pool *pgxpool.Pool
}

// NewAuditRepository creates a new audit repository.
func NewAuditRepository(pool *pgxpool.Pool) *AuditRepository {
	return &AuditRepository{pool: pool}
}

// AuditFilter represents filters for querying audit logs.
type AuditFilter struct {
	UserID        *uuid.UUID
	Action        *string
	EntityType    *string
	EntityID      *string
	Status        *string
	StartDate     *time.Time
	EndDate       *time.Time
	Search        *string
	Limit         int
	Offset        int
	SortBy        string
	SortDirection string
}

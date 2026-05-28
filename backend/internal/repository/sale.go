package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"dashpoint/backend/internal/models"
)

// SaleRepository handles sale database operations.
type SaleRepository struct {
	pool          *pgxpool.Pool
	inventoryRepo *InventoryRepository
}

// NewSaleRepository creates a new sale repository.
func NewSaleRepository(pool *pgxpool.Pool, inventoryRepo *InventoryRepository) *SaleRepository {
	return &SaleRepository{
		pool:          pool,
		inventoryRepo: inventoryRepo,
	}
}

// CreateSaleRequest contains all data needed to create a sale.
type CreateSaleRequest struct {
	Items          []CreateSaleItemRequest
	Payments       []CreatePaymentRequest
	EmployeeID     uuid.UUID
	ShiftID        *uuid.UUID
	DiscountType   *string
	DiscountValue  *decimal.Decimal
	DiscountReason *string
	Notes          *string
}

// CreateSaleItemRequest contains data for a sale item.
type CreateSaleItemRequest struct {
	ProductID      uuid.UUID
	Quantity       decimal.Decimal
	UnitPrice      decimal.Decimal
	DiscountType   *string
	DiscountValue  decimal.Decimal
	DiscountAmount decimal.Decimal
}

// CreatePaymentRequest contains data for a payment.
type CreatePaymentRequest struct {
	PaymentMethod  models.PaymentMethod
	Amount         decimal.Decimal
	AmountTendered *decimal.Decimal
	ChangeGiven    *decimal.Decimal
	CardType       *string
	CardLastFour   *string
	ReferenceNo    *string
	BankName       *string
	AccountNo      *string
	VoucherCode    *string
	Notes          *string
}

// SaleFilter contains sales listing filters.
type SaleFilter struct {
	EmployeeID    *uuid.UUID
	ShiftID       *uuid.UUID
	Status        *string
	StartDate     *time.Time
	EndDate       *time.Time
	InvoiceSearch *string
	Limit         int
	Offset        int
}

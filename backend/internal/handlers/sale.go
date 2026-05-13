package handlers

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"dashpoint/backend/internal/models"
	"dashpoint/backend/internal/repository"
)

type saleStore interface {
	Create(context.Context, *repository.CreateSaleRequest) (*models.Sale, error)
	GetByID(context.Context, uuid.UUID) (*models.Sale, error)
	GetByInvoiceNo(context.Context, string) (*models.Sale, error)
	List(context.Context, *repository.SaleFilter) ([]models.Sale, int, error)
	VoidSale(context.Context, uuid.UUID, uuid.UUID, string) error
	GetDailySummary(context.Context, time.Time) (map[string]interface{}, error)
}

type shiftLookupStore interface {
	GetOpenShiftByEmployee(context.Context, uuid.UUID) (*models.Shift, error)
}

// SaleHandler handles sale endpoints.
type SaleHandler struct {
	saleRepo  saleStore
	shiftRepo shiftLookupStore
}

// NewSaleHandler creates a new sale handler.
func NewSaleHandler(saleRepo saleStore, shiftRepo shiftLookupStore) *SaleHandler {
	return &SaleHandler{
		saleRepo:  saleRepo,
		shiftRepo: shiftRepo,
	}
}

// SaleItemRequest represents a sale item in the request.
type SaleItemRequest struct {
	ProductID      string  `json:"product_id"`
	Quantity       string  `json:"quantity"`
	UnitPrice      string  `json:"unit_price"`
	DiscountType   *string `json:"discount_type"`
	DiscountValue  string  `json:"discount_value"`
	DiscountAmount string  `json:"discount_amount"`
}

// PaymentRequest represents a payment in the request.
type PaymentRequest struct {
	PaymentMethod  string  `json:"payment_method"`
	Amount         string  `json:"amount"`
	AmountTendered *string `json:"amount_tendered"`
	ChangeGiven    *string `json:"change_given"`
	CardType       *string `json:"card_type"`
	CardLastFour   *string `json:"card_last_four"`
	ReferenceNo    *string `json:"reference_no"`
	BankName       *string `json:"bank_name"`
	AccountNo      *string `json:"account_no"`
	VoucherCode    *string `json:"voucher_code"`
	Notes          *string `json:"notes"`
}

// CreateSaleRequest represents the request to create a sale.
type CreateSaleRequest struct {
	Items          []SaleItemRequest `json:"items"`
	Payments       []PaymentRequest  `json:"payments"`
	CustomerName   *string           `json:"customer_name"`
	CustomerPhone  *string           `json:"customer_phone"`
	DiscountType   *string           `json:"discount_type"`
	DiscountValue  *string           `json:"discount_value"`
	DiscountReason *string           `json:"discount_reason"`
	Notes          *string           `json:"notes"`
}

// VoidSaleRequest represents the request to void a sale.
type VoidSaleRequest struct {
	Reason string `json:"reason"`
}

type saleCreateInput struct {
	items          []repository.CreateSaleItemRequest
	payments       []repository.CreatePaymentRequest
	customerName   *string
	customerPhone  *string
	discountType   *string
	discountValue  *decimal.Decimal
	discountReason *string
	notes          *string
}

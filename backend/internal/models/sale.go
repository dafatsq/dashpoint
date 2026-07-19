package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// SaleStatus represents the status of a sale.
type SaleStatus string

const (
	SaleStatusDraft     SaleStatus = "draft"
	SaleStatusCompleted SaleStatus = "completed"
	SaleStatusVoided    SaleStatus = "voided"
)

// PaymentStatus represents the payment status of a sale.
type PaymentStatus string

const (
	PaymentStatusPending  PaymentStatus = "pending"
	PaymentStatusPartial  PaymentStatus = "partial"
	PaymentStatusPaid     PaymentStatus = "paid"
	PaymentStatusRefunded PaymentStatus = "refunded"
	PaymentStatusVoided   PaymentStatus = "voided"
)

// Sale represents a sales transaction.
type Sale struct {
	ID        uuid.UUID `json:"id"`
	InvoiceNo string    `json:"invoice_no"`

	Subtotal       decimal.Decimal `json:"subtotal"`
	TaxAmount      decimal.Decimal `json:"tax_amount"`
	DiscountAmount decimal.Decimal `json:"discount_amount"`
	TotalAmount    decimal.Decimal `json:"total_amount"`
	ItemCount      int             `json:"item_count"`

	PaymentStatus PaymentStatus   `json:"payment_status"`
	AmountPaid    decimal.Decimal `json:"amount_paid"`
	ChangeAmount  decimal.Decimal `json:"change_amount"`

	DiscountType   *string          `json:"discount_type,omitempty"`
	DiscountValue  *decimal.Decimal `json:"discount_value,omitempty"`
	DiscountReason *string          `json:"discount_reason,omitempty"`

	EmployeeID uuid.UUID  `json:"employee_id"`
	ShiftID    *uuid.UUID `json:"shift_id,omitempty"`

	Status     SaleStatus `json:"status"`
	VoidedAt   *time.Time `json:"voided_at,omitempty"`
	VoidedBy   *uuid.UUID `json:"voided_by,omitempty"`
	VoidReason *string    `json:"void_reason,omitempty"`

	Notes *string `json:"notes,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Items        []SaleItem `json:"items,omitempty"`
	Payments     []Payment  `json:"payments,omitempty"`
	EmployeeName *string    `json:"employee_name,omitempty"`
}

// SaleItem represents a line item in a sale.
type SaleItem struct {
	ID     uuid.UUID `json:"id"`
	SaleID uuid.UUID `json:"sale_id"`

	ProductID      uuid.UUID `json:"product_id"`
	ProductName    string    `json:"product_name"`
	ProductSKU     *string   `json:"product_sku,omitempty"`
	ProductBarcode *string   `json:"product_barcode,omitempty"`

	Quantity  decimal.Decimal `json:"quantity"`
	UnitPrice decimal.Decimal `json:"unit_price"`
	CostPrice decimal.Decimal `json:"cost_price"`

	DiscountType   *string         `json:"discount_type,omitempty"`
	DiscountValue  decimal.Decimal `json:"discount_value"`
	DiscountAmount decimal.Decimal `json:"discount_amount"`

	TaxRate   decimal.Decimal `json:"tax_rate"`
	TaxAmount decimal.Decimal `json:"tax_amount"`
	Subtotal  decimal.Decimal `json:"subtotal"`
	Total     decimal.Decimal `json:"total"`

	CreatedAt time.Time `json:"created_at"`
}

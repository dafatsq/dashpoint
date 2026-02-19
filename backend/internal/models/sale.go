package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ShiftStatus represents the status of a shift
type ShiftStatus string

const (
	ShiftStatusOpen      ShiftStatus = "open"
	ShiftStatusClosed    ShiftStatus = "closed"
	ShiftStatusSuspended ShiftStatus = "suspended"
)

// Shift represents a cashier shift
type Shift struct {
	ID         uuid.UUID   `json:"id"`
	EmployeeID uuid.UUID   `json:"employee_id"`
	StartedAt  time.Time   `json:"started_at"`
	EndedAt    *time.Time  `json:"ended_at,omitempty"`
	Status     ShiftStatus `json:"status"`

	// Cash drawer
	OpeningCash    decimal.Decimal  `json:"opening_cash"`
	ClosingCash    *decimal.Decimal `json:"closing_cash,omitempty"`
	ExpectedCash   *decimal.Decimal `json:"expected_cash,omitempty"`
	CashDifference *decimal.Decimal `json:"cash_difference,omitempty"`

	// Summary
	TotalSales       decimal.Decimal `json:"total_sales"`
	TotalCashSales   decimal.Decimal `json:"total_cash_sales"`
	TotalRefunds     decimal.Decimal `json:"total_refunds"`
	TransactionCount int             `json:"transaction_count"`
	RefundCount      int             `json:"refund_count"`

	Notes *string `json:"notes,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Joined fields
	EmployeeName *string `json:"employee_name,omitempty"`
}

// SaleStatus represents the status of a sale
type SaleStatus string

const (
	SaleStatusDraft     SaleStatus = "draft"
	SaleStatusCompleted SaleStatus = "completed"
	SaleStatusVoided    SaleStatus = "voided"
	SaleStatusRefunded  SaleStatus = "refunded"
)

// PaymentStatus represents the payment status of a sale
type PaymentStatus string

const (
	PaymentStatusPending  PaymentStatus = "pending"
	PaymentStatusPartial  PaymentStatus = "partial"
	PaymentStatusPaid     PaymentStatus = "paid"
	PaymentStatusRefunded PaymentStatus = "refunded"
	PaymentStatusVoided   PaymentStatus = "voided"
)

// Sale represents a sales transaction
type Sale struct {
	ID        uuid.UUID `json:"id"`
	InvoiceNo string    `json:"invoice_no"`

	// Totals
	Subtotal       decimal.Decimal `json:"subtotal"`
	TaxAmount      decimal.Decimal `json:"tax_amount"`
	DiscountAmount decimal.Decimal `json:"discount_amount"`
	TotalAmount    decimal.Decimal `json:"total_amount"`
	ItemCount      int             `json:"item_count"`

	// Payment
	PaymentStatus PaymentStatus   `json:"payment_status"`
	AmountPaid    decimal.Decimal `json:"amount_paid"`
	ChangeAmount  decimal.Decimal `json:"change_amount"`

	// Discount
	DiscountType   *string          `json:"discount_type,omitempty"`
	DiscountValue  *decimal.Decimal `json:"discount_value,omitempty"`
	DiscountReason *string          `json:"discount_reason,omitempty"`

	// References
	EmployeeID    uuid.UUID  `json:"employee_id"`
	ShiftID       *uuid.UUID `json:"shift_id,omitempty"`
	CustomerName  *string    `json:"customer_name,omitempty"`
	CustomerPhone *string    `json:"customer_phone,omitempty"`

	// Status
	Status     SaleStatus `json:"status"`
	VoidedAt   *time.Time `json:"voided_at,omitempty"`
	VoidedBy   *uuid.UUID `json:"voided_by,omitempty"`
	VoidReason *string    `json:"void_reason,omitempty"`

	Notes *string `json:"notes,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Related data
	Items        []SaleItem `json:"items,omitempty"`
	Payments     []Payment  `json:"payments,omitempty"`
	EmployeeName *string    `json:"employee_name,omitempty"`
}

// SaleItem represents a line item in a sale
type SaleItem struct {
	ID     uuid.UUID `json:"id"`
	SaleID uuid.UUID `json:"sale_id"`

	// Product reference
	ProductID      uuid.UUID `json:"product_id"`
	ProductName    string    `json:"product_name"`
	ProductSKU     *string   `json:"product_sku,omitempty"`
	ProductBarcode *string   `json:"product_barcode,omitempty"`

	// Quantities and pricing
	Quantity  decimal.Decimal `json:"quantity"`
	UnitPrice decimal.Decimal `json:"unit_price"`
	CostPrice decimal.Decimal `json:"cost_price"`

	// Discount
	DiscountType   *string         `json:"discount_type,omitempty"`
	DiscountValue  decimal.Decimal `json:"discount_value"`
	DiscountAmount decimal.Decimal `json:"discount_amount"`

	// Tax
	TaxRate   decimal.Decimal `json:"tax_rate"`
	TaxAmount decimal.Decimal `json:"tax_amount"`

	// Totals
	Subtotal decimal.Decimal `json:"subtotal"`
	Total    decimal.Decimal `json:"total"`

	// Refund tracking
	IsRefunded       bool            `json:"is_refunded"`
	RefundedQuantity decimal.Decimal `json:"refunded_quantity"`

	CreatedAt time.Time `json:"created_at"`
}

// PaymentMethod represents payment method types
type PaymentMethod string

const (
	PaymentMethodCash     PaymentMethod = "cash"
	PaymentMethodCard     PaymentMethod = "card"
	PaymentMethodTransfer PaymentMethod = "transfer"
	PaymentMethodQRIS     PaymentMethod = "qris"
	PaymentMethodCredit   PaymentMethod = "credit"
	PaymentMethodVoucher  PaymentMethod = "voucher"
	PaymentMethodOther    PaymentMethod = "other"
)

// PaymentRecordStatus represents payment record status
type PaymentRecordStatus string

const (
	PaymentRecordPending   PaymentRecordStatus = "pending"
	PaymentRecordCompleted PaymentRecordStatus = "completed"
	PaymentRecordFailed    PaymentRecordStatus = "failed"
	PaymentRecordRefunded  PaymentRecordStatus = "refunded"
)

// Payment represents a payment record
type Payment struct {
	ID     uuid.UUID `json:"id"`
	SaleID uuid.UUID `json:"sale_id"`

	// Payment details
	PaymentMethod  PaymentMethod    `json:"payment_method"`
	Amount         decimal.Decimal  `json:"amount"`
	AmountTendered *decimal.Decimal `json:"amount_tendered,omitempty"`
	ChangeGiven    *decimal.Decimal `json:"change_given,omitempty"`

	// Card details
	CardType     *string `json:"card_type,omitempty"`
	CardLastFour *string `json:"card_last_four,omitempty"`
	ReferenceNo  *string `json:"reference_no,omitempty"`

	// Transfer/QRIS
	BankName  *string `json:"bank_name,omitempty"`
	AccountNo *string `json:"account_no,omitempty"`

	// Voucher
	VoucherCode *string `json:"voucher_code,omitempty"`

	Status      PaymentRecordStatus `json:"status"`
	Notes       *string             `json:"notes,omitempty"`
	ProcessedBy *uuid.UUID          `json:"processed_by,omitempty"`

	CreatedAt time.Time `json:"created_at"`
}

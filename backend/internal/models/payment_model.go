package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// PaymentMethod represents payment method types.
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

// PaymentRecordStatus represents payment record status.
type PaymentRecordStatus string

const (
	PaymentRecordPending   PaymentRecordStatus = "pending"
	PaymentRecordCompleted PaymentRecordStatus = "completed"
	PaymentRecordFailed    PaymentRecordStatus = "failed"
	PaymentRecordRefunded  PaymentRecordStatus = "refunded"
)

// Payment represents a payment record.
type Payment struct {
	ID     uuid.UUID `json:"id"`
	SaleID uuid.UUID `json:"sale_id"`

	PaymentMethod  PaymentMethod    `json:"payment_method"`
	Amount         decimal.Decimal  `json:"amount"`
	AmountTendered *decimal.Decimal `json:"amount_tendered,omitempty"`
	ChangeGiven    *decimal.Decimal `json:"change_given,omitempty"`

	ReferenceNo *string `json:"reference_no,omitempty"`

	Status PaymentRecordStatus `json:"status"`

	CreatedAt time.Time `json:"created_at"`
}

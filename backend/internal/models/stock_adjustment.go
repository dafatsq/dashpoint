package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type AdjustmentType string

const (
	AdjustmentInitial    AdjustmentType = "initial"
	AdjustmentPurchase   AdjustmentType = "purchase"
	AdjustmentSale       AdjustmentType = "sale"
	AdjustmentReturn     AdjustmentType = "return"
	AdjustmentAdjustment AdjustmentType = "adjustment"
	AdjustmentDamage     AdjustmentType = "damage"
	AdjustmentLoss       AdjustmentType = "loss"
	AdjustmentTransfer   AdjustmentType = "transfer"
	AdjustmentCount      AdjustmentType = "count"
)

// StockAdjustment represents a stock adjustment record
type StockAdjustment struct {
	ID             uuid.UUID       `json:"id"`
	ProductID      uuid.UUID       `json:"product_id"`
	AdjustmentType AdjustmentType  `json:"adjustment_type"`
	QuantityBefore decimal.Decimal `json:"quantity_before"`
	QuantityChange decimal.Decimal `json:"quantity_change"`
	QuantityAfter  decimal.Decimal `json:"quantity_after"`
	Reason         *string         `json:"reason,omitempty"`
	ReferenceType  *string         `json:"reference_type,omitempty"`
	ReferenceID    *uuid.UUID      `json:"reference_id,omitempty"`
	AdjustedBy     uuid.UUID       `json:"adjusted_by"`
	CreatedAt      time.Time       `json:"created_at"`

	Product        *Product `json:"product,omitempty"`
	AdjustedByUser *User    `json:"adjusted_by_user,omitempty"`
}

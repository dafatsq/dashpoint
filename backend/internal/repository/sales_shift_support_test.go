package repository

import (
	"testing"

	"github.com/shopspring/decimal"

	"dashpoint/backend/internal/models"
)

func TestCalculateSalePaymentStatusPaid(t *testing.T) {
	total := decimal.RequireFromString("12.50")
	payments := []CreatePaymentRequest{{Amount: decimal.RequireFromString("15.00")}}

	amountPaid, status, change := calculateSalePaymentStatus(total, payments)
	if !amountPaid.Equal(decimal.RequireFromString("15.00")) {
		t.Fatalf("unexpected amount_paid: %s", amountPaid)
	}
	if status != models.PaymentStatusPaid {
		t.Fatalf("expected paid status, got %s", status)
	}
	if !change.Equal(decimal.RequireFromString("2.50")) {
		t.Fatalf("unexpected change amount: %s", change)
	}
}

func TestCalculateExpectedCash(t *testing.T) {
	expected := calculateExpectedCash(
		decimal.RequireFromString("100"),
		decimal.RequireFromString("50"),
		decimal.RequireFromString("10"),
		decimal.RequireFromString("20"),
		decimal.RequireFromString("5"),
	)

	if !expected.Equal(decimal.RequireFromString("155")) {
		t.Fatalf("expected 155, got %s", expected)
	}
}

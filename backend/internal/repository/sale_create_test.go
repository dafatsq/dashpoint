package repository

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"dashpoint/backend/internal/models"
)

func TestBuildPreparedSaleItemUsesLiveProductPrice(t *testing.T) {
	product := &models.Product{
		ID:      uuid.New(),
		Name:    "Cola",
		Price:   decimal.RequireFromString("15.00"),
		Cost:    decimal.RequireFromString("8.00"),
		TaxRate: decimal.RequireFromString("10"),
	}
	clientItem := &CreateSaleItemRequest{
		ProductID:      product.ID,
		Quantity:       decimal.RequireFromString("2"),
		UnitPrice:      decimal.RequireFromString("15.00"),
		DiscountAmount: decimal.RequireFromString("1.00"),
	}

	prepared := buildPreparedSaleItem(product, clientItem, decimal.RequireFromString("10"), time.Unix(100, 0))

	if !prepared.SaleItem.UnitPrice.Equal(product.Price) {
		t.Fatalf("expected live product price %s, got %s", product.Price, prepared.SaleItem.UnitPrice)
	}
	if !prepared.SaleItem.Subtotal.Equal(decimal.RequireFromString("30.00")) {
		t.Fatalf("expected subtotal 30.00, got %s", prepared.SaleItem.Subtotal)
	}
	if !prepared.SaleItem.TaxAmount.Equal(decimal.RequireFromString("3.00")) {
		t.Fatalf("expected tax amount 3.00, got %s", prepared.SaleItem.TaxAmount)
	}
	if !prepared.SaleItem.Total.Equal(decimal.RequireFromString("32.00")) {
		t.Fatalf("expected total 32.00, got %s", prepared.SaleItem.Total)
	}
}

func TestValidateSaleItemUnitPriceRejectsStaleCartPrice(t *testing.T) {
	product := &models.Product{
		Name:  "Cola",
		Price: decimal.RequireFromString("4500"),
	}
	item := &CreateSaleItemRequest{
		UnitPrice: decimal.RequireFromString("450"),
	}

	err := validateSaleItemUnitPrice(product, item)
	if err == nil {
		t.Fatal("expected stale price to be rejected")
	}
	if err.Error() != "product price changed for Cola: current price is 4500, submitted price is 450" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateSaleItemUnitPriceAcceptsMatchingPriceWithDifferentScale(t *testing.T) {
	product := &models.Product{
		Name:  "Cola",
		Price: decimal.RequireFromString("4500.00"),
	}
	item := &CreateSaleItemRequest{
		UnitPrice: decimal.RequireFromString("4500"),
	}

	if err := validateSaleItemUnitPrice(product, item); err != nil {
		t.Fatalf("expected matching price to be accepted, got %v", err)
	}
}

func TestValidateSaleFinancialIntegrityRejectsPaymentMismatch(t *testing.T) {
	err := validateSaleFinancialIntegrity(
		decimal.RequireFromString("10000"),
		[]CreatePaymentRequest{{
			PaymentMethod: models.PaymentMethodCard,
			Amount:        decimal.RequireFromString("9000"),
		}},
	)

	if err == nil || err.Error() != "payment amount does not match sale total" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateSaleFinancialIntegrityRejectsCashChangeMismatch(t *testing.T) {
	tendered := decimal.RequireFromString("15000")
	change := decimal.RequireFromString("4000")
	err := validateSaleFinancialIntegrity(
		decimal.RequireFromString("10000"),
		[]CreatePaymentRequest{{
			PaymentMethod:  models.PaymentMethodCash,
			Amount:         decimal.RequireFromString("10000"),
			AmountTendered: &tendered,
			ChangeGiven:    &change,
		}},
	)

	if err == nil || err.Error() != "cash change does not match amount tendered" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateSaleFinancialIntegrityRejectsNonCashTenderFields(t *testing.T) {
	tendered := decimal.RequireFromString("10000")
	err := validateSaleFinancialIntegrity(
		decimal.RequireFromString("10000"),
		[]CreatePaymentRequest{{
			PaymentMethod:  models.PaymentMethodQRIS,
			Amount:         decimal.RequireFromString("10000"),
			AmountTendered: &tendered,
		}},
	)

	if err == nil || err.Error() != "amount tendered and change are only valid for cash payments" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateSaleFinancialIntegrityAcceptsValidCashTender(t *testing.T) {
	tendered := decimal.RequireFromString("15000")
	change := decimal.RequireFromString("5000")
	err := validateSaleFinancialIntegrity(
		decimal.RequireFromString("10000"),
		[]CreatePaymentRequest{{
			PaymentMethod:  models.PaymentMethodCash,
			Amount:         decimal.RequireFromString("10000"),
			AmountTendered: &tendered,
			ChangeGiven:    &change,
		}},
	)

	if err != nil {
		t.Fatalf("expected valid cash tender to pass, got %v", err)
	}
}

func TestValidateSaleFinancialIntegrityRejectsNegativeTotal(t *testing.T) {
	err := validateSaleFinancialIntegrity(
		decimal.RequireFromString("-1"),
		[]CreatePaymentRequest{{
			PaymentMethod: models.PaymentMethodCash,
			Amount:        decimal.RequireFromString("1"),
		}},
	)

	if err == nil || err.Error() != "sale total cannot be negative" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateSaleItemDiscountRejectsOutOfBoundsValues(t *testing.T) {
	product := &models.Product{Name: "Cola"}
	item := &CreateSaleItemRequest{
		DiscountAmount: decimal.RequireFromString("11"),
	}

	err := validateSaleItemDiscount(
		product,
		item,
		decimal.RequireFromString("10"),
		decimal.Zero,
	)

	if err == nil || err.Error() != "discount exceeds item total for Cola" {
		t.Fatalf("unexpected error: %v", err)
	}
}

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

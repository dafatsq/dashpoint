package models

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestProductCalculateTaxAndTotal(t *testing.T) {
	product := &Product{
		Price:   decimal.RequireFromString("12.50"),
		TaxRate: decimal.RequireFromString("10"),
	}

	quantity := decimal.RequireFromString("2")
	if got := product.CalculateTax(quantity); !got.Equal(decimal.RequireFromString("2.50")) {
		t.Fatalf("expected tax 2.50, got %s", got)
	}
	if got := product.CalculateTotal(quantity); !got.Equal(decimal.RequireFromString("27.50")) {
		t.Fatalf("expected total 27.50, got %s", got)
	}
}

func TestInventoryItemAvailableQuantityAndLowStock(t *testing.T) {
	item := &InventoryItem{
		Quantity:          decimal.RequireFromString("10"),
		LowStockThreshold: decimal.RequireFromString("10"),
	}

	if got := item.AvailableQuantity(); !got.Equal(decimal.RequireFromString("10")) {
		t.Fatalf("expected available quantity 10, got %s", got)
	}
	if !item.IsLowStock() {
		t.Fatal("expected item to be low stock at threshold")
	}
}

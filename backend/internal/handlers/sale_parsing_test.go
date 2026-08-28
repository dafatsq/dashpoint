package handlers

import (
	"errors"
	"testing"
)

func TestParseCreateSaleInputRejectsOversizedCart(t *testing.T) {
	req := CreateSaleRequest{Items: make([]SaleItemRequest, maxSaleItems+1)}

	_, err := parseCreateSaleInput(req)

	var apiErr *apiError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected an apiError for a %d-item cart, got %v", maxSaleItems+1, err)
	}
	if apiErr.code != "TOO_MANY_ITEMS" {
		t.Fatalf("expected TOO_MANY_ITEMS, got %s", apiErr.code)
	}
}

func TestParseCreateSaleInputAllowsMaxSizeCart(t *testing.T) {
	req := CreateSaleRequest{
		Items:    make([]SaleItemRequest, maxSaleItems),
		Payments: []PaymentRequest{{Amount: "10.00"}},
	}

	_, err := parseCreateSaleInput(req)

	// A cart at exactly the cap must pass the size gate; any failure after it
	// comes from item parsing, never from TOO_MANY_ITEMS.
	var apiErr *apiError
	if errors.As(err, &apiErr) && apiErr.code == "TOO_MANY_ITEMS" {
		t.Fatalf("cart of exactly %d items must not be rejected by the cap", maxSaleItems)
	}
}

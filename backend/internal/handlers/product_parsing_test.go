package handlers

import (
	"strings"
	"testing"
)

func TestValidateProductTextLengthRejectsOversize(t *testing.T) {
	if err := validateProductTextLength("name", strings.Repeat("x", 256), maxProductNameLen); err == nil {
		t.Fatalf("expected 256-char name to be rejected")
	}
	if err := validateProductTextLength("sku", strings.Repeat("x", 101), maxProductSKULen); err == nil {
		t.Fatalf("expected 101-char sku to be rejected")
	}
	if err := validateProductTextLength("barcode", strings.Repeat("x", 101), maxProductBarcodeLen); err == nil {
		t.Fatalf("expected 101-char barcode to be rejected")
	}
}

func TestValidateProductTextLengthAcceptsAtLimit(t *testing.T) {
	if err := validateProductTextLength("name", strings.Repeat("x", 255), maxProductNameLen); err != nil {
		t.Fatalf("expected 255-char name to pass, got %v", err)
	}
	if err := validateProductTextLength("sku", "", maxProductSKULen); err != nil {
		t.Fatalf("expected empty sku to pass, got %v", err)
	}
}

func TestParseCreateProductInputRejectsOversizedName(t *testing.T) {
	req := CreateProductRequest{
		Name:  strings.Repeat("x", 300),
		Price: "10.00",
	}

	if _, err := parseCreateProductInput(req); err == nil {
		t.Fatalf("expected oversized product name to be rejected before hitting the database")
	}
}

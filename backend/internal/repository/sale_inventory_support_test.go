package repository

import "testing"

func TestSaleInventoryReason(t *testing.T) {
	invoiceNo := "INV-20260601-00001"

	if got := saleInventoryReason(invoiceNo); got != "Sale invoice "+invoiceNo {
		t.Fatalf("expected normalized sale reason, got %q", got)
	}
	if got := voidedSaleInventoryReason(invoiceNo); got != "Voided sale invoice "+invoiceNo {
		t.Fatalf("expected normalized voided sale reason, got %q", got)
	}
}

package repository

import (
	"testing"

	"dashpoint/backend/internal/models"
)

func TestDecodeAuditJSONFieldsIgnoresInvalidJSON(t *testing.T) {
	logEntry := &models.AuditLog{}
	err := decodeAuditJSONFields(logEntry, []byte(`{`), nil, nil)
	if err != nil {
		t.Fatalf("expected invalid JSON to be ignored, got %v", err)
	}
}

func TestDecodeAuditJSONFieldsPopulatesMaps(t *testing.T) {
	logEntry := &models.AuditLog{}
	err := decodeAuditJSONFields(logEntry, []byte(`{"before":"x"}`), []byte(`{"after":"y"}`), []byte(`{"request_id":"r1"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if logEntry.OldValues["before"] != "x" {
		t.Fatalf("unexpected old_values: %#v", logEntry.OldValues)
	}
	if logEntry.NewValues["after"] != "y" {
		t.Fatalf("unexpected new_values: %#v", logEntry.NewValues)
	}
	if logEntry.Metadata["request_id"] != "r1" {
		t.Fatalf("unexpected metadata: %#v", logEntry.Metadata)
	}
}

package repository

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestBuildExpenseListBaseQueryIncludesFilters(t *testing.T) {
	categoryID := uuid.New()
	startDate := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 5, 31, 0, 0, 0, 0, time.UTC)

	query, args, argNum := buildExpenseListBaseQuery(&categoryID, &startDate, &endDate)
	if argNum != 4 {
		t.Fatalf("expected next arg index 4, got %d", argNum)
	}
	if len(args) != 3 {
		t.Fatalf("expected 3 args, got %d", len(args))
	}
	if query == "" {
		t.Fatal("expected non-empty query")
	}
}

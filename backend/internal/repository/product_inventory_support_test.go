package repository

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestBuildProductWhereClause(t *testing.T) {
	categoryID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	active := true

	clause, args := buildProductWhereClause(ProductFilter{
		Search:     "milk",
		CategoryID: &categoryID,
		IsActive:   &active,
		LowStock:   true,
	})

	for _, part := range []string{"p.name ILIKE", "p.category_id", "p.is_active", "i.quantity <= i.low_stock_threshold"} {
		if !strings.Contains(clause, part) {
			t.Fatalf("expected where clause to contain %q, got %q", part, clause)
		}
	}
	if len(args) != 3 {
		t.Fatalf("expected 3 args, got %d", len(args))
	}
}

package repository

import (
	"strings"
	"testing"
)

func TestExpenseSelectQueryIncludesFromClause(t *testing.T) {
	query := expenseSelectQuery(" WHERE e.id = $1")

	if !strings.Contains(query, "FROM expenses e") {
		t.Fatalf("expected expense select query to include FROM clause, got %q", query)
	}

	if !strings.Contains(query, "WHERE e.id = $1") {
		t.Fatalf("expected expense select query to include where clause, got %q", query)
	}
}

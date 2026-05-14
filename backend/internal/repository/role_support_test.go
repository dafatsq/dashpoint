package repository

import (
	"testing"

	"github.com/google/uuid"

	"dashpoint/backend/internal/models"
)

func TestGroupPermissionsByCategory(t *testing.T) {
	permissions := []*models.Permission{
		{ID: uuid.New(), Key: "can_view_users", Category: "users"},
		{ID: uuid.New(), Key: "can_view_sales", Category: "sales"},
		{ID: uuid.New(), Key: "can_edit_users", Category: "users"},
	}

	grouped := groupPermissionsByCategory(permissions)
	if len(grouped["users"]) != 2 {
		t.Fatalf("expected 2 users permissions, got %d", len(grouped["users"]))
	}
	if len(grouped["sales"]) != 1 {
		t.Fatalf("expected 1 sales permission, got %d", len(grouped["sales"]))
	}
}

func TestRoleSortWeight(t *testing.T) {
	if got := roleSortWeight("owner"); got != 1 {
		t.Fatalf("expected owner weight 1, got %d", got)
	}
	if got := roleSortWeight("manager"); got != 2 {
		t.Fatalf("expected manager weight 2, got %d", got)
	}
	if got := roleSortWeight("cashier"); got != 3 {
		t.Fatalf("expected cashier weight 3, got %d", got)
	}
	if got := roleSortWeight("custom"); got != 4 {
		t.Fatalf("expected custom weight 4, got %d", got)
	}
}

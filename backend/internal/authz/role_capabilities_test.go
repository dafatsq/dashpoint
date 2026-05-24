package authz

import "testing"

func TestPermissionsForRoleReturnsCopy(t *testing.T) {
	permissions := PermissionsForRole("manager")
	if len(permissions) == 0 {
		t.Fatal("expected manager permissions")
	}

	permissions[0] = "mutated"

	refetched := PermissionsForRole("manager")
	if refetched[0] == "mutated" {
		t.Fatal("expected defensive copy")
	}
}

func TestHasPermission(t *testing.T) {
	if !HasPermission("manager", "can_view_users") {
		t.Fatal("expected manager to have can_view_users")
	}
	if HasPermission("cashier", "can_view_reports") {
		t.Fatal("did not expect cashier to have can_view_reports")
	}
}

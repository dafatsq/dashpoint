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
	if !HasPermission("manager", "access_users_page") {
		t.Fatal("expected manager to have access_users_page")
	}
	if HasPermission("cashier", "access_reports_page") {
		t.Fatal("did not expect cashier to have access_reports_page")
	}
}

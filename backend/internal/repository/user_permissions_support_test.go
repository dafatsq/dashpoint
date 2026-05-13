package repository

import "testing"

func TestPermissionSetToSlice(t *testing.T) {
	permissionSet := map[string]bool{
		"can_view_users":  true,
		"can_edit_users":  true,
		"can_delete_user": true,
	}

	permissions := permissionSetToSlice(permissionSet)
	if len(permissions) != 3 {
		t.Fatalf("expected 3 permissions, got %d", len(permissions))
	}

	seen := make(map[string]bool, len(permissions))
	for _, permission := range permissions {
		seen[permission] = true
	}

	for expected := range permissionSet {
		if !seen[expected] {
			t.Fatalf("missing permission %q in slice output", expected)
		}
	}
}

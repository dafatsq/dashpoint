package authz

var managerPermissions = []string{
	"can_view_users",
	"can_create_cashier_users",
	"can_edit_cashier_users",
	"can_delete_cashier_users",
	"can_view_products",
	"can_create_product",
	"can_edit_product",
	"can_delete_product",
	"can_view_inventory",
	"can_edit_inventory",
	"can_add_stock",
	"can_remove_stock",
	"can_adjust_stock",
	"can_create_sale",
	"can_view_sales",
	"can_void_sale",
	"can_view_reports",
	"can_export_data",
	"can_view_expenses",
	"can_manage_expenses",
	"can_create_expenses",
	"can_edit_expenses",
	"can_delete_expenses",
	"can_view_audit_logs",
	"can_view_pos",
	"can_start_shift",
	"can_end_shift",
	"can_view_categories",
	"can_manage_categories",
	"can_create_categories",
	"can_edit_categories",
	"can_delete_categories",
}

var cashierPermissions = []string{
	"can_view_pos",
	"can_create_sale",
	"can_view_sales",
	"can_start_shift",
	"can_end_shift",
}

var allPermissions = []string{
	"can_view_users",
	"can_create_user",
	"can_edit_user",
	"can_delete_user",
	"can_manage_permissions",
	"can_create_manager_users",
	"can_edit_manager_users",
	"can_delete_manager_users",
	"can_manage_manager_permissions",
	"can_create_cashier_users",
	"can_edit_cashier_users",
	"can_delete_cashier_users",
	"can_manage_cashier_permissions",
	"can_view_products",
	"can_create_product",
	"can_edit_product",
	"can_delete_product",
	"can_view_inventory",
	"can_edit_inventory",
	"can_add_stock",
	"can_remove_stock",
	"can_adjust_stock",
	"can_create_sale",
	"can_view_sales",
	"can_void_sale",
	"can_view_reports",
	"can_export_data",
	"can_view_expenses",
	"can_manage_expenses",
	"can_create_expenses",
	"can_edit_expenses",
	"can_delete_expenses",
	"can_view_audit_logs",
	"can_view_pos",
	"can_start_shift",
	"can_end_shift",
	"can_view_categories",
	"can_manage_categories",
	"can_create_categories",
	"can_edit_categories",
	"can_delete_categories",
}

var rolePermissions = map[string][]string{
	"owner":   allPermissions,
	"manager": managerPermissions,
	"cashier": cashierPermissions,
}

func PermissionsForRole(roleName string) []string {
	permissions, ok := rolePermissions[roleName]
	if !ok {
		return []string{}
	}

	cloned := make([]string, len(permissions))
	copy(cloned, permissions)
	return cloned
}

func HasPermission(roleName, permission string) bool {
	for _, granted := range PermissionsForRole(roleName) {
		if granted == permission {
			return true
		}
	}
	return false
}

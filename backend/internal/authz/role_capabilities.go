package authz

var managerPermissions = []string{
	"access_pos_page",
	"manage_pos_page",
	"access_products_page",
	"manage_products_page",
	"access_inventory_page",
	"manage_inventory_page",
	"access_sales_page",
	"manage_sales_page",
	"access_reports_page",
	"manage_reports_page",
	"access_expenses_page",
	"manage_expenses_page",
	"access_categories_page",
	"manage_categories_page",
	"access_users_page",
	"manage_users_page",
	"access_shifts_page",
	"manage_shifts_page",
	"access_changes_page",
	"access_audit_page",
}

var cashierPermissions = []string{
	"access_pos_page",
	"manage_pos_page",
	"access_sales_page",
	"access_shifts_page",
	"manage_shifts_page",
}

var allPermissions = []string{
	"access_pos_page",
	"manage_pos_page",
	"access_products_page",
	"manage_products_page",
	"access_inventory_page",
	"manage_inventory_page",
	"access_sales_page",
	"manage_sales_page",
	"access_reports_page",
	"manage_reports_page",
	"access_expenses_page",
	"manage_expenses_page",
	"access_categories_page",
	"manage_categories_page",
	"access_users_page",
	"manage_users_page",
	"access_shifts_page",
	"manage_shifts_page",
	"access_changes_page",
	"access_audit_page",
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

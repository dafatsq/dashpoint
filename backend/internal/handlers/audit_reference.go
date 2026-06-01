package handlers

import "github.com/gofiber/fiber/v2"

// GetActions handles GET /api/v1/logs/actions - returns available action types
func (h *AuditHandler) GetActions(c *fiber.Ctx) error {
	actions := []map[string]string{
		{"action": "auth.login", "description": "User login"},
		{"action": "auth.login_failed", "description": "Failed login attempt"},
		{"action": "auth.logout", "description": "User logout"},
		{"action": "auth.pin_login", "description": "PIN login"},
		{"action": "user.create", "description": "User created"},
		{"action": "user.update", "description": "User updated"},
		{"action": "user.delete", "description": "User deleted"},
		{"action": "user.deactivate", "description": "User deactivated"},
		{"action": "user.password_change", "description": "Password changed"},
		{"action": "user.pin_change", "description": "PIN changed"},
		{"action": "user.permission_change", "description": "Permissions changed"},
		{"action": "product.create", "description": "Product created"},
		{"action": "product.update", "description": "Product updated"},
		{"action": "product.delete", "description": "Product deleted"},
		{"action": "inventory.adjust", "description": "Stock adjusted"},
		{"action": "inventory.count", "description": "Stock count"},
		{"action": "inventory.threshold_update", "description": "Low stock threshold updated"},
		{"action": "category.create", "description": "Category created"},
		{"action": "category.update", "description": "Category updated"},
		{"action": "category.delete", "description": "Category deleted"},
		{"action": "sale.create", "description": "Sale created"},
		{"action": "sale.void", "description": "Sale voided"},
		{"action": "shift.start", "description": "Shift started"},
		{"action": "shift.close", "description": "Shift closed"},
		{"action": "report.export", "description": "Report exported"},
	}

	entityTypes := []map[string]string{
		{"type": "user", "description": "User accounts"},
		{"type": "product", "description": "Products"},
		{"type": "category", "description": "Categories"},
		{"type": "inventory", "description": "Inventory"},
		{"type": "sale", "description": "Sales"},
		{"type": "shift", "description": "Shifts"},
		{"type": "payment", "description": "Payments"},
		{"type": "report", "description": "Reports"},
		{"type": "auth", "description": "Authentication"},
		{"type": "system", "description": "System"},
	}

	return c.JSON(fiber.Map{
		"actions":      actions,
		"entity_types": entityTypes,
	})
}

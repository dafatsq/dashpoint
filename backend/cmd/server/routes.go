package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"dashpoint/backend/internal/middleware"
)

func registerRoutes(app *fiber.App, deps *serverDependencies) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"name":    "DashPoint POS API",
			"version": "1.0.0",
			"status":  "running",
		})
	})

	api := app.Group("/api/v1")
	registerPublicRoutes(api, deps)
	registerProtectedRoutes(api, deps)
}

func registerPublicRoutes(api fiber.Router, deps *serverDependencies) {
	api.Get("/health", deps.healthHandler.Check)
	api.Get("/ping", deps.healthHandler.Ping)

	authGroup := api.Group("/auth")
	authGroup.Post("/login", deps.authHandler.Login)
	authGroup.Post("/pin-login", deps.authHandler.PINLogin)
	authGroup.Post("/refresh", deps.authHandler.Refresh)
	authGroup.Post("/logout", deps.authHandler.Logout)

	api.Get("/events/subscribe", deps.eventsHandler.Subscribe)
}

func registerProtectedRoutes(api fiber.Router, deps *serverDependencies) {
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(deps.jwtManager, deps.userRepo))

	protected.Get("/me", deps.authHandler.Me)
	protected.Get("/roles", deps.roleHandler.ListRoles)
	protected.Get("/roles/:id", deps.roleHandler.GetRole)
	protected.Get("/permissions", deps.roleHandler.ListPermissions)

	registerUserRoutes(protected, deps)
	registerCatalogRoutes(protected, deps)
	registerOperationsRoutes(protected, deps)
	registerReportsRoutes(protected, deps)
	registerExpenseRoutes(protected, deps)
	registerAuditRoutes(protected, deps)
	registerUploadRoutes(protected, deps)
}

func registerUserRoutes(protected fiber.Router, deps *serverDependencies) {
	users := protected.Group("/users")
	users.Get("/", middleware.RequirePermission(deps.permissionChecker, "can_view_users"), deps.userHandler.List)
	users.Get("/:id", middleware.RequirePermission(deps.permissionChecker, "can_view_users"), deps.userHandler.Get)
	users.Post("/", middleware.RequireAnyPermission(deps.permissionChecker, "can_create_user", "can_create_manager_users", "can_create_cashier_users"), deps.userHandler.Create)
	users.Patch("/:id", middleware.RequireAnyPermission(deps.permissionChecker, "can_edit_user", "can_edit_manager_users", "can_edit_cashier_users"), deps.userHandler.Update)
	users.Patch("/:id/password", middleware.RequireAnyPermission(deps.permissionChecker, "can_edit_user", "can_edit_manager_users", "can_edit_cashier_users"), deps.userHandler.UpdatePassword)
	users.Patch("/:id/pin", middleware.RequireAnyPermission(deps.permissionChecker, "can_edit_user", "can_edit_manager_users", "can_edit_cashier_users"), deps.userHandler.UpdatePIN)
	users.Delete("/:id", middleware.RequireAnyPermission(deps.permissionChecker, "can_delete_user", "can_delete_manager_users", "can_delete_cashier_users"), deps.userHandler.Delete)
	users.Delete("/:id/permanent", middleware.RequireAnyPermission(deps.permissionChecker, "can_delete_user", "can_delete_manager_users", "can_delete_cashier_users"), deps.userHandler.PermanentDelete)
	users.Get("/:id/permissions", middleware.RequireAnyPermission(deps.permissionChecker, "can_manage_permissions", "can_manage_manager_permissions", "can_manage_cashier_permissions"), deps.userHandler.GetPermissions)
	users.Patch("/:id/permissions", middleware.RequireAnyPermission(deps.permissionChecker, "can_manage_permissions", "can_manage_manager_permissions", "can_manage_cashier_permissions"), deps.userHandler.SetPermissions)
}

func registerCatalogRoutes(protected fiber.Router, deps *serverDependencies) {
	categories := protected.Group("/categories")
	categories.Get("/", middleware.RequirePermission(deps.permissionChecker, "can_view_categories"), deps.categoryHandler.List)
	categories.Get("/:id", middleware.RequirePermission(deps.permissionChecker, "can_view_categories"), deps.categoryHandler.Get)
	categories.Post("/", middleware.RequirePermission(deps.permissionChecker, "can_create_categories"), deps.categoryHandler.Create)
	categories.Patch("/:id", middleware.RequirePermission(deps.permissionChecker, "can_edit_categories"), deps.categoryHandler.Update)
	categories.Delete("/:id", middleware.RequirePermission(deps.permissionChecker, "can_delete_categories"), deps.categoryHandler.Delete)
	categories.Delete("/:id/permanent", middleware.RequirePermission(deps.permissionChecker, "can_delete_categories"), deps.categoryHandler.PermanentDelete)

	products := protected.Group("/products")
	products.Get("/", deps.productHandler.List)
	products.Get("/lookup", deps.productHandler.Lookup)
	products.Get("/:id", deps.productHandler.Get)
	products.Get("/:id/inventory", deps.productHandler.GetInventory)
	products.Post("/", middleware.RequirePermission(deps.permissionChecker, "can_create_product"), deps.productHandler.Create)
	products.Patch("/:id", middleware.RequirePermission(deps.permissionChecker, "can_edit_product"), deps.productHandler.Update)
	products.Delete("/:id", middleware.RequirePermission(deps.permissionChecker, "can_delete_product"), deps.productHandler.Delete)
	products.Delete("/:id/permanent", middleware.RequirePermission(deps.permissionChecker, "can_delete_product"), deps.productHandler.PermanentDelete)

	inventory := protected.Group("/inventory")
	inventory.Get("/low-stock", deps.productHandler.GetLowStock)
	inventory.Post("/adjust", requireInventoryAdjustmentPermission(deps.permissionChecker), deps.productHandler.AdjustStock)
}

func registerOperationsRoutes(protected fiber.Router, deps *serverDependencies) {
	shifts := protected.Group("/shifts")
	shifts.Get("/current", deps.shiftHandler.GetCurrentShift)
	shifts.Post("/start", deps.shiftHandler.StartShift)
	shifts.Post("/close", deps.shiftHandler.CloseShift)
	shifts.Post("/pay-in", deps.cashDrawerHandler.PayIn)
	shifts.Post("/pay-out", deps.cashDrawerHandler.PayOut)
	shifts.Get("/", deps.shiftHandler.ListShifts)
	shifts.Get("/:id", deps.shiftHandler.GetShift)
	shifts.Get("/:id/operations", deps.cashDrawerHandler.ListOperations)

	sales := protected.Group("/sales")
	sales.Post("/", middleware.RequirePermission(deps.permissionChecker, "can_create_sale"), deps.saleHandler.CreateSale)
	sales.Get("/", middleware.RequirePermission(deps.permissionChecker, "can_view_sales"), deps.saleHandler.ListSales)
	sales.Get("/summary/daily", middleware.RequirePermission(deps.permissionChecker, "can_view_sales"), deps.saleHandler.GetDailySummary)
	sales.Get("/invoice/:invoiceNo", middleware.RequirePermission(deps.permissionChecker, "can_view_sales"), deps.saleHandler.GetSaleByInvoice)
	sales.Get("/:id", middleware.RequirePermission(deps.permissionChecker, "can_view_sales"), deps.saleHandler.GetSale)
	sales.Post("/:id/void", middleware.RequirePermission(deps.permissionChecker, "can_void_sale"), deps.saleHandler.VoidSale)
}

func registerReportsRoutes(protected fiber.Router, deps *serverDependencies) {
	reports := protected.Group("/reports")
	reports.Get("/daily", deps.reportHandler.GetDailySalesReport)
	reports.Get("/sales", deps.reportHandler.GetSalesRangeReport)
	reports.Get("/top-sellers", deps.reportHandler.GetTopSellers)
	reports.Get("/inventory", deps.reportHandler.GetInventoryValuation)
	reports.Get("/cash", deps.reportHandler.GetCashReport)
	reports.Get("/by-employee", middleware.RequirePermission(deps.permissionChecker, "can_view_reports"), deps.reportHandler.GetEmployeeSalesReport)
	reports.Get("/by-category", deps.reportHandler.GetCategorySalesReport)
	reports.Get("/export/sales", middleware.RequirePermission(deps.permissionChecker, "can_export_data"), deps.reportHandler.ExportSalesCSV)
	reports.Get("/export/inventory", middleware.RequirePermission(deps.permissionChecker, "can_export_data"), deps.reportHandler.ExportInventoryCSV)
	reports.Get("/export/top-sellers", middleware.RequirePermission(deps.permissionChecker, "can_export_data"), deps.reportHandler.ExportTopSellersCSV)
	reports.Get("/export/comprehensive", middleware.RequirePermission(deps.permissionChecker, "can_export_data"), deps.reportHandler.ExportComprehensiveReportCSV)
}

func registerExpenseRoutes(protected fiber.Router, deps *serverDependencies) {
	expenses := protected.Group("/expenses")
	expenses.Get("/", middleware.RequirePermission(deps.permissionChecker, "can_view_expenses"), deps.expenseHandler.List)
	expenses.Get("/categories", middleware.RequireAnyPermission(deps.permissionChecker, "can_view_expenses", "can_view_categories"), deps.expenseHandler.ListCategories)
	expenses.Post("/categories", middleware.RequirePermission(deps.permissionChecker, "can_create_categories"), deps.expenseHandler.CreateCategory)
	expenses.Get("/categories/:id", middleware.RequireAnyPermission(deps.permissionChecker, "can_view_expenses", "can_view_categories"), deps.expenseHandler.GetCategory)
	expenses.Patch("/categories/:id", middleware.RequirePermission(deps.permissionChecker, "can_edit_categories"), deps.expenseHandler.UpdateCategory)
	expenses.Delete("/categories/:id", middleware.RequirePermission(deps.permissionChecker, "can_delete_categories"), deps.expenseHandler.DeleteCategory)
	expenses.Delete("/categories/:id/permanent", middleware.RequirePermission(deps.permissionChecker, "can_delete_categories"), deps.expenseHandler.PermanentDeleteCategory)
	expenses.Get("/summary", middleware.RequirePermission(deps.permissionChecker, "can_view_expenses"), deps.expenseHandler.GetSummary)
	expenses.Get("/monthly", middleware.RequirePermission(deps.permissionChecker, "can_view_expenses"), deps.expenseHandler.GetMonthlyTotals)
	expenses.Post("/", middleware.RequirePermission(deps.permissionChecker, "can_create_expenses"), deps.expenseHandler.Create)
	expenses.Get("/:id", middleware.RequirePermission(deps.permissionChecker, "can_view_expenses"), deps.expenseHandler.Get)
	expenses.Patch("/:id", middleware.RequirePermission(deps.permissionChecker, "can_edit_expenses"), deps.expenseHandler.Update)
	expenses.Delete("/:id", middleware.RequirePermission(deps.permissionChecker, "can_delete_expenses"), deps.expenseHandler.Delete)
}

func registerAuditRoutes(protected fiber.Router, deps *serverDependencies) {
	dashboard := protected.Group("/dashboard")
	dashboard.Get("/changes", deps.auditHandler.List)

	logs := protected.Group("/logs")
	logs.Get("/", middleware.RequirePermission(deps.permissionChecker, "can_view_audit_logs"), deps.auditHandler.List)
	logs.Get("/actions", middleware.RequirePermission(deps.permissionChecker, "can_view_audit_logs"), deps.auditHandler.GetActions)
	logs.Get("/summary", middleware.RequirePermission(deps.permissionChecker, "can_view_audit_logs"), deps.auditHandler.GetSummary)
	logs.Get("/entity/:type/:id", middleware.RequirePermission(deps.permissionChecker, "can_view_audit_logs"), deps.auditHandler.GetEntityHistory)
	logs.Get("/user/:id", middleware.RequirePermission(deps.permissionChecker, "can_view_audit_logs"), deps.auditHandler.GetUserActivity)
	logs.Get("/:id", middleware.RequirePermission(deps.permissionChecker, "can_view_audit_logs"), deps.auditHandler.Get)

	auditGroup := protected.Group("/audit")
	auditGroup.Get("/", middleware.RequirePermission(deps.permissionChecker, "can_view_audit_logs"), deps.auditHandler.List)
	auditGroup.Get("/:id", middleware.RequirePermission(deps.permissionChecker, "can_view_audit_logs"), deps.auditHandler.Get)
}

func registerUploadRoutes(protected fiber.Router, deps *serverDependencies) {
	upload := protected.Group("/upload")
	upload.Post("/image", deps.uploadHandler.UploadImage)
	upload.Delete("/image/:filename", deps.uploadHandler.DeleteImage)
}

func requireInventoryAdjustmentPermission(checker middleware.PermissionChecker) fiber.Handler {
	type inventoryAdjustmentRequest struct {
		AdjustmentType string `json:"adjustment_type"`
	}

	return func(c *fiber.Ctx) error {
		userID := middleware.GetUserID(c)
		if userID == uuid.Nil {
			return middleware.JSONError(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		}

		var req inventoryAdjustmentRequest
		if err := c.BodyParser(&req); err != nil {
			return middleware.JSONError(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		}

		requiredPermission, ok := adjustmentPermission(req.AdjustmentType)
		if !ok {
			return middleware.JSONError(c, fiber.StatusBadRequest, "INVALID_ADJUSTMENT_TYPE", "Invalid adjustment type")
		}

		hasPermission, err := checker(c, userID, requiredPermission)
		if err != nil {
			return middleware.JSONError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "Failed to check permissions")
		}
		if !hasPermission {
			return middleware.JSONError(c, fiber.StatusForbidden, "FORBIDDEN", "You do not have the required permission: "+requiredPermission)
		}

		return c.Next()
	}
}

func adjustmentPermission(adjustmentType string) (string, bool) {
	switch adjustmentType {
	case "purchase":
		return "can_add_stock", true
	case "damage", "loss":
		return "can_remove_stock", true
	case "adjustment", "count":
		return "can_adjust_stock", true
	default:
		return "", false
	}
}

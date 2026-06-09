package main

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"dashpoint/backend/internal/middleware"
)

var categoryMetadataPermissions = []string{
	"access_categories_page",
	"access_products_page",
	"manage_products_page",
	"access_pos_page",
}
var productReadPermissions = []string{
	"access_products_page",
	"access_inventory_page",
	"access_pos_page",
}
var currentShiftReadPermissions = []string{
	"access_pos_page",
	"access_shifts_page",
	"manage_shifts_page",
}

var (
	expenseCategoryReadPermissions = []string{
		"access_expenses_page",
		"access_categories_page",
	}
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
	authGroup.Post("/login", middleware.AuthRateLimit(), deps.authHandler.Login)
	authGroup.Post("/pin-login", middleware.AuthRateLimit(), deps.authHandler.PINLogin)
	authGroup.Post("/refresh", middleware.AuthRateLimit(), deps.authHandler.Refresh)
	authGroup.Post("/logout", deps.authHandler.Logout)

	api.Get("/events/subscribe", deps.eventsHandler.Subscribe)
}

func registerProtectedRoutes(api fiber.Router, deps *serverDependencies) {
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(deps.jwtManager, deps.userRepo))

	protected.Get("/me", deps.authHandler.Me)
	protected.Get("/roles", middleware.RequirePermission(deps.permissionChecker, "access_users_page"), deps.roleHandler.ListRoles)
	protected.Patch("/roles/:id/permissions", middleware.RequireRole("owner"), deps.roleHandler.UpdateRolePermissions)
	protected.Get("/roles/:id", middleware.RequirePermission(deps.permissionChecker, "access_users_page"), deps.roleHandler.GetRole)

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
	users.Get("/basic", deps.userHandler.ListBasic)
	users.Get("/", middleware.RequirePermission(deps.permissionChecker, "access_users_page"), deps.userHandler.List)
	users.Get("/:id", middleware.RequirePermission(deps.permissionChecker, "access_users_page"), deps.userHandler.Get)
	users.Post("/", middleware.RequirePermission(deps.permissionChecker, "manage_users_page"), deps.userHandler.Create)
	users.Patch("/:id", middleware.RequirePermissionOrSelfParam(deps.permissionChecker, "id", "manage_users_page"), deps.userHandler.Update)
	users.Patch("/:id/password", middleware.RequirePermissionOrSelfParam(deps.permissionChecker, "id", "manage_users_page"), deps.userHandler.UpdatePassword)
	users.Patch("/:id/pin", middleware.RequirePermissionOrSelfParam(deps.permissionChecker, "id", "manage_users_page"), deps.userHandler.UpdatePIN)
	users.Delete("/:id", middleware.RequirePermission(deps.permissionChecker, "manage_users_page"), deps.userHandler.Delete)
	users.Delete("/:id/permanent", middleware.RequirePermission(deps.permissionChecker, "manage_users_page"), deps.userHandler.PermanentDelete)
}

func registerCatalogRoutes(protected fiber.Router, deps *serverDependencies) {
	registerCategoryRoutes(protected, deps)
	registerProductRoutes(protected, deps)
	registerInventoryRoutes(protected, deps)
}

func registerCategoryRoutes(protected fiber.Router, deps *serverDependencies) {
	categories := protected.Group("/categories")
	categories.Get(
		"/",
		middleware.RequireAnyPermission(deps.permissionChecker, categoryMetadataPermissions...),
		deps.categoryHandler.List,
	)
	categories.Get(
		"/:id",
		middleware.RequireAnyPermission(deps.permissionChecker, categoryMetadataPermissions...),
		deps.categoryHandler.Get,
	)
	categories.Post("/", middleware.RequirePermission(deps.permissionChecker, "manage_categories_page"), deps.categoryHandler.Create)
	categories.Patch("/:id", middleware.RequirePermission(deps.permissionChecker, "manage_categories_page"), deps.categoryHandler.Update)
	categories.Delete("/:id", middleware.RequirePermission(deps.permissionChecker, "manage_categories_page"), deps.categoryHandler.Delete)
	categories.Delete("/:id/permanent", middleware.RequirePermission(deps.permissionChecker, "manage_categories_page"), deps.categoryHandler.PermanentDelete)
}

func registerProductRoutes(protected fiber.Router, deps *serverDependencies) {
	products := protected.Group("/products")
	products.Get("/", middleware.RequireAnyPermission(deps.permissionChecker, productReadPermissions...), deps.productHandler.List)
	products.Get("/lookup", middleware.RequirePermission(deps.permissionChecker, "access_pos_page"), deps.productHandler.Lookup)
	products.Get("/:id", middleware.RequireAnyPermission(deps.permissionChecker, productReadPermissions...), deps.productHandler.Get)
	products.Get("/:id/inventory", middleware.RequireAnyPermission(deps.permissionChecker, "access_inventory_page", "manage_inventory_page"), deps.productHandler.GetInventory)
	products.Patch("/:id/inventory", middleware.RequirePermission(deps.permissionChecker, "manage_inventory_page"), deps.productHandler.UpdateInventoryThreshold)
	products.Post("/", middleware.RequirePermission(deps.permissionChecker, "manage_products_page"), deps.productHandler.Create)
	products.Patch("/:id", middleware.RequirePermission(deps.permissionChecker, "manage_products_page"), deps.productHandler.Update)
	products.Delete("/:id", middleware.RequirePermission(deps.permissionChecker, "manage_products_page"), deps.productHandler.Delete)
	products.Delete("/:id/permanent", middleware.RequirePermission(deps.permissionChecker, "manage_products_page"), deps.productHandler.PermanentDelete)
}

func registerInventoryRoutes(protected fiber.Router, deps *serverDependencies) {
	inventory := protected.Group("/inventory")
	inventory.Get("/low-stock", middleware.RequirePermission(deps.permissionChecker, "access_inventory_page"), deps.productHandler.GetLowStock)
	inventory.Post("/adjust", requireInventoryAdjustmentPermission(deps.permissionChecker), deps.productHandler.AdjustStock)
}

func registerOperationsRoutes(protected fiber.Router, deps *serverDependencies) {
	registerShiftRoutes(protected, deps)
	registerSalesRoutes(protected, deps)
}

func registerShiftRoutes(protected fiber.Router, deps *serverDependencies) {
	shifts := protected.Group("/shifts")
	shifts.Get("/current", middleware.RequireAnyPermission(deps.permissionChecker, currentShiftReadPermissions...), deps.shiftHandler.GetCurrentShift)
	shifts.Post("/start", middleware.RequirePermission(deps.permissionChecker, "manage_shifts_page"), deps.shiftHandler.StartShift)
	shifts.Post("/close", middleware.RequirePermission(deps.permissionChecker, "manage_shifts_page"), deps.shiftHandler.CloseShift)
	shifts.Post("/pay-in", middleware.RequirePermission(deps.permissionChecker, "manage_shifts_page"), deps.cashDrawerHandler.PayIn)
	shifts.Post("/pay-out", middleware.RequirePermission(deps.permissionChecker, "manage_shifts_page"), deps.cashDrawerHandler.PayOut)
	shifts.Get("/", middleware.RequirePermission(deps.permissionChecker, "access_shifts_page"), deps.shiftHandler.ListShifts)
	shifts.Get("/:id", middleware.RequirePermission(deps.permissionChecker, "access_shifts_page"), deps.shiftHandler.GetShift)
	shifts.Get("/:id/operations", middleware.RequirePermission(deps.permissionChecker, "access_shifts_page"), deps.cashDrawerHandler.ListOperations)
}

func registerSalesRoutes(protected fiber.Router, deps *serverDependencies) {
	sales := protected.Group("/sales")
	sales.Post("/", middleware.RequirePermission(deps.permissionChecker, "manage_pos_page"), deps.saleHandler.CreateSale)
	sales.Post("/validate", middleware.RequirePermission(deps.permissionChecker, "manage_pos_page"), deps.saleHandler.ValidateCart)
	sales.Get("/", middleware.RequirePermission(deps.permissionChecker, "access_sales_page"), deps.saleHandler.ListSales)
	sales.Get("/summary/daily", middleware.RequirePermission(deps.permissionChecker, "access_sales_page"), deps.saleHandler.GetDailySummary)
	sales.Get("/invoice/:invoiceNo", middleware.RequirePermission(deps.permissionChecker, "access_sales_page"), deps.saleHandler.GetSaleByInvoice)
	sales.Get("/:id", middleware.RequirePermission(deps.permissionChecker, "access_sales_page"), deps.saleHandler.GetSale)
	sales.Post("/:id/void", middleware.RequirePermission(deps.permissionChecker, "manage_sales_page"), deps.saleHandler.VoidSale)
}

func registerReportsRoutes(protected fiber.Router, deps *serverDependencies) {
	reports := protected.Group("/reports")
	reports.Get("/daily", middleware.RequirePermission(deps.permissionChecker, "access_reports_page"), deps.reportHandler.GetDailySalesReport)
	reports.Get("/sales", middleware.RequirePermission(deps.permissionChecker, "access_reports_page"), deps.reportHandler.GetSalesRangeReport)
	reports.Get("/top-sellers", middleware.RequirePermission(deps.permissionChecker, "access_reports_page"), deps.reportHandler.GetTopSellers)
	reports.Get("/inventory", middleware.RequirePermission(deps.permissionChecker, "access_reports_page"), deps.reportHandler.GetInventoryValuation)
	reports.Get("/cash", middleware.RequirePermission(deps.permissionChecker, "access_reports_page"), deps.reportHandler.GetCashReport)
	reports.Get("/by-employee", middleware.RequirePermission(deps.permissionChecker, "access_reports_page"), deps.reportHandler.GetEmployeeSalesReport)
	reports.Get("/by-category", middleware.RequirePermission(deps.permissionChecker, "access_reports_page"), deps.reportHandler.GetCategorySalesReport)
	reports.Get("/export/sales", middleware.RequirePermission(deps.permissionChecker, "manage_reports_page"), deps.reportHandler.ExportSalesCSV)
	reports.Get("/export/inventory", middleware.RequirePermission(deps.permissionChecker, "manage_reports_page"), deps.reportHandler.ExportInventoryCSV)
	reports.Get("/export/top-sellers", middleware.RequirePermission(deps.permissionChecker, "manage_reports_page"), deps.reportHandler.ExportTopSellersCSV)
	reports.Get("/export/comprehensive", middleware.RequirePermission(deps.permissionChecker, "manage_reports_page"), deps.reportHandler.ExportComprehensiveReportCSV)
}

func registerExpenseRoutes(protected fiber.Router, deps *serverDependencies) {
	expenses := protected.Group("/expenses")
	expenses.Get("/", middleware.RequirePermission(deps.permissionChecker, "access_expenses_page"), deps.expenseHandler.List)
	expenses.Get("/categories", middleware.RequireAnyPermission(deps.permissionChecker, expenseCategoryReadPermissions...), deps.expenseHandler.ListCategories)
	expenses.Post("/categories", middleware.RequirePermission(deps.permissionChecker, "manage_categories_page"), deps.expenseHandler.CreateCategory)
	expenses.Get("/categories/:id", middleware.RequireAnyPermission(deps.permissionChecker, expenseCategoryReadPermissions...), deps.expenseHandler.GetCategory)
	expenses.Patch("/categories/:id", middleware.RequirePermission(deps.permissionChecker, "manage_categories_page"), deps.expenseHandler.UpdateCategory)
	expenses.Delete("/categories/:id", middleware.RequirePermission(deps.permissionChecker, "manage_categories_page"), deps.expenseHandler.DeleteCategory)
	expenses.Delete("/categories/:id/permanent", middleware.RequirePermission(deps.permissionChecker, "manage_categories_page"), deps.expenseHandler.PermanentDeleteCategory)
	expenses.Get("/summary", middleware.RequirePermission(deps.permissionChecker, "access_expenses_page"), deps.expenseHandler.GetSummary)
	expenses.Get("/monthly", middleware.RequirePermission(deps.permissionChecker, "access_expenses_page"), deps.expenseHandler.GetMonthlyTotals)
	expenses.Post("/", middleware.RequirePermission(deps.permissionChecker, "manage_expenses_page"), deps.expenseHandler.Create)
	expenses.Get("/:id", middleware.RequirePermission(deps.permissionChecker, "access_expenses_page"), deps.expenseHandler.Get)
	expenses.Patch("/:id", middleware.RequirePermission(deps.permissionChecker, "manage_expenses_page"), deps.expenseHandler.Update)
	expenses.Delete("/:id", middleware.RequirePermission(deps.permissionChecker, "manage_expenses_page"), deps.expenseHandler.Delete)
}

func registerAuditRoutes(protected fiber.Router, deps *serverDependencies) {
	dashboard := protected.Group("/dashboard")
	dashboard.Get("/changes", middleware.RequirePermission(deps.permissionChecker, "access_changes_page"), deps.auditHandler.List)

	logs := protected.Group("/logs")
	logs.Get("/", middleware.RequirePermission(deps.permissionChecker, "access_audit_page"), deps.auditHandler.List)
	logs.Get("/actions", middleware.RequirePermission(deps.permissionChecker, "access_audit_page"), deps.auditHandler.GetActions)
	logs.Get("/summary", middleware.RequirePermission(deps.permissionChecker, "access_audit_page"), deps.auditHandler.GetSummary)
	logs.Get("/entity/:type/:id", middleware.RequirePermission(deps.permissionChecker, "access_audit_page"), deps.auditHandler.GetEntityHistory)
	logs.Get("/user/:id", middleware.RequirePermission(deps.permissionChecker, "access_audit_page"), deps.auditHandler.GetUserActivity)
	logs.Get("/:id", middleware.RequirePermission(deps.permissionChecker, "access_audit_page"), deps.auditHandler.Get)

	auditGroup := protected.Group("/audit")
	auditGroup.Get("/", middleware.RequirePermission(deps.permissionChecker, "access_audit_page"), deps.auditHandler.List)
	auditGroup.Get("/:id", middleware.RequirePermission(deps.permissionChecker, "access_audit_page"), deps.auditHandler.Get)
}

func registerUploadRoutes(protected fiber.Router, deps *serverDependencies) {
	upload := protected.Group("/upload")
	upload.Post("/image", middleware.RequireAnyPermission(deps.permissionChecker, "manage_products_page", "manage_expenses_page"), deps.uploadHandler.UploadImage)
	upload.Delete("/image/:filename", middleware.RequireAnyPermission(deps.permissionChecker, "manage_products_page", "manage_expenses_page"), deps.uploadHandler.DeleteImage)
}

func requireInventoryAdjustmentPermission(checker middleware.PermissionChecker) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := middleware.GetUserID(c)
		if userID == uuid.Nil {
			return middleware.JSONError(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		}

		requiredPermission, err := parseRequiredInventoryPermission(c)
		if err != nil {
			return err
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

func parseRequiredInventoryPermission(c *fiber.Ctx) (string, error) {
	type inventoryAdjustmentRequest struct {
		AdjustmentType string `json:"adjustment_type"`
	}

	var req inventoryAdjustmentRequest
	if err := c.BodyParser(&req); err != nil {
		return "", middleware.JSONError(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
	}

	requiredPermission, ok := adjustmentPermission(req.AdjustmentType)
	if !ok {
		return "", middleware.JSONError(c, fiber.StatusBadRequest, "INVALID_ADJUSTMENT_TYPE", "Invalid adjustment type")
	}

	return requiredPermission, nil
}

func adjustmentPermission(adjustmentType string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(adjustmentType)) {
	case "purchase":
		return "manage_inventory_page", true
	case "damage", "loss":
		return "manage_inventory_page", true
	case "adjustment", "count":
		return "manage_inventory_page", true
	default:
		return "", false
	}
}

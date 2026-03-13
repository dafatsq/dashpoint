package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"dashpoint/backend/internal/audit"
	"dashpoint/backend/internal/auth"
	"dashpoint/backend/internal/config"
	"dashpoint/backend/internal/database"
	"dashpoint/backend/internal/handlers"
	"dashpoint/backend/internal/middleware"
	"dashpoint/backend/internal/repository"
)

func main() {
	// Setup logging
	setupLogging()

	log.Info().Msg("Starting DashPoint POS Backend...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	log.Info().
		Str("port", cfg.Port).
		Str("environment", cfg.Environment).
		Msg("Configuration loaded")

	// Run database migrations
	migrationsPath := "./migrations"
	if err := database.RunMigrations(cfg.DatabaseURL, migrationsPath); err != nil {
		log.Fatal().Err(err).Msg("Failed to run database migrations")
	}

	// Connect to database
	db, err := database.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	// Create JWT manager
	jwtManager := auth.NewJWTManager(cfg.JWTSecret, cfg.JWTExpiryMinutes, cfg.RefreshExpiryHours)

	// Create repositories
	userRepo := repository.NewUserRepository(db.Pool)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db.Pool)
	roleRepo := repository.NewRoleRepository(db.Pool)
	permissionRepo := repository.NewPermissionRepository(db.Pool)
	productRepo := repository.NewProductRepository(db.Pool)
	inventoryRepo := repository.NewInventoryRepository(db.Pool)
	categoryRepo := repository.NewCategoryRepository(db.Pool)
	shiftRepo := repository.NewShiftRepository(db.Pool)
	saleRepo := repository.NewSaleRepository(db.Pool, inventoryRepo)
	reportRepo := repository.NewReportRepository(db.Pool)
	auditRepo := repository.NewAuditRepository(db.Pool)
	expenseRepo := repository.NewExpenseRepository(db.Pool)
	cashDrawerRepo := repository.NewCashDrawerRepository(db.Pool)

	// Initialize audit service
	audit.Init(auditRepo)

	// Create upload directory
	uploadDir := "./uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Fatal().Err(err).Msg("Failed to create upload directory")
	}

	// Create handlers
	healthHandler := handlers.NewHealthHandler(db)
	authHandler := handlers.NewAuthHandler(userRepo, refreshTokenRepo, jwtManager)
	eventsHandler := handlers.NewEventsHandler(jwtManager)
	userHandler := handlers.NewUserHandler(userRepo, roleRepo, permissionRepo)
	userHandler.SetEventsHandler(eventsHandler) // Enable real-time user updates
	roleHandler := handlers.NewRoleHandler(roleRepo, permissionRepo)
	productHandler := handlers.NewProductHandler(productRepo, inventoryRepo, categoryRepo, uploadDir)
	categoryHandler := handlers.NewCategoryHandler(categoryRepo)
	shiftHandler := handlers.NewShiftHandler(shiftRepo)
	saleHandler := handlers.NewSaleHandler(saleRepo, shiftRepo)
	reportHandler := handlers.NewReportHandler(reportRepo)
	auditHandler := handlers.NewAuditHandler(auditRepo)
	expenseHandler := handlers.NewExpenseHandler(expenseRepo, inventoryRepo, productRepo)
	uploadHandler := handlers.NewUploadHandler(uploadDir)
	cashDrawerHandler := handlers.NewCashDrawerHandler(cashDrawerRepo, shiftRepo)

	// Create permission checker function
	permissionChecker := func(c *fiber.Ctx, userID uuid.UUID, permission string) (bool, error) {
		permissions, err := userRepo.GetUserPermissions(c.Context(), userID)
		if err != nil {
			return false, err
		}
		for _, perm := range permissions {
			if perm == permission {
				return true, nil
			}
		}
		return false, nil
	}

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "DashPoint POS API",
		ErrorHandler: errorHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 0, // Disable write timeout for SSE long-polling connections
		IdleTimeout:  120 * time.Second,
	})

	// Setup middleware
	app.Use(middleware.Recover())
	app.Use(middleware.Logger())
	app.Use(middleware.CORS())
	app.Use(middleware.RequestID())

	// Serve static files (uploaded images) with proper headers
	app.Static("/uploads", uploadDir, fiber.Static{
		Browse: false,
		MaxAge: 3600, // Cache for 1 hour
		ModifyResponse: func(c *fiber.Ctx) error {
			// Add CORS headers for images
			c.Set("Access-Control-Allow-Origin", "*")
			c.Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			return nil
		},
	})

	// Setup routes
	setupRoutes(app, jwtManager, userRepo, permissionChecker, healthHandler, authHandler, userHandler, roleHandler, productHandler, categoryHandler, shiftHandler, saleHandler, reportHandler, auditHandler, expenseHandler, eventsHandler, uploadHandler, cashDrawerHandler)

	// Start server in goroutine
	go func() {
		addr := fmt.Sprintf(":%s", cfg.Port)
		if err := app.Listen(addr); err != nil {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	log.Info().Str("port", cfg.Port).Msg("Server started successfully")

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Error().Err(err).Msg("Error during server shutdown")
	}

	log.Info().Msg("Server stopped")
}

func setupLogging() {
	// Pretty logging for development
	if os.Getenv("ENVIRONMENT") != "production" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
	}

	// Set global log level
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if os.Getenv("DEBUG") == "true" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
}

func setupRoutes(
	app *fiber.App,
	jwtManager *auth.JWTManager,
	userRepo *repository.UserRepository,
	permissionChecker middleware.PermissionChecker,
	healthHandler *handlers.HealthHandler,
	authHandler *handlers.AuthHandler,
	userHandler *handlers.UserHandler,
	roleHandler *handlers.RoleHandler,
	productHandler *handlers.ProductHandler,
	categoryHandler *handlers.CategoryHandler,
	shiftHandler *handlers.ShiftHandler,
	saleHandler *handlers.SaleHandler,
	reportHandler *handlers.ReportHandler,
	auditHandler *handlers.AuditHandler,
	expenseHandler *handlers.ExpenseHandler,
	eventsHandler *handlers.EventsHandler,
	uploadHandler *handlers.UploadHandler,
	cashDrawerHandler *handlers.CashDrawerHandler,
) {
	// Root endpoint
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"name":    "DashPoint POS API",
			"version": "1.0.0",
			"status":  "running",
		})
	})

	// API v1 group
	api := app.Group("/api/v1")

	// Health endpoints (public)
	api.Get("/health", healthHandler.Check)
	api.Get("/ping", healthHandler.Ping)

	// Auth endpoints (public)
	authGroup := api.Group("/auth")
	authGroup.Post("/login", authHandler.Login)
	authGroup.Post("/pin-login", authHandler.PINLogin)
	authGroup.Post("/refresh", authHandler.Refresh)
	authGroup.Post("/logout", authHandler.Logout)

	// SSE events endpoint (token passed via query param for EventSource compatibility)
	api.Get("/events/subscribe", eventsHandler.Subscribe)

	// Protected routes group
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(jwtManager, userRepo))

	// Current user endpoint
	// Current user endpoint (returns full user profile with permissions)
	protected.Get("/me", authHandler.Me)

	// Roles and Permissions endpoints
	protected.Get("/roles", roleHandler.ListRoles)
	protected.Get("/roles/:id", roleHandler.GetRole)
	protected.Get("/permissions", roleHandler.ListPermissions)

	// User management endpoints
	users := protected.Group("/users")
	users.Get("/", middleware.RequirePermission(permissionChecker, "can_view_users"), userHandler.List)
	users.Get("/:id", middleware.RequirePermission(permissionChecker, "can_view_users"), userHandler.Get)
	users.Post("/", middleware.RequirePermission(permissionChecker, "can_create_user"), userHandler.Create)
	users.Patch("/:id", middleware.RequirePermission(permissionChecker, "can_edit_user"), userHandler.Update)
	users.Patch("/:id/password", middleware.RequirePermission(permissionChecker, "can_edit_user"), userHandler.UpdatePassword)
	users.Patch("/:id/pin", middleware.RequirePermission(permissionChecker, "can_edit_user"), userHandler.UpdatePIN)
	users.Delete("/:id", middleware.RequirePermission(permissionChecker, "can_delete_user"), userHandler.Delete)
	users.Delete("/:id/permanent", middleware.RequirePermission(permissionChecker, "can_delete_user"), userHandler.PermanentDelete)

	// Permission management endpoints (requires can_manage_permissions)
	userPermissions := protected.Group("/users")
	userPermissions.Get("/:id/permissions", middleware.RequirePermission(permissionChecker, "can_manage_permissions"), userHandler.GetPermissions)
	userPermissions.Patch("/:id/permissions", middleware.RequirePermission(permissionChecker, "can_manage_permissions"), userHandler.SetPermissions)

	// Category endpoints (requires can_view_categories for viewing, can_manage_categories for modifications)
	categories := protected.Group("/categories")
	categories.Get("/", middleware.RequirePermission(permissionChecker, "can_view_categories"), categoryHandler.List)
	categories.Get("/:id", middleware.RequirePermission(permissionChecker, "can_view_categories"), categoryHandler.Get)
	categories.Post("/", middleware.RequirePermission(permissionChecker, "can_manage_categories"), categoryHandler.Create)
	categories.Patch("/:id", middleware.RequirePermission(permissionChecker, "can_manage_categories"), categoryHandler.Update)
	categories.Delete("/:id", middleware.RequirePermission(permissionChecker, "can_manage_categories"), categoryHandler.Delete)
	categories.Delete("/:id/permanent", middleware.RequirePermission(permissionChecker, "can_manage_categories"), categoryHandler.PermanentDelete)

	// Product endpoints
	products := protected.Group("/products")
	products.Get("/", productHandler.List)
	products.Get("/lookup", productHandler.Lookup)
	products.Get("/:id", productHandler.Get)
	products.Get("/:id/inventory", productHandler.GetInventory)
	products.Post("/", middleware.RequirePermission(permissionChecker, "can_create_product"), productHandler.Create)
	products.Patch("/:id", middleware.RequirePermission(permissionChecker, "can_edit_product"), productHandler.Update)
	products.Delete("/:id", middleware.RequirePermission(permissionChecker, "can_delete_product"), productHandler.Delete)
	products.Delete("/:id/permanent", middleware.RequirePermission(permissionChecker, "can_delete_product"), productHandler.PermanentDelete)

	// Inventory endpoints
	inventory := protected.Group("/inventory")
	inventory.Get("/low-stock", productHandler.GetLowStock)
	inventory.Post("/adjust", middleware.RequirePermission(permissionChecker, "can_edit_inventory"), productHandler.AdjustStock)

	// Shift endpoints
	shifts := protected.Group("/shifts")
	shifts.Get("/current", shiftHandler.GetCurrentShift)
	shifts.Post("/start", shiftHandler.StartShift)
	shifts.Post("/close", shiftHandler.CloseShift)
	shifts.Post("/pay-in", cashDrawerHandler.PayIn)
	shifts.Post("/pay-out", cashDrawerHandler.PayOut)
	shifts.Get("/", shiftHandler.ListShifts)
	shifts.Get("/:id", shiftHandler.GetShift)
	shifts.Get("/:id/operations", cashDrawerHandler.ListOperations)

	// Sales endpoints
	sales := protected.Group("/sales")
	sales.Post("/", middleware.RequirePermission(permissionChecker, "can_create_sale"), saleHandler.CreateSale)
	sales.Get("/", middleware.RequirePermission(permissionChecker, "can_view_sales"), saleHandler.ListSales)
	sales.Get("/summary/daily", middleware.RequirePermission(permissionChecker, "can_view_sales"), saleHandler.GetDailySummary)
	sales.Get("/invoice/:invoiceNo", middleware.RequirePermission(permissionChecker, "can_view_sales"), saleHandler.GetSaleByInvoice)
	sales.Get("/:id", middleware.RequirePermission(permissionChecker, "can_view_sales"), saleHandler.GetSale)
	sales.Post("/:id/void", middleware.RequirePermission(permissionChecker, "can_void_sale"), saleHandler.VoidSale)

	// Reports endpoints (owner/manager only for most)
	reports := protected.Group("/reports")
	reports.Get("/daily", reportHandler.GetDailySalesReport)
	reports.Get("/sales", reportHandler.GetSalesRangeReport)
	reports.Get("/top-sellers", reportHandler.GetTopSellers)
	reports.Get("/inventory", reportHandler.GetInventoryValuation)
	reports.Get("/cash", reportHandler.GetCashReport)
	reports.Get("/by-employee", middleware.RequirePermission(permissionChecker, "can_view_reports"), reportHandler.GetEmployeeSalesReport)
	reports.Get("/by-category", reportHandler.GetCategorySalesReport)

	// Export endpoints (requires can_export_data permission)
	reports.Get("/export/sales", middleware.RequirePermission(permissionChecker, "can_export_data"), reportHandler.ExportSalesCSV)
	reports.Get("/export/inventory", middleware.RequirePermission(permissionChecker, "can_export_data"), reportHandler.ExportInventoryCSV)
	reports.Get("/export/top-sellers", middleware.RequirePermission(permissionChecker, "can_export_data"), reportHandler.ExportTopSellersCSV)
	reports.Get("/export/comprehensive", middleware.RequirePermission(permissionChecker, "can_export_data"), reportHandler.ExportComprehensiveReportCSV)

	// Expense endpoints
	expenses := protected.Group("/expenses")
	expenses.Get("/", middleware.RequirePermission(permissionChecker, "can_view_expenses"), expenseHandler.List)
	expenses.Get("/categories", middleware.RequireAnyPermission(permissionChecker, "can_view_expenses", "can_view_categories"), expenseHandler.ListCategories)
	expenses.Post("/categories", middleware.RequireAnyPermission(permissionChecker, "can_manage_expenses", "can_manage_categories"), expenseHandler.CreateCategory)
	expenses.Get("/categories/:id", middleware.RequireAnyPermission(permissionChecker, "can_view_expenses", "can_view_categories"), expenseHandler.GetCategory)
	expenses.Patch("/categories/:id", middleware.RequireAnyPermission(permissionChecker, "can_manage_expenses", "can_manage_categories"), expenseHandler.UpdateCategory)
	expenses.Delete("/categories/:id", middleware.RequireAnyPermission(permissionChecker, "can_manage_expenses", "can_manage_categories"), expenseHandler.DeleteCategory)
	expenses.Delete("/categories/:id/permanent", middleware.RequireAnyPermission(permissionChecker, "can_manage_expenses", "can_manage_categories"), expenseHandler.PermanentDeleteCategory)
	expenses.Get("/summary", middleware.RequirePermission(permissionChecker, "can_view_expenses"), expenseHandler.GetSummary)
	expenses.Get("/monthly", middleware.RequirePermission(permissionChecker, "can_view_expenses"), expenseHandler.GetMonthlyTotals)
	expenses.Post("/", middleware.RequirePermission(permissionChecker, "can_manage_expenses"), expenseHandler.Create)
	expenses.Get("/:id", middleware.RequirePermission(permissionChecker, "can_view_expenses"), expenseHandler.Get)
	expenses.Patch("/:id", middleware.RequirePermission(permissionChecker, "can_manage_expenses"), expenseHandler.Update)
	expenses.Delete("/:id", middleware.RequirePermission(permissionChecker, "can_manage_expenses"), expenseHandler.Delete)

	// Dashboard endpoints (accessible to all authenticated users)
	dashboard := protected.Group("/dashboard")
	dashboard.Get("/changes", auditHandler.List) // reuses audit list, no permission required

	// Audit log endpoints (requires can_view_audit_logs permission)
	logs := protected.Group("/logs")
	logs.Get("/", middleware.RequirePermission(permissionChecker, "can_view_audit_logs"), auditHandler.List)
	logs.Get("/actions", middleware.RequirePermission(permissionChecker, "can_view_audit_logs"), auditHandler.GetActions)
	logs.Get("/summary", middleware.RequirePermission(permissionChecker, "can_view_audit_logs"), auditHandler.GetSummary)
	logs.Get("/entity/:type/:id", middleware.RequirePermission(permissionChecker, "can_view_audit_logs"), auditHandler.GetEntityHistory)
	logs.Get("/user/:id", middleware.RequirePermission(permissionChecker, "can_view_audit_logs"), auditHandler.GetUserActivity)
	logs.Get("/:id", middleware.RequirePermission(permissionChecker, "can_view_audit_logs"), auditHandler.Get)

	// Upload endpoints (authenticated users only)
	upload := protected.Group("/upload")
	upload.Post("/image", uploadHandler.UploadImage)
	upload.Delete("/image/:filename", uploadHandler.DeleteImage)
}

func errorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "Internal Server Error"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	log.Error().
		Err(err).
		Int("status", code).
		Str("path", c.Path()).
		Msg("Request error")

	return c.Status(code).JSON(fiber.Map{
		"code":    code,
		"message": message,
	})
}

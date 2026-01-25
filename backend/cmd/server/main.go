package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
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

	// Initialize audit service
	audit.Init(auditRepo)

	// Create handlers
	healthHandler := handlers.NewHealthHandler(db)
	authHandler := handlers.NewAuthHandler(userRepo, refreshTokenRepo, jwtManager)
	eventsHandler := handlers.NewEventsHandler(jwtManager)
	userHandler := handlers.NewUserHandler(userRepo, roleRepo, permissionRepo)
	userHandler.SetEventsHandler(eventsHandler) // Enable real-time user updates
	roleHandler := handlers.NewRoleHandler(roleRepo, permissionRepo)
	productHandler := handlers.NewProductHandler(productRepo, inventoryRepo, categoryRepo)
	categoryHandler := handlers.NewCategoryHandler(categoryRepo)
	shiftHandler := handlers.NewShiftHandler(shiftRepo)
	saleHandler := handlers.NewSaleHandler(saleRepo, shiftRepo)
	reportHandler := handlers.NewReportHandler(reportRepo)
	auditHandler := handlers.NewAuditHandler(auditRepo)
	expenseHandler := handlers.NewExpenseHandler(expenseRepo, inventoryRepo, productRepo)

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

	// Setup routes
	setupRoutes(app, jwtManager, userRepo, healthHandler, authHandler, userHandler, roleHandler, productHandler, categoryHandler, shiftHandler, saleHandler, reportHandler, auditHandler, expenseHandler, eventsHandler)

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

	// User management endpoints (owner/manager only)
	users := protected.Group("/users")
	users.Use(middleware.RequireRole("owner", "manager"))
	users.Get("/", userHandler.List)
	users.Get("/:id", userHandler.Get)
	users.Post("/", userHandler.Create)
	users.Patch("/:id", userHandler.Update)
	users.Patch("/:id/password", userHandler.UpdatePassword)
	users.Patch("/:id/pin", userHandler.UpdatePIN)
	users.Delete("/:id", userHandler.Delete)
	users.Delete("/:id/permanent", userHandler.PermanentDelete)
	users.Get("/:id/permissions", userHandler.GetPermissions)
	users.Patch("/:id/permissions", userHandler.SetPermissions)

	// Category endpoints (all authenticated users can view, owner/manager can modify)
	categories := protected.Group("/categories")
	categories.Get("/", categoryHandler.List)
	categories.Get("/:id", categoryHandler.Get)
	categories.Post("/", middleware.RequireRole("owner", "manager"), categoryHandler.Create)
	categories.Patch("/:id", middleware.RequireRole("owner", "manager"), categoryHandler.Update)
	categories.Delete("/:id", middleware.RequireRole("owner", "manager"), categoryHandler.Delete)

	// Product endpoints
	products := protected.Group("/products")
	products.Get("/", productHandler.List)
	products.Get("/lookup", productHandler.Lookup)
	products.Get("/:id", productHandler.Get)
	products.Get("/:id/inventory", productHandler.GetInventory)
	products.Post("/", middleware.RequireRole("owner", "manager"), productHandler.Create)
	products.Patch("/:id", middleware.RequireRole("owner", "manager"), productHandler.Update)
	products.Delete("/:id", middleware.RequireRole("owner", "manager"), productHandler.Delete)
	products.Delete("/:id/permanent", middleware.RequireRole("owner", "manager"), productHandler.PermanentDelete)

	// Inventory endpoints
	inventory := protected.Group("/inventory")
	inventory.Get("/low-stock", productHandler.GetLowStock)
	inventory.Post("/adjust", middleware.RequireRole("owner", "manager", "cashier"), productHandler.AdjustStock)

	// Shift endpoints
	shifts := protected.Group("/shifts")
	shifts.Get("/current", shiftHandler.GetCurrentShift)
	shifts.Post("/start", shiftHandler.StartShift)
	shifts.Post("/close", shiftHandler.CloseShift)
	shifts.Get("/", shiftHandler.ListShifts)
	shifts.Get("/:id", shiftHandler.GetShift)

	// Sales endpoints
	sales := protected.Group("/sales")
	sales.Post("/", saleHandler.CreateSale)
	sales.Get("/", saleHandler.ListSales)
	sales.Get("/summary/daily", saleHandler.GetDailySummary)
	sales.Get("/invoice/:invoiceNo", saleHandler.GetSaleByInvoice)
	sales.Get("/:id", saleHandler.GetSale)
	sales.Post("/:id/void", saleHandler.VoidSale)

	// Reports endpoints (owner/manager only for most)
	reports := protected.Group("/reports")
	reports.Get("/daily", reportHandler.GetDailySalesReport)
	reports.Get("/sales", reportHandler.GetSalesRangeReport)
	reports.Get("/top-sellers", reportHandler.GetTopSellers)
	reports.Get("/inventory", reportHandler.GetInventoryValuation)
	reports.Get("/cash", reportHandler.GetCashReport)
	reports.Get("/by-employee", middleware.RequireRole("owner", "manager"), reportHandler.GetEmployeeSalesReport)
	reports.Get("/by-category", reportHandler.GetCategorySalesReport)

	// Export endpoints (owner/manager only)
	reports.Get("/export/sales", reportHandler.ExportSalesCSV)
	reports.Get("/export/inventory", reportHandler.ExportInventoryCSV)
	reports.Get("/export/top-sellers", reportHandler.ExportTopSellersCSV)

	// Expense endpoints (owner/manager only)
	expenses := protected.Group("/expenses")
	expenses.Use(middleware.RequireRole("owner", "manager"))
	expenses.Get("/", expenseHandler.List)
	expenses.Get("/categories", expenseHandler.ListCategories)
	expenses.Post("/categories", expenseHandler.CreateCategory)
	expenses.Get("/summary", expenseHandler.GetSummary)
	expenses.Get("/monthly", expenseHandler.GetMonthlyTotals)
	expenses.Post("/", expenseHandler.Create)
	expenses.Get("/:id", expenseHandler.Get)
	expenses.Patch("/:id", expenseHandler.Update)
	expenses.Delete("/:id", expenseHandler.Delete)

	// Audit log endpoints (owner only)
	logs := protected.Group("/logs")
	logs.Use(middleware.RequireRole("owner"))
	logs.Get("/", auditHandler.List)
	logs.Get("/actions", auditHandler.GetActions)
	logs.Get("/summary", auditHandler.GetSummary)
	logs.Get("/entity/:type/:id", auditHandler.GetEntityHistory)
	logs.Get("/user/:id", auditHandler.GetUserActivity)
	logs.Get("/:id", auditHandler.Get)
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

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"dashpoint/backend/internal/audit"
	"dashpoint/backend/internal/auth"
	"dashpoint/backend/internal/config"
	"dashpoint/backend/internal/database"
	"dashpoint/backend/internal/handlers"
	"dashpoint/backend/internal/middleware"
	"dashpoint/backend/internal/repository"
)

const uploadDirectory = "./uploads"
const apiReadBufferSize = 32 * 1024

type serverDependencies struct {
	jwtManager        *auth.JWTManager
	userRepo          *repository.UserRepository
	permissionChecker middleware.PermissionChecker
	healthHandler     *handlers.HealthHandler
	authHandler       *handlers.AuthHandler
	setupHandler      *handlers.SetupHandler
	eventsHandler     *handlers.EventsHandler
	userHandler       *handlers.UserHandler
	roleHandler       *handlers.RoleHandler
	productHandler    *handlers.ProductHandler
	categoryHandler   *handlers.CategoryHandler
	shiftHandler      *handlers.ShiftHandler
	saleHandler       *handlers.SaleHandler
	reportHandler     *handlers.ReportHandler
	auditHandler      *handlers.AuditHandler
	expenseHandler    *handlers.ExpenseHandler
	uploadHandler     *handlers.UploadHandler
	cashDrawerHandler *handlers.CashDrawerHandler
	uploadDir         string
}

func buildServerDependencies(cfg *config.Config, db *database.DB) (*serverDependencies, error) {
	jwtManager := auth.NewJWTManager(cfg.JWTSecret, cfg.JWTExpiryMinutes, cfg.RefreshExpiryHours)

	userRepo := repository.NewUserRepository(db.Pool)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db.Pool)
	roleRepo := repository.NewRoleRepository(db.Pool)
	productRepo := repository.NewProductRepository(db.Pool)
	inventoryRepo := repository.NewInventoryRepository(db.Pool)
	categoryRepo := repository.NewCategoryRepository(db.Pool)
	shiftRepo := repository.NewShiftRepository(db.Pool)
	saleRepo := repository.NewSaleRepository(db.Pool, inventoryRepo)
	reportRepo := repository.NewReportRepository(db.Pool)
	auditRepo := repository.NewAuditRepository(db.Pool)
	expenseRepo := repository.NewExpenseRepository(db.Pool)
	cashDrawerRepo := repository.NewCashDrawerRepository(db.Pool)

	audit.Init(auditRepo)

	if err := os.MkdirAll(uploadDirectory, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	healthHandler := handlers.NewHealthHandler(db)
	authHandler := handlers.NewAuthHandler(userRepo, refreshTokenRepo, jwtManager)
	setupHandler := handlers.NewSetupHandler(userRepo, roleRepo)
	eventsHandler := handlers.NewEventsHandler(jwtManager, userRepo, cfg.CORSOrigins)
	userHandler := handlers.NewUserHandler(userRepo, roleRepo)
	userHandler.SetEventsHandler(eventsHandler)
	userHandler.SetRefreshTokenRevoker(refreshTokenRepo)
	roleHandler := handlers.NewRoleHandler(roleRepo)
	roleHandler.SetEventsHandler(eventsHandler)

	deps := &serverDependencies{
		jwtManager:        jwtManager,
		userRepo:          userRepo,
		permissionChecker: newPermissionChecker(userRepo),
		healthHandler:     healthHandler,
		authHandler:       authHandler,
		setupHandler:      setupHandler,
		eventsHandler:     eventsHandler,
		userHandler:       userHandler,
		roleHandler:       roleHandler,
		productHandler:    handlers.NewProductHandler(productRepo, inventoryRepo, categoryRepo, uploadDirectory),
		categoryHandler:   handlers.NewCategoryHandler(categoryRepo),
		shiftHandler:      handlers.NewShiftHandler(shiftRepo),
		saleHandler:       handlers.NewSaleHandler(saleRepo, shiftRepo),
		reportHandler:     handlers.NewReportHandler(reportRepo),
		auditHandler:      handlers.NewAuditHandler(auditRepo),
		expenseHandler:    handlers.NewExpenseHandler(expenseRepo, inventoryRepo, productRepo),
		uploadHandler:     handlers.NewUploadHandler(uploadDirectory),
		cashDrawerHandler: handlers.NewCashDrawerHandler(cashDrawerRepo, shiftRepo),
		uploadDir:         uploadDirectory,
	}

	return deps, nil
}

func newPermissionChecker(userRepo *repository.UserRepository) middleware.PermissionChecker {
	return func(c *fiber.Ctx, userID uuid.UUID, permission string) (bool, error) {
		permissions, err := userRepo.GetUserPermissions(c.Context(), userID)
		if err != nil {
			return false, err
		}
		for _, granted := range permissions {
			if granted == permission {
				return true, nil
			}
		}
		return false, nil
	}
}

func newServerApp(cfg *config.Config, deps *serverDependencies) *fiber.App {
	if cfg.Environment == "production" && len(cfg.TrustedProxies) == 0 {
		log.Warn().Msg("TRUSTED_PROXIES is empty in production: forwarded headers are ignored, so behind a reverse proxy cookies lose Secure and audit IPs are the proxy address")
	}

	app := fiber.New(serverFiberConfig(cfg.TrustedProxies))

	app.Use(middleware.Recover())
	app.Use(middleware.Logger())
	app.Use(middleware.CORS(cfg.CORSOrigins))
	app.Use(middleware.RequestID())
	app.Use(middleware.SecureHeaders())

	app.Static("/uploads", deps.uploadDir, fiber.Static{
		Browse: false,
		MaxAge: 3600,
		ModifyResponse: func(c *fiber.Ctx) error {
			middleware.ApplyCORSHeaders(c, cfg.CORSOrigins)
			c.Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			return nil
		},
	})

	registerRoutes(app, deps)
	return app
}

func serverFiberConfig(trustedProxies []string) fiber.Config {
	return fiber.Config{
		AppName:        "DashPoint POS API",
		ErrorHandler:   errorHandler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   0,
		IdleTimeout:    120 * time.Second,
		ReadBufferSize: apiReadBufferSize,
		// Uploads accept 5 MB; without an explicit limit Fiber's implicit 4 MB
		// default would reject them first. This is also the only size bound on
		// endpoints that skip the strict bounded JSON parsers.
		BodyLimit: 6 * 1024 * 1024,
		// Only honor X-Forwarded-For / X-Forwarded-Proto from proxies on this
		// allowlist. Empty by default: direct deployments ignore spoofed
		// forwarding headers entirely, and c.IP() is the real socket address.
		EnableTrustedProxyCheck: true,
		TrustedProxies:          trustedProxies,
		ProxyHeader:             "X-Forwarded-For",
	}
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
		Str("request_id", middleware.GetRequestID(c)).
		Msg("Request error")

	return c.Status(code).JSON(fiber.Map{
		"code":       code,
		"message":    message,
		"request_id": middleware.GetRequestID(c),
	})
}

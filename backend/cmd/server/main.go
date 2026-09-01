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

	"dashpoint/backend/internal/config"
	"dashpoint/backend/internal/database"
)

func main() {
	setupLogging()

	log.Info().Msg("Starting DashPoint POS Backend...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	log.Info().
		Str("port", cfg.Port).
		Str("environment", cfg.Environment).
		Strs("cors_origins", cfg.CORSOrigins).
		Msg("Configuration loaded")

	if err := runMigrations(cfg); err != nil {
		log.Fatal().Err(err).Msg("Failed to run database migrations")
	}

	db, err := database.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	// Demo branch only: keep the public demo accounts present and active.
	if err := database.EnsureDemoAccounts(context.Background(), db); err != nil {
		log.Fatal().Err(err).Msg("Failed to ensure demo accounts")
	}

	deps, err := buildServerDependencies(cfg, db)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to build server dependencies")
	}

	app := newServerApp(cfg, deps)
	startServer(app, cfg.Port)
	waitForShutdown(app)
}

func setupLogging() {
	if os.Getenv("ENVIRONMENT") != "production" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
	}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if os.Getenv("DEBUG") == "true" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
}

func runMigrations(cfg *config.Config) error {
	migrationsPath := resolveMigrationsPath()

	log.Info().Str("migrations_path", migrationsPath).Msg("Running database migrations")
	if err := database.RunMigrations(cfg.DatabaseURL, migrationsPath); err != nil {
		return err
	}

	return nil
}

func resolveMigrationsPath() string {
	migrationsPath := "./migrations"
	if _, err := os.Stat(migrationsPath); err == nil {
		return migrationsPath
	}

	fallbackPath := "./backend/migrations"
	if _, err := os.Stat(fallbackPath); err == nil {
		return fallbackPath
	}

	return migrationsPath
}

func startServer(app *fiber.App, port string) {
	go func() {
		addr := fmt.Sprintf(":%s", port)
		if err := app.Listen(addr); err != nil {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	log.Info().Str("port", port).Msg("Server started successfully")
}

func waitForShutdown(app *fiber.App) {
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

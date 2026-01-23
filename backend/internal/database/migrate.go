package database

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/rs/zerolog/log"
)

// RunMigrations runs all pending database migrations
func RunMigrations(databaseURL, migrationsPath string) error {
	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		databaseURL,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	// Run migrations
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Info().Msg("No new migrations to apply")
			return nil
		}
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	version, dirty, err := m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		return fmt.Errorf("failed to get migration version: %w", err)
	}

	log.Info().
		Uint("version", version).
		Bool("dirty", dirty).
		Msg("Migrations applied successfully")

	return nil
}

// MigrateDown rolls back the last migration
func MigrateDown(databaseURL, migrationsPath string) error {
	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		databaseURL,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	if err := m.Steps(-1); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Info().Msg("No migrations to rollback")
			return nil
		}
		return fmt.Errorf("failed to rollback migration: %w", err)
	}

	log.Info().Msg("Migration rolled back successfully")
	return nil
}

// GetMigrationVersion returns the current migration version
func GetMigrationVersion(databaseURL, migrationsPath string) (uint, bool, error) {
	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		databaseURL,
	)
	if err != nil {
		return 0, false, fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	version, dirty, err := m.Version()
	if err != nil {
		if errors.Is(err, migrate.ErrNilVersion) {
			return 0, false, nil
		}
		return 0, false, err
	}

	return version, dirty, nil
}

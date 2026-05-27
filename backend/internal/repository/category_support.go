package repository

import (
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"dashpoint/backend/internal/models"
)

func normalizeCategoryStatus(status string) string {
	switch status {
	case "all", "active", "archived":
		return status
	default:
		return "active"
	}
}

func collectCategoryIDs(ids []uuid.UUID) []uuid.UUID {
	collected := make([]uuid.UUID, len(ids))
	copy(collected, ids)
	return collected
}

func scanCategory(scanner interface {
	Scan(dest ...interface{}) error
}) (*models.Category, error) {
	category := &models.Category{}
	err := scanner.Scan(
		&category.ID,
		&category.Name,
		&category.Description,
		&category.IsActive,
		&category.CreatedAt,
		&category.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return category, nil
}

func scanCategoryRow(row pgx.Row) (*models.Category, error) {
	return scanCategory(row)
}

func isCategoryNotFoundError(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

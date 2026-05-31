package handlers

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"

	"dashpoint/backend/internal/models"
)

func expenseMessage(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{"message": message})
}

func expenseInternalError(c *fiber.Ctx, err error, message string) error {
	log.Error().Err(err).Msg(message)
	return expenseMessage(c, fiber.StatusInternalServerError, message)
}

func expenseToResponse(e *models.Expense) ExpenseResponse {
	resp := ExpenseResponse{
		ID:               e.ID.String(),
		Amount:           e.Amount.String(),
		Description:      e.Description,
		ExpenseDate:      e.ExpenseDate.Format(reportDateLayout),
		AppliesInventory: e.AppliesInventory,
		Vendor:           e.Vendor,
		ReferenceNumber:  e.ReferenceNumber,
		Notes:            e.Notes,
		CreatedBy:        e.CreatedBy.String(),
		CreatedByName:    e.CreatedByName,
		CreatedAt:        e.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        e.UpdatedAt.Format(time.RFC3339),
	}
	if e.CategoryID != nil {
		categoryID := e.CategoryID.String()
		resp.CategoryID = &categoryID
	}
	resp.CategoryName = e.CategoryName
	if e.ProductID != nil {
		productID := e.ProductID.String()
		resp.ProductID = &productID
	}
	resp.ProductName = e.ProductName
	if e.Quantity != nil {
		quantity := e.Quantity.String()
		resp.Quantity = &quantity
	}
	return resp
}

func parseExpenseParamID(c *fiber.Ctx, key, invalidMessage string) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Params(key))
	if err != nil {
		return uuid.Nil, expenseMessage(c, fiber.StatusBadRequest, invalidMessage)
	}
	return id, nil
}

func parseOptionalExpenseUUIDField(value *string, invalidMessage string) (*uuid.UUID, error) {
	if value == nil {
		return nil, nil
	}
	if *value == "" {
		return nil, nil
	}

	parsed, err := uuid.Parse(*value)
	if err != nil {
		return nil, fmt.Errorf("%s", invalidMessage)
	}

	return &parsed, nil
}

func parseOptionalExpensePositiveDecimalField(value *string, invalidMessage string) (*decimal.Decimal, error) {
	if value == nil {
		return nil, nil
	}
	if *value == "" {
		return nil, nil
	}

	parsed, err := decimal.NewFromString(*value)
	if err != nil || !parsed.GreaterThan(decimal.Zero) {
		return nil, fmt.Errorf("%s", invalidMessage)
	}

	return &parsed, nil
}

func parseRequiredAmount(value string) (decimal.Decimal, error) {
	amount, err := decimal.NewFromString(value)
	if err != nil || amount.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero, fmt.Errorf("Invalid amount")
	}
	return amount, nil
}

func parseExpenseDateField(value string, fieldMessage string) (time.Time, error) {
	parsed, err := parseReportDay(value, fieldMessage)
	if err != nil {
		return time.Time{}, fmt.Errorf("%s", fieldMessage)
	}
	return parsed, nil
}

func stringPtr(value string) *string {
	return &value
}

func expenseAuditValues(expense *models.Expense) map[string]interface{} {
	values := map[string]interface{}{
		"amount":       expense.Amount.String(),
		"description":  expense.Description,
		"expense_date": expense.ExpenseDate.Format(reportDateLayout),
	}
	if expense.Vendor != nil {
		values["vendor"] = *expense.Vendor
	}
	if expense.ReferenceNumber != nil {
		values["reference_number"] = *expense.ReferenceNumber
	}
	if expense.Notes != nil {
		values["notes"] = *expense.Notes
	}
	if expense.CategoryName != nil {
		values["category"] = *expense.CategoryName
	} else if expense.CategoryID != nil {
		values["category"] = expense.CategoryID.String()
	}
	if expense.ProductName != nil {
		values["product"] = *expense.ProductName
	} else if expense.ProductID != nil {
		values["product"] = expense.ProductID.String()
	}
	if expense.Quantity != nil {
		values["quantity"] = expense.Quantity.String()
	}
	values["applies_inventory"] = expense.AppliesInventory
	return values
}

func categoryNameByID(categories []models.ExpenseCategory, categoryID uuid.UUID) *string {
	for _, category := range categories {
		if category.ID == categoryID {
			return &category.Name
		}
	}
	return nil
}

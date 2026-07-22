package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"

	"dashpoint/backend/internal/middleware"
	"dashpoint/backend/internal/models"
)

const (
	expenseListDefaultLimit       = 50
	expenseListMaxLimit           = 100
	expenseCategoryNameMaxLength  = 100
	expenseDescriptionMaxLength   = 1000
	expenseVendorMaxLength        = 255
	expenseReferenceMaxLength     = 100
	expenseNotesMaxLength         = 1000
	expenseCategoryDescriptionMax = 1000
)

var (
	maxExpenseAmount   = decimal.RequireFromString("9999999999999.99")
	maxExpenseQuantity = decimal.RequireFromString("999999999999.999")
)

func expenseMessage(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{"message": message})
}

func expenseInternalError(c *fiber.Ctx, err error, message string) error {
	log.Error().Err(err).Msg(message)
	return expenseMessage(c, fiber.StatusInternalServerError, message)
}

func parseExpenseBody(c *fiber.Ctx, dest interface{}) error {
	body := c.Body()
	var rawFields map[string]json.RawMessage
	decoder := json.NewDecoder(bytes.NewReader(body))
	if err := decoder.Decode(&rawFields); err != nil {
		return expenseMessage(c, fiber.StatusBadRequest, "Invalid request body")
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return expenseMessage(c, fiber.StatusBadRequest, "Invalid request body")
	}
	allowedFields := expenseJSONFieldSet(dest)
	for field := range rawFields {
		if !allowedFields[field] {
			return expenseMessage(c, fiber.StatusBadRequest, "Invalid request body")
		}
	}
	if err := json.Unmarshal(body, dest); err != nil {
		return expenseMessage(c, fiber.StatusBadRequest, "Invalid request body")
	}
	return nil
}

func expenseJSONFieldSet(dest interface{}) map[string]bool {
	allowed := make(map[string]bool)
	typ := reflect.TypeOf(dest)
	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return allowed
	}
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		jsonName := strings.Split(field.Tag.Get("json"), ",")[0]
		if jsonName == "-" {
			continue
		}
		if jsonName == "" {
			jsonName = field.Name
		}
		allowed[jsonName] = true
	}
	return allowed
}

func requireExpenseExpectedUpdatedAt(c *fiber.Ctx, expectedUpdatedAt *string, actualUpdatedAt time.Time) (bool, error) {
	if expectedUpdatedAt == nil || strings.TrimSpace(*expectedUpdatedAt) == "" {
		return false, middleware.JSONError(c, fiber.StatusBadRequest, "EXPECTED_UPDATED_AT_REQUIRED", "expected_updated_at is required")
	}
	stale, staleErr := isStaleSubmit(expectedUpdatedAt, actualUpdatedAt)
	if staleErr != nil {
		return false, middleware.JSONError(c, fiber.StatusBadRequest, "INVALID_EXPECTED_UPDATED_AT", "Invalid expected_updated_at")
	}
	if stale {
		return false, middleware.JSONError(c, fiber.StatusConflict, "STALE_SUBMIT", staleSubmitMessage)
	}
	return true, nil
}

func requireExpenseExpectedProductUpdatedAt(c *fiber.Ctx, expectedUpdatedAt *string, actualUpdatedAt time.Time) (bool, error) {
	if expectedUpdatedAt == nil || strings.TrimSpace(*expectedUpdatedAt) == "" {
		return false, middleware.JSONError(c, fiber.StatusBadRequest, "EXPECTED_PRODUCT_UPDATED_AT_REQUIRED", "expected_product_updated_at is required")
	}
	stale, staleErr := isStaleSubmit(expectedUpdatedAt, actualUpdatedAt)
	if staleErr != nil {
		return false, middleware.JSONError(c, fiber.StatusBadRequest, "INVALID_EXPECTED_PRODUCT_UPDATED_AT", "Invalid expected_product_updated_at")
	}
	if stale {
		return false, middleware.JSONError(c, fiber.StatusConflict, "STALE_SUBMIT", staleSubmitMessage)
	}
	return true, nil
}

func (h *ExpenseHandler) requireExpenseProductCurrent(c *fiber.Ctx, productID uuid.UUID, expectedUpdatedAt *string) (bool, error) {
	if h.productRepo == nil {
		return false, expenseMessage(c, fiber.StatusInternalServerError, "Failed to load product")
	}
	product, err := h.productRepo.GetByID(c.Context(), productID)
	if err != nil {
		return false, expenseInternalError(c, err, "Failed to load product")
	}
	if product == nil {
		return false, expenseMessage(c, fiber.StatusBadRequest, "Product not found")
	}
	if !product.IsActive {
		return false, middleware.JSONError(c, fiber.StatusConflict, "PRODUCT_INACTIVE", "Archived products cannot be changed")
	}
	return requireExpenseExpectedProductUpdatedAt(c, expectedUpdatedAt, product.UpdatedAt)
}

func parseExpensePagination(c *fiber.Ctx) (int, int, error) {
	limit := expenseListDefaultLimit
	if rawLimit := strings.TrimSpace(c.Query("limit")); rawLimit != "" {
		parsedLimit, err := strconv.Atoi(rawLimit)
		if err != nil {
			return 0, 0, fmt.Errorf("Invalid limit")
		}
		limit = parsedLimit
	}

	offset := 0
	if rawOffset := strings.TrimSpace(c.Query("offset")); rawOffset != "" {
		parsedOffset, err := strconv.Atoi(rawOffset)
		if err != nil {
			return 0, 0, fmt.Errorf("Invalid offset")
		}
		offset = parsedOffset
	}

	if limit < 1 || limit > expenseListMaxLimit {
		return 0, 0, fmt.Errorf("Invalid limit")
	}
	if offset < 0 {
		return 0, 0, fmt.Errorf("Invalid offset")
	}
	return limit, offset, nil
}

func parseExpenseSort(c *fiber.Ctx) (string, string) {
	sortBy := strings.TrimSpace(c.Query("sort_by", "date"))
	sortDirection := strings.TrimSpace(c.Query("sort_direction", "desc"))
	validSortFields := map[string]bool{
		"date": true, "amount": true, "category": true, "description": true, "created_by": true,
	}
	if !validSortFields[sortBy] {
		sortBy = "date"
	}
	if sortDirection != "asc" && sortDirection != "desc" {
		sortDirection = "desc"
	}
	return sortBy, sortDirection
}

func parseExpenseCategoryStatus(c *fiber.Ctx) (string, error) {
	status := c.Query("status", "")
	if status == "" {
		if c.Query("active_only", "true") == "true" {
			status = "active"
		} else {
			status = "all"
		}
	}
	switch status {
	case "active", "archived", "all":
		return status, nil
	default:
		return "", fmt.Errorf("Invalid status")
	}
}

func normalizeRequiredExpenseString(value *string, fieldName string, maxLength int) error {
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	if len(trimmed) > maxLength {
		return fmt.Errorf("%s is too long", fieldName)
	}
	*value = trimmed
	return nil
}

func normalizeOptionalExpenseString(value **string, fieldName string, maxLength int) error {
	if value == nil || *value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(**value)
	if len(trimmed) > maxLength {
		return fmt.Errorf("%s is too long", fieldName)
	}
	*value = &trimmed
	return nil
}

func validateCreateExpenseRequest(req *CreateExpenseRequest) error {
	if err := normalizeRequiredExpenseString(&req.Amount, "Amount", 64); err != nil {
		return err
	}
	if err := normalizeRequiredExpenseString(&req.Description, "Description", expenseDescriptionMaxLength); err != nil {
		return err
	}
	if err := normalizeRequiredExpenseString(&req.ExpenseDate, "Expense date", len(reportDateLayout)); err != nil {
		return err
	}
	return validateExpenseOptionalText(&req.Vendor, &req.ReferenceNumber, &req.Notes)
}

func validateUpdateExpenseRequest(req *UpdateExpenseRequest) error {
	if req.Amount != nil {
		if err := normalizeRequiredExpenseString(req.Amount, "Amount", 64); err != nil {
			return err
		}
	}
	if req.Description != nil {
		if err := normalizeRequiredExpenseString(req.Description, "Description", expenseDescriptionMaxLength); err != nil {
			return err
		}
	}
	if req.ExpenseDate != nil {
		if err := normalizeRequiredExpenseString(req.ExpenseDate, "Expense date", len(reportDateLayout)); err != nil {
			return err
		}
	}
	return validateExpenseOptionalText(&req.Vendor, &req.ReferenceNumber, &req.Notes)
}

func validateExpenseOptionalText(vendor, referenceNumber, notes **string) error {
	if err := normalizeOptionalExpenseString(vendor, "Vendor", expenseVendorMaxLength); err != nil {
		return err
	}
	if err := normalizeOptionalExpenseString(referenceNumber, "Reference number", expenseReferenceMaxLength); err != nil {
		return err
	}
	if err := normalizeOptionalExpenseString(notes, "Notes", expenseNotesMaxLength); err != nil {
		return err
	}
	return nil
}

func validateExpenseCategoryName(name *string) error {
	return normalizeRequiredExpenseString(name, "Category name", expenseCategoryNameMaxLength)
}

func validateExpenseCategoryDescription(description **string) error {
	return normalizeOptionalExpenseString(description, "Category description", expenseCategoryDescriptionMax)
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
	if err != nil || !parsed.GreaterThan(decimal.Zero) || parsed.GreaterThan(maxExpenseQuantity) {
		return nil, fmt.Errorf("%s", invalidMessage)
	}

	return &parsed, nil
}

func parseRequiredAmount(value string) (decimal.Decimal, error) {
	amount, err := decimal.NewFromString(value)
	if err != nil || amount.LessThanOrEqual(decimal.Zero) || amount.GreaterThan(maxExpenseAmount) {
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

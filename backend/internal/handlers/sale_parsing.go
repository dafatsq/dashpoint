package handlers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"dashpoint/backend/internal/models"
	"dashpoint/backend/internal/repository"
)

func saleValidationAPIError(code, message string) error {
	return &apiError{status: fiber.StatusBadRequest, code: code, message: message}
}

func parseRequiredPositiveDecimal(raw, code, message string) (decimal.Decimal, error) {
	value, err := decimal.NewFromString(raw)
	if err != nil || value.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero, saleValidationAPIError(code, message)
	}
	return value, nil
}

func parseCreateSaleInput(req CreateSaleRequest) (*saleCreateInput, error) {
	if len(req.Items) == 0 {
		return nil, saleValidationAPIError("NO_ITEMS", "At least one item is required")
	}
	if len(req.Payments) == 0 {
		return nil, saleValidationAPIError("NO_PAYMENTS", "At least one payment is required")
	}

	items, err := parseSaleItems(req.Items)
	if err != nil {
		return nil, err
	}

	payments := make([]repository.CreatePaymentRequest, 0, len(req.Payments))
	for i, paymentReq := range req.Payments {
		amount, err := parseRequiredPositiveDecimal(
			paymentReq.Amount,
			"INVALID_PAYMENT_AMOUNT",
			fmt.Sprintf("Invalid payment amount at payment %d", i+1),
		)
		if err != nil {
			return nil, err
		}

		method := models.PaymentMethod(paymentReq.PaymentMethod)
		if !isValidPaymentMethod(method) {
			return nil, saleValidationAPIError("INVALID_PAYMENT_METHOD", fmt.Sprintf("Invalid payment method at payment %d", i+1))
		}

		payment := repository.CreatePaymentRequest{
			PaymentMethod: method,
			Amount:        amount,
			ReferenceNo:   paymentReq.ReferenceNo,
		}

		if paymentReq.AmountTendered != nil {
			tendered, err := decimal.NewFromString(*paymentReq.AmountTendered)
			if err != nil {
				return nil, saleValidationAPIError("INVALID_PAYMENT_AMOUNT", fmt.Sprintf("Invalid amount_tendered at payment %d", i+1))
			}
			payment.AmountTendered = &tendered
		}
		if paymentReq.ChangeGiven != nil {
			change, err := decimal.NewFromString(*paymentReq.ChangeGiven)
			if err != nil {
				return nil, saleValidationAPIError("INVALID_PAYMENT_AMOUNT", fmt.Sprintf("Invalid change_given at payment %d", i+1))
			}
			payment.ChangeGiven = &change
		}

		payments = append(payments, payment)
	}

	discountValue, err := parseOptionalSaleDecimalPointer(req.DiscountValue, "discount_value")
	if err != nil {
		return nil, err
	}
	if err := validateSaleDiscountRequest(req.DiscountType, discountValue); err != nil {
		return nil, err
	}
	var shiftID *uuid.UUID
	if req.ShiftID != nil && *req.ShiftID != "" {
		parsedShiftID, err := uuid.Parse(*req.ShiftID)
		if err != nil {
			return nil, saleValidationAPIError("INVALID_SHIFT_ID", "Invalid shift_id")
		}
		shiftID = &parsedShiftID
	}

	return &saleCreateInput{
		items:          items,
		payments:       payments,
		discountType:   req.DiscountType,
		discountValue:  discountValue,
		discountReason: req.DiscountReason,
		notes:          req.Notes,
		shiftID:        shiftID,
	}, nil
}

func parseSaleCartValidationInput(req ValidateSaleCartRequest) (*saleCartValidationInput, error) {
	if len(req.Items) == 0 {
		return nil, saleValidationAPIError("NO_ITEMS", "At least one item is required")
	}
	items, err := parseSaleItems(req.Items)
	if err != nil {
		return nil, err
	}

	var shiftID *uuid.UUID
	if req.ShiftID != nil && *req.ShiftID != "" {
		parsedShiftID, err := uuid.Parse(*req.ShiftID)
		if err != nil {
			return nil, saleValidationAPIError("INVALID_SHIFT_ID", "Invalid shift_id")
		}
		shiftID = &parsedShiftID
	}

	return &saleCartValidationInput{
		items:   items,
		shiftID: shiftID,
	}, nil
}

func parseSaleItems(itemRequests []SaleItemRequest) ([]repository.CreateSaleItemRequest, error) {
	items := make([]repository.CreateSaleItemRequest, 0, len(itemRequests))
	for i, itemReq := range itemRequests {
		productID, err := uuid.Parse(itemReq.ProductID)
		if err != nil {
			return nil, saleValidationAPIError("INVALID_PRODUCT_ID", fmt.Sprintf("Invalid product ID at item %d", i+1))
		}

		quantity, err := parseRequiredPositiveDecimal(
			itemReq.Quantity,
			"INVALID_QUANTITY",
			fmt.Sprintf("Invalid quantity at item %d", i+1),
		)
		if err != nil {
			return nil, err
		}

		unitPrice, err := decimal.NewFromString(itemReq.UnitPrice)
		if err != nil || unitPrice.LessThan(decimal.Zero) {
			return nil, saleValidationAPIError("INVALID_PRICE", fmt.Sprintf("Invalid unit price at item %d", i+1))
		}

		discountValue, err := parseOptionalSaleDecimal(itemReq.DiscountValue, fmt.Sprintf("discount_value at item %d", i+1))
		if err != nil {
			return nil, err
		}
		discountAmount, err := parseOptionalSaleDecimal(itemReq.DiscountAmount, fmt.Sprintf("discount_amount at item %d", i+1))
		if err != nil {
			return nil, err
		}

		items = append(items, repository.CreateSaleItemRequest{
			ProductID:      productID,
			Quantity:       quantity,
			UnitPrice:      unitPrice,
			DiscountType:   itemReq.DiscountType,
			DiscountValue:  discountValue,
			DiscountAmount: discountAmount,
		})
	}
	return items, nil
}

func parseSaleFilter(c *fiber.Ctx) (*repository.SaleFilter, error) {
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	filter := &repository.SaleFilter{Limit: limit, Offset: offset}
	filter.SortBy = c.Query("sort_by", "date")
	filter.SortDirection = c.Query("sort_direction", "desc")
	validSortFields := map[string]bool{
		"date": true, "invoice_no": true, "total": true, "employee": true, "status": true,
	}
	if !validSortFields[filter.SortBy] || (filter.SortDirection != "asc" && filter.SortDirection != "desc") {
		return nil, saleValidationAPIError("INVALID_SORT", "Invalid sort parameters")
	}

	if employeeID := c.Query("employee_id"); employeeID != "" {
		id, err := uuid.Parse(employeeID)
		if err != nil {
			return nil, saleValidationAPIError("INVALID_EMPLOYEE_ID", "Invalid employee ID format")
		}
		filter.EmployeeID = &id
	}
	if shiftID := c.Query("shift_id"); shiftID != "" {
		id, err := uuid.Parse(shiftID)
		if err != nil {
			return nil, saleValidationAPIError("INVALID_SHIFT_ID", "Invalid shift ID format")
		}
		filter.ShiftID = &id
	}
	if status := c.Query("status"); status != "" {
		filter.Status = &status
	}
	if startStr := c.Query("from"); startStr != "" {
		startDate, err := parseReportDay(startStr, "from")
		if err != nil {
			return nil, saleValidationAPIError("INVALID_START_DATE", "Invalid from format. Use YYYY-MM-DD")
		}
		startDate = reportDayStart(startDate)
		filter.StartDate = &startDate
	}
	if endStr := c.Query("to"); endStr != "" {
		endDate, err := parseReportDay(endStr, "to")
		if err != nil {
			return nil, saleValidationAPIError("INVALID_END_DATE", "Invalid to format. Use YYYY-MM-DD")
		}
		exclusiveEnd := reportDayStart(endDate).Add(24 * time.Hour)
		filter.EndDate = &exclusiveEnd
	}
	if invoiceSearch := c.Query("invoice_no"); invoiceSearch != "" {
		filter.InvoiceSearch = &invoiceSearch
	}

	return filter, nil
}

func parseOptionalSaleDecimal(raw string, field string) (decimal.Decimal, error) {
	if raw == "" {
		return decimal.Zero, nil
	}
	value, err := decimal.NewFromString(raw)
	if err != nil {
		return decimal.Zero, saleValidationAPIError("VALIDATION_ERROR", fmt.Sprintf("Invalid %s", field))
	}
	return value, nil
}

func parseOptionalSaleDecimalPointer(raw *string, field string) (*decimal.Decimal, error) {
	if raw == nil || *raw == "" {
		return nil, nil
	}
	value, err := decimal.NewFromString(*raw)
	if err != nil {
		return nil, saleValidationAPIError("VALIDATION_ERROR", fmt.Sprintf("Invalid %s", field))
	}
	return &value, nil
}

func isValidPaymentMethod(method models.PaymentMethod) bool {
	switch method {
	case models.PaymentMethodCash,
		models.PaymentMethodCard,
		models.PaymentMethodTransfer,
		models.PaymentMethodQRIS,
		models.PaymentMethodCredit,
		models.PaymentMethodVoucher,
		models.PaymentMethodOther:
		return true
	default:
		return false
	}
}

func validateSaleDiscountRequest(discountType *string, discountValue *decimal.Decimal) error {
	if discountType == nil && discountValue == nil {
		return nil
	}
	if discountType == nil || discountValue == nil {
		return saleValidationAPIError("INVALID_DISCOUNT", "Discount type and value must be provided together")
	}
	if discountValue.LessThan(decimal.Zero) {
		return saleValidationAPIError("INVALID_DISCOUNT", "Discount value cannot be negative")
	}

	switch *discountType {
	case "fixed":
		return nil
	case "percentage":
		if discountValue.GreaterThan(decimal.NewFromInt(100)) {
			return saleValidationAPIError("INVALID_DISCOUNT", "Discount percentage cannot exceed 100")
		}
		return nil
	default:
		return saleValidationAPIError("INVALID_DISCOUNT", "Invalid discount type")
	}
}

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

func parseCreateSaleInput(req CreateSaleRequest) (*saleCreateInput, error) {
	if len(req.Items) == 0 {
		return nil, &apiError{status: fiber.StatusBadRequest, code: "NO_ITEMS", message: "At least one item is required"}
	}
	if len(req.Payments) == 0 {
		return nil, &apiError{status: fiber.StatusBadRequest, code: "NO_PAYMENTS", message: "At least one payment is required"}
	}

	items := make([]repository.CreateSaleItemRequest, 0, len(req.Items))
	for i, itemReq := range req.Items {
		productID, err := uuid.Parse(itemReq.ProductID)
		if err != nil {
			return nil, &apiError{status: fiber.StatusBadRequest, code: "INVALID_PRODUCT_ID", message: fmt.Sprintf("Invalid product ID at item %d", i+1)}
		}

		quantity, err := decimal.NewFromString(itemReq.Quantity)
		if err != nil || quantity.LessThanOrEqual(decimal.Zero) {
			return nil, &apiError{status: fiber.StatusBadRequest, code: "INVALID_QUANTITY", message: fmt.Sprintf("Invalid quantity at item %d", i+1)}
		}

		unitPrice, err := decimal.NewFromString(itemReq.UnitPrice)
		if err != nil || unitPrice.LessThan(decimal.Zero) {
			return nil, &apiError{status: fiber.StatusBadRequest, code: "INVALID_PRICE", message: fmt.Sprintf("Invalid unit price at item %d", i+1)}
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

	payments := make([]repository.CreatePaymentRequest, 0, len(req.Payments))
	for i, paymentReq := range req.Payments {
		amount, err := decimal.NewFromString(paymentReq.Amount)
		if err != nil || amount.LessThanOrEqual(decimal.Zero) {
			return nil, &apiError{status: fiber.StatusBadRequest, code: "INVALID_PAYMENT_AMOUNT", message: fmt.Sprintf("Invalid payment amount at payment %d", i+1)}
		}

		method := models.PaymentMethod(paymentReq.PaymentMethod)
		if !isValidPaymentMethod(method) {
			return nil, &apiError{status: fiber.StatusBadRequest, code: "INVALID_PAYMENT_METHOD", message: fmt.Sprintf("Invalid payment method at payment %d", i+1)}
		}

		payment := repository.CreatePaymentRequest{
			PaymentMethod: method,
			Amount:        amount,
			CardType:      paymentReq.CardType,
			CardLastFour:  paymentReq.CardLastFour,
			ReferenceNo:   paymentReq.ReferenceNo,
			BankName:      paymentReq.BankName,
			AccountNo:     paymentReq.AccountNo,
			VoucherCode:   paymentReq.VoucherCode,
			Notes:         paymentReq.Notes,
		}

		if paymentReq.AmountTendered != nil {
			tendered, err := decimal.NewFromString(*paymentReq.AmountTendered)
			if err != nil {
				return nil, &apiError{status: fiber.StatusBadRequest, code: "INVALID_PAYMENT_AMOUNT", message: fmt.Sprintf("Invalid amount_tendered at payment %d", i+1)}
			}
			payment.AmountTendered = &tendered
		}
		if paymentReq.ChangeGiven != nil {
			change, err := decimal.NewFromString(*paymentReq.ChangeGiven)
			if err != nil {
				return nil, &apiError{status: fiber.StatusBadRequest, code: "INVALID_PAYMENT_AMOUNT", message: fmt.Sprintf("Invalid change_given at payment %d", i+1)}
			}
			payment.ChangeGiven = &change
		}

		payments = append(payments, payment)
	}

	discountValue, err := parseOptionalSaleDecimalPointer(req.DiscountValue, "discount_value")
	if err != nil {
		return nil, err
	}

	return &saleCreateInput{
		items:          items,
		payments:       payments,
		customerName:   req.CustomerName,
		customerPhone:  req.CustomerPhone,
		discountType:   req.DiscountType,
		discountValue:  discountValue,
		discountReason: req.DiscountReason,
		notes:          req.Notes,
	}, nil
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

	if employeeID := c.Query("employee_id"); employeeID != "" {
		id, err := uuid.Parse(employeeID)
		if err != nil {
			return nil, &apiError{status: fiber.StatusBadRequest, code: "INVALID_EMPLOYEE_ID", message: "Invalid employee ID format"}
		}
		filter.EmployeeID = &id
	}
	if shiftID := c.Query("shift_id"); shiftID != "" {
		id, err := uuid.Parse(shiftID)
		if err != nil {
			return nil, &apiError{status: fiber.StatusBadRequest, code: "INVALID_SHIFT_ID", message: "Invalid shift ID format"}
		}
		filter.ShiftID = &id
	}
	if status := c.Query("status"); status != "" {
		filter.Status = &status
	}
	if startStr := c.Query("start_date"); startStr != "" {
		startDate, err := time.Parse("2006-01-02", startStr)
		if err != nil {
			return nil, &apiError{status: fiber.StatusBadRequest, code: "INVALID_START_DATE", message: "Invalid start_date format. Use YYYY-MM-DD"}
		}
		filter.StartDate = &startDate
	}
	if endStr := c.Query("end_date"); endStr != "" {
		endDate, err := time.Parse("2006-01-02", endStr)
		if err != nil {
			return nil, &apiError{status: fiber.StatusBadRequest, code: "INVALID_END_DATE", message: "Invalid end_date format. Use YYYY-MM-DD"}
		}
		endOfDay := endDate.Add(24*time.Hour - time.Second)
		filter.EndDate = &endOfDay
	}

	return filter, nil
}

func parseOptionalSaleDecimal(raw string, field string) (decimal.Decimal, error) {
	if raw == "" {
		return decimal.Zero, nil
	}
	value, err := decimal.NewFromString(raw)
	if err != nil {
		return decimal.Zero, &apiError{status: fiber.StatusBadRequest, code: "VALIDATION_ERROR", message: fmt.Sprintf("Invalid %s", field)}
	}
	return value, nil
}

func parseOptionalSaleDecimalPointer(raw *string, field string) (*decimal.Decimal, error) {
	if raw == nil || *raw == "" {
		return nil, nil
	}
	value, err := decimal.NewFromString(*raw)
	if err != nil {
		return nil, &apiError{status: fiber.StatusBadRequest, code: "VALIDATION_ERROR", message: fmt.Sprintf("Invalid %s", field)}
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

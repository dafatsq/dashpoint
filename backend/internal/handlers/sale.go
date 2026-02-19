package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"

	"dashpoint/backend/internal/audit"
	"dashpoint/backend/internal/middleware"
	"dashpoint/backend/internal/models"
	"dashpoint/backend/internal/repository"
)

// SaleHandler handles sale endpoints
type SaleHandler struct {
	saleRepo  *repository.SaleRepository
	shiftRepo *repository.ShiftRepository
}

// NewSaleHandler creates a new sale handler
func NewSaleHandler(saleRepo *repository.SaleRepository, shiftRepo *repository.ShiftRepository) *SaleHandler {
	return &SaleHandler{
		saleRepo:  saleRepo,
		shiftRepo: shiftRepo,
	}
}

// SaleItemRequest represents a sale item in the request
type SaleItemRequest struct {
	ProductID      string  `json:"product_id"`
	Quantity       string  `json:"quantity"`
	UnitPrice      string  `json:"unit_price"`
	DiscountType   *string `json:"discount_type"`
	DiscountValue  string  `json:"discount_value"`
	DiscountAmount string  `json:"discount_amount"`
}

// PaymentRequest represents a payment in the request
type PaymentRequest struct {
	PaymentMethod  string  `json:"payment_method"`
	Amount         string  `json:"amount"`
	AmountTendered *string `json:"amount_tendered"`
	ChangeGiven    *string `json:"change_given"`
	CardType       *string `json:"card_type"`
	CardLastFour   *string `json:"card_last_four"`
	ReferenceNo    *string `json:"reference_no"`
	BankName       *string `json:"bank_name"`
	AccountNo      *string `json:"account_no"`
	VoucherCode    *string `json:"voucher_code"`
	Notes          *string `json:"notes"`
}

// CreateSaleRequest represents the request to create a sale
type CreateSaleRequest struct {
	Items          []SaleItemRequest `json:"items"`
	Payments       []PaymentRequest  `json:"payments"`
	CustomerName   *string           `json:"customer_name"`
	CustomerPhone  *string           `json:"customer_phone"`
	DiscountType   *string           `json:"discount_type"`
	DiscountValue  *string           `json:"discount_value"`
	DiscountReason *string           `json:"discount_reason"`
	Notes          *string           `json:"notes"`
}

// VoidSaleRequest represents the request to void a sale
type VoidSaleRequest struct {
	Reason string `json:"reason"`
}

// CreateSale handles POST /api/v1/sales
func (h *SaleHandler) CreateSale(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	var req CreateSaleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request body",
		})
	}

	// Validate items
	if len(req.Items) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "NO_ITEMS",
			"message": "At least one item is required",
		})
	}

	// Validate payments
	if len(req.Payments) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "NO_PAYMENTS",
			"message": "At least one payment is required",
		})
	}

	// Get current shift (required for cashiers)
	var shiftID *uuid.UUID
	shift, _ := h.shiftRepo.GetOpenShiftByEmployee(c.Context(), userID)
	if shift != nil {
		shiftID = &shift.ID
	} else {
		// Cashiers must have an open shift to create sales
		roleName := middleware.GetRoleName(c)
		if roleName == "cashier" {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"code":    "NO_OPEN_SHIFT",
				"message": "You must start a shift before processing sales",
			})
		}
	}

	// Parse items
	var items []repository.CreateSaleItemRequest
	for i, itemReq := range req.Items {
		productID, err := uuid.Parse(itemReq.ProductID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code":    "INVALID_PRODUCT_ID",
				"message": "Invalid product ID at item " + strconv.Itoa(i+1),
			})
		}

		quantity, err := decimal.NewFromString(itemReq.Quantity)
		if err != nil || quantity.LessThanOrEqual(decimal.Zero) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code":    "INVALID_QUANTITY",
				"message": "Invalid quantity at item " + strconv.Itoa(i+1),
			})
		}

		unitPrice, err := decimal.NewFromString(itemReq.UnitPrice)
		if err != nil || unitPrice.LessThan(decimal.Zero) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code":    "INVALID_PRICE",
				"message": "Invalid unit price at item " + strconv.Itoa(i+1),
			})
		}

		discountValue := decimal.Zero
		if itemReq.DiscountValue != "" {
			discountValue, _ = decimal.NewFromString(itemReq.DiscountValue)
		}

		discountAmount := decimal.Zero
		if itemReq.DiscountAmount != "" {
			discountAmount, _ = decimal.NewFromString(itemReq.DiscountAmount)
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

	// Parse payments
	var payments []repository.CreatePaymentRequest
	for i, paymentReq := range req.Payments {
		amount, err := decimal.NewFromString(paymentReq.Amount)
		if err != nil || amount.LessThanOrEqual(decimal.Zero) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code":    "INVALID_PAYMENT_AMOUNT",
				"message": "Invalid payment amount at payment " + strconv.Itoa(i+1),
			})
		}

		// Validate payment method
		method := models.PaymentMethod(paymentReq.PaymentMethod)
		validMethods := map[models.PaymentMethod]bool{
			models.PaymentMethodCash:     true,
			models.PaymentMethodCard:     true,
			models.PaymentMethodTransfer: true,
			models.PaymentMethodQRIS:     true,
			models.PaymentMethodCredit:   true,
			models.PaymentMethodVoucher:  true,
			models.PaymentMethodOther:    true,
		}
		if !validMethods[method] {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code":    "INVALID_PAYMENT_METHOD",
				"message": "Invalid payment method at payment " + strconv.Itoa(i+1),
			})
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

		// Parse cash specific fields
		if paymentReq.AmountTendered != nil {
			tendered, _ := decimal.NewFromString(*paymentReq.AmountTendered)
			payment.AmountTendered = &tendered
		}
		if paymentReq.ChangeGiven != nil {
			change, _ := decimal.NewFromString(*paymentReq.ChangeGiven)
			payment.ChangeGiven = &change
		}

		payments = append(payments, payment)
	}

	// Parse sale-level discount
	var discountValue *decimal.Decimal
	if req.DiscountValue != nil && *req.DiscountValue != "" {
		dv, err := decimal.NewFromString(*req.DiscountValue)
		if err == nil {
			discountValue = &dv
		}
	}

	// Create sale
	createReq := &repository.CreateSaleRequest{
		Items:          items,
		Payments:       payments,
		EmployeeID:     userID,
		ShiftID:        shiftID,
		CustomerName:   req.CustomerName,
		CustomerPhone:  req.CustomerPhone,
		DiscountType:   req.DiscountType,
		DiscountValue:  discountValue,
		DiscountReason: req.DiscountReason,
		Notes:          req.Notes,
	}

	sale, err := h.saleRepo.Create(c.Context(), createReq)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create sale")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "SALE_FAILED",
			"message": err.Error(),
		})
	}


	// Audit log
	audit.LogFromFiber(c, models.AuditActionSaleCreate, models.AuditEntitySale, sale.ID.String(), "Created sale")

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Sale completed successfully",
		"sale":    h.toSaleResponse(sale),
	})
}

// GetSale handles GET /api/v1/sales/:id
func (h *SaleHandler) GetSale(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_ID",
			"message": "Invalid sale ID format",
		})
	}

	sale, err := h.saleRepo.GetByID(c.Context(), id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get sale")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to retrieve sale",
		})
	}

	if sale == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"code":    "NOT_FOUND",
			"message": "Sale not found",
		})
	}

	return c.JSON(fiber.Map{
		"sale": h.toSaleResponse(sale),
	})
}

// GetSaleByInvoice handles GET /api/v1/sales/invoice/:invoiceNo
func (h *SaleHandler) GetSaleByInvoice(c *fiber.Ctx) error {
	invoiceNo := c.Params("invoiceNo")

	sale, err := h.saleRepo.GetByInvoiceNo(c.Context(), invoiceNo)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get sale by invoice")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to retrieve sale",
		})
	}

	if sale == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"code":    "NOT_FOUND",
			"message": "Sale not found",
		})
	}

	return c.JSON(fiber.Map{
		"sale": h.toSaleResponse(sale),
	})
}

// ListSales handles GET /api/v1/sales
func (h *SaleHandler) ListSales(c *fiber.Ctx) error {
	// Parse query parameters
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	var employeeID, shiftID *uuid.UUID
	var status *string
	var startDate, endDate *time.Time

	if empIDStr := c.Query("employee_id"); empIDStr != "" {
		id, err := uuid.Parse(empIDStr)
		if err == nil {
			employeeID = &id
		}
	}

	if shiftIDStr := c.Query("shift_id"); shiftIDStr != "" {
		id, err := uuid.Parse(shiftIDStr)
		if err == nil {
			shiftID = &id
		}
	}

	if statusStr := c.Query("status"); statusStr != "" {
		status = &statusStr
	}

	if startStr := c.Query("start_date"); startStr != "" {
		t, err := time.Parse("2006-01-02", startStr)
		if err == nil {
			startDate = &t
		}
	}

	if endStr := c.Query("end_date"); endStr != "" {
		t, err := time.Parse("2006-01-02", endStr)
		if err == nil {
			// Set to end of day
			endOfDay := t.Add(24*time.Hour - time.Second)
			endDate = &endOfDay
		}
	}

	// For cashiers, only show their own sales
	roleName := middleware.GetRoleName(c)
	if roleName == "cashier" {
		userID := middleware.GetUserID(c)
		employeeID = &userID
	}

	sales, total, err := h.saleRepo.List(c.Context(), employeeID, shiftID, status, startDate, endDate, limit, offset)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list sales")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to retrieve sales",
		})
	}

	// Convert to response format
	var salesResponse []fiber.Map
	for _, s := range sales {
		salesResponse = append(salesResponse, h.toSaleListResponse(&s))
	}

	return c.JSON(fiber.Map{
		"sales":  salesResponse,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// VoidSale handles POST /api/v1/sales/:id/void
func (h *SaleHandler) VoidSale(c *fiber.Ctx) error {
	// Check permission (owner/manager only)
	roleName := middleware.GetRoleName(c)
	if roleName != "owner" && roleName != "manager" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"code":    "FORBIDDEN",
			"message": "Only owner or manager can void sales",
		})
	}

	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_ID",
			"message": "Invalid sale ID format",
		})
	}

	var req VoidSaleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_REQUEST",
			"message": "Invalid request body",
		})
	}

	if req.Reason == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "REASON_REQUIRED",
			"message": "Reason is required to void a sale",
		})
	}

	userID := middleware.GetUserID(c)

	if err := h.saleRepo.VoidSale(c.Context(), id, userID, req.Reason); err != nil {
		log.Error().Err(err).Msg("Failed to void sale")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "VOID_FAILED",
			"message": err.Error(),
		})
	}

	// Fetch the updated sale for the response
	sale, _ := h.saleRepo.GetByID(c.Context(), id)

	// Audit log
	audit.LogFromFiber(c, models.AuditActionSaleVoid, models.AuditEntitySale, id.String(), "Voided sale: "+req.Reason)

	return c.JSON(fiber.Map{
		"message": "Sale voided successfully",
		"sale":    h.toSaleResponse(sale),
	})
}

// GetDailySummary handles GET /api/v1/sales/summary/daily
func (h *SaleHandler) GetDailySummary(c *fiber.Ctx) error {
	dateStr := c.Query("date", time.Now().Format("2006-01-02"))
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "INVALID_DATE",
			"message": "Invalid date format. Use YYYY-MM-DD",
		})
	}

	summary, err := h.saleRepo.GetDailySummary(c.Context(), date)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get daily summary")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to retrieve daily summary",
		})
	}

	return c.JSON(fiber.Map{
		"summary": summary,
	})
}

// Helper functions

func (h *SaleHandler) toSaleResponse(s *models.Sale) fiber.Map {
	response := fiber.Map{
		"id":              s.ID.String(),
		"invoice_no":      s.InvoiceNo,
		"subtotal":        s.Subtotal.String(),
		"tax_amount":      s.TaxAmount.String(),
		"discount_amount": s.DiscountAmount.String(),
		"total_amount":    s.TotalAmount.String(),
		"item_count":      s.ItemCount,
		"payment_status":  s.PaymentStatus,
		"amount_paid":     s.AmountPaid.String(),
		"change_amount":   s.ChangeAmount.String(),
		"employee_id":     s.EmployeeID.String(),
		"status":          s.Status,
		"created_at":      s.CreatedAt.Format(time.RFC3339),
		"updated_at":      s.UpdatedAt.Format(time.RFC3339),
	}

	if s.EmployeeName != nil {
		response["employee_name"] = *s.EmployeeName
	}
	if s.ShiftID != nil {
		response["shift_id"] = s.ShiftID.String()
	}
	if s.CustomerName != nil {
		response["customer_name"] = *s.CustomerName
	}
	if s.CustomerPhone != nil {
		response["customer_phone"] = *s.CustomerPhone
	}
	if s.DiscountType != nil {
		response["discount_type"] = *s.DiscountType
	}
	if s.DiscountValue != nil {
		response["discount_value"] = s.DiscountValue.String()
	}
	if s.DiscountReason != nil {
		response["discount_reason"] = *s.DiscountReason
	}
	if s.Notes != nil {
		response["notes"] = *s.Notes
	}
	if s.VoidedAt != nil {
		response["voided_at"] = s.VoidedAt.Format(time.RFC3339)
	}
	if s.VoidedBy != nil {
		response["voided_by"] = s.VoidedBy.String()
	}
	if s.VoidReason != nil {
		response["void_reason"] = *s.VoidReason
	}

	// Items
	if len(s.Items) > 0 {
		var items []fiber.Map
		for _, item := range s.Items {
			itemMap := fiber.Map{
				"id":              item.ID.String(),
				"product_id":      item.ProductID.String(),
				"product_name":    item.ProductName,
				"quantity":        item.Quantity.String(),
				"unit_price":      item.UnitPrice.String(),
				"discount_amount": item.DiscountAmount.String(),
				"tax_rate":        item.TaxRate.String(),
				"tax_amount":      item.TaxAmount.String(),
				"subtotal":        item.Subtotal.String(),
				"total":           item.Total.String(),
			}
			if item.ProductSKU != nil {
				itemMap["product_sku"] = *item.ProductSKU
			}
			if item.ProductBarcode != nil {
				itemMap["product_barcode"] = *item.ProductBarcode
			}
			items = append(items, itemMap)
		}
		response["items"] = items
	}

	// Payments
	if len(s.Payments) > 0 {
		var payments []fiber.Map
		for _, payment := range s.Payments {
			paymentMap := fiber.Map{
				"id":             payment.ID.String(),
				"payment_method": payment.PaymentMethod,
				"amount":         payment.Amount.String(),
				"status":         payment.Status,
			}
			if payment.AmountTendered != nil {
				paymentMap["amount_tendered"] = payment.AmountTendered.String()
			}
			if payment.ChangeGiven != nil {
				paymentMap["change_given"] = payment.ChangeGiven.String()
			}
			if payment.ReferenceNo != nil {
				paymentMap["reference_no"] = *payment.ReferenceNo
			}
			payments = append(payments, paymentMap)
		}
		response["payments"] = payments
	}

	return response
}

func (h *SaleHandler) toSaleListResponse(s *models.Sale) fiber.Map {
	response := fiber.Map{
		"id":             s.ID.String(),
		"invoice_no":     s.InvoiceNo,
		"total_amount":   s.TotalAmount.String(),
		"item_count":     s.ItemCount,
		"payment_status": s.PaymentStatus,
		"status":         s.Status,
		"created_at":     s.CreatedAt.Format(time.RFC3339),
	}

	if s.EmployeeName != nil {
		response["employee_name"] = *s.EmployeeName
	}
	if s.CustomerName != nil {
		response["customer_name"] = *s.CustomerName
	}

	// Include primary payment method if available
	if len(s.Payments) > 0 {
		var payments []fiber.Map
		for _, payment := range s.Payments {
			paymentMap := fiber.Map{
				"payment_method": payment.PaymentMethod,
			}
			payments = append(payments, paymentMap)
		}
		response["payments"] = payments
	}

	return response
}

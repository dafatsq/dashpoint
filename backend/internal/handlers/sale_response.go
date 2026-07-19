package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"dashpoint/backend/internal/models"
)

func saleItemResponse(item models.SaleItem) fiber.Map {
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
	return itemMap
}

func paymentResponse(payment models.Payment) fiber.Map {
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
	return paymentMap
}

func (h *SaleHandler) toSaleResponse(s *models.Sale) fiber.Map {
	if s == nil {
		return nil
	}

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
	if len(s.Items) > 0 {
		items := make([]fiber.Map, 0, len(s.Items))
		for _, item := range s.Items {
			items = append(items, saleItemResponse(item))
		}
		response["items"] = items
	}
	if len(s.Payments) > 0 {
		payments := make([]fiber.Map, 0, len(s.Payments))
		for _, payment := range s.Payments {
			payments = append(payments, paymentResponse(payment))
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
		"employee_id":    s.EmployeeID.String(),
		"created_at":     s.CreatedAt.Format(time.RFC3339),
	}

	if s.EmployeeName != nil {
		response["employee_name"] = *s.EmployeeName
	}
	if len(s.Payments) > 0 {
		payments := make([]fiber.Map, 0, len(s.Payments))
		for _, payment := range s.Payments {
			payments = append(payments, fiber.Map{"payment_method": payment.PaymentMethod})
		}
		response["payments"] = payments
	}

	return response
}

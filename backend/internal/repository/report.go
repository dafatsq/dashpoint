package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

// ReportRepository handles report database operations.
type ReportRepository struct {
	pool *pgxpool.Pool
}

// NewReportRepository creates a new report repository.
func NewReportRepository(pool *pgxpool.Pool) *ReportRepository {
	return &ReportRepository{pool: pool}
}

type DailySalesReport struct {
	Date             string            `json:"date"`
	TotalSales       decimal.Decimal   `json:"total_sales"`
	TotalTax         decimal.Decimal   `json:"total_tax"`
	TotalDiscount    decimal.Decimal   `json:"total_discount"`
	TotalAmount      decimal.Decimal   `json:"total_amount"`
	TransactionCount int               `json:"transaction_count"`
	ItemCount        int               `json:"item_count"`
	VoidedCount      int               `json:"voided_count"`
	VoidedAmount     decimal.Decimal   `json:"voided_amount"`
	PaymentBreakdown map[string]string `json:"payment_breakdown"`
	HourlySales      []HourlySales     `json:"hourly_sales,omitempty"`
}

type HourlySales struct {
	Hour         int             `json:"hour"`
	Sales        decimal.Decimal `json:"sales"`
	Transactions int             `json:"transactions"`
}

type TopSellerItem struct {
	ProductID    string          `json:"product_id"`
	ProductName  string          `json:"product_name"`
	ProductSKU   *string         `json:"product_sku,omitempty"`
	CategoryName *string         `json:"category_name,omitempty"`
	QuantitySold decimal.Decimal `json:"quantity_sold"`
	TotalRevenue decimal.Decimal `json:"total_revenue"`
	TotalProfit  decimal.Decimal `json:"total_profit"`
}

type InventoryValuation struct {
	TotalProducts    int             `json:"total_products"`
	TotalQuantity    decimal.Decimal `json:"total_quantity"`
	TotalCostValue   decimal.Decimal `json:"total_cost_value"`
	TotalRetailValue decimal.Decimal `json:"total_retail_value"`
	PotentialProfit  decimal.Decimal `json:"potential_profit"`
	Items            []InventoryItem `json:"items,omitempty"`
}

type InventoryItem struct {
	ProductID    string          `json:"product_id"`
	ProductName  string          `json:"product_name"`
	ProductSKU   *string         `json:"product_sku,omitempty"`
	CategoryName *string         `json:"category_name,omitempty"`
	Quantity     decimal.Decimal `json:"quantity"`
	CostPrice    decimal.Decimal `json:"cost_price"`
	SellPrice    decimal.Decimal `json:"sell_price"`
	CostValue    decimal.Decimal `json:"cost_value"`
	RetailValue  decimal.Decimal `json:"retail_value"`
}

type SalesReportItem struct {
	InvoiceNo     string          `json:"invoice_no"`
	Date          string          `json:"date"`
	Time          string          `json:"time"`
	EmployeeName  string          `json:"employee_name"`
	ItemCount     int             `json:"item_count"`
	Subtotal      decimal.Decimal `json:"subtotal"`
	Tax           decimal.Decimal `json:"tax"`
	Discount      decimal.Decimal `json:"discount"`
	Total         decimal.Decimal `json:"total"`
	PaymentMethod string          `json:"payment_method"`
	Status        string          `json:"status"`
}

type CashReport struct {
	Date            string          `json:"date"`
	OpeningCash     decimal.Decimal `json:"opening_cash"`
	CashSales       decimal.Decimal `json:"cash_sales"`
	CashVoidedSales decimal.Decimal `json:"cash_voided_sales"`
	PayInTotal      decimal.Decimal `json:"pay_in_total"`
	PayOutTotal     decimal.Decimal `json:"pay_out_total"`
	ExpectedCash    decimal.Decimal `json:"expected_cash"`
	ActualCash      decimal.Decimal `json:"actual_cash"`
	Difference      decimal.Decimal `json:"difference"`
	ShiftCount      int             `json:"shift_count"`
}

type SalesRangeSummary struct {
	TotalTransactions int             `json:"total_transactions"`
	TotalItems        int             `json:"total_items"`
	TotalAmount       decimal.Decimal `json:"total_amount"`
	TotalTax          decimal.Decimal `json:"total_tax"`
	TotalDiscount     decimal.Decimal `json:"total_discount"`
}

func inclusiveEndDate(endDate time.Time) time.Time {
	return endDate.Add(24 * time.Hour)
}

func inventoryWhereClause(categoryID *uuid.UUID) (string, []interface{}) {
	whereClause := "WHERE p.is_active = true"
	if categoryID == nil {
		return whereClause, nil
	}
	return whereClause + " AND p.category_id = $1", []interface{}{*categoryID}
}

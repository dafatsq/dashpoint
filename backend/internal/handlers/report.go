package handlers

import (
	"context"
	"time"

	"github.com/google/uuid"

	"dashpoint/backend/internal/repository"
)

type reportStore interface {
	GetDailySalesReport(context.Context, time.Time) (*repository.DailySalesReport, error)
	GetSalesRangeReport(context.Context, time.Time, time.Time) ([]repository.DailySalesReport, error)
	GetSalesRangeSummary(context.Context, time.Time, time.Time) (*repository.SalesRangeSummary, error)
	GetTopSellers(context.Context, time.Time, time.Time, int) ([]repository.TopSellerItem, error)
	GetInventoryValuation(context.Context, *uuid.UUID, bool) (*repository.InventoryValuation, error)
	GetCashReport(context.Context, time.Time, time.Time) (*repository.CashReport, error)
	GetEmployeeSalesReport(context.Context, time.Time, time.Time) ([]map[string]interface{}, error)
	GetCategorySalesReport(context.Context, time.Time, time.Time) ([]map[string]interface{}, error)
	GetSalesForExport(context.Context, time.Time, time.Time) ([]repository.SalesReportItem, error)
}

// ReportHandler handles report endpoints.
type ReportHandler struct {
	reportRepo reportStore
}

// NewReportHandler creates a new report handler.
func NewReportHandler(reportRepo reportStore) *ReportHandler {
	return &ReportHandler{reportRepo: reportRepo}
}

package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"dashpoint/backend/internal/repository"
)

type fakeReportStore struct {
	getDailySalesReportFunc   func(context.Context, time.Time) (*repository.DailySalesReport, error)
	getSalesRangeReportFunc   func(context.Context, time.Time, time.Time) ([]repository.DailySalesReport, error)
	getSalesRangeSummaryFunc  func(context.Context, time.Time, time.Time) (*repository.SalesRangeSummary, error)
	getTopSellersFunc         func(context.Context, time.Time, time.Time, int) ([]repository.TopSellerItem, error)
	getInventoryValuationFunc func(context.Context, *uuid.UUID, bool) (*repository.InventoryValuation, error)
	getCashReportFunc         func(context.Context, time.Time, time.Time) (*repository.CashReport, error)
	getEmployeeSalesReportFun func(context.Context, time.Time, time.Time) ([]map[string]interface{}, error)
	getCategorySalesReportFun func(context.Context, time.Time, time.Time) ([]map[string]interface{}, error)
	getSalesForExportFunc     func(context.Context, time.Time, time.Time) ([]repository.SalesReportItem, error)
}

func (f *fakeReportStore) GetDailySalesReport(ctx context.Context, date time.Time) (*repository.DailySalesReport, error) {
	if f.getDailySalesReportFunc != nil {
		return f.getDailySalesReportFunc(ctx, date)
	}
	return nil, nil
}

func (f *fakeReportStore) GetSalesRangeReport(ctx context.Context, startDate, endDate time.Time) ([]repository.DailySalesReport, error) {
	if f.getSalesRangeReportFunc != nil {
		return f.getSalesRangeReportFunc(ctx, startDate, endDate)
	}
	return nil, nil
}

func (f *fakeReportStore) GetSalesRangeSummary(ctx context.Context, startDate, endDate time.Time) (*repository.SalesRangeSummary, error) {
	if f.getSalesRangeSummaryFunc != nil {
		return f.getSalesRangeSummaryFunc(ctx, startDate, endDate)
	}
	return nil, nil
}

func (f *fakeReportStore) GetTopSellers(ctx context.Context, startDate, endDate time.Time, limit int) ([]repository.TopSellerItem, error) {
	if f.getTopSellersFunc != nil {
		return f.getTopSellersFunc(ctx, startDate, endDate, limit)
	}
	return nil, nil
}

func (f *fakeReportStore) GetInventoryValuation(ctx context.Context, categoryID *uuid.UUID, includeItems bool) (*repository.InventoryValuation, error) {
	if f.getInventoryValuationFunc != nil {
		return f.getInventoryValuationFunc(ctx, categoryID, includeItems)
	}
	return &repository.InventoryValuation{}, nil
}

func (f *fakeReportStore) GetCashReport(ctx context.Context, startDate, endDate time.Time) (*repository.CashReport, error) {
	if f.getCashReportFunc != nil {
		return f.getCashReportFunc(ctx, startDate, endDate)
	}
	return &repository.CashReport{}, nil
}

func (f *fakeReportStore) GetEmployeeSalesReport(ctx context.Context, startDate, endDate time.Time) ([]map[string]interface{}, error) {
	if f.getEmployeeSalesReportFun != nil {
		return f.getEmployeeSalesReportFun(ctx, startDate, endDate)
	}
	return nil, nil
}

func (f *fakeReportStore) GetCategorySalesReport(ctx context.Context, startDate, endDate time.Time) ([]map[string]interface{}, error) {
	if f.getCategorySalesReportFun != nil {
		return f.getCategorySalesReportFun(ctx, startDate, endDate)
	}
	return nil, nil
}

func (f *fakeReportStore) GetSalesForExport(ctx context.Context, startDate, endDate time.Time) ([]repository.SalesReportItem, error) {
	if f.getSalesForExportFunc != nil {
		return f.getSalesForExportFunc(ctx, startDate, endDate)
	}
	return nil, nil
}

func TestGetInventoryValuationRejectsInvalidCategoryID(t *testing.T) {
	handler := NewReportHandler(&fakeReportStore{})
	app := fiber.New()
	app.Get("/reports/inventory", handler.GetInventoryValuation)

	req := httptest.NewRequest(http.MethodGet, "/reports/inventory?category_id=bad-uuid", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestGetCashReportUsesExclusiveNextDayUpperBound(t *testing.T) {
	store := &fakeReportStore{
		getCashReportFunc: func(_ context.Context, startDate, endDate time.Time) (*repository.CashReport, error) {
			expectedStart := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
			expectedEnd := expectedStart.Add(24 * time.Hour)
			if !startDate.Equal(expectedStart) {
				t.Fatalf("unexpected startDate: %v", startDate)
			}
			if !endDate.Equal(expectedEnd) {
				t.Fatalf("expected exclusive next-day endDate %v, got %v", expectedEnd, endDate)
			}
			return &repository.CashReport{}, nil
		},
	}
	handler := NewReportHandler(store)
	app := fiber.New()
	app.Get("/reports/cash", handler.GetCashReport)

	req := httptest.NewRequest(http.MethodGet, "/reports/cash?start_date=2026-05-01&end_date=2026-05-01", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

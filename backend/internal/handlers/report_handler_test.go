package handlers

import (
	"bytes"
	"context"
	"encoding/csv"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"dashpoint/backend/internal/audit"
	"dashpoint/backend/internal/models"
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

func TestGetCashReportUsesJakartaLocalDateBounds(t *testing.T) {
	jakarta := time.FixedZone("WIB", 7*60*60)
	store := &fakeReportStore{
		getCashReportFunc: func(_ context.Context, startDate, endDate time.Time) (*repository.CashReport, error) {
			expectedStart := time.Date(2026, 5, 1, 0, 0, 0, 0, jakarta)
			expectedEnd := expectedStart
			if !startDate.Equal(expectedStart) {
				t.Fatalf("unexpected startDate: %v", startDate)
			}
			if !endDate.Equal(expectedEnd) {
				t.Fatalf("expected inclusive local endDate %v, got %v", expectedEnd, endDate)
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

func TestExportInventoryCSVUsesJakartaDateFilename(t *testing.T) {
	handler := NewReportHandler(&fakeReportStore{})
	app := fiber.New()
	app.Get("/reports/export/inventory", handler.ExportInventoryCSV)

	req := httptest.NewRequest(http.MethodGet, "/reports/export/inventory", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	expectedFilename := "attachment; filename=inventory_" + time.Now().In(reportBusinessLocation).Format("20060102") + ".csv"
	if got := resp.Header.Get("Content-Disposition"); got != expectedFilename {
		t.Fatalf("expected content disposition %q, got %q", expectedFilename, got)
	}
}

func TestExportComprehensiveReportIncludesGeneratedLine(t *testing.T) {
	handler := NewReportHandler(&fakeReportStore{
		getSalesRangeSummaryFunc: func(context.Context, time.Time, time.Time) (*repository.SalesRangeSummary, error) {
			return &repository.SalesRangeSummary{}, nil
		},
		getSalesRangeReportFunc: func(context.Context, time.Time, time.Time) ([]repository.DailySalesReport, error) {
			return []repository.DailySalesReport{}, nil
		},
		getTopSellersFunc: func(context.Context, time.Time, time.Time, int) ([]repository.TopSellerItem, error) {
			return []repository.TopSellerItem{}, nil
		},
		getEmployeeSalesReportFun: func(context.Context, time.Time, time.Time) ([]map[string]interface{}, error) {
			return []map[string]interface{}{}, nil
		},
		getCategorySalesReportFun: func(context.Context, time.Time, time.Time) ([]map[string]interface{}, error) {
			return []map[string]interface{}{}, nil
		},
	})
	app := fiber.New()
	app.Get("/reports/export/comprehensive", handler.ExportComprehensiveReportCSV)

	req := httptest.NewRequest(http.MethodGet, "/reports/export/comprehensive?start_date=2026-05-01&end_date=2026-05-01", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	reader := csv.NewReader(resp.Body)
	reader.Comma = ';'
	reader.FieldsPerRecord = -1
	rows, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("failed to read csv: %v", err)
	}
	if len(rows) < 2 || len(rows[1]) < 2 || rows[1][0] != "Generated:" {
		t.Fatalf("expected generated row, got %v", rows)
	}
	if strings.TrimSpace(rows[1][1]) == "" {
		t.Fatalf("expected generated timestamp, got %q", rows[1][1])
	}
}

func TestExportSalesCSVSanitizesTextCellsAndLogsAudit(t *testing.T) {
	auditRepo := &stubAuditRepository{}
	audit.Init(auditRepo)
	store := &fakeReportStore{
		getSalesForExportFunc: func(context.Context, time.Time, time.Time) ([]repository.SalesReportItem, error) {
			return []repository.SalesReportItem{{
				InvoiceNo:     "=INV-1",
				Date:          "2026-05-01",
				Time:          "09:00:00",
				EmployeeName:  "\t+Cashier\nOne",
				ItemCount:     1,
				Subtotal:      decimal.NewFromInt(1000),
				Tax:           decimal.Zero,
				Discount:      decimal.Zero,
				Total:         decimal.NewFromInt(1000),
				PaymentMethod: "@cash",
				Status:        "-completed",
			}}, nil
		},
		getSalesRangeSummaryFunc: func(context.Context, time.Time, time.Time) (*repository.SalesRangeSummary, error) {
			return &repository.SalesRangeSummary{
				TotalTransactions: 1,
				TotalItems:        1,
				TotalAmount:       decimal.NewFromInt(1000),
			}, nil
		},
	}
	handler := NewReportHandler(store)
	app := fiber.New()
	app.Get("/reports/export/sales", handler.ExportSalesCSV)

	req := httptest.NewRequest(http.MethodGet, "/reports/export/sales?start_date=2026-05-01&end_date=2026-05-01", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	if !strings.Contains(string(bodyBytes), "Invoice No;Date;Time;Employee") {
		t.Fatalf("expected semicolon-delimited CSV, got %q", string(bodyBytes))
	}

	reader := csv.NewReader(bytes.NewReader(bodyBytes))
	reader.Comma = ';'
	reader.FieldsPerRecord = -1
	rows, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("failed to read csv: %v", err)
	}
	itemRow := rows[len(rows)-1]
	if itemRow[0] != "'=INV-1" {
		t.Fatalf("expected sanitized invoice, got %q", itemRow[0])
	}
	if !strings.HasPrefix(itemRow[3], "'") || strings.ContainsAny(itemRow[3], "\r\n\t") {
		t.Fatalf("expected sanitized employee name, got %q", itemRow[3])
	}
	if itemRow[9] != "'@cash" || itemRow[10] != "'-completed" {
		t.Fatalf("expected sanitized payment/status, got %q and %q", itemRow[9], itemRow[10])
	}

	if auditRepo.last == nil {
		t.Fatal("expected export audit log")
	}
	if auditRepo.last.Action != models.AuditActionReportExport {
		t.Fatalf("expected action %q, got %q", models.AuditActionReportExport, auditRepo.last.Action)
	}
	if auditRepo.last.EntityType != models.AuditEntityReport || auditRepo.last.EntityID != "sales" {
		t.Fatalf("expected report sales audit entity, got %q %q", auditRepo.last.EntityType, auditRepo.last.EntityID)
	}
	if auditRepo.last.NewValues["export_type"] != "sales" {
		t.Fatalf("expected export_type sales, got %#v", auditRepo.last.NewValues["export_type"])
	}
	if _, ok := auditRepo.last.NewValues["csv"]; ok {
		t.Fatal("audit metadata must not include CSV contents")
	}
}

func TestTopSellersRejectsInvalidLimits(t *testing.T) {
	handler := NewReportHandler(&fakeReportStore{})
	app := fiber.New()
	app.Get("/reports/top-sellers", handler.GetTopSellers)
	app.Get("/reports/export/top-sellers", handler.ExportTopSellersCSV)

	tests := []string{
		"/reports/top-sellers?limit=abc",
		"/reports/top-sellers?limit=101",
		"/reports/export/top-sellers?limit=0",
		"/reports/export/top-sellers?limit=101",
	}

	for _, path := range tests {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("%s returned error: %v", path, err)
		}
		if resp.StatusCode != fiber.StatusBadRequest {
			t.Fatalf("%s expected 400, got %d", path, resp.StatusCode)
		}
	}
}

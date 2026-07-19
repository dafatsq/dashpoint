package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"dashpoint/backend/internal/models"
	"dashpoint/backend/internal/repository"
)

type fakeSaleStore struct {
	createFunc          func(context.Context, *repository.CreateSaleRequest) (*models.Sale, error)
	validateCartFunc    func(context.Context, *repository.ValidateSaleCartRequest) error
	getByIDFunc         func(context.Context, uuid.UUID) (*models.Sale, error)
	getByInvoiceNoFunc  func(context.Context, string) (*models.Sale, error)
	listFunc            func(context.Context, *repository.SaleFilter) ([]models.Sale, int, error)
	voidSaleFunc        func(context.Context, uuid.UUID, uuid.UUID, string) error
	getDailySummaryFunc func(context.Context, time.Time) (map[string]interface{}, error)
}

func (f *fakeSaleStore) Create(ctx context.Context, req *repository.CreateSaleRequest) (*models.Sale, error) {
	if f.createFunc != nil {
		return f.createFunc(ctx, req)
	}
	return nil, nil
}
func (f *fakeSaleStore) ValidateCart(ctx context.Context, req *repository.ValidateSaleCartRequest) error {
	if f.validateCartFunc != nil {
		return f.validateCartFunc(ctx, req)
	}
	return nil
}
func (f *fakeSaleStore) GetByID(ctx context.Context, id uuid.UUID) (*models.Sale, error) {
	if f.getByIDFunc != nil {
		return f.getByIDFunc(ctx, id)
	}
	return nil, nil
}
func (f *fakeSaleStore) GetByInvoiceNo(ctx context.Context, invoice string) (*models.Sale, error) {
	if f.getByInvoiceNoFunc != nil {
		return f.getByInvoiceNoFunc(ctx, invoice)
	}
	return nil, nil
}
func (f *fakeSaleStore) List(ctx context.Context, filter *repository.SaleFilter) ([]models.Sale, int, error) {
	if f.listFunc != nil {
		return f.listFunc(ctx, filter)
	}
	return nil, 0, nil
}
func (f *fakeSaleStore) VoidSale(ctx context.Context, saleID, userID uuid.UUID, reason string) error {
	if f.voidSaleFunc != nil {
		return f.voidSaleFunc(ctx, saleID, userID, reason)
	}
	return nil
}
func (f *fakeSaleStore) GetDailySummary(ctx context.Context, date time.Time) (map[string]interface{}, error) {
	if f.getDailySummaryFunc != nil {
		return f.getDailySummaryFunc(ctx, date)
	}
	return nil, nil
}

type fakeShiftLookup struct {
	shift *models.Shift
	err   error
}

func (f *fakeShiftLookup) GetCurrentOpenShift(context.Context) (*models.Shift, error) {
	return f.shift, f.err
}

func TestCreateSaleRejectsInvalidSaleDiscountValue(t *testing.T) {
	handler := NewSaleHandler(&fakeSaleStore{}, &fakeShiftLookup{})
	app := saleTestApp(handler, uuid.New(), "owner")
	body := `{
		"items":[{"product_id":"00000000-0000-0000-0000-000000000001","quantity":"1","unit_price":"10.00","discount_value":"0","discount_amount":"0"}],
		"payments":[{"payment_method":"cash","amount":"10.00"}],
		"discount_type":"fixed",
		"discount_value":"oops"
	}`

	resp := performJSONRequest(t, app, http.MethodPost, "/sales", body)
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestCreateSaleRejectsUnsupportedSaleDiscountType(t *testing.T) {
	handler := NewSaleHandler(&fakeSaleStore{}, &fakeShiftLookup{})
	app := saleTestApp(handler, uuid.New(), "owner")
	body := `{
		"items":[{"product_id":"00000000-0000-0000-0000-000000000001","quantity":"1","unit_price":"10.00","discount_value":"0","discount_amount":"0"}],
		"payments":[{"payment_method":"cash","amount":"10.00"}],
		"discount_type":"free",
		"discount_value":"10"
	}`

	resp := performJSONRequest(t, app, http.MethodPost, "/sales", body)
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestCreateSaleRejectsOutOfRangeSaleDiscountPercentage(t *testing.T) {
	handler := NewSaleHandler(&fakeSaleStore{}, &fakeShiftLookup{})
	app := saleTestApp(handler, uuid.New(), "owner")
	body := `{
		"items":[{"product_id":"00000000-0000-0000-0000-000000000001","quantity":"1","unit_price":"10.00","discount_value":"0","discount_amount":"0"}],
		"payments":[{"payment_method":"cash","amount":"10.00"}],
		"discount_type":"percentage",
		"discount_value":"101"
	}`

	resp := performJSONRequest(t, app, http.MethodPost, "/sales", body)
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestCreateSaleRejectsInvalidAmountTendered(t *testing.T) {
	handler := NewSaleHandler(&fakeSaleStore{}, &fakeShiftLookup{})
	app := saleTestApp(handler, uuid.New(), "owner")
	body := `{
		"items":[{"product_id":"00000000-0000-0000-0000-000000000001","quantity":"1","unit_price":"10.00","discount_value":"0","discount_amount":"0"}],
		"payments":[{"payment_method":"cash","amount":"10.00","amount_tendered":"oops"}]
	}`

	resp := performJSONRequest(t, app, http.MethodPost, "/sales", body)
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestCreateSaleRequiresOpenShift(t *testing.T) {
	handler := NewSaleHandler(&fakeSaleStore{}, &fakeShiftLookup{})
	app := saleTestApp(handler, uuid.New(), "owner")
	body := `{
		"items":[{"product_id":"00000000-0000-0000-0000-000000000001","quantity":"1","unit_price":"10.00","discount_value":"0","discount_amount":"0"}],
		"payments":[{"payment_method":"cash","amount":"10.00"}]
	}`

	resp := performJSONRequest(t, app, http.MethodPost, "/sales", body)
	if resp.StatusCode != fiber.StatusConflict {
		t.Fatalf("expected 409, got %d", resp.StatusCode)
	}
}

func TestCreateSaleRejectsMismatchedShiftID(t *testing.T) {
	currentShiftID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	staleShiftID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	createCalled := false
	handler := NewSaleHandler(&fakeSaleStore{
		createFunc: func(context.Context, *repository.CreateSaleRequest) (*models.Sale, error) {
			createCalled = true
			return nil, nil
		},
	}, &fakeShiftLookup{shift: &models.Shift{ID: currentShiftID}})
	app := saleTestApp(handler, uuid.New(), "owner")
	body := `{
		"shift_id":"` + staleShiftID.String() + `",
		"items":[{"product_id":"00000000-0000-0000-0000-000000000001","quantity":"1","unit_price":"10.00","discount_value":"0","discount_amount":"0"}],
		"payments":[{"payment_method":"cash","amount":"10.00"}]
	}`

	resp := performJSONRequest(t, app, http.MethodPost, "/sales", body)
	if resp.StatusCode != fiber.StatusConflict {
		t.Fatalf("expected 409, got %d", resp.StatusCode)
	}
	if createCalled {
		t.Fatalf("expected mismatched shift submit to be blocked")
	}
}

func TestValidateCartRejectsMismatchedShiftID(t *testing.T) {
	currentShiftID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	staleShiftID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	validateCalled := false
	handler := NewSaleHandler(&fakeSaleStore{
		validateCartFunc: func(context.Context, *repository.ValidateSaleCartRequest) error {
			validateCalled = true
			return nil
		},
	}, &fakeShiftLookup{shift: &models.Shift{ID: currentShiftID}})
	app := saleTestApp(handler, uuid.New(), "owner")
	body := `{
		"shift_id":"` + staleShiftID.String() + `",
		"items":[{"product_id":"00000000-0000-0000-0000-000000000001","quantity":"1","unit_price":"10.00","discount_value":"0","discount_amount":"0"}]
	}`

	resp := performJSONRequest(t, app, http.MethodPost, "/sales/validate", body)
	if resp.StatusCode != fiber.StatusConflict {
		t.Fatalf("expected 409, got %d", resp.StatusCode)
	}
	if validateCalled {
		t.Fatalf("expected mismatched shift validation to be blocked")
	}
}

func TestValidateCartCallsRepositoryWithoutPayments(t *testing.T) {
	shiftID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	var capturedItems []repository.CreateSaleItemRequest
	handler := NewSaleHandler(&fakeSaleStore{
		validateCartFunc: func(_ context.Context, req *repository.ValidateSaleCartRequest) error {
			capturedItems = req.Items
			return nil
		},
	}, &fakeShiftLookup{shift: &models.Shift{ID: shiftID}})
	app := saleTestApp(handler, uuid.New(), "owner")
	body := `{
		"shift_id":"` + shiftID.String() + `",
		"items":[{"product_id":"00000000-0000-0000-0000-000000000001","quantity":"1","unit_price":"10.00","discount_value":"0","discount_amount":"0"}]
	}`

	resp := performJSONRequest(t, app, http.MethodPost, "/sales/validate", body)
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if len(capturedItems) != 1 {
		t.Fatalf("expected one validated item, got %d", len(capturedItems))
	}
}

func TestListSalesRejectsInvalidShiftID(t *testing.T) {
	handler := NewSaleHandler(&fakeSaleStore{}, &fakeShiftLookup{})
	app := saleTestApp(handler, uuid.New(), "owner")

	req := httptest.NewRequest(http.MethodGet, "/sales?shift_id=bad-uuid", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestListSalesUsesJakartaExclusiveDateBounds(t *testing.T) {
	var got *repository.SaleFilter
	handler := NewSaleHandler(&fakeSaleStore{
		listFunc: func(_ context.Context, filter *repository.SaleFilter) ([]models.Sale, int, error) {
			got = filter
			return []models.Sale{}, 0, nil
		},
	}, &fakeShiftLookup{})
	app := saleTestApp(handler, uuid.New(), "owner")

	req := httptest.NewRequest(http.MethodGet, "/sales?from=2026-05-29&to=2026-05-29", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	jakarta := time.FixedZone("WIB", 7*60*60)
	expectedStart := time.Date(2026, 5, 29, 0, 0, 0, 0, jakarta)
	expectedEnd := expectedStart.Add(24 * time.Hour)
	if got == nil || got.StartDate == nil || !got.StartDate.Equal(expectedStart) {
		t.Fatalf("expected startDate %v, got %+v", expectedStart, got)
	}
	if got.EndDate == nil || !got.EndDate.Equal(expectedEnd) {
		t.Fatalf("expected exclusive endDate %v, got %v", expectedEnd, got.EndDate)
	}
}

func TestGetDailySummaryParsesDateInJakartaTimezone(t *testing.T) {
	var got time.Time
	handler := NewSaleHandler(&fakeSaleStore{
		getDailySummaryFunc: func(_ context.Context, date time.Time) (map[string]interface{}, error) {
			got = date
			return map[string]interface{}{}, nil
		},
	}, &fakeShiftLookup{})
	app := saleTestApp(handler, uuid.New(), "owner")
	app.Get("/sales/summary/daily", handler.GetDailySummary)

	req := httptest.NewRequest(http.MethodGet, "/sales/summary/daily?date=2026-05-29", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	jakarta := time.FixedZone("WIB", 7*60*60)
	expected := time.Date(2026, 5, 29, 0, 0, 0, 0, jakarta)
	if !got.Equal(expected) {
		t.Fatalf("expected %v, got %v", expected, got)
	}
}

func saleTestApp(handler *SaleHandler, userID uuid.UUID, roleName string) *fiber.App {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		c.Locals("role_name", roleName)
		return c.Next()
	})
	app.Post("/sales", handler.CreateSale)
	app.Post("/sales/validate", handler.ValidateCart)
	app.Get("/sales", handler.ListSales)
	return app
}

func performJSONRequest(t *testing.T, app *fiber.App, method, path, body string) *http.Response {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	return resp
}

func TestToSaleResponseIncludesPaymentsAndItems(t *testing.T) {
	handler := NewSaleHandler(&fakeSaleStore{}, &fakeShiftLookup{})
	productID := uuid.New()
	sale := &models.Sale{
		ID:             uuid.New(),
		InvoiceNo:      "INV-1",
		Subtotal:       decimal.RequireFromString("10"),
		TaxAmount:      decimal.Zero,
		DiscountAmount: decimal.Zero,
		TotalAmount:    decimal.RequireFromString("10"),
		ItemCount:      1,
		PaymentStatus:  models.PaymentStatusPaid,
		AmountPaid:     decimal.RequireFromString("10"),
		ChangeAmount:   decimal.Zero,
		EmployeeID:     uuid.New(),
		Status:         models.SaleStatusCompleted,
		CreatedAt:      time.Unix(100, 0),
		UpdatedAt:      time.Unix(100, 0),
		Items: []models.SaleItem{{
			ID:           uuid.New(),
			ProductID:    productID,
			ProductName:  "Item",
			Quantity:     decimal.RequireFromString("1"),
			UnitPrice:    decimal.RequireFromString("10"),
			DiscountType: nil,
			TaxRate:      decimal.Zero,
			TaxAmount:    decimal.Zero,
			Subtotal:     decimal.RequireFromString("10"),
			Total:        decimal.RequireFromString("10"),
		}},
		Payments: []models.Payment{{
			ID:            uuid.New(),
			PaymentMethod: models.PaymentMethodCash,
			Amount:        decimal.RequireFromString("10"),
			Status:        models.PaymentRecordCompleted,
		}},
	}

	payload := handler.toSaleResponse(sale)
	if _, ok := payload["items"]; !ok {
		t.Fatal("expected items in response")
	}
	if _, ok := payload["payments"]; !ok {
		t.Fatal("expected payments in response")
	}
}

func TestListSalesResponseSerializesPayments(t *testing.T) {
	handler := NewSaleHandler(&fakeSaleStore{
		listFunc: func(context.Context, *repository.SaleFilter) ([]models.Sale, int, error) {
			return []models.Sale{{
				ID:            uuid.New(),
				InvoiceNo:     "INV-1",
				TotalAmount:   decimal.RequireFromString("25"),
				ItemCount:     1,
				PaymentStatus: models.PaymentStatusPaid,
				Status:        models.SaleStatusCompleted,
				CreatedAt:     time.Unix(100, 0),
				Payments:      []models.Payment{{PaymentMethod: models.PaymentMethodCash}},
			}}, 1, nil
		},
	}, &fakeShiftLookup{})
	app := saleTestApp(handler, uuid.New(), "owner")

	resp := performJSONRequest(t, app, http.MethodGet, "/sales", "")
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var payload struct {
		Sales []map[string]any `json:"sales"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode payload: %v", err)
	}
	if len(payload.Sales) != 1 {
		t.Fatalf("expected one sale, got %d", len(payload.Sales))
	}
}

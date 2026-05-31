package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	managerPasswordHash                = "$2a$12$Lk7MKaGpZkn/ZM2.DGwele0oBQSFI7znHnkyvP0KG83c2UMMg9cLa"
	cashierPasswordHash                = "$2a$12$.HFs8tIjNMFIKE6My.5DIe6FHrkO5O3Qdj30PElSwWCBLCTn.fzEu"
	managerPinHash                     = "$2a$12$GHnsR25t75pwOjuwus7cKOC3c8F.KBzBSNOso29RXYctuQm/glSeS"
	cashierPinHash                     = "$2a$12$AXQP.ErpkUyp9k/U1lOBYOW8A8HqXl5jxl5PTWlJ9aFKAqp7QSYh."
	seedMarker                         = "[seed]"
	rupiahScale                        = 1000
	inventoryPurchaseCategorySystemKey = "inventory_purchase"
)

type role struct {
	ID   uuid.UUID
	Name string
}

type category struct {
	ID        uuid.UUID
	Name      string
	SystemKey *string
}

type userSeed struct {
	ID           uuid.UUID
	Email        string
	Name         string
	RoleName     string
	PasswordHash string
	PinHash      string
}

type productSeed struct {
	ID          uuid.UUID
	SKU         string
	Barcode     string
	Name        string
	Description string
	CategoryID  uuid.UUID
	Price       float64
	Cost        float64
	TaxRate     float64
	IsActive    bool
	ImageURL    string
	Quantity    float64
	Threshold   float64
	CreatedAt   time.Time
}

type shiftSeed struct {
	ID               uuid.UUID
	EmployeeID       uuid.UUID
	StartedAt        time.Time
	EndedAt          *time.Time
	Status           string
	OpeningCash      float64
	ClosingCash      *float64
	ExpectedCash     *float64
	CashDifference   *float64
	TotalSales       float64
	TotalVoided      float64
	TransactionCount int
	VoidCount        int
	Notes            string
	ClosedBy         *uuid.UUID
}

type saleSeed struct {
	ID             uuid.UUID
	InvoiceNo      string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Subtotal       float64
	TaxAmount      float64
	DiscountAmount float64
	TotalAmount    float64
	ItemCount      int
	PaymentStatus  string
	AmountPaid     float64
	ChangeAmount   float64
	DiscountType   *string
	DiscountValue  *float64
	DiscountReason *string
	EmployeeID     uuid.UUID
	ShiftID        uuid.UUID
	CustomerName   string
	CustomerPhone  string
	Status         string
	VoidedAt       *time.Time
	VoidedBy       *uuid.UUID
	VoidReason     *string
	Notes          string
	Items          []saleItemSeed
	Payment        paymentSeed
}

type saleItemSeed struct {
	ID             uuid.UUID
	ProductID      uuid.UUID
	ProductName    string
	ProductSKU     string
	ProductBarcode string
	Quantity       float64
	UnitPrice      float64
	CostPrice      float64
	DiscountType   *string
	DiscountValue  *float64
	DiscountAmount float64
	TaxRate        float64
	TaxAmount      float64
	Subtotal       float64
	Total          float64
}

type paymentSeed struct {
	ID             uuid.UUID
	Method         string
	Amount         float64
	AmountTendered *float64
	ChangeGiven    *float64
	ReferenceNo    *string
	Status         string
	CreatedAt      time.Time
}

type expenseSeed struct {
	ID               uuid.UUID
	CategoryID       uuid.UUID
	Amount           float64
	Description      string
	ExpenseDate      time.Time
	Vendor           string
	ReferenceNumber  string
	Notes            string
	CreatedBy        uuid.UUID
	ProductID        *uuid.UUID
	Quantity         *float64
	AppliesInventory bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type auditSeed struct {
	ID          uuid.UUID
	CreatedAt   time.Time
	UserID      uuid.UUID
	UserName    string
	UserRole    string
	Action      string
	EntityType  string
	EntityID    string
	Description string
	OldValues   string
	NewValues   string
	Metadata    string
	Status      string
}

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://dashpoint:dashpoint_dev@localhost:5432/dashpoint_dev?sslmode=disable"
	}
	seedMode := os.Getenv("SEED_MODE")
	if seedMode == "" {
		seedMode = "full"
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer pool.Close()

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		log.Fatalf("begin tx: %v", err)
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	roles, err := loadRoles(ctx, tx)
	if err != nil {
		log.Fatalf("load roles: %v", err)
	}
	categories, err := loadCategories(ctx, tx)
	if err != nil {
		log.Fatalf("load categories: %v", err)
	}
	owner, err := loadOwner(ctx, tx)
	if err != nil {
		log.Fatalf("load owner: %v", err)
	}
	var expenseCategories []category
	if seedMode != "core" {
		expenseCategories, err = loadExpenseCategories(ctx, tx)
		if err != nil {
			log.Fatalf("load expense categories: %v", err)
		}
	}

	if err := cleanupSeedData(ctx, tx, seedMode); err != nil {
		log.Fatalf("cleanup: %v", err)
	}

	users := buildSeedUsers()
	if err := insertUsers(ctx, tx, roles, users); err != nil {
		log.Fatalf("insert users: %v", err)
	}
	if seedMode != "core" {
		if err := insertRefreshTokens(ctx, tx, users); err != nil {
			log.Fatalf("insert refresh tokens: %v", err)
		}
	}

	products := buildProducts(categories)
	if err := insertProducts(ctx, tx, products); err != nil {
		log.Fatalf("insert products: %v", err)
	}
	if err := insertInventory(ctx, tx, products); err != nil {
		log.Fatalf("insert inventory: %v", err)
	}
	if err := insertStockAdjustments(ctx, tx, owner.ID, products); err != nil {
		log.Fatalf("insert stock adjustments: %v", err)
	}

	shifts := buildShifts(users, owner.ID)
	var sales []saleSeed
	if seedMode == "full" {
		sales = buildSales(users, products, shifts)
		applyShiftSummaries(shifts, sales)
	}

	if err := insertShifts(ctx, tx, shifts); err != nil {
		log.Fatalf("insert shifts: %v", err)
	}
	if seedMode == "full" {
		if err := insertSales(ctx, tx, sales); err != nil {
			log.Fatalf("insert sales: %v", err)
		}
		if err := insertCashDrawerOps(ctx, tx, shifts, users); err != nil {
			log.Fatalf("insert cash drawer ops: %v", err)
		}

		expenses := buildExpenses(users, expenseCategories, products)
		if err := insertExpenses(ctx, tx, expenses); err != nil {
			log.Fatalf("insert expenses: %v", err)
		}

		auditLogs := buildAuditLogs(owner, users, products, shifts, sales, expenses)
		if err := insertAuditLogs(ctx, tx, auditLogs); err != nil {
			log.Fatalf("insert audit logs: %v", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		log.Fatalf("commit seed: %v", err)
	}
	tx = nil

	fmt.Println("Seed complete")
	fmt.Printf("Mode: %s\n", seedMode)
	fmt.Printf("Users: %d\n", len(users)+1)
	fmt.Printf("Products: %d\n", len(products))
	fmt.Printf("Inventory items: %d\n", len(products))
	fmt.Printf("Shifts: %d\n", len(shifts))
	if seedMode == "full" {
		fmt.Printf("Sales: %d\n", len(sales))
		fmt.Printf("Expenses: %d\n", 22)
		fmt.Printf("Audit logs: %d\n", 45)
	}
}

func loadRoles(ctx context.Context, tx pgx.Tx) (map[string]role, error) {
	rows, err := tx.Query(ctx, `SELECT id, name FROM roles`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[string]role{}
	for rows.Next() {
		var r role
		if err := rows.Scan(&r.ID, &r.Name); err != nil {
			return nil, err
		}
		out[r.Name] = r
	}
	return out, rows.Err()
}

func loadCategories(ctx context.Context, tx pgx.Tx) ([]category, error) {
	rows, err := tx.Query(ctx, `SELECT id, name FROM categories WHERE is_active = true ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []category
	for rows.Next() {
		var c category
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func loadExpenseCategories(ctx context.Context, tx pgx.Tx) ([]category, error) {
	rows, err := tx.Query(ctx, `SELECT id, name, system_key FROM expense_categories WHERE is_active = true ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []category
	for rows.Next() {
		var c category
		if err := rows.Scan(&c.ID, &c.Name, &c.SystemKey); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func loadOwner(ctx context.Context, tx pgx.Tx) (userSeed, error) {
	row := tx.QueryRow(ctx, `
		SELECT u.id, u.email, u.name, r.name
		FROM users u
		JOIN roles r ON r.id = u.role_id
		WHERE r.name = 'owner'
		ORDER BY u.created_at
		LIMIT 1
	`)
	var user userSeed
	if err := row.Scan(&user.ID, &user.Email, &user.Name, &user.RoleName); err != nil {
		return userSeed{}, err
	}
	return user, nil
}

func cleanupSeedData(ctx context.Context, tx pgx.Tx, seedMode string) error {
	queries := []string{
		`DELETE FROM inventory_items WHERE product_id IN (SELECT id FROM products WHERE sku LIKE 'SEED-SKU-%')`,
		`DELETE FROM products WHERE sku LIKE 'SEED-SKU-%'`,
		`DELETE FROM shifts WHERE notes LIKE '[seed]%'`,
		`DELETE FROM users WHERE email IN ('manager@dashpoint.local','manager2@dashpoint.local','cashier@dashpoint.local','cashier2@dashpoint.local','cashier3@dashpoint.local','cashier4@dashpoint.local')`,
	}
	if seedMode == "full" {
		queries = append([]string{
			`DELETE FROM refresh_tokens WHERE token_hash LIKE 'seed:%'`,
			`DELETE FROM audit_logs WHERE description LIKE '[seed]%'`,
			`DELETE FROM payments WHERE sale_id IN (SELECT id FROM sales WHERE invoice_no LIKE 'SEED-INV-%')`,
			`DELETE FROM sale_items WHERE sale_id IN (SELECT id FROM sales WHERE invoice_no LIKE 'SEED-INV-%')`,
			`DELETE FROM sales WHERE invoice_no LIKE 'SEED-INV-%'`,
			`DELETE FROM cash_drawer_operations WHERE reason LIKE '[seed]%'`,
			`DELETE FROM stock_adjustments WHERE reason LIKE '[seed]%'`,
			`DELETE FROM expenses WHERE notes LIKE '[seed]%'`,
		}, queries...)
	}
	for _, q := range queries {
		if _, err := tx.Exec(ctx, q); err != nil {
			return err
		}
	}
	return nil
}

func buildSeedUsers() []userSeed {
	return []userSeed{
		{
			ID:           uuid.MustParse("00000000-0000-0000-0000-000000000002"),
			Email:        "manager@dashpoint.local",
			Name:         "Manager One",
			RoleName:     "manager",
			PasswordHash: managerPasswordHash,
			PinHash:      managerPinHash,
		},
		{
			ID:           uuid.MustParse("00000000-0000-0000-0000-000000000003"),
			Email:        "cashier@dashpoint.local",
			Name:         "Cashier One",
			RoleName:     "cashier",
			PasswordHash: cashierPasswordHash,
			PinHash:      cashierPinHash,
		},
		{
			ID:           uuid.MustParse("00000000-0000-0000-0000-000000000004"),
			Email:        "manager2@dashpoint.local",
			Name:         "Manager Two",
			RoleName:     "manager",
			PasswordHash: managerPasswordHash,
			PinHash:      managerPinHash,
		},
		{
			ID:           uuid.MustParse("00000000-0000-0000-0000-000000000005"),
			Email:        "cashier2@dashpoint.local",
			Name:         "Cashier Two",
			RoleName:     "cashier",
			PasswordHash: cashierPasswordHash,
			PinHash:      cashierPinHash,
		},
		{
			ID:           uuid.MustParse("00000000-0000-0000-0000-000000000006"),
			Email:        "cashier3@dashpoint.local",
			Name:         "Cashier Three",
			RoleName:     "cashier",
			PasswordHash: cashierPasswordHash,
			PinHash:      cashierPinHash,
		},
		{
			ID:           uuid.MustParse("00000000-0000-0000-0000-000000000007"),
			Email:        "cashier4@dashpoint.local",
			Name:         "Cashier Four",
			RoleName:     "cashier",
			PasswordHash: cashierPasswordHash,
			PinHash:      cashierPinHash,
		},
	}
}

func insertUsers(ctx context.Context, tx pgx.Tx, roles map[string]role, users []userSeed) error {
	now := time.Now().UTC()
	for _, user := range users {
		roleID := roles[user.RoleName].ID
		if _, err := tx.Exec(ctx, `
			INSERT INTO users (
				id, email, name, password_hash, pin_hash, role_id, is_active, last_login_at, created_at, updated_at
			) VALUES ($1,$2,$3,$4,$5,$6,true,$7,$8,$9)
		`, user.ID, user.Email, user.Name, user.PasswordHash, user.PinHash, roleID, now.Add(-2*time.Hour), now, now); err != nil {
			return err
		}
	}
	return nil
}

func insertRefreshTokens(ctx context.Context, tx pgx.Tx, users []userSeed) error {
	now := time.Now().UTC()
	for idx, user := range users[:4] {
		tokenID := seedUUID("refresh-token", idx)
		expiresAt := now.Add(7 * 24 * time.Hour)
		var revokedAt any
		var revokedReason any
		if idx == 3 {
			revokedAt = now.Add(-6 * time.Hour)
			revokedReason = "seed_logout"
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO refresh_tokens (
				id, user_id, token_hash, expires_at, created_at, revoked_at, revoked_reason
			) VALUES ($1,$2,$3,$4,$5,$6,$7)
		`, tokenID, user.ID, fmt.Sprintf("seed:%s", user.Email), expiresAt, now.Add(time.Duration(-idx)*time.Hour), revokedAt, revokedReason); err != nil {
			return err
		}
	}
	return nil
}

func buildProducts(categories []category) []productSeed {
	var products []productSeed
	now := time.Now().UTC()
	productCount := 42
	for i := 0; i < productCount; i++ {
		category := categories[i%len(categories)]
		price := rupiah(1.25 + float64((i%9)+1)*0.85 + float64(i)/15)
		cost := round2(price * 0.58)
		qty := round3(18 + float64((i*7)%65))
		if i%6 == 0 {
			qty += 12
		}
		threshold := float64(4 + (i % 5))
		product := productSeed{
			ID:          seedUUID("product", i+1),
			SKU:         fmt.Sprintf("SEED-SKU-%03d", i+1),
			Barcode:     fmt.Sprintf("880000000%03d", i+1),
			Name:        fmt.Sprintf("%s Demo Item %02d", category.Name, (i%6)+1),
			Description: fmt.Sprintf("%s demo catalog item %02d %s", seedMarker, i+1, stringsForCategory(category.Name, i)),
			CategoryID:  category.ID,
			Price:       price,
			Cost:        cost,
			TaxRate:     []float64{0, 5, 10}[i%3],
			IsActive:    i%13 != 0,
			ImageURL:    "",
			Quantity:    qty,
			Threshold:   threshold,
			CreatedAt:   now.Add(-time.Duration((i+1)*6) * time.Hour),
		}
		products = append(products, product)
	}
	return products
}

func stringsForCategory(category string, idx int) string {
	suffixes := map[string][]string{
		"Food":      {"snack", "biscuit", "noodle", "cracker", "candy", "bread"},
		"Beverages": {"juice", "soda", "tea", "coffee", "water", "milk"},
		"Tobacco":   {"pack", "menthol", "filter", "slim", "mild", "classic"},
		"Groceries": {"rice", "oil", "flour", "sugar", "salt", "sauce"},
		"Hygiene":   {"soap", "shampoo", "detergent", "toothpaste", "tissue", "cleaner"},
		"Medicine":  {"vitamin", "tablet", "balm", "syrup", "capsule", "ointment"},
		"Others":    {"battery", "lighter", "stationery", "bag", "cup", "container"},
	}
	values := suffixes[category]
	if len(values) == 0 {
		return "item"
	}
	return values[idx%len(values)]
}

func insertProducts(ctx context.Context, tx pgx.Tx, products []productSeed) error {
	for _, product := range products {
		if _, err := tx.Exec(ctx, `
			INSERT INTO products (
				id, sku, barcode, name, description, category_id, price, cost, tax_rate,
				is_active, image_url, created_at, updated_at
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,NULLIF($11,''),$12,$13)
		`, product.ID, product.SKU, product.Barcode, product.Name, product.Description, product.CategoryID, money(product.Price), money(product.Cost), money(product.TaxRate), product.IsActive, product.ImageURL, product.CreatedAt, product.CreatedAt); err != nil {
			return err
		}
	}
	return nil
}

func insertInventory(ctx context.Context, tx pgx.Tx, products []productSeed) error {
	now := time.Now().UTC()
	for _, product := range products {
		if _, err := tx.Exec(ctx, `
			INSERT INTO inventory_items (
				product_id, quantity, low_stock_threshold, updated_at
			) VALUES ($1,$2,$3,$4)
		`, product.ID, qty(product.Quantity), qty(product.Threshold), now); err != nil {
			return err
		}
	}
	return nil
}

func insertStockAdjustments(ctx context.Context, tx pgx.Tx, adjustedBy uuid.UUID, products []productSeed) error {
	now := time.Now().UTC()
	for i, product := range products {
		if _, err := tx.Exec(ctx, `
			INSERT INTO stock_adjustments (
				id, product_id, adjustment_type, quantity_before, quantity_change, quantity_after,
				reason, reference_type, reference_id, adjusted_by, created_at
			) VALUES ($1,$2,'initial',$3,$4,$5,$6,$7,$8,$9,$10)
		`, seedUUID("stock-initial", i+1), product.ID, qty(0), qty(product.Quantity), qty(product.Quantity), seedMarker+" initial stock load", "seed", nil, adjustedBy, now.Add(-time.Duration(240-i)*time.Minute)); err != nil {
			return err
		}
		if i%4 == 0 {
			before := product.Quantity
			change := 5.0 + float64(i%3)
			if _, err := tx.Exec(ctx, `
				INSERT INTO stock_adjustments (
					id, product_id, adjustment_type, quantity_before, quantity_change, quantity_after,
					reason, reference_type, reference_id, adjusted_by, created_at
				) VALUES ($1,$2,'purchase',$3,$4,$5,$6,$7,$8,$9,$10)
			`, seedUUID("stock-purchase", i+1), product.ID, qty(before-change), qty(change), qty(before), seedMarker+" replenishment batch", "seed", nil, adjustedBy, now.Add(-time.Duration(180-i)*time.Minute)); err != nil {
				return err
			}
		}
	}
	return nil
}

func buildShifts(users []userSeed, ownerID uuid.UUID) []shiftSeed {
	var shifts []shiftSeed
	base := time.Now().UTC().AddDate(0, 0, -9)
	cashierIDs := []uuid.UUID{users[1].ID, users[3].ID, users[4].ID, users[5].ID}
	for i := 0; i < 12; i++ {
		started := base.Add(time.Duration(i*14) * time.Hour)
		ended := started.Add(9*time.Hour + time.Duration((i%3))*20*time.Minute)
		opening := rupiah(150 + float64((i%4)*25))
		shifts = append(shifts, shiftSeed{
			ID:          seedUUID("shift-closed", i+1),
			EmployeeID:  cashierIDs[i%len(cashierIDs)],
			StartedAt:   started,
			EndedAt:     &ended,
			Status:      "closed",
			OpeningCash: opening,
			Notes:       seedMarker + fmt.Sprintf(" closed shift %02d", i+1),
			ClosedBy:    &ownerID,
		})
	}
	openStarted := time.Now().UTC().Add(-4 * time.Hour)
	shifts = append(shifts, shiftSeed{
		ID:          seedUUID("shift-open", 1),
		EmployeeID:  users[1].ID,
		StartedAt:   openStarted,
		Status:      "open",
		OpeningCash: rupiah(200),
		Notes:       seedMarker + " active shift for POS testing",
	})
	return shifts
}

func buildSales(users []userSeed, products []productSeed, shifts []shiftSeed) []saleSeed {
	var sales []saleSeed
	methods := []string{"cash", "card", "qris", "transfer", "voucher"}
	closedShifts := make([]shiftSeed, 0, len(shifts))
	for _, shift := range shifts {
		if shift.Status == "closed" {
			closedShifts = append(closedShifts, shift)
		}
	}

	for i := 0; i < 36; i++ {
		shift := closedShifts[i%len(closedShifts)]
		productA := products[i%len(products)]
		productB := products[(i*3+7)%len(products)]
		qtyA := float64((i % 3) + 1)
		qtyB := float64(((i + 1) % 2) + 1)
		subA := round2(productA.Price * qtyA)
		subB := round2(productB.Price * qtyB)
		taxA := round2(subA * productA.TaxRate / 100)
		taxB := round2(subB * productB.TaxRate / 100)
		subtotal := round2(subA + subB)
		taxAmount := round2(taxA + taxB)
		var discountType *string
		var discountValue *float64
		var discountReason *string
		discountAmount := 0.0
		if i%5 == 0 {
			kind := "fixed"
			value := rupiah(0.75 + float64(i%4))
			reason := seedMarker + " promo"
			discountType = &kind
			discountValue = &value
			discountReason = &reason
			discountAmount = value
		}
		total := round2(subtotal + taxAmount - discountAmount)
		method := methods[i%len(methods)]
		createdAt := shift.StartedAt.Add(time.Duration(30+i*11) * time.Minute)
		status := "completed"
		paymentStatus := "paid"
		var voidedAt *time.Time
		var voidedBy *uuid.UUID
		var voidReason *string
		if i > 0 && i%11 == 0 {
			status = "voided"
			paymentStatus = "voided"
			t := createdAt.Add(25 * time.Minute)
			voidedAt = &t
			voidedBy = &users[0].ID
			reason := seedMarker + " voided training transaction"
			voidReason = &reason
		}

		amountTendered := total
		change := 0.0
		if method == "cash" {
			amountTendered = round2(math.Ceil(total/5000) * 5000)
			change = round2(amountTendered - total)
		}

		itemOne := saleItemSeed{
			ID:             seedUUID("sale-item-a", i+1),
			ProductID:      productA.ID,
			ProductName:    productA.Name,
			ProductSKU:     productA.SKU,
			ProductBarcode: productA.Barcode,
			Quantity:       qtyA,
			UnitPrice:      productA.Price,
			CostPrice:      productA.Cost,
			DiscountType:   nil,
			DiscountValue:  nil,
			DiscountAmount: 0,
			TaxRate:        productA.TaxRate,
			TaxAmount:      taxA,
			Subtotal:       subA,
			Total:          round2(subA + taxA),
		}
		itemTwoDiscount := discountAmount
		itemTwo := saleItemSeed{
			ID:             seedUUID("sale-item-b", i+1),
			ProductID:      productB.ID,
			ProductName:    productB.Name,
			ProductSKU:     productB.SKU,
			ProductBarcode: productB.Barcode,
			Quantity:       qtyB,
			UnitPrice:      productB.Price,
			CostPrice:      productB.Cost,
			DiscountType:   discountType,
			DiscountValue:  discountValue,
			DiscountAmount: itemTwoDiscount,
			TaxRate:        productB.TaxRate,
			TaxAmount:      taxB,
			Subtotal:       subB,
			Total:          round2(subB + taxB - itemTwoDiscount),
		}

		var referenceNo *string
		switch method {
		case "transfer":
			v := fmt.Sprintf("SEED-TF-%03d", i+1)
			referenceNo = &v
		case "qris":
			v := fmt.Sprintf("SEED-QR-%03d", i+1)
			referenceNo = &v
		}
		paymentStatusRow := "completed"
		if status == "voided" {
			paymentStatusRow = "refunded"
		}

		sales = append(sales, saleSeed{
			ID:             seedUUID("sale", i+1),
			InvoiceNo:      fmt.Sprintf("SEED-INV-%03d", i+1),
			CreatedAt:      createdAt,
			UpdatedAt:      createdAt,
			Subtotal:       subtotal,
			TaxAmount:      taxAmount,
			DiscountAmount: discountAmount,
			TotalAmount:    total,
			ItemCount:      int(qtyA + qtyB),
			PaymentStatus:  paymentStatus,
			AmountPaid:     total,
			ChangeAmount:   change,
			DiscountType:   discountType,
			DiscountValue:  discountValue,
			DiscountReason: discountReason,
			EmployeeID:     shift.EmployeeID,
			ShiftID:        shift.ID,
			CustomerName:   fmt.Sprintf("Customer %02d", i+1),
			CustomerPhone:  fmt.Sprintf("08123%06d", 100000+i),
			Status:         status,
			VoidedAt:       voidedAt,
			VoidedBy:       voidedBy,
			VoidReason:     voidReason,
			Notes:          seedMarker + " demo sale",
			Items:          []saleItemSeed{itemOne, itemTwo},
			Payment: paymentSeed{
				ID:             seedUUID("payment", i+1),
				Method:         method,
				Amount:         total,
				AmountTendered: nullableFloat(amountTendered, method == "cash"),
				ChangeGiven:    nullableFloat(change, method == "cash" && change > 0),
				ReferenceNo:    referenceNo,
				Status:         paymentStatusRow,
				CreatedAt:      createdAt,
			},
		})
	}

	return sales
}

func applyShiftSummaries(shifts []shiftSeed, sales []saleSeed) {
	shiftIndex := map[uuid.UUID]int{}
	for i := range shifts {
		shiftIndex[shifts[i].ID] = i
	}

	type cashOps struct {
		payIn  float64
		payOut float64
	}
	opTotals := map[uuid.UUID]cashOps{}
	for i, shift := range shifts {
		if shift.Status != "closed" {
			continue
		}
		if i%3 == 0 {
			opTotals[shift.ID] = cashOps{payIn: 20 + float64(i%4)*5, payOut: 10 + float64(i%3)*4}
		}
	}

	for _, sale := range sales {
		idx, ok := shiftIndex[sale.ShiftID]
		if !ok {
			continue
		}
		shift := &shifts[idx]
		if sale.Status == "voided" {
			shift.TotalVoided = round2(shift.TotalVoided + sale.TotalAmount)
			shift.VoidCount++
			continue
		}
		shift.TotalSales = round2(shift.TotalSales + sale.TotalAmount)
		shift.TransactionCount++
	}

	for i := range shifts {
		if shifts[i].Status != "closed" {
			continue
		}
		cashSales := 0.0
		cashRefunds := 0.0
		for _, sale := range sales {
			if sale.ShiftID != shifts[i].ID || sale.Payment.Method != "cash" {
				continue
			}
			if sale.Status == "voided" {
				cashRefunds += sale.TotalAmount
			} else {
				cashSales += sale.TotalAmount
			}
		}
		expected := round2(shifts[i].OpeningCash + cashSales - cashRefunds + opTotals[shifts[i].ID].payIn - opTotals[shifts[i].ID].payOut)
		diff := round2([]float64{-1.25, 0, 2.10}[i%3])
		closing := round2(expected + diff)
		shifts[i].ExpectedCash = &expected
		shifts[i].ClosingCash = &closing
		shifts[i].CashDifference = &diff
	}
}

func insertShifts(ctx context.Context, tx pgx.Tx, shifts []shiftSeed) error {
	for _, shift := range shifts {
		var updatedAt time.Time
		if shift.EndedAt != nil {
			updatedAt = *shift.EndedAt
		} else {
			updatedAt = shift.StartedAt
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO shifts (
				id, opened_by, started_at, ended_at, opening_cash, closing_cash, expected_cash, cash_difference,
				total_sales, total_voided, transaction_count, void_count, status, notes, created_at, updated_at, closed_by
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)
		`, shift.ID, shift.EmployeeID, shift.StartedAt, shift.EndedAt, money(shift.OpeningCash), nullableMoney(shift.ClosingCash), nullableMoney(shift.ExpectedCash), nullableMoney(shift.CashDifference), money(shift.TotalSales), money(shift.TotalVoided), shift.TransactionCount, shift.VoidCount, shift.Status, shift.Notes, shift.StartedAt, updatedAt, shift.ClosedBy); err != nil {
			return err
		}
	}
	return nil
}

func insertSales(ctx context.Context, tx pgx.Tx, sales []saleSeed) error {
	for _, sale := range sales {
		if _, err := tx.Exec(ctx, `
			INSERT INTO sales (
				id, invoice_no, created_at, updated_at, subtotal, tax_amount, discount_amount, total_amount, item_count,
				payment_status, amount_paid, change_amount, discount_type, discount_value, discount_reason, employee_id,
				shift_id, status, voided_at, voided_by, void_reason, notes
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22)
		`, sale.ID, sale.InvoiceNo, sale.CreatedAt, sale.UpdatedAt, money(sale.Subtotal), money(sale.TaxAmount), money(sale.DiscountAmount), money(sale.TotalAmount), sale.ItemCount, sale.PaymentStatus, money(sale.AmountPaid), money(sale.ChangeAmount), sale.DiscountType, nullableMoney(sale.DiscountValue), sale.DiscountReason, sale.EmployeeID, sale.ShiftID, sale.Status, sale.VoidedAt, sale.VoidedBy, sale.VoidReason, sale.Notes); err != nil {
			return err
		}
		for _, item := range sale.Items {
			if _, err := tx.Exec(ctx, `
				INSERT INTO sale_items (
					id, sale_id, product_id, product_name, product_sku, product_barcode, quantity, unit_price, cost_price,
					discount_type, discount_value, discount_amount, tax_rate, tax_amount, subtotal, total, created_at
				) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)
			`, item.ID, sale.ID, item.ProductID, item.ProductName, item.ProductSKU, item.ProductBarcode, qty(item.Quantity), money(item.UnitPrice), money(item.CostPrice), item.DiscountType, nullableMoney(item.DiscountValue), money(item.DiscountAmount), money(item.TaxRate), money(item.TaxAmount), money(item.Subtotal), money(item.Total), sale.CreatedAt); err != nil {
				return err
			}
		}
		payment := sale.Payment
		if _, err := tx.Exec(ctx, `
			INSERT INTO payments (
				id, sale_id, payment_method, amount, amount_tendered, change_given, reference_no, status, created_at
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		`, payment.ID, sale.ID, payment.Method, money(payment.Amount), nullableMoney(payment.AmountTendered), nullableMoney(payment.ChangeGiven), payment.ReferenceNo, payment.Status, payment.CreatedAt); err != nil {
			return err
		}
	}
	return nil
}

func insertCashDrawerOps(ctx context.Context, tx pgx.Tx, shifts []shiftSeed, users []userSeed) error {
	now := time.Now().UTC()
	actor := users[0].ID
	for i, shift := range shifts {
		if shift.Status != "closed" && shift.Status != "open" {
			continue
		}
		if shift.Status == "closed" && i%3 != 0 {
			continue
		}
		payIn := rupiah(20.0 + float64(i%4)*5)
		payOut := rupiah(10.0 + float64(i%3)*4)
		if _, err := tx.Exec(ctx, `
			INSERT INTO cash_drawer_operations (id, shift_id, type, amount, reason, performed_by, created_at)
			VALUES ($1,$2,'pay_in',$3,$4,$5,$6)
		`, seedUUID("cash-op-in", i+1), shift.ID, money(payIn), seedMarker+" pay in float top-up", actor, now.Add(-time.Duration(90-i)*time.Minute)); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO cash_drawer_operations (id, shift_id, type, amount, reason, performed_by, created_at)
			VALUES ($1,$2,'pay_out',$3,$4,$5,$6)
		`, seedUUID("cash-op-out", i+1), shift.ID, money(payOut), seedMarker+" petty cash run", actor, now.Add(-time.Duration(70-i)*time.Minute)); err != nil {
			return err
		}
	}
	return nil
}

func buildExpenses(users []userSeed, categories []category, products []productSeed) []expenseSeed {
	var expenses []expenseSeed
	now := time.Now().UTC()
	var inventoryPurchase category
	for _, category := range categories {
		if category.SystemKey != nil && *category.SystemKey == inventoryPurchaseCategorySystemKey {
			inventoryPurchase = category
			break
		}
	}
	otherCategories := make([]category, 0, len(categories))
	for _, category := range categories {
		if category.ID != inventoryPurchase.ID {
			otherCategories = append(otherCategories, category)
		}
	}

	for i := 0; i < 22; i++ {
		createdBy := users[i%len(users)].ID
		expenseDate := now.AddDate(0, 0, -(i % 18))
		expense := expenseSeed{
			ID:              seedUUID("expense", i+1),
			Amount:          rupiah(12 + float64(i%7)*8.5),
			Description:     fmt.Sprintf("%s operating expense %02d", seedMarker, i+1),
			ExpenseDate:     expenseDate,
			Vendor:          fmt.Sprintf("Vendor %02d", (i%8)+1),
			ReferenceNumber: fmt.Sprintf("SEED-EXP-%03d", i+1),
			Notes:           seedMarker + " generated demo expense",
			CreatedBy:       createdBy,
			CreatedAt:       expenseDate.Add(2 * time.Hour),
			UpdatedAt:       expenseDate.Add(2 * time.Hour),
		}

		if i%4 == 0 && inventoryPurchase.ID != uuid.Nil {
			product := products[(i*2)%len(products)]
			quantity := round3(6 + float64(i%5))
			expense.CategoryID = inventoryPurchase.ID
			expense.ProductID = &product.ID
			expense.Quantity = &quantity
			expense.AppliesInventory = true
			expense.Amount = round2(product.Cost * quantity)
			expense.Description = fmt.Sprintf("%s inventory restock for %s", seedMarker, product.Name)
		} else {
			category := otherCategories[i%len(otherCategories)]
			expense.CategoryID = category.ID
		}

		expenses = append(expenses, expense)
	}

	return expenses
}

func insertExpenses(ctx context.Context, tx pgx.Tx, expenses []expenseSeed) error {
	for _, expense := range expenses {
		if _, err := tx.Exec(ctx, `
			INSERT INTO expenses (
				id, category_id, amount, description, expense_date, vendor, reference_number, notes,
				created_by, created_at, updated_at, product_id, quantity, applies_inventory
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		`, expense.ID, expense.CategoryID, money(expense.Amount), expense.Description, expense.ExpenseDate.Format("2006-01-02"), expense.Vendor, expense.ReferenceNumber, expense.Notes, expense.CreatedBy, expense.CreatedAt, expense.UpdatedAt, expense.ProductID, nullableQty(expense.Quantity), expense.AppliesInventory); err != nil {
			return err
		}
	}
	return nil
}

func buildAuditLogs(owner userSeed, users []userSeed, products []productSeed, shifts []shiftSeed, sales []saleSeed, expenses []expenseSeed) []auditSeed {
	var logs []auditSeed
	now := time.Now().UTC()
	actors := append([]userSeed{owner}, users...)

	appendLog := func(idx int, actor userSeed, action, entityType, entityID, desc, oldValues, newValues, metadata string) {
		logs = append(logs, auditSeed{
			ID:          seedUUID("audit", idx),
			CreatedAt:   now.Add(-time.Duration(idx) * 17 * time.Minute),
			UserID:      actor.ID,
			UserName:    actor.Name,
			UserRole:    actor.RoleName,
			Action:      action,
			EntityType:  entityType,
			EntityID:    entityID,
			Description: desc,
			OldValues:   oldValues,
			NewValues:   newValues,
			Metadata:    metadata,
			Status:      "success",
		})
	}

	idx := 1
	for _, product := range products[:12] {
		actor := actors[idx%len(actors)]
		appendLog(idx, actor, "product.update", "product", product.ID.String(), fmt.Sprintf("%s updated %s", seedMarker, product.Name), `{"price":"old"}`, fmt.Sprintf(`{"price":"%s","sku":"%s"}`, money(product.Price), product.SKU), `{"source":"seed"}`)
		idx++
	}
	for _, sale := range sales[:10] {
		actor := actors[idx%len(actors)]
		appendLog(idx, actor, "sale.create", "sale", sale.ID.String(), fmt.Sprintf("%s created sale %s", seedMarker, sale.InvoiceNo), `{}`, fmt.Sprintf(`{"invoice_no":"%s","total_amount":"%s"}`, sale.InvoiceNo, money(sale.TotalAmount)), `{"source":"seed"}`)
		idx++
	}
	for _, expense := range expenses[:8] {
		actor := actors[idx%len(actors)]
		appendLog(idx, actor, "expense.create", "expense", expense.ID.String(), fmt.Sprintf("%s created expense %s", seedMarker, expense.ReferenceNumber), `{}`, fmt.Sprintf(`{"amount":"%s","reference_number":"%s"}`, money(expense.Amount), expense.ReferenceNumber), `{"source":"seed"}`)
		idx++
	}
	for _, shift := range shifts[:8] {
		actor := actors[idx%len(actors)]
		appendLog(idx, actor, "shift.close", "shift", shift.ID.String(), fmt.Sprintf("%s processed shift %s", seedMarker, shift.ID.String()[:8]), `{"status":"open"}`, fmt.Sprintf(`{"status":"%s"}`, shift.Status), `{"source":"seed"}`)
		idx++
	}
	for idx <= 45 {
		actor := actors[idx%len(actors)]
		appendLog(idx, actor, "user.login", "session", seedUUID("session", idx).String(), fmt.Sprintf("%s user signed in", seedMarker), `{}`, `{"result":"success"}`, `{"source":"seed"}`)
		idx++
	}

	sort.Slice(logs, func(i, j int) bool {
		return logs[i].CreatedAt.Before(logs[j].CreatedAt)
	})
	return logs
}

func insertAuditLogs(ctx context.Context, tx pgx.Tx, logs []auditSeed) error {
	for _, entry := range logs {
		if _, err := tx.Exec(ctx, `
			INSERT INTO audit_logs (
				id, created_at, user_id, user_name, user_role, action, entity_type, entity_id,
				description, old_values, new_values, metadata, status
			) VALUES (
				$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,
				$11::jsonb,$12::jsonb,$13::jsonb,$14
			)
		`, entry.ID, entry.CreatedAt, entry.UserID, entry.UserName, entry.UserRole, entry.Action, entry.EntityType, entry.EntityID, entry.Description, entry.OldValues, entry.NewValues, entry.Metadata, entry.Status); err != nil {
			return err
		}
	}
	return nil
}

func seedUUID(kind string, idx int) uuid.UUID {
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(fmt.Sprintf("dashpoint-%s-%03d", kind, idx)))
}

func money(value float64) string {
	return fmt.Sprintf("%.2f", round2(value))
}

func qty(value float64) string {
	return fmt.Sprintf("%.3f", round3(value))
}

func nullableMoney(value *float64) any {
	if value == nil {
		return nil
	}
	return money(*value)
}

func nullableQty(value *float64) any {
	if value == nil {
		return nil
	}
	return qty(*value)
}

func nullableFloat(value float64, use bool) *float64 {
	if !use {
		return nil
	}
	v := round2(value)
	return &v
}

func rupiah(value float64) float64 {
	return round2(value * rupiahScale)
}

func round2(value float64) float64 {
	return math.Round(value*100) / 100
}

func round3(value float64) float64 {
	return math.Round(value*1000) / 1000
}

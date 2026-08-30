//go:build integration

package repository

// Concurrency reproduction for the sale-checkout vs stock-adjustment race
// (audit L2): before the fix, checkout read inventory through an unlocked
// join and wrote back a computed remainder, silently clobbering concurrent
// stock adjustments.
//
// Run against a live database:
//
//	TEST_DATABASE_URL=postgres://dashpoint:dashpoint_dev@127.0.0.1:5432/dashpoint_dev?sslmode=disable \
//	  go test -tags=integration ./internal/repository/ -run TestCheckoutStockRace -v
//
// Without TEST_DATABASE_URL the test skips, so plain `go test ./...` (CI)
// is unaffected. Both sides of the race call the real production code paths:
// loadSaleProductForUpdate + setInventoryQuantityTx for checkout, and
// AdjustStockWithTx for adjustments.

import (
	"context"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"dashpoint/backend/internal/models"
)

func TestCheckoutStockRaceNoLostUpdates(t *testing.T) {
	ctx := context.Background()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; integration race test requires a live database")
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer pool.Close()

	productID := uuid.New()
	if _, err := pool.Exec(ctx,
		`INSERT INTO products (id, name, price) VALUES ($1, 'race-test-product', 1)`, productID); err != nil {
		t.Fatalf("failed to create test product: %v", err)
	}
	if _, err := pool.Exec(ctx,
		`INSERT INTO inventory_items (product_id, quantity) VALUES ($1, 0)`, productID); err != nil {
		t.Fatalf("failed to create test inventory: %v", err)
	}
	t.Cleanup(func() {
		pool.Exec(context.Background(), `DELETE FROM stock_adjustments WHERE product_id = $1`, productID)
		pool.Exec(context.Background(), `DELETE FROM inventory_items WHERE product_id = $1`, productID)
		pool.Exec(context.Background(), `DELETE FROM products WHERE id = $1`, productID)
	})

	inventoryRepo := NewInventoryRepository(pool)
	one := decimal.NewFromInt(1)
	five := decimal.NewFromInt(5)
	adjustmentsPerRound, checkoutsPerRound := 4, 6
	// 4 adjustments add +5 each, 6 checkouts subtract 1 each.
	expected := decimal.NewFromInt(int64(adjustmentsPerRound*5 - checkoutsPerRound))

	// stock_adjustments.adjusted_by references users; any existing user works.
	var adjustedBy uuid.UUID
	if err := pool.QueryRow(ctx, `SELECT id FROM users ORDER BY created_at LIMIT 1`).Scan(&adjustedBy); err != nil {
		t.Fatalf("failed to find a user for stock adjustments: %v", err)
	}

	// checkout runs the exact fixed inventory sequence of a sale checkout.
	// Hitting the oversell guard is a legitimate business outcome (the sale
	// would fail at the POS), so it reports success without a decrement.
	var checkoutsOK atomic.Int64
	checkout := func() error {
		tx, err := pool.Begin(ctx)
		if err != nil {
			return err
		}
		product, err := loadSaleProductForUpdate(ctx, tx, productID)
		if err != nil {
			tx.Rollback(ctx)
			return err
		}
		newQty := product.Inventory.Quantity.Sub(one)
		if newQty.LessThan(decimal.Zero) {
			tx.Rollback(ctx)
			return nil
		}
		if err := setInventoryQuantityTx(ctx, tx, productID, newQty, time.Now()); err != nil {
			tx.Rollback(ctx)
			return err
		}
		if err := tx.Commit(ctx); err != nil {
			return err
		}
		checkoutsOK.Add(1)
		return nil
	}

	// adjust runs the real stock-adjustment path (locks the inventory row).
	adjust := func() error {
		tx, err := pool.Begin(ctx)
		if err != nil {
			return err
		}
		if _, err := inventoryRepo.AdjustStockWithTx(ctx, tx, productID, models.AdjustmentPurchase, five, nil, nil, nil, adjustedBy); err != nil {
			tx.Rollback(ctx)
			return err
		}
		return tx.Commit(ctx)
	}

	readQty := func() decimal.Decimal {
		var qty decimal.Decimal
		pool.QueryRow(ctx, `SELECT quantity FROM inventory_items WHERE product_id = $1`, productID).Scan(&qty)
		return qty
	}
	resetQty := func() {
		if _, err := pool.Exec(ctx, `UPDATE inventory_items SET quantity = 0 WHERE product_id = $1`, productID); err != nil {
			t.Fatalf("failed to reset quantity: %v", err)
		}
	}

	runConcurrently := func(workers ...func() error) {
		var wg sync.WaitGroup
		errCh := make(chan error, len(workers))
		for _, worker := range workers {
			wg.Add(1)
			go func(w func() error) {
				defer wg.Done()
				errCh <- w()
			}(worker)
		}
		wg.Wait()
		close(errCh)
		for err := range errCh {
			if err != nil {
				t.Fatalf("concurrent worker failed: %v", err)
			}
		}
	}

	const rounds = 12

	// Fixed path: every round must land on the exact expected quantity,
	// accounting for checkouts the oversell guard legitimately rejected.
	for round := 0; round < rounds; round++ {
		resetQty()
		checkoutsOK.Store(0)
		workers := make([]func() error, 0, adjustmentsPerRound+checkoutsPerRound)
		for i := 0; i < adjustmentsPerRound; i++ {
			workers = append(workers, adjust)
		}
		for i := 0; i < checkoutsPerRound; i++ {
			workers = append(workers, checkout)
		}
		runConcurrently(workers...)

		expectedThisRound := five.Mul(decimal.NewFromInt(int64(adjustmentsPerRound))).Sub(one.Mul(decimal.NewFromInt(checkoutsOK.Load())))
		if final := readQty(); !final.Equal(expectedThisRound) {
			t.Fatalf("round %d: lost update with fixed locking: expected %s (%d successful checkouts), got %s",
				round, expectedThisRound, checkoutsOK.Load(), final)
		}
	}

	// Legacy detection proof: the pre-fix pattern (unlocked read, compute,
	// write) under the same load must diverge at least sometimes — otherwise
	// the assertions above would prove nothing. Not a hard failure: this is
	// timing-dependent evidence that the test can see the race.
	legacyLost := 0
	for round := 0; round < rounds; round++ {
		resetQty()
		legacyCheckout := func() error {
			var qty decimal.Decimal
			if err := pool.QueryRow(ctx,
				`SELECT quantity FROM inventory_items WHERE product_id = $1`, productID).Scan(&qty); err != nil {
				return err
			}
			// Widen the read-compute-write gap the way the original bug did
			// (the sale-item insertion loop sat between read and write).
			time.Sleep(2 * time.Millisecond)
			_, err := pool.Exec(ctx,
				`UPDATE inventory_items SET quantity = $1 WHERE product_id = $2`, qty.Sub(one), productID)
			return err
		}
		workers := make([]func() error, 0, adjustmentsPerRound+checkoutsPerRound)
		for i := 0; i < adjustmentsPerRound; i++ {
			workers = append(workers, adjust)
		}
		for i := 0; i < checkoutsPerRound; i++ {
			workers = append(workers, legacyCheckout)
		}
		runConcurrently(workers...)

		if final := readQty(); !final.Equal(expected) {
			legacyLost++
		}
	}
	t.Logf("legacy unlocked pattern diverged in %d/%d rounds — the race is detectable, and the fixed path never hit it", legacyLost, rounds)
	if legacyLost == 0 {
		t.Log("no legacy divergence observed this run (timing-dependent); fixed-path rounds still all matched exactly")
	}
}

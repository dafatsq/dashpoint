package repository

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"dashpoint/backend/internal/models"
)

type fakeRefreshTx struct {
	execCalls   int
	commitCalls int
	execErrAt   int
	execErr     error
	commitErr   error
}

func (f *fakeRefreshTx) Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error) {
	f.execCalls++
	if f.execErrAt == f.execCalls {
		return pgconn.CommandTag{}, f.execErr
	}
	return pgconn.CommandTag{}, nil
}

func (f *fakeRefreshTx) Commit(context.Context) error {
	f.commitCalls++
	return f.commitErr
}

func (f *fakeRefreshTx) Rollback(context.Context) error { return nil }

type fakeBoolRow struct {
	err   error
	value any
}

func (r fakeBoolRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	switch ptr := dest[0].(type) {
	case *bool:
		*ptr = r.value.(bool)
	case *int:
		*ptr = r.value.(int)
	default:
		return errors.New("unsupported scan target")
	}
	return nil
}

type fakeSalesQuerier struct {
	rows []fakeBoolRow
	idx  int
}

func (f *fakeSalesQuerier) QueryRow(context.Context, string, ...any) pgx.Row {
	row := f.rows[f.idx]
	f.idx++
	return row
}

type fakeCleanupTx struct {
	execCount    int
	failAt       int
	execErr      error
	commitCalled bool
}

func (f *fakeCleanupTx) Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error) {
	f.execCount++
	if f.failAt == f.execCount {
		return pgconn.CommandTag{}, f.execErr
	}
	return pgconn.CommandTag{}, nil
}

func (f *fakeCleanupTx) Commit(context.Context) error {
	f.commitCalled = true
	return nil
}

func (f *fakeCleanupTx) Rollback(context.Context) error { return nil }

func TestRotateRefreshTokenTxSetsReplacementAndCommits(t *testing.T) {
	tx := &fakeRefreshTx{}
	replacement := &models.RefreshToken{
		UserID:    uuid.New(),
		TokenHash: "new-token-hash",
		ExpiresAt: time.Now().Add(time.Hour),
	}

	if err := rotateRefreshTokenTx(context.Background(), tx, "old-token-hash", "token_refresh", replacement); err != nil {
		t.Fatalf("rotateRefreshTokenTx returned error: %v", err)
	}
	if replacement.ID == uuid.Nil {
		t.Fatalf("expected replacement ID to be generated")
	}
	if replacement.CreatedAt.IsZero() {
		t.Fatalf("expected replacement CreatedAt to be set")
	}
	if tx.execCalls != 2 {
		t.Fatalf("expected two Exec calls, got %d", tx.execCalls)
	}
	if tx.commitCalls != 1 {
		t.Fatalf("expected one Commit call, got %d", tx.commitCalls)
	}
}

func TestRotateRefreshTokenTxReturnsStepSpecificError(t *testing.T) {
	tx := &fakeRefreshTx{execErrAt: 2, execErr: errors.New("insert failed")}
	replacement := &models.RefreshToken{
		UserID:    uuid.New(),
		TokenHash: "new-token-hash",
		ExpiresAt: time.Now().Add(time.Hour),
	}

	err := rotateRefreshTokenTx(context.Background(), tx, "old-token-hash", "token_refresh", replacement)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "failed to store rotated refresh token") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHasSalesHistoryReturnsCount(t *testing.T) {
	querier := &fakeSalesQuerier{
		rows: []fakeBoolRow{
			{value: 1},
		},
	}

	hasSales, err := hasSalesHistory(context.Background(), querier, uuid.New())
	if err != nil {
		t.Fatalf("hasSalesHistory returned error: %v", err)
	}
	if !hasSales {
		t.Fatalf("expected sales history to be detected")
	}
}

func TestHasSalesHistoryReturnsQueryError(t *testing.T) {
	querier := &fakeSalesQuerier{
		rows: []fakeBoolRow{
			{err: errors.New("db unavailable")},
		},
	}

	_, err := hasSalesHistory(context.Background(), querier, uuid.New())
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "failed to query sales history") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPermanentDeleteTxReturnsNamedCleanupStep(t *testing.T) {
	tx := &fakeCleanupTx{failAt: 4, execErr: errors.New("constraint failure")}

	err := permanentDeleteTx(context.Background(), tx, uuid.New())
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "failed to nullify sales voids during permanent delete") {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx.commitCalled {
		t.Fatalf("did not expect commit on failure")
	}
}

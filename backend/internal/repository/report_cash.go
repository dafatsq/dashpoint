package repository

import (
	"context"
	"fmt"
	"time"
)

// GetCashReport gets cash report for a date range.
func (r *ReportRepository) GetCashReport(ctx context.Context, startDate, endDate time.Time) (*CashReport, error) {
	start := startOfReportDay(startDate)
	end := exclusiveEndDate(endDate)
	report := &CashReport{
		Date: fmt.Sprintf("%s to %s", start.Format("2006-01-02"), startOfReportDay(endDate).Format("2006-01-02")),
	}

	if err := r.pool.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COALESCE(SUM(opening_cash), 0),
			COALESCE(SUM(closing_cash), 0)
		FROM shifts
		WHERE started_at >= $1 AND started_at < $2 AND status = 'closed'
	`, start, end).Scan(&report.ShiftCount, &report.OpeningCash, &report.ActualCash); err != nil {
		return nil, err
	}

	if err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(p.amount), 0)
		FROM payments p
		JOIN sales s ON p.sale_id = s.id
		WHERE s.created_at >= $1 AND s.created_at < $2
		AND s.status = 'completed' AND p.payment_method = 'cash'
	`, start, end).Scan(&report.CashSales); err != nil {
		return nil, err
	}

	if err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(p.amount), 0)
		FROM payments p
		JOIN sales s ON p.sale_id = s.id
		WHERE s.created_at >= $1 AND s.created_at < $2
		AND s.status = 'voided' AND p.payment_method = 'cash' AND p.status = 'refunded'
	`, start, end).Scan(&report.CashVoidedSales); err != nil {
		return nil, err
	}

	if err := r.pool.QueryRow(ctx, `
		SELECT
			COALESCE(SUM(CASE WHEN cdo.type = 'pay_in' THEN cdo.amount ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN cdo.type = 'pay_out' THEN cdo.amount ELSE 0 END), 0)
		FROM cash_drawer_operations cdo
		JOIN shifts sh ON cdo.shift_id = sh.id
		WHERE sh.started_at >= $1 AND sh.started_at < $2
	`, start, end).Scan(&report.PayInTotal, &report.PayOutTotal); err != nil {
		return nil, err
	}

	report.ExpectedCash = report.OpeningCash.
		Add(report.CashSales).
		Add(report.PayInTotal).
		Sub(report.PayOutTotal)
	report.Difference = report.ActualCash.Sub(report.ExpectedCash)

	return report, nil
}

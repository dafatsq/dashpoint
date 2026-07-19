import { describe, expect, test } from "vitest";

import type { CashDrawerOperation, Shift } from "@/types";

import {
  buildShiftQueryParams,
  formatShiftCurrency,
  getShiftCashDifferenceTone,
  getShiftOperationLabel,
  getShiftSummary,
} from "./shifts-helpers";

function buildShift(overrides: Partial<Shift> = {}): Shift {
  return {
    id: "shift-1",
    opened_by: "user-1",
    started_at: "2026-01-01T00:00:00Z",
    status: "closed",
    opening_cash: "100000",
    closing_cash: "120000",
    expected_cash: "110000",
    cash_difference: "10000",
    total_sales: "50000",
    total_cash_sales: "50000",
    total_voided: "0",
    transaction_count: 3,
    void_count: 0,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
    operations: [],
    ...overrides,
  };
}

describe("shifts helpers", () => {
  test("builds typed shifts query params", () => {
    expect(
      buildShiftQueryParams({
        page: 2,
        limit: 10,
        dateRange: { start: "2026-01-01", end: "2026-01-02" },
        selectedUser: "user-1",
      }),
    ).toEqual({
      limit: 10,
      offset: 10,
      from: "2026-01-01",
      to: "2026-01-02",
      opened_by_id: "user-1",
    });
  });

  test("derives shift cash summaries and tones", () => {
    const summary = getShiftSummary(buildShift());
    expect(summary.isOpen).toBe(false);
    expect(summary.openingCash).toBe(100000);
    expect(summary.cashDifference).toBe(10000);
    expect(getShiftCashDifferenceTone(summary.cashDifference)).toBe("text-green-600");
    expect(getShiftCashDifferenceTone(-1)).toBe("text-red-600");
  });

  test("formats operation labels and currency", () => {
    const operation = { id: "op-1", type: "pay_in", amount: "5000", reason: "Float", shift_id: "shift-1", performed_by: "u1", created_at: "2026-01-01T00:00:00Z" } as CashDrawerOperation;
    expect(getShiftOperationLabel(operation)).toBe("+ PAY IN");
    expect(formatShiftCurrency(5000)).toContain("Rp");
  });
});

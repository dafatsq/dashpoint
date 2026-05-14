import { describe, expect, test } from "vitest";

import type { Sale } from "@/types";

import {
  filterSalesBySearch,
  formatSalesCurrency,
  getPrimarySalePaymentMethod,
  getSalesStatusBadge,
} from "./sales-helpers";

function buildSale(overrides: Partial<Sale> = {}): Sale {
  return {
    id: "sale-1",
    invoice_no: "INV-001",
    employee_id: "user-1",
    employee_name: "Alice",
    status: "completed",
    subtotal: "10000",
    tax_amount: "0",
    discount_amount: "0",
    total_amount: "10000",
    amount_paid: "10000",
    change_amount: "0",
    item_count: 1,
    payment_status: "paid",
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
    payments: [{ payment_method: "cash", amount: "10000" }],
    ...overrides,
  };
}

describe("sales helpers", () => {
  test("returns the primary payment method with a cash fallback", () => {
    expect(getPrimarySalePaymentMethod(buildSale())).toBe("cash");
    expect(getPrimarySalePaymentMethod(buildSale({ payments: [] }))).toBe("cash");
  });

  test("maps status badges", () => {
    expect(getSalesStatusBadge("completed")).toContain("bg-green-600");
    expect(getSalesStatusBadge("voided")).toContain("bg-red-600");
  });

  test("filters sales by invoice and employee name", () => {
    const sales = [
      buildSale(),
      buildSale({ id: "sale-2", invoice_no: "INV-002", employee_name: "Bob" }),
    ];

    expect(filterSalesBySearch(sales, "inv-001")).toHaveLength(1);
    expect(filterSalesBySearch(sales, "bob")).toHaveLength(1);
    expect(filterSalesBySearch(sales, "")).toHaveLength(2);
  });

  test("formats sales currency consistently", () => {
    expect(formatSalesCurrency("15000")).toContain("Rp");
  });
});

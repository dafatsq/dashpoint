import { describe, expect, test } from "vitest";

import type { AuditLog } from "@/types";

import {
  formatDashboardFieldName,
  formatDashboardFieldValue,
  getDashboardActionBadgeColor,
  getDashboardActionVerb,
  getDashboardChangeDescription,
  getDashboardFieldChanges,
  isDashboardImageField,
} from "./dashboard-helpers";

function buildAuditLog(overrides: Partial<AuditLog> = {}): AuditLog {
  return {
    id: "log-1",
    user_id: "user-1",
    user_name: "Alice",
    action: "product.update",
    entity_type: "product",
    entity_id: "entity-1",
    old_values: { name: "Old Coffee", tax_rate: "10" },
    new_values: { name: "New Coffee", tax_rate: "11" },
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
    ...overrides,
  } as AuditLog;
}

describe("dashboard helpers", () => {
  test("parses the action verb from dotted actions", () => {
    expect(getDashboardActionVerb("product.create")).toBe("create");
    expect(getDashboardActionVerb("inventory.threshold_update")).toBe("threshold_update");
    expect(getDashboardActionVerb("archive")).toBe("archive");
  });

  test("maps badge colors by action verb", () => {
    expect(getDashboardActionBadgeColor("product.create")).toBe("bg-green-600 text-white");
    expect(getDashboardActionBadgeColor("sale.void")).toBe("bg-orange-600 text-white");
  });

  test("builds human-readable change descriptions", () => {
    expect(
      getDashboardChangeDescription(
        buildAuditLog({
          entity_type: "inventory",
          action: "inventory.adjust",
          new_values: { product_name: "Beans", adjustment_type: "adjustment", quantity: "-2", new_quantity: "8" },
          old_values: {},
        }),
      ),
    ).toContain("Beans");
    expect(
      getDashboardChangeDescription(
        buildAuditLog({
          entity_type: "inventory",
          action: "inventory.threshold_update",
          old_values: { product_name: "Beans", low_stock_threshold: "5" },
          new_values: { product_name: "Beans", low_stock_threshold: "12" },
        }),
      ),
    ).toBe("Updated low stock threshold: Beans (5 → 12)");
    expect(
      getDashboardChangeDescription(
        buildAuditLog({
          entity_type: "sale",
          action: "sale.create",
          new_values: { invoice_no: "INV-1" },
          old_values: {},
        }),
      ),
    ).toBe("INV-1");
  });

  test("collects changed fields while skipping non-display keys", () => {
    const changes = getDashboardFieldChanges(
      buildAuditLog({
        old_values: { name: "Old", image_url: "/old.jpg", product_id: "skip-me" },
        new_values: { name: "New", image_url: "/new.jpg", product_id: "skip-me" },
      }),
    );

    expect(changes.map((change) => change.key)).toEqual(["name", "image_url"]);
  });

  test("formats field names and values for display", () => {
    expect(formatDashboardFieldName("image_url", "product.update")).toBe("Photo");
    expect(formatDashboardFieldName("reason", "sale.void")).toBe("Reason");
    expect(formatDashboardFieldValue("tax_rate", "11")).toBe("11%");
    expect(formatDashboardFieldValue("amount", "15000")).toContain("Rp");
  });

  test("flags image fields correctly", () => {
    expect(isDashboardImageField("image_url")).toBe(true);
    expect(isDashboardImageField("name")).toBe(false);
  });
});

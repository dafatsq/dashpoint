import { describe, expect, test } from "vitest";

import type { AuditLog } from "@/types";

import {
  buildActivityDescription,
  buildActivityFieldChanges,
  buildAuditLogParams,
  buildChangesListResetKey,
  buildDashboardChangeParams,
  formatActivityFieldValue,
  getActivityActionLabel,
  getActivityActionVerb,
  getActivityBadgeColor,
  getActivityEntityLabel,
  incrementActivityRefreshKey,
} from "./activity-helpers";

function normalizeCurrency(value: string): string {
  return value.replace(/\s+/g, "");
}

const baseLog: AuditLog = {
  id: "1",
  user_id: "u1",
  user_name: "Alice",
  action: "product.update",
  entity_type: "product",
  entity_id: "p1",
  old_values: {},
  new_values: {},
  created_at: "2026-05-14T12:00:00Z",
};

describe("activity helpers", () => {
  test("parses dotted action verbs and labels", () => {
    expect(getActivityActionVerb("inventory.adjust")).toBe("adjust");
    expect(getActivityActionLabel("auth.login_failed")).toBe("Login Failed");
  });

  test("maps entity labels and badge colors", () => {
    expect(getActivityEntityLabel("auth")).toBe("Authentication");
    expect(getActivityBadgeColor("category.archive")).toContain(
      "bg-orange-500",
    );
    expect(getActivityBadgeColor("sale.void")).toContain("bg-orange-600");
  });

  test("builds descriptive activity summaries", () => {
    expect(
      buildActivityDescription({
        ...baseLog,
        entity_type: "inventory",
        action: "inventory.adjust",
        new_values: {
          product_name: "Milk",
          adjustment_type: "restock",
          quantity: 5,
          new_quantity: 12,
        },
      }),
    ).toBe("Restock stock: Milk -> 12 (Δ5)");

    expect(
      buildActivityDescription({
        ...baseLog,
        entity_type: "expense",
        action: "expense.create",
        new_values: {
          description: "Delivery",
          amount: "15000",
        },
      }),
    ).toBe("Created expense: Delivery — Rp 15.000");
  });

  test("builds filtered field changes", () => {
    const changes = buildActivityFieldChanges({
      ...baseLog,
      old_values: {
        name: "Old Name",
        image_url: "/uploads/old.jpg",
        affected_product: "Ignored",
      },
      new_values: {
        name: "New Name",
        image_url: "/uploads/new.jpg",
        affected_product: "Ignored",
      },
    });

    expect(changes).toEqual([
      { key: "name", oldVal: "Old Name", newVal: "New Name" },
      {
        key: "image_url",
        oldVal: "/uploads/old.jpg",
        newVal: "/uploads/new.jpg",
      },
    ]);
  });

  test("formats activity field values consistently", () => {
    expect(normalizeCurrency(formatActivityFieldValue("amount", "15000"))).toBe(
      "Rp15.000",
    );
    expect(formatActivityFieldValue("tax_rate", 11)).toBe("11%");
    expect(formatActivityFieldValue("is_active", true)).toBe("Yes");
    expect(formatActivityFieldValue("metadata", { foo: "bar" })).toBe(
      '{"foo":"bar"}',
    );
  });

  test("builds typed audit and dashboard change params", () => {
    expect(
      buildAuditLogParams({
        page: 2,
        limit: 50,
        selectedAction: "create",
        selectedEntity: "product",
        dateRange: { start: "2026-05-01", end: "2026-05-14" },
      }),
    ).toEqual({
      action: "create",
      entity_type: "product",
      from: "2026-05-01",
      to: "2026-05-14",
      limit: 50,
      offset: 50,
    });

    expect(
      buildDashboardChangeParams({
        entityType: "sale",
        page: 3,
        limit: 10,
        dateRange: { start: "2026-05-01", end: "" },
        selectedUser: "user-1",
      }),
    ).toEqual({
      entity_type: "sale",
      user_id: "user-1",
      from: "2026-05-01",
      limit: 10,
      offset: 20,
    });
  });

  test("builds stable reset keys for changes lists", () => {
    expect(
      buildChangesListResetKey({
        entityType: "product",
        dateRange: { start: "", end: "" },
        selectedUser: "all",
      }),
    ).toBe("product|all||");

    expect(
      buildChangesListResetKey({
        entityType: "product",
        dateRange: { start: "2026-05-01", end: "2026-05-14" },
        selectedUser: "user-1",
      }),
    ).toBe("product|user-1|2026-05-01|2026-05-14");
  });

  test("increments activity refresh keys for retry flows", () => {
    expect(incrementActivityRefreshKey(0)).toBe(1);
    expect(incrementActivityRefreshKey(7)).toBe(8);
  });
});

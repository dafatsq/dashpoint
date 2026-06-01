import { describe, expect, test } from "vitest";

import type { Product } from "@/types";

import {
  ADJUSTMENT_TYPE_OPTIONS,
  INVENTORY_HISTORY_FILTER_OPTIONS,
  buildInventoryAdjustmentRequest,
  canSubmitInventoryAdjustment,
  classifyInventoryStock,
  createEmptyAdjustmentFormState,
  getDefaultAdjustmentType,
  getInventoryAdjustmentChangeLabel,
  getInventoryAdjustmentTypeLabel,
  getInventoryProductMinQuantity,
  getInventoryProductPrice,
  getInventoryProductQuantity,
  getInventoryProductImageUrl,
  getPermittedInventoryActions,
  requiresInventoryAdjustmentReason,
} from "./inventory-helpers";

function buildProduct(overrides: Partial<Product> = {}): Product {
  return {
    id: "product-1",
    name: "Coffee",
    price: "15000",
    cost: "10000",
    is_active: true,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
    inventory: {
      quantity: "5",
      available_quantity: "5",
      low_stock_threshold: "2",
      is_low_stock: false,
    },
    ...overrides,
  };
}

describe("inventory helpers", () => {
  test("parses quantity, minimum quantity, and price safely", () => {
    const product = buildProduct();

    expect(getInventoryProductQuantity(product)).toBe(5);
    expect(getInventoryProductMinQuantity(product)).toBe(2);
    expect(getInventoryProductPrice(product)).toBe(15000);
  });

  test("classifies low-stock and out-of-stock states", () => {
    expect(
      classifyInventoryStock(
        buildProduct({
          inventory: { quantity: "0", available_quantity: "0", low_stock_threshold: "2", is_low_stock: false },
        }),
      ),
    ).toEqual({ isLowStock: false, isOutOfStock: true });

    expect(
      classifyInventoryStock(
        buildProduct({
          inventory: { quantity: "2", available_quantity: "2", low_stock_threshold: "2", is_low_stock: false },
        }),
      ),
    ).toEqual({ isLowStock: true, isOutOfStock: false });
  });

  test("returns permitted inventory actions from permissions", () => {
    expect(getPermittedInventoryActions({ canAddStock: true, canRemoveStock: false, canAdjustStock: true })).toEqual([
      "add",
      "count",
    ]);
  });

  test("requires at least one allowed action and a permitted selected action before submit", () => {
    expect(
      canSubmitInventoryAdjustment({
        allowedActions: [],
        action: "add",
        adjustmentType: "purchase",
        quantity: "5",
        notes: "",
        isSubmitting: false,
      }),
    ).toBe(false);

    expect(
      canSubmitInventoryAdjustment({
        allowedActions: ["count"],
        action: "add",
        adjustmentType: "purchase",
        quantity: "5",
        notes: "",
        isSubmitting: false,
      }),
    ).toBe(false);

    expect(
      canSubmitInventoryAdjustment({
        allowedActions: ["count"],
        action: "count",
        adjustmentType: "count",
        quantity: "5",
        notes: "",
        isSubmitting: false,
      }),
    ).toBe(true);
    expect(
      canSubmitInventoryAdjustment({
        allowedActions: ["count"],
        action: "count",
        adjustmentType: "count",
        quantity: "",
        notes: "",
        isSubmitting: false,
      }),
    ).toBe(false);
  });

  test("requires notes for destructive inventory adjustments", () => {
    expect(requiresInventoryAdjustmentReason("damage", "remove")).toBe(true);
    expect(requiresInventoryAdjustmentReason("loss", "remove")).toBe(true);
    expect(requiresInventoryAdjustmentReason("adjustment", "remove")).toBe(true);
    expect(requiresInventoryAdjustmentReason("purchase", "add")).toBe(false);
    expect(requiresInventoryAdjustmentReason("adjustment", "add")).toBe(false);

    expect(
      canSubmitInventoryAdjustment({
        allowedActions: ["remove"],
        action: "remove",
        adjustmentType: "damage",
        quantity: "2",
        notes: "",
        isSubmitting: false,
      }),
    ).toBe(false);

    expect(
      canSubmitInventoryAdjustment({
        allowedActions: ["remove"],
        action: "remove",
        adjustmentType: "damage",
        quantity: "2",
        notes: "Broken seal",
        isSubmitting: false,
      }),
    ).toBe(true);
  });

  test("returns the correct default backend adjustment type", () => {
    expect(getDefaultAdjustmentType("add")).toBe("purchase");
    expect(getDefaultAdjustmentType("remove")).toBe("damage");
    expect(getDefaultAdjustmentType("count")).toBe("count");
  });

  test("exposes adjustment type options by action", () => {
    expect(ADJUSTMENT_TYPE_OPTIONS.add.map((item) => item.value)).toEqual(["purchase", "adjustment"]);
    expect(ADJUSTMENT_TYPE_OPTIONS.remove.map((item) => item.value)).toEqual(["damage", "loss", "adjustment"]);
    expect(ADJUSTMENT_TYPE_OPTIONS.count.map((item) => item.value)).toEqual(["count"]);
  });

  test("exposes history filter options for the drawer", () => {
    expect(INVENTORY_HISTORY_FILTER_OPTIONS.map((item) => item.value)).toEqual([
      "all",
      "purchase",
      "sale",
      "return",
      "adjustment",
      "damage",
      "loss",
      "count",
      "initial",
      "transfer",
    ]);
  });

  test("builds stock-count requests as absolute quantity", () => {
    expect(
      buildInventoryAdjustmentRequest({
        productId: "product-1",
        action: "count",
        adjustmentType: "count",
        quantity: "7",
        currentStock: 5,
        notes: "Cycle count",
      }),
    ).toEqual({
      product_id: "product-1",
      adjustment_type: "count",
      quantity: "7",
      reason: "Cycle count",
    });
  });

  test("builds removal adjustments as negative deltas for correction type", () => {
    expect(
      buildInventoryAdjustmentRequest({
        productId: "product-1",
        action: "remove",
        adjustmentType: "adjustment",
        quantity: "3",
        currentStock: 5,
      }),
    ).toEqual({
      product_id: "product-1",
      adjustment_type: "adjustment",
      quantity: "-3",
      reason: undefined,
    });
  });

  test("builds damage/loss requests as positive quantities for backend handling", () => {
    expect(
      buildInventoryAdjustmentRequest({
        productId: "product-1",
        action: "remove",
        adjustmentType: "damage",
        quantity: "2",
        currentStock: 5,
      }),
    ).toEqual({
      product_id: "product-1",
      adjustment_type: "damage",
      quantity: "2",
      reason: undefined,
    });
  });

  test("creates a reset adjustment form state", () => {
    expect(createEmptyAdjustmentFormState("add")).toEqual({
      action: "add",
      quantity: "",
      adjustmentType: "purchase",
      notes: "",
    });
  });

  test("formats adjustment type and quantity labels for stock history", () => {
    expect(getInventoryAdjustmentTypeLabel("purchase")).toBe("Restock");
    expect(getInventoryAdjustmentTypeLabel("sale")).toBe("Sale");
    expect(getInventoryAdjustmentTypeLabel("return")).toBe("Return");
    expect(getInventoryAdjustmentTypeLabel("adjustment")).toBe("Correction");
    expect(getInventoryAdjustmentTypeLabel("initial")).toBe("Initial Stock");
    expect(getInventoryAdjustmentTypeLabel("transfer")).toBe("Transfer");
    expect(getInventoryAdjustmentTypeLabel("count")).toBe("Stock Count");

    expect(
      getInventoryAdjustmentChangeLabel({
        adjustment_type: "purchase",
        quantity_change: "4",
        quantity_after: "9",
      }),
    ).toBe("+4");

    expect(
      getInventoryAdjustmentChangeLabel({
        adjustment_type: "loss",
        quantity_change: "-2",
        quantity_after: "7",
      }),
    ).toBe("-2");

    expect(
      getInventoryAdjustmentChangeLabel({
        adjustment_type: "count",
        quantity_change: "3",
        quantity_after: "12",
      }),
    ).toBe("Set to 12");
  });

  test("builds image urls only when a path exists", () => {
    expect(getInventoryProductImageUrl("/uploads/a.jpg", (path) => `https://x${path}`)).toBe("https://x/uploads/a.jpg");
    expect(getInventoryProductImageUrl("", (path) => path)).toBe("");
  });
});

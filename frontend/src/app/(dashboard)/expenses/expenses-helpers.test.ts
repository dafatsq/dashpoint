import { describe, expect, test } from "vitest";

import type { Expense, ExpenseCategory, Product } from "@/types";

import {
  buildExpenseRequest,
  canSubmitExpenseForm,
  createEmptyExpenseFormData,
  deriveInventoryPurchaseAmount,
  deriveInventoryPurchaseDescription,
  hasExpenseFormChanges,
  isInventoryPurchaseCategory,
  mapExpenseToFormData,
  todayDateString,
} from "./expenses-helpers";

function buildCategory(overrides: Partial<ExpenseCategory> = {}): ExpenseCategory {
  return {
    id: "category-1",
    name: "Inventory Purchase",
    system_key: "inventory_purchase",
    is_active: true,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
    ...overrides,
  };
}

function buildProduct(overrides: Partial<Product> = {}): Product {
  return {
    id: "product-1",
    name: "Beans",
    description: "",
    price: "12000",
    cost: "7000",
    is_active: true,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
    ...overrides,
  } as Product;
}

function buildExpense(overrides: Partial<Expense> = {}): Expense {
  return {
    id: "expense-1",
    category_id: "category-1",
    category_name: "Inventory Purchase",
    product_id: "product-1",
    product_name: "Beans",
    quantity: "3",
    applies_inventory: true,
    amount: "21000",
    description: "Beans x 3",
    expense_date: "2026-05-14",
    vendor: "Vendor A",
    reference_number: "REF-1",
    notes: "test",
    created_by: "user-1",
    created_at: "2026-05-14T00:00:00Z",
    updated_at: "2026-05-14T00:00:00Z",
    ...overrides,
  };
}

describe("expenses helpers", () => {
  test("detects inventory purchase category by current category name", () => {
    expect(isInventoryPurchaseCategory("category-1", [buildCategory()])).toBe(true);
    expect(
      isInventoryPurchaseCategory("category-2", [buildCategory({ id: "category-2", name: "Utilities", system_key: undefined })]),
    ).toBe(false);
  });

  test("derives amount from product cost and quantity", () => {
    expect(deriveInventoryPurchaseAmount("product-1", "3", [buildProduct()])).toBe("21000.00");
  });

  test("derives description from product name and quantity", () => {
    expect(deriveInventoryPurchaseDescription("product-1", "3", [buildProduct()])).toBe("Beans x 3");
  });

  test("maps an expense into editable form data", () => {
    expect(mapExpenseToFormData(buildExpense())).toEqual({
      category_id: "category-1",
      product_id: "product-1",
      quantity: "3",
      applies_inventory: true,
      amount: "21000",
      description: "Beans x 3",
      expense_date: "2026-05-14",
      vendor: "Vendor A",
      reference_number: "REF-1",
      notes: "test",
    });
  });

  test("detects expense form changes", () => {
    const expense = buildExpense();
    const sameForm = mapExpenseToFormData(expense);
    const changedForm = { ...sameForm, notes: "changed" };

    expect(hasExpenseFormChanges(sameForm, expense)).toBe(false);
    expect(hasExpenseFormChanges(changedForm, expense)).toBe(true);
  });

  test("derives today using Jakarta local date semantics", () => {
    expect(todayDateString(new Date("2026-05-28T20:00:00.000Z"))).toBe("2026-05-29");
  });

  test("builds expense request with optional empty fields removed", () => {
    const formData = {
      ...createEmptyExpenseFormData("2026-05-14"),
      category_id: "category-1",
      amount: "21000",
      description: "Beans x 3",
      product_id: "",
      quantity: "",
      applies_inventory: false,
      vendor: "",
      reference_number: "",
      notes: "",
    };

    expect(buildExpenseRequest(formData)).toEqual({
      category_id: "category-1",
      applies_inventory: false,
      amount: "21000",
      description: "Beans x 3",
      expense_date: "2026-05-14",
    });
  });

  test("requires inventory purchase dialogs to keep a valid product selection before submit", () => {
    expect(
      canSubmitExpenseForm({
        isSubmitting: false,
        hasChanges: true,
        categoryId: "category-1",
        isInventoryPurchase: true,
        productId: "",
        quantity: "2",
      }),
    ).toBe(false);

    expect(
      canSubmitExpenseForm({
        isSubmitting: false,
        hasChanges: true,
        categoryId: "category-1",
        isInventoryPurchase: true,
        productId: "product-1",
        quantity: "",
      }),
    ).toBe(false);

    expect(
      canSubmitExpenseForm({
        isSubmitting: false,
        hasChanges: true,
        categoryId: "category-1",
        isInventoryPurchase: true,
        productId: "product-1",
        quantity: "2",
      }),
    ).toBe(true);

    expect(
      canSubmitExpenseForm({
        isSubmitting: false,
        hasChanges: true,
        categoryId: "none",
        isInventoryPurchase: false,
        productId: "",
        quantity: "",
      }),
    ).toBe(false);

    expect(
      canSubmitExpenseForm({
        isSubmitting: false,
        hasChanges: true,
        categoryId: "category-2",
        isInventoryPurchase: false,
        productId: "",
        quantity: "",
      }),
    ).toBe(true);
  });
});

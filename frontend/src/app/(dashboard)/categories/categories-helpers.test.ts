import { describe, expect, test } from "vitest";

import type { Category, ExpenseCategory } from "@/types";

import {
  createEmptyCategoryFormData,
  filterCategories,
  hasCategoryFormChanges,
  mapCategoryToFormData,
  resolveCategoryActionLabels,
  type CategoryFormData,
} from "./categories-helpers";

function buildProductCategory(overrides: Partial<Category> = {}): Category {
  return {
    id: "product-category-1",
    name: "Coffee",
    description: "Coffee items",
    product_count: 5,
    is_active: true,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
    ...overrides,
  };
}

function buildExpenseCategory(overrides: Partial<ExpenseCategory> = {}): ExpenseCategory {
  return {
    id: "expense-category-1",
    name: "Utilities",
    description: "Bills",
    is_active: true,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
    ...overrides,
  };
}

describe("categories helpers", () => {
  test("creates an empty category form", () => {
    expect(createEmptyCategoryFormData()).toEqual({
      name: "",
      description: "",
    });
  });

  test("maps category data into editable form data", () => {
    expect(mapCategoryToFormData(buildProductCategory())).toEqual({
      name: "Coffee",
      description: "Coffee items",
    });
  });

  test("detects category form changes", () => {
    const category = buildProductCategory();
    const sameForm = mapCategoryToFormData(category);
    const changedForm: CategoryFormData = { ...sameForm, description: "Updated" };

    expect(hasCategoryFormChanges(sameForm, category)).toBe(false);
    expect(hasCategoryFormChanges(changedForm, category)).toBe(true);
  });

  test("filters categories by search and active view", () => {
    const categories = [
      buildProductCategory(),
      buildProductCategory({ id: "2", name: "Tea", is_active: false }),
    ];

    expect(filterCategories(categories, "cof", "active")).toHaveLength(1);
    expect(filterCategories(categories, "", "archived")).toHaveLength(1);
  });

  test("filters expense categories with the same helper path", () => {
    const categories = [
      buildExpenseCategory(),
      buildExpenseCategory({ id: "2", name: "Travel", is_active: false }),
    ];

    expect(filterCategories(categories, "util", "active")).toHaveLength(1);
    expect(filterCategories(categories, "", "archived")).toHaveLength(1);
  });

  test("returns action labels based on archived vs active mode", () => {
    expect(resolveCategoryActionLabels("active")).toEqual({
      emptyTitle: "No active categories found",
      emptyDescription: "Try adjusting your search or add a new category to get started.",
      deleteActionLabel: "Archive Category",
    });
    expect(resolveCategoryActionLabels("archived")).toEqual({
      emptyTitle: "No archived categories found",
      emptyDescription: "Archived categories will appear here.",
      deleteActionLabel: "Delete Permanently",
    });
  });
});

import { describe, expect, test } from "vitest";

import type { Product } from "@/types";

import {
  buildProductCreateRequest,
  buildProductUpdateRequest,
  createEmptyProductFormData,
  getProductDisplayImageUrl,
  getProductLowStockThreshold,
  getProductPriceValue,
  getProductQuantityValue,
  hasProductFormChanges,
  mapProductToFormData,
  type ProductFormData,
} from "./products-helpers";

function buildProduct(overrides: Partial<Product> = {}): Product {
  return {
    id: "product-1",
    name: "Milk",
    description: "Fresh milk",
    sku: "SKU-1",
    barcode: "BAR-1",
    price: "15000",
    cost: "10000",
    tax_rate: "11",
    category_id: "category-1",
    category_name: "Dairy",
    is_active: true,
    track_inventory: true,
    allow_negative_stock: false,
    image_url: "/uploads/milk.jpg",
    inventory: {
      quantity: "12",
      available_quantity: "12",
      low_stock_threshold: "5",
      is_low_stock: false,
    },
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
    ...overrides,
  };
}

describe("products helpers", () => {
  test("parses price, quantity, and low stock threshold from product data", () => {
    const product = buildProduct();

    expect(getProductPriceValue(product)).toBe(15000);
    expect(getProductQuantityValue(product)).toBe(12);
    expect(getProductLowStockThreshold(product)).toBe(5);
  });

  test("falls back safely for missing or invalid numeric product data", () => {
    const product = buildProduct({
      price: "bad",
      inventory: {
        quantity: "bad",
        available_quantity: "0",
        low_stock_threshold: "",
        is_low_stock: false,
      },
    });

    expect(getProductPriceValue(product)).toBe(0);
    expect(getProductQuantityValue(product)).toBe(0);
    expect(getProductLowStockThreshold(product)).toBe(0);
  });

  test("maps an existing product into editable form data", () => {
    const product = buildProduct();

    expect(mapProductToFormData(product)).toEqual({
      name: "Milk",
      description: "Fresh milk",
      sku: "SKU-1",
      barcode: "BAR-1",
      price: "15000",
      cost: "10000",
      tax_rate: "11",
      initial_quantity: "",
      low_stock_threshold: "5",
      category_id: "category-1",
      image_url: "/uploads/milk.jpg",
    });
  });

  test("detects edit-form changes including image clear operations", () => {
    const product = buildProduct();
    const sameForm = mapProductToFormData(product);
    const changedForm: ProductFormData = { ...sameForm, image_url: "" };

    expect(hasProductFormChanges(sameForm, product)).toBe(false);
    expect(hasProductFormChanges(changedForm, product)).toBe(true);
  });

  test("builds create product payload with optional empty fields removed", () => {
    const formData: ProductFormData = {
      ...createEmptyProductFormData(),
      name: "Milk",
      price: "15000",
      tax_rate: "",
      image_url: "",
    };

    expect(buildProductCreateRequest(formData)).toEqual({
      name: "Milk",
      price: "15000",
      tax_rate: "0",
      low_stock_threshold: "5",
    });
  });

  test("builds update product payload while preserving explicit image clear", () => {
    const formData: ProductFormData = {
      ...createEmptyProductFormData(),
      name: "Milk",
      price: "15000",
      image_url: "",
    };

    expect(buildProductUpdateRequest(formData)).toEqual({
      name: "Milk",
      price: "15000",
      tax_rate: "0",
      image_url: "",
    });
  });

  test("builds display image url only when an image path exists", () => {
    expect(getProductDisplayImageUrl("/uploads/milk.jpg", (path) => `https://cdn.test${path}`)).toBe(
      "https://cdn.test/uploads/milk.jpg",
    );
    expect(getProductDisplayImageUrl("", (path) => path)).toBe("");
  });
});

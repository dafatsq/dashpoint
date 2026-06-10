import { describe, expect, test } from "vitest";

import type { CartItem, Product } from "@/types";

import {
  addCartItem,
  buildSaleRequest,
  buildSaleCartValidationRequest,
  canSubmitEndShift,
  canSubmitStartShift,
  calculateCartTotals,
  classifyStock,
  formatCurrencyInputValue,
  getProductMinQuantity,
  getProductPrice,
  getProductQuantity,
  roundCurrencyAmount,
  removeCartItem,
  updateCartItemQuantity,
} from "./pos-helpers";

function createProduct(overrides: Partial<Product> = {}): Product {
  return {
    id: "product-1",
    name: "Coffee",
    price: "15000",
    cost: "10000",
    tax_rate: "10",
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

describe("POS helpers", () => {
  test("parses product string values safely", () => {
    const product = createProduct();

    expect(getProductPrice(product)).toBe(15000);
    expect(getProductQuantity(product)).toBe(5);
    expect(getProductMinQuantity(product)).toBe(2);
  });

  test("calculates subtotal tax discount and change", () => {
    const items: CartItem[] = [{ product: createProduct(), quantity: 2 }];

    expect(calculateCartTotals(items, 10, 50000)).toEqual({
      subtotal: 30000,
      totalTax: 3000,
      discountAmount: 3000,
      total: 30000,
      change: 20000,
    });
  });

  test("normalizes decimal POS totals to two decimal places", () => {
    const items: CartItem[] = [
      {
        product: createProduct({ price: "3016.67", tax_rate: "5" }),
        quantity: 2,
      },
    ];

    expect(calculateCartTotals(items, 0, 6335.01)).toEqual({
      subtotal: 6033.34,
      totalTax: 301.67,
      discountAmount: 0,
      total: 6335.01,
      change: 0,
    });

    expect(roundCurrencyAmount(6335.0070000000005)).toBe(6335.01);
    expect(formatCurrencyInputValue(6335.0070000000005)).toBe("6335.01");
  });

  test("adds and updates cart items", () => {
    const product = createProduct();

    const afterAdd = addCartItem([], product);
    expect(afterAdd).toEqual([{ product, quantity: 1 }]);

    const afterSecondAdd = addCartItem(afterAdd, product);
    expect(afterSecondAdd[0].quantity).toBe(2);

    const afterIncrement = updateCartItemQuantity(afterSecondAdd, product.id, 1);
    expect(afterIncrement[0].quantity).toBe(3);

    const afterRemovalDelta = updateCartItemQuantity(afterIncrement, product.id, -3);
    expect(afterRemovalDelta).toEqual([]);
  });

  test("removes cart item directly", () => {
    const product = createProduct();
    const items: CartItem[] = [{ product, quantity: 2 }];

    expect(removeCartItem(items, product.id)).toEqual([]);
  });

  test("builds sale request for cash payment with change", () => {
    const items: CartItem[] = [{ product: createProduct(), quantity: 2 }];

    expect(buildSaleRequest(items, "cash", 10, 30000, 50000)).toEqual({
      items: [
        {
          product_id: "product-1",
          quantity: "2",
          unit_price: "15000",
        },
      ],
      payments: [
        {
          payment_method: "cash",
          amount: "30000",
          amount_tendered: "50000",
          change_given: "20000",
        },
      ],
      discount_value: "10",
      discount_type: "percentage",
    });
  });

  test("builds sale request for non cash payment without tendered amount", () => {
    const items: CartItem[] = [{ product: createProduct(), quantity: 1 }];

    expect(buildSaleRequest(items, "card", 0, 16500, 0)).toEqual({
      items: [
        {
          product_id: "product-1",
          quantity: "1",
          unit_price: "15000",
        },
      ],
      payments: [
        {
          payment_method: "card",
          amount: "16500",
          amount_tendered: undefined,
          change_given: undefined,
        },
      ],
      discount_value: undefined,
      discount_type: undefined,
    });
  });

  test("builds cart validation request without payment fields", () => {
    const items: CartItem[] = [{ product: createProduct(), quantity: 2 }];

    expect(buildSaleCartValidationRequest(items, "shift-1")).toEqual({
      items: [
        {
          product_id: "product-1",
          quantity: "2",
          unit_price: "15000",
        },
      ],
      shift_id: "shift-1",
    });
  });

  test("classifies stock states", () => {
    expect(classifyStock(createProduct({ inventory: { quantity: "0", available_quantity: "0", low_stock_threshold: "2", is_low_stock: false } }))).toEqual({
      isOutOfStock: true,
      isLowStock: false,
    });

    expect(classifyStock(createProduct({ inventory: { quantity: "2", available_quantity: "2", low_stock_threshold: "2", is_low_stock: false } }))).toEqual({
      isOutOfStock: false,
      isLowStock: true,
    });
  });

  test("requires live permission and required input before starting a shift", () => {
    expect(
      canSubmitStartShift({
        canStartShift: false,
        startingCash: "100000",
      }),
    ).toBe(false);

    expect(
      canSubmitStartShift({
        canStartShift: true,
        startingCash: "",
      }),
    ).toBe(false);

    expect(
      canSubmitStartShift({
        canStartShift: true,
        startingCash: "100000",
      }),
    ).toBe(true);
  });

  test("requires live permission, input, and idle state before ending a shift", () => {
    expect(
      canSubmitEndShift({
        canEndShift: false,
        endingCash: "100000",
        isProcessing: false,
      }),
    ).toBe(false);

    expect(
      canSubmitEndShift({
        canEndShift: true,
        endingCash: "",
        isProcessing: false,
      }),
    ).toBe(false);

    expect(
      canSubmitEndShift({
        canEndShift: true,
        endingCash: "100000",
        isProcessing: true,
      }),
    ).toBe(false);

    expect(
      canSubmitEndShift({
        canEndShift: true,
        endingCash: "100000",
        isProcessing: false,
      }),
    ).toBe(true);
  });
});

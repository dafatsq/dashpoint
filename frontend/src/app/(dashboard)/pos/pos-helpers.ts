import { buildBackendUrl } from "@/lib/config";
import type { CartItem, CreateSaleRequest, PaymentMethod, Product } from "@/types";

export const QUICK_CASH_AMOUNTS = [10000, 20000, 50000, 100000] as const;

export function getImageUrl(path: string | null | undefined): string {
  return path ? buildBackendUrl(path) : "";
}

export function getProductPrice(product: Product): number {
  return parseFloat(product.price) || 0;
}

export function getProductQuantity(product: Product): number {
  return product.inventory?.quantity ? parseFloat(product.inventory.quantity) || 0 : 0;
}

export function getProductMinQuantity(product: Product): number {
  return product.inventory?.low_stock_threshold
    ? parseFloat(product.inventory.low_stock_threshold) || 0
    : 0;
}

export function formatCurrency(amount: number): string {
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    minimumFractionDigits: 0,
  }).format(amount);
}

export function calculateCartTotals(
  cartItems: CartItem[],
  discount: number,
  amountPaid: number,
): {
  subtotal: number;
  totalTax: number;
  discountAmount: number;
  total: number;
  change: number;
} {
  const subtotal = cartItems.reduce(
    (sum, item) => sum + getProductPrice(item.product) * item.quantity,
    0,
  );

  const totalTax = cartItems.reduce((sum, item) => {
    const itemSubtotal = getProductPrice(item.product) * item.quantity;
    const taxRate = item.product.tax_rate ? parseFloat(item.product.tax_rate) || 0 : 0;
    return sum + (itemSubtotal * taxRate) / 100;
  }, 0);

  const discountAmount = (subtotal * discount) / 100;
  const total = subtotal + totalTax - discountAmount;
  const change = amountPaid - total;

  return {
    subtotal,
    totalTax,
    discountAmount,
    total,
    change,
  };
}

export function addCartItem(items: CartItem[], product: Product): CartItem[] {
  const existingItem = items.find((item) => item.product.id === product.id);
  if (existingItem) {
    return items.map((item) =>
      item.product.id === product.id ? { ...item, quantity: item.quantity + 1 } : item,
    );
  }

  return [...items, { product, quantity: 1 }];
}

export function updateCartItemQuantity(
  items: CartItem[],
  productId: string,
  delta: number,
): CartItem[] {
  return items
    .map((item) =>
      item.product.id === productId
        ? { ...item, quantity: Math.max(0, item.quantity + delta) }
        : item,
    )
    .filter((item) => item.quantity > 0);
}

export function removeCartItem(items: CartItem[], productId: string): CartItem[] {
  return items.filter((item) => item.product.id !== productId);
}

export function classifyStock(product: Product): {
  isOutOfStock: boolean;
  isLowStock: boolean;
} {
  const quantity = getProductQuantity(product);
  const minQuantity = getProductMinQuantity(product);

  return {
    isOutOfStock: quantity <= 0,
    isLowStock: quantity <= minQuantity && quantity > 0,
  };
}

export function buildSaleRequest(
  cartItems: CartItem[],
  paymentMethod: PaymentMethod,
  discount: number,
  total: number,
  amountPaid: number,
): CreateSaleRequest {
  const change = amountPaid - total;

  return {
    items: cartItems.map((item) => ({
      product_id: item.product.id,
      quantity: item.quantity.toString(),
      unit_price: item.product.price,
    })),
    payments: [
      {
        payment_method: paymentMethod,
        amount: total.toString(),
        amount_tendered: paymentMethod === "cash" ? amountPaid.toString() : undefined,
        change_given: paymentMethod === "cash" && change > 0 ? change.toString() : undefined,
      },
    ],
    discount_value: discount > 0 ? discount.toString() : undefined,
    discount_type: discount > 0 ? "percentage" : undefined,
  };
}

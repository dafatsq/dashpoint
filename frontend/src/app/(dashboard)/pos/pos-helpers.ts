import { buildBackendUrl } from "@/lib/config";
import type { CartItem, CreateSaleRequest, PaymentMethod, Product, ValidateSaleCartRequest } from "@/types";

export const QUICK_CASH_AMOUNTS = [10000, 20000, 50000, 100000] as const;

export function parseNumericInput(value: string): number {
  return parseFloat(value || "0");
}

export function roundCurrencyAmount(value: number): number {
  if (!Number.isFinite(value)) return 0;
  return Math.round((value + Number.EPSILON) * 100) / 100;
}

export function formatCurrencyInputValue(value: number): string {
  return roundCurrencyAmount(value)
    .toFixed(2)
    .replace(/\.00$/, "")
    .replace(/(\.\d)0$/, "$1");
}

export function formatCurrencyRequestValue(value: number): string {
  return formatCurrencyInputValue(value);
}

export function getImageUrl(path: string | null | undefined): string {
  return path ? buildBackendUrl(path) : "";
}

export function getProductPrice(product: Product): number {
  return roundCurrencyAmount(parseFloat(product.price) || 0);
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

export function getCashDrawerDialogCopy(
  operationType: "pay_in" | "pay_out",
): {
  title: string;
  description: string;
  confirmLabel: string;
  confirmClassName?: string;
} {
  if (operationType === "pay_out") {
    return {
      title: "Pay Out",
      description:
        "Remove cash from the drawer (e.g., petty cash, withdrawal).",
      confirmLabel: "Confirm Pay Out",
      confirmClassName: "bg-red-600 hover:bg-red-700",
    };
  }

  return {
    title: "Pay In",
    description:
      "Add cash to the drawer (e.g., change float, deposit).",
    confirmLabel: "Confirm Pay In",
  };
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
  const subtotal = roundCurrencyAmount(
    cartItems.reduce(
      (sum, item) => sum + getProductPrice(item.product) * item.quantity,
      0,
    ),
  );

  const totalTax = roundCurrencyAmount(cartItems.reduce((sum, item) => {
    const itemSubtotal = roundCurrencyAmount(
      getProductPrice(item.product) * item.quantity,
    );
    const taxRate = item.product.tax_rate ? parseFloat(item.product.tax_rate) || 0 : 0;
    return sum + roundCurrencyAmount((itemSubtotal * taxRate) / 100);
  }, 0));

  const discountAmount = roundCurrencyAmount((subtotal * discount) / 100);
  const total = roundCurrencyAmount(subtotal + totalTax - discountAmount);
  const change = roundCurrencyAmount(amountPaid - total);

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

export function canSubmitStartShift(input: {
  canStartShift: boolean;
  startingCash: string;
}): boolean {
  return input.canStartShift && Boolean(input.startingCash);
}

export function canSubmitEndShift(input: {
  canEndShift: boolean;
  endingCash: string;
  isProcessing: boolean;
}): boolean {
  return input.canEndShift && Boolean(input.endingCash) && !input.isProcessing;
}

export function buildSaleRequest(
  cartItems: CartItem[],
  paymentMethod: PaymentMethod,
  discount: number,
  total: number,
  amountPaid: number,
  shiftId?: string,
): CreateSaleRequest {
  const normalizedTotal = roundCurrencyAmount(total);
  const normalizedAmountPaid = roundCurrencyAmount(amountPaid);
  const change = roundCurrencyAmount(normalizedAmountPaid - normalizedTotal);

  const request: CreateSaleRequest = {
    items: cartItems.map((item) => ({
      product_id: item.product.id,
      quantity: item.quantity.toString(),
      unit_price: item.product.price,
    })),
    payments: [
      {
        payment_method: paymentMethod,
        amount: formatCurrencyRequestValue(normalizedTotal),
        amount_tendered:
          paymentMethod === "cash"
            ? formatCurrencyRequestValue(normalizedAmountPaid)
            : undefined,
        change_given:
          paymentMethod === "cash" && change > 0
            ? formatCurrencyRequestValue(change)
            : undefined,
      },
    ],
    discount_value: discount > 0 ? discount.toString() : undefined,
    discount_type: discount > 0 ? "percentage" : undefined,
  };
  if (shiftId) {
    request.shift_id = shiftId;
  }
  return request;
}

export function buildSaleCartValidationRequest(
  cartItems: CartItem[],
  shiftId?: string,
): ValidateSaleCartRequest {
  const request: ValidateSaleCartRequest = {
    items: cartItems.map((item) => ({
      product_id: item.product.id,
      quantity: item.quantity.toString(),
      unit_price: item.product.price,
    })),
  };
  if (shiftId) {
    request.shift_id = shiftId;
  }
  return request;
}

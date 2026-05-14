import type { RefObject } from "react";

import type { CartItem, CashDrawerOperation, PaymentMethod, Product, Shift } from "@/types";

export interface PosErrorState {
  title: string;
  message: string;
}

export interface PosCartViewProps {
  cartItems: CartItem[];
  subtotal: number;
  totalTax: number;
  discount: number;
  discountAmount: number;
  total: number;
  currentShift: Shift | null;
  canApplyDiscount: boolean;
  canCreateSale: boolean;
  cashierName?: string;
  onClear: () => void;
  onUpdateQuantity: (id: string, delta: number) => void;
  onRemove: (id: string) => void;
  onDiscountChange: (discount: number) => void;
  onCheckout: () => Promise<void>;
}

export interface PosCheckoutState {
  paymentMethod: PaymentMethod;
  amountPaid: string;
  isProcessing: boolean;
  saleComplete: boolean;
  lastInvoice: string;
  lastChange: number;
}

export interface ShiftOperationSummary {
  operations: CashDrawerOperation[];
  totals: {
    pay_in_total: string;
    pay_out_total: string;
  };
}

export interface ProductGridProps {
  products: Product[];
  categories: { id: string; name: string }[];
  currentShift: Shift | null;
  isLoadingProducts: boolean;
  isFetchingMore: boolean;
  hasMore: boolean;
  searchQuery: string;
  selectedCategory: string;
  onSearchChange: (value: string) => void;
  onCategoryChange: (value: string) => void;
  onAddToCart: (product: Product) => void;
  loadMoreRef: RefObject<HTMLDivElement | null>;
}

import type { Product } from "./catalog";

export type SaleStatus = "completed" | "voided" | "pending";
export type PaymentMethod = "cash" | "card" | "qris" | "transfer";

export interface SaleItem {
  id: string;
  product_id: string;
  product_name: string;
  product_sku?: string;
  product_barcode?: string;
  quantity: string;
  unit_price: string;
  discount_amount: string;
  tax_rate: string;
  tax_amount: string;
  subtotal: string;
  total: string;
}

export interface Sale {
  id: string;
  invoice_no: string;
  employee_id: string;
  employee_name?: string;
  shift_id?: string;
  status: SaleStatus;
  subtotal: string;
  tax_amount: string;
  discount_amount: string;
  total_amount: string;
  amount_paid: string;
  change_amount: string;
  item_count: number;
  payment_status: string;
  notes?: string;
  void_reason?: string;
  created_at: string;
  updated_at: string;
  items?: SaleItem[];
  payments?: SalePayment[];
}

export interface SalePayment {
  id?: string;
  payment_method: PaymentMethod;
  amount: string;
  amount_tendered?: string;
  change_given?: string;
  reference_no?: string;
  notes?: string;
}

export interface CreateSaleItem {
  product_id: string;
  quantity: string;
  unit_price: string;
  discount_value?: string;
  discount_amount?: string;
}

export interface CreatePayment {
  payment_method: PaymentMethod;
  amount: string;
  amount_tendered?: string;
  change_given?: string;
  reference_no?: string;
  notes?: string;
}

export interface CreateSaleRequest {
  items: CreateSaleItem[];
  payments: CreatePayment[];
  discount_type?: string;
  discount_value?: string;
  discount_reason?: string;
  notes?: string;
}

export interface CartItem {
  product: Product;
  quantity: number;
}

export interface Cart {
  items: CartItem[];
  subtotal: number;
  discount: number;
  tax: number;
  total: number;
}

// User types
export type UserRole = 'owner' | 'manager' | 'cashier';

export interface Permission {
  id: string;
  key: string;
  name: string;
  description?: string;
  category: string;
}

export interface PermissionOverride {
  permission_id: string;
  permission_key: string;
  permission_name: string;
  allowed: boolean;
  granted_by?: string;
  granted_by_name?: string;
  created_at?: string;
}

export interface Role {
  id: string;
  name: string;
  description?: string;
}

export interface User {
  id: string;
  email?: string;
  name: string;
  role_id: string;
  role_name: UserRole;  // The string role name from backend
  role?: Role;          // Optional full role object
  is_active: boolean;
  has_pin: boolean;
  permissions?: string[];
  last_login_at?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateUserRequest {
  email: string;
  password: string;
  name: string;
  role?: UserRole;        // For UI convenience
  role_id?: string;       // What backend actually expects
  pin?: string;
  permissions?: string[];
}

export interface UpdateUserRequest {
  email?: string;
  name?: string;
  role?: UserRole;        // For UI convenience
  role_id?: string;       // What backend actually expects
  is_active?: boolean;
  password?: string;
  pin?: string;
  permissions?: string[];
}

// Category types
export interface Category {
  id: string;
  name: string;
  description: string;
  product_count: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

// Product types - Backend returns decimals as strings
export interface ProductInventory {
  quantity: string;
  available_quantity: string;
  low_stock_threshold: string;
  is_low_stock: boolean;
}

export interface Product {
  id: string;
  name: string;
  description?: string;
  sku?: string;
  barcode?: string;
  price: string;
  cost: string;
  tax_rate?: string;
  unit?: string;
  category_id?: string;
  category_name?: string;
  is_active: boolean;
  track_inventory?: boolean;
  allow_negative_stock?: boolean;
  image_url?: string;
  inventory?: ProductInventory;
  created_at: string;
  updated_at: string;
}

export interface CreateProductRequest {
  name: string;
  description?: string;
  sku?: string;
  barcode?: string;
  price: string;
  cost?: string;
  tax_rate?: string;
  initial_quantity?: string;
  low_stock_threshold?: string;
  category_id?: string;
  image_url?: string;
}

export interface UpdateProductRequest {
  name?: string;
  description?: string;
  sku?: string;
  barcode?: string;
  price?: string;
  cost?: string;
  tax_rate?: string;
  category_id?: string;
  is_active?: boolean;
  image_url?: string;
}

// Inventory types
// Backend returns ProductWithInventory which embeds Product
export interface LowStockItem {
  id: string;
  name: string;
  sku: string | null;
  barcode: string | null;
  price: string;
  category_id: string | null;
  is_active: boolean;
  quantity: string;
  available_quantity: string;
  is_low_stock: boolean;
}

export type AdjustmentType = 'purchase' | 'adjustment' | 'damage' | 'loss' | 'count';

export interface InventoryAdjustment {
  product_id: string;
  adjustment_type: AdjustmentType;
  quantity: string;
  reason?: string;
}

// Shift types
export type ShiftStatus = 'open' | 'closed' | 'suspended';

export interface Shift {
  id: string;
  employee_id: string;
  started_at: string;
  ended_at?: string;
  status: ShiftStatus;
  // Cash drawer
  opening_cash: string;
  closing_cash?: string;
  expected_cash?: string;
  cash_difference?: string;
  // Summary
  total_sales: string;
  total_cash_sales: string;
  total_refunds: string;
  transaction_count: number;
  refund_count: number;
  notes?: string;
  created_at: string;
  updated_at: string;
  // Joined fields
  employee_name?: string;
  closed_by?: string;
  closed_by_name?: string;
  operations?: CashDrawerOperation[];
}

// Cash drawer operation types
export type CashDrawerOpType = 'pay_in' | 'pay_out';

export interface CashDrawerOperation {
  id: string;
  shift_id: string;
  type: CashDrawerOpType;
  amount: string;
  reason: string;
  performed_by: string;
  created_at: string;
  performed_by_name?: string;
}

export interface CashDrawerOperationsResponse {
  operations: CashDrawerOperation[];
  pay_in_total: string;
  pay_out_total: string;
}

// Sale types
export type SaleStatus = 'completed' | 'voided' | 'pending';
export type PaymentMethod = 'cash' | 'card' | 'qris' | 'transfer';

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
  customer_name?: string;
  customer_phone?: string;
  notes?: string;
  voided_by?: string;
  void_reason?: string;
  voided_at?: string;
  items?: SaleItem[];
  payments?: SalePayment[];
  created_at: string;
  updated_at: string;
}

export interface SalePayment {
  id: string;
  payment_method: PaymentMethod;
  amount: string;
  amount_tendered?: string;
  change_given?: string;
  reference_no?: string;
  status: string;
}

export interface CreateSaleItem {
  product_id: string;
  quantity: string; // decimal as string
  unit_price: string; // decimal as string
  discount_value?: string;
  discount_amount?: string;
}

export interface CreatePayment {
  payment_method: PaymentMethod;
  amount: string; // decimal as string
  amount_tendered?: string;
  change_given?: string;
  reference_no?: string;
  notes?: string;
}

export interface CreateSaleRequest {
  items: CreateSaleItem[];
  payments: CreatePayment[];
  customer_name?: string;
  customer_phone?: string;
  discount_type?: string;
  discount_value?: string;
  discount_reason?: string;
  notes?: string;
}

// Report types - Backend returns decimal values as strings
export interface DailySummary {
  date: string;
  total_sales: string;
  total_amount: string;
  total_tax: string;
  total_discount: string;
  transaction_count: number;
  item_count: number;
  payment_breakdown: Record<string, string>;
}

export interface HourlySales {
  hour: number;
  sales: string;
  transactions: number;
}

export interface DailyReport {
  date: string;
  total_sales: string;
  total_tax: string;
  total_discount: string;
  total_amount: string;
  transaction_count: number;
  item_count: number;
  voided_count: number;
  voided_amount: string;
  payment_breakdown: Record<string, string>;
  hourly_sales?: HourlySales[];
}

export interface SalesReport {
  from: string;
  to: string;
  total_revenue: string;
  total_transactions: number;
  data: { period: string; revenue: string; transactions: number }[];
}

export interface TopSeller {
  product_id: string;
  product_name: string;
  product_sku?: string;
  category_name?: string;
  quantity_sold: string;
  total_revenue: string;
  total_profit: string;
}

export interface InventoryValuation {
  total_products: number;
  total_quantity: string;
  total_cost_value: string;
  total_retail_value: string;
  potential_profit: string;
  items?: InventoryValuationItem[];
}

export interface InventoryValuationItem {
  product_id: string;
  product_name: string;
  product_sku?: string;
  category_name?: string;
  quantity: string;
  cost_price: string;
  sell_price: string;
  cost_value: string;
  retail_value: string;
}

// Keep InventoryReport as alias for backwards compatibility
export type InventoryReport = InventoryValuation;

// Cash report types
export interface CashReport {
  date: string;
  opening_cash: string;
  cash_sales: string;
  cash_refunds: string;
  pay_in_total: string;
  pay_out_total: string;
  expected_cash: string;
  actual_cash: string;
  difference: string;
  shift_count: number;
}

// Employee sales report
export interface EmployeeSales {
  employee_id: string;
  employee_name: string;
  transaction_count: number;
  item_count: number;
  total_sales: string;
  avg_transaction: string;
}

// Category sales report
export interface CategorySales {
  category_id: string;
  category_name: string;
  items_sold: number;
  quantity_sold: string;
  total_revenue: string;
}

// Sales range report
export interface SalesRangeReport {
  start_date: string;
  end_date: string;
  summary: {
    total_sales: string;
    total_tax: string;
    total_discount: string;
    total_amount: string;
    total_transactions: number;
    total_items: number;
  };
  daily_reports: DailyReport[];
}

// Audit log types
export interface AuditLog {
  id: string;
  user_id: string;
  user_name: string;
  action: string;
  entity_type: string;
  entity_id: string;
  old_values: Record<string, unknown>;
  new_values: Record<string, unknown>;
  ip_address: string;
  user_agent: string;
  created_at: string;
}

// Cart types for POS
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

// Expense types
export interface ExpenseCategory {
  id: string;
  name: string;
  description?: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface Expense {
  id: string;
  category_id?: string;
  category_name?: string;
  product_id?: string;
  product_name?: string;
  quantity?: string;
  amount: string;
  description: string;
  expense_date: string;
  vendor?: string;
  reference_number?: string;
  notes?: string;
  created_by: string;
  created_by_name?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateExpenseRequest {
  category_id?: string;
  product_id?: string;
  quantity?: string;
  amount: string;
  description: string;
  expense_date: string;
  vendor?: string;
  reference_number?: string;
  notes?: string;
}

export interface UpdateExpenseRequest {
  category_id?: string;
  product_id?: string;
  quantity?: string;
  amount?: string;
  description?: string;
  expense_date?: string;
  vendor?: string;
  reference_number?: string;
  notes?: string;
}

export interface ExpenseSummary {
  total_amount: string;
  expense_count: number;
  by_category: CategoryExpenseSummary[];
  start_date: string;
  end_date: string;
}

export interface CategoryExpenseSummary {
  category_id?: string;
  category_name: string;
  total_amount: string;
  count: number;
}

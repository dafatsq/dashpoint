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

export interface CashReport {
  date: string;
  opening_cash: string;
  cash_sales: string;
  cash_voided_sales: string;
  pay_in_total: string;
  pay_out_total: string;
  expected_cash: string;
  actual_cash: string;
  difference: string;
  shift_count: number;
}

export interface EmployeeSales {
  employee_id: string;
  employee_name: string;
  transaction_count: number;
  item_count: number;
  total_sales: string;
  avg_transaction: string;
}

export interface CategorySales {
  category_id: string;
  category_name: string;
  items_sold: number;
  quantity_sold: string;
  total_revenue: string;
}

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

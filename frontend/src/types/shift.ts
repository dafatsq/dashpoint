export type ShiftStatus = "open" | "closed" | "suspended";

export interface Shift {
  id: string;
  employee_id: string;
  started_at: string;
  ended_at?: string;
  status: ShiftStatus;
  opening_cash: string;
  closing_cash?: string;
  expected_cash?: string;
  cash_difference?: string;
  total_sales: string;
  total_cash_sales: string;
  total_refunds: string;
  transaction_count: number;
  refund_count: number;
  notes?: string;
  created_at: string;
  updated_at: string;
  employee_name?: string;
  closed_by?: string;
  closed_by_name?: string;
  operations?: CashDrawerOperation[];
}

export type CashDrawerOpType = "pay_in" | "pay_out";

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

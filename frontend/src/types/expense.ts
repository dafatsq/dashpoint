export interface ExpenseCategory {
  id: string;
  name: string;
  system_key?: string;
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
  applies_inventory: boolean;
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
  applies_inventory?: boolean;
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
  applies_inventory?: boolean;
  amount?: string;
  description?: string;
  expense_date?: string;
  vendor?: string;
  reference_number?: string;
  notes?: string;
  expected_updated_at?: string;
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

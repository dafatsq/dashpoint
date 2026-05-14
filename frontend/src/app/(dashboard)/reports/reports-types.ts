import type {
  CashReport,
  CategorySales,
  EmployeeSales,
  ExpenseSummary,
  InventoryValuation,
  SalesRangeReport,
  TopSeller,
} from '@/types';

export type ReportType =
  | 'overview'
  | 'sales'
  | 'top-sellers'
  | 'inventory'
  | 'cash'
  | 'employees'
  | 'categories';

export type DatePresetKey =
  | 'today'
  | 'yesterday'
  | 'last7'
  | 'last30'
  | 'last90'
  | 'thisMonth'
  | 'lastMonth';

export interface ReportDateRange {
  start: string;
  end: string;
}

export interface ReportState {
  salesRangeReport: SalesRangeReport | null;
  topSellers: TopSeller[];
  inventoryReport: InventoryValuation | null;
  cashReport: CashReport | null;
  employeeSales: EmployeeSales[];
  categorySales: CategorySales[];
  expenseSummary: ExpenseSummary | null;
}

export interface ReportTabActionConfig {
  refreshLabel?: string;
  exportLabel?: string;
  showDateRange?: boolean;
}

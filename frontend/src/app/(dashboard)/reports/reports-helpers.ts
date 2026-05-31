import type { CategorySales, ExpenseSummary, SalesRangeReport } from '@/types';
import { getJakartaDateString } from '@/lib/date-local';

import type { DatePresetKey, ReportDateRange, ReportType } from './reports-types';

export const DATE_PRESETS: Record<DatePresetKey, { label: string }> = {
  today: { label: 'Today' },
  yesterday: { label: 'Yesterday' },
  last7: { label: 'Last 7 Days' },
  last30: { label: 'Last 30 Days' },
  last90: { label: 'Last 90 Days' },
  thisMonth: { label: 'This Month' },
  lastMonth: { label: 'Last Month' },
};

function toDateString(date: Date): string {
  const year = date.getFullYear();
  const month = `${date.getMonth() + 1}`.padStart(2, '0');
  const day = `${date.getDate()}`.padStart(2, '0');
  return `${year}-${month}-${day}`;
}

export function getDateRange(preset: DatePresetKey, now = new Date()): ReportDateRange {
  const today = new Date(now);
  const endDate = new Date(today);
  let startDate = new Date(today);

  switch (preset) {
    case 'today':
      break;
    case 'yesterday':
      startDate.setDate(today.getDate() - 1);
      endDate.setDate(today.getDate() - 1);
      break;
    case 'last7':
      startDate.setDate(today.getDate() - 6);
      break;
    case 'last30':
      startDate.setDate(today.getDate() - 29);
      break;
    case 'last90':
      startDate.setDate(today.getDate() - 89);
      break;
    case 'thisMonth':
      startDate = new Date(today.getFullYear(), today.getMonth(), 1);
      break;
    case 'lastMonth':
      startDate = new Date(today.getFullYear(), today.getMonth() - 1, 1);
      endDate.setDate(0);
      break;
  }

  return {
    start: toDateString(startDate),
    end: toDateString(endDate),
  };
}

export function formatCurrency(amount: number | string): string {
  const value = typeof amount === 'string' ? parseFloat(amount) : amount;
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
  }).format(value || 0);
}

export function formatNumber(value: number | string): string {
  const numberValue = typeof value === 'string' ? parseFloat(value) : value;
  return new Intl.NumberFormat('id-ID').format(numberValue || 0);
}

export interface OverviewMetrics {
  totalRevenue: number;
  totalTax: number;
  totalExpenses: number;
  netRevenue: number;
  netProfit: number;
  averageDailyRevenue: number;
}

export function calculateOverviewMetrics(
  salesRangeReport: SalesRangeReport,
  expenseSummary: ExpenseSummary | null,
): OverviewMetrics {
  const totalRevenue = parseFloat(salesRangeReport.summary.total_amount) || 0;
  const totalTax = parseFloat(salesRangeReport.summary.total_tax) || 0;
  const totalExpenses = expenseSummary ? parseFloat(expenseSummary.total_amount) || 0 : 0;
  const netRevenue = totalRevenue - totalTax;
  const netProfit = netRevenue - totalExpenses;
  const averageDailyRevenue = totalRevenue / Math.max(salesRangeReport.daily_reports.length, 1);

  return {
    totalRevenue,
    totalTax,
    totalExpenses,
    netRevenue,
    netProfit,
    averageDailyRevenue,
  };
}

export function calculateCategoryRevenuePercentages(
  categorySales: CategorySales[],
): Record<string, number> {
  const totalRevenue = categorySales.reduce(
    (sum, category) => sum + (parseFloat(category.total_revenue) || 0),
    0,
  );

  return categorySales.reduce<Record<string, number>>((acc, category) => {
    const revenue = parseFloat(category.total_revenue) || 0;
    acc[category.category_id] = totalRevenue > 0 ? (revenue / totalRevenue) * 100 : 0;
    return acc;
  }, {});
}

export function buildExportFilename(
  tab: ReportType,
  dateRange: ReportDateRange,
  today = getJakartaDateString(),
): string {
  switch (tab) {
    case 'overview':
      return `comprehensive_report_${dateRange.start}_to_${dateRange.end}.csv`;
    case 'sales':
      return `sales_${dateRange.start}_to_${dateRange.end}.csv`;
    case 'top-sellers':
      return `top_sellers_${dateRange.start}_to_${dateRange.end}.csv`;
    case 'inventory':
      return `inventory_${today}.csv`;
    default:
      return `report_${tab}_${dateRange.start}_to_${dateRange.end}.csv`;
  }
}

export function getFetchKeyForTab(tab: ReportType): ReportType {
  return tab;
}

export function triggerDownload(url: string, filename: string): void {
  const anchor = document.createElement('a');
  anchor.href = url;
  anchor.download = filename;
  anchor.click();
  URL.revokeObjectURL(url);
}

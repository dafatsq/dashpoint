import { describe, expect, test } from 'vitest';

import {
  buildExportFilename,
  calculateCategoryRevenuePercentages,
  calculateOverviewMetrics,
  formatCurrency,
  formatNumber,
  getDateRange,
  getFetchKeyForTab,
} from './reports-helpers';

describe('getDateRange', () => {
  test('returns bounded last30 range', () => {
    const range = getDateRange('last30', new Date('2026-05-14T12:00:00.000Z'));
    expect(range).toEqual({ start: '2026-04-15', end: '2026-05-14' });
  });

  test('returns this month range', () => {
    const range = getDateRange('thisMonth', new Date('2026-05-14T12:00:00.000Z'));
    expect(range).toEqual({ start: '2026-05-01', end: '2026-05-14' });
  });

  test('returns last month range', () => {
    const range = getDateRange('lastMonth', new Date('2026-05-14T12:00:00.000Z'));
    expect(range).toEqual({ start: '2026-04-01', end: '2026-04-30' });
  });
});

describe('overview metrics', () => {
  test('calculates summary metrics from sales and expenses', () => {
    const metrics = calculateOverviewMetrics(
      {
        start_date: '2026-05-01',
        end_date: '2026-05-14',
        summary: {
          total_sales: '15',
          total_tax: '1000',
          total_discount: '500',
          total_amount: '10000',
          total_transactions: 12,
          total_items: 30,
        },
        daily_reports: [
          {
            date: '2026-05-01',
            total_sales: '10',
            total_tax: '500',
            total_discount: '100',
            total_amount: '6000',
            transaction_count: 7,
            item_count: 20,
            voided_count: 0,
            voided_amount: '0',
            payment_breakdown: {},
          },
          {
            date: '2026-05-02',
            total_sales: '5',
            total_tax: '500',
            total_discount: '400',
            total_amount: '4000',
            transaction_count: 5,
            item_count: 10,
            voided_count: 0,
            voided_amount: '0',
            payment_breakdown: {},
          },
        ],
      },
      {
        total_amount: '2500',
        expense_count: 2,
        by_category: [],
        start_date: '2026-05-01',
        end_date: '2026-05-14',
      },
    );

    expect(metrics).toEqual({
      totalRevenue: 10000,
      totalTax: 1000,
      totalExpenses: 2500,
      netRevenue: 9000,
      netProfit: 6500,
      averageDailyRevenue: 5000,
    });
  });
});

describe('category percentages', () => {
  test('calculates revenue share by category id', () => {
    expect(
      calculateCategoryRevenuePercentages([
        { category_id: 'a', category_name: 'A', items_sold: 1, quantity_sold: '2', total_revenue: '40' },
        { category_id: 'b', category_name: 'B', items_sold: 1, quantity_sold: '3', total_revenue: '60' },
      ]),
    ).toEqual({ a: 40, b: 60 });
  });
});

describe('formatting helpers', () => {
  test('formats numbers and currency from string inputs', () => {
    expect(formatNumber('12000')).toBe('12.000');
    expect(formatCurrency('12000')).toContain('12.000');
  });
});

describe('report mapping helpers', () => {
  test('maps tabs to fetch keys', () => {
    expect(getFetchKeyForTab('overview')).toBe('overview');
    expect(getFetchKeyForTab('inventory')).toBe('inventory');
  });

  test('builds export filenames per tab', () => {
    expect(buildExportFilename('overview', { start: '2026-05-01', end: '2026-05-14' })).toBe(
      'comprehensive_report_2026-05-01_to_2026-05-14.csv',
    );
    expect(buildExportFilename('inventory', { start: '2026-05-01', end: '2026-05-14' }, '2026-05-14')).toBe(
      'inventory_2026-05-14.csv',
    );
  });
});

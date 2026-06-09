'use client';

import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import {
  ArrowDownRight,
  ArrowUpRight,
  Banknote,
  BarChart3,
  DollarSign,
  TrendingDown,
} from 'lucide-react';
import type { ExpenseSummary, SalesRangeReport, TopSeller } from '@/types';

import {
  calculateOverviewMetrics,
  formatCurrency,
  formatNumber,
} from './reports-helpers';
import { ReportsEmptyState, ReportsLoadingState } from './reports-feedback';
import type { ReportDateRange } from './reports-types';

interface ReportsOverviewProps {
  isLoading: boolean;
  salesRangeReport: SalesRangeReport | null;
  topSellers: TopSeller[];
  expenseSummary: ExpenseSummary | null;
  dateRange: ReportDateRange;
  onViewAllTopSellers: () => void;
}

export function ReportsOverview({
  isLoading,
  salesRangeReport,
  topSellers,
  expenseSummary,
  dateRange,
  onViewAllTopSellers,
}: ReportsOverviewProps) {
  if (isLoading) {
    return <ReportsLoadingState />;
  }

  if (!salesRangeReport) {
    return <ReportsEmptyState icon={BarChart3} message="No data for selected period" />;
  }

  const metrics = calculateOverviewMetrics(salesRangeReport, expenseSummary);

  return (
    <div className="space-y-6">
      <div className="grid gap-4 grid-cols-1 sm:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium">Net Revenue</CardTitle>
            <DollarSign className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold font-mono">{formatCurrency(metrics.netRevenue)}</div>
            <p className="text-xs text-muted-foreground mt-1">Sales after discount</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium">Total Expenses</CardTitle>
            <TrendingDown className="h-4 w-4 text-red-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-red-600 font-mono">{formatCurrency(metrics.totalExpenses)}</div>
            <p className="text-xs text-muted-foreground mt-1">
              {expenseSummary ? expenseSummary.expense_count : 0} records
            </p>
          </CardContent>
        </Card>

        <Card className={metrics.netProfit >= 0 ? 'border-2 border-green-500' : 'border-2 border-red-500'}>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium">Net Profit</CardTitle>
            {metrics.netProfit >= 0 ? (
              <ArrowUpRight className="h-5 w-5 text-green-500" />
            ) : (
              <ArrowDownRight className="h-5 w-5 text-red-500" />
            )}
          </CardHeader>
          <CardContent>
            <div className={`text-2xl font-bold font-mono ${metrics.netProfit >= 0 ? 'text-green-600' : 'text-red-600'}`}>
              {formatCurrency(metrics.netProfit)}
            </div>
            <p className="text-xs text-muted-foreground mt-1">Revenue - Expenses</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium">Tax Collected</CardTitle>
            <Banknote className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold font-mono">{formatCurrency(metrics.totalTax)}</div>
            <p className="text-xs text-muted-foreground mt-1">Not included in net revenue</p>
          </CardContent>
        </Card>
      </div>

      {salesRangeReport.daily_reports.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Daily Sales Trend</CardTitle>
            <CardDescription>
              {dateRange.start} to {dateRange.end}
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex items-end gap-1 h-32">
              {(() => {
                const maxAmount = Math.max(
                  ...salesRangeReport.daily_reports.map((day) => parseFloat(day.total_amount) || 0),
                  1,
                );
                return salesRangeReport.daily_reports.map((day) => {
                  const height = ((parseFloat(day.total_amount) || 0) / maxAmount) * 100;
                  return (
                    <div
                      key={day.date}
                      className="flex-1 min-w-[4px] max-w-[24px] bg-primary rounded-t hover:bg-primary/80 transition-colors cursor-pointer"
                      style={{ height: `${Math.max(height, 2)}%` }}
                      title={`${day.date}: ${formatCurrency(day.total_amount)} (${day.transaction_count} tx)`}
                    />
                  );
                });
              })()}
            </div>
            <div className="flex justify-between mt-2 text-xs text-muted-foreground">
              <span>{salesRangeReport.daily_reports[0]?.date}</span>
              <span>{salesRangeReport.daily_reports[salesRangeReport.daily_reports.length - 1]?.date}</span>
            </div>
          </CardContent>
        </Card>
      )}

      {topSellers.length > 0 && (
        <Card>
          <CardHeader className="flex flex-row items-center justify-between">
            <div>
              <CardTitle className="text-base">Top Sellers</CardTitle>
              <CardDescription>Best performing products in this period</CardDescription>
            </div>
            <Button variant="ghost" size="sm" onClick={onViewAllTopSellers}>
              View All
              <ArrowUpRight className="h-4 w-4 ml-1" />
            </Button>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {topSellers.slice(0, 5).map((item, index) => (
                <div key={item.product_id} className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <span className="flex h-6 w-6 items-center justify-center rounded-full bg-primary text-xs font-bold text-primary-foreground">
                      {index + 1}
                    </span>
                    <div>
                      <p className="font-medium text-sm">{item.product_name}</p>
                      <p className="text-xs text-muted-foreground">{formatNumber(item.quantity_sold)} sold</p>
                    </div>
                  </div>
                  <span className="font-bold text-sm font-mono">{formatCurrency(item.total_revenue)}</span>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}

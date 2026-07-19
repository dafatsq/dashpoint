'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { BarChart3 } from 'lucide-react';
import type { SalesRangeReport } from '@/types';

import { formatCurrency, formatNumber } from './reports-helpers';
import { ReportsEmptyState, ReportsLoadingState } from './reports-feedback';

interface ReportsSalesProps {
  isLoading: boolean;
  salesRangeReport: SalesRangeReport | null;
}

export function ReportsSales({ isLoading, salesRangeReport }: ReportsSalesProps) {
  if (isLoading) {
    return <ReportsLoadingState />;
  }

  if (!salesRangeReport) {
    return <ReportsEmptyState icon={BarChart3} message="No sales data available" />;
  }

  return (
    <div className="space-y-6">
      <div className="grid gap-4 grid-cols-1 sm:grid-cols-2 lg:grid-cols-3">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Total Revenue</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-xl font-bold truncate font-mono" title={formatCurrency(salesRangeReport.summary.total_amount)}>
              {formatCurrency(salesRangeReport.summary.total_amount)}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Transactions</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-xl font-bold truncate">
              {formatNumber(salesRangeReport.summary.total_transactions)}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Items Sold</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-xl font-bold truncate">{formatNumber(salesRangeReport.summary.total_items)}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Avg/Day</CardTitle>
          </CardHeader>
          <CardContent>
            <div
              className="text-xl font-bold truncate font-mono"
              title={formatCurrency(
                (parseFloat(salesRangeReport.summary.total_amount) || 0) /
                  Math.max(salesRangeReport.daily_reports.length, 1),
              )}
            >
              {formatCurrency(
                (parseFloat(salesRangeReport.summary.total_amount) || 0) /
                  Math.max(salesRangeReport.daily_reports.length, 1),
              )}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Tax</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-xl font-bold truncate font-mono" title={formatCurrency(salesRangeReport.summary.total_tax)}>
              {formatCurrency(salesRangeReport.summary.total_tax)}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Discounts</CardTitle>
          </CardHeader>
          <CardContent>
            <div
              className="text-xl font-bold text-orange-600 truncate font-mono"
              title={formatCurrency(salesRangeReport.summary.total_discount)}
            >
              {formatCurrency(salesRangeReport.summary.total_discount)}
            </div>
          </CardContent>
        </Card>
      </div>

      <Card className="hidden lg:block">
        <CardHeader>
          <CardTitle>Daily Breakdown</CardTitle>
          <CardDescription>
            {salesRangeReport.start_date} to {salesRangeReport.end_date}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="overflow-x-auto">
            <table className="w-full min-w-max">
              <thead>
                <tr className="border-b text-left text-sm text-muted-foreground">
                  <th className="pb-3 font-medium">Date</th>
                  <th className="pb-3 font-medium text-right">Transactions</th>
                  <th className="pb-3 font-medium text-right">Items</th>
                  <th className="pb-3 font-medium text-right">Revenue</th>
                  <th className="pb-3 font-medium text-right">Tax</th>
                  <th className="pb-3 font-medium text-right">Discounts</th>
                </tr>
              </thead>
              <tbody>
                {salesRangeReport.daily_reports.map((day) => (
                  <tr key={day.date} className="border-b last:border-0 hover:bg-muted/50">
                    <td className="py-3 font-medium">{day.date}</td>
                    <td className="py-3 text-right">{day.transaction_count}</td>
                    <td className="py-3 text-right">{day.item_count}</td>
                    <td className="py-3 text-right font-bold font-mono">{formatCurrency(day.total_amount)}</td>
                    <td className="py-3 text-right font-mono">{formatCurrency(day.total_tax)}</td>
                    <td className="py-3 text-right text-orange-600 font-mono">{formatCurrency(day.total_discount)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </CardContent>
      </Card>

      <div className="lg:hidden space-y-4">
        <h3 className="font-semibold text-lg">Daily Breakdown</h3>
        {salesRangeReport.daily_reports.map((day) => (
          <Card key={day.date}>
            <CardContent className="p-4 space-y-3">
              <div className="flex justify-between items-center border-b pb-2">
                <span className="font-bold">{day.date}</span>
                <span className="text-sm text-muted-foreground">{day.transaction_count} txns</span>
              </div>
              <div className="grid grid-cols-2 gap-2 text-sm">
                <div className="flex flex-col">
                  <span className="text-muted-foreground text-xs">Revenue</span>
                  <span className="font-bold font-mono">{formatCurrency(day.total_amount)}</span>
                </div>
                <div className="flex flex-col items-end">
                  <span className="text-muted-foreground text-xs">Items</span>
                  <span>{day.item_count}</span>
                </div>
                <div className="flex flex-col">
                  <span className="text-muted-foreground text-xs">Tax</span>
                  <span className="font-mono">{formatCurrency(day.total_tax)}</span>
                </div>
                <div className="flex flex-col items-end">
                  <span className="text-muted-foreground text-xs">Discounts</span>
                  <span className="text-orange-600 font-mono">{formatCurrency(day.total_discount)}</span>
                </div>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}

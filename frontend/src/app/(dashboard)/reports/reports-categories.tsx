'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { FolderOpen } from 'lucide-react';
import type { CategorySales } from '@/types';
import { DataSortSelect } from '@/components/shared/data-sort-select';
import { useMemo, useState } from 'react';

import {
  calculateCategoryRevenuePercentages,
  formatCurrency,
  formatNumber,
} from './reports-helpers';
import { ReportsEmptyState, ReportsLoadingState } from './reports-feedback';
import type { ReportDateRange } from './reports-types';

interface ReportsCategoriesProps {
  isLoading: boolean;
  categorySales: CategorySales[];
  dateRange: ReportDateRange;
}

export function ReportsCategories({ isLoading, categorySales, dateRange }: ReportsCategoriesProps) {
  const [sort, setSort] = useState('revenue_desc');
  const sortedCategorySales = useMemo(
    () => [...categorySales].sort((left, right) => {
      const direction = sort.endsWith('_desc') ? -1 : 1;
      const sortBy = sort.replace(/_(asc|desc)$/, '');
      if (sortBy === 'name') return left.category_name.localeCompare(right.category_name) * direction;
      if (sortBy === 'quantity') return (Number.parseFloat(left.quantity_sold) - Number.parseFloat(right.quantity_sold)) * direction;
      if (sortBy === 'items') return (left.items_sold - right.items_sold) * direction;
      return (Number.parseFloat(left.total_revenue) - Number.parseFloat(right.total_revenue)) * direction;
    }),
    [categorySales, sort],
  );

  if (isLoading) {
    return <ReportsLoadingState />;
  }

  if (categorySales.length === 0) {
    return <ReportsEmptyState icon={FolderOpen} message="No category sales data available" />;
  }

  const revenuePercentages = calculateCategoryRevenuePercentages(categorySales);

  return (
    <div className="space-y-6">
      <div className="grid gap-4 grid-cols-1 sm:grid-cols-2 lg:grid-cols-3">
        {sortedCategorySales.map((category, index) => (
          <Card key={category.category_id} className="hover:shadow-md transition-shadow">
            <CardHeader className="pb-2">
              <div className="flex items-center justify-between">
                <CardTitle className="text-base">{category.category_name}</CardTitle>
                <span
                  className={`flex h-6 w-6 items-center justify-center rounded-full text-xs font-bold ${
                    index < 3 ? 'bg-primary text-primary-foreground' : 'bg-muted'
                  }`}
                >
                  {index + 1}
                </span>
              </div>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold font-mono">{formatCurrency(category.total_revenue)}</div>
              <div className="flex items-center gap-4 mt-2 text-sm text-muted-foreground">
                <span>{formatNumber(category.quantity_sold)} units</span>
                <span>•</span>
                <span>{category.items_sold} line items</span>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      <Card className="hidden lg:block">
        <CardHeader>
          <div className="flex flex-wrap items-center justify-between gap-4">
            <div>
              <CardTitle>Category Breakdown</CardTitle>
              <CardDescription>
                {dateRange.start} to {dateRange.end}
              </CardDescription>
            </div>
            <DataSortSelect value={sort} options={[{ value: 'revenue_desc', label: 'Revenue (high-low)' }, { value: 'revenue_asc', label: 'Revenue (low-high)' }, { value: 'quantity_desc', label: 'Qty sold (high-low)' }, { value: 'items_desc', label: 'Line items (high-low)' }, { value: 'name_asc', label: 'Category (A-Z)' }]} onChange={setSort} />
          </div>
        </CardHeader>
        <CardContent>
          <div className="overflow-x-auto">
            <table className="w-full min-w-max">
              <thead>
                <tr className="border-b text-left text-sm text-muted-foreground">
                  <th className="pb-3 font-medium">#</th>
                  <th className="pb-3 font-medium">Category</th>
                  <th className="pb-3 font-medium text-right">Line Items</th>
                  <th className="pb-3 font-medium text-right">Qty Sold</th>
                  <th className="pb-3 font-medium text-right">Revenue</th>
                  <th className="pb-3 font-medium text-right">% of Total</th>
                </tr>
              </thead>
              <tbody>
                {sortedCategorySales.map((category, index) => (
                  <tr key={category.category_id} className="border-b last:border-0 hover:bg-muted/50">
                    <td className="py-3">
                      <span
                        className={`flex h-7 w-7 items-center justify-center rounded-full text-xs font-bold ${
                          index < 3 ? 'bg-primary text-primary-foreground' : 'bg-muted'
                        }`}
                      >
                        {index + 1}
                      </span>
                    </td>
                    <td className="py-3 font-medium">{category.category_name}</td>
                    <td className="py-3 text-right">{formatNumber(category.items_sold)}</td>
                    <td className="py-3 text-right">{formatNumber(category.quantity_sold)}</td>
                    <td className="py-3 text-right font-bold font-mono">{formatCurrency(category.total_revenue)}</td>
                    <td className="py-3 text-right">
                      <span className="inline-flex items-center rounded-md px-2 py-1 text-xs font-medium ring-1 ring-inset bg-transparent text-muted-foreground ring-muted-foreground/30">
                        {revenuePercentages[category.category_id].toFixed(1)}%
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </CardContent>
      </Card>

      <div className="lg:hidden space-y-4">
        <div className="flex flex-wrap items-center justify-between gap-4">
          <h3 className="font-semibold text-lg">Category Breakdown</h3>
          <DataSortSelect value={sort} options={[{ value: 'revenue_desc', label: 'Revenue (high-low)' }, { value: 'revenue_asc', label: 'Revenue (low-high)' }, { value: 'quantity_desc', label: 'Qty sold (high-low)' }, { value: 'items_desc', label: 'Line items (high-low)' }, { value: 'name_asc', label: 'Category (A-Z)' }]} onChange={setSort} />
        </div>
        {sortedCategorySales.map((category, index) => (
          <Card key={category.category_id}>
            <CardContent className="p-4 space-y-3">
              <div className="flex items-center gap-3 border-b pb-3">
                <span
                  className={`flex h-8 w-8 items-center justify-center rounded-full text-sm font-bold ${
                    index < 3 ? 'bg-primary text-primary-foreground' : 'bg-muted'
                  }`}
                >
                  {index + 1}
                </span>
                <p className="font-bold">{category.category_name}</p>
              </div>
              <div className="grid grid-cols-2 gap-3 text-sm">
                <div>
                  <span className="text-xs text-muted-foreground block">Revenue</span>
                  <span className="font-bold font-mono">{formatCurrency(category.total_revenue)}</span>
                </div>
                <div className="text-right">
                  <span className="text-xs text-muted-foreground block">Qty Sold</span>
                  <span className="font-medium">{formatNumber(category.quantity_sold)}</span>
                </div>
                <div>
                  <span className="text-xs text-muted-foreground block">Line Items</span>
                  <span className="font-medium">{formatNumber(category.items_sold)}</span>
                </div>
                <div className="text-right">
                  <span className="text-xs text-muted-foreground block">% of Total</span>
                  <span className="inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium ring-1 ring-inset bg-transparent text-muted-foreground ring-muted-foreground/30">
                    {revenuePercentages[category.category_id].toFixed(1)}%
                  </span>
                </div>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}

'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { TrendingUp } from 'lucide-react';
import type { TopSeller } from '@/types';

import { formatCurrency, formatNumber } from './reports-helpers';
import { ReportsEmptyState, ReportsLoadingState } from './reports-feedback';
import type { ReportDateRange } from './reports-types';

interface ReportsTopSellersProps {
  isLoading: boolean;
  topSellers: TopSeller[];
  dateRange: ReportDateRange;
}

export function ReportsTopSellers({ isLoading, topSellers, dateRange }: ReportsTopSellersProps) {
  if (isLoading) {
    return <ReportsLoadingState />;
  }

  if (topSellers.length === 0) {
    return <ReportsEmptyState icon={TrendingUp} message="No sales data available" />;
  }

  return (
    <div className="space-y-6">
      <Card className="hidden lg:block">
        <CardHeader>
          <CardTitle>Top Selling Products</CardTitle>
          <CardDescription>
            {dateRange.start} to {dateRange.end}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="overflow-x-auto">
            <table className="w-full min-w-max">
              <thead>
                <tr className="border-b text-left text-sm text-muted-foreground">
                  <th className="pb-3 font-medium w-12">#</th>
                  <th className="pb-3 font-medium">Product</th>
                  <th className="pb-3 font-medium">Category</th>
                  <th className="pb-3 font-medium text-right">Qty Sold</th>
                  <th className="pb-3 font-medium text-right">Revenue</th>
                </tr>
              </thead>
              <tbody>
                {topSellers.map((item, index) => (
                  <tr key={`${item.product_id}-${index}`} className="border-b last:border-0 hover:bg-muted/50">
                    <td className="py-3">
                      <span
                        className={`flex h-7 w-7 items-center justify-center rounded-full text-xs font-bold ${
                          index < 3 ? 'bg-primary text-primary-foreground' : 'bg-muted'
                        }`}
                      >
                        {index + 1}
                      </span>
                    </td>
                    <td className="py-3">
                      <div>
                        <p className="font-medium">{item.product_name}</p>
                        {item.product_sku ? (
                          <p className="text-xs text-muted-foreground">{item.product_sku}</p>
                        ) : null}
                      </div>
                    </td>
                    <td className="py-3 text-muted-foreground">{item.category_name || '-'}</td>
                    <td className="py-3 text-right font-medium">{formatNumber(item.quantity_sold)}</td>
                    <td className="py-3 text-right font-bold font-mono">{formatCurrency(item.total_revenue)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </CardContent>
      </Card>

      <div className="lg:hidden space-y-4">
        <h3 className="font-semibold text-lg">Top Selling Products</h3>
        {topSellers.map((item, index) => (
          <Card key={`${item.product_id}-mobile-${index}`}>
            <CardContent className="p-4 space-y-3">
              <div className="flex items-start gap-3 border-b pb-3">
                <span
                  className={`flex-shrink-0 flex h-8 w-8 items-center justify-center rounded-full text-sm font-bold ${
                    index < 3 ? 'bg-primary text-primary-foreground' : 'bg-muted'
                  }`}
                >
                  {index + 1}
                </span>
                <div className="flex-1 min-w-0">
                  <p className="font-bold break-words">{item.product_name}</p>
                  <div className="flex items-center gap-2 text-xs text-muted-foreground mt-0.5">
                    {item.product_sku ? <span>{item.product_sku}</span> : null}
                    {item.product_sku ? <span>•</span> : null}
                    <span>{item.category_name || 'Uncategorized'}</span>
                  </div>
                </div>
              </div>

              <div className="grid grid-cols-2 gap-3 text-sm">
                <div>
                  <span className="text-xs text-muted-foreground block">Revenue</span>
                  <span className="font-bold font-mono">{formatCurrency(item.total_revenue)}</span>
                </div>
                <div className="text-right">
                  <span className="text-xs text-muted-foreground block">Qty Sold</span>
                  <span className="font-medium">{formatNumber(item.quantity_sold)}</span>
                </div>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}

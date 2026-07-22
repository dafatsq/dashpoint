'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { DataSortSelect } from '@/components/shared/data-sort-select';
import { Package } from 'lucide-react';
import type { InventoryValuation } from '@/types';
import { useMemo, useState } from 'react';

import { formatCurrency, formatNumber } from './reports-helpers';
import { ReportsEmptyState, ReportsLoadingState } from './reports-feedback';

interface ReportsInventoryProps {
  isLoading: boolean;
  inventoryReport: InventoryValuation | null;
}

export function ReportsInventory({ isLoading, inventoryReport }: ReportsInventoryProps) {
  const [sort, setSort] = useState('retail_value_desc');
  const sortedItems = useMemo(
    () => [...(inventoryReport?.items || [])].sort((left, right) => {
      const direction = sort.endsWith('_desc') ? -1 : 1;
      const sortBy = sort.replace(/_(asc|desc)$/, '');
      if (sortBy === 'name') return left.product_name.localeCompare(right.product_name) * direction;
      if (sortBy === 'quantity') return (Number.parseFloat(left.quantity) - Number.parseFloat(right.quantity)) * direction;
      return (Number.parseFloat(left.retail_value) - Number.parseFloat(right.retail_value)) * direction;
    }),
    [inventoryReport?.items, sort],
  );

  if (isLoading) {
    return <ReportsLoadingState />;
  }

  if (!inventoryReport) {
    return <ReportsEmptyState icon={Package} message="No inventory data available" />;
  }

  return (
    <div className="space-y-6">
      <div className="grid gap-4 grid-cols-1 sm:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Total Products</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{formatNumber(inventoryReport.total_products)}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Total Units</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{formatNumber(inventoryReport.total_quantity)}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Retail Value</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold font-mono">{formatCurrency(inventoryReport.total_retail_value)}</div>
          </CardContent>
        </Card>
      </div>

      {inventoryReport.items?.length ? (
        <>
          <Card className="hidden lg:block">
            <CardHeader>
              <div className="flex flex-wrap items-center justify-between gap-4">
                <div>
                  <CardTitle>Inventory Items</CardTitle>
                  <CardDescription>Current inventory valuation</CardDescription>
                </div>
                <DataSortSelect value={sort} options={[{ value: 'retail_value_desc', label: 'Value (high-low)' }, { value: 'retail_value_asc', label: 'Value (low-high)' }, { value: 'quantity_desc', label: 'Qty (high-low)' }, { value: 'quantity_asc', label: 'Qty (low-high)' }, { value: 'name_asc', label: 'Product (A-Z)' }]} onChange={setSort} />
              </div>
            </CardHeader>
            <CardContent>
              <div className="overflow-x-auto">
                <table className="w-full min-w-max">
                  <thead>
                    <tr className="border-b text-left text-sm text-muted-foreground">
                      <th className="pb-3 font-medium">Product</th>
                      <th className="pb-3 font-medium">Category</th>
                      <th className="pb-3 font-medium text-right">Qty</th>
                      <th className="pb-3 font-medium text-right">Price</th>
                      <th className="pb-3 font-medium text-right">Value</th>
                    </tr>
                  </thead>
                  <tbody>
                    {sortedItems.map((item) => (
                      <tr key={item.product_id} className="border-b last:border-0 hover:bg-muted/50">
                        <td className="py-3">
                          <div>
                            <p className="font-medium">{item.product_name}</p>
                            {item.product_sku ? (
                              <p className="text-xs text-muted-foreground">{item.product_sku}</p>
                            ) : null}
                          </div>
                        </td>
                        <td className="py-3 text-muted-foreground">{item.category_name || '-'}</td>
                        <td className="py-3 text-right">{formatNumber(item.quantity)}</td>
                        <td className="py-3 text-right font-mono">{formatCurrency(item.sell_price)}</td>
                        <td className="py-3 text-right font-bold font-mono">{formatCurrency(item.retail_value)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </CardContent>
          </Card>

          <div className="lg:hidden space-y-4">
            <div className="flex flex-wrap items-center justify-between gap-4">
              <h3 className="font-semibold text-lg">Inventory Items</h3>
              <DataSortSelect value={sort} options={[{ value: 'retail_value_desc', label: 'Value (high-low)' }, { value: 'retail_value_asc', label: 'Value (low-high)' }, { value: 'quantity_desc', label: 'Qty (high-low)' }, { value: 'quantity_asc', label: 'Qty (low-high)' }, { value: 'name_asc', label: 'Product (A-Z)' }]} onChange={setSort} />
            </div>
            {sortedItems.map((item) => (
              <Card key={item.product_id}>
                <CardContent className="p-4 space-y-3">
                  <div className="border-b pb-2">
                    <p className="font-bold break-words">{item.product_name}</p>
                    <div className="flex items-center gap-2 text-xs text-muted-foreground mt-0.5">
                      {item.product_sku ? <span>{item.product_sku}</span> : null}
                      {item.product_sku ? <span>•</span> : null}
                      <span>{item.category_name || 'Uncategorized'}</span>
                    </div>
                  </div>
                  <div className="grid grid-cols-2 gap-3 text-sm">
                    <div>
                      <span className="text-xs text-muted-foreground block">Quantity</span>
                      <span className="font-medium">{formatNumber(item.quantity)}</span>
                    </div>
                    <div className="text-right">
                      <span className="text-xs text-muted-foreground block">Retail Value</span>
                      <span className="font-bold font-mono">{formatCurrency(item.retail_value)}</span>
                    </div>
                    <div className="col-span-2 text-right">
                      <span className="text-xs text-muted-foreground block">Price</span>
                      <span className="font-medium font-mono">{formatCurrency(item.sell_price)}</span>
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        </>
      ) : null}
    </div>
  );
}

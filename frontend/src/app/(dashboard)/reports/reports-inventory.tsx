'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Package } from 'lucide-react';
import type { InventoryValuation } from '@/types';

import { formatCurrency, formatNumber } from './reports-helpers';
import { ReportsEmptyState, ReportsLoadingState } from './reports-feedback';

interface ReportsInventoryProps {
  isLoading: boolean;
  inventoryReport: InventoryValuation | null;
}

export function ReportsInventory({ isLoading, inventoryReport }: ReportsInventoryProps) {
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
            <div className="text-2xl font-bold">{formatCurrency(inventoryReport.total_retail_value)}</div>
          </CardContent>
        </Card>
      </div>

      {inventoryReport.items?.length ? (
        <>
          <Card className="hidden lg:block">
            <CardHeader>
              <CardTitle>Inventory Items</CardTitle>
              <CardDescription>Sorted by retail value</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="overflow-x-auto">
                <table className="w-full">
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
                    {inventoryReport.items.map((item) => (
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
                        <td className="py-3 text-right">{formatCurrency(item.sell_price)}</td>
                        <td className="py-3 text-right font-bold">{formatCurrency(item.retail_value)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </CardContent>
          </Card>

          <div className="lg:hidden space-y-4">
            <h3 className="font-semibold text-lg">Inventory Items</h3>
            {inventoryReport.items.map((item) => (
              <Card key={item.product_id}>
                <CardContent className="p-4 space-y-3">
                  <div className="border-b pb-2">
                    <p className="font-bold truncate">{item.product_name}</p>
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
                      <span className="font-bold">{formatCurrency(item.retail_value)}</span>
                    </div>
                    <div className="col-span-2 text-right">
                      <span className="text-xs text-muted-foreground block">Price</span>
                      <span className="font-medium">{formatCurrency(item.sell_price)}</span>
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

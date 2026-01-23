'use client';

import { useEffect, useState } from 'react';
import { Header } from '@/components/layout/header';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import {
  Loader2,
  BarChart3,
  TrendingUp,
  Package,
  DollarSign,
  Calendar,
  ShoppingCart,
} from 'lucide-react';
import api from '@/lib/api';
import { DailyReport, TopSeller, InventoryValuation } from '@/types';

type ReportType = 'daily' | 'top-sellers' | 'inventory';

export default function ReportsPage() {
  const [activeReport, setActiveReport] = useState<ReportType>('daily');
  const [isLoading, setIsLoading] = useState(true);
  const [selectedDate, setSelectedDate] = useState(
    new Date().toISOString().split('T')[0]
  );
  const [dateFrom, setDateFrom] = useState('');
  const [dateTo, setDateTo] = useState('');

  // Report data
  const [dailyReport, setDailyReport] = useState<DailyReport | null>(null);
  const [topSellers, setTopSellers] = useState<TopSeller[]>([]);
  const [inventoryReport, setInventoryReport] = useState<InventoryValuation | null>(null);

  // Fetch daily report
  useEffect(() => {
    if (activeReport !== 'daily') return;

    const fetchDailyReport = async () => {
      setIsLoading(true);
      try {
        const result = await api.getDailyReport(selectedDate);
        if (result.data) setDailyReport(result.data);
      } catch (error) {
        console.error('Failed to fetch daily report:', error);
      } finally {
        setIsLoading(false);
      }
    };

    fetchDailyReport();
  }, [activeReport, selectedDate]);

  // Fetch top sellers
  useEffect(() => {
    if (activeReport !== 'top-sellers') return;

    const fetchTopSellers = async () => {
      setIsLoading(true);
      try {
        const params: { from?: string; to?: string; limit?: number } = { limit: 20 };
        if (dateFrom) params.from = dateFrom;
        if (dateTo) params.to = dateTo;

        const result = await api.getTopSellers(params);
        if (result.data) setTopSellers(result.data);
      } catch (error) {
        console.error('Failed to fetch top sellers:', error);
      } finally {
        setIsLoading(false);
      }
    };

    fetchTopSellers();
  }, [activeReport, dateFrom, dateTo]);

  // Fetch inventory report
  useEffect(() => {
    if (activeReport !== 'inventory') return;

    const fetchInventoryReport = async () => {
      setIsLoading(true);
      try {
        const result = await api.getInventoryReport();
        if (result.data) setInventoryReport(result.data);
      } catch (error) {
        console.error('Failed to fetch inventory report:', error);
      } finally {
        setIsLoading(false);
      }
    };

    fetchInventoryReport();
  }, [activeReport]);

  const formatCurrency = (amount: number | string) => {
    const num = typeof amount === 'string' ? parseFloat(amount) : amount;
    return new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: 'IDR',
      minimumFractionDigits: 0,
    }).format(num || 0);
  };

  return (
    <div className="flex flex-col h-full">
      <Header title="Reports" />

      <div className="flex-1 p-6 overflow-auto">
        {/* Report type selector */}
        <div className="flex gap-2 mb-6">
          <Button
            variant={activeReport === 'daily' ? 'default' : 'outline'}
            onClick={() => setActiveReport('daily')}
          >
            <BarChart3 className="h-4 w-4 mr-2" />
            Daily Report
          </Button>
          <Button
            variant={activeReport === 'top-sellers' ? 'default' : 'outline'}
            onClick={() => setActiveReport('top-sellers')}
          >
            <TrendingUp className="h-4 w-4 mr-2" />
            Top Sellers
          </Button>
          <Button
            variant={activeReport === 'inventory' ? 'default' : 'outline'}
            onClick={() => setActiveReport('inventory')}
          >
            <Package className="h-4 w-4 mr-2" />
            Inventory
          </Button>
        </div>

        {/* Daily Report */}
        {activeReport === 'daily' && (
          <>
            <div className="flex items-center gap-4 mb-6">
              <div className="flex items-center gap-2">
                <Calendar className="h-4 w-4 text-muted-foreground" />
                <Input
                  type="date"
                  value={selectedDate}
                  onChange={(e) => setSelectedDate(e.target.value)}
                  className="w-48"
                />
              </div>
            </div>

            {isLoading ? (
              <div className="flex items-center justify-center h-64">
                <Loader2 className="h-8 w-8 animate-spin text-primary" />
              </div>
            ) : dailyReport ? (
              <div className="space-y-6">
                {/* Summary cards */}
                <div className="grid gap-4 md:grid-cols-4">
                  <Card>
                    <CardHeader className="flex flex-row items-center justify-between pb-2">
                      <CardTitle className="text-sm font-medium">Total Sales</CardTitle>
                      <DollarSign className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                      <div className="text-2xl font-bold">
                        {formatCurrency(dailyReport.total_amount)}
                      </div>
                    </CardContent>
                  </Card>

                  <Card>
                    <CardHeader className="flex flex-row items-center justify-between pb-2">
                      <CardTitle className="text-sm font-medium">Transactions</CardTitle>
                      <ShoppingCart className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                      <div className="text-2xl font-bold">
                        {dailyReport.transaction_count}
                      </div>
                    </CardContent>
                  </Card>

                  <Card>
                    <CardHeader className="flex flex-row items-center justify-between pb-2">
                      <CardTitle className="text-sm font-medium">Items Sold</CardTitle>
                      <Package className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                      <div className="text-2xl font-bold">
                        {dailyReport.item_count}
                      </div>
                    </CardContent>
                  </Card>

                  <Card>
                    <CardHeader className="flex flex-row items-center justify-between pb-2">
                      <CardTitle className="text-sm font-medium">Avg Transaction</CardTitle>
                      <BarChart3 className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                      <div className="text-2xl font-bold">
                        {dailyReport.transaction_count > 0
                          ? formatCurrency(
                              parseFloat(dailyReport.total_amount) / dailyReport.transaction_count
                            )
                          : formatCurrency(0)}
                      </div>
                    </CardContent>
                  </Card>
                </div>

                {/* Tax and Discount */}
                <div className="grid gap-4 md:grid-cols-3">
                  <Card>
                    <CardHeader className="pb-2">
                      <CardTitle className="text-sm font-medium">Tax Collected</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="text-xl font-bold">
                        {formatCurrency(dailyReport.total_tax)}
                      </div>
                    </CardContent>
                  </Card>

                  <Card>
                    <CardHeader className="pb-2">
                      <CardTitle className="text-sm font-medium">Discounts Given</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="text-xl font-bold text-orange-600">
                        {formatCurrency(dailyReport.total_discount)}
                      </div>
                    </CardContent>
                  </Card>

                  <Card>
                    <CardHeader className="pb-2">
                      <CardTitle className="text-sm font-medium">Voided Sales</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="text-xl font-bold text-red-600">
                        {dailyReport.voided_count} ({formatCurrency(dailyReport.voided_amount)})
                      </div>
                    </CardContent>
                  </Card>
                </div>

                {/* Payment breakdown */}
                {dailyReport.payment_breakdown && Object.keys(dailyReport.payment_breakdown).length > 0 && (
                  <Card>
                    <CardHeader>
                      <CardTitle>Payment Methods</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="grid gap-4 md:grid-cols-4">
                        {Object.entries(dailyReport.payment_breakdown).map(
                          ([method, amount]) => (
                            <div
                              key={method}
                              className="flex items-center justify-between rounded-lg border p-4"
                            >
                              <span className="capitalize font-medium">{method}</span>
                              <span className="font-bold">{formatCurrency(amount)}</span>
                            </div>
                          )
                        )}
                      </div>
                    </CardContent>
                  </Card>
                )}

                {/* Hourly sales */}
                {dailyReport.hourly_sales && dailyReport.hourly_sales.length > 0 && (
                  <Card>
                    <CardHeader>
                      <CardTitle>Hourly Breakdown</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="grid gap-2 grid-cols-6 md:grid-cols-12">
                        {dailyReport.hourly_sales.map((hourData) => (
                          <div
                            key={hourData.hour}
                            className="text-center p-2 rounded border"
                          >
                            <div className="text-xs text-muted-foreground">
                              {hourData.hour.toString().padStart(2, '0')}:00
                            </div>
                            <div className="font-bold">{hourData.transactions}</div>
                            <div className="text-xs">{formatCurrency(hourData.sales)}</div>
                          </div>
                        ))}
                      </div>
                    </CardContent>
                  </Card>
                )}
              </div>
            ) : (
              <Card>
                <CardContent className="flex flex-col items-center justify-center py-12">
                  <BarChart3 className="h-12 w-12 text-muted-foreground mb-4" />
                  <p className="text-muted-foreground">No data for selected date</p>
                </CardContent>
              </Card>
            )}
          </>
        )}

        {/* Top Sellers Report */}
        {activeReport === 'top-sellers' && (
          <>
            <div className="flex items-center gap-4 mb-6">
              <div className="flex items-center gap-2">
                <Calendar className="h-4 w-4 text-muted-foreground" />
                <Input
                  type="date"
                  value={dateFrom}
                  onChange={(e) => setDateFrom(e.target.value)}
                  placeholder="From"
                  className="w-40"
                />
                <span className="text-muted-foreground">to</span>
                <Input
                  type="date"
                  value={dateTo}
                  onChange={(e) => setDateTo(e.target.value)}
                  placeholder="To"
                  className="w-40"
                />
              </div>
            </div>

            {isLoading ? (
              <div className="flex items-center justify-center h-64">
                <Loader2 className="h-8 w-8 animate-spin text-primary" />
              </div>
            ) : topSellers.length === 0 ? (
              <Card>
                <CardContent className="flex flex-col items-center justify-center py-12">
                  <TrendingUp className="h-12 w-12 text-muted-foreground mb-4" />
                  <p className="text-muted-foreground">No sales data available</p>
                </CardContent>
              </Card>
            ) : (
              <Card>
                <CardHeader>
                  <CardTitle>Top Selling Products</CardTitle>
                  <CardDescription>
                    {dateFrom && dateTo
                      ? `From ${dateFrom} to ${dateTo}`
                      : 'Last 30 days'}
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="overflow-x-auto">
                    <table className="w-full">
                      <thead>
                        <tr className="border-b text-left text-sm text-muted-foreground">
                          <th className="pb-3 font-medium w-12">#</th>
                          <th className="pb-3 font-medium">Product</th>
                          <th className="pb-3 font-medium">Category</th>
                          <th className="pb-3 font-medium text-right">Qty Sold</th>
                          <th className="pb-3 font-medium text-right">Revenue</th>
                          <th className="pb-3 font-medium text-right">Profit</th>
                        </tr>
                      </thead>
                      <tbody>
                        {topSellers.map((item, index) => (
                          <tr key={`${item.product_id}-${index}`} className="border-b last:border-0">
                            <td className="py-3">
                              <span className="flex h-6 w-6 items-center justify-center rounded-full bg-primary text-xs font-bold text-primary-foreground">
                                {index + 1}
                              </span>
                            </td>
                            <td className="py-3 font-medium">{item.product_name}</td>
                            <td className="py-3 text-muted-foreground">
                              {item.category_name || '-'}
                            </td>
                            <td className="py-3 text-right">{parseFloat(item.quantity_sold)}</td>
                            <td className="py-3 text-right font-bold">
                              {formatCurrency(item.total_revenue)}
                            </td>
                            <td className="py-3 text-right text-green-600">
                              {formatCurrency(item.total_profit)}
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </CardContent>
              </Card>
            )}
          </>
        )}

        {/* Inventory Report */}
        {activeReport === 'inventory' && (
          <>
            {isLoading ? (
              <div className="flex items-center justify-center h-64">
                <Loader2 className="h-8 w-8 animate-spin text-primary" />
              </div>
            ) : inventoryReport ? (
              <div className="space-y-6">
                {/* Summary cards */}
                <div className="grid gap-4 md:grid-cols-4">
                  <Card>
                    <CardHeader className="flex flex-row items-center justify-between pb-2">
                      <CardTitle className="text-sm font-medium">Total Products</CardTitle>
                      <Package className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                      <div className="text-2xl font-bold">
                        {inventoryReport.total_products}
                      </div>
                    </CardContent>
                  </Card>

                  <Card>
                    <CardHeader className="flex flex-row items-center justify-between pb-2">
                      <CardTitle className="text-sm font-medium">Total Quantity</CardTitle>
                      <Package className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                      <div className="text-2xl font-bold">
                        {parseFloat(inventoryReport.total_quantity)}
                      </div>
                    </CardContent>
                  </Card>

                  <Card>
                    <CardHeader className="flex flex-row items-center justify-between pb-2">
                      <CardTitle className="text-sm font-medium">Cost Value</CardTitle>
                      <DollarSign className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                      <div className="text-2xl font-bold">
                        {formatCurrency(inventoryReport.total_cost_value)}
                      </div>
                    </CardContent>
                  </Card>

                  <Card>
                    <CardHeader className="flex flex-row items-center justify-between pb-2">
                      <CardTitle className="text-sm font-medium">Retail Value</CardTitle>
                      <DollarSign className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                      <div className="text-2xl font-bold">
                        {formatCurrency(inventoryReport.total_retail_value)}
                      </div>
                    </CardContent>
                  </Card>
                </div>

                {/* Potential profit */}
                <Card className="border-green-500">
                  <CardHeader>
                    <CardTitle className="text-base">Potential Profit</CardTitle>
                    <CardDescription>
                      If all inventory is sold at retail price
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div className="text-3xl font-bold text-green-600">
                      {formatCurrency(inventoryReport.potential_profit)}
                    </div>
                  </CardContent>
                </Card>

                {/* Inventory items */}
                {inventoryReport.items && inventoryReport.items.length > 0 && (
                  <Card>
                    <CardHeader>
                      <CardTitle>Inventory Items</CardTitle>
                    </CardHeader>
                    <CardContent>
                      <div className="overflow-x-auto">
                        <table className="w-full">
                          <thead>
                            <tr className="border-b text-left text-sm text-muted-foreground">
                              <th className="pb-3 font-medium">Product</th>
                              <th className="pb-3 font-medium">Category</th>
                              <th className="pb-3 font-medium text-right">Qty</th>
                              <th className="pb-3 font-medium text-right">Cost</th>
                              <th className="pb-3 font-medium text-right">Price</th>
                              <th className="pb-3 font-medium text-right">Value</th>
                            </tr>
                          </thead>
                          <tbody>
                            {inventoryReport.items.map((item, index) => (
                              <tr key={`${item.product_id}-${index}`} className="border-b last:border-0">
                                <td className="py-3 font-medium">{item.product_name}</td>
                                <td className="py-3 text-muted-foreground">
                                  {item.category_name || '-'}
                                </td>
                                <td className="py-3 text-right">{parseFloat(item.quantity)}</td>
                                <td className="py-3 text-right">
                                  {formatCurrency(item.cost_price)}
                                </td>
                                <td className="py-3 text-right">
                                  {formatCurrency(item.sell_price)}
                                </td>
                                <td className="py-3 text-right font-bold">
                                  {formatCurrency(item.retail_value)}
                                </td>
                              </tr>
                            ))}
                          </tbody>
                        </table>
                      </div>
                    </CardContent>
                  </Card>
                )}
              </div>
            ) : (
              <Card>
                <CardContent className="flex flex-col items-center justify-center py-12">
                  <Package className="h-12 w-12 text-muted-foreground mb-4" />
                  <p className="text-muted-foreground">No inventory data available</p>
                </CardContent>
              </Card>
            )}
          </>
        )}
      </div>
    </div>
  );
}

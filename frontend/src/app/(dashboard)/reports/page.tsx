'use client';

import { useEffect, useState, useCallback } from 'react';
import { Header } from '@/components/layout/header';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { DateRangePicker } from '@/components/ui/date-range-picker';
import {
  Loader2,
  BarChart3,
  TrendingUp,
  TrendingDown,
  Package,
  DollarSign,
  Calendar,
  ShoppingCart,
  Users,
  Download,
  RefreshCw,
  Banknote,
  CreditCard,
  ArrowUpRight,
  ArrowDownRight,
  Minus,
  FolderOpen,
  AlertTriangle,
} from 'lucide-react';
import api from '@/lib/api';
import {
  DailyReport,
  TopSeller,
  InventoryValuation,
  CashReport,
  EmployeeSales,
  CategorySales,
  SalesRangeReport,
} from '@/types';
import { useAuth, PERMISSIONS } from '@/contexts/auth-context';

type ReportType = 'overview' | 'sales' | 'top-sellers' | 'inventory' | 'cash' | 'employees' | 'categories';

// Date range presets
const DATE_PRESETS = {
  today: { label: 'Today', days: 0 },
  yesterday: { label: 'Yesterday', days: 1 },
  last7: { label: 'Last 7 Days', days: 7 },
  last30: { label: 'Last 30 Days', days: 30 },
  last90: { label: 'Last 90 Days', days: 90 },
  thisMonth: { label: 'This Month', days: -1 }, // Special case
  lastMonth: { label: 'Last Month', days: -2 }, // Special case
};

function getDateRange(preset: string): { start: string; end: string } {
  const today = new Date();
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
      endDate.setDate(0); // Last day of previous month
      break;
    default:
      startDate.setDate(today.getDate() - 29);
  }

  return {
    start: startDate.toISOString().split('T')[0],
    end: endDate.toISOString().split('T')[0],
  };
}

export default function ReportsPage() {
  const { hasPermission, user } = useAuth();
  const canExport = hasPermission(PERMISSIONS.REPORTS_EXPORT);

  // Debug: log permission check
  console.log('[Reports] User:', user?.role_name, 'Permissions:', user?.permissions, 'canExport:', canExport);

  const [activeTab, setActiveTab] = useState<ReportType>('overview');
  const [isLoading, setIsLoading] = useState(true);
  const [datePreset, setDatePreset] = useState('last30');
  const [dateRange, setDateRange] = useState(getDateRange('last30'));


  // Report data
  const [dailyReport, setDailyReport] = useState<DailyReport | null>(null);
  const [salesRangeReport, setSalesRangeReport] = useState<SalesRangeReport | null>(null);
  const [topSellers, setTopSellers] = useState<TopSeller[]>([]);
  const [inventoryReport, setInventoryReport] = useState<InventoryValuation | null>(null);
  const [cashReport, setCashReport] = useState<CashReport | null>(null);
  const [employeeSales, setEmployeeSales] = useState<EmployeeSales[]>([]);
  const [categorySales, setCategorySales] = useState<CategorySales[]>([]);

  const formatCurrency = (amount: number | string) => {
    const num = typeof amount === 'string' ? parseFloat(amount) : amount;
    return new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: 'IDR',
      minimumFractionDigits: 0,
    }).format(num || 0);
  };

  const formatNumber = (num: number | string) => {
    const n = typeof num === 'string' ? parseFloat(num) : num;
    return new Intl.NumberFormat('id-ID').format(n || 0);
  };

  // Handle date preset change
  const handlePresetChange = (preset: string) => {
    setDatePreset(preset);
    setDateRange(getDateRange(preset));
  };

  // Fetch overview data
  const fetchOverviewData = useCallback(async () => {
    setIsLoading(true);
    try {
      const [salesRes, topRes] = await Promise.all([
        api.getSalesRangeReport({
          start_date: dateRange.start,
          end_date: dateRange.end,
        }),
        api.getTopSellers({ from: dateRange.start, to: dateRange.end, limit: 5 }),
      ]);

      if (salesRes.data) setSalesRangeReport(salesRes.data);
      if (topRes.data) setTopSellers(topRes.data);
    } catch (error) {
      console.error('Failed to fetch overview data:', error);
    } finally {
      setIsLoading(false);
    }
  }, [dateRange]);

  // Fetch sales range data
  const fetchSalesData = useCallback(async () => {
    setIsLoading(true);
    try {
      const result = await api.getSalesRangeReport({
        start_date: dateRange.start,
        end_date: dateRange.end,
      });
      if (result.data) setSalesRangeReport(result.data);
    } catch (error) {
      console.error('Failed to fetch sales data:', error);
    } finally {
      setIsLoading(false);
    }
  }, [dateRange]);

  // Fetch top sellers
  const fetchTopSellers = useCallback(async () => {
    setIsLoading(true);
    try {
      const result = await api.getTopSellers({
        from: dateRange.start,
        to: dateRange.end,
        limit: 50,
      });
      if (result.data) setTopSellers(result.data);
    } catch (error) {
      console.error('Failed to fetch top sellers:', error);
    } finally {
      setIsLoading(false);
    }
  }, [dateRange]);

  // Fetch inventory report
  const fetchInventoryReport = useCallback(async () => {
    setIsLoading(true);
    try {
      const result = await api.getInventoryReport(true);
      if (result.data) setInventoryReport(result.data);
    } catch (error) {
      console.error('Failed to fetch inventory report:', error);
    } finally {
      setIsLoading(false);
    }
  }, []);

  // Fetch cash report
  const fetchCashReport = useCallback(async () => {
    setIsLoading(true);
    try {
      const result = await api.getCashReport({
        start_date: dateRange.start,
        end_date: dateRange.end,
      });
      if (result.data) setCashReport(result.data);
    } catch (error) {
      console.error('Failed to fetch cash report:', error);
    } finally {
      setIsLoading(false);
    }
  }, [dateRange]);

  // Fetch employee sales
  const fetchEmployeeSales = useCallback(async () => {
    setIsLoading(true);
    try {
      const result = await api.getEmployeeSalesReport({
        start_date: dateRange.start,
        end_date: dateRange.end,
      });
      if (result.data) setEmployeeSales(result.data);
    } catch (error) {
      console.error('Failed to fetch employee sales:', error);
    } finally {
      setIsLoading(false);
    }
  }, [dateRange]);

  // Fetch category sales
  const fetchCategorySales = useCallback(async () => {
    setIsLoading(true);
    try {
      const result = await api.getCategorySalesReport({
        start_date: dateRange.start,
        end_date: dateRange.end,
      });
      if (result.data) setCategorySales(result.data);
    } catch (error) {
      console.error('Failed to fetch category sales:', error);
    } finally {
      setIsLoading(false);
    }
  }, [dateRange]);

  // Fetch data based on active tab
  useEffect(() => {
    switch (activeTab) {
      case 'overview':
        fetchOverviewData();
        break;
      case 'sales':
        fetchSalesData();
        break;
      case 'top-sellers':
        fetchTopSellers();
        break;
      case 'inventory':
        fetchInventoryReport();
        break;
      case 'cash':
        fetchCashReport();
        break;
      case 'employees':
        fetchEmployeeSales();
        break;
      case 'categories':
        fetchCategorySales();
        break;
    }
  }, [activeTab, fetchOverviewData, fetchSalesData, fetchTopSellers, fetchInventoryReport, fetchCashReport, fetchEmployeeSales, fetchCategorySales]);

  // Export handlers
  const handleExportComprehensive = async () => {
    try {
      const url = await api.exportComprehensiveReportCSV({
        start_date: dateRange.start,
        end_date: dateRange.end,
      });
      const a = document.createElement('a');
      a.href = url;
      a.download = `comprehensive_report_${dateRange.start}_to_${dateRange.end}.csv`;
      a.click();
      URL.revokeObjectURL(url);
    } catch (error) {
      console.error('Failed to export comprehensive report:', error);
    }
  };

  const handleExportSales = async () => {
    try {
      const url = await api.exportSalesCSV({
        start_date: dateRange.start,
        end_date: dateRange.end,
      });
      const a = document.createElement('a');
      a.href = url;
      a.download = `sales_${dateRange.start}_to_${dateRange.end}.csv`;
      a.click();
      URL.revokeObjectURL(url);
    } catch (error) {
      console.error('Failed to export sales data:', error);
    }
  };

  const handleExportInventory = async () => {
    try {
      const url = await api.exportInventoryCSV();
      const a = document.createElement('a');
      a.href = url;
      a.download = `inventory_${new Date().toISOString().split('T')[0]}.csv`;
      a.click();
      URL.revokeObjectURL(url);
    } catch (error) {
      console.error('Failed to export inventory data:', error);
    }
  };

  const handleExportTopSellers = async () => {
    try {
      const url = await api.exportTopSellersCSV({
        start_date: dateRange.start,
        end_date: dateRange.end,
        limit: 100,
      });
      const a = document.createElement('a');
      a.href = url;
      a.download = `top_sellers_${dateRange.start}_to_${dateRange.end}.csv`;
      a.click();
      URL.revokeObjectURL(url);
    } catch (error) {
      console.error('Failed to export top sellers data:', error);
    }
  };

  const renderOverview = () => (
    <div className="space-y-6">
      {/* Date range controls */}
      {/* Date range controls */}
      <Card className="mb-6">
        <CardContent className="p-6">
          <div className="flex flex-col sm:flex-row items-stretch sm:items-center gap-4">
            <Select value={datePreset} onValueChange={handlePresetChange}>
              <SelectTrigger className="w-full sm:w-40">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {Object.entries(DATE_PRESETS).map(([key, { label }]) => (
                  <SelectItem key={key} value={key}>{label}</SelectItem>
                ))}
              </SelectContent>
            </Select>
            <DateRangePicker
              value={dateRange}
              onChange={setDateRange}
              placeholder="Select date range"
              className="w-full sm:w-[280px]"
            />
            <div className="flex gap-2 w-full sm:w-auto">
              <Button variant="outline" size="sm" onClick={fetchOverviewData} disabled={isLoading} className="flex-1 sm:flex-none">
                <RefreshCw className={`h-4 w-4 mr-2 ${isLoading ? 'animate-spin' : ''}`} />
                Refresh
              </Button>
              {canExport && (
                <Button variant="outline" size="sm" onClick={handleExportComprehensive} className="flex-1 sm:flex-none">
                  <Download className="h-4 w-4 mr-2" />
                  Export
                </Button>
              )}
            </div>
          </div>
        </CardContent>
      </Card>

      {isLoading ? (
        <div className="flex items-center justify-center h-64">
          <Loader2 className="h-8 w-8 animate-spin text-primary" />
        </div>
      ) : salesRangeReport ? (
        <>
          {/* Summary cards */}
          <div className="grid gap-4 grid-cols-1 sm:grid-cols-2 lg:grid-cols-4">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between pb-2">
                <CardTitle className="text-sm font-medium">Total Sales</CardTitle>
                <DollarSign className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{formatCurrency(salesRangeReport.summary.total_amount)}</div>
                <p className="text-xs text-muted-foreground mt-1">
                  {salesRangeReport.summary.total_transactions} transactions
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between pb-2">
                <CardTitle className="text-sm font-medium">Items Sold</CardTitle>
                <Package className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{formatNumber(salesRangeReport.summary.total_items)}</div>
                <p className="text-xs text-muted-foreground mt-1">
                  Avg {salesRangeReport.summary.total_transactions > 0
                    ? (salesRangeReport.summary.total_items / salesRangeReport.summary.total_transactions).toFixed(1)
                    : 0} per sale
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between pb-2">
                <CardTitle className="text-sm font-medium">Tax Collected</CardTitle>
                <Banknote className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{formatCurrency(salesRangeReport.summary.total_tax)}</div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between pb-2">
                <CardTitle className="text-sm font-medium">Discounts</CardTitle>
                <TrendingDown className="h-4 w-4 text-orange-500" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold text-orange-600">
                  {formatCurrency(salesRangeReport.summary.total_discount)}
                </div>
              </CardContent>
            </Card>
          </div>

          {/* Daily breakdown mini chart */}
          {salesRangeReport.daily_reports && salesRangeReport.daily_reports.length > 0 && (
            <Card>
              <CardHeader>
                <CardTitle className="text-base">Daily Sales Trend</CardTitle>
                <CardDescription>{dateRange.start} to {dateRange.end}</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="flex items-end gap-1 h-32">
                  {(() => {
                    const maxAmount = Math.max(...salesRangeReport.daily_reports.map(d => parseFloat(d.total_amount) || 0), 1);
                    return salesRangeReport.daily_reports.map((day) => {
                      const height = (parseFloat(day.total_amount) / maxAmount) * 100;
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

          {/* Quick top sellers */}
          {topSellers.length > 0 && (
            <Card>
              <CardHeader className="flex flex-row items-center justify-between">
                <div>
                  <CardTitle className="text-base">Top Sellers</CardTitle>
                  <CardDescription>Best performing products in this period</CardDescription>
                </div>
                <Button variant="ghost" size="sm" onClick={() => setActiveTab('top-sellers')}>
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
                          <p className="text-xs text-muted-foreground">
                            {formatNumber(item.quantity_sold)} sold
                          </p>
                        </div>
                      </div>
                      <span className="font-bold text-sm">{formatCurrency(item.total_revenue)}</span>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          )}
        </>
      ) : (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <BarChart3 className="h-12 w-12 text-muted-foreground mb-4" />
            <p className="text-muted-foreground">No data for selected period</p>
          </CardContent>
        </Card>
      )}
    </div>
  );

  const renderSalesReport = () => (
    <div className="space-y-6">
      {/* Date range controls */}
      {/* Date range controls */}
      <Card className="mb-6">
        <CardContent className="p-6">
          <div className="flex flex-col sm:flex-row items-stretch sm:items-center gap-4">
            <Select value={datePreset} onValueChange={handlePresetChange}>
              <SelectTrigger className="w-full sm:w-40">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {Object.entries(DATE_PRESETS).map(([key, { label }]) => (
                  <SelectItem key={key} value={key}>{label}</SelectItem>
                ))}
              </SelectContent>
            </Select>
            <DateRangePicker
              value={dateRange}
              onChange={setDateRange}
              placeholder="Select date range"
              className="w-full sm:w-[280px]"
            />
            {canExport && (
              <Button variant="outline" size="sm" onClick={handleExportSales} className="w-full sm:w-auto">
                <Download className="h-4 w-4 mr-2" />
                Export CSV
              </Button>
            )}
          </div>
        </CardContent>
      </Card>

      {isLoading ? (
        <div className="flex items-center justify-center h-64">
          <Loader2 className="h-8 w-8 animate-spin text-primary" />
        </div>
      ) : salesRangeReport ? (
        <>
          {/* Summary cards */}
          <div className="grid gap-4 grid-cols-1 sm:grid-cols-2 lg:grid-cols-3">
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium">Total Revenue</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-xl font-bold truncate" title={formatCurrency(salesRangeReport.summary.total_amount)}>
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
                <div className="text-xl font-bold truncate">
                  {formatNumber(salesRangeReport.summary.total_items)}
                </div>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium">Avg/Day</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-xl font-bold truncate" title={formatCurrency(parseFloat(salesRangeReport.summary.total_amount) / Math.max(salesRangeReport.daily_reports.length, 1))}>
                  {formatCurrency(
                    parseFloat(salesRangeReport.summary.total_amount) / Math.max(salesRangeReport.daily_reports.length, 1)
                  )}
                </div>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium">Tax</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-xl font-bold truncate" title={formatCurrency(salesRangeReport.summary.total_tax)}>
                  {formatCurrency(salesRangeReport.summary.total_tax)}
                </div>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium">Discounts</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-xl font-bold text-orange-600 truncate" title={formatCurrency(salesRangeReport.summary.total_discount)}>
                  {formatCurrency(salesRangeReport.summary.total_discount)}
                </div>
              </CardContent>
            </Card>
          </div>

          {/* Daily breakdown table */}
          {/* Daily breakdown table - Desktop */}
          <Card className="hidden lg:block">
            <CardHeader>
              <CardTitle>Daily Breakdown</CardTitle>
              <CardDescription>
                {salesRangeReport.start_date} to {salesRangeReport.end_date}
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="overflow-x-auto">
                <table className="w-full">
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
                        <td className="py-3 text-right font-bold">{formatCurrency(day.total_amount)}</td>
                        <td className="py-3 text-right">{formatCurrency(day.total_tax)}</td>
                        <td className="py-3 text-right text-orange-600">{formatCurrency(day.total_discount)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </CardContent>
          </Card>

          {/* Daily breakdown - Mobile Card View */}
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
                      <span className="font-bold">{formatCurrency(day.total_amount)}</span>
                    </div>
                    <div className="flex flex-col items-end">
                      <span className="text-muted-foreground text-xs">Items</span>
                      <span>{day.item_count}</span>
                    </div>
                    <div className="flex flex-col">
                      <span className="text-muted-foreground text-xs">Tax</span>
                      <span>{formatCurrency(day.total_tax)}</span>
                    </div>
                    <div className="flex flex-col items-end">
                      <span className="text-muted-foreground text-xs">Discounts</span>
                      <span className="text-orange-600">{formatCurrency(day.total_discount)}</span>
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        </>
      ) : (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <BarChart3 className="h-12 w-12 text-muted-foreground mb-4" />
            <p className="text-muted-foreground">No sales data available</p>
          </CardContent>
        </Card>
      )}
    </div>
  );

  const renderTopSellers = () => (
    <div className="space-y-6">
      {/* Controls */}
      {/* Controls */}
      <Card className="mb-6">
        <CardContent className="p-6">
          <div className="flex flex-col sm:flex-row items-stretch sm:items-center gap-4">
            <Select value={datePreset} onValueChange={handlePresetChange}>
              <SelectTrigger className="w-full sm:w-40">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {Object.entries(DATE_PRESETS).map(([key, { label }]) => (
                  <SelectItem key={key} value={key}>{label}</SelectItem>
                ))}
              </SelectContent>
            </Select>
            <DateRangePicker
              value={dateRange}
              onChange={setDateRange}
              placeholder="Select date range"
              className="w-full sm:w-[280px]"
            />
            {canExport && (
              <Button variant="outline" size="sm" onClick={handleExportTopSellers} className="w-full sm:w-auto">
                <Download className="h-4 w-4 mr-2" />
                Export CSV
              </Button>
            )}
          </div>
        </CardContent>
      </Card>

      {isLoading ? (
        <div className="flex items-center justify-center h-64">
          <Loader2 className="h-8 w-8 animate-spin text-primary" />
        </div>
      ) : topSellers.length > 0 ? (
        <>
          {/* Desktop Table View */}
          <Card className="hidden lg:block">
            <CardHeader>
              <CardTitle>Top Selling Products</CardTitle>
              <CardDescription>
                {dateRange.start} to {dateRange.end}
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
                      <th className="pb-3 font-medium text-right">Margin</th>
                    </tr>
                  </thead>
                  <tbody>
                    {topSellers.map((item, index) => {
                      const margin = parseFloat(item.total_revenue) > 0
                        ? (parseFloat(item.total_profit) / parseFloat(item.total_revenue)) * 100
                        : 0;
                      return (
                        <tr key={`${item.product_id}-${index}`} className="border-b last:border-0 hover:bg-muted/50">
                          <td className="py-3">
                            <span className={`flex h-7 w-7 items-center justify-center rounded-full text-xs font-bold ${index < 3 ? 'bg-primary text-primary-foreground' : 'bg-muted'
                              }`}>
                              {index + 1}
                            </span>
                          </td>
                          <td className="py-3">
                            <div>
                              <p className="font-medium">{item.product_name}</p>
                              {item.product_sku && (
                                <p className="text-xs text-muted-foreground">{item.product_sku}</p>
                              )}
                            </div>
                          </td>
                          <td className="py-3 text-muted-foreground">{item.category_name || '-'}</td>
                          <td className="py-3 text-right font-medium">{formatNumber(item.quantity_sold)}</td>
                          <td className="py-3 text-right font-bold">{formatCurrency(item.total_revenue)}</td>
                          <td className="py-3 text-right text-green-600">{formatCurrency(item.total_profit)}</td>
                          <td className="py-3 text-right">
                            <span className={`inline-flex items-center rounded-md px-2 py-1 text-xs font-medium ring-1 ring-inset ${margin > 30
                              ? 'bg-primary text-primary-foreground ring-primary/20'
                              : margin > 15
                                ? 'bg-muted text-muted-foreground ring-muted-foreground/20'
                                : 'bg-transparent text-muted-foreground ring-muted-foreground/30'
                              }`}>
                              {margin.toFixed(1)}%
                            </span>
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>
            </CardContent>
          </Card>

          {/* Mobile Card View */}
          <div className="lg:hidden space-y-4">
            <h3 className="font-semibold text-lg">Top Selling Products</h3>
            {topSellers.map((item, index) => {
              const margin = parseFloat(item.total_revenue) > 0
                ? (parseFloat(item.total_profit) / parseFloat(item.total_revenue)) * 100
                : 0;

              return (
                <Card key={`${item.product_id}-mobile-${index}`}>
                  <CardContent className="p-4 space-y-3">
                    <div className="flex items-start gap-3 border-b pb-3">
                      <span className={`flex-shrink-0 flex h-8 w-8 items-center justify-center rounded-full text-sm font-bold ${index < 3 ? 'bg-primary text-primary-foreground' : 'bg-muted'
                        }`}>
                        {index + 1}
                      </span>
                      <div className="flex-1 min-w-0">
                        <p className="font-bold truncate">{item.product_name}</p>
                        <div className="flex items-center gap-2 text-xs text-muted-foreground mt-0.5">
                          {item.product_sku && <span>{item.product_sku}</span>}
                          {item.product_sku && <span>•</span>}
                          <span>{item.category_name || 'Uncategorized'}</span>
                        </div>
                      </div>
                    </div>

                    <div className="grid grid-cols-2 gap-3 text-sm">
                      <div>
                        <span className="text-xs text-muted-foreground block">Revenue</span>
                        <span className="font-bold">{formatCurrency(item.total_revenue)}</span>
                      </div>
                      <div className="text-right">
                        <span className="text-xs text-muted-foreground block">Qty Sold</span>
                        <span className="font-medium">{formatNumber(item.quantity_sold)}</span>
                      </div>
                      <div>
                        <span className="text-xs text-muted-foreground block">Profit</span>
                        <span className="font-medium text-green-600">{formatCurrency(item.total_profit)}</span>
                      </div>
                      <div className="text-right">
                        <span className="text-xs text-muted-foreground block">Margin</span>
                        <span className={`inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium ring-1 ring-inset ${margin > 30
                          ? 'bg-primary/10 text-primary ring-primary/20'
                          : margin > 15
                            ? 'bg-muted text-muted-foreground ring-muted-foreground/20'
                            : 'bg-transparent text-muted-foreground ring-muted-foreground/30'
                          }`}>
                          {margin.toFixed(1)}%
                        </span>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              )
            })}
          </div>
        </>
      ) : (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <TrendingUp className="h-12 w-12 text-muted-foreground mb-4" />
            <p className="text-muted-foreground">No sales data available</p>
          </CardContent>
        </Card>
      )}
    </div>
  );

  const renderInventory = () => (
    <div className="space-y-6">
      {/* Controls */}
      {/* Controls */}
      <Card className="mb-6">
        <CardContent className="p-6">
          <div className="flex flex-col sm:flex-row items-stretch sm:items-center gap-4">
            <Button variant="outline" size="sm" onClick={fetchInventoryReport} disabled={isLoading} className="w-full sm:w-auto">
              <RefreshCw className={`h-4 w-4 mr-2 ${isLoading ? 'animate-spin' : ''}`} />
              Refresh
            </Button>
            {canExport && (
              <Button variant="outline" size="sm" onClick={handleExportInventory} className="w-full sm:w-auto">
                <Download className="h-4 w-4 mr-2" />
                Export CSV
              </Button>
            )}
          </div>
        </CardContent>
      </Card>

      {isLoading ? (
        <div className="flex items-center justify-center h-64">
          <Loader2 className="h-8 w-8 animate-spin text-primary" />
        </div>
      ) : inventoryReport ? (
        <>
          {/* Summary cards */}
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
                <CardTitle className="text-sm font-medium">Cost Value</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{formatCurrency(inventoryReport.total_cost_value)}</div>
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

          {/* Potential profit */}
          <Card className="border-l-4 border-l-green-500 shadow-sm">
            <CardHeader>
              <CardTitle className="text-base">Potential Profit</CardTitle>
              <CardDescription>If all inventory is sold at retail price</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="text-3xl font-bold text-green-600">
                {formatCurrency(inventoryReport.potential_profit)}
              </div>
              <p className="text-xs text-muted-foreground mt-2">
                Margin: {parseFloat(inventoryReport.total_retail_value) > 0 ? ((parseFloat(inventoryReport.potential_profit) / parseFloat(inventoryReport.total_retail_value)) * 100).toFixed(1) : 0}%
              </p>
            </CardContent>
          </Card>

          {/* Inventory items table - Desktop */}
          {inventoryReport.items && inventoryReport.items.length > 0 && (
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
                        <th className="pb-3 font-medium text-right">Cost</th>
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
                              {item.product_sku && (
                                <p className="text-xs text-muted-foreground">{item.product_sku}</p>
                              )}
                            </div>
                          </td>
                          <td className="py-3 text-muted-foreground">{item.category_name || '-'}</td>
                          <td className="py-3 text-right">{formatNumber(item.quantity)}</td>
                          <td className="py-3 text-right">{formatCurrency(item.cost_price)}</td>
                          <td className="py-3 text-right">{formatCurrency(item.sell_price)}</td>
                          <td className="py-3 text-right font-bold">{formatCurrency(item.retail_value)}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </CardContent>
            </Card>
          )}

          {/* Inventory Items - Mobile Card View */}
          {inventoryReport.items && inventoryReport.items.length > 0 && (
            <div className="lg:hidden space-y-4">
              <h3 className="font-semibold text-lg">Inventory Items</h3>
              {inventoryReport.items.map((item) => (
                <Card key={item.product_id}>
                  <CardContent className="p-4 space-y-3">
                    <div className="border-b pb-2">
                      <p className="font-bold truncate">{item.product_name}</p>
                      <div className="flex items-center gap-2 text-xs text-muted-foreground mt-0.5">
                        {item.product_sku && <span>{item.product_sku}</span>}
                        {item.product_sku && <span>•</span>}
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
                      <div>
                        <span className="text-xs text-muted-foreground block">Cost</span>
                        <span className="font-medium">{formatCurrency(item.cost_price)}</span>
                      </div>
                      <div className="text-right">
                        <span className="text-xs text-muted-foreground block">Price</span>
                        <span className="font-medium">{formatCurrency(item.sell_price)}</span>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          )}
        </>
      ) : (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <Package className="h-12 w-12 text-muted-foreground mb-4" />
            <p className="text-muted-foreground">No inventory data available</p>
          </CardContent>
        </Card>
      )}
    </div>
  );

  const renderCashReport = () => (
    <div className="space-y-6">
      {/* Date selector */}
      {/* Date selector */}
      {/* Date selector */}
      <Card className="mb-6">
        <CardContent className="p-6">
          <div className="flex flex-col sm:flex-row items-stretch sm:items-center gap-4">
            <Select value={datePreset} onValueChange={handlePresetChange}>
              <SelectTrigger className="w-full sm:w-40">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {Object.entries(DATE_PRESETS).map(([key, { label }]) => (
                  <SelectItem key={key} value={key}>{label}</SelectItem>
                ))}
              </SelectContent>
            </Select>
            <DateRangePicker
              value={dateRange}
              onChange={setDateRange}
              placeholder="Select date range"
              className="w-full sm:w-[280px]"
            />
            <Button variant="outline" size="sm" onClick={fetchCashReport} disabled={isLoading} className="w-full sm:w-auto">
              <RefreshCw className={`h-4 w-4 mr-2 ${isLoading ? 'animate-spin' : ''}`} />
              Refresh
            </Button>
          </div>
        </CardContent>
      </Card>

      {isLoading ? (
        <div className="flex items-center justify-center h-64">
          <Loader2 className="h-8 w-8 animate-spin text-primary" />
        </div>
      ) : cashReport ? (
        <>
          {/* Cash flow summary */}
          <div className="grid gap-4 grid-cols-1 sm:grid-cols-3">
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium">Opening Cash</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{formatCurrency(cashReport.opening_cash)}</div>
                <p className="text-xs text-muted-foreground mt-1">
                  From {cashReport.shift_count} shift(s)
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium">Cash Sales</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold text-green-600">
                  +{formatCurrency(cashReport.cash_sales)}
                </div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium">Cash Refunds</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold text-red-600">
                  -{formatCurrency(cashReport.cash_refunds)}
                </div>
              </CardContent>
            </Card>
          </div>

          {/* Expected vs Actual */}
          <div className="grid gap-4 grid-cols-1 sm:grid-cols-2">
            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium">Expected Cash</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{formatCurrency(cashReport.expected_cash)}</div>
                <p className="text-xs text-muted-foreground mt-1">
                  Opening + Sales - Refunds
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm font-medium">Actual Cash</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{formatCurrency(cashReport.actual_cash)}</div>
                <p className="text-xs text-muted-foreground mt-1">
                  Closing cash from shifts
                </p>
              </CardContent>
            </Card>
          </div>

          {/* Difference */}
          <Card className={`border-l-4 shadow-sm ${parseFloat(cashReport.difference) === 0 ? 'border-l-green-500' :
            parseFloat(cashReport.difference) > 0 ? 'border-l-yellow-500' : 'border-l-red-500'
            }`}>
            <CardHeader>
              <CardTitle className="text-base">Cash Difference</CardTitle>
              <CardDescription>Actual minus Expected</CardDescription>
            </CardHeader>
            <CardContent>
              <div className={`text-3xl font-bold flex items-center gap-2 ${parseFloat(cashReport.difference) > 0
                ? 'text-green-600'
                : parseFloat(cashReport.difference) < 0
                  ? 'text-red-600'
                  : 'text-green-600'
                }`}>
                {parseFloat(cashReport.difference) > 0 && <ArrowUpRight className="h-8 w-8" />}
                {parseFloat(cashReport.difference) < 0 && <ArrowDownRight className="h-8 w-8" />}
                {parseFloat(cashReport.difference) === 0 && <Minus className="h-8 w-8" />}
                {formatCurrency(Math.abs(parseFloat(cashReport.difference)))}
                {parseFloat(cashReport.difference) > 0 && ' over'}
                {parseFloat(cashReport.difference) < 0 && ' short'}
                {parseFloat(cashReport.difference) === 0 && ' (balanced)'}
              </div>
              <p className="text-xs text-muted-foreground mt-2">
                Expected: {formatCurrency(cashReport.expected_cash)} • Actual: {formatCurrency(cashReport.actual_cash)}
              </p>
            </CardContent>
          </Card>
        </>
      ) : (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <Banknote className="h-12 w-12 text-muted-foreground mb-4" />
            <p className="text-muted-foreground">No cash data available</p>
            <p className="text-xs text-muted-foreground mt-1">Shifts must be closed to appear here</p>
          </CardContent>
        </Card>
      )
      }
    </div >
  );

  const renderEmployeeReport = () => (
    <div className="space-y-6">
      {/* Controls */}
      <Card className="mb-6">
        <CardContent className="p-6">
          <div className="flex flex-col sm:flex-row items-stretch sm:items-center gap-4">
            <Select value={datePreset} onValueChange={handlePresetChange}>
              <SelectTrigger className="w-full sm:w-40">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {Object.entries(DATE_PRESETS).map(([key, { label }]) => (
                  <SelectItem key={key} value={key}>{label}</SelectItem>
                ))}
              </SelectContent>
            </Select>
            <DateRangePicker
              value={dateRange}
              onChange={setDateRange}
              placeholder="Select date range"
              className="w-full sm:w-[280px]"
            />
          </div>
        </CardContent>
      </Card>

      {isLoading ? (
        <div className="flex items-center justify-center h-64">
          <Loader2 className="h-8 w-8 animate-spin text-primary" />
        </div>
      ) : employeeSales.length > 0 ? (
        <>
          {/* Desktop Table View */}
          <Card className="hidden lg:block">
            <CardHeader>
              <CardTitle>Sales by Employee</CardTitle>
              <CardDescription>
                {dateRange.start} to {dateRange.end}
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead>
                    <tr className="border-b text-left text-sm text-muted-foreground">
                      <th className="pb-3 font-medium">#</th>
                      <th className="pb-3 font-medium">Employee</th>
                      <th className="pb-3 font-medium text-right">Transactions</th>
                      <th className="pb-3 font-medium text-right">Items Sold</th>
                      <th className="pb-3 font-medium text-right">Total Sales</th>
                      <th className="pb-3 font-medium text-right">Avg / Transaction</th>
                    </tr>
                  </thead>
                  <tbody>
                    {employeeSales.map((emp, index) => (
                      <tr key={emp.employee_id} className="border-b last:border-0 hover:bg-muted/50">
                        <td className="py-3">
                          <span className={`flex h-7 w-7 items-center justify-center rounded-full text-xs font-bold ${index < 3 ? 'bg-primary text-primary-foreground' : 'bg-muted'
                            }`}>
                            {index + 1}
                          </span>
                        </td>
                        <td className="py-3 font-medium">{emp.employee_name}</td>
                        <td className="py-3 text-right">{formatNumber(emp.transaction_count)}</td>
                        <td className="py-3 text-right">{formatNumber(emp.item_count)}</td>
                        <td className="py-3 text-right font-bold">{formatCurrency(emp.total_sales)}</td>
                        <td className="py-3 text-right">{formatCurrency(emp.avg_transaction)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </CardContent>
          </Card>

          {/* Mobile Card View */}
          <div className="lg:hidden space-y-4">
            <h3 className="font-semibold text-lg">Sales by Employee</h3>
            {employeeSales.map((emp, index) => (
              <Card key={emp.employee_id}>
                <CardContent className="p-4 space-y-3">
                  <div className="flex items-center gap-3 border-b pb-3">
                    <span className={`flex h-8 w-8 items-center justify-center rounded-full text-sm font-bold ${index < 3 ? 'bg-primary text-primary-foreground' : 'bg-muted'
                      }`}>
                      {index + 1}
                    </span>
                    <p className="font-bold">{emp.employee_name}</p>
                  </div>

                  <div className="grid grid-cols-2 gap-3 text-sm">
                    <div>
                      <span className="text-xs text-muted-foreground block">Total Sales</span>
                      <span className="font-bold">{formatCurrency(emp.total_sales)}</span>
                    </div>
                    <div className="text-right">
                      <span className="text-xs text-muted-foreground block">Transactions</span>
                      <span className="font-medium">{formatNumber(emp.transaction_count)}</span>
                    </div>
                    <div>
                      <span className="text-xs text-muted-foreground block">Items Sold</span>
                      <span className="font-medium">{formatNumber(emp.item_count)}</span>
                    </div>
                    <div className="text-right">
                      <span className="text-xs text-muted-foreground block">Avg / Txn</span>
                      <span className="font-medium">{formatCurrency(emp.avg_transaction)}</span>
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        </>
      ) : (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <Users className="h-12 w-12 text-muted-foreground mb-4" />
            <p className="text-muted-foreground">No employee sales data available</p>
          </CardContent>
        </Card>
      )}
    </div>
  );

  const renderCategoryReport = () => (
    <div className="space-y-6">
      {/* Controls */}
      <Card className="mb-6">
        <CardContent className="p-6">
          <div className="flex flex-col sm:flex-row items-stretch sm:items-center gap-4">
            <Select value={datePreset} onValueChange={handlePresetChange}>
              <SelectTrigger className="w-full sm:w-40">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {Object.entries(DATE_PRESETS).map(([key, { label }]) => (
                  <SelectItem key={key} value={key}>{label}</SelectItem>
                ))}
              </SelectContent>
            </Select>
            <DateRangePicker
              value={dateRange}
              onChange={setDateRange}
              placeholder="Select date range"
              className="w-full sm:w-[280px]"
            />
          </div>
        </CardContent>
      </Card>

      {isLoading ? (
        <div className="flex items-center justify-center h-64">
          <Loader2 className="h-8 w-8 animate-spin text-primary" />
        </div>
      ) : categorySales.length > 0 ? (
        <>
          {/* Category cards */}
          <div className="grid gap-4 grid-cols-1 sm:grid-cols-2 lg:grid-cols-3">
            {categorySales.map((cat, index) => (
              <Card key={cat.category_id} className="hover:shadow-md transition-shadow">
                <CardHeader className="pb-2">
                  <div className="flex items-center justify-between">
                    <CardTitle className="text-base">{cat.category_name}</CardTitle>
                    <span className={`flex h-6 w-6 items-center justify-center rounded-full text-xs font-bold ${index < 3 ? 'bg-primary text-primary-foreground' : 'bg-muted'
                      }`}>
                      {index + 1}
                    </span>
                  </div>
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">{formatCurrency(cat.total_revenue)}</div>
                  <div className="flex items-center gap-4 mt-2 text-sm text-muted-foreground">
                    <span>{formatNumber(cat.quantity_sold)} units</span>
                    <span>•</span>
                    <span>{cat.items_sold} line items</span>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>

          {/* Table view - Desktop */}
          <Card className="hidden lg:block">
            <CardHeader>
              <CardTitle>Category Breakdown</CardTitle>
              <CardDescription>
                {dateRange.start} to {dateRange.end}
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="overflow-x-auto">
                <table className="w-full">
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
                    {(() => {
                      const totalRevenue = categorySales.reduce((sum, c) => sum + parseFloat(c.total_revenue), 0);
                      return categorySales.map((cat, index) => {
                        const percentage = totalRevenue > 0
                          ? (parseFloat(cat.total_revenue) / totalRevenue) * 100
                          : 0;
                        return (
                          <tr key={cat.category_id} className="border-b last:border-0 hover:bg-muted/50">
                            <td className="py-3">
                              <span className={`flex h-7 w-7 items-center justify-center rounded-full text-xs font-bold ${index < 3 ? 'bg-primary text-primary-foreground' : 'bg-muted'
                                }`}>
                                {index + 1}
                              </span>
                            </td>
                            <td className="py-3 font-medium">{cat.category_name}</td>
                            <td className="py-3 text-right">{formatNumber(cat.items_sold)}</td>
                            <td className="py-3 text-right">{formatNumber(cat.quantity_sold)}</td>
                            <td className="py-3 text-right font-bold">{formatCurrency(cat.total_revenue)}</td>
                            <td className="py-3 text-right">
                              <span className="inline-flex items-center rounded-md px-2 py-1 text-xs font-medium ring-1 ring-inset bg-transparent text-muted-foreground ring-muted-foreground/30">
                                {percentage.toFixed(1)}%
                              </span>
                            </td>
                          </tr>
                        );
                      });
                    })()}
                  </tbody>
                </table>
              </div>
            </CardContent>
          </Card>

          {/* Table view - Mobile Card View */}
          <div className="lg:hidden space-y-4">
            <h3 className="font-semibold text-lg">Category Breakdown</h3>
            {(() => {
              const totalRevenue = categorySales.reduce((sum, c) => sum + parseFloat(c.total_revenue), 0);
              return categorySales.map((cat, index) => {
                const percentage = totalRevenue > 0
                  ? (parseFloat(cat.total_revenue) / totalRevenue) * 100
                  : 0;
                return (
                  <Card key={cat.category_id}>
                    <CardContent className="p-4 space-y-3">
                      <div className="flex items-center gap-3 border-b pb-3">
                        <span className={`flex h-8 w-8 items-center justify-center rounded-full text-sm font-bold ${index < 3 ? 'bg-primary text-primary-foreground' : 'bg-muted'
                          }`}>
                          {index + 1}
                        </span>
                        <p className="font-bold">{cat.category_name}</p>
                      </div>

                      <div className="grid grid-cols-2 gap-3 text-sm">
                        <div>
                          <span className="text-xs text-muted-foreground block">Revenue</span>
                          <span className="font-bold">{formatCurrency(cat.total_revenue)}</span>
                        </div>
                        <div className="text-right">
                          <span className="text-xs text-muted-foreground block">Qty Sold</span>
                          <span className="font-medium">{formatNumber(cat.quantity_sold)}</span>
                        </div>
                        <div>
                          <span className="text-xs text-muted-foreground block">Line Items</span>
                          <span className="font-medium">{formatNumber(cat.items_sold)}</span>
                        </div>
                        <div className="text-right">
                          <span className="text-xs text-muted-foreground block">% of Total</span>
                          <span className="inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium ring-1 ring-inset bg-transparent text-muted-foreground ring-muted-foreground/30">
                            {percentage.toFixed(1)}%
                          </span>
                        </div>
                      </div>
                    </CardContent>
                  </Card>
                );
              });
            })()}
          </div>
        </>
      ) : (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <FolderOpen className="h-12 w-12 text-muted-foreground mb-4" />
            <p className="text-muted-foreground">No category sales data available</p>
          </CardContent>
        </Card>
      )}
    </div>
  );

  return (
    <div className="flex flex-col h-full">
      <Header title="Reports" />

      <div className="flex-1 p-6 overflow-auto">
        <Tabs value={activeTab} onValueChange={(v) => setActiveTab(v as ReportType)}>
          <TabsList className="w-full justify-start flex overflow-x-auto lg:w-auto lg:inline-flex mb-6 no-scrollbar pb-1">
            <TabsTrigger value="overview" className="gap-2">
              <BarChart3 className="h-4 w-4" />
              <span className="ml-2">Overview</span>
            </TabsTrigger>
            <TabsTrigger value="sales">
              <ShoppingCart className="h-4 w-4" />
              <span className="ml-2">Sales</span>
            </TabsTrigger>
            <TabsTrigger value="top-sellers">
              <TrendingUp className="h-4 w-4" />
              <span className="ml-2">Top Sellers</span>
            </TabsTrigger>
            <TabsTrigger value="inventory">
              <Package className="h-4 w-4" />
              <span className="ml-2">Inventory</span>
            </TabsTrigger>
            <TabsTrigger value="cash">
              <Banknote className="h-4 w-4" />
              <span className="ml-2">Cash</span>
            </TabsTrigger>
            <TabsTrigger value="employees">
              <Users className="h-4 w-4" />
              <span className="ml-2">Employees</span>
            </TabsTrigger>
            <TabsTrigger value="categories">
              <FolderOpen className="h-4 w-4" />
              <span className="ml-2">Categories</span>
            </TabsTrigger>
          </TabsList>

          <TabsContent value="overview">{renderOverview()}</TabsContent>
          <TabsContent value="sales">{renderSalesReport()}</TabsContent>
          <TabsContent value="top-sellers">{renderTopSellers()}</TabsContent>
          <TabsContent value="inventory">{renderInventory()}</TabsContent>
          <TabsContent value="cash">{renderCashReport()}</TabsContent>
          <TabsContent value="employees">{renderEmployeeReport()}</TabsContent>
          <TabsContent value="categories">{renderCategoryReport()}</TabsContent>
        </Tabs>
      </div>
    </div>
  );
}

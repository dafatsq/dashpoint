'use client';

import { useCallback, useEffect, useMemo, useState } from 'react';
import { Header } from '@/components/layout/header';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
  Banknote,
  BarChart3,
  FolderOpen,
  Package,
  ShoppingCart,
  TrendingUp,
  Users,
} from 'lucide-react';
import api from '@/lib/api';
import type { ApiResponse } from '@/lib/api/types';
import { PERMISSIONS, useAuth } from '@/contexts/auth-context';
import { useGlobalError } from '@/contexts/error-context';

import { ReportsCategories } from './reports-categories';
import { ReportsControls } from './reports-controls';
import { ReportsCash } from './reports-cash';
import { ReportsEmployees } from './reports-employees';
import { ReportsErrorBanner } from './reports-feedback';
import {
  buildExportFilename,
  getDateRange,
  triggerDownload,
} from './reports-helpers';
import { ReportsInventory } from './reports-inventory';
import { ReportsOverview } from './reports-overview';
import { ReportsSales } from './reports-sales';
import { ReportsTopSellers } from './reports-top-sellers';
import type { DatePresetKey, ReportState, ReportType } from './reports-types';

const INITIAL_DATE_PRESET: DatePresetKey = 'last30';

const INITIAL_REPORT_STATE: ReportState = {
  salesRangeReport: null,
  topSellers: [],
  inventoryReport: null,
  cashReport: null,
  employeeSales: [],
  categorySales: [],
  expenseSummary: null,
};

function requireApiData<T>(response: ApiResponse<T>, fallbackMessage: string): T | undefined {
  if (response.error) {
    throw new Error(response.error || fallbackMessage);
  }

  return response.data;
}

export function ReportsScreen() {
  const { hasPermission } = useAuth();
  const canExport = hasPermission(PERMISSIONS.REPORTS_EXPORT);
  const { showError } = useGlobalError();

  const [activeTab, setActiveTab] = useState<ReportType>('overview');
  const [isLoading, setIsLoading] = useState(true);
  const [datePreset, setDatePreset] = useState<DatePresetKey>(INITIAL_DATE_PRESET);
  const [dateRange, setDateRange] = useState(getDateRange(INITIAL_DATE_PRESET));
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [reportState, setReportState] = useState<ReportState>(INITIAL_REPORT_STATE);

  const setPartialState = useCallback((partialState: Partial<ReportState>) => {
    setReportState((currentState) => ({ ...currentState, ...partialState }));
  }, []);

  const handlePresetChange = useCallback((preset: DatePresetKey) => {
    setDatePreset(preset);
    setDateRange(getDateRange(preset));
  }, []);

  const runFetch = useCallback(async (onSuccess: () => Promise<void>, failureMessage: string) => {
    setIsLoading(true);
    setErrorMessage(null);
    try {
      await onSuccess();
    } catch {
      setErrorMessage(failureMessage);
    } finally {
      setIsLoading(false);
    }
  }, []);

  const fetchOverviewData = useCallback(async () => {
    await runFetch(async () => {
      const [salesResult, topSellersResult, expenseSummaryResult] = await Promise.all([
        api.getSalesRangeReport({
          start_date: dateRange.start,
          end_date: dateRange.end,
        }),
        api.getTopSellers({
          from: dateRange.start,
          to: dateRange.end,
          limit: 5,
        }),
        api.getExpenseSummary({
          start_date: dateRange.start,
          end_date: dateRange.end,
        }),
      ]);

      setPartialState({
        salesRangeReport: requireApiData(salesResult, 'Failed to fetch overview sales data.') || null,
        topSellers: requireApiData(topSellersResult, 'Failed to fetch overview top sellers.') || [],
        expenseSummary:
          requireApiData(expenseSummaryResult, 'Failed to fetch overview expense summary.') || null,
      });
    }, 'Failed to fetch overview data.');
  }, [dateRange.end, dateRange.start, runFetch, setPartialState]);

  const fetchSalesData = useCallback(async () => {
    await runFetch(async () => {
      const result = await api.getSalesRangeReport({
        start_date: dateRange.start,
        end_date: dateRange.end,
      });
      setPartialState({
        salesRangeReport: requireApiData(result, 'Failed to fetch sales data.') || null,
      });
    }, 'Failed to fetch sales data.');
  }, [dateRange.end, dateRange.start, runFetch, setPartialState]);

  const fetchTopSellers = useCallback(async () => {
    await runFetch(async () => {
      const result = await api.getTopSellers({
        from: dateRange.start,
        to: dateRange.end,
        limit: 50,
      });
      setPartialState({
        topSellers: requireApiData(result, 'Failed to fetch top sellers.') || [],
      });
    }, 'Failed to fetch top sellers.');
  }, [dateRange.end, dateRange.start, runFetch, setPartialState]);

  const fetchInventoryReport = useCallback(async () => {
    await runFetch(async () => {
      const result = await api.getInventoryReport(true);
      setPartialState({
        inventoryReport: requireApiData(result, 'Failed to fetch inventory report.') || null,
      });
    }, 'Failed to fetch inventory report.');
  }, [runFetch, setPartialState]);

  const fetchCashReport = useCallback(async () => {
    await runFetch(async () => {
      const result = await api.getCashReport({
        start_date: dateRange.start,
        end_date: dateRange.end,
      });
      setPartialState({
        cashReport: requireApiData(result, 'Failed to fetch cash report.') || null,
      });
    }, 'Failed to fetch cash report.');
  }, [dateRange.end, dateRange.start, runFetch, setPartialState]);

  const fetchEmployeeSales = useCallback(async () => {
    await runFetch(async () => {
      const result = await api.getEmployeeSalesReport({
        start_date: dateRange.start,
        end_date: dateRange.end,
      });
      setPartialState({
        employeeSales: requireApiData(result, 'Failed to fetch employee sales.') || [],
      });
    }, 'Failed to fetch employee sales.');
  }, [dateRange.end, dateRange.start, runFetch, setPartialState]);

  const fetchCategorySales = useCallback(async () => {
    await runFetch(async () => {
      const result = await api.getCategorySalesReport({
        start_date: dateRange.start,
        end_date: dateRange.end,
      });
      setPartialState({
        categorySales: requireApiData(result, 'Failed to fetch category sales.') || [],
      });
    }, 'Failed to fetch category sales.');
  }, [dateRange.end, dateRange.start, runFetch, setPartialState]);

  const fetchers = useMemo<Record<ReportType, () => Promise<void>>>(
    () => ({
      overview: fetchOverviewData,
      sales: fetchSalesData,
      'top-sellers': fetchTopSellers,
      inventory: fetchInventoryReport,
      cash: fetchCashReport,
      employees: fetchEmployeeSales,
      categories: fetchCategorySales,
    }),
    [
      fetchCategorySales,
      fetchCashReport,
      fetchEmployeeSales,
      fetchInventoryReport,
      fetchOverviewData,
      fetchSalesData,
      fetchTopSellers,
    ],
  );

  useEffect(() => {
    void fetchers[activeTab]();
  }, [activeTab, fetchers]);

  const exportReport = useCallback(
    async (tab: 'overview' | 'sales' | 'inventory' | 'top-sellers') => {
      if (!canExport) {
        showError("Permission Denied", "You do not have permission to export reports");
        return;
      }

      setErrorMessage(null);
      try {
        let url = '';
        switch (tab) {
          case 'overview':
            url = await api.exportComprehensiveReportCSV({
              start_date: dateRange.start,
              end_date: dateRange.end,
            });
            break;
          case 'sales':
            url = await api.exportSalesCSV({
              start_date: dateRange.start,
              end_date: dateRange.end,
            });
            break;
          case 'inventory':
            url = await api.exportInventoryCSV();
            break;
          case 'top-sellers':
            url = await api.exportTopSellersCSV({
              start_date: dateRange.start,
              end_date: dateRange.end,
              limit: 100,
            });
            break;
        }

        triggerDownload(url, buildExportFilename(tab, dateRange));
      } catch {
        const message = `Failed to export ${tab.replace('-', ' ')} report.`;
        setErrorMessage(message);
        showError("Export Failed", message);
      }
    },
    [canExport, dateRange, showError],
  );

  return (
    <div className="flex flex-col h-full">
      <Header title="Reports" />

      <div className="flex-1 p-6 overflow-auto">
        <ReportsErrorBanner message={errorMessage} />

        <Tabs value={activeTab} onValueChange={(value) => setActiveTab(value as ReportType)}>
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

          <TabsContent value="overview">
            <ReportsControls
              datePreset={datePreset}
              onPresetChange={handlePresetChange}
              dateRange={dateRange}
              onDateRangeChange={setDateRange}
              onRefresh={() => void fetchOverviewData()}
              onExport={() => void exportReport('overview')}
              canExport={canExport}
              isLoading={isLoading}
              exportLabel="Export"
            />
            <ReportsOverview
              isLoading={isLoading}
              salesRangeReport={reportState.salesRangeReport}
              topSellers={reportState.topSellers}
              expenseSummary={reportState.expenseSummary}
              dateRange={dateRange}
              onViewAllTopSellers={() => setActiveTab('top-sellers')}
            />
          </TabsContent>

          <TabsContent value="sales">
            <ReportsControls
              datePreset={datePreset}
              onPresetChange={handlePresetChange}
              dateRange={dateRange}
              onDateRangeChange={setDateRange}
              onExport={() => void exportReport('sales')}
              canExport={canExport}
              exportLabel="Export CSV"
            />
            <ReportsSales isLoading={isLoading} salesRangeReport={reportState.salesRangeReport} />
          </TabsContent>

          <TabsContent value="top-sellers">
            <ReportsControls
              datePreset={datePreset}
              onPresetChange={handlePresetChange}
              dateRange={dateRange}
              onDateRangeChange={setDateRange}
              onExport={() => void exportReport('top-sellers')}
              canExport={canExport}
              exportLabel="Export CSV"
            />
            <ReportsTopSellers
              isLoading={isLoading}
              topSellers={reportState.topSellers}
              dateRange={dateRange}
            />
          </TabsContent>

          <TabsContent value="inventory">
            <ReportsControls
              datePreset={datePreset}
              onPresetChange={handlePresetChange}
              dateRange={dateRange}
              onDateRangeChange={setDateRange}
              onRefresh={() => void fetchInventoryReport()}
              onExport={() => void exportReport('inventory')}
              canExport={canExport}
              isLoading={isLoading}
              showDateRange={false}
              exportLabel="Export CSV"
            />
            <ReportsInventory isLoading={isLoading} inventoryReport={reportState.inventoryReport} />
          </TabsContent>

          <TabsContent value="cash">
            <ReportsControls
              datePreset={datePreset}
              onPresetChange={handlePresetChange}
              dateRange={dateRange}
              onDateRangeChange={setDateRange}
              onRefresh={() => void fetchCashReport()}
              isLoading={isLoading}
            />
            <ReportsCash isLoading={isLoading} cashReport={reportState.cashReport} />
          </TabsContent>

          <TabsContent value="employees">
            <ReportsControls
              datePreset={datePreset}
              onPresetChange={handlePresetChange}
              dateRange={dateRange}
              onDateRangeChange={setDateRange}
            />
            <ReportsEmployees
              isLoading={isLoading}
              employeeSales={reportState.employeeSales}
              dateRange={dateRange}
            />
          </TabsContent>

          <TabsContent value="categories">
            <ReportsControls
              datePreset={datePreset}
              onPresetChange={handlePresetChange}
              dateRange={dateRange}
              onDateRangeChange={setDateRange}
            />
            <ReportsCategories
              isLoading={isLoading}
              categorySales={reportState.categorySales}
              dateRange={dateRange}
            />
          </TabsContent>
        </Tabs>
      </div>
    </div>
  );
}

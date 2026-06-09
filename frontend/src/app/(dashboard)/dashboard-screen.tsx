"use client";

import Link from "next/link";
import {
  useCallback,
  useEffect,
  useMemo,
  useState,
  type Dispatch,
  type ReactNode,
  type SetStateAction,
} from "react";
import {
  Boxes,
  Clock,
  History,
  Loader2,
  Package,
  Receipt,
  Wallet,
} from "lucide-react";

import { Header } from "@/components/layout/header";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useAuth, PERMISSIONS } from "@/contexts/auth-context";
import api from "@/lib/api";
import type { AuditLog, LowStockItem, Shift } from "@/types";

import { DashboardChangesList } from "./dashboard-changes-list";
import {
  buildDashboardStats,
  type DashboardStats as DashboardStatsData,
} from "./dashboard-helpers";
import { DashboardLowStock } from "./dashboard-low-stock";
import { DashboardShiftHistory } from "./dashboard-shift-history";
import { DashboardStats } from "./dashboard-stats";

type ChangeTab = "product" | "inventory" | "sale" | "expense";

const CHANGE_TABS: { value: ChangeTab; label: string; icon: ReactNode }[] = [
  {
    value: "product",
    label: "Products",
    icon: <Package className="h-4 w-4" />,
  },
  {
    value: "inventory",
    label: "Inventory",
    icon: <Boxes className="h-4 w-4" />,
  },
  { value: "sale", label: "Sales", icon: <Receipt className="h-4 w-4" /> },
  { value: "expense", label: "Expenses", icon: <Wallet className="h-4 w-4" /> },
];

export function DashboardScreen() {
  const { user, hasPermission } = useAuth();
  const canViewSales = hasPermission(PERMISSIONS.SALES_VIEW);
  const canViewInventory = hasPermission(PERMISSIONS.INVENTORY_VIEW);
  const canViewShifts = hasPermission(PERMISSIONS.SHIFTS_VIEW);
  const canViewChanges = hasPermission(PERMISSIONS.CHANGES_VIEW);
  const [stats, setStats] = useState<DashboardStatsData | null>(null);
  const [lowStockItems, setLowStockItems] = useState<LowStockItem[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [dashboardError, setDashboardError] = useState<string | null>(null);
  const [shifts, setShifts] = useState<Shift[]>([]);
  const [isShiftsLoading, setIsShiftsLoading] = useState(true);
  const [shiftsError, setShiftsError] = useState<string | null>(null);
  const [activeChangeTab, setActiveChangeTab] = useState<ChangeTab>("product");
  const [changeLogs, setChangeLogs] = useState<Record<ChangeTab, AuditLog[]>>({
    product: [],
    inventory: [],
    sale: [],
    expense: [],
  });
  const [changeLoading, setChangeLoading] = useState<
    Record<ChangeTab, boolean>
  >({
    product: false,
    inventory: false,
    sale: false,
    expense: false,
  });
  const [changeLoaded, setChangeLoaded] = useState<Record<ChangeTab, boolean>>({
    product: false,
    inventory: false,
    sale: false,
    expense: false,
  });
  const [changeErrors, setChangeErrors] = useState<
    Record<ChangeTab, string | null>
  >({
    product: null,
    inventory: null,
    sale: null,
    expense: null,
  });

  const setChangeTabState = useCallback(
    <T,>(
      setter: Dispatch<SetStateAction<Record<ChangeTab, T>>>,
      entityType: ChangeTab,
      value: T,
    ) => {
      setter((prev) => ({ ...prev, [entityType]: value }));
    },
    [],
  );

  const fetchDashboardData = useCallback(async () => {
    setIsLoading(true);
    setDashboardError(null);

    const summaryPromise = canViewSales
      ? api.getDailySummary()
      : Promise.resolve({ data: null, error: null });
    const lowStockPromise = canViewInventory
      ? api.getLowStock()
      : Promise.resolve({ data: [], error: null });

    const [summaryResult, lowStockResult] = await Promise.all([
      summaryPromise,
      lowStockPromise,
    ]);

    if (canViewInventory && lowStockResult.error) {
      setDashboardError(
        lowStockResult.error || "Could not load dashboard data",
      );
      setIsLoading(false);
      return;
    }

    const lowStock = lowStockResult.data || [];
    setStats(
      buildDashboardStats(
        summaryResult.data?.total_sales,
        summaryResult.data?.transaction_count,
        lowStock.length,
      ),
    );
    setLowStockItems(lowStock.slice(0, 5));
    setIsLoading(false);
  }, [canViewInventory, canViewSales]);

  const fetchShiftPreview = useCallback(async () => {
    if (!canViewShifts) {
      setShifts([]);
      setIsShiftsLoading(false);
      setShiftsError(null);
      return;
    }

    setIsShiftsLoading(true);
    setShiftsError(null);
    const result = await api.getShifts();
    if (result.error) {
      setShiftsError("Could not load shifts");
    } else {
      setShifts((result.data || []).slice(0, 5));
    }
    setIsShiftsLoading(false);
  }, [canViewShifts]);

  const fetchChangeLogs = useCallback(
    async (entityType: ChangeTab) => {
      if (!canViewChanges) {
        setChangeTabState<AuditLog[]>(setChangeLogs, entityType, []);
        setChangeTabState<boolean>(setChangeLoaded, entityType, true);
        setChangeTabState<boolean>(setChangeLoading, entityType, false);
        setChangeTabState<string | null>(setChangeErrors, entityType, null);
        return;
      }

      setChangeTabState<boolean>(setChangeLoading, entityType, true);
      setChangeTabState<string | null>(setChangeErrors, entityType, null);
      const result = await api.getDashboardChanges({
        entity_type: entityType,
        limit: 5,
      });
      if (result.error) {
        setChangeTabState<string | null>(
          setChangeErrors,
          entityType,
          "Could not load changes",
        );
      } else {
        setChangeTabState<AuditLog[]>(
          setChangeLogs,
          entityType,
          result.data || [],
        );
      }
      setChangeTabState<boolean>(setChangeLoaded, entityType, true);
      setChangeTabState<boolean>(setChangeLoading, entityType, false);
    },
    [canViewChanges, setChangeTabState],
  );

  useEffect(() => {
    const timer = window.setTimeout(() => {
      void fetchDashboardData();
      void fetchShiftPreview();
      if (canViewChanges) {
        void fetchChangeLogs("product");
      }
    }, 0);

    return () => window.clearTimeout(timer);
  }, [canViewChanges, fetchChangeLogs, fetchDashboardData, fetchShiftPreview]);

  useEffect(() => {
    if (canViewChanges && !changeLoaded[activeChangeTab] && !changeLoading[activeChangeTab]) {
      const timer = window.setTimeout(() => {
        void fetchChangeLogs(activeChangeTab);
      }, 0);

      return () => window.clearTimeout(timer);
    }
  }, [activeChangeTab, canViewChanges, changeLoaded, changeLoading, fetchChangeLogs]);

  const activeLogs = useMemo(
    () => changeLogs[activeChangeTab],
    [activeChangeTab, changeLogs],
  );

  return (
    <div className="flex flex-col h-full">
      <Header title="Dashboard" />

      <div className="flex-1 p-6 overflow-auto">
        <div className="mb-6">
          <h2 className="text-2xl font-bold">Welcome back, {user?.name}!</h2>
          <p className="text-muted-foreground">
            Here&apos;s what&apos;s happening with your store today.
          </p>
        </div>

        {isLoading ? (
          <div className="flex items-center justify-center h-64">
            <Loader2 className="h-8 w-8 animate-spin text-primary" />
          </div>
        ) : dashboardError ? (
          <div className="rounded-md border border-destructive/20 bg-destructive/10 p-4 text-sm text-destructive mb-6">
            {dashboardError}
          </div>
        ) : (
          <>
            <DashboardStats stats={stats} showLowStock={canViewInventory} />
            <DashboardLowStock items={lowStockItems} />
          </>
        )}

        {canViewShifts ? (
        <Card className="mb-6 flex flex-col border-0 shadow-none bg-transparent md:border md:shadow md:bg-card">
          <CardHeader className="px-0 pt-0 pb-4 md:p-6">
            <div className="flex items-center gap-2">
              <Clock className="h-5 w-5 text-primary" />
              <CardTitle>Shift History</CardTitle>
            </div>
            <CardDescription>
              Recent and active shared store shifts
            </CardDescription>
          </CardHeader>
          <CardContent className="flex-1 px-0 pb-4 md:px-6 md:pb-6 md:pt-0">
            <DashboardShiftHistory
              shifts={shifts}
              isLoading={isShiftsLoading}
              error={shiftsError}
              onRetry={() => void fetchShiftPreview()}
            />
          </CardContent>
          <div className="mt-auto px-0 pb-0 md:p-4 md:pt-0">
            <Button
              variant="outline"
              className="w-full bg-background md:bg-transparent"
              asChild
            >
              <Link href="/shifts">View More</Link>
            </Button>
          </div>
        </Card>
        ) : null}

        {canViewChanges ? (
        <Card className="flex flex-col border-0 shadow-none bg-transparent md:border md:shadow md:bg-card">
          <CardHeader className="px-0 pt-0 pb-4 md:p-6">
            <div className="flex items-center gap-2">
              <History className="h-5 w-5 text-primary" />
              <CardTitle>Recent Changes</CardTitle>
            </div>
            <CardDescription>
              Activity log across all areas of your store
            </CardDescription>
          </CardHeader>
          <CardContent className="flex-1 px-0 pb-4 md:px-6 md:pb-6 md:pt-0">
            <Tabs
              defaultValue="product"
              value={activeChangeTab}
              onValueChange={(value) => setActiveChangeTab(value as ChangeTab)}
            >
              <TabsList className="w-full justify-start flex overflow-x-auto lg:w-auto lg:inline-flex mb-4 no-scrollbar pb-1">
                {CHANGE_TABS.map((tab) => (
                  <TabsTrigger
                    key={tab.value}
                    value={tab.value}
                    className="gap-1.5"
                  >
                    {tab.icon}
                    {tab.label}
                  </TabsTrigger>
                ))}
              </TabsList>

              {CHANGE_TABS.map((tab) => (
                <TabsContent key={tab.value} value={tab.value}>
                  <DashboardChangesList
                    logs={
                      tab.value === activeChangeTab
                        ? activeLogs
                        : changeLogs[tab.value]
                    }
                    isLoading={changeLoading[tab.value]}
                    error={changeErrors[tab.value]}
                    onRetry={() => void fetchChangeLogs(tab.value)}
                  />
                </TabsContent>
              ))}
            </Tabs>
          </CardContent>
          <div className="mt-auto px-0 pb-0 md:p-4 md:pt-0">
            <Button
              variant="outline"
              className="w-full bg-background md:bg-transparent"
              asChild
            >
              <Link href="/changes">View More</Link>
            </Button>
          </div>
        </Card>
        ) : null}
      </div>
    </div>
  );
}

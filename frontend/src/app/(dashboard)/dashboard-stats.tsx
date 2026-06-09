'use client';

import { DollarSign, Package, ShoppingCart, TrendingUp } from "lucide-react";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

import type { DashboardStats as DashboardStatsData } from "./dashboard-helpers";
import { formatDashboardCurrency } from "./dashboard-helpers";

interface DashboardStatsProps {
  stats: DashboardStatsData | null;
  showLowStock?: boolean;
}

export function DashboardStats({ stats, showLowStock = true }: DashboardStatsProps) {
  return (
    <div className={`grid gap-4 md:grid-cols-2 ${showLowStock ? "lg:grid-cols-4" : "lg:grid-cols-3"} mb-6`}>
      <Card>
        <CardHeader className="flex flex-row items-center justify-between pb-2">
          <CardTitle className="text-sm font-medium">Today&apos;s Sales</CardTitle>
          <DollarSign className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold font-mono">{formatDashboardCurrency(stats?.todaySales || 0)}</div>
          <p className="text-xs text-muted-foreground">Total revenue today</p>
        </CardContent>
      </Card>

      <Card>
        <CardHeader className="flex flex-row items-center justify-between pb-2">
          <CardTitle className="text-sm font-medium">Transactions</CardTitle>
          <ShoppingCart className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">{stats?.todayTransactions || 0}</div>
          <p className="text-xs text-muted-foreground">Sales completed today</p>
        </CardContent>
      </Card>

      <Card>
        <CardHeader className="flex flex-row items-center justify-between pb-2">
          <CardTitle className="text-sm font-medium">Average Sale</CardTitle>
          <TrendingUp className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold font-mono">{formatDashboardCurrency(stats?.averageSale || 0)}</div>
          <p className="text-xs text-muted-foreground">Per transaction</p>
        </CardContent>
      </Card>

      {showLowStock ? (
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium">Low Stock</CardTitle>
            <Package className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats?.lowStockCount || 0}</div>
            <p className="text-xs text-muted-foreground">Products need attention</p>
          </CardContent>
        </Card>
      ) : null}
    </div>
  );
}

'use client';

import { useEffect, useState } from 'react';
import { useAuth } from '@/contexts/auth-context';
import { Header } from '@/components/layout/header';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { 
  DollarSign, 
  ShoppingCart, 
  Package, 
  TrendingUp,
  AlertTriangle,
  Loader2 
} from 'lucide-react';
import api from '@/lib/api';
import { DailySummary, LowStockItem } from '@/types';

interface DashboardStats {
  todaySales: number;
  todayTransactions: number;
  averageSale: number;
  lowStockCount: number;
}

export default function DashboardPage() {
  const { user } = useAuth();
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [lowStockItems, setLowStockItems] = useState<LowStockItem[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const fetchDashboardData = async () => {
      try {
        const [summaryResult, lowStockResult] = await Promise.all([
          api.getDailySummary(),
          api.getLowStock(),
        ]);

        const summary = summaryResult.data;
        const lowStock = lowStockResult.data || [];

        // Calculate average sale
        const totalSales = summary?.total_sales ? parseFloat(summary.total_sales) : 0;
        const transactionCount = summary?.transaction_count || 0;
        const averageSale = transactionCount > 0 ? totalSales / transactionCount : 0;

        setStats({
          todaySales: totalSales,
          todayTransactions: transactionCount,
          averageSale: averageSale,
          lowStockCount: lowStock.length,
        });

        setLowStockItems(lowStock.slice(0, 5));
      } catch (error) {
        console.error('Failed to fetch dashboard data:', error);
      } finally {
        setIsLoading(false);
      }
    };

    fetchDashboardData();
  }, []);

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: 'IDR',
      minimumFractionDigits: 0,
    }).format(amount);
  };

  return (
    <div className="flex flex-col h-full">
      <Header title="Dashboard" />
      
      <div className="flex-1 p-6">
        {/* Welcome message */}
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
        ) : (
          <>
            {/* Stats cards */}
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4 mb-6">
              <Card>
                <CardHeader className="flex flex-row items-center justify-between pb-2">
                  <CardTitle className="text-sm font-medium">Today&apos;s Sales</CardTitle>
                  <DollarSign className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">
                    {formatCurrency(stats?.todaySales || 0)}
                  </div>
                  <p className="text-xs text-muted-foreground">
                    Total revenue today
                  </p>
                </CardContent>
              </Card>

              <Card>
                <CardHeader className="flex flex-row items-center justify-between pb-2">
                  <CardTitle className="text-sm font-medium">Transactions</CardTitle>
                  <ShoppingCart className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">
                    {stats?.todayTransactions || 0}
                  </div>
                  <p className="text-xs text-muted-foreground">
                    Sales completed today
                  </p>
                </CardContent>
              </Card>

              <Card>
                <CardHeader className="flex flex-row items-center justify-between pb-2">
                  <CardTitle className="text-sm font-medium">Average Sale</CardTitle>
                  <TrendingUp className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">
                    {formatCurrency(stats?.averageSale || 0)}
                  </div>
                  <p className="text-xs text-muted-foreground">
                    Per transaction
                  </p>
                </CardContent>
              </Card>

              <Card>
                <CardHeader className="flex flex-row items-center justify-between pb-2">
                  <CardTitle className="text-sm font-medium">Low Stock</CardTitle>
                  <Package className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">
                    {stats?.lowStockCount || 0}
                  </div>
                  <p className="text-xs text-muted-foreground">
                    Products need attention
                  </p>
                </CardContent>
              </Card>
            </div>

            {/* Low stock alert */}
            {lowStockItems.length > 0 && (
              <Card>
                <CardHeader>
                  <div className="flex items-center gap-2">
                    <AlertTriangle className="h-5 w-5 text-yellow-500" />
                    <CardTitle>Low Stock Alert</CardTitle>
                  </div>
                  <CardDescription>
                    These products are running low and need to be restocked
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="space-y-3">
                    {lowStockItems.map((item) => (
                      <div
                        key={item.id}
                        className="flex items-center justify-between rounded-lg border p-3"
                      >
                        <div>
                          <p className="font-medium">{item.name}</p>
                          <p className="text-sm text-muted-foreground">
                            SKU: {item.sku || 'N/A'}
                          </p>
                        </div>
                        <div className="text-right">
                          <p className="font-bold text-destructive">
                            {parseFloat(item.quantity)} left
                          </p>
                          <p className="text-xs text-muted-foreground">
                            Available: {parseFloat(item.available_quantity)}
                          </p>
                        </div>
                      </div>
                    ))}
                  </div>
                </CardContent>
              </Card>
            )}
          </>
        )}
      </div>
    </div>
  );
}

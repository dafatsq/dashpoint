'use client';

import { useEffect, useState, useCallback } from 'react';
import { useAuth } from '@/contexts/auth-context';
import { Header } from '@/components/layout/header';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Button } from '@/components/ui/button';
import Link from 'next/link';
import {
  DollarSign,
  ShoppingCart,
  Package,
  TrendingUp,
  AlertTriangle,
  Loader2,
  Clock,
  Boxes,
  Receipt,
  Wallet,
  History,
  User,
  ArrowRightLeft,
  CircleDot,
  CheckCircle2,
} from 'lucide-react';
import api from '@/lib/api';
import { DailySummary, LowStockItem, AuditLog, Shift } from '@/types';

interface DashboardStats {
  todaySales: number;
  todayTransactions: number;
  averageSale: number;
  lowStockCount: number;
}

// ─── Audit Log Helpers ───────────────────────────────────────────────

const ACTION_LABELS: Record<string, string> = {
  'create': 'Created',
  'update': 'Updated',
  'delete': 'Deleted',
  'void': 'Voided',
  'adjust': 'Adjusted',
  'count': 'Counted',
  'start': 'Started',
  'close': 'Closed',
};

const getActionVerb = (action: string): string => {
  // Actions may be dotted like "product.create" or plain like "create"
  const parts = action.split('.');
  return parts[parts.length - 1];
};

const getActionBadgeColor = (action: string) => {
  const verb = getActionVerb(action);
  switch (verb) {
    case 'create':
    case 'start':
      return 'bg-green-600 text-white';
    case 'update':
    case 'close':
      return 'bg-yellow-600 text-white';
    case 'delete':
      return 'bg-red-600 text-white';
    case 'void':
      return 'bg-orange-600 text-white';
    case 'adjust':
      return 'bg-purple-600 text-white';
    case 'count':
      return 'bg-blue-600 text-white';
    default:
      return 'bg-gray-600 text-white';
  }
};

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleString('id-ID', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
};

const formatDateTime = (dateString: string) => {
  return new Date(dateString).toLocaleString('id-ID', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
};

function formatCurrencyShort(amount: number): string {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
  }).format(amount);
}

function getChangeDescription(log: AuditLog): string {
  const newVals = log.new_values || {};
  const oldVals = log.old_values || {};
  const verb = getActionVerb(log.action);

  if (log.entity_type === 'product') {
    const name = String(newVals.name || oldVals.name || '');
    if (verb === 'create') return name || 'New product';
    if (verb === 'update') return name || 'Updated product';
    if (verb === 'delete') return name || 'Deleted product';
  }

  if (log.entity_type === 'inventory') {
    const productName = newVals.product_name || oldVals.product_name || newVals.affected_product || oldVals.affected_product || '';
    const adjType = newVals.adjustment_type as string | undefined;
    const qty = newVals.quantity;
    const newQty = newVals.new_quantity;
    const adjTypeLabel = adjType
      ? adjType.charAt(0).toUpperCase() + adjType.slice(1)
      : 'Adjusted';
    if (productName && qty !== undefined && newQty !== undefined) {
      return `${adjTypeLabel}: ${productName} → ${newQty} (Δ${qty})`;
    }
    if (productName && qty !== undefined) return `${adjTypeLabel} ${productName}: ${qty}`;
    if (productName) return `Stock change: ${productName}`;
    return 'Stock adjustment';
  }

  if (log.entity_type === 'sale') {
    const invoice = String(newVals.invoice_no || '');
    if (verb === 'create') {
      return invoice || 'New sale created';
    }
    if (verb === 'void') return invoice || 'Voided sale';
    if (verb === 'delete') return invoice || 'Deleted sale';
  }

  if (log.entity_type === 'expense') {
    const desc = String(newVals.description || newVals.affected_expense || oldVals.affected_expense || '');
    const amount = newVals.amount;
    if (verb === 'create') return `${desc}${amount ? ` — ${formatCurrencyShort(Number(amount))}` : ''}`;
    if (verb === 'update') return desc || 'Updated expense';
    if (verb === 'delete') return desc || 'Deleted expense';
  }

  return `${ACTION_LABELS[verb] || verb} ${log.entity_type}`;
}

// ─── Change Tabs Config (without Shifts) ─────────────────────────────

type ChangeTab = 'product' | 'inventory' | 'sale' | 'expense';

const CHANGE_TABS: { value: ChangeTab; label: string; icon: React.ReactNode }[] = [
  { value: 'product', label: 'Products', icon: <Package className="h-4 w-4" /> },
  { value: 'inventory', label: 'Inventory', icon: <Boxes className="h-4 w-4" /> },
  { value: 'sale', label: 'Sales', icon: <Receipt className="h-4 w-4" /> },
  { value: 'expense', label: 'Expenses', icon: <Wallet className="h-4 w-4" /> },
];

// ─── ChangesList Component ───────────────────────────────────────────

function ChangesList({ entityType }: { entityType: ChangeTab }) {
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const fetchLogs = async () => {
      setIsLoading(true);
      try {
        const result = await api.getDashboardChanges({ entity_type: entityType, limit: 5 });
        if (result.data) {
          setLogs(result.data);
        }
      } catch (error) {
        console.error('Failed to fetch change logs:', error);
      } finally {
        setIsLoading(false);
      }
    };
    fetchLogs();
  }, [entityType]);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="h-6 w-6 animate-spin text-primary" />
      </div>
    );
  }

  if (logs.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12 text-muted-foreground">
        <History className="h-10 w-10 mb-2" />
        <p className="text-sm">No changes recorded yet</p>
      </div>
    );
  }

  // Fields to skip (already shown in the description or not useful)
  const SKIP_FIELDS = new Set(['affected_product', 'affected_category', 'affected_expense', 'affected_user', 'product_name', 'invoice_no', 'category_id', 'product_id']);

  const renderFieldChanges = (log: AuditLog) => {
    const oldVals = log.old_values || {};
    const newVals = log.new_values || {};
    const verb = getActionVerb(log.action);

    // Collect changed fields
    const allKeys = new Set([...Object.keys(oldVals), ...Object.keys(newVals)]);
    const changes: { key: string; oldVal: unknown; newVal: unknown }[] = [];

    allKeys.forEach((key) => {
      if (SKIP_FIELDS.has(key)) return;
      const oldVal = oldVals[key];
      const newVal = newVals[key];

      if (verb === 'update' || verb === 'close') {
        // Only show fields that actually changed
        if (JSON.stringify(oldVal) !== JSON.stringify(newVal)) {
          changes.push({ key, oldVal, newVal });
        }
      } else if (verb === 'create' || verb === 'start' || verb === 'adjust' || verb === 'count' || verb === 'void') {
        // Show new values
        if (newVal !== undefined && newVal !== null && newVal !== '') {
          changes.push({ key, oldVal: undefined, newVal });
        }
      } else if (verb === 'delete') {
        // Show old values for deleted items
        if (oldVal !== undefined && oldVal !== null && oldVal !== '') {
          changes.push({ key, oldVal, newVal: undefined });
        }
      }
    });

    if (changes.length === 0) return null;

    const formatFieldName = (key: string): string => {
      if (key === 'image_url') return 'Photo';
      if (key === 'is_active') return 'Active';
      if (key === 'adjustment_type') return 'Type';
      if (key === 'new_quantity') return 'New Stock';
      if (key === 'quantity') return 'Change';
      if (key === 'reason') return verb === 'void' ? 'Reason' : 'Notes';
      return key.replace(/_/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase());
    };

    const isImageField = (key: string) => key === 'image_url';

    const formatValue = (key: string, val: unknown): string => {
      if (val === null || val === undefined) return '—';
      if (typeof val === 'boolean') return val ? 'Yes' : 'No';
      if (typeof val === 'object') return JSON.stringify(val);
      if (key === 'tax_rate') return `${String(val)}%`;
      if (key === 'total' || key === 'amount') return formatCurrencyShort(Number(val));
      return String(val);
    };

    return (
      <div className="mt-2 ml-0 space-y-0.5">
        {changes.slice(0, 6).map(({ key, oldVal, newVal }) => (
          <div key={key} className="text-xs flex items-start gap-1.5 text-muted-foreground">
            <span className="font-medium text-foreground/70 shrink-0 pt-0.5">{formatFieldName(key)}:</span>
            {isImageField(key) ? (
              // Render image thumbnails for photo changes
              <div className="flex items-center gap-1.5">
                {oldVal !== undefined && (
                  <div className="flex flex-col items-center gap-0.5">
                    {/* eslint-disable-next-line @next/next/no-img-element */}
                    <img src={String(oldVal)} alt="old" className="h-8 w-8 object-cover rounded border border-red-300 opacity-60" />
                    <span className="text-[9px] text-red-500">before</span>
                  </div>
                )}
                {oldVal !== undefined && newVal !== undefined && (
                  <span className="text-muted-foreground">→</span>
                )}
                {newVal !== undefined && (
                  <div className="flex flex-col items-center gap-0.5">
                    {/* eslint-disable-next-line @next/next/no-img-element */}
                    <img src={String(newVal)} alt="new" className="h-8 w-8 object-cover rounded border border-green-300" />
                    <span className="text-[9px] text-green-600">after</span>
                  </div>
                )}
                {newVal === undefined && oldVal !== undefined && (
                  <span className="text-red-500 text-[10px]">(removed)</span>
                )}
              </div>
            ) : oldVal !== undefined && newVal !== undefined ? (
              <>
                <span className="text-red-500 line-through">{formatValue(key, oldVal)}</span>
                <span className="text-muted-foreground">→</span>
                <span className="text-green-600">{formatValue(key, newVal)}</span>
              </>
            ) : newVal !== undefined ? (
              <span className="text-foreground/60">{formatValue(key, newVal)}</span>
            ) : (
              <span className="text-red-500 line-through">{formatValue(key, oldVal)}</span>
            )}
          </div>
        ))}
        {changes.length > 6 && (
          <span className="text-xs text-muted-foreground">+{changes.length - 6} more</span>
        )}
      </div>
    );
  };

  return (
    <div className="space-y-4">
      {logs.map((log) => (
        <div
          key={log.id}
          className="rounded-xl border p-4 bg-card text-card-foreground shadow-sm hover:shadow-md transition-shadow"
        >
          {/* Top row: status + employee + time */}
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 mb-3">
            <div className="flex items-center gap-2 flex-wrap">
              <span
                className={`inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-semibold uppercase tracking-wide whitespace-nowrap ${getActionBadgeColor(
                  log.action
                )}`}
              >
                {ACTION_LABELS[getActionVerb(log.action)] || getActionVerb(log.action)}
              </span>
              <div className="flex items-center gap-1.5 text-sm font-medium">
                <User className="h-3.5 w-3.5 text-muted-foreground shrink-0" />
                <span className="truncate max-w-[150px] sm:max-w-[200px]">{log.user_name || 'System'}</span>
              </div>
            </div>
            <span className="text-xs text-muted-foreground whitespace-nowrap">
              {formatDate(log.created_at)}
            </span>
          </div>

          {/* Description and Details */}
          <div>
            <p className="text-base font-medium leading-snug mb-3">
              {getChangeDescription(log)}
            </p>
            {renderFieldChanges(log)}
          </div>
        </div>
      ))}
    </div>
  );
}

// ─── ShiftHistory Component ──────────────────────────────────────────

function ShiftHistory() {
  const [shifts, setShifts] = useState<Shift[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [fetchError, setFetchError] = useState(false);

  const fetchShifts = useCallback(async () => {
    setIsLoading(true);
    setFetchError(false);
    try {
      const result = await api.getShifts();
      if (result.data) {
        setShifts(result.data.slice(0, 5));
      } else if (result.error) {
        setFetchError(true);
      }
    } catch (error) {
      console.error('Failed to fetch shifts:', error);
      setFetchError(true);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchShifts();
  }, [fetchShifts]);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="h-6 w-6 animate-spin text-primary" />
      </div>
    );
  }

  if (fetchError) {
    return (
      <div className="flex flex-col items-center justify-center py-12 text-muted-foreground gap-2">
        <Clock className="h-10 w-10 mb-2" />
        <p className="text-sm">Could not load shifts</p>
        <button
          onClick={fetchShifts}
          className="text-xs text-primary hover:underline"
        >
          Retry
        </button>
      </div>
    );
  }

  if (shifts.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12 text-muted-foreground">
        <Clock className="h-10 w-10 mb-2" />
        <p className="text-sm">No shifts recorded yet</p>
      </div>
    );
  }

  return (
    <div className="space-y-3">
      {shifts.map((shift) => {
        const isOpen = shift.status === 'open';
        const openingCash = parseFloat(shift.opening_cash || '0');
        const closingCash = shift.closing_cash ? parseFloat(shift.closing_cash) : null;

        return (
          <div
            key={shift.id}
            className={`rounded-xl border p-4 shadow-sm transition-shadow hover:shadow-md ${isOpen
              ? 'border-green-500/50 bg-green-50/50 dark:bg-green-500/10'
              : 'bg-card text-card-foreground'
              }`}
          >
            {/* Top row: status + employee + time */}
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 mb-3">
              <div className="flex items-center gap-2">
                {isOpen ? (
                  <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-semibold uppercase tracking-wide bg-green-600 text-white whitespace-nowrap">
                    <CircleDot className="h-3 w-3" />
                    Open
                  </span>
                ) : (
                  <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-semibold uppercase tracking-wide bg-gray-600 text-white whitespace-nowrap">
                    <CheckCircle2 className="h-3 w-3" />
                    Closed
                  </span>
                )}
                <div className="flex items-center gap-1 text-sm font-medium flex-wrap">
                  <User className="h-3.5 w-3.5 text-muted-foreground flex-shrink-0" />
                  <span className="truncate max-w-[120px] sm:max-w-[200px]">{shift.employee_name || 'Unknown'}</span>
                  {!isOpen && shift.closed_by_name && (
                    <span className="text-muted-foreground text-[10px] sm:text-xs font-normal">
                      (Closed by {shift.closed_by_name})
                    </span>
                  )}
                </div>
              </div>
              <span className="text-xs text-muted-foreground whitespace-nowrap">
                {formatDateTime(shift.started_at)}
              </span>
            </div>

            {/* Cash info row */}
            <div className="flex flex-wrap items-center gap-x-4 gap-y-2 text-sm mt-3">
              <div className="flex items-center gap-1.5 whitespace-nowrap">
                <span className="text-muted-foreground text-xs">Open:</span>
                <span className="font-medium">{formatCurrencyShort(openingCash)}</span>
              </div>
              {!isOpen && closingCash !== null && (
                <>
                  <div className="flex items-center gap-1.5 whitespace-nowrap">
                    <ArrowRightLeft className="h-3.5 w-3.5 text-muted-foreground hidden sm:block" />
                    <span className="text-muted-foreground text-xs sm:ml-0">Close:</span>
                    <span className="font-medium">{formatCurrencyShort(closingCash)}</span>
                    {shift.cash_difference && (
                      <span className={`text-xs ml-1 font-medium ${parseFloat(shift.cash_difference) > 0 ? 'text-green-600' : parseFloat(shift.cash_difference) < 0 ? 'text-red-600' : 'text-muted-foreground'}`}>
                        ({parseFloat(shift.cash_difference) > 0 ? '+' : ''}{formatCurrencyShort(parseFloat(shift.cash_difference))})
                      </span>
                    )}
                  </div>
                </>
              )}
              {!isOpen && shift.ended_at && (
                <div className="w-full sm:w-auto sm:ml-auto text-[11px] sm:text-xs text-muted-foreground mt-1 sm:mt-0">
                  Ended: {formatDateTime(shift.ended_at)}
                </div>
              )}
            </div>

            {/* Operations List */}
            {shift.operations && shift.operations.length > 0 && (
              <div className="mt-3 pt-3 border-t border-dashed space-y-2">
                <p className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider mb-2">Cash Drawer Log</p>
                {shift.operations.map((op) => (
                  <div key={op.id} className="flex justify-between items-center bg-muted/30 p-2 rounded-md">
                    <div className="flex flex-col gap-0.5">
                      <div className="flex items-center gap-1.5">
                        <span className={`text-[10px] font-bold uppercase ${op.type === 'pay_in' ? 'text-green-600 dark:text-green-500' : 'text-red-600 dark:text-red-500'}`}>
                          {op.type === 'pay_in' ? '+ Pay In' : '- Pay Out'}
                        </span>
                        <span className="text-xs font-medium text-foreground">{op.reason || 'No reason specified'}</span>
                      </div>
                      <span className="text-[10px] text-muted-foreground">
                        {op.performed_by_name || 'System'} • {formatDateTime(op.created_at)}
                      </span>
                    </div>
                    <div className="flex flex-col items-end">
                      <span className={`text-sm font-semibold tabular-nums tracking-tight ${op.type === 'pay_in' ? 'text-green-700 dark:text-green-400' : 'text-red-700 dark:text-red-400'}`}>
                        {formatCurrencyShort(parseFloat(op.amount))}
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        );
      })}
    </div>
  );
}

// ─── Main Dashboard Page ─────────────────────────────────────────────

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

      <div className="flex-1 p-6 overflow-auto">
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
                  <div className="text-2xl font-bold">
                    {formatCurrency(stats?.averageSale || 0)}
                  </div>
                  <p className="text-xs text-muted-foreground">Per transaction</p>
                </CardContent>
              </Card>

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
            </div>

            {/* Low stock alert */}
            {lowStockItems.length > 0 && (
              <Card className="mb-6">
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

        {/* Shift History — dedicated section */}
        <Card className="mb-6 flex flex-col border-0 shadow-none bg-transparent md:border md:shadow md:bg-card">
          <CardHeader className="px-0 pt-0 pb-4 md:p-6">
            <div className="flex items-center gap-2">
              <Clock className="h-5 w-5 text-primary" />
              <CardTitle>Shift History</CardTitle>
            </div>
            <CardDescription>
              Recent and active shifts across all employees
            </CardDescription>
          </CardHeader>
          <CardContent className="flex-1 px-0 pb-4 md:px-6 md:pb-6 md:pt-0">
            <ShiftHistory />
          </CardContent>
          <div className="mt-auto px-0 pb-0 md:p-4 md:pt-0">
            <Button variant="outline" className="w-full bg-background md:bg-transparent" asChild>
              <Link href="/shifts">View More</Link>
            </Button>
          </div>
        </Card>

        {/* Recent Changes — audit-based tabs */}
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
            <Tabs defaultValue="product">
              <TabsList className="w-full justify-start flex overflow-x-auto lg:w-auto lg:inline-flex mb-4 no-scrollbar pb-1">
                {CHANGE_TABS.map((tab) => (
                  <TabsTrigger key={tab.value} value={tab.value} className="gap-1.5">
                    {tab.icon}
                    {tab.label}
                  </TabsTrigger>
                ))}
              </TabsList>

              {CHANGE_TABS.map((tab) => (
                <TabsContent key={tab.value} value={tab.value}>
                  <ChangesList entityType={tab.value} />
                </TabsContent>
              ))}
            </Tabs>
          </CardContent>
          <div className="mt-auto px-0 pb-0 md:p-4 md:pt-0">
            <Button variant="outline" className="w-full bg-background md:bg-transparent" asChild>
              <Link href="/changes">View More</Link>
            </Button>
          </div>
        </Card>
      </div>
    </div>
  );
}

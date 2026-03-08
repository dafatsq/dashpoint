'use client';

import { useEffect, useState } from 'react';
import { Header } from '@/components/layout/header';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
    Package,
    Boxes,
    Receipt,
    Wallet,
    Loader2,
    History,
    User as UserIcon,
    ChevronLeft,
    ChevronRight,
} from 'lucide-react';
import { DateRangePicker } from '@/components/ui/date-range-picker';
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import api from '@/lib/api';
import { AuditLog, User } from '@/types';

type ChangeTab = 'product' | 'inventory' | 'sale' | 'expense';

const CHANGE_TABS: { value: ChangeTab; label: string; icon: React.ReactNode }[] = [
    { value: 'product', label: 'Products', icon: <Package className="h-4 w-4" /> },
    { value: 'inventory', label: 'Inventory', icon: <Boxes className="h-4 w-4" /> },
    { value: 'sale', label: 'Sales', icon: <Receipt className="h-4 w-4" /> },
    { value: 'expense', label: 'Expenses', icon: <Wallet className="h-4 w-4" /> },
];

const ACTION_LABELS: Record<string, string> = {
    create: 'Created',
    update: 'Updated',
    delete: 'Deleted',
    void: 'Voided',
    adjust: 'Adjusted',
    count: 'Counted',
    start: 'Started',
    close: 'Closed',
    archive: 'Archived',
    restore: 'Restored',
};

const getActionVerb = (action: string): string => {
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
        case 'restore':
            return 'bg-yellow-600 text-white';
        case 'archive':
            return 'bg-orange-500 text-white';
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
        if (verb === 'archive') return name || 'Archived product';
        if (verb === 'restore') return name || 'Restored product';
        if (verb === 'delete') return name || 'Deleted product';
    }

    if (log.entity_type === 'inventory') {
        const productName = newVals.product_name || oldVals.product_name || newVals.affected_product || oldVals.affected_product || '';
        const adjType = newVals.adjustment_type as string | undefined;
        const qty = newVals.quantity;
        const newQty = newVals.new_quantity;
        const adjTypeLabel = adjType ? adjType.charAt(0).toUpperCase() + adjType.slice(1) : 'Adjusted';
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

    if (log.entity_type === 'user') {
        const name = String(newVals.name || oldVals.name || oldVals.affected_user || newVals.affected_user || '');
        if (verb === 'create') return name || 'New user';
        if (verb === 'update') return name || 'Updated user';
        if (verb === 'archive') return name || 'Archived user';
        if (verb === 'restore') return name || 'Restored user';
        if (verb === 'delete') return name || 'Deleted user';
    }

    return `${ACTION_LABELS[verb] || verb} ${log.entity_type}`;
}

function ChangesList({ entityType, dateRange, selectedUser }: { entityType: ChangeTab; dateRange: { start: string; end: string }; selectedUser: string }) {
    const [logs, setLogs] = useState<AuditLog[]>([]);
    const [isLoading, setIsLoading] = useState(true);

    // Pagination
    const [page, setPage] = useState(1);
    const [hasMore, setHasMore] = useState(true);
    const [limit, setLimit] = useState(10);
    const [total, setTotal] = useState(0);

    useEffect(() => {
        setPage(1); // Reset page when tab changes
    }, [entityType]);

    useEffect(() => {
        const fetchLogs = async () => {
            setIsLoading(true);
            try {
                const params: Record<string, any> = {
                    entity_type: entityType,
                    limit,
                    offset: (page - 1) * limit,
                };
                if (dateRange.start) params.from = dateRange.start;
                if (dateRange.end) params.to = dateRange.end;
                if (selectedUser && selectedUser !== 'all') params.user_id = selectedUser;

                const result = await api.getDashboardChanges(params);
                if (result.data) {
                    setLogs(result.data);
                    setHasMore(result.data.length === limit);
                    setTotal(result.total || 0);
                }
            } catch (error) {
                console.error('Failed to fetch change logs:', error);
            } finally {
                setIsLoading(false);
            }
        };
        fetchLogs();
    }, [entityType, page, limit, dateRange, selectedUser]);

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

    const SKIP_FIELDS = new Set(['affected_product', 'affected_category', 'affected_expense', 'affected_user', 'product_name', 'invoice_no', 'category_id', 'product_id']);

    const renderFieldChanges = (log: AuditLog) => {
        const oldVals = log.old_values || {};
        const newVals = log.new_values || {};
        const verb = getActionVerb(log.action);

        const allKeys = new Set([...Object.keys(oldVals), ...Object.keys(newVals)]);
        const changes: { key: string; oldVal: unknown; newVal: unknown }[] = [];

        allKeys.forEach((key) => {
            if (SKIP_FIELDS.has(key)) return;
            const oldVal = oldVals[key];
            const newVal = newVals[key];

            if (verb === 'update' || verb === 'close' || verb === 'restore') {
                if (JSON.stringify(oldVal) !== JSON.stringify(newVal)) {
                    changes.push({ key, oldVal, newVal });
                }
            } else if (verb === 'create' || verb === 'start' || verb === 'adjust' || verb === 'count' || verb === 'void') {
                if (newVal !== undefined && newVal !== null && newVal !== '') {
                    changes.push({ key, oldVal: undefined, newVal });
                }
            } else if (verb === 'delete' || verb === 'archive') {
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
                {changes.slice(0, 8).map(({ key, oldVal, newVal }) => (
                    <div key={key} className="text-xs flex items-start gap-1.5 text-muted-foreground">
                        <span className="font-medium text-foreground/70 shrink-0 pt-0.5">{formatFieldName(key)}:</span>
                        {isImageField(key) ? (
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
                {changes.length > 8 && (
                    <span className="text-xs text-muted-foreground">+{changes.length - 8} more</span>
                )}
            </div>
        );
    };

    return (
        <div className="space-y-4">
            <div className="flex flex-col md:flex-row gap-4 items-center justify-between mb-4">
                <div className="flex items-center gap-2">
                    <span className="text-sm text-muted-foreground">Show</span>
                    <Select value={String(limit)} onValueChange={(v) => { setLimit(Number(v)); setPage(1); }}>
                        <SelectTrigger className="w-[80px]">
                            <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                            <SelectItem value="10">10</SelectItem>
                            <SelectItem value="20">20</SelectItem>
                            <SelectItem value="50">50</SelectItem>
                            <SelectItem value="100">100</SelectItem>
                        </SelectContent>
                    </Select>
                    <span className="text-sm text-muted-foreground">entries</span>
                </div>
                {total > 0 && (
                    <span className="text-sm text-muted-foreground">
                        {Math.min(logs.length, limit)} entries of {total}
                    </span>
                )}
            </div>

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
                                    <UserIcon className="h-3.5 w-3.5 text-muted-foreground shrink-0" />
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

                {/* Pagination Controls */}
                <div className="flex items-center justify-between mt-6 pt-4 border-t">
                    <p className="text-sm text-muted-foreground">Page {page}</p>
                    <div className="flex gap-2">
                        <Button
                            variant="outline"
                            size="sm"
                            onClick={() => setPage((p) => Math.max(1, p - 1))}
                            disabled={page === 1}
                        >
                            <ChevronLeft className="h-4 w-4 mr-1" />
                            Previous
                        </Button>
                        <Button
                            variant="outline"
                            size="sm"
                            onClick={() => setPage((p) => p + 1)}
                            disabled={!hasMore}
                        >
                            Next
                            <ChevronRight className="h-4 w-4 ml-1" />
                        </Button>
                    </div>
                </div>
            </div>
        </div>
    );
}

export default function ChangesHistoryPage() {
    const [dateRange, setDateRange] = useState<{ start: string; end: string }>({ start: '', end: '' });
    const [selectedUser, setSelectedUser] = useState<string>('all');
    const [users, setUsers] = useState<User[]>([]);

    useEffect(() => {
        const fetchUsers = async () => {
            const result = await api.getUsers();
            if (result.data) {
                setUsers(result.data);
            }
        };
        fetchUsers();
    }, []);

    return (
        <div className="flex flex-col h-full">
            <Header title="Changes" />

            <div className="flex-1 p-6 overflow-auto">
                <div className="mx-auto w-full">

                    <Card className="mb-6">
                        <CardHeader>
                            <CardTitle className="text-base">Filters</CardTitle>
                        </CardHeader>
                        <CardContent>
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                <div className="space-y-1.5">
                                    <label className="text-xs font-medium text-muted-foreground">Date Range</label>
                                    <DateRangePicker
                                        value={dateRange}
                                        onChange={(newRange) => {
                                            setDateRange(newRange);
                                        }}
                                        placeholder="Filter by date..."
                                        className="w-full"
                                    />
                                </div>
                                <div className="space-y-1.5">
                                    <label className="text-xs font-medium text-muted-foreground">Employee</label>
                                    <Select value={selectedUser} onValueChange={(v) => setSelectedUser(v)}>
                                        <SelectTrigger className="w-full">
                                            <SelectValue placeholder="All Employees" />
                                        </SelectTrigger>
                                        <SelectContent>
                                            <SelectItem value="all">All Employees</SelectItem>
                                            {users.map(user => (
                                                <SelectItem key={user.id} value={user.id}>{user.name}</SelectItem>
                                            ))}
                                        </SelectContent>
                                    </Select>
                                </div>
                            </div>
                        </CardContent>
                    </Card>

                    <Card className="flex flex-col border-0 shadow-none bg-transparent md:border md:shadow md:bg-card">
                        <CardContent className="flex-1 px-0 py-0 md:p-6">
                            <Tabs defaultValue="product" className="w-full">
                                <TabsList className="w-full justify-start flex overflow-x-auto lg:w-auto lg:inline-flex mb-6 no-scrollbar pb-1">
                                    {CHANGE_TABS.map((tab) => (
                                        <TabsTrigger key={tab.value} value={tab.value} className="gap-1.5 px-4">
                                            {tab.icon}
                                            {tab.label}
                                        </TabsTrigger>
                                    ))}
                                </TabsList>

                                {CHANGE_TABS.map((tab) => (
                                    <TabsContent key={tab.value} value={tab.value}>
                                        <ChangesList entityType={tab.value} dateRange={dateRange} selectedUser={selectedUser} />
                                    </TabsContent>
                                ))}
                            </Tabs>
                        </CardContent>
                    </Card>
                </div>
            </div>
        </div>
    );
}

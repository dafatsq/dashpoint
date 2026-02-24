'use client';

import { useEffect, useState, useCallback } from 'react';
import { Header } from '@/components/layout/header';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import {
    Clock,
    Loader2,
    CircleDot,
    CheckCircle2,
    User as UserIcon,
    ArrowRightLeft,
    ChevronLeft,
    ChevronRight,
} from 'lucide-react';
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import { DateRangePicker } from '@/components/ui/date-range-picker';
import api from '@/lib/api';
import { Shift, CashDrawerOperation, User } from '@/types';

export default function ShiftsHistoryPage() {
    const [shifts, setShifts] = useState<Shift[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const [fetchError, setFetchError] = useState(false);

    // Pagination
    const [page, setPage] = useState(1);
    const [hasMore, setHasMore] = useState(true);
    const [limit, setLimit] = useState(10);
    const [total, setTotal] = useState(0);

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

    const fetchShifts = useCallback(async () => {
        setIsLoading(true);
        setFetchError(false);
        try {
            const params: Record<string, any> = {
                limit,
                offset: (page - 1) * limit,
            };
            if (dateRange.start) params.from = dateRange.start;
            if (dateRange.end) params.to = dateRange.end;
            if (selectedUser && selectedUser !== 'all') params.user_id = selectedUser;

            const result = await api.getShifts(params);

            if (result.data) {
                setShifts(result.data);
                setHasMore(result.data.length === limit);
                setTotal(result.total || 0);
            } else if (result.error) {
                setFetchError(true);
            }
        } catch (error) {
            console.error('Failed to fetch shifts:', error);
            setFetchError(true);
        } finally {
            setIsLoading(false);
        }
    }, [page, limit, dateRange, selectedUser]);

    useEffect(() => {
        fetchShifts();
    }, [fetchShifts]);

    const formatDateTime = (dateString: string) => {
        return new Date(dateString).toLocaleString('id-ID', {
            day: '2-digit',
            month: 'short',
            hour: '2-digit',
            minute: '2-digit',
        });
    };

    const formatCurrencyShort = (amount: number) => {
        return new Intl.NumberFormat('id-ID', {
            style: 'currency',
            currency: 'IDR',
            minimumFractionDigits: 0,
            maximumFractionDigits: 0,
        }).format(amount);
    };

    const renderOperations = (operations: CashDrawerOperation[]) => {
        if (!operations || operations.length === 0) return null;

        return (
            <div className="mt-4 pt-3 border-t border-dashed space-y-2">
                <p className="text-[10px] font-semibold text-muted-foreground uppercase tracking-wider mb-2">
                    Cash Drawer Log
                </p>
                <div className="grid gap-2">
                    {operations.map((op) => (
                        <div key={op.id} className="text-xs flex items-start gap-2 bg-muted/30 p-2 rounded">
                            <span
                                className={`font-semibold shrink-0 ${op.type === 'pay_in' ? 'text-green-600' : 'text-red-600'
                                    }`}
                            >
                                {op.type === 'pay_in' ? '+ PAY IN' : '- PAY OUT'}
                            </span>
                            <div className="flex-1 min-w-0">
                                <span className="font-medium text-foreground">{op.reason}</span>
                                <div className="flex items-center gap-1.5 text-[10px] text-muted-foreground mt-0.5">
                                    <span>{op.performed_by_name || 'User'}</span>
                                    <span>·</span>
                                    <span>{formatDateTime(op.created_at)}</span>
                                </div>
                            </div>
                            <span className="font-medium whitespace-nowrap">
                                {formatCurrencyShort(parseFloat(op.amount))}
                            </span>
                        </div>
                    ))}
                </div>
            </div>
        );
    };

    return (
        <div className="flex flex-col h-full">
            <Header title="Shifts History" />

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
                                            setPage(1);
                                        }}
                                        placeholder="Filter by date..."
                                        className="w-full"
                                    />
                                </div>
                                <div className="space-y-1.5">
                                    <label className="text-xs font-medium text-muted-foreground">Employee</label>
                                    <Select value={selectedUser} onValueChange={(v) => { setSelectedUser(v); setPage(1); }}>
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
                                        {Math.min(shifts.length, limit)} entries of {total}
                                    </span>
                                )}
                            </div>

                            {isLoading ? (
                                <div className="flex items-center justify-center py-12">
                                    <Loader2 className="h-6 w-6 animate-spin text-primary" />
                                </div>
                            ) : fetchError ? (
                                <div className="flex flex-col items-center justify-center py-12 text-muted-foreground gap-2">
                                    <Clock className="h-10 w-10 mb-2" />
                                    <p className="text-sm">Could not load shifts</p>
                                    <Button
                                        variant="outline"
                                        size="sm"
                                        onClick={fetchShifts}
                                        className="mt-2"
                                    >
                                        Retry
                                    </Button>
                                </div>
                            ) : shifts.length === 0 ? (
                                <Card>
                                    <CardContent className="flex flex-col items-center justify-center py-12 text-muted-foreground">
                                        <Clock className="h-10 w-10 mb-2" />
                                        <p className="text-sm">No shifts recorded yet</p>
                                    </CardContent>
                                </Card>
                            ) : (
                                <div className="space-y-4">
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
                                                            <UserIcon className="h-3.5 w-3.5 text-muted-foreground flex-shrink-0" />
                                                            <span className="truncate max-w-[120px] sm:max-w-[200px]">
                                                                {shift.employee_name || 'Unknown'}
                                                            </span>
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
                                                                    <span
                                                                        className={`text-xs ml-1 font-medium ${parseFloat(shift.cash_difference) > 0
                                                                            ? 'text-green-600'
                                                                            : parseFloat(shift.cash_difference) < 0
                                                                                ? 'text-red-600'
                                                                                : 'text-muted-foreground'
                                                                            }`}
                                                                    >
                                                                        ({parseFloat(shift.cash_difference) > 0 ? '+' : ''}
                                                                        {formatCurrencyShort(parseFloat(shift.cash_difference))})
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
                                                {shift.operations && shift.operations.length > 0 && renderOperations(shift.operations)}
                                            </div>
                                        );
                                    })}

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
                            )}
                        </CardContent>
                    </Card>
                </div>
            </div>
        </div>
    );
}

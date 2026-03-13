'use client';

import { useEffect, useState } from 'react';
import { Header } from '@/components/layout/header';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { DateRangePicker } from '@/components/ui/date-range-picker';
import {
  Search,
  Loader2,
  ScrollText,
  Eye,
  ChevronLeft,
  ChevronRight,
} from 'lucide-react';
import api from '@/lib/api';
import { AuditLog } from '@/types';

const ACTION_LABELS: Record<string, string> = {
  'login': 'Login',
  'login_failed': 'Login Failed',
  'logout': 'Logout',
  'create': 'Create',
  'update': 'Update',
  'delete': 'Delete',
  'void': 'Void',
  'adjust': 'Adjust',
  'auth.login': 'Login',
  'auth.login_failed': 'Login Failed',
  'auth.logout': 'Logout',
  'auth.pin_login': 'PIN Login',
  'user.create': 'Create User',
  'user.update': 'Update User',
  'user.delete': 'Delete User',
  'user.archive': 'Archive User',
  'user.restore': 'Restore User',
  'product.create': 'Create Product',
  'product.update': 'Update Product',
  'product.delete': 'Delete Product',
  'product.archive': 'Archive Product',
  'product.restore': 'Restore Product',
  'inventory.adjust': 'Adjust Stock',
  'inventory.count': 'Stock Count',
  'category.create': 'Create Category',
  'category.update': 'Update Category',
  'category.delete': 'Delete Category',
  'category.archive': 'Archive Category',
  'category.restore': 'Restore Category',
  'expense.create': 'Create Expense',
  'expense.update': 'Update Expense',
  'expense.delete': 'Delete Expense',
  'expense.archive': 'Archive Expense',
  'expense.restore': 'Restore Expense',
  'sale.create': 'Sale',
  'sale.void': 'Void Sale',
  'sale.refund': 'Refund',
  'shift.start': 'Start Shift',
  'shift.close': 'Close Shift',
};

const ENTITY_LABELS: Record<string, string> = {
  'user': 'User',
  'product': 'Product',
  'category': 'Category',
  'sale': 'Sale',
  'shift': 'Shift',
  'inventory': 'Inventory',
  'auth': 'Auth',
  'expense': 'Expense',
};

export default function AuditLogsPage() {
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedAction, setSelectedAction] = useState<string>('all');
  const [selectedEntity, setSelectedEntity] = useState<string>('all');
  const [dateRange, setDateRange] = useState<{ start: string; end: string }>({ start: '', end: '' });

  // Pagination
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);
  const limit = 50;

  // Detail dialog
  const [selectedLog, setSelectedLog] = useState<AuditLog | null>(null);
  const [detailDialogOpen, setDetailDialogOpen] = useState(false);

  // Fetch logs
  useEffect(() => {
    const fetchLogs = async () => {
      setIsLoading(true);
      try {
        const params: {
          action?: string;
          entity_type?: string;
          from?: string;
          to?: string;
          limit: number;
          offset: number;
        } = {
          limit,
          offset: (page - 1) * limit,
        };

        if (selectedAction !== 'all') params.action = selectedAction;
        if (selectedEntity !== 'all') params.entity_type = selectedEntity;
        if (dateRange.start) params.from = dateRange.start;
        if (dateRange.end) params.to = dateRange.end;

        const result = await api.getAuditLogs(params);
        if (result.data) {
          setLogs(result.data);
          setHasMore(result.data.length === limit);
        }
      } catch (error) {
        console.error('Failed to fetch logs:', error);
      } finally {
        setIsLoading(false);
      }
    };

    fetchLogs();
  }, [page, selectedAction, selectedEntity, dateRange]);

  // Filter logs by search
  const filteredLogs = logs.filter((log) => {
    const matchesSearch =
      log.user_name?.toLowerCase().includes(searchQuery.toLowerCase()) ||
      log.action?.toLowerCase().includes(searchQuery.toLowerCase()) ||
      log.entity_type?.toLowerCase().includes(searchQuery.toLowerCase());
    return matchesSearch;
  });

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString('id-ID', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
    });
  };

  const getActionBadgeColor = (action: string) => {
    if (action.includes('login')) {
      return action.includes('failed')
        ? 'bg-red-600 text-white dark:bg-red-600/90 dark:text-white'
        : 'bg-blue-600 text-white dark:bg-blue-600/90 dark:text-white';
    }
    switch (action) {
      case 'create':
        return 'bg-green-600 text-white dark:bg-green-600/90 dark:text-white';
      case 'update':
        return 'bg-yellow-600 text-white dark:bg-yellow-600/90 dark:text-white';
      case 'delete':
        return 'bg-red-600 text-white dark:bg-red-600/90 dark:text-white';
      case 'void':
        return 'bg-orange-600 text-white dark:bg-orange-600/90 dark:text-white';
    }
    if (action.includes('archive')) {
      return 'bg-orange-500 text-white dark:bg-orange-500/90 dark:text-white';
    }
    if (action.includes('restore')) {
      return 'bg-emerald-600 text-white dark:bg-emerald-600/90 dark:text-white';
    }
    return 'bg-gray-600 text-white dark:bg-gray-600/90 dark:text-white';
  };

  const openDetailDialog = (log: AuditLog) => {
    setSelectedLog(log);
    setDetailDialogOpen(true);
  };

  const renderChanges = (oldValues: Record<string, unknown>, newValues: Record<string, unknown>) => {
    const allKeys = new Set([...Object.keys(oldValues || {}), ...Object.keys(newValues || {})]);

    if (allKeys.size === 0) {
      return <p className="text-muted-foreground text-sm">No detailed changes recorded</p>;
    }

    // Fields to always show at the top (even if unchanged) as context
    const contextFields = ['affected_user', 'affected_product', 'affected_category', 'affected_expense'];
    const contextItems: { key: string; value: unknown }[] = [];

    contextFields.forEach(field => {
      if (oldValues?.[field] || newValues?.[field]) {
        contextItems.push({ key: field, value: oldValues?.[field] || newValues?.[field] });
      }
    });

    return (
      <div className="space-y-2">
        {/* Show context fields (affected entity) at the top */}
        {contextItems.length > 0 && (
          <div className="bg-muted/50 rounded-md p-2 mb-2">
            {contextItems.map(({ key, value }) => (
              <div key={key} className="text-sm">
                <span className="font-medium capitalize">{key.replace(/_/g, ' ')}: </span>
                <span className="text-blue-600 dark:text-blue-400 font-semibold">
                  {typeof value === 'object' ? JSON.stringify(value) : String(value)}
                </span>
              </div>
            ))}
          </div>
        )}

        {/* Show changed fields */}
        {Array.from(allKeys).map((key) => {
          // Skip context fields as they're shown above
          if (contextFields.includes(key)) return null;

          const oldVal = oldValues?.[key];
          const newVal = newValues?.[key];

          // Skip if both are the same
          if (JSON.stringify(oldVal) === JSON.stringify(newVal)) return null;

          return (
            <div key={key} className="text-sm border-b pb-2">
              <span className="font-medium capitalize">{key.replace(/_/g, ' ')}:</span>
              <div className="ml-4 grid grid-cols-2 gap-2 mt-1">
                {oldVal !== undefined && (
                  <div className="text-red-600 dark:text-red-400">
                    <span className="text-xs text-muted-foreground">Old: </span>
                    {typeof oldVal === 'object' ? JSON.stringify(oldVal) : String(oldVal)}
                  </div>
                )}
                {newVal !== undefined && (
                  <div className="text-green-600 dark:text-green-400">
                    <span className="text-xs text-muted-foreground">New: </span>
                    {typeof newVal === 'object' ? JSON.stringify(newVal) : String(newVal)}
                  </div>
                )}
              </div>
            </div>
          );
        })}
      </div>
    );
  };

  return (
    <div className="flex flex-col h-full">
      <Header title="Audit" />

      <div className="flex-1 p-6 overflow-auto">
        {/* Toolbar */}
        {/* Toolbar */}
        <Card className="mb-6">
          <CardContent className="p-4">
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                <Input
                  placeholder="Search logs..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="pl-9 w-full"
                />
              </div>
              <DateRangePicker
                value={dateRange}
                onChange={(newRange) => {
                  setDateRange(newRange);
                  setPage(1);
                }}
                placeholder="Filter by date..."
                className="w-full"
              />
              <Select value={selectedAction} onValueChange={(v) => { setSelectedAction(v); setPage(1); }}>
                <SelectTrigger className="w-full">
                  <SelectValue placeholder="All Actions" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Actions</SelectItem>
                  <SelectItem value="login">Login</SelectItem>
                  <SelectItem value="login_failed">Login Failed</SelectItem>
                  <SelectItem value="create">Create</SelectItem>
                  <SelectItem value="update">Update</SelectItem>
                  <SelectItem value="archive">Archive</SelectItem>
                  <SelectItem value="restore">Restore</SelectItem>
                  <SelectItem value="delete">Delete</SelectItem>
                  <SelectItem value="void">Void</SelectItem>
                </SelectContent>
              </Select>
              <Select value={selectedEntity} onValueChange={(v) => { setSelectedEntity(v); setPage(1); }}>
                <SelectTrigger className="w-full">
                  <SelectValue placeholder="All Entities" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Entities</SelectItem>
                  <SelectItem value="auth">Authentication</SelectItem>
                  <SelectItem value="user">User</SelectItem>
                  <SelectItem value="product">Product</SelectItem>
                  <SelectItem value="category">Category</SelectItem>
                  <SelectItem value="sale">Sale</SelectItem>
                  <SelectItem value="shift">Shift</SelectItem>
                  <SelectItem value="inventory">Inventory</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </CardContent>
        </Card>

        {/* Logs table */}
        {isLoading ? (
          <div className="flex items-center justify-center h-64">
            <Loader2 className="h-8 w-8 animate-spin text-primary" />
          </div>
        ) : filteredLogs.length === 0 ? (
          <Card>
            <CardContent className="flex flex-col items-center justify-center py-12">
              <ScrollText className="h-12 w-12 text-muted-foreground mb-4" />
              <p className="text-muted-foreground">No audit logs found</p>
            </CardContent>
          </Card>
        ) : (
          <>
            {/* Desktop View */}
            <Card className="hidden lg:block">
              <CardHeader>
                <CardTitle>Activity Log</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="overflow-x-auto">
                  <table className="w-full">
                    <thead>
                      <tr className="border-b text-left text-sm text-muted-foreground">
                        <th className="pb-3 font-medium">Timestamp</th>
                        <th className="pb-3 font-medium">User</th>
                        <th className="pb-3 font-medium">Action</th>
                        <th className="pb-3 font-medium">Entity</th>
                        <th className="pb-3 font-medium">IP Address</th>
                        <th className="pb-3 font-medium text-right">Details</th>
                      </tr>
                    </thead>
                    <tbody>
                      {filteredLogs.map((log) => (
                        <tr key={log.id} className="border-b last:border-0">
                          <td className="py-3 text-sm">
                            {formatDate(log.created_at)}
                          </td>
                          <td className="py-3">
                            <p className="font-medium text-sm">{log.user_name || 'System'}</p>
                          </td>
                          <td className="py-3">
                            <span
                              className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${getActionBadgeColor(
                                log.action
                              )}`}
                            >
                              {ACTION_LABELS[log.action] || log.action}
                            </span>
                          </td>
                          <td className="py-3 text-sm capitalize">
                            {ENTITY_LABELS[log.entity_type] || log.entity_type}
                            {log.entity_id && (
                              <span className="text-xs text-muted-foreground ml-1">
                                ({log.entity_id.slice(0, 8)}...)
                              </span>
                            )}
                          </td>
                          <td className="py-3 text-sm text-muted-foreground">
                            {log.ip_address || '-'}
                          </td>
                          <td className="py-3 text-right">
                            <Button
                              variant="ghost"
                              size="icon"
                              onClick={() => openDetailDialog(log)}
                            >
                              <Eye className="h-4 w-4" />
                            </Button>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>

                {/* Pagination */}
                <div className="flex items-center justify-between mt-4 pt-4 border-t">
                  <p className="text-sm text-muted-foreground">
                    Page {page}
                  </p>
                  <div className="flex gap-2">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setPage((p) => Math.max(1, p - 1))}
                      disabled={page === 1}
                    >
                      <ChevronLeft className="h-4 w-4" />
                      Previous
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setPage((p) => p + 1)}
                      disabled={!hasMore}
                    >
                      Next
                      <ChevronRight className="h-4 w-4" />
                    </Button>
                  </div>
                </div>
              </CardContent>
            </Card>

            {/* Mobile View */}
            <div className="lg:hidden space-y-4">
              {filteredLogs.map((log) => (
                <Card key={log.id}>
                  <CardContent className="p-4">
                    <div className="flex items-center justify-between mb-3">
                      <div className="flex items-center gap-2">
                        <span
                          className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${getActionBadgeColor(
                            log.action
                          )}`}
                        >
                          {ACTION_LABELS[log.action] || log.action}
                        </span>
                        <span className="text-xs text-muted-foreground">
                          {formatDate(log.created_at)}
                        </span>
                      </div>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8"
                        onClick={() => openDetailDialog(log)}
                      >
                        <Eye className="h-4 w-4" />
                      </Button>
                    </div>

                    <div className="grid gap-1">
                      <div className="flex justify-between text-sm">
                        <span className="text-muted-foreground">User:</span>
                        <span className="font-medium">{log.user_name || 'System'}</span>
                      </div>
                      <div className="flex justify-between text-sm">
                        <span className="text-muted-foreground">Entity:</span>
                        <span className="font-medium capitalize">
                          {ENTITY_LABELS[log.entity_type] || log.entity_type}
                        </span>
                      </div>
                      {log.entity_id && (
                        <div className="flex justify-between text-sm">
                          <span className="text-muted-foreground">ID:</span>
                          <span className="font-mono text-xs">{log.entity_id.slice(0, 8)}...</span>
                        </div>
                      )}
                    </div>
                  </CardContent>
                </Card>
              ))}

              {/* Pagination for mobile */}
              <div className="flex items-center justify-between pt-4">
                <p className="text-sm text-muted-foreground">
                  Page {page}
                </p>
                <div className="flex gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setPage((p) => Math.max(1, p - 1))}
                    disabled={page === 1}
                  >
                    <ChevronLeft className="h-4 w-4" />
                    Previous
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setPage((p) => p + 1)}
                    disabled={!hasMore}
                  >
                    Next
                    <ChevronRight className="h-4 w-4" />
                  </Button>
                </div>
              </div>
            </div>
          </>
        )}
      </div>

      {/* Detail Dialog */}
      <Dialog open={detailDialogOpen} onOpenChange={setDetailDialogOpen}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>Log Details</DialogTitle>
            <DialogDescription>
              {selectedLog && formatDate(selectedLog.created_at)}
            </DialogDescription>
          </DialogHeader>

          {selectedLog && (
            <div className="space-y-4">
              <div className="grid grid-cols-2 gap-4 text-sm">
                <div>
                  <span className="text-muted-foreground">User:</span>
                  <p className="font-medium">{selectedLog.user_name || 'System'}</p>
                </div>
                <div>
                  <span className="text-muted-foreground">Action:</span>
                  <p className="font-medium capitalize">{selectedLog.action}</p>
                </div>
                <div>
                  <span className="text-muted-foreground">Entity Type:</span>
                  <p className="font-medium capitalize">{selectedLog.entity_type}</p>
                </div>
                <div>
                  <span className="text-muted-foreground">Entity ID:</span>
                  <p className="font-medium font-mono text-xs">{selectedLog.entity_id || '-'}</p>
                </div>
                <div>
                  <span className="text-muted-foreground">IP Address:</span>
                  <p className="font-medium">{selectedLog.ip_address || '-'}</p>
                </div>
                <div>
                  <span className="text-muted-foreground">User Agent:</span>
                  <p className="font-medium text-xs truncate" title={selectedLog.user_agent}>
                    {selectedLog.user_agent || '-'}
                  </p>
                </div>
              </div>

              <div className="border-t pt-4">
                <h4 className="font-medium mb-2">Changes</h4>
                {renderChanges(selectedLog.old_values, selectedLog.new_values)}
              </div>
            </div>
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}

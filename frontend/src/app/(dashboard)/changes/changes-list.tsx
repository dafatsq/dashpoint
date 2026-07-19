'use client';

import { useEffect, useState } from "react";
import { ChevronLeft, ChevronRight, History, Loader2, User as UserIcon } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import api from "@/lib/api";
import type { AuditLog } from "@/types";

import { ActivityFieldChanges } from "../activity-field-changes";
import {
  type ActivityChangeTab,
  type ActivityDateRange,
  buildActivityDescription,
  buildDashboardChangeParams,
  formatActivityDate,
  getActivityActionLabel,
  getActivityActionVerb,
  getActivityBadgeColor,
  incrementActivityRefreshKey,
} from "../activity-helpers";

interface ChangesListProps {
  entityType: ActivityChangeTab;
  dateRange: ActivityDateRange;
  selectedUser: string;
}

export function ChangesList({ entityType, dateRange, selectedUser }: ChangesListProps) {
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [fetchError, setFetchError] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);
  const [limit, setLimit] = useState(10);
  const [total, setTotal] = useState(0);
  const [refreshKey, setRefreshKey] = useState(0);

  useEffect(() => {
    const timer = window.setTimeout(async () => {
      setIsLoading(true);
      setFetchError(null);

      const result = await api.getDashboardChanges(
        buildDashboardChangeParams({
          entityType,
          page,
          limit,
          dateRange,
          selectedUser,
        }),
      );

      if (result.error) {
        setFetchError(result.error);
      } else {
        setLogs(result.data || []);
        setHasMore((result.data || []).length === limit);
        setTotal(result.total || 0);
      }

      setIsLoading(false);
    }, 0);

    return () => window.clearTimeout(timer);
  }, [dateRange, entityType, limit, page, refreshKey, selectedUser]);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="h-6 w-6 animate-spin text-primary" />
      </div>
    );
  }

  if (fetchError) {
    return (
      <div className="flex flex-col items-center justify-center gap-3 py-12 text-muted-foreground">
        <History className="h-10 w-10" />
        <p className="text-sm">{fetchError}</p>
        <Button variant="outline" size="sm" onClick={() => setRefreshKey(incrementActivityRefreshKey)}>
          Retry
        </Button>
      </div>
    );
  }

  if (logs.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-12 text-muted-foreground">
        <History className="mb-2 h-10 w-10" />
        <p className="text-sm">No changes recorded yet</p>
      </div>
    );
  }

  return (
    <div>
      <div className="mb-4 flex flex-row items-center justify-between gap-4">
        <div className="flex items-center gap-2">
          <span className="text-sm text-muted-foreground">Show</span>
          <Select value={String(limit)} onValueChange={(value) => { setLimit(Number(value)); setPage(1); }}>
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
        {total > 0 ? (
          <span className="text-sm text-muted-foreground">
            {Math.min(logs.length, limit)} entries of {total}
          </span>
        ) : null}
      </div>

      <div className="space-y-4">
        {logs.map((log) => (
          <div key={log.id} className="rounded-xl border bg-card p-4 text-card-foreground shadow-sm transition-shadow hover:shadow-md">
            <div className="mb-3 flex flex-col justify-between gap-2 sm:flex-row sm:items-center">
              <div className="flex flex-wrap items-center gap-2">
                <span className={`inline-flex whitespace-nowrap rounded-full px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wide ${getActivityBadgeColor(log.action)}`}>
                  {getActivityActionLabel(log.action) || getActivityActionVerb(log.action)}
                </span>
                <div className="flex items-center gap-1.5 text-sm font-medium">
                  <UserIcon className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
                  <span className="break-words">{log.user_name || "System"}</span>
                </div>
              </div>
              <span className="whitespace-nowrap text-xs text-muted-foreground">{formatActivityDate(log.created_at)}</span>
            </div>

            <div>
              <p className="mb-3 text-base font-medium leading-snug">{buildActivityDescription(log)}</p>
              <ActivityFieldChanges log={log} />
            </div>
          </div>
        ))}

        <div className="mt-6 flex items-center justify-between border-t pt-4">
          <p className="text-sm text-muted-foreground">Page {page}</p>
          <div className="flex gap-2">
            <Button variant="outline" size="sm" onClick={() => setPage((current) => Math.max(1, current - 1))} disabled={page === 1}>
              <ChevronLeft className="mr-1 h-4 w-4" />
              Previous
            </Button>
            <Button variant="outline" size="sm" onClick={() => setPage((current) => current + 1)} disabled={!hasMore}>
              Next
              <ChevronRight className="ml-1 h-4 w-4" />
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}

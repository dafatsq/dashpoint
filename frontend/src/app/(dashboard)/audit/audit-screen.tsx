'use client';

import { useEffect, useMemo, useState } from "react";

import { Header } from "@/components/layout/header";
import api from "@/lib/api";
import type { AuditLog } from "@/types";

import {
  type ActivityDateRange,
  buildAuditLogParams,
  incrementActivityRefreshKey,
} from "../activity-helpers";
import { AuditDetailDialog } from "./audit-detail-dialog";
import { AuditFilters } from "./audit-filters";
import { AuditList } from "./audit-list";

export function AuditScreen() {
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [fetchError, setFetchError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedAction, setSelectedAction] = useState("all");
  const [selectedEntity, setSelectedEntity] = useState("all");
  const [dateRange, setDateRange] = useState<ActivityDateRange>({ start: "", end: "" });
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);
  const [refreshKey, setRefreshKey] = useState(0);
  const [selectedLog, setSelectedLog] = useState<AuditLog | null>(null);
  const [detailDialogOpen, setDetailDialogOpen] = useState(false);

  const limit = 50;

  useEffect(() => {
    const timer = window.setTimeout(async () => {
      setIsLoading(true);
      setFetchError(null);

      const result = await api.getAuditLogs(
        buildAuditLogParams({
          page,
          limit,
          selectedAction,
          selectedEntity,
          dateRange,
        }),
      );

      if (result.error) {
        setFetchError(result.error);
      } else {
        setLogs(result.data || []);
        setHasMore((result.data || []).length === limit);
      }

      setIsLoading(false);
    }, 0);

    return () => window.clearTimeout(timer);
  }, [dateRange, page, refreshKey, selectedAction, selectedEntity]);

  const filteredLogs = useMemo(() => {
    const normalizedSearch = searchQuery.trim().toLowerCase();
    if (!normalizedSearch) return logs;

    return logs.filter((log) => {
      return (
        log.user_name?.toLowerCase().includes(normalizedSearch) ||
        log.action?.toLowerCase().includes(normalizedSearch) ||
        log.entity_type?.toLowerCase().includes(normalizedSearch)
      );
    });
  }, [logs, searchQuery]);

  return (
    <div className="flex h-full flex-col">
      <Header title="Audit" />

      <div className="flex-1 overflow-auto p-6">
        <AuditFilters
          searchQuery={searchQuery}
          selectedAction={selectedAction}
          selectedEntity={selectedEntity}
          dateRange={dateRange}
          onSearchChange={setSearchQuery}
          onActionChange={(value) => {
            setSelectedAction(value);
            setPage(1);
          }}
          onEntityChange={(value) => {
            setSelectedEntity(value);
            setPage(1);
          }}
          onDateRangeChange={(value) => {
            setDateRange(value);
            setPage(1);
          }}
        />

        <AuditList
          logs={filteredLogs}
          isLoading={isLoading}
          error={fetchError}
          page={page}
          hasMore={hasMore}
          onRetry={() => setRefreshKey(incrementActivityRefreshKey)}
          onPageChange={setPage}
          onViewLog={(log) => {
            setSelectedLog(log);
            setDetailDialogOpen(true);
          }}
        />
      </div>

      <AuditDetailDialog
        open={detailDialogOpen}
        log={selectedLog}
        onOpenChange={setDetailDialogOpen}
      />
    </div>
  );
}

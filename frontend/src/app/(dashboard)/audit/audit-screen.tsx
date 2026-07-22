"use client";

import { useCallback, useEffect, useState } from "react";

import { Header } from "@/components/layout/header";
import { DataTableContainer } from "@/components/shared/data-table-container";
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
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [selectedAction, setSelectedAction] = useState("all");
  const [selectedEntity, setSelectedEntity] = useState("all");
  const [sort, setSort] = useState("date_desc");
  const [dateRange, setDateRange] = useState<ActivityDateRange>({
    start: "",
    end: "",
  });
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);
  const [total, setTotal] = useState(0);
  const [limit, setLimit] = useState(50);
  const [refreshKey, setRefreshKey] = useState(0);
  const [selectedLog, setSelectedLog] = useState<AuditLog | null>(null);
  const [detailDialogOpen, setDetailDialogOpen] = useState(false);

  const resetToFirstPage = useCallback(() => {
    setPage(1);
  }, []);

  const loadAuditLogs = useCallback(async () => {
    setIsLoading(true);
    setFetchError(null);

    const result = await api.getAuditLogs(
      buildAuditLogParams({
        page,
        limit,
        selectedAction,
        selectedEntity,
        dateRange,
        searchQuery: debouncedSearch,
        sort,
      }),
    );

    if (result.error) {
      setFetchError(result.error);
    } else {
      const nextLogs = result.data || [];
      setLogs(nextLogs);
      setHasMore(nextLogs.length === limit);
      setTotal(result.total || 0);
    }

    setIsLoading(false);
  }, [dateRange, debouncedSearch, limit, page, selectedAction, selectedEntity, sort]);

  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(searchQuery);
      resetToFirstPage();
    }, 300);
    return () => clearTimeout(timer);
  }, [resetToFirstPage, searchQuery]);

  useEffect(() => {
    const timer = window.setTimeout(() => {
      void loadAuditLogs();
    }, 0);

    return () => window.clearTimeout(timer);
  }, [loadAuditLogs, refreshKey]);

  return (
    <div className="flex h-full flex-col">
      <Header title="Audit" />

      <div className="flex-1 overflow-auto p-6">
        <AuditFilters
          searchQuery={searchQuery}
          selectedAction={selectedAction}
          selectedEntity={selectedEntity}
          dateRange={dateRange}
          sort={sort}
          onSearchChange={setSearchQuery}
          onActionChange={(value) => {
            setSelectedAction(value);
            resetToFirstPage();
          }}
          onEntityChange={(value) => {
            setSelectedEntity(value);
            resetToFirstPage();
          }}
          onDateRangeChange={(value) => {
            setDateRange(value);
            resetToFirstPage();
          }}
          onSortChange={(value) => {
            setSort(value);
            resetToFirstPage();
          }}
        />

        <DataTableContainer
          limit={limit}
          onLimitChange={(value) => {
            setLimit(value);
            resetToFirstPage();
          }}
          total={total}
          currentCount={logs.length}
        >
          <AuditList
            logs={logs}
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
        </DataTableContainer>
      </div>

      <AuditDetailDialog
        open={detailDialogOpen}
        log={selectedLog}
        onOpenChange={setDetailDialogOpen}
      />
    </div>
  );
}

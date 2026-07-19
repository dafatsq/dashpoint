"use client";

import { useCallback, useEffect, useState } from "react";

import { Header } from "@/components/layout/header";
import { DataTableContainer } from "@/components/shared/data-table-container";
import api from "@/lib/api";
import type { Shift } from "@/types";

import { buildShiftQueryParams } from "./shifts-helpers";
import { ShiftsFilters } from "./shifts-filters";
import { ShiftsList } from "./shifts-list";

export function ShiftsScreen() {
  const [shifts, setShifts] = useState<Shift[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [fetchError, setFetchError] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);
  const [limit, setLimit] = useState(10);
  const [total, setTotal] = useState(0);
  const [dateRange, setDateRange] = useState<{ start: string; end: string }>({
    start: "",
    end: "",
  });
  const [selectedUser, setSelectedUser] = useState<string>("all");
  const [users, setUsers] = useState<{ id: string; name: string }[]>([]);

  const resetToFirstPage = useCallback(() => {
    setPage(1);
  }, []);

  const loadUsers = useCallback(async () => {
    const result = await api.getBasicUsers();
    if (!result.error && result.data) {
      setUsers([...result.data].sort((a, b) => a.name.localeCompare(b.name)));
    }
  }, []);

  // Fetch active employees for filter (no permission gate, active users only)
  useEffect(() => {
    const timer = window.setTimeout(() => {
      void loadUsers();
    }, 0);

    return () => window.clearTimeout(timer);
  }, [loadUsers]);

  const fetchShifts = useCallback(async () => {
    setIsLoading(true);
    setFetchError(null);

    const result = await api.getShifts(
      buildShiftQueryParams({
        page,
        limit,
        dateRange,
        selectedUser,
      }),
    );

    if (result.error) {
      setFetchError("Could not load shifts");
    } else {
      setShifts(result.data || []);
      setHasMore((result.data || []).length === limit);
      setTotal(result.total || 0);
    }

    setIsLoading(false);
  }, [dateRange, limit, page, selectedUser]);

  useEffect(() => {
    const timer = window.setTimeout(() => {
      void fetchShifts();
    }, 0);

    return () => window.clearTimeout(timer);
  }, [fetchShifts]);

  return (
    <div className="flex flex-col h-full">
      <Header title="Shifts History" />

      <div className="flex-1 p-6 overflow-auto">
        <div className="mx-auto w-full">
          <ShiftsFilters
            dateRange={dateRange}
            selectedUser={selectedUser}
            users={users}
            onDateRangeChange={(newRange) => {
              setDateRange(newRange);
              resetToFirstPage();
            }}
            onSelectedUserChange={(value) => {
              setSelectedUser(value);
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
            currentCount={shifts.length}
          >
            <ShiftsList
              shifts={shifts}
              isLoading={isLoading}
              error={fetchError}
              page={page}
              limit={limit}
              total={total}
              hasMore={hasMore}
              onRetry={() => void fetchShifts()}
              onPageChange={setPage}
            />
          </DataTableContainer>
        </div>
      </div>
    </div>
  );
}

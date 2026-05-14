'use client';

import { useCallback, useEffect, useState } from "react";

import { Header } from "@/components/layout/header";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import api from "@/lib/api";
import type { Shift, User } from "@/types";

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
  const [dateRange, setDateRange] = useState<{ start: string; end: string }>({ start: "", end: "" });
  const [selectedUser, setSelectedUser] = useState<string>("all");
  const [users, setUsers] = useState<User[]>([]);
  const [usersError, setUsersError] = useState<string | null>(null);

  const fetchUsers = useCallback(async () => {
    setUsersError(null);
    const result = await api.getUsers();
    if (result.error) {
      setUsersError("Could not load employees");
      return;
    }

    if (result.data) {
      setUsers(result.data);
    }
  }, []);

  useEffect(() => {
    const timer = window.setTimeout(() => {
      void fetchUsers();
    }, 0);

    return () => window.clearTimeout(timer);
  }, [fetchUsers]);

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
          {usersError ? (
            <div className="mb-4 rounded-md border border-destructive/20 bg-destructive/10 p-3 text-sm text-destructive">
              <div className="flex items-center justify-between gap-3">
                <span>{usersError}</span>
                <Button variant="outline" size="sm" onClick={() => void fetchUsers()}>
                  Retry
                </Button>
              </div>
            </div>
          ) : null}

          <ShiftsFilters
            dateRange={dateRange}
            selectedUser={selectedUser}
            users={users}
            onDateRangeChange={(newRange) => {
              setDateRange(newRange);
              setPage(1);
            }}
            onSelectedUserChange={(value) => {
              setSelectedUser(value);
              setPage(1);
            }}
          />

          <Card className="flex flex-col border-0 shadow-none bg-transparent md:border md:shadow md:bg-card">
            <CardContent className="flex-1 px-0 py-0 md:p-6">
              <div className="flex flex-col md:flex-row gap-4 items-center justify-between mb-4">
                <div className="flex items-center gap-2">
                  <span className="text-sm text-muted-foreground">Show</span>
                  <Select
                    value={String(limit)}
                    onValueChange={(value) => {
                      setLimit(Number(value));
                      setPage(1);
                    }}
                  >
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
                {total > 0 ? <span className="text-sm text-muted-foreground">{Math.min(shifts.length, limit)} entries of {total}</span> : null}
              </div>

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
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}

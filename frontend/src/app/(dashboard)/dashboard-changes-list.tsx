'use client';

import { History, Loader2, User } from "lucide-react";
import { useMemo, useState } from "react";

import type { AuditLog } from "@/types";
import { DataSortSelect } from "@/components/shared/data-sort-select";

import {
  DASHBOARD_ACTION_LABELS,
  formatDashboardDate,
  formatDashboardFieldName,
  formatDashboardFieldValue,
  getDashboardActionBadgeColor,
  getDashboardActionVerb,
  getDashboardChangeDescription,
  getDashboardFieldChanges,
  isDashboardImageField,
} from "./dashboard-helpers";

interface DashboardChangesListProps {
  logs: AuditLog[];
  isLoading: boolean;
  error: string | null;
  onRetry: () => void;
}

export function DashboardChangesList({ logs, isLoading, error, onRetry }: DashboardChangesListProps) {
  const [sort, setSort] = useState("date_desc");
  const sortedLogs = useMemo(
    () => [...logs].sort((left, right) => {
      const direction = sort.endsWith("_desc") ? -1 : 1;
      const sortBy = sort.replace(/_(asc|desc)$/, "");
      if (sortBy === "user") return (left.user_name || "System").localeCompare(right.user_name || "System") * direction;
      if (sortBy === "action") return left.action.localeCompare(right.action) * direction;
      return (new Date(left.created_at).getTime() - new Date(right.created_at).getTime()) * direction;
    }),
    [logs, sort],
  );

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="h-6 w-6 animate-spin text-primary" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex flex-col items-center justify-center py-12 text-muted-foreground gap-2">
        <History className="h-10 w-10 mb-2" />
        <p className="text-sm">{error}</p>
        <button onClick={onRetry} className="text-xs text-primary hover:underline">
          Retry
        </button>
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

  return (
    <div className="space-y-4">
      <div className="flex justify-end">
        <DataSortSelect
          value={sort}
          options={[
            { value: "date_desc", label: "Date (newest)" },
            { value: "date_asc", label: "Date (oldest)" },
            { value: "user_asc", label: "User (A-Z)" },
            { value: "action_asc", label: "Action (A-Z)" },
          ]}
          onChange={setSort}
        />
      </div>
      {sortedLogs.map((log) => {
        const fieldChanges = getDashboardFieldChanges(log);
        return (
          <div
            key={log.id}
            className="rounded-xl border p-4 bg-card text-card-foreground shadow-sm hover:shadow-md transition-shadow"
          >
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 mb-3">
              <div className="flex items-center gap-2 flex-wrap">
                <span
                  className={`inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-semibold uppercase tracking-wide whitespace-nowrap ${getDashboardActionBadgeColor(
                    log.action,
                  )}`}
                >
                  {DASHBOARD_ACTION_LABELS[getDashboardActionVerb(log.action)] || getDashboardActionVerb(log.action)}
                </span>
                <div className="flex items-center gap-1.5 text-sm font-medium">
                  <User className="h-3.5 w-3.5 text-muted-foreground shrink-0" />
                  <span className="break-words">{log.user_name || "System"}</span>
                </div>
              </div>
              <span className="text-xs text-muted-foreground whitespace-nowrap">{formatDashboardDate(log.created_at)}</span>
            </div>

            <div>
              <p className="text-base font-medium leading-snug mb-3">{getDashboardChangeDescription(log)}</p>
              {fieldChanges.length > 0 ? (
                <div className="mt-2 ml-0 space-y-0.5">
                  {fieldChanges.map(({ key, oldVal, newVal, label }) => (
                    <div key={`${key}-${label ?? "default"}`} className="text-xs flex items-start gap-1.5 text-muted-foreground">
                      <span className="font-medium text-foreground/70 shrink-0 pt-0.5">
                        {label ?? formatDashboardFieldName(key, log.action)}:
                      </span>
                      {isDashboardImageField(key) ? (
                        <div className="flex items-center gap-1.5">
                          {oldVal !== undefined ? (
                            <div className="flex flex-col items-center gap-0.5">
                              {/* eslint-disable-next-line @next/next/no-img-element */}
                              <img src={String(oldVal)} alt="old" className="h-8 w-8 object-cover rounded border border-red-300 opacity-60" />
                              <span className="text-[9px] text-red-500">before</span>
                            </div>
                          ) : null}
                          {oldVal !== undefined && newVal !== undefined ? <span className="text-muted-foreground">→</span> : null}
                          {newVal !== undefined ? (
                            <div className="flex flex-col items-center gap-0.5">
                              {/* eslint-disable-next-line @next/next/no-img-element */}
                              <img src={String(newVal)} alt="new" className="h-8 w-8 object-cover rounded border border-green-300" />
                              <span className="text-[9px] text-green-600">after</span>
                            </div>
                          ) : null}
                          {newVal === undefined && oldVal !== undefined ? <span className="text-red-500 text-[10px]">(removed)</span> : null}
                        </div>
                      ) : oldVal !== undefined && newVal !== undefined ? (
                        <>
                          <span className="text-red-500 line-through">{formatDashboardFieldValue(key, oldVal)}</span>
                          <span className="text-muted-foreground">→</span>
                          <span className="text-green-600">{formatDashboardFieldValue(key, newVal)}</span>
                        </>
                      ) : newVal !== undefined ? (
                        <span className="text-foreground/60">{formatDashboardFieldValue(key, newVal)}</span>
                      ) : (
                        <span className="text-red-500 line-through">{formatDashboardFieldValue(key, oldVal)}</span>
                      )}
                    </div>
                  ))}
                </div>
              ) : null}
            </div>
          </div>
        );
      })}
    </div>
  );
}

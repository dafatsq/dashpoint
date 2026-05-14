'use client';

import { ArrowRightLeft, CheckCircle2, CircleDot, Clock, Loader2, User } from "lucide-react";

import { Button } from "@/components/ui/button";
import type { Shift } from "@/types";

import { formatDashboardCurrency, formatDashboardDateTime, getShiftPreview } from "./dashboard-helpers";

interface DashboardShiftHistoryProps {
  shifts: Shift[];
  isLoading: boolean;
  error: string | null;
  onRetry: () => void;
}

export function DashboardShiftHistory({ shifts, isLoading, error, onRetry }: DashboardShiftHistoryProps) {
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
        <Clock className="h-10 w-10 mb-2" />
        <p className="text-sm">{error}</p>
        <Button variant="outline" size="sm" onClick={onRetry}>
          Retry
        </Button>
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
        const preview = getShiftPreview(shift);
        return (
          <div
            key={shift.id}
            className={`rounded-xl border p-4 shadow-sm transition-shadow hover:shadow-md ${
              preview.isOpen ? "border-green-500/50 bg-green-50/50 dark:bg-green-500/10" : "bg-card text-card-foreground"
            }`}
          >
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 mb-3">
              <div className="flex items-center gap-2">
                {preview.isOpen ? (
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
                  <span className="truncate max-w-[120px] sm:max-w-[200px]">{shift.employee_name || "Unknown"}</span>
                  {!preview.isOpen && shift.closed_by_name ? (
                    <span className="text-muted-foreground text-[10px] sm:text-xs font-normal">(Closed by {shift.closed_by_name})</span>
                  ) : null}
                </div>
              </div>
              <span className="text-xs text-muted-foreground whitespace-nowrap">{formatDashboardDateTime(shift.started_at)}</span>
            </div>

            <div className="flex flex-wrap items-center gap-x-4 gap-y-2 text-sm mt-3">
              <div className="flex items-center gap-1.5 whitespace-nowrap">
                <span className="text-muted-foreground text-xs">Open:</span>
                <span className="font-medium">{formatDashboardCurrency(preview.openingCash)}</span>
              </div>
              {!preview.isOpen && preview.closingCash !== null ? (
                <div className="flex items-center gap-1.5 whitespace-nowrap">
                  <ArrowRightLeft className="h-3.5 w-3.5 text-muted-foreground hidden sm:block" />
                  <span className="text-muted-foreground text-xs sm:ml-0">Close:</span>
                  <span className="font-medium">{formatDashboardCurrency(preview.closingCash)}</span>
                  {preview.cashDifference !== null ? (
                    <span
                      className={`text-xs ml-1 font-medium ${
                        preview.cashDifference > 0
                          ? "text-green-600"
                          : preview.cashDifference < 0
                            ? "text-red-600"
                            : "text-muted-foreground"
                      }`}
                    >
                      ({preview.cashDifference > 0 ? "+" : ""}
                      {formatDashboardCurrency(preview.cashDifference)})
                    </span>
                  ) : null}
                </div>
              ) : null}
              {!preview.isOpen && shift.ended_at ? (
                <div className="w-full sm:w-auto sm:ml-auto text-[11px] sm:text-xs text-muted-foreground mt-1 sm:mt-0">
                  Ended: {formatDashboardDateTime(shift.ended_at)}
                </div>
              ) : null}
            </div>

            {shift.operations && shift.operations.length > 0 ? (
              <div className="mt-3 pt-3 border-t border-dashed space-y-2">
                <p className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider mb-2">Cash Drawer Log</p>
                {shift.operations.map((operation) => (
                  <div key={operation.id} className="flex justify-between items-center bg-muted/30 p-2 rounded-md">
                    <div className="flex flex-col gap-0.5">
                      <div className="flex items-center gap-1.5">
                        <span className={`text-[10px] font-bold uppercase ${operation.type === "pay_in" ? "text-green-600 dark:text-green-500" : "text-red-600 dark:text-red-500"}`}>
                          {operation.type === "pay_in" ? "+ Pay In" : "- Pay Out"}
                        </span>
                        <span className="text-xs font-medium text-foreground">{operation.reason || "No reason specified"}</span>
                      </div>
                      <span className="text-[10px] text-muted-foreground">
                        {operation.performed_by_name || "System"} • {formatDashboardDateTime(operation.created_at)}
                      </span>
                    </div>
                    <span className={`text-sm font-semibold tabular-nums tracking-tight ${operation.type === "pay_in" ? "text-green-700 dark:text-green-400" : "text-red-700 dark:text-red-400"}`}>
                      {formatDashboardCurrency(Number.parseFloat(operation.amount))}
                    </span>
                  </div>
                ))}
              </div>
            ) : null}
          </div>
        );
      })}
    </div>
  );
}

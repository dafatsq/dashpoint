'use client';

import { ArrowRightLeft, CheckCircle2, ChevronLeft, ChevronRight, CircleDot, Clock, Loader2, User as UserIcon } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import type { Shift } from "@/types";

import { formatShiftCurrency, formatShiftDateTime, getShiftCashDifferenceTone, getShiftSummary } from "./shifts-helpers";

interface ShiftsListProps {
  shifts: Shift[];
  isLoading: boolean;
  error: string | null;
  page: number;
  limit: number;
  total: number;
  hasMore: boolean;
  onRetry: () => void;
  onPageChange: (page: number) => void;
}

export function ShiftsList({ shifts, isLoading, error, page, limit, total, hasMore, onRetry, onPageChange }: ShiftsListProps) {
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
        <Button variant="outline" size="sm" onClick={onRetry} className="mt-2">
          Retry
        </Button>
      </div>
    );
  }

  if (shifts.length === 0) {
    return (
      <Card>
        <CardContent className="flex flex-col items-center justify-center py-12 text-muted-foreground">
          <Clock className="h-10 w-10 mb-2" />
          <p className="text-sm">No shifts recorded yet</p>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-4">
      {shifts.map((shift) => {
        const summary = getShiftSummary(shift);
        return (
          <div
            key={shift.id}
            className={`rounded-xl border p-4 shadow-sm transition-shadow hover:shadow-md ${
              summary.isOpen ? "border-green-500/50 bg-green-50/50 dark:bg-green-500/10" : "bg-card text-card-foreground"
            }`}
          >
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 mb-3">
              <div className="flex items-center gap-2">
                {summary.isOpen ? (
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
                  <span className="truncate max-w-[120px] sm:max-w-[200px]">{shift.employee_name || "Unknown"}</span>
                  {!summary.isOpen && shift.closed_by_name ? (
                    <span className="text-muted-foreground text-[10px] sm:text-xs font-normal">(Closed by {shift.closed_by_name})</span>
                  ) : null}
                </div>
              </div>
              <span className="text-xs text-muted-foreground whitespace-nowrap">{formatShiftDateTime(shift.started_at)}</span>
            </div>

            <div className="flex flex-wrap items-center gap-x-4 gap-y-2 text-sm mt-3">
              <div className="flex items-center gap-1.5 whitespace-nowrap">
                <span className="text-muted-foreground text-xs">Open:</span>
                <span className="font-medium">{formatShiftCurrency(summary.openingCash)}</span>
              </div>
              {!summary.isOpen && summary.closingCash !== null ? (
                <div className="flex items-center gap-1.5 whitespace-nowrap">
                  <ArrowRightLeft className="h-3.5 w-3.5 text-muted-foreground hidden sm:block" />
                  <span className="text-muted-foreground text-xs sm:ml-0">Close:</span>
                  <span className="font-medium">{formatShiftCurrency(summary.closingCash)}</span>
                  {summary.cashDifference !== null ? (
                    <span className={`text-xs ml-1 font-medium ${getShiftCashDifferenceTone(summary.cashDifference)}`}>
                      ({summary.cashDifference > 0 ? "+" : ""}
                      {formatShiftCurrency(summary.cashDifference)})
                    </span>
                  ) : null}
                </div>
              ) : null}
              {!summary.isOpen && shift.ended_at ? (
                <div className="w-full sm:w-auto sm:ml-auto text-[11px] sm:text-xs text-muted-foreground mt-1 sm:mt-0">
                  Ended: {formatShiftDateTime(shift.ended_at)}
                </div>
              ) : null}
            </div>

            {shift.operations && shift.operations.length > 0 ? (
              <div className="mt-4 pt-3 border-t border-dashed space-y-2">
                <p className="text-[10px] font-semibold text-muted-foreground uppercase tracking-wider mb-2">Cash Drawer Log</p>
                <div className="grid gap-2">
                  {shift.operations.map((operation) => (
                    <div key={operation.id} className="text-xs flex items-start gap-2 bg-muted/30 p-2 rounded">
                      <span className={`font-semibold shrink-0 ${operation.type === "pay_in" ? "text-green-600" : "text-red-600"}`}>
                        {operation.type === "pay_in" ? "+ PAY IN" : "- PAY OUT"}
                      </span>
                      <div className="flex-1 min-w-0">
                        <span className="font-medium text-foreground">{operation.reason}</span>
                        <div className="flex items-center gap-1.5 text-[10px] text-muted-foreground mt-0.5">
                          <span>{operation.performed_by_name || "User"}</span>
                          <span>·</span>
                          <span>{formatShiftDateTime(operation.created_at)}</span>
                        </div>
                      </div>
                      <span className="font-medium whitespace-nowrap">{formatShiftCurrency(Number.parseFloat(operation.amount))}</span>
                    </div>
                  ))}
                </div>
              </div>
            ) : null}
          </div>
        );
      })}

      <div className="flex items-center justify-between mt-6 pt-4 border-t">
        <p className="text-sm text-muted-foreground">Page {page} • {Math.min(shifts.length, limit)} entries of {total}</p>
        <div className="flex gap-2">
          <Button variant="outline" size="sm" onClick={() => onPageChange(Math.max(1, page - 1))} disabled={page === 1}>
            <ChevronLeft className="h-4 w-4 mr-1" />
            Previous
          </Button>
          <Button variant="outline" size="sm" onClick={() => onPageChange(page + 1)} disabled={!hasMore}>
            Next
            <ChevronRight className="h-4 w-4 ml-1" />
          </Button>
        </div>
      </div>
    </div>
  );
}

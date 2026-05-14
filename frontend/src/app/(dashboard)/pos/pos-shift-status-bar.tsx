"use client";

import { Button } from "@/components/ui/button";
import { AlertCircle, ArrowDownCircle, ArrowUpCircle, CheckCircle, Clock } from "lucide-react";
import type { Shift } from "@/types";

interface PosShiftStatusBarProps {
  currentShift: Shift | null;
  canStartShift: boolean;
  canEndShift: boolean;
  onStartShift: () => void;
  onOpenShiftDetails: () => Promise<void>;
  onOpenCashDrawerDialog: (type: "pay_in" | "pay_out") => void;
}

export function PosShiftStatusBar({
  currentShift,
  canStartShift,
  canEndShift,
  onStartShift,
  onOpenShiftDetails,
  onOpenCashDrawerDialog,
}: PosShiftStatusBarProps) {
  if (!currentShift) {
    return (
      <div className="bg-card border-b border-l-4 border-l-yellow-500 px-6 py-4 flex items-center justify-between flex-shrink-0 shadow-sm">
        <div className="flex items-center gap-2">
          <AlertCircle className="h-5 w-5 text-yellow-600" />
          <span className="text-sm font-medium">No active shift. Start a shift to begin selling.</span>
        </div>
        {canStartShift ? (
          <Button size="sm" onClick={onStartShift}>
            <Clock className="h-4 w-4 mr-2" />
            Start Shift
          </Button>
        ) : null}
      </div>
    );
  }

  return (
    <div className="bg-card border-b border-l-4 border-l-green-500 px-6 py-4 flex items-center justify-between flex-shrink-0 shadow-sm">
      <div className="flex items-center gap-2">
        <CheckCircle className="h-5 w-5 text-green-600" />
        <span className="text-sm font-medium">
          Shift active <span className="text-muted-foreground mx-2">|</span>
          Started: {new Date(currentShift.started_at).toLocaleTimeString()}
        </span>
      </div>
      <div className="flex items-center gap-2">
        <Button
          size="sm"
          variant="outline"
          onClick={() => onOpenCashDrawerDialog("pay_in")}
          className="gap-1.5"
        >
          <ArrowDownCircle className="h-4 w-4 text-green-600" />
          <span className="hidden sm:inline">Pay In</span>
        </Button>
        <Button
          size="sm"
          variant="outline"
          onClick={() => onOpenCashDrawerDialog("pay_out")}
          className="gap-1.5"
        >
          <ArrowUpCircle className="h-4 w-4 text-red-600" />
          <span className="hidden sm:inline">Pay Out</span>
        </Button>
        {canEndShift ? (
          <Button size="sm" variant="outline" onClick={() => void onOpenShiftDetails()}>
            End Shift
          </Button>
        ) : null}
      </div>
    </div>
  );
}

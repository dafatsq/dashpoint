"use client";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Separator } from "@/components/ui/separator";
import { ArrowDownCircle, ArrowUpCircle, CheckCircle } from "lucide-react";
import type { CashDrawerOperation, Shift } from "@/types";

import { formatCurrency } from "./pos-helpers";

interface PosShiftDetailsDialogProps {
  open: boolean;
  currentShift: Shift | null;
  shiftClosed: boolean;
  closedShiftData: Shift | null;
  endingCash: string;
  closingNotes: string;
  cashDrawerOps: CashDrawerOperation[];
  cashDrawerTotals: {
    pay_in_total: string;
    pay_out_total: string;
  };
  isProcessing: boolean;
  onOpenChange: (open: boolean) => void;
  onEndingCashChange: (value: string) => void;
  onClosingNotesChange: (value: string) => void;
  onSubmit: () => Promise<void>;
  onDone: () => void;
}

export function PosShiftDetailsDialog({
  open,
  currentShift,
  shiftClosed,
  closedShiftData,
  endingCash,
  closingNotes,
  cashDrawerOps,
  cashDrawerTotals,
  isProcessing,
  onOpenChange,
  onEndingCashChange,
  onClosingNotesChange,
  onSubmit,
  onDone,
}: PosShiftDetailsDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        {shiftClosed && closedShiftData ? (
          <>
            <DialogHeader>
              <DialogTitle className="text-center">
                <CheckCircle className="h-12 w-12 text-green-500 mx-auto mb-2" />
                Shift Closed
              </DialogTitle>
              <DialogDescription className="text-center">
                Here is your shift reconciliation summary.
              </DialogDescription>
            </DialogHeader>
            <div className="space-y-3 py-4">
              <div className="grid grid-cols-2 gap-3 text-sm">
                <div>
                  <span className="text-muted-foreground block text-xs">Opening Cash</span>
                  <span className="font-medium">
                    {formatCurrency(parseFloat(closedShiftData.opening_cash) || 0)}
                  </span>
                </div>
                <div>
                  <span className="text-muted-foreground block text-xs">Closing Cash</span>
                  <span className="font-medium">
                    {formatCurrency(parseFloat(closedShiftData.closing_cash || "0"))}
                  </span>
                </div>
                <div>
                  <span className="text-muted-foreground block text-xs">
                    Total Sales ({closedShiftData.transaction_count} txn)
                  </span>
                  <span className="font-medium text-green-600">
                    +{formatCurrency(parseFloat(closedShiftData.total_sales) || 0)}
                  </span>
                </div>
                <div>
                  <span className="text-muted-foreground block text-xs">Total Cash Sales</span>
                  <span className="font-medium text-blue-600">
                    +{formatCurrency(parseFloat(closedShiftData.total_cash_sales) || 0)}
                  </span>
                </div>
                <div>
                  <span className="text-muted-foreground block text-xs">
                    Total Refunds ({closedShiftData.refund_count})
                  </span>
                  <span className="font-medium text-red-600">
                    -{formatCurrency(parseFloat(closedShiftData.total_refunds) || 0)}
                  </span>
                </div>
                <div>
                  <span className="text-muted-foreground block text-xs">Expected Cash</span>
                  <span className="font-bold">
                    {formatCurrency(parseFloat(closedShiftData.expected_cash || "0"))}
                  </span>
                </div>
                <div>
                  <span className="text-muted-foreground block text-xs">Difference</span>
                  <span
                    className={`font-bold ${
                      parseFloat(closedShiftData.cash_difference || "0") >= 0
                        ? "text-green-600"
                        : "text-red-600"
                    }`}
                  >
                    {formatCurrency(parseFloat(closedShiftData.cash_difference || "0"))}
                  </span>
                </div>
              </div>
            </div>
            <DialogFooter>
              <Button className="w-full" onClick={onDone}>
                Done
              </Button>
            </DialogFooter>
          </>
        ) : currentShift ? (
          <>
            <DialogHeader>
              <DialogTitle>End Shift</DialogTitle>
              <DialogDescription>
                Count your cash and enter the total below. Shift details will be shown after closing.
              </DialogDescription>
            </DialogHeader>

            <div className="space-y-4 py-4">
              <div className="grid grid-cols-1 gap-4 text-sm">
                <div>
                  <span className="text-muted-foreground block text-xs">Started At</span>
                  <span className="font-medium">
                    {new Date(currentShift.started_at).toLocaleString()}
                  </span>
                </div>
              </div>

              {cashDrawerOps.length > 0 ? (
                <>
                  <Separator />
                  <div>
                    <h4 className="text-sm font-medium mb-2">Cash Drawer Activity</h4>
                    <div className="space-y-1.5 max-h-32 overflow-y-auto">
                      {cashDrawerOps.map((operation) => (
                        <div
                          key={operation.id}
                          className="flex items-center justify-between text-xs"
                        >
                          <div className="flex items-center gap-1.5">
                            {operation.type === "pay_in" ? (
                              <ArrowDownCircle className="h-3.5 w-3.5 text-green-600" />
                            ) : (
                              <ArrowUpCircle className="h-3.5 w-3.5 text-red-600" />
                            )}
                            <span className="text-muted-foreground">{operation.reason}</span>
                          </div>
                          <span
                            className={
                              operation.type === "pay_in"
                                ? "text-green-600 font-medium"
                                : "text-red-600 font-medium"
                            }
                          >
                            {operation.type === "pay_in" ? "+" : "-"}
                            {formatCurrency(parseFloat(operation.amount))}
                          </span>
                        </div>
                      ))}
                    </div>
                    <div className="flex justify-between text-xs mt-2 pt-2 border-t">
                      <span className="text-green-600">
                        Pay-In: +{formatCurrency(parseFloat(cashDrawerTotals.pay_in_total))}
                      </span>
                      <span className="text-red-600">
                        Pay-Out: -{formatCurrency(parseFloat(cashDrawerTotals.pay_out_total))}
                      </span>
                    </div>
                  </div>
                </>
              ) : null}

              <Separator />

              <div>
                <label className="text-sm font-medium">Ending Cash (Counted)</label>
                <Input
                  type="number"
                  value={endingCash}
                  onChange={(event) => onEndingCashChange(event.target.value)}
                  placeholder="Count the cash in your drawer..."
                  className="mt-1.5"
                />
                <p className="text-xs text-muted-foreground mt-1">
                  Enter the actual amount of cash you have in the drawer.
                </p>
              </div>

              <div>
                <label className="text-sm font-medium">Notes (Optional)</label>
                <Input
                  value={closingNotes}
                  onChange={(event) => onClosingNotesChange(event.target.value)}
                  placeholder="Any discrepancies or comments..."
                  className="mt-1.5"
                />
              </div>
            </div>

            <DialogFooter>
              <Button variant="outline" onClick={() => onOpenChange(false)}>
                Cancel
              </Button>
              <Button variant="destructive" onClick={() => void onSubmit()} disabled={!endingCash || isProcessing}>
                {isProcessing ? "Ending Shift..." : "End Shift"}
              </Button>
            </DialogFooter>
          </>
        ) : (
          <>
            <DialogHeader>
              <DialogTitle>No Active Shift</DialogTitle>
              <DialogDescription>There is no active shift to end.</DialogDescription>
            </DialogHeader>
            <DialogFooter>
              <Button onClick={() => onOpenChange(false)}>Close</Button>
            </DialogFooter>
          </>
        )}
      </DialogContent>
    </Dialog>
  );
}

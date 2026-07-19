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
import { ArrowDownCircle, ArrowUpCircle } from "lucide-react";
import { getCashDrawerDialogCopy } from "./pos-helpers";

interface PosCashDrawerDialogProps {
  open: boolean;
  operationType: "pay_in" | "pay_out";
  amount: string;
  reason: string;
  isSubmitting: boolean;
  onOpenChange: (open: boolean) => void;
  onAmountChange: (value: string) => void;
  onReasonChange: (value: string) => void;
  onSubmit: () => Promise<void>;
}

export function PosCashDrawerDialog({
  open,
  operationType,
  amount,
  reason,
  isSubmitting,
  onOpenChange,
  onAmountChange,
  onReasonChange,
  onSubmit,
}: PosCashDrawerDialogProps) {
  const copy = getCashDrawerDialogCopy(operationType);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-sm">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            {operationType === "pay_in" ? (
              <>
                <ArrowDownCircle className="h-5 w-5 text-green-600" /> {copy.title}
              </>
            ) : (
              <>
                <ArrowUpCircle className="h-5 w-5 text-red-600" /> {copy.title}
              </>
            )}
          </DialogTitle>
          <DialogDescription>{copy.description}</DialogDescription>
        </DialogHeader>
        <div className="space-y-4 py-4">
          <div>
            <label className="text-sm font-medium">Amount (IDR)</label>
            <Input
              type="number"
              value={amount}
              onChange={(event) => onAmountChange(event.target.value)}
              placeholder="Enter amount..."
              className="mt-1.5"
            />
          </div>
          <div>
            <label className="text-sm font-medium">Reason</label>
            <Input
              value={reason}
              onChange={(event) => onReasonChange(event.target.value)}
              placeholder="e.g., Petty cash withdrawal"
              className="mt-1.5"
            />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button
            onClick={() => void onSubmit()}
            disabled={isSubmitting}
            className={copy.confirmClassName}
          >
            {isSubmitting ? "Processing..." : copy.confirmLabel}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

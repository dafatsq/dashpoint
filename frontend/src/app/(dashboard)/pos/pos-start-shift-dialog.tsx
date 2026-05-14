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

interface PosStartShiftDialogProps {
  open: boolean;
  startingCash: string;
  onOpenChange: (open: boolean) => void;
  onStartingCashChange: (value: string) => void;
  onSubmit: () => Promise<void>;
}

export function PosStartShiftDialog({
  open,
  startingCash,
  onOpenChange,
  onStartingCashChange,
  onSubmit,
}: PosStartShiftDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Start New Shift</DialogTitle>
          <DialogDescription>
            Enter the starting cash amount in your drawer to begin your shift.
          </DialogDescription>
        </DialogHeader>
        <div className="py-4">
          <label className="text-sm font-medium">Starting Cash (IDR)</label>
          <Input
            type="number"
            value={startingCash}
            onChange={(event) => onStartingCashChange(event.target.value)}
            placeholder="0"
            className="mt-1"
          />
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={() => void onSubmit()} disabled={!startingCash}>
            Start Shift
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

"use client";

import { Loader2 } from "lucide-react";

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
import { Label } from "@/components/ui/label";

interface SalesVoidDialogProps {
  open: boolean;
  reason: string;
  isVoiding: boolean;
  onOpenChange: (open: boolean) => void;
  onReasonChange: (reason: string) => void;
  onConfirm: () => void;
}

export function SalesVoidDialog({
  open,
  reason,
  isVoiding,
  onOpenChange,
  onReasonChange,
  onConfirm,
}: SalesVoidDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Void Sale</DialogTitle>
          <DialogDescription>
            This will void the sale and restore inventory. This action cannot be
            undone.
          </DialogDescription>
        </DialogHeader>

        <div className="py-4">
          <Label htmlFor="reason">Reason for voiding *</Label>
          <Input
            id="reason"
            value={reason}
            onChange={(event) => onReasonChange(event.target.value)}
            placeholder="Enter reason..."
            className="mt-2"
          />
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button
            variant="destructive"
            onClick={onConfirm}
            disabled={isVoiding}
          >
            {isVoiding ? (
              <>
                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                Voiding...
              </>
            ) : (
              "Confirm Void"
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

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
import { AlertCircle } from "lucide-react";

import type { PosErrorState } from "./pos-types";

interface PosErrorDialogProps {
  error: PosErrorState | null;
  onClose: () => void;
}

export function PosErrorDialog({ error, onClose }: PosErrorDialogProps) {
  return (
    <Dialog open={!!error} onOpenChange={(open) => !open && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle className="text-destructive flex items-center gap-2">
            <AlertCircle className="h-5 w-5" />
            {error?.title}
          </DialogTitle>
          <DialogDescription className="pt-2 text-base text-foreground">
            {error?.message}
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button onClick={onClose}>OK</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

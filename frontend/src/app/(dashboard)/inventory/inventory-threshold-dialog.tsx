'use client';

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
import type { Product } from "@/types";

interface InventoryThresholdDialogProps {
  open: boolean;
  product: Product | null;
  value: string;
  isSubmitting: boolean;
  onOpenChange: (open: boolean) => void;
  onValueChange: (value: string) => void;
  onSubmit: () => void;
}

export function InventoryThresholdDialog({
  open,
  product,
  value,
  isSubmitting,
  onOpenChange,
  onValueChange,
  onSubmit,
}: InventoryThresholdDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Edit Low Stock Threshold</DialogTitle>
          <DialogDescription>
            {product ? `${product.name}${product.sku ? ` • ${product.sku}` : ""}` : "Selected product"}
          </DialogDescription>
        </DialogHeader>

        <div className="grid gap-2 py-4">
          <Label htmlFor="low_stock_threshold">Low Stock Threshold</Label>
          <Input
            id="low_stock_threshold"
            type="number"
            min="0"
            step="0.01"
            value={value}
            onChange={(event) => onValueChange(event.target.value)}
            placeholder="Enter threshold"
          />
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={onSubmit} disabled={isSubmitting}>
            {isSubmitting ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Saving...
              </>
            ) : (
              "Save Threshold"
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

'use client';

import { AlertTriangle, Loader2 } from "lucide-react";

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
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import type { AdjustmentType, Product, ProductInventoryDetails } from "@/types";

import {
  ADJUSTMENT_TYPE_OPTIONS,
  getInventoryAdjustmentChangeLabel,
  getInventoryAdjustmentTypeLabel,
  getInventoryProductQuantity,
  type AdjustmentFormState,
  type InventoryAction,
} from "./inventory-helpers";

interface InventoryAdjustDialogProps {
  open: boolean;
  product: Product | null;
  inventoryDetails: ProductInventoryDetails | null;
  formState: AdjustmentFormState;
  allowedActions: InventoryAction[];
  isLoadingInventoryDetails: boolean;
  isSubmitting: boolean;
  onOpenChange: (open: boolean) => void;
  onActionChange: (action: InventoryAction) => void;
  onFormStateChange: (formState: AdjustmentFormState) => void;
  onSubmit: () => void;
}

export function InventoryAdjustDialog({
  open,
  product,
  inventoryDetails,
  formState,
  allowedActions,
  isLoadingInventoryDetails,
  isSubmitting,
  onOpenChange,
  onActionChange,
  onFormStateChange,
  onSubmit,
}: InventoryAdjustDialogProps) {
  const currentStock = product ? getInventoryProductQuantity(product) : 0;
  const quantity = Number.parseInt(formState.quantity || "0", 10) || 0;
  const nextStock =
    formState.adjustmentType === "count"
      ? quantity
      : currentStock + (formState.action === "add" ? 1 : -1) * quantity;
  const recentAdjustments = inventoryDetails?.recent_adjustments || [];

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>
            {formState.action === "add" ? "Add Stock" : formState.action === "remove" ? "Remove Stock" : "Stock Count"}
          </DialogTitle>
          <DialogDescription asChild>
            <div className="space-y-2">
              <div>
                {product?.name} - Current stock: {currentStock}
              </div>
              {formState.action === "count" ? (
                <div className="flex items-center gap-2 text-blue-600 dark:text-blue-400 text-xs">
                  <svg className="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  <span>Enter the exact quantity counted. This will update inventory to match your physical count.</span>
                </div>
              ) : (
                <div className="flex items-center gap-2 text-yellow-600 dark:text-yellow-500 text-xs">
                  <AlertTriangle className="h-3 w-3" />
                  <span>Stock updates in real-time. Current value may change if others adjust inventory.</span>
                </div>
              )}
            </div>
          </DialogDescription>
        </DialogHeader>

        <div className="grid gap-4 py-4">


          <div className="grid gap-2">
            <Label htmlFor="actionType">Action</Label>
            {allowedActions.length === 0 ? (
              <p className="text-xs text-muted-foreground">
                No inventory actions are currently available with your permissions.
              </p>
            ) : null}
            <Select value={formState.action} onValueChange={(value) => onActionChange(value as InventoryAction)}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {allowedActions.includes("add") ? <SelectItem value="add">Add Stock</SelectItem> : null}
                {allowedActions.includes("remove") ? <SelectItem value="remove">Remove Stock</SelectItem> : null}
                {allowedActions.includes("count") ? <SelectItem value="count">Stock Count</SelectItem> : null}
              </SelectContent>
            </Select>
          </div>

          <div className="grid gap-2">
            <Label htmlFor="quantity">Quantity</Label>
            <Input
              id="quantity"
              type="number"
              min="1"
              value={formState.quantity}
              onChange={(event) => onFormStateChange({ ...formState, quantity: event.target.value })}
              placeholder="Enter quantity"
            />
          </div>

          {formState.action !== "count" ? (
            <div className="grid gap-2">
              <Label htmlFor="adjustmentType">Adjustment Type</Label>
              <Select
                value={formState.adjustmentType}
                onValueChange={(value) => onFormStateChange({ ...formState, adjustmentType: value as AdjustmentType })}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {ADJUSTMENT_TYPE_OPTIONS[formState.action].map((option) => (
                    <SelectItem key={option.value} value={option.value}>
                      {option.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          ) : null}

          <div className="grid gap-2">
            <Label htmlFor="notes">Notes (optional)</Label>
            <Input
              id="notes"
              value={formState.notes}
              onChange={(event) => onFormStateChange({ ...formState, notes: event.target.value })}
              placeholder="Additional notes..."
            />
          </div>

          {formState.quantity ? (
            <div className="rounded-lg bg-muted p-3">
              <p className="text-sm">
                {formState.adjustmentType === "count" ? (
                  <>
                    Setting stock to: <span className="font-bold">{quantity}</span>
                    <span className="text-muted-foreground ml-2">(Current: {currentStock})</span>
                  </>
                ) : (
                  <>
                    New stock level: <span className="font-bold">{nextStock}</span>
                  </>
                )}
              </p>
            </div>
          ) : null}

          <div className="grid gap-2">
            <Label>Recent Stock History</Label>
            <div className="rounded-lg border bg-muted/30">
              {isLoadingInventoryDetails ? (
                <div className="flex items-center gap-2 px-3 py-4 text-sm text-muted-foreground">
                  <Loader2 className="h-4 w-4 animate-spin" />
                  Loading recent adjustments...
                </div>
              ) : recentAdjustments.length === 0 ? (
                <div className="px-3 py-4 text-sm text-muted-foreground">
                  No recent stock adjustments for this product.
                </div>
              ) : (
                <div className="divide-y">
                  {recentAdjustments.slice(0, 5).map((adjustment) => (
                    <div key={adjustment.id} className="px-3 py-3 text-sm">
                      <div className="flex items-start justify-between gap-3">
                        <div className="min-w-0">
                          <div className="font-medium">
                            {getInventoryAdjustmentTypeLabel(adjustment.adjustment_type)}
                          </div>
                          <div className="text-xs text-muted-foreground">
                            {adjustment.adjusted_by_user?.name || "Former user"} •{" "}
                            {new Date(adjustment.created_at).toLocaleString("id-ID", {
                              dateStyle: "medium",
                              timeStyle: "short",
                            })}
                          </div>
                        </div>
                        <div className="text-right">
                          <div
                            className={`font-semibold ${
                              adjustment.adjustment_type === "count"
                                ? "text-blue-600"
                                : adjustment.quantity_change.startsWith("-")
                                  ? "text-red-600"
                                  : "text-green-600"
                            }`}
                          >
                            {getInventoryAdjustmentChangeLabel(adjustment)}
                          </div>
                          <div className="text-xs text-muted-foreground">
                            Stock {adjustment.quantity_before} → {adjustment.quantity_after}
                          </div>
                        </div>
                      </div>
                      {adjustment.reason ? (
                        <div className="mt-2 text-xs text-muted-foreground">{adjustment.reason}</div>
                      ) : null}
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button
            onClick={onSubmit}
            disabled={isSubmitting}
            variant={formState.action === "remove" ? "destructive" : "default"}
          >
            {isSubmitting ? (
              <>
                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                Saving...
              </>
            ) : formState.action === "add" ? (
              "Add Stock"
            ) : formState.action === "remove" ? (
              "Remove Stock"
            ) : (
              "Update Count"
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

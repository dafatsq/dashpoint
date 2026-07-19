'use client';

import { Loader2 } from "lucide-react";

import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { DateRangePicker } from "@/components/ui/date-range-picker";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import type { Product, ProductInventoryDetails } from "@/types";

import {
  INVENTORY_HISTORY_FILTER_OPTIONS,
  getInventoryAdjustmentChangeLabel,
  getInventoryAdjustmentTypeLabel,
  getInventoryProductMinQuantity,
  getInventoryProductQuantity,
  type InventoryHistoryFilter,
} from "./inventory-helpers";

interface InventoryHistoryDrawerProps {
  open: boolean;
  product: Product | null;
  inventoryDetails: ProductInventoryDetails | null;
  users: { id: string; name: string }[];
  isLoading: boolean;
  offset: number;
  selectedType: InventoryHistoryFilter;
  selectedUserId: string;
  selectedDateRange: { start: string; end: string };
  onOpenChange: (open: boolean) => void;
  onTypeChange: (value: InventoryHistoryFilter) => void;
  onUserChange: (value: string) => void;
  onDateRangeChange: (value: { start: string; end: string }) => void;
  onPrevious: () => void;
  onNext: () => void;
}

export function InventoryHistoryDrawer({
  open,
  product,
  inventoryDetails,
  users,
  isLoading,
  offset,
  selectedType,
  selectedUserId,
  selectedDateRange,
  onOpenChange,
  onTypeChange,
  onUserChange,
  onDateRangeChange,
  onPrevious,
  onNext,
}: InventoryHistoryDrawerProps) {
  const adjustments = inventoryDetails?.recent_adjustments || [];
  const totalAdjustments = inventoryDetails?.total_adjustments || 0;
  const currentStock = product ? getInventoryProductQuantity(product) : 0;
  const minStock = product ? getInventoryProductMinQuantity(product) : 0;
  const hasPrevious = offset > 0;
  const hasNext = offset + adjustments.length < totalAdjustments;

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent side="right" className="w-full sm:max-w-xl p-0 bg-card gap-0">
        <div className="flex h-full flex-col">
          <SheetHeader className="border-b px-6 py-5">
            <SheetTitle>Stock History</SheetTitle>
            <SheetDescription>
              {product ? `${product.name}${product.sku ? ` • ${product.sku}` : ""}` : "Selected product"}
            </SheetDescription>
          </SheetHeader>

          <div className="flex-1 overflow-y-auto px-6 py-4 space-y-4">
            {product ? (
              <Card>
                <CardContent className="grid grid-cols-2 gap-3 p-4 text-sm">
                  <div>
                    <div className="text-xs font-medium text-muted-foreground">Current Stock</div>
                    <div className="mt-1 font-semibold">{currentStock}</div>
                  </div>
                  <div>
                    <div className="text-xs font-medium text-muted-foreground">Low Stock Threshold</div>
                    <div className="mt-1 font-semibold">{minStock}</div>
                  </div>
                </CardContent>
              </Card>
            ) : null}

            <Card>
              <CardContent className="p-4">
                <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                  <div className="space-y-1.5">
                    <label className="text-xs font-medium text-muted-foreground">Type</label>
                    <Select value={selectedType} onValueChange={(value) => onTypeChange(value as InventoryHistoryFilter)}>
                      <SelectTrigger className="w-full">
                        <SelectValue placeholder="All Activity" />
                      </SelectTrigger>
                      <SelectContent>
                        {INVENTORY_HISTORY_FILTER_OPTIONS.map((option) => (
                          <SelectItem key={option.value} value={option.value}>
                            {option.label}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                  <div className="space-y-1.5">
                    <label className="text-xs font-medium text-muted-foreground">Employee</label>
                    <Select value={selectedUserId} onValueChange={onUserChange}>
                      <SelectTrigger className="w-full">
                        <SelectValue placeholder="All Employees" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="all">All Employees</SelectItem>
                        {users.map((user) => (
                          <SelectItem key={user.id} value={user.id}>
                            {user.name}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                  <div className="space-y-1.5 sm:col-span-2">
                    <label className="text-xs font-medium text-muted-foreground">Date Range</label>
                    <DateRangePicker
                      value={selectedDateRange}
                      onChange={onDateRangeChange}
                      placeholder="Filter by date..."
                      className="w-full"
                    />
                  </div>
                </div>
              </CardContent>
            </Card>

            {isLoading ? (
              <div className="flex items-center gap-2 text-sm text-muted-foreground">
                <Loader2 className="h-4 w-4 animate-spin" />
                Loading stock history...
              </div>
            ) : adjustments.length === 0 ? (
              <div className="rounded-md border border-dashed px-4 py-6 text-sm text-muted-foreground">
                No stock adjustments match the current filters.
              </div>
            ) : (
              <div className="space-y-3">
                {adjustments.map((adjustment) => (
                  <Card key={adjustment.id}>
                    <CardContent className="px-4 py-3">
                    <div className="flex items-start justify-between gap-3">
                      <div className="min-w-0">
                        <div className="font-medium">{getInventoryAdjustmentTypeLabel(adjustment.adjustment_type)}</div>
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
                      <div className="mt-2 text-sm text-muted-foreground">{adjustment.reason}</div>
                    ) : null}
                    </CardContent>
                  </Card>
                ))}
              </div>
            )}
          </div>

          <div className="border-t px-6 py-4">
            <div className="mb-3 text-xs text-muted-foreground">
              {totalAdjustments > 0
                ? `Showing ${offset + 1}-${Math.min(offset + adjustments.length, totalAdjustments)} of ${totalAdjustments} adjustments`
                : "No adjustments to display"}
            </div>
            <div className="flex items-center justify-end gap-2">
              <Button variant="outline" onClick={onPrevious} disabled={isLoading || !hasPrevious}>
                Previous
              </Button>
              <Button variant="outline" onClick={onNext} disabled={isLoading || !hasNext}>
                Next
              </Button>
            </div>
            <Button className="mt-3 w-full" onClick={() => onOpenChange(false)}>
              Close
            </Button>
          </div>
        </div>
      </SheetContent>
    </Sheet>
  );
}

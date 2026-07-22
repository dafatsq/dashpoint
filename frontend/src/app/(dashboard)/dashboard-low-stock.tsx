'use client';

import { AlertTriangle } from "lucide-react";
import { useMemo, useState } from "react";

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import type { LowStockItem } from "@/types";
import { DataSortSelect } from "@/components/shared/data-sort-select";

interface DashboardLowStockProps {
  items: LowStockItem[];
}

export function DashboardLowStock({ items }: DashboardLowStockProps) {
  const [sort, setSort] = useState("stock_asc");
  const sortedItems = useMemo(
    () => [...items].sort((left, right) => {
      const direction = sort.endsWith("_desc") ? -1 : 1;
      const sortBy = sort.replace(/_(asc|desc)$/, "");
      if (sortBy === "name") return left.name.localeCompare(right.name) * direction;
      return (Number.parseFloat(left.quantity) - Number.parseFloat(right.quantity)) * direction;
    }),
    [items, sort],
  );

  if (items.length === 0) {
    return null;
  }

  return (
    <Card className="mb-6">
      <CardHeader>
        <div className="flex flex-wrap items-center justify-between gap-4">
          <div className="flex items-center gap-2">
            <AlertTriangle className="h-5 w-5 text-yellow-500" />
            <CardTitle>Low Stock Alert</CardTitle>
          </div>
          <DataSortSelect
            value={sort}
            options={[
              { value: "stock_asc", label: "Stock (low-high)" },
              { value: "stock_desc", label: "Stock (high-low)" },
              { value: "name_asc", label: "Name (A-Z)" },
            ]}
            onChange={setSort}
          />
        </div>
        <CardDescription>These products are running low and need to be restocked</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-3">
          {sortedItems.map((item) => (
            <div key={item.id} className="flex items-center justify-between rounded-lg border p-3">
              <div>
                <p className="font-medium">{item.name}</p>
                <p className="text-sm text-muted-foreground">SKU: {item.sku || "N/A"}</p>
              </div>
              <div className="text-right">
                <p className="font-bold text-destructive">{Number.parseFloat(item.quantity)} left</p>
                <p className="text-xs text-muted-foreground">Available: {Number.parseFloat(item.available_quantity)}</p>
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}

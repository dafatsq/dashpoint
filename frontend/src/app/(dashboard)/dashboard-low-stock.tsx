'use client';

import { AlertTriangle } from "lucide-react";

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import type { LowStockItem } from "@/types";

interface DashboardLowStockProps {
  items: LowStockItem[];
}

export function DashboardLowStock({ items }: DashboardLowStockProps) {
  if (items.length === 0) {
    return null;
  }

  return (
    <Card className="mb-6">
      <CardHeader>
        <div className="flex items-center gap-2">
          <AlertTriangle className="h-5 w-5 text-yellow-500" />
          <CardTitle>Low Stock Alert</CardTitle>
        </div>
        <CardDescription>These products are running low and need to be restocked</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-3">
          {items.map((item) => (
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

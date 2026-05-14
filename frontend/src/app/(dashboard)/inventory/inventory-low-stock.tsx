'use client';

import { Boxes, ImageIcon, Package, Settings2 } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import type { LowStockItem, Product } from "@/types";

import { getInventoryProductImageUrl } from "./inventory-helpers";

interface InventoryLowStockProps {
  items: LowStockItem[];
  products: Product[];
  canModifyStock: boolean;
  onAdjust: (productId: string) => void;
}

export function InventoryLowStock({ items, products, canModifyStock, onAdjust }: InventoryLowStockProps) {
  if (items.length === 0) {
    return (
      <Card>
        <CardContent className="flex flex-col items-center justify-center py-12">
          <Boxes className="h-12 w-12 text-muted-foreground mb-4" />
          <p className="text-muted-foreground">All products are well-stocked!</p>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Low Stock Items</CardTitle>
        <CardDescription>Products that are below their minimum stock level</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-3">
          {items.map((item) => {
            const product = products.find((entry) => entry.id === item.id);
            return (
              <div key={item.id} className="flex items-center justify-between rounded-lg border p-4">
                <div className="flex items-center gap-3 min-w-0">
                  {product?.image_url ? (
                    <div className="relative h-12 w-12 rounded border overflow-hidden flex-shrink-0 bg-muted">
                      <img
                        src={getInventoryProductImageUrl(product.image_url)}
                        alt={item.name}
                        className="h-full w-full object-cover"
                        onError={(event) => {
                          (event.target as HTMLImageElement).style.display = "none";
                          ((event.target as HTMLImageElement).nextSibling as HTMLElement).style.display = "flex";
                        }}
                      />
                      <div className="absolute inset-0 hidden items-center justify-center bg-muted">
                        <ImageIcon className="h-5 w-5 text-muted-foreground" />
                      </div>
                    </div>
                  ) : (
                    <div className="h-12 w-12 rounded border flex items-center justify-center bg-muted flex-shrink-0">
                      <Package className="h-5 w-5 text-muted-foreground" />
                    </div>
                  )}

                  <div className="min-w-0">
                    <p className="font-medium truncate">{item.name}</p>
                    <p className="text-sm text-muted-foreground">SKU: {item.sku || "N/A"}</p>
                  </div>
                </div>
                <div className="flex items-center gap-4">
                  <div className="text-right">
                    <p className="font-bold text-destructive">{Number.parseFloat(item.quantity)} left</p>
                    <p className="text-xs text-muted-foreground">Available: {Number.parseFloat(item.available_quantity)}</p>
                  </div>
                  {canModifyStock ? (
                    <Button size="sm" variant="outline" onClick={() => onAdjust(item.id)}>
                      <Settings2 className="h-4 w-4 mr-1" />
                      Adjust
                    </Button>
                  ) : null}
                </div>
              </div>
            );
          })}
        </div>
      </CardContent>
    </Card>
  );
}

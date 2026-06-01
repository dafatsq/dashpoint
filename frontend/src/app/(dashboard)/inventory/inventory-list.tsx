'use client';

import { ImageIcon, Loader2, Package, Settings2 } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import type { Product } from "@/types";

import {
  classifyInventoryStock,
  getInventoryProductImageUrl,
  getInventoryProductMinQuantity,
  getInventoryProductQuantity,
} from "./inventory-helpers";

interface InventoryListProps {
  products: Product[];
  totalProducts: number;
  hasMore: boolean;
  isFetchingMore: boolean;
  canModifyStock: boolean;
  canEditThreshold: boolean;
  loadMoreRef: React.RefObject<HTMLDivElement | null>;
  onAdjust: (product: Product) => void;
  onHistory: (product: Product) => void;
  onEditThreshold: (product: Product) => void;
}

export function InventoryList({
  products,
  totalProducts,
  hasMore,
  isFetchingMore,
  canModifyStock,
  canEditThreshold,
  loadMoreRef,
  onAdjust,
  onHistory,
  onEditThreshold,
}: InventoryListProps) {
  return (
    <>
      <div className="lg:hidden space-y-4">
        {products.map((product) => {
          const quantity = getInventoryProductQuantity(product);
          const minQuantity = getInventoryProductMinQuantity(product);
          const { isLowStock, isOutOfStock } = classifyInventoryStock(product);

          return (
            <div key={product.id} className="border rounded-lg p-4 bg-card text-card-foreground shadow-sm">
              <div className="flex items-start gap-4">
                {product.image_url ? (
                  <div className="relative h-20 w-20 rounded border overflow-hidden flex-shrink-0 bg-muted">
                    <img
                      src={getInventoryProductImageUrl(product.image_url)}
                      alt={product.name}
                      className="h-full w-full object-cover"
                      onError={(event) => {
                        (event.target as HTMLImageElement).style.display = "none";
                        ((event.target as HTMLImageElement).nextSibling as HTMLElement).style.display = "flex";
                      }}
                    />
                    <div className="absolute inset-0 hidden items-center justify-center bg-muted">
                      <ImageIcon className="h-8 w-8 text-muted-foreground" />
                    </div>
                  </div>
                ) : (
                  <div className="h-20 w-20 rounded border flex items-center justify-center bg-muted flex-shrink-0">
                    <Package className="h-8 w-8 text-muted-foreground" />
                  </div>
                )}

                <div className="flex-1 min-w-0">
                  <div className="flex justify-between items-start mb-2">
                    <div>
                      <h3 className="font-semibold line-clamp-1">{product.name}</h3>
                      <p className="text-sm text-muted-foreground">{product.sku || "-"}</p>
                    </div>
                    <span
                      className={`inline-flex items-center px-2 py-1 rounded-full text-[10px] font-bold uppercase tracking-wider ${
                        isOutOfStock ? "bg-red-600 text-white" : isLowStock ? "bg-yellow-600 text-white" : "bg-green-600 text-white"
                      }`}
                    >
                      {isOutOfStock ? "Out" : isLowStock ? "Low" : "In"}
                    </span>
                  </div>

                  <div className="grid grid-cols-2 gap-2 text-sm mb-3">
                    <div>
                      <span className="text-muted-foreground">Stock: </span>
                      <span className={`font-bold ${isLowStock ? "text-destructive" : ""}`}>{quantity}</span>
                    </div>
                    <div>
                      <span className="text-muted-foreground">Min: </span>
                      <span>{minQuantity}</span>
                    </div>
                  </div>

                  <div className="flex items-center justify-end gap-2 border-t pt-3 mt-1">
                    <Button variant="outline" size="sm" onClick={() => onHistory(product)} className="h-8">
                      History
                    </Button>
                    {canEditThreshold ? (
                      <Button variant="outline" size="sm" onClick={() => onEditThreshold(product)} className="h-8">
                        Edit Threshold
                      </Button>
                    ) : null}
                    {canModifyStock ? (
                      <Button variant="outline" size="sm" onClick={() => onAdjust(product)} className="h-8">
                        <Settings2 className="h-3.5 w-3.5 mr-1" />
                        Adjust
                      </Button>
                    ) : null}
                  </div>
                </div>
              </div>
            </div>
          );
        })}
      </div>

      <Card className="hidden lg:block">
        <CardHeader>
          <CardTitle>Inventory ({totalProducts || products.length})</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b text-left text-sm text-muted-foreground">
                  <th className="pb-3 font-medium">Product</th>
                  <th className="pb-3 font-medium hidden lg:table-cell">SKU</th>
                  <th className="pb-3 font-medium text-right">Stock</th>
                  <th className="pb-3 font-medium text-right hidden lg:table-cell">Min</th>
                  <th className="pb-3 font-medium text-center">Status</th>
                  <th className="pb-3 font-medium text-right">Actions</th>
                </tr>
              </thead>
              <tbody>
                {products.map((product) => {
                  const quantity = getInventoryProductQuantity(product);
                  const minQuantity = getInventoryProductMinQuantity(product);
                  const { isLowStock, isOutOfStock } = classifyInventoryStock(product);

                  return (
                    <tr key={product.id} className="border-b last:border-0">
                      <td className="py-3">
                        <div className="flex items-center gap-3">
                          {product.image_url ? (
                            <div className="relative h-10 w-10 rounded border overflow-hidden flex-shrink-0 bg-muted">
                              <img
                                src={getInventoryProductImageUrl(product.image_url)}
                                alt={product.name}
                                className="h-full w-full object-cover"
                                onError={(event) => {
                                  (event.target as HTMLImageElement).style.display = "none";
                                  ((event.target as HTMLImageElement).nextSibling as HTMLElement).style.display = "flex";
                                }}
                              />
                              <div className="absolute inset-0 hidden items-center justify-center bg-muted">
                                <ImageIcon className="h-4 w-4 text-muted-foreground" />
                              </div>
                            </div>
                          ) : (
                            <div className="h-10 w-10 rounded border flex items-center justify-center bg-muted flex-shrink-0">
                              <Package className="h-4 w-4 text-muted-foreground" />
                            </div>
                          )}

                          <div className="min-w-0">
                            <p className="font-medium truncate">{product.name}</p>
                            <p className="text-xs text-muted-foreground">{product.category_name || "No category"}</p>
                          </div>
                        </div>
                      </td>
                      <td className="py-3 text-sm text-muted-foreground hidden lg:table-cell">{product.sku || "-"}</td>
                      <td className={`py-3 text-right font-medium ${isLowStock ? "text-destructive" : ""}`}>{quantity}</td>
                      <td className="py-3 text-right text-muted-foreground hidden lg:table-cell">{minQuantity}</td>
                      <td className="py-3 text-center">
                        <span
                          className={`inline-flex items-center px-2.5 py-1 rounded-full text-xs font-semibold ${
                            isOutOfStock
                              ? "bg-red-600 text-white dark:bg-red-600/90 dark:text-white"
                              : isLowStock
                                ? "bg-yellow-600 text-white dark:bg-yellow-600/90 dark:text-white"
                                : "bg-green-600 text-white dark:bg-green-600/90 dark:text-white"
                          }`}
                        >
                          {isOutOfStock ? "Out of Stock" : isLowStock ? "Low Stock" : "In Stock"}
                        </span>
                      </td>
                      <td className="py-3 text-right">
                        <div className="flex items-center justify-end gap-2">
                          <Button variant="outline" size="sm" onClick={() => onHistory(product)}>
                            History
                          </Button>
                          {canEditThreshold ? (
                            <Button variant="outline" size="sm" onClick={() => onEditThreshold(product)}>
                              Edit Threshold
                            </Button>
                          ) : null}
                          {canModifyStock ? (
                            <Button variant="outline" size="sm" onClick={() => onAdjust(product)}>
                              <Settings2 className="h-3 w-3 mr-1" />
                              Adjust
                            </Button>
                          ) : null}
                        </div>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        </CardContent>
      </Card>

      <div ref={loadMoreRef} className="h-1 w-full" />
      <div className="py-4 flex items-center justify-center">
        {isFetchingMore ? (
          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            <Loader2 className="h-4 w-4 animate-spin" />
            Loading more inventory...
          </div>
        ) : !hasMore && products.length > 0 ? (
          <p className="text-xs text-muted-foreground">End of inventory list</p>
        ) : null}
      </div>
    </>
  );
}

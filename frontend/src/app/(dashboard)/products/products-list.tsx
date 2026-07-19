import { Archive, ImageIcon, Loader2, Package, Pencil, RotateCcw, Trash2 } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Product } from "@/types";

import { getProductLowStockThreshold, getProductPriceValue, getProductQuantityValue } from "./products-helpers";

interface ProductsListProps {
  canCreate: boolean;
  canEdit: boolean;
  canDelete: boolean;
  products: Product[];
  viewMode: "active" | "archived";
  isLoading: boolean;
  isFetchingMore: boolean;
  hasMore: boolean;
  isSubmitting: boolean;
  loadMoreRef: React.RefObject<HTMLDivElement | null>;
  onCreate: () => void;
  onEdit: (product: Product) => void;
  onArchive: (product: Product) => void;
  onPermanentDelete: (product: Product) => void;
  onRestore: (product: Product) => void;
  formatCurrency: (amount: number) => string;
  getImageUrl: (path: string | null | undefined) => string;
}

export function ProductsList({
  canCreate,
  canEdit,
  canDelete,
  products,
  viewMode,
  isLoading,
  isFetchingMore,
  hasMore,
  isSubmitting,
  loadMoreRef,
  onCreate,
  onEdit,
  onArchive,
  onPermanentDelete,
  onRestore,
  formatCurrency,
  getImageUrl,
}: ProductsListProps) {
  if (isLoading) {
    return (
      <div className="flex h-64 items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  if (products.length === 0) {
    return (
      <Card>
        <CardContent className="flex flex-col items-center justify-center py-12">
          <Package className="mb-4 h-12 w-12 text-muted-foreground" />
          <p className="text-muted-foreground">
            {viewMode === "active" ? "No products found" : "No archived products"}
          </p>
          {canCreate && viewMode === "active" && (
            <Button variant="link" onClick={onCreate}>
              Add your first product
            </Button>
          )}
        </CardContent>
      </Card>
    );
  }

  return (
    <>
      <div className="grid gap-4 lg:hidden">
        {products.map((product) => {
          const quantity = getProductQuantityValue(product);
          const minQuantity = getProductLowStockThreshold(product);
          const isLowStock = quantity <= minQuantity;

          return (
            <Card key={product.id} className="@container">
              <CardContent className="p-4">
                <div className="flex items-start gap-4">
                  {product.image_url ? (
                    <div className="relative h-20 w-20 flex-shrink-0 overflow-hidden rounded border bg-muted">
                      <img
                        src={getImageUrl(product.image_url)}
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
                    <div className="flex h-20 w-20 flex-shrink-0 items-center justify-center rounded border bg-muted">
                      <Package className="h-8 w-8 text-muted-foreground" />
                    </div>
                  )}
                  <div className="min-w-0 flex-1">
                    <div className="mb-2 flex items-start justify-between">
                      <div>
                        <h3 className="break-words font-semibold">{product.name}</h3>
                        <p className="text-sm text-muted-foreground">{product.sku || "-"}</p>
                      </div>
                      <span className="inline-flex items-center rounded-full bg-blue-600 px-2.5 py-1 text-xs font-medium text-white dark:bg-blue-600/90 dark:text-white">
                        {product.category_name || "Uncategorized"}
                      </span>
                    </div>

                    <div className="mb-3 grid grid-cols-2 gap-2 text-sm">
                      <div>
                        <span className="text-muted-foreground">Price: </span>
                        <span className="font-medium font-mono">{formatCurrency(getProductPriceValue(product))}</span>
                      </div>
                      <div>
                        <span className="text-muted-foreground">Stock: </span>
                        <span className={isLowStock ? "font-medium text-destructive" : "font-medium"}>
                          {quantity}
                        </span>
                      </div>
                      <div>
                        <span className="text-muted-foreground">Tax: </span>
                        <span className="font-medium">
                          {product.tax_rate ? `${parseFloat(product.tax_rate)}%` : "0%"}
                        </span>
                      </div>
                    </div>

                    {(canEdit || canDelete) && (
                      <div className="mt-1 flex items-center justify-end gap-2 border-t pt-3">
                        {viewMode === "active" ? (
                          <>
                            {canEdit && (
                              <Button
                                variant="outline"
                                size="sm"
                                className="h-8 w-8 p-0 @[250px]:w-auto @[250px]:px-3"
                                onClick={() => onEdit(product)}
                                title="Edit"
                              >
                                <Pencil className="h-3.5 w-3.5 @[250px]:mr-1" />
                                <span className="hidden @[250px]:inline">Edit</span>
                              </Button>
                            )}
                            {canDelete && (
                              <Button
                                variant="outline"
                                size="sm"
                                className="h-8 w-8 p-0 @[250px]:w-auto @[250px]:px-3 text-amber-600 hover:bg-amber-50 hover:text-amber-700 dark:text-amber-500 dark:hover:bg-amber-950/20"
                                onClick={() => onArchive(product)}
                                title="Archive"
                              >
                                <Archive className="h-3.5 w-3.5 @[250px]:mr-1" />
                                <span className="hidden @[250px]:inline">Archive</span>
                              </Button>
                            )}
                          </>
                        ) : (
                          <>
                            {canEdit && (
                              <Button
                                variant="outline"
                                size="sm"
                                className="h-8 w-8 p-0 @[250px]:w-auto @[250px]:px-3"
                                onClick={() => onRestore(product)}
                                disabled={isSubmitting}
                                title="Restore"
                              >
                                <RotateCcw className="h-3.5 w-3.5 @[250px]:mr-1" />
                                <span className="hidden @[250px]:inline">Restore</span>
                              </Button>
                            )}
                            {canDelete && (
                              <Button
                                variant="outline"
                                size="sm"
                                className="h-8 w-8 p-0 @[250px]:w-auto @[250px]:px-3 text-destructive hover:bg-destructive/10 hover:text-destructive"
                                onClick={() => onPermanentDelete(product)}
                                title="Delete"
                              >
                                <Trash2 className="h-3.5 w-3.5 @[250px]:mr-1" />
                                <span className="hidden @[250px]:inline">Delete</span>
                              </Button>
                            )}
                          </>
                        )}
                      </div>
                    )}
                  </div>
                </div>
              </CardContent>
            </Card>
          );
        })}
      </div>

      <Card className="hidden lg:block">
        <CardContent className="p-6">
          <div className="overflow-x-auto">
            <table className="w-full min-w-max">
              <thead>
                <tr className="border-b text-left text-sm text-muted-foreground">
                  <th className="pb-3 font-medium">Product</th>
                  <th className="pb-3 font-medium">SKU</th>
                  <th className="pb-3 font-medium">Category</th>
                  <th className="pb-3 text-right font-medium">Price</th>
                  <th className="pb-3 text-right font-medium">Tax Rate</th>
                  <th className="pb-3 text-right font-medium">Stock</th>
                  {(canEdit || canDelete) && <th className="pb-3 text-right font-medium">Actions</th>}
                </tr>
              </thead>
              <tbody>
                {products.map((product) => {
                  const quantity = getProductQuantityValue(product);
                  const minQuantity = getProductLowStockThreshold(product);
                  const isLowStock = quantity <= minQuantity;

                  return (
                    <tr key={product.id} className="border-b last:border-0">
                      <td className="py-3">
                        <div className="flex items-center gap-3">
                          {product.image_url ? (
                            <div className="relative h-10 w-10 flex-shrink-0 overflow-hidden rounded border bg-muted">
                              <img
                                src={getImageUrl(product.image_url)}
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
                            <div className="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded border bg-muted">
                              <Package className="h-4 w-4 text-muted-foreground" />
                            </div>
                          )}
                          <div className="min-w-0">
                            <p className="font-medium">{product.name}</p>
                            {product.description && (
                              <p className="max-w-xs truncate text-xs text-muted-foreground">{product.description}</p>
                            )}
                          </div>
                        </div>
                      </td>
                      <td className="py-3 text-sm text-muted-foreground">{product.sku || "-"}</td>
                      <td className="py-3 text-sm">
                        <span className="inline-flex items-center rounded-full bg-blue-600 px-2.5 py-1 text-xs font-medium text-white dark:bg-blue-600/90 dark:text-white">
                          {product.category_name || "Uncategorized"}
                        </span>
                      </td>
                      <td className="py-3 text-right font-medium font-mono">{formatCurrency(getProductPriceValue(product))}</td>
                      <td className="py-3 text-right font-medium text-muted-foreground">
                        {product.tax_rate ? `${parseFloat(product.tax_rate)}%` : "0%"}
                      </td>
                      <td className="py-3 text-right">
                        <span className={isLowStock ? "font-medium text-destructive" : ""}>{quantity}</span>
                      </td>
                      {(canEdit || canDelete) && (
                        <td className="py-3 text-right">
                          <div className="flex items-center justify-end gap-1">
                            {viewMode === "active" ? (
                              <>
                                {canEdit && (
                                  <Button variant="ghost" size="icon" onClick={() => onEdit(product)}>
                                    <Pencil className="h-4 w-4" />
                                  </Button>
                                )}
                                {canDelete && (
                                  <Button
                                    variant="ghost"
                                    size="icon"
                                    onClick={() => onArchive(product)}
                                    title="Archive product"
                                  >
                                    <Archive className="h-4 w-4 text-amber-600 hover:text-amber-700 dark:text-amber-500" />
                                  </Button>
                                )}
                              </>
                            ) : (
                              <>
                                {canEdit && (
                                  <Button
                                    variant="ghost"
                                    size="icon"
                                    onClick={() => onRestore(product)}
                                    disabled={isSubmitting}
                                    title="Restore product"
                                  >
                                    <RotateCcw className="h-4 w-4" />
                                  </Button>
                                )}
                                {canDelete && (
                                  <Button
                                    variant="ghost"
                                    size="icon"
                                    className="text-destructive hover:text-destructive"
                                    onClick={() => onPermanentDelete(product)}
                                    title="Delete product permanently"
                                  >
                                    <Trash2 className="h-4 w-4" />
                                  </Button>
                                )}
                              </>
                            )}
                          </div>
                        </td>
                      )}
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        </CardContent>
      </Card>

      <div ref={loadMoreRef} className="h-1 w-full" />
      <div className="flex items-center justify-center py-4">
        {isFetchingMore ? (
          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            <Loader2 className="h-4 w-4 animate-spin" />
            Loading more products...
          </div>
        ) : !hasMore && products.length > 0 ? (
          <p className="text-xs text-muted-foreground">End of products list</p>
        ) : null}
      </div>
    </>
  );
}

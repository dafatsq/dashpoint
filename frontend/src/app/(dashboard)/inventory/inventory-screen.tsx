'use client';

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { AlertTriangle, Boxes, Loader2, Package, TrendingDown } from "lucide-react";

import { Header } from "@/components/layout/header";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useAuth, PERMISSIONS } from "@/contexts/auth-context";
import api from "@/lib/api";
import type { LowStockItem, Product } from "@/types";

import { InventoryAdjustDialog } from "./inventory-adjust-dialog";
import { InventoryControls } from "./inventory-controls";
import {
  buildInventoryAdjustmentRequest,
  createEmptyAdjustmentFormState,
  getInventoryProductPrice,
  getInventoryProductQuantity,
  getPermittedInventoryActions,
  type AdjustmentFormState,
} from "./inventory-helpers";
import { InventoryList } from "./inventory-list";
import { InventoryLowStock } from "./inventory-low-stock";

export function InventoryScreen() {
  const PAGE_SIZE = 24;
  const { hasPermission } = useAuth();
  const canAddStock = hasPermission(PERMISSIONS.INVENTORY_ADD_STOCK);
  const canRemoveStock = hasPermission(PERMISSIONS.INVENTORY_REMOVE_STOCK);
  const canAdjustStock = hasPermission(PERMISSIONS.INVENTORY_ADJUST_STOCK);
  const canModifyStock = canAddStock || canRemoveStock || canAdjustStock;

  const [products, setProducts] = useState<Product[]>([]);
  const [lowStockItems, setLowStockItems] = useState<LowStockItem[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [activeTab, setActiveTab] = useState<"all" | "low-stock">("all");
  const [pageError, setPageError] = useState<string | null>(null);
  const [adjustmentError, setAdjustmentError] = useState<string | null>(null);
  const [lastRefreshed, setLastRefreshed] = useState<Date>(new Date());
  const [page, setPage] = useState(1);
  const [totalProducts, setTotalProducts] = useState(0);
  const [hasMore, setHasMore] = useState(true);
  const [isFetchingMore, setIsFetchingMore] = useState(false);
  const loadMoreRef = useRef<HTMLDivElement | null>(null);

  const [adjustDialogOpen, setAdjustDialogOpen] = useState(false);
  const [selectedProduct, setSelectedProduct] = useState<Product | null>(null);
  const [adjustmentForm, setAdjustmentForm] = useState<AdjustmentFormState>(createEmptyAdjustmentFormState("add"));
  const [isSubmitting, setIsSubmitting] = useState(false);

  const allowedActions = useMemo(
    () => getPermittedInventoryActions({ canAddStock, canRemoveStock, canAdjustStock }),
    [canAddStock, canAdjustStock, canRemoveStock],
  );

  useEffect(() => {
    const timeout = setTimeout(() => setDebouncedSearch(searchQuery), 300);
    return () => clearTimeout(timeout);
  }, [searchQuery]);

  const fetchLowStock = useCallback(async () => {
    const result = await api.getLowStock();
    if (result.error) {
      setPageError(result.error);
      return;
    }
    setPageError(null);
    setLowStockItems(result.data || []);
    setLastRefreshed(new Date());
  }, []);

  const fetchProductsPage = useCallback(
    async (pageToLoad: number, replace: boolean) => {
      if (replace) {
        setIsLoading(true);
      } else {
        setIsFetchingMore(true);
      }

      const result = await api.getProductsPage({
        active: true,
        search: debouncedSearch || undefined,
        page: pageToLoad,
        per_page: PAGE_SIZE,
      });

      if (result.error) {
        setPageError(result.error);
      } else {
        setPageError(null);
        setProducts((prev) => (replace ? result.data || [] : [...prev, ...(result.data || [])]));
        setTotalProducts(result.total || 0);
        if (result.total_pages !== undefined) {
          setHasMore(pageToLoad < result.total_pages);
        } else {
          setHasMore((result.data || []).length === PAGE_SIZE);
        }
        setPage(pageToLoad);
      }

      if (replace) {
        setIsLoading(false);
      } else {
        setIsFetchingMore(false);
      }
    },
    [debouncedSearch],
  );

  useEffect(() => {
    const timeout = window.setTimeout(() => {
      void fetchProductsPage(1, true);
    }, 0);
    return () => window.clearTimeout(timeout);
  }, [fetchProductsPage]);

  useEffect(() => {
    const timeout = window.setTimeout(() => {
      void fetchLowStock();
    }, 0);
    const interval = setInterval(() => {
      void fetchLowStock();
    }, 10000);
    return () => {
      window.clearTimeout(timeout);
      clearInterval(interval);
    };
  }, [fetchLowStock]);

  useEffect(() => {
    const node = loadMoreRef.current;
    if (!node || activeTab !== "all" || !hasMore || isLoading || isFetchingMore) {
      return;
    }

    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting) {
          void fetchProductsPage(page + 1, false);
        }
      },
      { rootMargin: "200px" },
    );

    observer.observe(node);
    return () => observer.disconnect();
  }, [activeTab, fetchProductsPage, hasMore, isFetchingMore, isLoading, page]);

  const openAdjustDialog = useCallback(
    (product: Product) => {
      if (!canModifyStock || allowedActions.length === 0) {
        return;
      }

      const defaultAction = allowedActions[0];
      setSelectedProduct(product);
      setAdjustmentForm(createEmptyAdjustmentFormState(defaultAction));
      setAdjustmentError(null);
      setAdjustDialogOpen(true);
    },
    [allowedActions, canModifyStock],
  );

  const openAdjustDialogFromLowStock = useCallback(
    async (productId: string) => {
      const existing = products.find((product) => product.id === productId);
      if (existing) {
        openAdjustDialog(existing);
        return;
      }

      const result = await api.getProduct(productId);
      if (result.error) {
        setPageError(result.error);
        return;
      }
      if (result.data) {
        openAdjustDialog(result.data);
      }
    },
    [openAdjustDialog, products],
  );

  const handleAdjustment = useCallback(async () => {
    if (!selectedProduct || !adjustmentForm.quantity) {
      return;
    }

    const inputQuantity = Number.parseInt(adjustmentForm.quantity, 10);
    const currentStock = getInventoryProductQuantity(selectedProduct);

    if (Number.isNaN(inputQuantity) || inputQuantity <= 0) {
      setAdjustmentError("Quantity must be greater than zero");
      return;
    }
    if (adjustmentForm.adjustmentType === "count" && inputQuantity < 0) {
      setAdjustmentError("Stock quantity cannot be negative");
      return;
    }
    if (
      adjustmentForm.adjustmentType === "adjustment" &&
      adjustmentForm.action === "remove" &&
      currentStock-inputQuantity < 0
    ) {
      setAdjustmentError(`Cannot remove ${inputQuantity} items. Only ${currentStock} available.`);
      return;
    }
    if (
      (adjustmentForm.adjustmentType === "damage" || adjustmentForm.adjustmentType === "loss") &&
      inputQuantity > currentStock
    ) {
      setAdjustmentError(`Cannot remove ${inputQuantity} items. Only ${currentStock} available.`);
      return;
    }

    setIsSubmitting(true);
    setAdjustmentError(null);

    const result = await api.adjustInventory(
      buildInventoryAdjustmentRequest({
        productId: selectedProduct.id,
        action: adjustmentForm.action,
        adjustmentType: adjustmentForm.adjustmentType,
        quantity: adjustmentForm.quantity,
        currentStock,
        notes: adjustmentForm.notes,
      }),
    );

    if (result.error) {
      setAdjustmentError(result.error);
      setIsSubmitting(false);
      return;
    }

    setProducts([]);
    setPage(1);
    setHasMore(true);
    await Promise.all([fetchProductsPage(1, true), fetchLowStock()]);
    setAdjustDialogOpen(false);
    setSelectedProduct(null);
    setIsSubmitting(false);
  }, [adjustmentForm, fetchLowStock, fetchProductsPage, selectedProduct]);

  const totalStock = products.reduce((sum, product) => sum + getInventoryProductQuantity(product), 0);
  const totalValue = products.reduce(
    (sum, product) => sum + getInventoryProductPrice(product) * getInventoryProductQuantity(product),
    0,
  );
  const lowStockCount = lowStockItems.length;

  const formatCurrency = (amount: number) =>
    new Intl.NumberFormat("id-ID", {
      style: "currency",
      currency: "IDR",
      minimumFractionDigits: 0,
    }).format(amount);

  return (
    <div className="flex flex-col h-full">
      <Header title="Inventory" />

      <div className="flex-1 p-6 overflow-auto">
        <div className="grid gap-4 grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 mb-6">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium">Total Products</CardTitle>
              <Package className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{totalProducts}</div>
              <p className="text-xs text-muted-foreground">Loaded {products.length}</p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium">Loaded Stock</CardTitle>
              <Boxes className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{totalStock.toLocaleString()}</div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium">Loaded Inventory Value</CardTitle>
              <TrendingDown className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-xl sm:text-2xl font-bold truncate" title={formatCurrency(totalValue)}>
                {formatCurrency(totalValue)}
              </div>
            </CardContent>
          </Card>

          <Card className={lowStockCount > 0 ? "border-yellow-500" : ""}>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium">Low Stock</CardTitle>
              <AlertTriangle className={`h-4 w-4 ${lowStockCount > 0 ? "text-yellow-500" : "text-muted-foreground"}`} />
            </CardHeader>
            <CardContent>
              <div className={`text-2xl font-bold ${lowStockCount > 0 ? "text-yellow-600" : ""}`}>{lowStockCount}</div>
            </CardContent>
          </Card>
        </div>

        <InventoryControls
          searchQuery={searchQuery}
          activeTab={activeTab}
          lowStockCount={lowStockCount}
          onSearchChange={setSearchQuery}
          onTabChange={setActiveTab}
        />

        <div className="flex justify-between items-center mb-4 gap-3">
          {pageError ? <div className="rounded-md border border-destructive/20 bg-destructive/10 p-3 text-sm text-destructive">{pageError}</div> : <div />}
          <div className="text-xs text-muted-foreground whitespace-nowrap">
            Last updated: {lastRefreshed.toLocaleTimeString()} • Auto-refreshes every 10s
          </div>
        </div>

        {isLoading ? (
          <div className="flex items-center justify-center h-64">
            <Loader2 className="h-8 w-8 animate-spin text-primary" />
          </div>
        ) : activeTab === "low-stock" ? (
          <InventoryLowStock items={lowStockItems} products={products} canModifyStock={canModifyStock} onAdjust={(id) => void openAdjustDialogFromLowStock(id)} />
        ) : (
          <InventoryList
            products={products}
            totalProducts={totalProducts}
            hasMore={hasMore}
            isFetchingMore={isFetchingMore}
            canModifyStock={canModifyStock}
            loadMoreRef={loadMoreRef}
            onAdjust={openAdjustDialog}
          />
        )}
      </div>

      <InventoryAdjustDialog
        open={adjustDialogOpen}
        product={selectedProduct}
        formState={adjustmentForm}
        allowedActions={allowedActions}
        error={adjustmentError}
        isSubmitting={isSubmitting}
        onOpenChange={(open) => {
          setAdjustDialogOpen(open);
          if (!open) {
            setSelectedProduct(null);
            setAdjustmentError(null);
          }
        }}
        onActionChange={(action) =>
          setAdjustmentForm({
            ...createEmptyAdjustmentFormState(action),
            notes: adjustmentForm.notes,
          })
        }
        onFormStateChange={setAdjustmentForm}
        onSubmit={() => void handleAdjustment()}
      />
    </div>
  );
}

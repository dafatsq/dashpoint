'use client';

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { AlertTriangle, Boxes, Loader2, Package, TrendingDown } from "lucide-react";

import { Header } from "@/components/layout/header";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useAuth, PERMISSIONS } from "@/contexts/auth-context";
import { useGlobalError } from "@/contexts/error-context";
import api from "@/lib/api";
import type { Category, LowStockItem, Product } from "@/types";

import { InventoryAdjustDialog } from "./inventory-adjust-dialog";
import { InventoryControls } from "./inventory-controls";
import {
  buildInventoryAdjustmentRequest,
  canSubmitInventoryAdjustment,
  createEmptyAdjustmentFormState,
  getInventoryProductPrice,
  getInventoryProductQuantity,
  getPermittedInventoryActions,
  isInventoryActionAllowed,
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
  const [categories, setCategories] = useState<Category[]>([]);
  const [lowStockItems, setLowStockItems] = useState<LowStockItem[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [selectedCategory, setSelectedCategory] = useState<string>("all");
  const [activeTab, setActiveTab] = useState<"all" | "low-stock">("all");
  const [pageError, setPageError] = useState<string | null>(null);
  const { showError } = useGlobalError();
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

  const loadCategories = useCallback(async () => {
    const result = await api.getCategories();
    if (result.data) {
      setCategories(result.data);
    }
  }, []);

  useEffect(() => {
    const timeoutId = window.setTimeout(() => {
      void loadCategories();
    }, 0);

    return () => window.clearTimeout(timeoutId);
  }, [loadCategories]);

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
    async (pageToLoad: number, replace: boolean, silent = false) => {
      if (replace && !silent) {
        setIsLoading(true);
      } else if (!replace && !silent) {
        setIsFetchingMore(true);
      }

      const fetchPage = silent ? 1 : pageToLoad;
      const fetchPerPage = silent ? PAGE_SIZE * pageToLoad : PAGE_SIZE;

      const result = await api.getProductsPage({
        active: true,
        search: debouncedSearch || undefined,
        category_id: selectedCategory !== "all" ? selectedCategory : undefined,
        page: fetchPage,
        per_page: fetchPerPage,
      });

      if (result.error) {
        if (!silent) setPageError(result.error);
      } else {
        if (!silent) setPageError(null);
        setProducts((prev) => {
          if (silent || replace) return result.data || [];
          return [...prev, ...(result.data || [])];
        });
        setTotalProducts(result.total || 0);

        const realTotalPages = Math.ceil((result.total || 0) / PAGE_SIZE);
        setHasMore(pageToLoad < realTotalPages);

        if (!silent) setPage(pageToLoad);
        setLastRefreshed(new Date());
      }

      if (replace && !silent) {
        setIsLoading(false);
      } else if (!replace && !silent) {
        setIsFetchingMore(false);
      }
    },
    [debouncedSearch, selectedCategory],
  );

  const resetProductPagination = useCallback(() => {
    setProducts([]);
    setPage(1);
    setHasMore(true);
  }, []);

  useEffect(() => {
    let currentPage = 1;
    const timeout = window.setTimeout(() => {
      void fetchProductsPage(currentPage, true);
    }, 0);
    const interval = setInterval(() => {
      setPage((p) => {
        currentPage = p;
        return p;
      });
      void fetchProductsPage(currentPage, false, true);
    }, 5000);
    return () => {
      window.clearTimeout(timeout);
      clearInterval(interval);
    };
  }, [fetchProductsPage]);

  useEffect(() => {
    const timeout = window.setTimeout(() => {
      void fetchLowStock();
    }, 0);
    const interval = setInterval(() => {
      void fetchLowStock();
    }, 5000);
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
    if (!selectedProduct) {
      showError("Product Required", "Select a product before saving.");
      return;
    }

    if (isSubmitting) {
      return;
    }

    if (allowedActions.length === 0) {
      showError("Action Denied", "You do not have any inventory actions available.");
      return;
    }

    if (!adjustmentForm.quantity) {
      showError("Quantity Required", "Enter a quantity before saving.");
      return;
    }

    if (
      !canSubmitInventoryAdjustment({
        allowedActions,
        action: adjustmentForm.action,
        quantity: adjustmentForm.quantity,
        isSubmitting,
      })
    ) {
      return;
    }

    if (!isInventoryActionAllowed(adjustmentForm.action, allowedActions)) {
      showError("Permission Denied", "You do not have permission for that inventory action.");
      return;
    }

    const inputQuantity = Number.parseInt(adjustmentForm.quantity, 10);
    const currentStock = getInventoryProductQuantity(selectedProduct);

    if (Number.isNaN(inputQuantity) || inputQuantity <= 0) {
      showError("Invalid Quantity", "Quantity must be greater than zero");
      return;
    }
    if (adjustmentForm.adjustmentType === "count" && inputQuantity < 0) {
      showError("Invalid Quantity", "Stock quantity cannot be negative");
      return;
    }
    if (
      adjustmentForm.adjustmentType === "adjustment" &&
      adjustmentForm.action === "remove" &&
      currentStock - inputQuantity < 0
    ) {
      showError("Insufficient Stock", `Cannot remove ${inputQuantity} items. Only ${currentStock} available.`);
      return;
    }
    if (
      (adjustmentForm.adjustmentType === "damage" || adjustmentForm.adjustmentType === "loss") &&
      inputQuantity > currentStock
    ) {
      showError("Insufficient Stock", `Cannot remove ${inputQuantity} items. Only ${currentStock} available.`);
      return;
    }

    setIsSubmitting(true);

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
      showError("Adjustment Failed", result.error);
      setIsSubmitting(false);
      return;
    }

    resetProductPagination();
    await Promise.all([fetchProductsPage(1, true), fetchLowStock()]);
    setAdjustDialogOpen(false);
    setSelectedProduct(null);
    setIsSubmitting(false);
  }, [
    adjustmentForm,
    allowedActions,
    fetchLowStock,
    fetchProductsPage,
    isSubmitting,
    resetProductPagination,
    selectedProduct,
    showError,
  ]);

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
    <div className="flex h-full flex-col">
      <Header title="Inventory" />

      <div className="flex-1 overflow-auto p-6">
        <div className="mb-6 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
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
              <div className="truncate text-xl font-bold sm:text-2xl" title={formatCurrency(totalValue)}>
                {formatCurrency(totalValue)}
              </div>
            </CardContent>
          </Card>

          <Card className={lowStockCount > 0 ? "border-yellow-500" : ""}>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium">Low Stock</CardTitle>
              <AlertTriangle
                className={`h-4 w-4 ${lowStockCount > 0 ? "text-yellow-500" : "text-muted-foreground"}`}
              />
            </CardHeader>
            <CardContent>
              <div className={`text-2xl font-bold ${lowStockCount > 0 ? "text-yellow-600" : ""}`}>
                {lowStockCount}
              </div>
            </CardContent>
          </Card>
        </div>

        <InventoryControls
          searchQuery={searchQuery}
          activeTab={activeTab}
          lowStockCount={lowStockCount}
          categories={categories}
          selectedCategory={selectedCategory}
          onSearchChange={setSearchQuery}
          onTabChange={setActiveTab}
          onCategoryChange={setSelectedCategory}
        />

        <div className="mb-4 flex items-center justify-between gap-3">
          {pageError ? (
            <div className="rounded-md border border-destructive/20 bg-destructive/10 p-3 text-sm text-destructive">
              {pageError}
            </div>
          ) : (
            <div />
          )}
          <div className="whitespace-nowrap text-xs text-muted-foreground">
            Last updated: {lastRefreshed.toLocaleTimeString()} • Auto-refreshes every 5s
          </div>
        </div>

        {isLoading ? (
          <div className="flex h-64 items-center justify-center">
            <Loader2 className="h-8 w-8 animate-spin text-primary" />
          </div>
        ) : activeTab === "low-stock" ? (
          <InventoryLowStock
            items={lowStockItems}
            products={products}
            canModifyStock={canModifyStock}
            onAdjust={(id) => void openAdjustDialogFromLowStock(id)}
          />
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
        isSubmitting={isSubmitting}
        onOpenChange={(open) => {
          setAdjustDialogOpen(open);
          if (!open) {
            setSelectedProduct(null);
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

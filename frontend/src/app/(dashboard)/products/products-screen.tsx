'use client';

import { useCallback, useEffect, useMemo, useRef, useState } from "react";

import { Header } from "@/components/layout/header";
import api from "@/lib/api";
import { buildBackendUrl } from "@/lib/config";
import { useAuth, PERMISSIONS } from "@/contexts/auth-context";
import { useGlobalError } from "@/contexts/error-context";
import { Category, Product } from "@/types";

import { ProductsControls } from "./products-controls";
import { ConfirmationDialog } from "@/components/shared/confirmation-dialog";
import { ProductsFormDialog } from "./products-form-dialog";
import {
  buildProductCreateRequest,
  buildProductUpdateRequest,
  createEmptyProductFormData,
  getProductDisplayImageUrl,
  hasProductFormChanges,
  mapProductToFormData,
  type ProductFormData,
} from "./products-helpers";
import { ProductsList } from "./products-list";
import { parseSortValue } from "@/lib/sorting";

export function ProductsScreen() {
  const PAGE_SIZE = 24;
  const { hasPermission } = useAuth();
  const canCreate = hasPermission(PERMISSIONS.PRODUCTS_CREATE);
  const canEdit = hasPermission(PERMISSIONS.PRODUCTS_EDIT);
  const canDelete = hasPermission(PERMISSIONS.PRODUCTS_DELETE);

  const [products, setProducts] = useState<Product[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [categoryError, setCategoryError] = useState<string | null>(null);
  const [productsError, setProductsError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [selectedCategory, setSelectedCategory] = useState<string>("all");
  const [sort, setSort] = useState("name_asc");
  const [viewMode, setViewMode] = useState<"active" | "archived">("active");
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);
  const [isFetchingMore, setIsFetchingMore] = useState(false);
  const loadMoreRef = useRef<HTMLDivElement | null>(null);

  const [dialogOpen, setDialogOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [permanentDeleteDialogOpen, setPermanentDeleteDialogOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [editingProduct, setEditingProduct] = useState<Product | null>(null);
  const [deletingProduct, setDeletingProduct] = useState<Product | null>(null);
  const { showError } = useGlobalError();
  const [formData, setFormData] = useState<ProductFormData>(createEmptyProductFormData());
  const [formErrors, setFormErrors] = useState<{
    name?: string;
    price?: string;
    sku?: string;
    barcode?: string;
    general?: string;
  }>({});

  useEffect(() => {
    const timeout = setTimeout(() => {
      setDebouncedSearch(searchQuery);
    }, 300);
    return () => clearTimeout(timeout);
  }, [searchQuery]);

  const loadCategories = useCallback(async () => {
    const result = await api.getCategories();
    if (result.error) {
      setCategoryError(result.error);
      return;
    }
    if (result.data) {
      setCategoryError(null);
      setCategories(result.data);
    }
  }, []);

  useEffect(() => {
    const timeoutId = window.setTimeout(() => {
      void loadCategories();
    }, 0);

    return () => window.clearTimeout(timeoutId);
  }, [loadCategories]);

  const fetchProductsPage = useCallback(
    async (pageToLoad: number, replace: boolean) => {
      if (replace) {
        setIsLoading(true);
      } else {
        setIsFetchingMore(true);
      }

      const result = await api.getProductsPage({
        active: viewMode === "active",
        category_id: selectedCategory !== "all" ? selectedCategory : undefined,
        search: debouncedSearch || undefined,
        page: pageToLoad,
        per_page: PAGE_SIZE,
        ...parseSortValue(sort),
      });

      if (result.error) {
        setProductsError(result.error);
      } else if (result.data) {
        setProductsError(null);
        setProducts((prev) => (replace ? result.data! : [...prev, ...result.data!]));
        if (result.total_pages !== undefined) {
          setHasMore(pageToLoad < result.total_pages);
        } else {
          setHasMore(result.data.length === PAGE_SIZE);
        }
        setPage(pageToLoad);
      }

      if (replace) {
        setIsLoading(false);
      } else {
        setIsFetchingMore(false);
      }
    },
    [debouncedSearch, selectedCategory, sort, viewMode],
  );

  const resetPagination = useCallback(() => {
    setProducts([]);
    setPage(1);
    setHasMore(true);
  }, []);

  useEffect(() => {
    resetPagination();

    const timeoutId = window.setTimeout(() => {
      void fetchProductsPage(1, true);
    }, 0);

    return () => window.clearTimeout(timeoutId);
  }, [fetchProductsPage, resetPagination]);

  useEffect(() => {
    const node = loadMoreRef.current;
    if (!node || !hasMore || isLoading || isFetchingMore) {
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
  }, [fetchProductsPage, hasMore, isFetchingMore, isLoading, page]);

  const resetForm = useCallback(() => {
    setFormData(createEmptyProductFormData());
    setEditingProduct(null);
    setFormErrors({});
  }, []);

  const openCreateDialog = useCallback(() => {
    resetForm();
    setDialogOpen(true);
  }, [resetForm]);

  const openEditDialog = useCallback((product: Product) => {
    setEditingProduct(product);
    setFormData(mapProductToFormData(product));
    setFormErrors({});
    setDialogOpen(true);
  }, []);

  const hasChanges = useMemo(() => hasProductFormChanges(formData, editingProduct), [editingProduct, formData]);

  const handleSubmit = async () => {
    if (editingProduct && !canEdit) {
      showError("Permission Denied", "You do not have permission to edit products");
      return;
    }
    if (!editingProduct && !canCreate) {
      showError("Permission Denied", "You do not have permission to create products");
      return;
    }
    if (editingProduct && !hasChanges) {
      showError("No Changes", "Make a change before saving.");
      return;
    }

    if (!formData.name.trim()) {
      showError("Name Required", "Product name is required");
      return;
    }
    if (!formData.price || parseFloat(formData.price) <= 0) {
      showError("Price Required", "Valid price is required");
      return;
    }

    setFormErrors({});
    setIsSubmitting(true);

    try {
      if (editingProduct) {
        const result = await api.updateProduct(editingProduct.id, {
          ...buildProductUpdateRequest(formData),
          expected_updated_at: editingProduct.updated_at,
        });
        if (result.error) {
          if (result.error.includes("SKU")) {
            showError("SKU Error", result.error);
          } else if (result.error.includes("Barcode") || result.error.includes("barcode")) {
            showError("Barcode Error", result.error);
          } else {
            showError("Save Failed", result.error);
          }
          return;
        }
        if (result.data) {
          setProducts((prev) => prev.map((product) => (product.id === editingProduct.id ? result.data! : product)));
        }
      } else {
        const result = await api.createProduct(buildProductCreateRequest(formData));
        if (result.error) {
          if (result.error.includes("SKU")) {
            showError("SKU Error", result.error);
          } else if (result.error.includes("Barcode") || result.error.includes("barcode")) {
            showError("Barcode Error", result.error);
          } else {
            showError("Create Failed", result.error);
          }
          return;
        }
        if (result.data) {
          setProducts((prev) => [result.data!, ...prev]);
        }
      }

      setDialogOpen(false);
      resetForm();
      void fetchProductsPage(1, true);
    } catch {
      showError("Save Error", "Failed to save product. Please try again.");
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleDelete = async () => {
    if (!deletingProduct) {
      return;
    }
    if (!canDelete) {
      showError("Permission Denied", "You do not have permission to archive products");
      return;
    }
    setIsSubmitting(true);
    const result = await api.deleteProduct(deletingProduct.id, deletingProduct.updated_at);
    if (result.error) {
      showError("Archive Failed", result.error);
    } else {
      setProductsError(null);
      setProducts((prev) => prev.filter((product) => product.id !== deletingProduct.id));
      setDeleteDialogOpen(false);
      setDeletingProduct(null);
      void fetchProductsPage(1, true);
    }
    setIsSubmitting(false);
  };

  const handleRestore = async (product: Product) => {
    if (!canEdit) {
      showError("Permission Denied", "You do not have permission to restore products");
      return;
    }
    setIsSubmitting(true);
    const result = await api.updateProduct(product.id, {
      is_active: true,
      expected_updated_at: product.updated_at,
    });
    if (result.error) {
      showError("Restore Failed", result.error);
    } else if (result.data) {
      setProductsError(null);
      setProducts((prev) => prev.filter((item) => item.id !== product.id));
      void fetchProductsPage(1, true);
    }
    setIsSubmitting(false);
  };

  const handlePermanentDelete = async () => {
    if (!deletingProduct) {
      return;
    }
    if (!canDelete) {
      showError("Permission Denied", "You do not have permission to delete products");
      return;
    }
    setIsSubmitting(true);
    const result = await api.permanentDeleteProduct(
      deletingProduct.id,
      deletingProduct.updated_at,
    );
    if (result.error) {
      showError("Delete Failed", result.error);
    } else {
      setProductsError(null);
      setProducts((prev) => prev.filter((product) => product.id !== deletingProduct.id));
      setPermanentDeleteDialogOpen(false);
      setDeletingProduct(null);
      void fetchProductsPage(1, true);
    }
    setIsSubmitting(false);
  };

  const formatCurrency = (amount: number) =>
    new Intl.NumberFormat("id-ID", {
      style: "currency",
      currency: "IDR",
      minimumFractionDigits: 0,
    }).format(amount);

  const getImageUrl = (path: string | null | undefined) => getProductDisplayImageUrl(path, buildBackendUrl);
  const pageError = categoryError || productsError;

  return (
    <div className="flex h-full flex-col">
      <Header title="Products" />

      <div className="flex-1 overflow-auto p-6">
        {pageError && (
          <div className="mb-6 rounded-md border border-destructive/20 bg-destructive/10 p-3">
            <p className="text-sm text-destructive">{pageError}</p>
          </div>
        )}

        <ProductsControls
          canCreate={canCreate}
          categories={categories}
          searchQuery={searchQuery}
          selectedCategory={selectedCategory}
          sort={sort}
          viewMode={viewMode}
          onCreate={openCreateDialog}
          onSearchChange={setSearchQuery}
          onCategoryChange={setSelectedCategory}
          onSortChange={setSort}
          onViewModeChange={setViewMode}
        />

        <ProductsList
          canCreate={canCreate}
          canEdit={canEdit}
          canDelete={canDelete}
          products={products}
          viewMode={viewMode}
          isLoading={isLoading}
          isFetchingMore={isFetchingMore}
          hasMore={hasMore}
          isSubmitting={isSubmitting}
          loadMoreRef={loadMoreRef}
          onCreate={openCreateDialog}
          onEdit={openEditDialog}
          onArchive={(product) => {
            setDeletingProduct(product);
            setDeleteDialogOpen(true);
          }}
          onPermanentDelete={(product) => {
            setDeletingProduct(product);
            setPermanentDeleteDialogOpen(true);
          }}
          onRestore={handleRestore}
          formatCurrency={formatCurrency}
          getImageUrl={getImageUrl}
        />
      </div>

      <ProductsFormDialog
        open={dialogOpen}
        editingProduct={editingProduct}
        categories={categories}
        formData={formData}
        formErrors={formErrors}
        isSubmitting={isSubmitting}
        hasChanges={hasChanges}
        onOpenChange={setDialogOpen}
        onFormDataChange={setFormData}
        onFormErrorsChange={setFormErrors}
        onSubmit={handleSubmit}
      />

      <ConfirmationDialog
        open={deleteDialogOpen}
        onOpenChange={setDeleteDialogOpen}
        title="Archive Product"
        description={`Are you sure you want to archive "${deletingProduct?.name}"? The product will be moved to the Archived tab and can be restored later.`}
        confirmText="Archive"
        isSubmitting={isSubmitting}
        loadingText="Archiving..."
        onConfirm={handleDelete}
      />

      <ConfirmationDialog
        open={permanentDeleteDialogOpen}
        onOpenChange={setPermanentDeleteDialogOpen}
        title="Permanently Delete Product"
        description={`Are you sure you want to permanently delete "${deletingProduct?.name}"? This action cannot be undone. All data associated with this product will be lost.`}
        confirmText="Delete Permanently"
        confirmVariant="destructive"
        isSubmitting={isSubmitting}
        loadingText="Deleting..."
        onConfirm={handlePermanentDelete}
      />
    </div>
  );
}

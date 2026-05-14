'use client';

import { useCallback, useEffect, useMemo, useRef, useState } from "react";

import { Header } from "@/components/layout/header";
import api from "@/lib/api";
import { buildBackendUrl } from "@/lib/config";
import { useAuth, PERMISSIONS } from "@/contexts/auth-context";
import { Category, Product } from "@/types";

import { ProductsControls } from "./products-controls";
import { ProductsDeleteDialogs } from "./products-delete-dialogs";
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
  const [viewMode, setViewMode] = useState<"active" | "archived">("active");
  const [page, setPage] = useState(1);
  const [totalProducts, setTotalProducts] = useState(0);
  const [hasMore, setHasMore] = useState(true);
  const [isFetchingMore, setIsFetchingMore] = useState(false);
  const loadMoreRef = useRef<HTMLDivElement | null>(null);

  const [dialogOpen, setDialogOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [permanentDeleteDialogOpen, setPermanentDeleteDialogOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [editingProduct, setEditingProduct] = useState<Product | null>(null);
  const [deletingProduct, setDeletingProduct] = useState<Product | null>(null);
  const [deleteError, setDeleteError] = useState<string | null>(null);
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

  useEffect(() => {
    const fetchCategories = async () => {
      const result = await api.getCategories();
      if (result.error) {
        setCategoryError(result.error);
        return;
      }
      if (result.data) {
        setCategoryError(null);
        setCategories(result.data);
      }
    };

    void fetchCategories();
  }, []);

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
      });

      if (result.error) {
        setProductsError(result.error);
      } else if (result.data) {
        setProductsError(null);
        setProducts((prev) => (replace ? result.data! : [...prev, ...result.data!]));
        setTotalProducts(result.total || 0);
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
    [debouncedSearch, selectedCategory, viewMode],
  );

  useEffect(() => {
    setProducts([]);
    setPage(1);
    setHasMore(true);
    void fetchProductsPage(1, true);
  }, [fetchProductsPage]);

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
    const errors: typeof formErrors = {};
    if (!formData.name.trim()) {
      errors.name = "Product name is required";
    }
    if (!formData.price || parseFloat(formData.price) <= 0) {
      errors.price = "Valid price is required";
    }
    if (Object.keys(errors).length > 0) {
      setFormErrors(errors);
      return;
    }

    setFormErrors({});
    setIsSubmitting(true);

    try {
      if (editingProduct) {
        const result = await api.updateProduct(editingProduct.id, buildProductUpdateRequest(formData));
        if (result.error) {
          if (result.error.includes("SKU")) {
            setFormErrors({ sku: result.error });
          } else if (result.error.includes("Barcode") || result.error.includes("barcode")) {
            setFormErrors({ barcode: result.error });
          } else {
            setFormErrors({ general: result.error });
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
            setFormErrors({ sku: result.error });
          } else if (result.error.includes("Barcode") || result.error.includes("barcode")) {
            setFormErrors({ barcode: result.error });
          } else {
            setFormErrors({ general: result.error });
          }
          return;
        }
        if (result.data) {
          setProducts((prev) => [result.data!, ...prev]);
          setTotalProducts((prev) => prev + 1);
        }
      }

      setDialogOpen(false);
      resetForm();
      void fetchProductsPage(1, true);
    } catch {
      setFormErrors({ general: "Failed to save product. Please try again." });
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleDelete = async () => {
    if (!deletingProduct) {
      return;
    }
    setIsSubmitting(true);
    const result = await api.deleteProduct(deletingProduct.id);
    if (result.error) {
      setProductsError(result.error);
    } else {
      setProductsError(null);
      setProducts((prev) => prev.filter((product) => product.id !== deletingProduct.id));
      setTotalProducts((prev) => Math.max(0, prev - 1));
      setDeleteDialogOpen(false);
      setDeletingProduct(null);
      void fetchProductsPage(1, true);
    }
    setIsSubmitting(false);
  };

  const handleRestore = async (product: Product) => {
    setIsSubmitting(true);
    const result = await api.updateProduct(product.id, { is_active: true });
    if (result.error) {
      setProductsError(result.error);
    } else if (result.data) {
      setProductsError(null);
      setProducts((prev) => prev.filter((item) => item.id !== product.id));
      setTotalProducts((prev) => Math.max(0, prev - 1));
      void fetchProductsPage(1, true);
    }
    setIsSubmitting(false);
  };

  const handlePermanentDelete = async () => {
    if (!deletingProduct) {
      return;
    }
    setIsSubmitting(true);
    setDeleteError(null);
    const result = await api.permanentDeleteProduct(deletingProduct.id);
    if (result.error) {
      setDeleteError(result.error);
    } else {
      setProductsError(null);
      setProducts((prev) => prev.filter((product) => product.id !== deletingProduct.id));
      setTotalProducts((prev) => Math.max(0, prev - 1));
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
          viewMode={viewMode}
          onCreate={openCreateDialog}
          onSearchChange={setSearchQuery}
          onCategoryChange={setSelectedCategory}
          onViewModeChange={setViewMode}
        />

        <ProductsList
          canCreate={canCreate}
          canEdit={canEdit}
          canDelete={canDelete}
          products={products}
          totalProducts={totalProducts}
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

      <ProductsDeleteDialogs
        deletingProduct={deletingProduct}
        deleteDialogOpen={deleteDialogOpen}
        permanentDeleteDialogOpen={permanentDeleteDialogOpen}
        isSubmitting={isSubmitting}
        deleteError={deleteError}
        onDeleteDialogOpenChange={setDeleteDialogOpen}
        onPermanentDeleteDialogOpenChange={setPermanentDeleteDialogOpen}
        onArchive={handleDelete}
        onPermanentDelete={handlePermanentDelete}
        onClearDeleteError={() => setDeleteError(null)}
      />
    </div>
  );
}

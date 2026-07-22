'use client';

import { useCallback, useEffect, useMemo, useState } from "react";
import { ShieldAlert } from "lucide-react";

import { Header } from "@/components/layout/header";
import { Card, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { useAuth, PERMISSIONS } from "@/contexts/auth-context";
import { useGlobalError } from "@/contexts/error-context";
import api from "@/lib/api";
import type { Category, ExpenseCategory } from "@/types";

import { CategoriesControls } from "./categories-controls";
import { CategoriesDeleteDialogs } from "./categories-delete-dialogs";
import { CategoriesFormDialog } from "./categories-form-dialog";
import {
  createEmptyCategoryFormData,
  filterCategories,
  hasCategoryFormChanges,
  isSpecialExpenseCategory,
  mapCategoryToFormData,
  sortCategories,
  type CategoryFormData,
  type CategoryType,
  type CategoryViewMode,
  type ManagedCategory,
} from "./categories-helpers";
import { CategoriesList } from "./categories-list";

interface CategoryTarget {
  id: string;
  name: string;
  type: CategoryType;
  updated_at: string;
}

export function CategoriesScreen() {
  const { hasPermission } = useAuth();
  const canViewCategories = hasPermission(PERMISSIONS.CATEGORIES_VIEW);
  const canCreateCategories = hasPermission(PERMISSIONS.CATEGORIES_CREATE);
  const canEditCategories = hasPermission(PERMISSIONS.CATEGORIES_EDIT);
  const canDeleteCategories = hasPermission(PERMISSIONS.CATEGORIES_DELETE);

  const [activeTab, setActiveTab] = useState<CategoryType>("product");
  const [viewMode, setViewMode] = useState<CategoryViewMode>("active");
  const [productCategories, setProductCategories] = useState<Category[]>([]);
  const [expenseCategories, setExpenseCategories] = useState<ExpenseCategory[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState("");
  const [sort, setSort] = useState("name_asc");
  const [pageError, setPageError] = useState<string | null>(null);

  const [dialogOpen, setDialogOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [permanentDeleteDialogOpen, setPermanentDeleteDialogOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [editingCategory, setEditingCategory] = useState<CategoryTarget | null>(null);
  const [deletingCategory, setDeletingCategory] = useState<CategoryTarget | null>(null);
  const [formData, setFormData] = useState<CategoryFormData>(createEmptyCategoryFormData());
  const { showError } = useGlobalError();
  const resetCategoryLists = useCallback(() => {
    if (activeTab === "product") {
      setProductCategories([]);
      return;
    }
    setExpenseCategories([]);
  }, [activeTab]);

  const currentCategories = useMemo<ManagedCategory[]>(
    () => (activeTab === "product" ? productCategories : expenseCategories),
    [activeTab, expenseCategories, productCategories],
  );

  const fetchData = useCallback(async () => {
    setIsLoading(true);
    setPageError(null);

    const status = viewMode === "active" ? "active" : "archived";
    if (activeTab === "product") {
      const result = await api.getCategories(status);
      if (result.error) {
        setPageError(result.error);
      } else {
        setProductCategories(result.data || []);
      }
    } else {
      const result = await api.getExpenseCategories(status);
      if (result.error) {
        setPageError(result.error);
      } else {
        setExpenseCategories(result.data || []);
      }
    }

    setIsLoading(false);
  }, [activeTab, viewMode]);

  useEffect(() => {
    resetCategoryLists();
    void fetchData();
  }, [fetchData, resetCategoryLists, viewMode]);

  const sortOptions = useMemo(
    () => [
      { value: "name_asc", label: "Name (A-Z)" },
      { value: "name_desc", label: "Name (Z-A)" },
      { value: "updated_at_desc", label: "Recently updated" },
      { value: "updated_at_asc", label: "Oldest updated" },
      ...(activeTab === "product"
        ? [
            { value: "products_desc", label: "Products (high-low)" },
            { value: "products_asc", label: "Products (low-high)" },
          ]
        : []),
    ],
    [activeTab],
  );

  const filteredCategories = useMemo(
    () => filterCategories(sortCategories(currentCategories, sort), searchQuery, viewMode),
    [currentCategories, searchQuery, sort, viewMode],
  );

  const resetForm = useCallback(() => {
    setFormData(createEmptyCategoryFormData());
    setEditingCategory(null);
  }, []);

  const setDeleteTarget = useCallback((category: ManagedCategory) => {
    setDeletingCategory({
      id: category.id,
      name: category.name,
      type: activeTab,
      updated_at: category.updated_at,
    });
  }, [activeTab]);

  const openCreateDialog = useCallback(() => {
    if (!canCreateCategories) {
      return;
    }
    resetForm();
    setDialogOpen(true);
  }, [canCreateCategories, resetForm]);

  const openEditDialog = useCallback(
    (category: ManagedCategory) => {
      if (!canEditCategories) {
        return;
      }
      setEditingCategory({
        id: category.id,
        name: category.name,
        type: activeTab,
        updated_at: category.updated_at,
      });
      setFormData(mapCategoryToFormData(category));
      setDialogOpen(true);
    },
    [activeTab, canEditCategories],
  );

  const hasChanges = useMemo(() => {
    const category = currentCategories.find((item) => item.id === editingCategory?.id) || null;
    return hasCategoryFormChanges(formData, category);
  }, [currentCategories, editingCategory?.id, formData]);

  const submitProductCategory = useCallback(async () => {
    if (editingCategory) {
      return api.updateCategory(editingCategory.id, {
        name: formData.name,
        description: formData.description,
        expected_updated_at: editingCategory.updated_at,
      });
    }

    return api.createCategory({
      name: formData.name,
      description: formData.description,
    });
  }, [editingCategory, formData.description, formData.name]);

  const submitExpenseCategory = useCallback(async () => {
    if (editingCategory) {
      return api.updateExpenseCategory(editingCategory.id, {
        name: formData.name,
        description: formData.description,
        expected_updated_at: editingCategory.updated_at,
      });
    }

    return api.createExpenseCategory(formData.name, formData.description);
  }, [editingCategory, formData.description, formData.name]);

  const handleSubmit = useCallback(async () => {
    if (!hasChanges) {
      showError("No Changes", "Make a change before saving.");
      return;
    }
    if (editingCategory && !canEditCategories) {
      showError("Permission Denied", "You do not have permission to edit categories");
      return;
    }
    if (!editingCategory && !canCreateCategories) {
      showError("Permission Denied", "You do not have permission to create categories");
      return;
    }
    if (!formData.name.trim()) {
      showError("Name Required", "Category name is required");
      return;
    }

    setIsSubmitting(true);

    try {
      const result = activeTab === "product" ? await submitProductCategory() : await submitExpenseCategory();
      if (result.error) {
        showError("Save Failed", result.error);
        return;
      }

      setDialogOpen(false);
      resetForm();
      await fetchData();
    } finally {
      setIsSubmitting(false);
    }
  }, [
    activeTab,
    canCreateCategories,
    canEditCategories,
    editingCategory,
    fetchData,
    formData.name,
    hasChanges,
    resetForm,
    showError,
    submitExpenseCategory,
    submitProductCategory,
  ]);

  const handleArchive = useCallback(async () => {
    if (!deletingCategory) {
      return;
    }
    if (!canDeleteCategories) {
      showError("Permission Denied", "You do not have permission to archive categories");
      return;
    }

    setIsSubmitting(true);

    try {
      const result =
        deletingCategory.type === "product"
          ? await api.deleteCategory(deletingCategory.id, deletingCategory.updated_at)
          : await api.deleteExpenseCategory(
            deletingCategory.id,
            deletingCategory.updated_at,
          );

      if (result.error) {
        showError("Archive Failed", result.error);
        return;
      }

      setDeleteDialogOpen(false);
      setDeletingCategory(null);
      await fetchData();
    } finally {
      setIsSubmitting(false);
    }
  }, [canDeleteCategories, deletingCategory, fetchData, showError]);

  const handleRestore = useCallback(
    async (category: ManagedCategory) => {
      if (!canDeleteCategories) {
        showError("Permission Denied", "You do not have permission to restore categories");
        return;
      }

      setIsLoading(true);
      setPageError(null);
      try {
        const result =
          activeTab === "product"
            ? await api.updateCategory(category.id, {
              name: category.name,
              is_active: true,
              expected_updated_at: category.updated_at,
            })
            : await api.updateExpenseCategory(category.id, {
              is_active: true,
              expected_updated_at: category.updated_at,
            });

        if (result.error) {
          showError("Restore Failed", result.error);
          return;
        }

        await fetchData();
      } finally {
        setIsLoading(false);
      }
    },
    [activeTab, canDeleteCategories, fetchData, showError],
  );

  const handlePermanentDelete = useCallback(async () => {
    if (!deletingCategory) {
      return;
    }
    if (!canDeleteCategories) {
      showError("Permission Denied", "You do not have permission to delete categories");
      return;
    }

    setIsSubmitting(true);

    try {
      const result =
        deletingCategory.type === "product"
          ? await api.permanentDeleteCategory(
            deletingCategory.id,
            deletingCategory.updated_at,
          )
          : await api.permanentDeleteExpenseCategory(
            deletingCategory.id,
            deletingCategory.updated_at,
          );

      if (result.error) {
        showError("Delete Failed", result.error);
        return;
      }

      setPermanentDeleteDialogOpen(false);
      setDeletingCategory(null);
      await fetchData();
    } finally {
      setIsSubmitting(false);
    }
  }, [canDeleteCategories, deletingCategory, fetchData, showError]);

  if (!canViewCategories) {
    return (
      <div className="flex flex-col h-full">
        <Header title="Categories" />
        <div className="flex-1 flex items-center justify-center p-6">
          <Card className="max-w-md w-full">
            <CardHeader className="text-center">
              <div className="mx-auto w-12 h-12 bg-red-100 rounded-full flex items-center justify-center mb-4 text-red-600">
                <ShieldAlert className="h-6 w-6" />
              </div>
              <CardTitle>Access Denied</CardTitle>
              <CardDescription>
                You do not have permission to view categories. Please contact your administrator.
              </CardDescription>
            </CardHeader>
          </Card>
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-full">
      <Header title="Categories" />

      <div className="flex-1 p-6 overflow-auto">
        <div className="space-y-6">
          <CategoriesControls
            activeTab={activeTab}
            searchQuery={searchQuery}
            canCreateCategories={canCreateCategories}
            viewMode={viewMode}
            sort={sort}
            sortOptions={sortOptions}
            onActiveTabChange={(value) => {
              setActiveTab(value);
              setSort("name_asc");
            }}
            onSearchChange={setSearchQuery}
            onViewModeChange={setViewMode}
            onSortChange={setSort}
            onCreate={openCreateDialog}
          />

          {pageError ? (
            <div className="rounded-md border border-destructive/20 bg-destructive/10 p-3 text-sm text-destructive">
              {pageError}
            </div>
          ) : null}

          <CategoriesList
            isLoading={isLoading}
            categories={filteredCategories}
            type={activeTab}
            viewMode={viewMode}
            canEditCategories={canEditCategories}
            canDeleteCategories={canDeleteCategories}
            onEdit={openEditDialog}
            onArchive={(category) => {
              setDeleteTarget(category);
              setDeleteDialogOpen(true);
            }}
            onRestore={handleRestore}
            onPermanentDelete={(category) => {
              if (isSpecialExpenseCategory(category, activeTab)) return;
              setDeleteTarget(category);
              setPermanentDeleteDialogOpen(true);
            }}
          />
        </div>
      </div>

      <CategoriesFormDialog
        open={dialogOpen}
        activeTab={activeTab}
        editing={editingCategory !== null}
        formData={formData}
        isSubmitting={isSubmitting}
        hasChanges={hasChanges}
        onOpenChange={setDialogOpen}
        onFormDataChange={setFormData}
        onSubmit={() => void handleSubmit()}
      />

      <CategoriesDeleteDialogs
        archiveOpen={deleteDialogOpen}
        permanentDeleteOpen={permanentDeleteDialogOpen}
        categoryName={deletingCategory?.name || ""}
        isSubmitting={isSubmitting}
        onArchiveOpenChange={setDeleteDialogOpen}
        onPermanentDeleteOpenChange={setPermanentDeleteDialogOpen}
        onArchive={() => void handleArchive()}
        onPermanentDelete={() => void handlePermanentDelete()}
      />
    </div>
  );
}

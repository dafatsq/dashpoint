'use client';

import { useCallback, useEffect, useMemo, useState } from "react";
import { ShieldAlert } from "lucide-react";

import { Header } from "@/components/layout/header";
import { Card, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useAuth, PERMISSIONS } from "@/contexts/auth-context";
import api from "@/lib/api";
import type { Category, ExpenseCategory } from "@/types";

import { CategoriesControls } from "./categories-controls";
import { CategoriesDeleteDialogs } from "./categories-delete-dialogs";
import { CategoriesFormDialog } from "./categories-form-dialog";
import {
  createEmptyCategoryFormData,
  filterCategories,
  hasCategoryFormChanges,
  mapCategoryToFormData,
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
  const [pageError, setPageError] = useState<string | null>(null);

  const [dialogOpen, setDialogOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [permanentDeleteDialogOpen, setPermanentDeleteDialogOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [editingCategory, setEditingCategory] = useState<CategoryTarget | null>(null);
  const [deletingCategory, setDeletingCategory] = useState<CategoryTarget | null>(null);
  const [formData, setFormData] = useState<CategoryFormData>(createEmptyCategoryFormData());
  const [dialogError, setDialogError] = useState<string | null>(null);

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
    if (activeTab === "product") {
      setProductCategories([]);
    } else {
      setExpenseCategories([]);
    }
    void fetchData();
  }, [activeTab, fetchData, viewMode]);

  const filteredCategories = useMemo(
    () => filterCategories(currentCategories, searchQuery, viewMode),
    [currentCategories, searchQuery, viewMode],
  );

  const resetForm = useCallback(() => {
    setFormData(createEmptyCategoryFormData());
    setEditingCategory(null);
    setDialogError(null);
  }, []);

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
      setEditingCategory({ id: category.id, name: category.name, type: activeTab });
      setFormData(mapCategoryToFormData(category));
      setDialogError(null);
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
      });
    }

    return api.createExpenseCategory(formData.name, formData.description);
  }, [editingCategory, formData.description, formData.name]);

  const handleSubmit = useCallback(async () => {
    if (editingCategory && !canEditCategories) {
      setDialogError("You do not have permission to edit categories");
      return;
    }
    if (!editingCategory && !canCreateCategories) {
      setDialogError("You do not have permission to create categories");
      return;
    }
    if (!formData.name.trim()) {
      setDialogError("Category name is required");
      return;
    }

    setIsSubmitting(true);
    setDialogError(null);

    try {
      const result = activeTab === "product" ? await submitProductCategory() : await submitExpenseCategory();
      if (result.error) {
        setDialogError(result.error);
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
    resetForm,
    submitExpenseCategory,
    submitProductCategory,
  ]);

  const handleArchive = useCallback(async () => {
    if (!deletingCategory || !canDeleteCategories) {
      return;
    }

    setIsSubmitting(true);
    setDialogError(null);

    try {
      const result =
        deletingCategory.type === "product"
          ? await api.deleteCategory(deletingCategory.id)
          : await api.deleteExpenseCategory(deletingCategory.id);

      if (result.error) {
        setDialogError(result.error);
        return;
      }

      setDeleteDialogOpen(false);
      setDeletingCategory(null);
      await fetchData();
    } finally {
      setIsSubmitting(false);
    }
  }, [canDeleteCategories, deletingCategory, fetchData]);

  const handleRestore = useCallback(
    async (category: ManagedCategory) => {
      if (!canDeleteCategories) {
        return;
      }

      setIsLoading(true);
      setPageError(null);
      try {
        const result =
          activeTab === "product"
            ? await api.updateCategory(category.id, { name: category.name, is_active: true })
            : await api.updateExpenseCategory(category.id, { is_active: true });

        if (result.error) {
          setPageError(result.error);
          return;
        }

        await fetchData();
      } finally {
        setIsLoading(false);
      }
    },
    [activeTab, canDeleteCategories, fetchData],
  );

  const handlePermanentDelete = useCallback(async () => {
    if (!deletingCategory || !canDeleteCategories) {
      return;
    }

    setIsSubmitting(true);
    setDialogError(null);

    try {
      const result =
        deletingCategory.type === "product"
          ? await api.permanentDeleteCategory(deletingCategory.id)
          : await api.permanentDeleteExpenseCategory(deletingCategory.id);

      if (result.error) {
        setDialogError(result.error);
        return;
      }

      setPermanentDeleteDialogOpen(false);
      setDeletingCategory(null);
      await fetchData();
    } finally {
      setIsSubmitting(false);
    }
  }, [canDeleteCategories, deletingCategory, fetchData]);

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
        <div className="max-w-5xl mx-auto space-y-6">
          <CategoriesControls
            activeTab={activeTab}
            searchQuery={searchQuery}
            canCreateCategories={canCreateCategories}
            onActiveTabChange={setActiveTab}
            onSearchChange={setSearchQuery}
            onCreate={openCreateDialog}
          />

          {pageError ? (
            <div className="rounded-md border border-destructive/20 bg-destructive/10 p-3 text-sm text-destructive">
              {pageError}
            </div>
          ) : null}

          <Card className="border-none shadow-none bg-transparent">
            <Tabs value={viewMode} onValueChange={(value) => setViewMode(value as CategoryViewMode)}>
              <div className="flex items-center justify-between border-b mb-6">
                <TabsList className="bg-transparent border-none p-0 h-auto">
                  <TabsTrigger
                    value="active"
                    className="rounded-none border-b-2 border-transparent data-[state=active]:border-primary data-[state=active]:bg-transparent shadow-none h-10 px-4"
                  >
                    Active
                  </TabsTrigger>
                  <TabsTrigger
                    value="archived"
                    className="rounded-none border-b-2 border-transparent data-[state=active]:border-primary data-[state=active]:bg-transparent shadow-none h-10 px-4"
                  >
                    Archived
                  </TabsTrigger>
                </TabsList>
              </div>

              <CategoriesList
                isLoading={isLoading}
                categories={filteredCategories}
                type={activeTab}
                viewMode={viewMode}
                canEditCategories={canEditCategories}
                canDeleteCategories={canDeleteCategories}
                onEdit={openEditDialog}
                onArchive={(category) => {
                  setDeletingCategory({ id: category.id, name: category.name, type: activeTab });
                  setDialogError(null);
                  setDeleteDialogOpen(true);
                }}
                onRestore={handleRestore}
                onPermanentDelete={(category) => {
                  setDeletingCategory({ id: category.id, name: category.name, type: activeTab });
                  setDialogError(null);
                  setPermanentDeleteDialogOpen(true);
                }}
              />
            </Tabs>
          </Card>
        </div>
      </div>

      <CategoriesFormDialog
        open={dialogOpen}
        activeTab={activeTab}
        editing={editingCategory !== null}
        formData={formData}
        error={dialogError}
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
        error={dialogError}
        isSubmitting={isSubmitting}
        onArchiveOpenChange={setDeleteDialogOpen}
        onPermanentDeleteOpenChange={setPermanentDeleteDialogOpen}
        onArchive={() => void handleArchive()}
        onPermanentDelete={() => void handlePermanentDelete()}
      />
    </div>
  );
}

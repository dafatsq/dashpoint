import type { Category, ExpenseCategory } from "@/types";

export type CategoryType = "product" | "expense";
export type CategoryViewMode = "active" | "archived";
export type ManagedCategory = Category | ExpenseCategory;
type CategoryActionLabels = {
  emptyTitle: string;
  emptyDescription: string;
  deleteActionLabel: string;
};

/** The one expense category that is system-managed and must never be permanently deleted. */
export const SPECIAL_EXPENSE_CATEGORY_KEY = "inventory_purchase";

/**
 * Returns true if the given category is the protected inventory-purchase system category.
 * Only applicable to expense categories.
 */
export function isSpecialExpenseCategory(category: ManagedCategory, type: CategoryType): boolean {
  return type === "expense" && "system_key" in category && category.system_key === SPECIAL_EXPENSE_CATEGORY_KEY;
}

export interface CategoryFormData {
  name: string;
  description: string;
}

export function createEmptyCategoryFormData(): CategoryFormData {
  return {
    name: "",
    description: "",
  };
}

export function mapCategoryToFormData(category: ManagedCategory): CategoryFormData {
  return {
    name: category.name,
    description: category.description || "",
  };
}

export function hasCategoryFormChanges(formData: CategoryFormData, category: ManagedCategory | null): boolean {
  if (!category) {
    return true;
  }

  return (
    formData.name !== category.name ||
    formData.description !== (category.description || "")
  );
}

export function filterCategories<T extends ManagedCategory>(
  categories: T[],
  searchQuery: string,
  viewMode: CategoryViewMode,
): T[] {
  const normalizedSearch = searchQuery.trim().toLowerCase();

  return categories.filter((category) => {
    const matchesSearch = normalizedSearch === "" || category.name.toLowerCase().includes(normalizedSearch);
    const matchesStatus = viewMode === "active" ? category.is_active : !category.is_active;
    return matchesSearch && matchesStatus;
  });
}

export function resolveCategoryActionLabels(viewMode: CategoryViewMode) {
  const labelsByViewMode: Record<CategoryViewMode, CategoryActionLabels> = {
    active: {
      emptyTitle: "No active categories found",
      emptyDescription:
        "Try adjusting your search or add a new category to get started.",
      deleteActionLabel: "Archive Category",
    },
    archived: {
      emptyTitle: "No archived categories found",
      emptyDescription: "Archived categories will appear here.",
      deleteActionLabel: "Delete Permanently",
    },
  };

  return labelsByViewMode[viewMode];
}

'use client';

import { useCallback, useEffect, useMemo, useState } from "react";
import { useRouter } from "next/navigation";

import { DataTableContainer } from "@/components/shared/data-table-container";
import { ConfirmationDialog } from "@/components/shared/confirmation-dialog";
import { Header } from "@/components/layout/header";
import { useAuth, PERMISSIONS } from "@/contexts/auth-context";
import { useGlobalError } from "@/contexts/error-context";
import api from "@/lib/api";
import { CreateExpenseRequest, Expense, ExpenseCategory, ExpenseSummary, Product } from "@/types";

import {
  buildExpenseRequest,
  canSubmitExpenseForm,
  createEmptyExpenseFormData,
  deriveInventoryPurchaseAmount,
  deriveInventoryPurchaseDescription,
  hasExpenseFormChanges,
  isInventoryPurchaseCategory,
  mapExpenseToFormData,
  todayDateString,
} from "./expenses-helpers";
import { ExpensesFormDialog } from "./expenses-form-dialog";
import { ExpensesList } from "./expenses-list";
import { ExpensesSummaryCards } from "./expenses-summary";
import { ExpensesToolbar } from "./expenses-toolbar";

export function ExpensesScreen() {
  const { hasPermission, isLoading: isAuthLoading } = useAuth();
  const router = useRouter();
  const canCreateExpense = hasPermission(PERMISSIONS.EXPENSES_CREATE);
  const canEditExpense = hasPermission(PERMISSIONS.EXPENSES_EDIT);
  const canDeleteExpense = hasPermission(PERMISSIONS.EXPENSES_DELETE);
  const canManageAnyExpense = canCreateExpense || canEditExpense || canDeleteExpense;

  const [expenses, setExpenses] = useState<Expense[]>([]);
  const [categories, setCategories] = useState<ExpenseCategory[]>([]);
  const [products, setProducts] = useState<Product[]>([]);
  const [summary, setSummary] = useState<ExpenseSummary | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [pageError, setPageError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedCategory, setSelectedCategory] = useState<string>("all");
  const [dateRange, setDateRange] = useState({ start: "", end: "" });
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(20);
  const [hasMore, setHasMore] = useState(true);
  const [total, setTotal] = useState(0);

  const [dialogOpen, setDialogOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [editingExpense, setEditingExpense] = useState<Expense | null>(null);
  const [deletingExpense, setDeletingExpense] = useState<Expense | null>(null);
  const { showError } = useGlobalError();
  const [formData, setFormData] = useState<CreateExpenseRequest>(createEmptyExpenseFormData());
  const [formErrors, setFormErrors] = useState<{ amount?: string; description?: string; general?: string }>({});
  const [isManualAmount, setIsManualAmount] = useState(false);
  const [isManualDescription, setIsManualDescription] = useState(false);

  useEffect(() => {
    if (!isAuthLoading && !hasPermission(PERMISSIONS.EXPENSES_VIEW)) {
      router.push("/");
    }
  }, [hasPermission, isAuthLoading, router]);

  const refreshSummary = useCallback(async () => {
    const summaryResult = await api.getExpenseSummary({
      start_date: dateRange.start,
      end_date: dateRange.end,
    });
    if (summaryResult.error) {
      setPageError(summaryResult.error);
      return;
    }
    if (summaryResult.data) {
      setSummary(summaryResult.data);
    }
  }, [dateRange.end, dateRange.start]);

  const loadExpensesData = useCallback(async () => {
    setIsLoading(true);
    const [expensesResult, categoriesResult, productsResult, summaryResult] = await Promise.all([
      api.getExpenses({
        category_id: selectedCategory !== "all" ? selectedCategory : undefined,
        start_date: dateRange.start,
        end_date: dateRange.end,
        limit,
        offset: (page - 1) * limit,
      }),
      api.getExpenseCategories(),
      api.getProducts({ active: true }),
      api.getExpenseSummary({
        start_date: dateRange.start,
        end_date: dateRange.end,
      }),
    ]);

    const firstError =
      expensesResult.error || categoriesResult.error || productsResult.error || summaryResult.error || null;
    setPageError(firstError);

    if (expensesResult.data) {
      setExpenses(expensesResult.data.expenses);
      setTotal(expensesResult.data.total || 0);
      setHasMore(expensesResult.data.expenses.length === limit);
    }
    if (categoriesResult.data) {
      setCategories(categoriesResult.data);
    }
    if (productsResult.data) {
      setProducts(productsResult.data);
    }
    if (summaryResult.data) {
      setSummary(summaryResult.data);
    }
    setIsLoading(false);
  }, [dateRange.end, dateRange.start, limit, page, selectedCategory]);

  const resetToFirstPage = useCallback(() => {
    setPage(1);
  }, []);

  useEffect(() => {
    const timeoutId = window.setTimeout(() => {
      void loadExpensesData();
    }, 0);

    return () => window.clearTimeout(timeoutId);
  }, [loadExpensesData]);

  const filteredExpenses = useMemo(
    () =>
      expenses.filter((expense) => {
        const search = searchQuery.toLowerCase();
        return (
          expense.description.toLowerCase().includes(search) ||
          expense.vendor?.toLowerCase().includes(search) ||
          expense.category_name?.toLowerCase().includes(search)
        );
      }),
    [expenses, searchQuery],
  );

  const resetForm = useCallback(() => {
    setFormData(createEmptyExpenseFormData(todayDateString()));
    setEditingExpense(null);
    setFormErrors({});
    setIsManualAmount(false);
    setIsManualDescription(false);
  }, []);

  const inventoryPurchase = useMemo(
    () => isInventoryPurchaseCategory(formData.category_id, categories),
    [categories, formData.category_id],
  );

  useEffect(() => {
    if (inventoryPurchase && formData.product_id && formData.quantity && !isManualAmount) {
      const amount = deriveInventoryPurchaseAmount(formData.product_id, formData.quantity, products);
      if (amount) {
        setFormData((prev) => ({ ...prev, amount }));
      }
    }
  }, [formData.product_id, formData.quantity, inventoryPurchase, isManualAmount, products]);

  useEffect(() => {
    if (inventoryPurchase && formData.product_id && formData.quantity && !isManualDescription) {
      const description = deriveInventoryPurchaseDescription(formData.product_id, formData.quantity, products);
      if (description) {
        setFormData((prev) => ({ ...prev, description }));
      }
    }
  }, [formData.product_id, formData.quantity, inventoryPurchase, isManualDescription, products]);

  const openCreateDialog = useCallback(() => {
    if (!canCreateExpense) {
      return;
    }
    resetForm();
    setDialogOpen(true);
  }, [canCreateExpense, resetForm]);

  const openEditDialog = useCallback(
    (expense: Expense) => {
      if (!canEditExpense) {
        return;
      }
      setEditingExpense(expense);
      setFormData(mapExpenseToFormData(expense));
      setIsManualAmount(!!expense.product_id);
      setIsManualDescription(true);
      setFormErrors({});
      setDialogOpen(true);
    },
    [canEditExpense],
  );

  const hasChanges = useMemo(() => hasExpenseFormChanges(formData, editingExpense), [editingExpense, formData]);
  const canSubmitForm = useMemo(
    () =>
      canSubmitExpenseForm({
        isSubmitting,
        hasChanges,
        categoryId: formData.category_id,
        isInventoryPurchase: inventoryPurchase,
        productId: formData.product_id,
        quantity: formData.quantity,
      }),
    [
      formData.category_id,
      formData.product_id,
      formData.quantity,
      hasChanges,
      inventoryPurchase,
      isSubmitting,
    ],
  );

  const handleSubmit = async () => {
    if (editingExpense && !canEditExpense) {
      showError("Permission Denied", "You do not have permission to edit expenses");
      return;
    }
    if (!editingExpense && !canCreateExpense) {
      showError("Permission Denied", "You do not have permission to create expenses");
      return;
    }
    if (!hasChanges) {
      showError("No Changes", "Make a change before saving.");
      return;
    }
    if (!canSubmitForm) {
      if (inventoryPurchase && !formData.product_id) {
        showError("Product Required", "Select a product before saving this inventory purchase.");
      } else if (inventoryPurchase && !formData.quantity) {
        showError("Quantity Required", "Enter a quantity before saving this inventory purchase.");
      } else if (!formData.category_id) {
        showError("Category Required", "Select a category before saving.");
      }
      return;
    }

    if (!formData.description.trim()) {
      showError("Description Required", "Please provide a description for this expense.");
      return;
    }
    if (!formData.amount || parseFloat(formData.amount) <= 0) {
      showError("Amount Required", "Please provide a valid amount greater than 0.");
      return;
    }

    setFormErrors({});
    setIsSubmitting(true);

    try {
      const expenseData = buildExpenseRequest(formData);
      if (editingExpense) {
        const result = await api.updateExpense(editingExpense.id, {
          ...expenseData,
          expected_updated_at: editingExpense.updated_at,
        });
        if (result.error) {
          showError("Save Failed", result.error);
          return;
        }
        if (result.data) {
          setExpenses((prev) => prev.map((expense) => (expense.id === editingExpense.id ? result.data! : expense)));
        }
      } else {
        const result = await api.createExpense(expenseData);
        if (result.error) {
          showError("Create Failed", result.error);
          return;
        }
        if (result.data) {
          setExpenses((prev) => [result.data!, ...prev]);
        }
      }

      await refreshSummary();
      setDialogOpen(false);
      resetForm();
    } catch {
      showError("Save Error", "Failed to save expense. Please try again.");
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleDelete = async () => {
    if (!deletingExpense) {
      return;
    }
    if (!canDeleteExpense) {
      showError("Permission Denied", "You do not have permission to delete expenses");
      return;
    }

    setIsSubmitting(true);
    const result = await api.deleteExpense(
      deletingExpense.id,
      deletingExpense.updated_at,
    );
    if (result.error) {
      showError("Delete Failed", result.error);
      setIsSubmitting(false);
      return;
    }

    setExpenses((prev) => prev.filter((expense) => expense.id !== deletingExpense.id));
    await refreshSummary();
    setDeleteDialogOpen(false);
    setDeletingExpense(null);
    setIsSubmitting(false);
  };

  const formatCurrency = (amount: string | number) =>
    new Intl.NumberFormat("id-ID", {
      style: "currency",
      currency: "IDR",
      minimumFractionDigits: 0,
    }).format(typeof amount === "string" ? parseFloat(amount) : amount);

  const formatDate = (dateString: string) =>
    new Date(dateString).toLocaleDateString("id-ID", {
      year: "numeric",
      month: "short",
      day: "numeric",
    });

  return (
    <div className="flex h-full flex-col">
      <Header title="Expenses" />

      <div className="flex-1 overflow-auto p-6">
        {pageError && (
          <div className="mb-6 rounded-md border border-destructive/20 bg-destructive/10 p-3">
            <p className="text-sm text-destructive">{pageError}</p>
          </div>
        )}

        <ExpensesSummaryCards
          summary={summary}
          dateRange={dateRange}
          formatCurrency={formatCurrency}
          formatDate={formatDate}
        />

        <ExpensesToolbar
          categories={categories}
          searchQuery={searchQuery}
          selectedCategory={selectedCategory}
          dateRange={dateRange}
          canCreateExpense={canCreateExpense}
          onCreate={openCreateDialog}
          onSearchChange={(value) => {
            setSearchQuery(value);
            resetToFirstPage();
          }}
          onCategoryChange={(value) => {
            setSelectedCategory(value);
            resetToFirstPage();
          }}
          onDateRangeChange={(range) => {
            setDateRange(range);
            resetToFirstPage();
          }}
        />

        <DataTableContainer
          limit={limit}
          onLimitChange={(value) => {
            setLimit(value);
            resetToFirstPage();
          }}
          total={total}
          currentCount={expenses.length}
        >
          <ExpensesList
            expenses={filteredExpenses}
            isLoading={isLoading}
            canCreateExpense={canCreateExpense}
            canEditExpense={canEditExpense}
            canDeleteExpense={canDeleteExpense}
            canManageAnyExpense={canManageAnyExpense}
            page={page}
            hasMore={hasMore}
            onCreate={openCreateDialog}
            onEdit={openEditDialog}
            onDelete={(expense) => {
              setDeletingExpense(expense);
              setDeleteDialogOpen(true);
            }}
            onPreviousPage={() => setPage((prev) => Math.max(1, prev - 1))}
            onNextPage={() => setPage((prev) => prev + 1)}
            formatCurrency={formatCurrency}
            formatDate={formatDate}
          />
        </DataTableContainer>
      </div>

      <ExpensesFormDialog
        open={dialogOpen}
        editingExpense={editingExpense}
        categories={categories}
        products={products}
        formData={formData}
        formErrors={formErrors}
        isSubmitting={isSubmitting}
        hasChanges={hasChanges}
        isInventoryPurchase={inventoryPurchase}
        isManualAmount={isManualAmount}
        isManualDescription={isManualDescription}
        onOpenChange={setDialogOpen}
        onFormDataChange={setFormData}
        onFormErrorsChange={setFormErrors}
        onManualAmountChange={setIsManualAmount}
        onManualDescriptionChange={setIsManualDescription}
        onSubmit={handleSubmit}
        formatCurrency={formatCurrency}
      />

      <ConfirmationDialog
        open={deleteDialogOpen}
        onOpenChange={setDeleteDialogOpen}
        title="Delete Expense"
        description="Are you sure you want to delete this expense? This action cannot be undone."
        confirmText="Delete"
        isSubmitting={isSubmitting}
        loadingText="Deleting..."
        onConfirm={handleDelete}
      />
    </div>
  );
}

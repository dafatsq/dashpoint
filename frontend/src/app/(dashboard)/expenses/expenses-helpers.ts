import { CreateExpenseRequest, Expense, ExpenseCategory, Product } from "@/types";

export function todayDateString(): string {
  return new Date().toISOString().split("T")[0];
}

export function createEmptyExpenseFormData(expenseDate = todayDateString()): CreateExpenseRequest {
  return {
    category_id: "",
    product_id: "",
    quantity: "",
    amount: "",
    description: "",
    expense_date: expenseDate,
    vendor: "",
    reference_number: "",
    notes: "",
  };
}

export function isInventoryPurchaseCategory(
  categoryID: string | undefined,
  categories: ExpenseCategory[],
): boolean {
  if (!categoryID) {
    return false;
  }
  const category = categories.find((item) => item.id === categoryID);
  return category?.name === "Inventory Purchase";
}

export function deriveInventoryPurchaseAmount(
  productID: string | undefined,
  quantity: string | undefined,
  products: Product[],
): string {
  if (!productID || !quantity) {
    return "";
  }
  const product = products.find((item) => item.id === productID);
  if (!product?.cost) {
    return "";
  }
  const qty = parseFloat(quantity);
  const cost = parseFloat(product.cost);
  if (!Number.isFinite(qty) || !Number.isFinite(cost)) {
    return "";
  }
  return (qty * cost).toFixed(2);
}

export function deriveInventoryPurchaseDescription(
  productID: string | undefined,
  quantity: string | undefined,
  products: Product[],
): string {
  if (!productID || !quantity) {
    return "";
  }
  const product = products.find((item) => item.id === productID);
  return product ? `${product.name} x ${quantity}` : "";
}

export function mapExpenseToFormData(expense: Expense): CreateExpenseRequest {
  return {
    category_id: expense.category_id || "",
    product_id: expense.product_id || "",
    quantity: expense.quantity || "",
    amount: expense.amount,
    description: expense.description,
    expense_date: expense.expense_date,
    vendor: expense.vendor || "",
    reference_number: expense.reference_number || "",
    notes: expense.notes || "",
  };
}

export function hasExpenseFormChanges(formData: CreateExpenseRequest, editingExpense: Expense | null): boolean {
  if (!editingExpense) {
    return true;
  }
  return (
    formData.category_id !== (editingExpense.category_id || "") ||
    formData.product_id !== (editingExpense.product_id || "") ||
    formData.quantity !== (editingExpense.quantity || "") ||
    formData.amount !== editingExpense.amount ||
    formData.description !== editingExpense.description ||
    formData.expense_date !== editingExpense.expense_date ||
    formData.vendor !== (editingExpense.vendor || "") ||
    formData.reference_number !== (editingExpense.reference_number || "") ||
    formData.notes !== (editingExpense.notes || "")
  );
}

export function canSubmitExpenseForm(input: {
  isSubmitting: boolean;
  hasChanges: boolean;
  categoryId: string | undefined;
  isInventoryPurchase: boolean;
  productId: string | undefined;
  quantity: string | undefined;
}): boolean {
  if (input.isSubmitting || !input.hasChanges) {
    return false;
  }

  if (!input.categoryId) {
    return false;
  }

  if (input.isInventoryPurchase && !input.productId) {
    return false;
  }

  if (input.isInventoryPurchase && !input.quantity) {
    return false;
  }

  return true;
}

export function buildExpenseRequest(formData: CreateExpenseRequest): CreateExpenseRequest {
  return {
    ...formData,
    category_id: formData.category_id && formData.category_id !== "none" ? formData.category_id : undefined,
    product_id: formData.product_id || undefined,
    quantity: formData.quantity || undefined,
    vendor: formData.vendor || undefined,
    reference_number: formData.reference_number || undefined,
    notes: formData.notes || undefined,
  };
}

import { Loader2 } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { DatePicker } from "@/components/ui/date-picker";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { CreateExpenseRequest, Expense, ExpenseCategory, Product } from "@/types";

interface ExpensesFormDialogProps {
  open: boolean;
  editingExpense: Expense | null;
  categories: ExpenseCategory[];
  products: Product[];
  formData: CreateExpenseRequest;
  formErrors: { amount?: string; description?: string; general?: string };
  isSubmitting: boolean;
  isInventoryPurchase: boolean;
  isManualAmount: boolean;
  isManualDescription: boolean;
  onOpenChange: (open: boolean) => void;
  onFormDataChange: (value: CreateExpenseRequest) => void;
  onFormErrorsChange: (value: { amount?: string; description?: string; general?: string }) => void;
  onManualAmountChange: (value: boolean) => void;
  onManualDescriptionChange: (value: boolean) => void;
  onSubmit: () => void;
  formatCurrency: (amount: string | number) => string;
}

export function ExpensesFormDialog({
  open,
  editingExpense,
  categories,
  products,
  formData,
  formErrors,
  isSubmitting,
  isInventoryPurchase,
  isManualAmount,
  isManualDescription,
  onOpenChange,
  onFormDataChange,
  onFormErrorsChange,
  onManualAmountChange,
  onManualDescriptionChange,
  onSubmit,
  formatCurrency,
}: ExpensesFormDialogProps) {
  const updateFormData = (patch: Partial<CreateExpenseRequest>) => {
    onFormDataChange({ ...formData, ...patch });
  };
  const hasCategory = !!formData.category_id && formData.category_id !== "none";
  const inventoryFieldsReady = !!formData.product_id && !!formData.quantity;
  const canEditRemainingFields = hasCategory && (!isInventoryPurchase || inventoryFieldsReady);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[85vh] overflow-y-auto sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{editingExpense ? "Edit Expense" : "Add Expense"}</DialogTitle>
          <DialogDescription>
            {editingExpense ? "Update the expense details below." : "Fill in the details for the new expense."}
          </DialogDescription>
        </DialogHeader>

        <div className="grid gap-4 py-4">
          {formErrors.general && (
            <div className="rounded-md border border-destructive/20 bg-destructive/10 p-3">
              <p className="text-sm text-destructive">{formErrors.general}</p>
            </div>
          )}

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div className="grid gap-2">
              <Label htmlFor="category">Category</Label>
              <Select
                value={formData.category_id || "none"}
                onValueChange={(value) => {
                  const nextCategoryId = value === "none" ? "" : value;
                  if (!editingExpense) {
                    onFormDataChange({
                      category_id: nextCategoryId,
                      product_id: "",
                      quantity: "",
                      applies_inventory: false,
                      amount: "",
                      description: "",
                      expense_date: formData.expense_date,
                      vendor: "",
                      reference_number: "",
                      notes: "",
                    });
                  } else {
                    updateFormData({ category_id: nextCategoryId, product_id: "", quantity: "", applies_inventory: false });
                  }
                  onManualAmountChange(false);
                  onManualDescriptionChange(false);
                }}
              >
                <SelectTrigger>
                  <SelectValue placeholder="Select category" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="none">No Category</SelectItem>
                  {categories.map((category) => (
                    <SelectItem key={category.id} value={category.id}>
                      {category.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="grid gap-2">
              {isInventoryPurchase ? (
                <>
                  <Label htmlFor="quantity" className={!hasCategory ? "opacity-50" : ""}>
                    Quantity *
                  </Label>
                  <Input
                    id="quantity"
                    type="number"
                    min="1"
                    value={formData.quantity}
                    onChange={(event) => {
                      updateFormData({ quantity: event.target.value });
                      onManualAmountChange(false);
                      onManualDescriptionChange(false);
                    }}
                    placeholder="Enter quantity"
                    disabled={!hasCategory}
                  />
                  <div className="mt-2 flex items-center gap-2">
                    <Checkbox
                      id="applies_inventory"
                      checked={!!formData.applies_inventory}
                      onCheckedChange={(checked) => updateFormData({ applies_inventory: checked === true })}
                      disabled={!canEditRemainingFields}
                    />
                    <Label htmlFor="applies_inventory" className={!canEditRemainingFields ? "opacity-50" : ""}>
                      Add to product inventory
                    </Label>
                  </div>
                </>
              ) : (
                <>
                  <Label htmlFor="vendor" className={!hasCategory ? "opacity-50" : ""}>
                    Vendor
                  </Label>
                  <Input
                    id="vendor"
                    value={formData.vendor}
                    onChange={(event) => onFormDataChange({ ...formData, vendor: event.target.value })}
                    placeholder="e.g., PLN"
                    disabled={!canEditRemainingFields}
                  />
                </>
              )}
            </div>
          </div>

          <div className="grid gap-2">
            {isInventoryPurchase ? (
              <>
                <Label htmlFor="product" className={!hasCategory ? "opacity-50" : ""}>
                  Product *
                </Label>
                <Select
                  value={formData.product_id || "none"}
                  onValueChange={(value) => {
                    updateFormData({ product_id: value === "none" ? "" : value });
                    onManualAmountChange(false);
                    onManualDescriptionChange(false);
                  }}
                  disabled={!hasCategory}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select product" />
                  </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="none">Select a product</SelectItem>
                      {products.map((product) => (
                        <SelectItem key={product.id} value={product.id}>
                        {product.name} - {formatCurrency(parseFloat(product.cost || "0"))}
                        </SelectItem>
                      ))}
                  </SelectContent>
                </Select>
              </>
            ) : (
              <>
                <Label htmlFor="description" className={!hasCategory ? "opacity-50" : ""}>
                  Description *
                </Label>
                <Input
                  id="description"
                  value={formData.description}
                  onChange={(event) => {
                    updateFormData({ description: event.target.value });
                    if (formErrors.description) {
                      onFormErrorsChange({ ...formErrors, description: undefined });
                    }
                  }}
                  placeholder="e.g., Monthly electricity bill"
                  disabled={!canEditRemainingFields}
                />
              </>
            )}
          </div>

          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div className="grid gap-2">
              <div className="flex items-center justify-between">
                <Label htmlFor="amount" className={!hasCategory ? "opacity-50" : ""}>
                  Amount (IDR) *
                </Label>
                {isInventoryPurchase && formData.product_id && formData.quantity && (
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    className="h-8 text-xs"
                    onClick={() => onManualAmountChange(!isManualAmount)}
                    disabled={!canEditRemainingFields}
                  >
                    {isManualAmount ? "Auto-calculate" : "Manual edit"}
                  </Button>
                )}
              </div>
              <Input
                id="amount"
                type="number"
                value={formData.amount}
                onChange={(event) => {
                  updateFormData({ amount: event.target.value });
                  if (formErrors.amount) {
                    onFormErrorsChange({ ...formErrors, amount: undefined });
                  }
                  if (isInventoryPurchase) {
                    onManualAmountChange(true);
                  }
                }}
                placeholder="100000"
                disabled={
                  !canEditRemainingFields ||
                  (isInventoryPurchase && !isManualAmount && inventoryFieldsReady)
                }
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="expense_date" className={!canEditRemainingFields ? "opacity-50" : ""}>
                Date *
              </Label>
              <DatePicker
                date={formData.expense_date}
                onSelect={(date) => updateFormData({ expense_date: date })}
                disabled={!canEditRemainingFields}
              />
            </div>
          </div>

          {isInventoryPurchase && (
            <>
              <div className="grid gap-2">
                <div className="flex items-center justify-between">
                  <Label htmlFor="description_inventory" className={!canEditRemainingFields ? "opacity-50" : ""}>
                    Description *
                  </Label>
                  {formData.product_id && formData.quantity && (
                  <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      className="h-8 text-xs"
                      onClick={() => onManualDescriptionChange(!isManualDescription)}
                      disabled={!canEditRemainingFields}
                    >
                      {isManualDescription ? "Auto-generate" : "Manual edit"}
                    </Button>
                  )}
                </div>
                <Input
                  id="description_inventory"
                  value={formData.description}
                  onChange={(event) => {
                    updateFormData({ description: event.target.value });
                    if (formErrors.description) {
                      onFormErrorsChange({ ...formErrors, description: undefined });
                    }
                    onManualDescriptionChange(true);
                  }}
                  placeholder="e.g., Stock for February"
                  disabled={!canEditRemainingFields || (!isManualDescription && inventoryFieldsReady)}
                />
                {!isManualDescription && inventoryFieldsReady && (
                  <p className="text-xs text-muted-foreground">Auto-generated from product and quantity</p>
                )}
              </div>

              <div className="grid gap-2">
                <Label htmlFor="vendor_inventory" className={!canEditRemainingFields ? "opacity-50" : ""}>
                  Vendor
                </Label>
                <Input
                  id="vendor_inventory"
                  value={formData.vendor}
                  onChange={(event) => updateFormData({ vendor: event.target.value })}
                  placeholder="e.g., Supplier Name"
                  disabled={!canEditRemainingFields}
                />
              </div>
            </>
          )}

          <div className="grid gap-2">
            <Label htmlFor="reference_number" className={!canEditRemainingFields ? "opacity-50" : ""}>
              Reference Number
            </Label>
            <Input
              id="reference_number"
              value={formData.reference_number}
              onChange={(event) => updateFormData({ reference_number: event.target.value })}
              placeholder="e.g., Invoice #12345"
              disabled={!canEditRemainingFields}
            />
          </div>

          <div className="grid gap-2">
            <Label htmlFor="notes" className={!canEditRemainingFields ? "opacity-50" : ""}>
              Notes
            </Label>
            <Input
              id="notes"
              value={formData.notes}
              onChange={(event) => updateFormData({ notes: event.target.value })}
              placeholder="Additional notes..."
              disabled={!canEditRemainingFields}
            />
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={onSubmit} disabled={isSubmitting}>
            {isSubmitting ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Saving...
              </>
            ) : editingExpense ? (
              "Update Expense"
            ) : (
              "Create Expense"
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

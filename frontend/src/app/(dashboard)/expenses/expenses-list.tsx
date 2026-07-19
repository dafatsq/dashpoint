import { ChevronLeft, ChevronRight, Loader2, Pencil, Receipt, Trash2 } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Expense } from "@/types";

interface ExpensesListProps {
  expenses: Expense[];
  isLoading: boolean;
  canCreateExpense: boolean;
  canEditExpense: boolean;
  canDeleteExpense: boolean;
  canManageAnyExpense: boolean;
  page: number;
  hasMore: boolean;
  onCreate: () => void;
  onEdit: (expense: Expense) => void;
  onDelete: (expense: Expense) => void;
  onPreviousPage: () => void;
  onNextPage: () => void;
  formatCurrency: (amount: string | number) => string;
  formatDate: (value: string) => string;
}

export function ExpensesList({
  expenses,
  isLoading,
  canCreateExpense,
  canEditExpense,
  canDeleteExpense,
  canManageAnyExpense,
  page,
  hasMore,
  onCreate,
  onEdit,
  onDelete,
  onPreviousPage,
  onNextPage,
  formatCurrency,
  formatDate,
}: ExpensesListProps) {
  if (isLoading) {
    return (
      <div className="flex h-64 items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  if (expenses.length === 0) {
    return (
      <Card>
        <CardContent className="flex flex-col items-center justify-center py-12">
          <Receipt className="mb-4 h-12 w-12 text-muted-foreground" />
          <p className="text-muted-foreground">No expenses found</p>
          {canCreateExpense && (
            <Button variant="link" onClick={onCreate}>
              Add your first expense
            </Button>
          )}
        </CardContent>
      </Card>
    );
  }

  return (
    <>
      <div className="hidden lg:block overflow-x-auto">
        <table className="w-full min-w-max">
          <thead>
            <tr className="border-b text-left text-sm text-muted-foreground">
              <th className="pb-3 font-medium">Date</th>
              <th className="pb-3 font-medium">Description</th>
              <th className="pb-3 font-medium">Category</th>
              <th className="pb-3 font-medium">User</th>
              <th className="pb-3 font-medium">Vendor</th>
              <th className="pb-3 text-right font-medium">Amount</th>
              {canManageAnyExpense && <th className="pb-3 text-right font-medium">Actions</th>}
            </tr>
          </thead>
          <tbody>
            {expenses.map((expense) => (
              <tr key={expense.id} className="border-b last:border-0 hover:bg-muted/50">
                <td className="py-3 text-sm">{formatDate(expense.expense_date)}</td>
                <td className="py-3">
                  <div>
                    <p className="font-medium">{expense.description}</p>
                    {expense.reference_number && (
                      <p className="text-xs text-muted-foreground">Ref: {expense.reference_number}</p>
                    )}
                  </div>
                </td>
                <td className="py-3 text-sm">
                  <span className="inline-flex items-center rounded-full bg-blue-600 px-2 py-1 text-xs font-medium text-white dark:bg-blue-600/90 dark:text-white">
                    {expense.category_name || "Uncategorized"}
                  </span>
                </td>
                <td className="py-3 text-sm">{expense.created_by_name || "Unknown"}</td>
                <td className="py-3 text-sm text-muted-foreground">{expense.vendor || "-"}</td>
                <td className="py-3 text-right font-medium text-destructive font-mono">{formatCurrency(expense.amount)}</td>
                {canManageAnyExpense && (
                  <td className="py-3 text-right">
                    <div className="flex items-center justify-end gap-1">
                      {canEditExpense && (
                        <Button variant="ghost" size="icon" onClick={() => onEdit(expense)}>
                          <Pencil className="h-4 w-4" />
                        </Button>
                      )}
                      {canDeleteExpense && (
                        <Button variant="ghost" size="icon" onClick={() => onDelete(expense)}>
                          <Trash2 className="h-4 w-4 text-destructive" />
                        </Button>
                      )}
                    </div>
                  </td>
                )}
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <div className="space-y-4 lg:hidden">
        {expenses.map((expense) => (
          <Card key={expense.id} className="@container">
            <CardContent className="space-y-3 p-4">
              <div className="flex items-start justify-between border-b pb-2">
                <div>
                  <span className="text-sm font-bold text-destructive font-mono">{formatCurrency(expense.amount)}</span>
                  <span className="block text-xs text-muted-foreground">{formatDate(expense.expense_date)}</span>
                </div>
                <span className="inline-flex items-center rounded-full bg-blue-600 px-2 py-1 text-xs font-medium text-white dark:bg-blue-600/90 dark:text-white">
                  {expense.category_name || "Uncategorized"}
                </span>
              </div>
              <div>
                <p className="font-medium">{expense.description}</p>
                {expense.reference_number && (
                  <p className="text-xs text-muted-foreground">Ref: {expense.reference_number}</p>
                )}
              </div>
              <div className="grid grid-cols-2 gap-2 text-sm text-muted-foreground">
                <div className="flex flex-col">
                  <span className="text-xs">Vendor</span>
                  <span className="text-foreground">{expense.vendor || "-"}</span>
                </div>
                <div className="flex flex-col">
                  <span className="text-xs">User</span>
                  <span className="text-foreground">{expense.created_by_name || "Unknown"}</span>
                </div>
              </div>
              {canManageAnyExpense && (
                <div className="flex justify-end gap-2 border-t pt-2">
                  {canEditExpense && (
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => onEdit(expense)}
                      className="h-8 w-8 p-0 @[250px]:w-auto @[250px]:px-3"
                      title="Edit"
                    >
                      <Pencil className="h-3.5 w-3.5 @[250px]:mr-1" />
                      <span className="hidden @[250px]:inline">Edit</span>
                    </Button>
                  )}
                  {canDeleteExpense && (
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => onDelete(expense)}
                      className="h-8 w-8 p-0 @[250px]:w-auto @[250px]:px-3 text-destructive hover:bg-destructive/10 hover:text-destructive"
                      title="Delete"
                    >
                      <Trash2 className="h-3.5 w-3.5 @[250px]:mr-1" />
                      <span className="hidden @[250px]:inline">Delete</span>
                    </Button>
                  )}
                </div>
              )}
            </CardContent>
          </Card>
        ))}
      </div>

      <div className="mt-6 flex items-center justify-between border-t pt-4">
        <div className="flex gap-2">
          <Button variant="outline" size="sm" onClick={onPreviousPage} disabled={page === 1}>
            <ChevronLeft className="mr-1 h-4 w-4" />
            Previous
          </Button>
          <Button variant="outline" size="sm" onClick={onNextPage} disabled={!hasMore}>
            Next
            <ChevronRight className="ml-1 h-4 w-4" />
          </Button>
        </div>
      </div>
    </>
  );
}

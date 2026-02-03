'use client';

import { useEffect, useState, useCallback } from 'react';
import { Header } from '@/components/layout/header';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Search,
  Plus,
  Pencil,
  Trash2,
  Loader2,
  Receipt,
  TrendingDown,
  Calendar,
  Info,
} from 'lucide-react';
import api from '@/lib/api';
import { Expense, ExpenseCategory, CreateExpenseRequest, ExpenseSummary, Product } from '@/types';

export default function ExpensesPage() {
  const [expenses, setExpenses] = useState<Expense[]>([]);
  const [categories, setCategories] = useState<ExpenseCategory[]>([]);
  const [products, setProducts] = useState<Product[]>([]);
  const [summary, setSummary] = useState<ExpenseSummary | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedCategory, setSelectedCategory] = useState<string>('all');
  const [dateRange, setDateRange] = useState({
    start: '',
    end: '',
  });

  // Dialog states
  const [dialogOpen, setDialogOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [editingExpense, setEditingExpense] = useState<Expense | null>(null);
  const [deletingExpense, setDeletingExpense] = useState<Expense | null>(null);

  const [deleteError, setDeleteError] = useState<string | null>(null);

  // Form state
  const [formData, setFormData] = useState<CreateExpenseRequest>({
    category_id: '',
    product_id: '',
    quantity: '',
    amount: '',
    description: '',
    expense_date: new Date().toISOString().split('T')[0],
    vendor: '',
    reference_number: '',
    notes: '',
  });
  const [formErrors, setFormErrors] = useState<{ amount?: string; description?: string; general?: string }>({});
  const [isManualAmount, setIsManualAmount] = useState(false);
  const [isManualDescription, setIsManualDescription] = useState(false);

  // Fetch data
  useEffect(() => {
    const fetchData = async () => {
      setIsLoading(true);
      try {
        const [expensesResult, categoriesResult, productsResult, summaryResult] = await Promise.all([
          api.getExpenses({
            category_id: selectedCategory !== 'all' ? selectedCategory : undefined,
            start_date: dateRange.start,
            end_date: dateRange.end,
          }),
          api.getExpenseCategories(),
          api.getProducts({ active: true }),
          api.getExpenseSummary({
            start_date: dateRange.start,
            end_date: dateRange.end,
          }),
        ]);

        if (expensesResult.data) setExpenses(expensesResult.data.expenses);
        if (categoriesResult.data) setCategories(categoriesResult.data);
        if (productsResult.data) setProducts(productsResult.data);
        if (summaryResult.data) setSummary(summaryResult.data);
      } catch (error) {
        console.error('Failed to fetch data:', error);
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();
  }, [selectedCategory, dateRange]);

  // Filter expenses by search
  const filteredExpenses = expenses.filter((expense) => {
    const matchesSearch =
      expense.description.toLowerCase().includes(searchQuery.toLowerCase()) ||
      expense.vendor?.toLowerCase().includes(searchQuery.toLowerCase()) ||
      expense.category_name?.toLowerCase().includes(searchQuery.toLowerCase());
    return matchesSearch;
  });

  // Reset form
  const resetForm = () => {
    setFormData({
      category_id: '',
      product_id: '',
      quantity: '',
      amount: '',
      description: '',
      expense_date: new Date().toISOString().split('T')[0],
      vendor: '',
      reference_number: '',
      notes: '',
    });
    setEditingExpense(null);
    setFormErrors({});
    setIsManualAmount(false);
    setIsManualDescription(false);
  };

  // Check if selected category is Inventory Purchase
  const isInventoryPurchase = useCallback(() => {
    const category = categories.find(c => c.id === formData.category_id);
    return category?.name === 'Inventory Purchase';
  }, [categories, formData.category_id]);

  // Auto-calculate amount when product or quantity changes
  useEffect(() => {
    if (isInventoryPurchase() && formData.product_id && formData.quantity && !isManualAmount) {
      const product = products.find(p => p.id === formData.product_id);
      if (product && product.cost) {
        const qty = parseFloat(formData.quantity);
        const cost = parseFloat(product.cost);
        const calculatedAmount = (qty * cost).toFixed(2);
        setFormData(prev => ({ ...prev, amount: calculatedAmount }));
      }
    }
  }, [formData.product_id, formData.quantity, isManualAmount, products, isInventoryPurchase]);

  // Auto-generate description when product or quantity changes
  useEffect(() => {
    if (isInventoryPurchase() && formData.product_id && formData.quantity && !isManualDescription) {
      const product = products.find(p => p.id === formData.product_id);
      if (product) {
        const description = `${product.name} x ${formData.quantity}`;
        setFormData(prev => ({ ...prev, description }));
      }
    }
  }, [formData.product_id, formData.quantity, isManualDescription, products, isInventoryPurchase]);

  // Open create dialog
  const openCreateDialog = () => {
    resetForm();
    setDialogOpen(true);
  };

  // Open edit dialog
  const openEditDialog = (expense: Expense) => {
    setEditingExpense(expense);
    setFormData({
      category_id: expense.category_id || '',
      product_id: expense.product_id || '',
      quantity: expense.quantity || '',
      amount: expense.amount,
      description: expense.description,
      expense_date: expense.expense_date,
      vendor: expense.vendor || '',
      reference_number: expense.reference_number || '',
      notes: expense.notes || '',
    });
    setIsManualAmount(!!expense.product_id); // If it has product, user may have overridden
    setIsManualDescription(true); // Default to manual for edits to preserve existing description
    setDialogOpen(true);
  };

  // Handle submit
  const handleSubmit = async () => {
    // Validate form
    const errors: { amount?: string; description?: string; general?: string } = {};
    if (!formData.amount || parseFloat(formData.amount) <= 0) {
      errors.amount = 'Valid amount is required';
    }
    if (!formData.description.trim()) {
      errors.description = 'Description is required';
    }

    if (Object.keys(errors).length > 0) {
      setFormErrors(errors);
      return;
    }

    setFormErrors({});
    setIsSubmitting(true);

    try {
      const expenseData: CreateExpenseRequest = {
        ...formData,
        category_id: (formData.category_id && formData.category_id !== 'none') ? formData.category_id : undefined,
        product_id: formData.product_id || undefined,
        quantity: formData.quantity || undefined,
        vendor: formData.vendor || undefined,
        reference_number: formData.reference_number || undefined,
        notes: formData.notes || undefined,
      };

      if (editingExpense) {
        const result = await api.updateExpense(editingExpense.id, expenseData);
        if (result.error) {
          // Log as warning for validation errors, error for system errors
          console.warn('Update expense failed:', result.error);
          setFormErrors({ general: result.error });
          return;
        }
        if (result.data) {
          setExpenses((prev) =>
            prev.map((e) => (e.id === editingExpense.id ? result.data! : e))
          );
        }
      } else {
        const result = await api.createExpense(expenseData);
        if (result.error) {
          console.warn('Create expense failed:', result.error);
          setFormErrors({ general: result.error });
          return;
        }
        if (result.data) {
          setExpenses((prev) => [result.data!, ...prev]);
        }
      }

      // Refresh summary
      const summaryResult = await api.getExpenseSummary({
        start_date: dateRange.start,
        end_date: dateRange.end,
      });
      if (summaryResult.data) setSummary(summaryResult.data);

      setDialogOpen(false);
      resetForm();
    } catch (error) {
      console.error('Failed to save expense:', error);
      setFormErrors({ general: 'Failed to save expense. Please try again.' });
    } finally {
      setIsSubmitting(false);
    }
  };

  // Handle delete
  const handleDelete = async () => {
    if (!deletingExpense) return;

    setIsSubmitting(true);
    setDeleteError(null);

    try {
      const result = await api.deleteExpense(deletingExpense.id);
      
      if (result.error) {
        setDeleteError(result.error);
        return;
      }

      setExpenses((prev) => prev.filter((e) => e.id !== deletingExpense.id));

      // Refresh summary
      const summaryResult = await api.getExpenseSummary({
        start_date: dateRange.start,
        end_date: dateRange.end,
      });
      if (summaryResult.data) setSummary(summaryResult.data);

      setDeleteDialogOpen(false);
      setDeletingExpense(null);
    } catch (error) {
      console.error('Failed to delete expense:', error);
      setDeleteError('An unexpected error occurred');
    } finally {
      setIsSubmitting(false);
    }
  };

  const formatCurrency = (amount: string | number) => {
    const num = typeof amount === 'string' ? parseFloat(amount) : amount;
    return new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: 'IDR',
      minimumFractionDigits: 0,
    }).format(num);
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('id-ID', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  };

  return (
    <div className="flex flex-col h-full">
      <Header title="Expenses" />

      <div className="flex-1 p-6 overflow-auto">
        {/* Summary Cards */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Total Expenses</CardTitle>
              <TrendingDown className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-destructive">
                {summary ? formatCurrency(summary.total_amount) : 'Rp 0'}
              </div>
              <p className="text-xs text-muted-foreground">
                {summary?.expense_count || 0} transactions
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Top Category</CardTitle>
              <Receipt className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {summary?.by_category?.[0]?.category_name || '-'}
              </div>
              <p className="text-xs text-muted-foreground">
                {summary?.by_category?.[0]
                  ? formatCurrency(summary.by_category[0].total_amount)
                  : 'No expenses yet'}
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Period</CardTitle>
              <Calendar className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-lg font-medium">
                {dateRange.start && dateRange.end 
                  ? `${formatDate(dateRange.start)} - ${formatDate(dateRange.end)}`
                  : 'No period selected'}
              </div>
              <p className="text-xs text-muted-foreground">
                {summary?.by_category?.length || 0} categories
              </p>
            </CardContent>
          </Card>
        </div>

        {/* Toolbar */}
        <div className="flex flex-col sm:flex-row gap-4 mb-6">
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Search expenses..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-9"
            />
          </div>
          <Select value={selectedCategory} onValueChange={setSelectedCategory}>
            <SelectTrigger className="w-48">
              <SelectValue placeholder="All Categories" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Categories</SelectItem>
              {categories.map((category) => (
                <SelectItem key={category.id} value={category.id}>
                  {category.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <div className="flex gap-2">
            <Input
              type="date"
              value={dateRange.start}
              onChange={(e) => setDateRange((prev) => ({ ...prev, start: e.target.value }))}
              className="w-40"
            />
            <Input
              type="date"
              value={dateRange.end}
              onChange={(e) => setDateRange((prev) => ({ ...prev, end: e.target.value }))}
              className="w-40"
            />
          </div>
          <Button onClick={openCreateDialog}>
            <Plus className="h-4 w-4 mr-2" />
            Add Expense
          </Button>
        </div>

        {/* Expenses table */}
        {isLoading ? (
          <div className="flex items-center justify-center h-64">
            <Loader2 className="h-8 w-8 animate-spin text-primary" />
          </div>
        ) : filteredExpenses.length === 0 ? (
          <Card>
            <CardContent className="flex flex-col items-center justify-center py-12">
              <Receipt className="h-12 w-12 text-muted-foreground mb-4" />
              <p className="text-muted-foreground">No expenses found</p>
              <Button variant="link" onClick={openCreateDialog}>
                Add your first expense
              </Button>
            </CardContent>
          </Card>
        ) : (
          <Card>
            <CardHeader>
              <CardTitle>Expenses ({filteredExpenses.length})</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead>
                    <tr className="border-b text-left text-sm text-muted-foreground">
                      <th className="pb-3 font-medium">Date</th>
                      <th className="pb-3 font-medium">Description</th>
                      <th className="pb-3 font-medium">Category</th>
                      <th className="pb-3 font-medium">User</th>
                      <th className="pb-3 font-medium">Vendor</th>
                      <th className="pb-3 font-medium text-right">Amount</th>
                      <th className="pb-3 font-medium text-right">Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {filteredExpenses.map((expense) => (
                      <tr key={expense.id} className="border-b last:border-0">
                        <td className="py-3 text-sm">
                          {formatDate(expense.expense_date)}
                        </td>
                        <td className="py-3">
                          <div>
                            <p className="font-medium">{expense.description}</p>
                            {expense.reference_number && (
                              <p className="text-xs text-muted-foreground">
                                Ref: {expense.reference_number}
                              </p>
                            )}
                          </div>
                        </td>
                        <td className="py-3 text-sm">
                          <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-blue-600 text-white dark:bg-blue-600/90 dark:text-white">
                            {expense.category_name || 'Uncategorized'}
                          </span>
                        </td>
                        <td className="py-3 text-sm">
                          {expense.created_by_name || 'Unknown'}
                        </td>
                        <td className="py-3 text-sm text-muted-foreground">
                          {expense.vendor || '-'}
                        </td>
                        <td className="py-3 text-right font-medium text-destructive">
                          {formatCurrency(expense.amount)}
                        </td>
                        <td className="py-3 text-right">
                          <div className="flex items-center justify-end gap-1">
                            <Button
                              variant="ghost"
                              size="icon"
                              onClick={() => openEditDialog(expense)}
                            >
                              <Pencil className="h-4 w-4" />
                            </Button>
                            <Button
                              variant="ghost"
                              size="icon"
                              onClick={() => {
                                setDeletingExpense(expense);
                                setDeleteDialogOpen(true);
                              }}
                            >
                              <Trash2 className="h-4 w-4 text-destructive" />
                            </Button>
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </CardContent>
          </Card>
        )}
      </div>

      {/* Create/Edit Dialog */}
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="sm:max-w-lg max-h-[85vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>
              {editingExpense ? 'Edit Expense' : 'Add Expense'}
            </DialogTitle>
            <DialogDescription>
              {editingExpense
                ? 'Update the expense details below.'
                : 'Fill in the details for the new expense.'}
            </DialogDescription>
          </DialogHeader>

          <div className="grid gap-4 py-4">
            {formErrors.general && (
              <div className="p-3 rounded-md bg-destructive/10 border border-destructive/20">
                <p className="text-sm text-destructive">{formErrors.general}</p>
              </div>
            )}

            <div className="grid grid-cols-2 gap-4">
              <div className="grid gap-2">
                <Label htmlFor="category">Category</Label>
                <Select
                  value={formData.category_id}
                  onValueChange={(value) => {
                    if (!editingExpense) {
                      // For New Expense: reset everything except date when category changes
                      setFormData({
                        category_id: value,
                        product_id: '',
                        quantity: '',
                        amount: '',
                        description: '',
                        expense_date: formData.expense_date,
                        vendor: '',
                        reference_number: '',
                        notes: '',
                      });
                    } else {
                      // For Editing: just clear product/quantity-specific fields to prevent logic conflicts
                      setFormData({ ...formData, category_id: value, product_id: '', quantity: '' });
                    }
                    setIsManualAmount(false);
                    setIsManualDescription(false);
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
                {isInventoryPurchase() ? (
                  <>
                    <Label htmlFor="quantity" className={!formData.category_id ? 'opacity-50' : ''}>Quantity *</Label>
                    <Input
                      id="quantity"
                      type="number"
                      min="1"
                      value={formData.quantity}
                      onChange={(e) =>
                        setFormData({ ...formData, quantity: e.target.value })
                      }
                      placeholder="Enter quantity"
                      disabled={!formData.category_id}
                    />
                  </>
                ) : (
                  <>
                    <Label htmlFor="vendor" className={!formData.category_id ? 'opacity-50' : ''}>Vendor</Label>
                    <Input
                      id="vendor"
                      value={formData.vendor}
                      onChange={(e) =>
                        setFormData({ ...formData, vendor: e.target.value })
                      }
                      placeholder="e.g., PLN"
                      disabled={!formData.category_id}
                    />
                  </>
                )}
              </div>
            </div>

            <div className="grid gap-2">
              {isInventoryPurchase() ? (
                <>
                  <Label htmlFor="product" className={!formData.category_id ? 'opacity-50' : ''}>Product *</Label>
                  <Select
                    value={formData.product_id || 'none'}
                    onValueChange={(value) =>
                      setFormData({ ...formData, product_id: value === 'none' ? '' : value })
                    }
                    disabled={!formData.category_id}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="Select product" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="none">Select a product</SelectItem>
                      {products.map((product) => (
                        <SelectItem key={product.id} value={product.id}>
                          {product.name} - {formatCurrency(parseFloat(product.cost || '0'))}/unit
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </>
              ) : (
                <>
                  <div className="flex items-center justify-between">
                    <Label htmlFor="description" className={!formData.category_id ? 'opacity-50' : ''}>Description *</Label>
                  </div>
                  <Input
                    id="description"
                    value={formData.description}
                    onChange={(e) => {
                      setFormData({ ...formData, description: e.target.value });
                      if (formErrors.description) setFormErrors({ ...formErrors, description: undefined });
                    }}
                    placeholder="e.g., Monthly electricity bill"
                    className={formErrors.description ? 'border-destructive' : ''}
                    disabled={!formData.category_id}
                  />
                  {formErrors.description && (
                    <p className="text-sm text-destructive">{formErrors.description}</p>
                  )}
                </>
              )}
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="grid gap-2">
                <div className="flex items-center justify-between">
                  <Label htmlFor="amount" className={!formData.category_id ? 'opacity-50' : ''}>Amount (IDR) *</Label>
                  {isInventoryPurchase() && formData.product_id && formData.quantity && (
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      className="h-6 text-xs"
                      onClick={() => setIsManualAmount(!isManualAmount)}
                      disabled={!formData.category_id}
                    >
                      {isManualAmount ? 'Auto-calculate' : 'Manual edit'}
                    </Button>
                  )}
                </div>
                <Input
                  id="amount"
                  type="number"
                  value={formData.amount}
                  onChange={(e) => {
                    setFormData({ ...formData, amount: e.target.value });
                    if (formErrors.amount) setFormErrors({ ...formErrors, amount: undefined });
                    if (isInventoryPurchase()) setIsManualAmount(true);
                  }}
                  placeholder="100000"
                  className={formErrors.amount ? 'border-destructive' : ''}
                  disabled={!formData.category_id || (isInventoryPurchase() && !isManualAmount && !!formData.product_id && !!formData.quantity)}
                />
                {isInventoryPurchase() && !isManualAmount && formData.product_id && formData.quantity && (
                  <p className="text-xs text-muted-foreground">Auto-calculated from product cost × quantity</p>
                )}
                {formErrors.amount && (
                  <p className="text-sm text-destructive">{formErrors.amount}</p>
                )}
              </div>
              <div className="grid gap-2">
                <Label htmlFor="expense_date" className={!formData.category_id ? 'opacity-50' : ''}>Date *</Label>
                <Input
                  id="expense_date"
                  type="date"
                  value={formData.expense_date}
                  onChange={(e) =>
                    setFormData({ ...formData, expense_date: e.target.value })
                  }
                  disabled={!formData.category_id}
                />
              </div>
            </div>

            {/* Conditional: Remaining sections for Inventory Purchase */}
            {isInventoryPurchase() && (
              <>
                <div className="flex items-center gap-3 p-3 my-4 rounded-md bg-primary/10 border border-primary/30">
                  <Info className="h-5 w-5 text-primary flex-shrink-0" />
                  <p className="text-sm text-primary">
                    Inventory Purchase: Amount is auto-calculated based on the selected Product and Quantity.
                  </p>
                </div>

                <div className="grid gap-2">
                  <Label htmlFor="vendor" className={!formData.category_id ? 'opacity-50' : ''}>Vendor</Label>
                  <Input
                    id="vendor"
                    value={formData.vendor}
                    onChange={(e) =>
                      setFormData({ ...formData, vendor: e.target.value })
                    }
                    placeholder="e.g., Supplier Name"
                    disabled={!formData.category_id}
                  />
                </div>

                <div className="grid gap-2">
                  <div className="flex items-center justify-between">
                    <Label htmlFor="description" className={!formData.category_id ? 'opacity-50' : ''}>Description *</Label>
                    {formData.product_id && formData.quantity && (
                      <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        className="h-6 text-xs"
                        onClick={() => setIsManualDescription(!isManualDescription)}
                        disabled={!formData.category_id}
                      >
                        {isManualDescription ? 'Auto-generate' : 'Manual edit'}
                      </Button>
                    )}
                  </div>
                  <Input
                    id="description"
                    value={formData.description}
                    onChange={(e) => {
                      setFormData({ ...formData, description: e.target.value });
                      if (formErrors.description) setFormErrors({ ...formErrors, description: undefined });
                      setIsManualDescription(true);
                    }}
                    placeholder="e.g., Stock for February"
                    className={formErrors.description ? 'border-destructive' : ''}
                    disabled={!formData.category_id || (!isManualDescription && !!formData.product_id && !!formData.quantity)}
                  />
                  {!isManualDescription && formData.product_id && formData.quantity && (
                    <p className="text-xs text-muted-foreground">Auto-generated from product and quantity</p>
                  )}
                </div>
              </>
            )}

            <div className="grid gap-2">
              <Label htmlFor="reference_number" className={!formData.category_id ? 'opacity-50' : ''}>Reference Number</Label>
              <Input
                id="reference_number"
                value={formData.reference_number}
                onChange={(e) =>
                  setFormData({ ...formData, reference_number: e.target.value })
                }
                placeholder="e.g., Invoice #12345"
                disabled={!formData.category_id}
              />
            </div>

            <div className="grid gap-2">
              <Label htmlFor="notes" className={!formData.category_id ? 'opacity-50' : ''}>Notes</Label>
              <Input
                id="notes"
                value={formData.notes}
                onChange={(e) =>
                  setFormData({ ...formData, notes: e.target.value })
                }
                placeholder="Additional notes..."
                disabled={!formData.category_id}
              />
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setDialogOpen(false)}>
              Cancel
            </Button>
            <Button onClick={handleSubmit} disabled={isSubmitting || !formData.category_id}>
              {isSubmitting ? (
                <>
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  Saving...
                </>
              ) : editingExpense ? (
                'Update Expense'
              ) : (
                'Create Expense'
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation Dialog */}
      <Dialog open={deleteDialogOpen} onOpenChange={(open) => {
        setDeleteDialogOpen(open);
        if (!open) setDeleteError(null);
      }}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete Expense</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete this expense? This action cannot be undone.
            </DialogDescription>
          </DialogHeader>

          {deleteError && (
            <div className="p-3 rounded-md bg-destructive/10 border border-destructive/20">
              <p className="text-sm text-destructive">{deleteError}</p>
            </div>
          )}

          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteDialogOpen(false)}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={handleDelete} disabled={isSubmitting || !!deleteError}>
              {isSubmitting ? (
                <>
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  Deleting...
                </>
              ) : (
                'Delete'
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

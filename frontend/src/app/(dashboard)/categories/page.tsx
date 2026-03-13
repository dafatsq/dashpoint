'use client';

import { useEffect, useState } from 'react';
import { Header } from '@/components/layout/header';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from '@/components/ui/tabs';
import {
  Search,
  Plus,
  Pencil,
  Trash2,
  Loader2,
  Boxes,
  Wallet,
  AlertCircle,
  RefreshCcw,
  ShieldAlert,
  RotateCcw,
} from 'lucide-react';
import api from '@/lib/api';
import { Category, ExpenseCategory } from '@/types';
import { useAuth, PERMISSIONS } from '@/contexts/auth-context';

type CategoryType = 'product' | 'expense';
type ViewMode = 'active' | 'archived';

interface CategoryFormData {
  name: string;
  description: string;
}

export default function CategoriesPage() {
  const { hasPermission } = useAuth();
  const canViewCategories = hasPermission(PERMISSIONS.CATEGORIES_VIEW);
  const canManageCategories = hasPermission(PERMISSIONS.CATEGORIES_MANAGE);
  
  const [activeTab, setActiveTab] = useState<CategoryType>('product');
  const [viewMode, setViewMode] = useState<ViewMode>('active');
  const [productCategories, setProductCategories] = useState<Category[]>([]);
  const [expenseCategories, setExpenseCategories] = useState<ExpenseCategory[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');

  // States
  const [dialogOpen, setDialogOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [permanentDeleteDialogOpen, setPermanentDeleteDialogOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [editingCategory, setEditingCategory] = useState<{ id: string; type: CategoryType } | null>(null);
  const [deletingCategory, setDeletingCategory] = useState<{ id: string; name: string; type: CategoryType } | null>(null);
  const [error, setError] = useState<string | null>(null);

  // Form state
  const [formData, setFormData] = useState<CategoryFormData>({
    name: '',
    description: '',
  });

  // Fetch data
  const fetchData = async () => {
    setIsLoading(true);
    try {
      const status = viewMode === 'active' ? 'active' : 'archived';
      if (activeTab === 'product') {
        const result = await api.getCategories(status);
        if (result.data) {
          setProductCategories(result.data);
        }
      } else {
        const result = await api.getExpenseCategories(status);
        if (result.data) {
          setExpenseCategories(result.data);
        }
      }
    } catch (err) {
      console.error('Failed to fetch categories:', err);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    // Clear the current view's categories when switching mode/tab to avoid stale data
    if (activeTab === 'product') {
      setProductCategories([]);
    } else {
      setExpenseCategories([]);
    }
    fetchData();
  }, [activeTab, viewMode]);

  // Filter categories and ensure they match the current viewMode
  const filteredCategories = (activeTab === 'product' ? productCategories : expenseCategories)
    .filter(c => {
      const matchesSearch = c.name.toLowerCase().includes(searchQuery.toLowerCase());
      const matchesStatus = viewMode === 'active' ? c.is_active : !c.is_active;
      return matchesSearch && matchesStatus;
    });

  // Reset form
  const resetForm = () => {
    setFormData({
      name: '',
      description: '',
    });
    setEditingCategory(null);
    setError(null);
  };

  // Open create dialog
  const openCreateDialog = () => {
    resetForm();
    setDialogOpen(true);
  };

  // Open edit dialog
  const openEditDialog = (cat: Category | ExpenseCategory) => {
    setEditingCategory({ id: cat.id, type: activeTab });
    setFormData({
      name: cat.name,
      description: cat.description || '',
    });
    setDialogOpen(true);
  };

  // Handle submit
  const handleSubmit = async () => {
    if (!formData.name.trim()) {
      setError('Category name is required');
      return;
    }

    setIsSubmitting(true);
    setError(null);

    try {
      if (activeTab === 'product') {
        if (editingCategory) {
          const result = await api.updateCategory(editingCategory.id, {
            name: formData.name,
            description: formData.description,
          });
          if (result.error) throw new Error(result.error);
        } else {
          const result = await api.createCategory({
            name: formData.name,
            description: formData.description,
          });
          if (result.error) throw new Error(result.error);
        }
      } else {
        if (editingCategory) {
          const result = await api.updateExpenseCategory(editingCategory.id, {
            name: formData.name,
            description: formData.description,
          });
          if (result.error) throw new Error(result.error);
        } else {
          const result = await api.createExpenseCategory(formData.name, formData.description);
          if (result.error) throw new Error(result.error);
        }
      }

      setDialogOpen(false);
      resetForm();
      fetchData();
    } catch (err: any) {
      setError(err.message || 'Failed to save category');
    } finally {
      setIsSubmitting(false);
    }
  };

  // Handle archive (soft delete)
  const handleArchive = async () => {
    if (!deletingCategory) return;

    setIsSubmitting(true);
    setError(null);

    try {
      const result = deletingCategory.type === 'product'
        ? await api.deleteCategory(deletingCategory.id)
        : await api.deleteExpenseCategory(deletingCategory.id);

      if (result.error) throw new Error(result.error);

      setDeleteDialogOpen(false);
      setDeletingCategory(null);
      fetchData();
    } catch (err: any) {
      setError(err.message || 'Failed to archive category');
    } finally {
      setIsSubmitting(false);
    }
  };

  // Handle restore
  const handleRestore = async (id: string, name: string) => {
    setIsLoading(true);
    try {
      const result = activeTab === 'product'
        ? await api.updateCategory(id, { name, is_active: true } as any)
        : await api.updateExpenseCategory(id, { is_active: true });

      if (result.error) throw new Error(result.error);
      fetchData();
    } catch (err: any) {
      alert(err.message || 'Failed to restore category');
    } finally {
      setIsLoading(false);
    }
  };

  // Handle permanent delete
  const handlePermanentDelete = async () => {
    if (!deletingCategory) return;

    setIsSubmitting(true);
    setError(null);

    try {
      const result = deletingCategory.type === 'product'
        ? await api.permanentDeleteCategory(deletingCategory.id)
        : await api.permanentDeleteExpenseCategory(deletingCategory.id);

      if (result.error) throw new Error(result.error);

      setPermanentDeleteDialogOpen(false);
      setDeletingCategory(null);
      fetchData();
    } catch (err: any) {
      setError(err.message || 'Failed to permanently delete category');
    } finally {
      setIsSubmitting(false);
    }
  };

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
          <div className="flex flex-col xl:flex-row xl:items-center justify-between gap-4">
            <Tabs value={activeTab} onValueChange={(v) => setActiveTab(v as CategoryType)} className="w-full xl:w-auto">
              <TabsList className="grid grid-cols-2 w-full xl:min-w-[400px]">
                <TabsTrigger value="product" className="flex items-center gap-2">
                  <Boxes className="h-4 w-4" />
                  Product Categories
                </TabsTrigger>
                <TabsTrigger value="expense" className="flex items-center gap-2">
                  <Wallet className="h-4 w-4" />
                  Expense Categories
                </TabsTrigger>
              </TabsList>
            </Tabs>

            <div className="flex items-center gap-3 w-full xl:w-auto flex-wrap sm:flex-nowrap">
              <div className="relative flex-1 md:min-w-[240px]">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                <Input
                  placeholder="Search categories..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="pl-9 w-full"
                />
              </div>
              {canManageCategories && (
                <Button onClick={openCreateDialog}>
                  <Plus className="h-4 w-4 mr-2" />
                  Add Category
                </Button>
              )}
            </div>
          </div>

          <Card className="border-none shadow-none bg-transparent">
            <Tabs value={viewMode} onValueChange={(v) => setViewMode(v as ViewMode)}>
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

              <CategoryList
                isLoading={isLoading}
                categories={filteredCategories}
                onEdit={openEditDialog}
                onArchive={(id, name) => {
                  setDeletingCategory({ id, name, type: activeTab });
                  setDeleteDialogOpen(true);
                }}
                onRestore={(id, name) => handleRestore(id, name)}
                onPermanentDelete={(id, name) => {
                  setDeletingCategory({ id, name, type: activeTab });
                  setPermanentDeleteDialogOpen(true);
                }}
                viewMode={viewMode}
                type={activeTab}
                canManageCategories={canManageCategories}
              />
            </Tabs>
          </Card>
        </div>
      </div>

      {/* Create/Edit Dialog */}
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>
              {editingCategory ? 'Edit Category' : 'Add New Category'}
            </DialogTitle>
            <DialogDescription>
              {activeTab === 'product' ? 'Product categories help organize your inventory.' : 'Expense categories track your business spending.'}
            </DialogDescription>
          </DialogHeader>

          <div className="grid gap-4 py-4">
            {error && (
              <div className="flex items-center gap-2 p-3 rounded-md bg-destructive/10 text-destructive text-sm border border-destructive/20">
                <AlertCircle className="h-4 w-4" />
                {error}
              </div>
            )}
            <div className="grid gap-2">
              <Label htmlFor="name">Name <span className="text-destructive">*</span></Label>
              <Input
                id="name"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                placeholder="Category name"
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="description">Description (Optional)</Label>
              <Input
                id="description"
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                placeholder="Brief description"
              />
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setDialogOpen(false)} disabled={isSubmitting}>
              Cancel
            </Button>
            <Button onClick={handleSubmit} disabled={isSubmitting}>
              {isSubmitting && <Loader2 className="h-4 w-4 mr-2 animate-spin" />}
              {editingCategory ? 'Update' : 'Create'} Category
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Archive Dialog */}
      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Archive Category</DialogTitle>
            <DialogDescription>
              Are you sure you want to archive <span className="font-semibold text-foreground">{deletingCategory?.name}</span>? 
              It will be moved to the Archived tab.
            </DialogDescription>
          </DialogHeader>

          {error && (
            <div className="flex items-center gap-2 p-3 rounded-md bg-destructive/10 text-destructive text-sm border border-destructive/20">
              <AlertCircle className="h-4 w-4" />
              {error}
            </div>
          )}

          <DialogFooter className="gap-2 sm:gap-0">
            <Button variant="outline" onClick={() => setDeleteDialogOpen(false)} disabled={isSubmitting}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={handleArchive} disabled={isSubmitting}>
              {isSubmitting && <Loader2 className="h-4 w-4 mr-2 animate-spin" />}
              Archive Category
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Permanent Delete Dialog */}
      <Dialog open={permanentDeleteDialogOpen} onOpenChange={setPermanentDeleteDialogOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle className="text-destructive flex items-center gap-2">
              <ShieldAlert className="h-5 w-5" />
              Permanent Delete
            </DialogTitle>
            <DialogDescription>
              This will permanently delete <span className="font-semibold text-foreground">{deletingCategory?.name}</span>.
              This action <span className="font-bold text-destructive">cannot be undone</span>.
            </DialogDescription>
          </DialogHeader>

          {error && (
            <div className="flex items-center gap-2 p-3 rounded-md bg-destructive/10 text-destructive text-sm border border-destructive/20">
              <AlertCircle className="h-4 w-4" />
              {error}
            </div>
          )}

          <DialogFooter className="gap-2 sm:gap-0">
            <Button variant="outline" onClick={() => setPermanentDeleteDialogOpen(false)} disabled={isSubmitting}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={handlePermanentDelete} disabled={isSubmitting}>
              {isSubmitting && <Loader2 className="h-4 w-4 mr-2 animate-spin" />}
              Delete Permanently
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

interface ListProps {
  isLoading: boolean;
  categories: any[];
  onEdit: (cat: any) => void;
  onArchive: (id: string, name: string) => void;
  onRestore: (id: string, name: string) => void;
  onPermanentDelete: (id: string, name: string) => void;
  viewMode: ViewMode;
  type: CategoryType;
  canManageCategories: boolean;
}

function CategoryList({ isLoading, categories, onEdit, onArchive, onRestore, onPermanentDelete, viewMode, type, canManageCategories }: ListProps) {
  if (isLoading) {
    return (
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {[1, 2, 3].map((i) => (
          <Card key={i} className="animate-pulse">
            <div className="h-32 bg-muted/50 rounded-lg"></div>
          </Card>
        ))}
      </div>
    );
  }

  if (categories.length === 0) {
    return (
      <Card className="border-dashed">
        <CardContent className="flex flex-col items-center justify-center py-12 text-center">
          <div className="h-12 w-12 rounded-full bg-muted flex items-center justify-center mb-4">
            {viewMode === 'active' ? <Plus className="h-6 w-6 text-muted-foreground" /> : <Trash2 className="h-6 w-6 text-muted-foreground" />}
          </div>
          <h3 className="font-medium text-lg">No {viewMode} categories found</h3>
          <p className="text-muted-foreground max-w-sm">
            {viewMode === 'active' 
              ? 'Try adjusting your search or add a new category to get started.' 
              : 'Archived categories will appear here.'}
          </p>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      {categories.map((cat) => (
        <Card key={cat.id} className="group hover:border-primary/50 transition-all duration-200 shadow-sm hover:shadow-md">
          <CardHeader className="pb-2">
            <div className="flex items-start justify-between">
              <div className="space-y-1">
                <CardTitle className="text-lg">
                  {cat.name}
                </CardTitle>
                <CardDescription className="line-clamp-1 h-5">
                  {cat.description || 'No description'}
                </CardDescription>
              </div>
              {canManageCategories && (
                <div className="flex items-center gap-1">
                  {viewMode === 'active' ? (
                    <>
                      <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => onEdit(cat)}>
                        <Pencil className="h-4 w-4" />
                      </Button>
                      <Button variant="ghost" size="icon" className="h-8 w-8 text-destructive" onClick={() => onArchive(cat.id, cat.name)}>
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </>
                  ) : (
                    <>
                      <Button variant="ghost" size="icon" className="h-8 w-8 text-blue-600" onClick={() => onRestore(cat.id, cat.name)}>
                        <RotateCcw className="h-4 w-4" />
                      </Button>
                      <Button variant="ghost" size="icon" className="h-8 w-8 text-destructive" onClick={() => onPermanentDelete(cat.id, cat.name)}>
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </>
                  )}
                </div>
              )}
            </div>
          </CardHeader>
          <CardContent>
            <div className="flex items-center justify-between text-xs text-muted-foreground mt-2 pt-2 border-t">
              <div className="flex items-center gap-1.5">
                {type === 'product' && (
                  <span className="px-1.5 py-0.5 rounded bg-primary/10 text-primary font-medium">
                    {cat.product_count || 0} Products
                  </span>
                )}
                {viewMode === 'archived' && (
                  <span className="px-1.5 py-0.5 rounded bg-amber-100 text-amber-700 font-medium">
                    Archived
                  </span>
                )}
              </div>
              <span>Updated {new Date(cat.updated_at).toLocaleDateString()}</span>
            </div>
          </CardContent>
        </Card>
      ))}
    </div>
  );
}

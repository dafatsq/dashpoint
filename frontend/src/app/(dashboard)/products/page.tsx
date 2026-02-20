'use client';

import { useEffect, useState } from 'react';
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
  Package,
  RotateCcw,
  Archive,
  ImageIcon,
} from 'lucide-react';
import api from '@/lib/api';
import { Product, Category, CreateProductRequest, UpdateProductRequest } from '@/types';
import { useAuth, PERMISSIONS } from '@/contexts/auth-context';
import { ImageUpload } from '@/components/ui/image-upload';

// Helper to get full image URL
function getImageUrl(path: string | null | undefined): string {
  if (!path) return '';
  if (path.startsWith('http://') || path.startsWith('https://')) {
    return path;
  }
  // Use window.location for dynamic base URL, fallback to localhost:8080
  const baseUrl = typeof window !== 'undefined' && window.location.hostname !== 'localhost'
    ? `${window.location.protocol}//${window.location.hostname}:8080`
    : 'http://localhost:8080';
  return `${baseUrl}${path}`;
}

// Helper to get numeric values from Product
function getProductPrice(product: Product): number {
  return parseFloat(product.price) || 0;
}

function getProductQuantity(product: Product): number {
  if (product.inventory?.quantity) {
    return parseFloat(product.inventory.quantity) || 0;
  }
  return 0;
}

function getProductMinQuantity(product: Product): number {
  if (product.inventory?.low_stock_threshold) {
    return parseFloat(product.inventory.low_stock_threshold) || 0;
  }
  return 0;
}

interface FormData {
  name: string;
  description: string;
  sku: string;
  barcode: string;
  price: string;
  cost: string;
  tax_rate: string;
  initial_quantity: string;
  low_stock_threshold: string;
  category_id: string;
  image_url: string;
}

export default function ProductsPage() {
  const { hasPermission } = useAuth();
  const canCreate = hasPermission(PERMISSIONS.PRODUCTS_CREATE);
  const canEdit = hasPermission(PERMISSIONS.PRODUCTS_EDIT);
  const canDelete = hasPermission(PERMISSIONS.PRODUCTS_DELETE);

  const [products, setProducts] = useState<Product[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedCategory, setSelectedCategory] = useState<string>('all');
  const [viewMode, setViewMode] = useState<'active' | 'archived'>('active');

  // Dialog states
  const [dialogOpen, setDialogOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [permanentDeleteDialogOpen, setPermanentDeleteDialogOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [editingProduct, setEditingProduct] = useState<Product | null>(null);
  const [deletingProduct, setDeletingProduct] = useState<Product | null>(null);
  const [deleteError, setDeleteError] = useState<string | null>(null);

  // Form state
  const [formData, setFormData] = useState<FormData>({
    name: '',
    description: '',
    sku: '',
    barcode: '',
    price: '',
    cost: '',
    tax_rate: '0',
    initial_quantity: '',
    low_stock_threshold: '5',
    category_id: '',
    image_url: '',
  });
  const [formErrors, setFormErrors] = useState<{ name?: string; price?: string; sku?: string; barcode?: string; general?: string }>({});

  // Fetch data
  useEffect(() => {
    const fetchData = async () => {
      setIsLoading(true);
      try {
        const [productsResult, categoriesResult] = await Promise.all([
          api.getProducts({ active: viewMode === 'active' }),
          api.getCategories(),
        ]);

        if (productsResult.data) setProducts(productsResult.data);
        if (categoriesResult.data) setCategories(categoriesResult.data);
      } catch (error) {
        console.error('Failed to fetch data:', error);
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();
  }, [viewMode]);

  // Filter products
  const filteredProducts = products.filter((product) => {
    const matchesSearch =
      product.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      product.sku?.toLowerCase().includes(searchQuery.toLowerCase()) ||
      product.barcode?.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesCategory =
      selectedCategory === 'all' || product.category_id === selectedCategory;
    return matchesSearch && matchesCategory;
  });

  // Reset form
  const resetForm = () => {
    setFormData({
      name: '',
      description: '',
      sku: '',
      barcode: '',
      price: '',
      cost: '',
      tax_rate: '0',
      initial_quantity: '',
      low_stock_threshold: '5',
      category_id: '',
      image_url: '',
    });
    setEditingProduct(null);
    setFormErrors({});
  };

  // Open create dialog
  const openCreateDialog = () => {
    resetForm();
    setDialogOpen(true);
  };

  // Open edit dialog
  const openEditDialog = (product: Product) => {
    setEditingProduct(product);
    setFormData({
      name: product.name,
      description: product.description || '',
      sku: product.sku || '',
      barcode: product.barcode || '',
      price: product.price,
      cost: product.cost || '',
      tax_rate: product.tax_rate?.toString() || '0',
      initial_quantity: '',  // Not editable on update
      low_stock_threshold: product.inventory?.low_stock_threshold || '5',
      category_id: product.category_id || '',
      image_url: product.image_url || '',
    });
    setDialogOpen(true);
  };

  // Handle submit
  const handleSubmit = async () => {
    // Validate form
    const errors: { name?: string; price?: string; sku?: string; barcode?: string; general?: string } = {};
    if (!formData.name.trim()) {
      errors.name = 'Product name is required';
    }
    if (!formData.price || parseFloat(formData.price) <= 0) {
      errors.price = 'Valid price is required';
    }

    if (Object.keys(errors).length > 0) {
      setFormErrors(errors);
      return;
    }

    setFormErrors({});
    setIsSubmitting(true);

    try {
      if (editingProduct) {
        const updateData: UpdateProductRequest = {
          name: formData.name,
          description: formData.description || undefined,
          sku: formData.sku || undefined,
          barcode: formData.barcode || undefined,
          price: formData.price,
          cost: formData.cost || undefined,
          tax_rate: formData.tax_rate || undefined,
          category_id: formData.category_id || undefined,
          image_url: formData.image_url, // Empty string clears, undefined skips update
        };
        const result = await api.updateProduct(editingProduct.id, updateData);
        if (result.error) {
          // Handle specific error codes
          if (result.error.includes('SKU')) {
            setFormErrors({ sku: result.error });
          } else if (result.error.includes('Barcode') || result.error.includes('barcode')) {
            setFormErrors({ barcode: result.error });
          } else {
            setFormErrors({ general: result.error });
          }
          return;
        }
        if (result.data) {
          setProducts((prev) =>
            prev.map((p) => (p.id === editingProduct.id ? result.data! : p))
          );
        }
      } else {
        const createData: CreateProductRequest = {
          name: formData.name,
          description: formData.description || undefined,
          sku: formData.sku || undefined,
          barcode: formData.barcode || undefined,
          price: formData.price,
          cost: formData.cost || undefined,
          tax_rate: formData.tax_rate || undefined,
          initial_quantity: formData.initial_quantity || undefined,
          low_stock_threshold: formData.low_stock_threshold || undefined,
          category_id: formData.category_id || undefined,
          image_url: formData.image_url || undefined,
        };
        const result = await api.createProduct(createData);
        if (result.error) {
          // Handle specific error codes
          if (result.error.includes('SKU')) {
            setFormErrors({ sku: result.error });
          } else if (result.error.includes('Barcode') || result.error.includes('barcode')) {
            setFormErrors({ barcode: result.error });
          } else {
            setFormErrors({ general: result.error });
          }
          return;
        }
        if (result.data) {
          setProducts((prev) => [...prev, result.data!]);
        }
      }
      setDialogOpen(false);
      resetForm();
    } catch (error) {
      console.error('Failed to save product:', error);
      setFormErrors({ general: 'Failed to save product. Please try again.' });
    } finally {
      setIsSubmitting(false);
    }
  };

  // Handle delete (soft delete / archive)
  const handleDelete = async () => {
    if (!deletingProduct) return;

    setIsSubmitting(true);

    try {
      await api.deleteProduct(deletingProduct.id);
      setProducts((prev) => prev.filter((p) => p.id !== deletingProduct.id));
      setDeleteDialogOpen(false);
      setDeletingProduct(null);
    } catch (error) {
      console.error('Failed to delete product:', error);
    } finally {
      setIsSubmitting(false);
    }
  };

  // Handle restore
  const handleRestore = async (product: Product) => {
    setIsSubmitting(true);
    try {
      const result = await api.updateProduct(product.id, { is_active: true });
      if (result.data) {
        setProducts((prev) => prev.filter((p) => p.id !== product.id));
      }
    } catch (error) {
      console.error('Failed to restore product:', error);
    } finally {
      setIsSubmitting(false);
    }
  };

  // Handle permanent delete
  const handlePermanentDelete = async () => {
    if (!deletingProduct) return;

    setIsSubmitting(true);
    setDeleteError(null);
    try {
      const result = await api.permanentDeleteProduct(deletingProduct.id);
      if (result.error) {
        setDeleteError(result.error);
      } else {
        setProducts((prev) => prev.filter((p) => p.id !== deletingProduct.id));
        setPermanentDeleteDialogOpen(false);
        setDeletingProduct(null);
      }
    } catch (error) {
      console.error('Failed to permanently delete product:', error);
      setDeleteError('Failed to delete product. Please try again.');
    } finally {
      setIsSubmitting(false);
    }
  };

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: 'IDR',
      minimumFractionDigits: 0,
    }).format(amount);
  };

  return (
    <div className="flex flex-col h-full">
      <Header title="Products" />

      <div className="flex-1 p-6 overflow-auto">
        {/* Filters and Actions */}
        <Card className="mb-6">
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-4">
            <CardTitle className="text-base font-semibold">Filters</CardTitle>
            {canCreate && viewMode === 'active' && (
              <Button onClick={openCreateDialog} size="sm">
                <Plus className="h-4 w-4 mr-2" />
                Add Product
              </Button>
            )}
          </CardHeader>
          <CardContent>
            <div className="flex flex-col gap-4">
              {/* Tab Toggle */}
              <div className="flex gap-1 p-1 bg-muted rounded-lg w-fit">
                <button
                  onClick={() => setViewMode('active')}
                  className={`px-4 py-2 rounded-md text-sm font-medium transition-colors ${viewMode === 'active'
                    ? 'bg-background text-foreground shadow-sm'
                    : 'text-muted-foreground hover:text-foreground'
                    }`}
                >
                  Active
                </button>
                <button
                  onClick={() => setViewMode('archived')}
                  className={`px-4 py-2 rounded-md text-sm font-medium transition-colors flex items-center gap-2 ${viewMode === 'archived'
                    ? 'bg-background text-foreground shadow-sm'
                    : 'text-muted-foreground hover:text-foreground'
                    }`}
                >
                  <Archive className="h-4 w-4" />
                  Archived
                </button>
              </div>

              {/* Search and Category */}
              <div className="flex flex-col md:flex-row gap-4">
                <div className="relative flex-1">
                  <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                  <Input
                    placeholder="Search products..."
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    className="pl-9 w-full"
                  />
                </div>
                <Select value={selectedCategory} onValueChange={setSelectedCategory}>
                  <SelectTrigger className="w-full md:w-48">
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
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Products table */}
        {isLoading ? (
          <div className="flex items-center justify-center h-64">
            <Loader2 className="h-8 w-8 animate-spin text-primary" />
          </div>
        ) : filteredProducts.length === 0 ? (
          <Card>
            <CardContent className="flex flex-col items-center justify-center py-12">
              <Package className="h-12 w-12 text-muted-foreground mb-4" />
              <p className="text-muted-foreground">
                {viewMode === 'active' ? 'No products found' : 'No archived products'}
              </p>
              {canCreate && viewMode === 'active' && (
                <Button variant="link" onClick={openCreateDialog}>
                  Add your first product
                </Button>
              )}
            </CardContent>
          </Card>
        ) : (
          <>
            {/* Mobile Card View */}
            <div className="md:hidden grid gap-4">
              {filteredProducts.map((product) => {
                const quantity = getProductQuantity(product);
                const minQuantity = getProductMinQuantity(product);
                const isLowStock = quantity <= minQuantity;

                return (
                  <Card key={product.id}>
                    <CardContent className="p-4">
                      <div className="flex items-start gap-4">
                        {product.image_url ? (
                          <div className="relative h-20 w-20 rounded border overflow-hidden flex-shrink-0 bg-muted">
                            <img
                              src={getImageUrl(product.image_url)}
                              alt={product.name}
                              className="h-full w-full object-cover"
                              onError={(e) => {
                                (e.target as HTMLImageElement).style.display = 'none';
                                ((e.target as HTMLImageElement).nextSibling as HTMLElement).style.display = 'flex';
                              }}
                            />
                            <div className="absolute inset-0 hidden items-center justify-center bg-muted">
                              <ImageIcon className="h-8 w-8 text-muted-foreground" />
                            </div>
                          </div>
                        ) : (
                          <div className="h-20 w-20 rounded border flex items-center justify-center bg-muted flex-shrink-0">
                            <Package className="h-8 w-8 text-muted-foreground" />
                          </div>
                        )}
                        <div className="flex-1 min-w-0">
                          <div className="flex items-start justify-between mb-2">
                            <div>
                              <h3 className="font-semibold line-clamp-1">{product.name}</h3>
                              <p className="text-sm text-muted-foreground">{product.sku || '-'}</p>
                            </div>
                            <span
                              className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${product.is_active
                                ? 'bg-green-600 text-white dark:bg-green-600/90 dark:text-white'
                                : 'bg-gray-600 text-white dark:bg-gray-600/90 dark:text-white'
                                }`}
                            >
                              {product.is_active ? 'Active' : 'Inactive'}
                            </span>
                          </div>

                          <div className="grid grid-cols-2 gap-2 text-sm mb-3">
                            <div>
                              <span className="text-muted-foreground">Price: </span>
                              <span className="font-medium">{formatCurrency(getProductPrice(product))}</span>
                            </div>
                            <div>
                              <span className="text-muted-foreground">Stock: </span>
                              <span className={isLowStock ? 'text-destructive font-medium' : 'font-medium'}>
                                {quantity}
                              </span>
                            </div>
                            <div>
                              <span className="text-muted-foreground">Tax: </span>
                              <span className="font-medium">{product.tax_rate ? `${parseFloat(product.tax_rate)}%` : '0%'}</span>
                            </div>
                          </div>

                          {(canEdit || canDelete) && (
                            <div className="flex items-center justify-end gap-2 border-t pt-3 mt-1">
                              {viewMode === 'active' ? (
                                <>
                                  {canEdit && (
                                    <Button
                                      variant="outline"
                                      size="sm"
                                      className="h-8"
                                      onClick={() => openEditDialog(product)}
                                    >
                                      <Pencil className="h-3.5 w-3.5 mr-1" />
                                      Edit
                                    </Button>
                                  )}
                                  {canDelete && (
                                    <Button
                                      variant="outline"
                                      size="sm"
                                      className="h-8 text-destructive hover:text-destructive hover:bg-destructive/10"
                                      onClick={() => {
                                        setDeletingProduct(product);
                                        setDeleteDialogOpen(true);
                                      }}
                                    >
                                      <Trash2 className="h-3.5 w-3.5 mr-1" />
                                      Delete
                                    </Button>
                                  )}
                                </>
                              ) : (
                                <>
                                  {canEdit && (
                                    <Button
                                      variant="outline"
                                      size="sm"
                                      className="h-8"
                                      onClick={() => handleRestore(product)}
                                      disabled={isSubmitting}
                                    >
                                      <RotateCcw className="h-3.5 w-3.5 mr-1" />
                                      Restore
                                    </Button>
                                  )}
                                  {canDelete && (
                                    <Button
                                      variant="outline"
                                      size="sm"
                                      className="h-8 text-destructive hover:text-destructive hover:bg-destructive/10"
                                      onClick={() => {
                                        setDeletingProduct(product);
                                        setPermanentDeleteDialogOpen(true);
                                      }}
                                    >
                                      <Trash2 className="h-3.5 w-3.5 mr-1" />
                                      Delete
                                    </Button>
                                  )}
                                </>
                              )}
                            </div>
                          )}
                        </div>
                      </div>
                    </CardContent>
                  </Card>
                );
              })}
            </div>

            {/* Desktop Table View */}
            <Card className="hidden md:block">
              <CardHeader>
                <CardTitle>
                  {viewMode === 'active' ? 'Products' : 'Archived Products'} ({filteredProducts.length})
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="overflow-x-auto">
                  <table className="w-full">
                    <thead>
                      <tr className="border-b text-left text-sm text-muted-foreground">
                        <th className="pb-3 font-medium">Product</th>
                        <th className="pb-3 font-medium">SKU</th>
                        <th className="pb-3 font-medium">Category</th>
                        <th className="pb-3 font-medium text-right">Price</th>
                        <th className="pb-3 font-medium text-right">Tax Rate</th>
                        <th className="pb-3 font-medium text-right">Stock</th>
                        {viewMode !== 'active' && (
                          <th className="pb-3 font-medium text-center">Status</th>
                        )}
                        {(canEdit || canDelete) && (
                          <th className="pb-3 font-medium text-right">Actions</th>
                        )}
                      </tr>
                    </thead>
                    <tbody>
                      {filteredProducts.map((product) => {
                        const quantity = getProductQuantity(product);
                        const minQuantity = getProductMinQuantity(product);
                        const isLowStock = quantity <= minQuantity;
                        return (
                          <tr key={product.id} className="border-b last:border-0">
                            <td className="py-3">
                              <div className="flex items-center gap-3">
                                {product.image_url ? (
                                  <div className="relative h-10 w-10 rounded border overflow-hidden flex-shrink-0 bg-muted">
                                    <img
                                      src={getImageUrl(product.image_url)}
                                      alt={product.name}
                                      className="h-full w-full object-cover"
                                      onError={(e) => {
                                        (e.target as HTMLImageElement).style.display = 'none';
                                        ((e.target as HTMLImageElement).nextSibling as HTMLElement).style.display = 'flex';
                                      }}
                                    />
                                    <div className="absolute inset-0 hidden items-center justify-center bg-muted">
                                      <ImageIcon className="h-4 w-4 text-muted-foreground" />
                                    </div>
                                  </div>
                                ) : (
                                  <div className="h-10 w-10 rounded border flex items-center justify-center bg-muted flex-shrink-0">
                                    <Package className="h-4 w-4 text-muted-foreground" />
                                  </div>
                                )}
                                <div className="min-w-0">
                                  <p className="font-medium">{product.name}</p>
                                  {product.description && (
                                    <p className="text-xs text-muted-foreground truncate max-w-xs">
                                      {product.description}
                                    </p>
                                  )}
                                </div>
                              </div>
                            </td>
                            <td className="py-3 text-sm text-muted-foreground">
                              {product.sku || '-'}
                            </td>
                            <td className="py-3 text-sm">
                              {product.category_name || '-'}
                            </td>
                            <td className="py-3 text-right font-medium">
                              {formatCurrency(getProductPrice(product))}
                            </td>
                            <td className="py-3 text-right font-medium text-muted-foreground">
                              {product.tax_rate ? `${parseFloat(product.tax_rate)}%` : '0%'}
                            </td>
                            <td className="py-3 text-right">
                              <span className={isLowStock ? 'text-destructive font-medium' : ''}>
                                {quantity}
                              </span>
                            </td>
                            {viewMode !== 'active' && (
                              <td className="py-3 text-center">
                                <span
                                  className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${product.is_active
                                    ? 'bg-green-600 text-white dark:bg-green-600/90 dark:text-white'
                                    : 'bg-gray-600 text-white dark:bg-gray-600/90 dark:text-white'
                                    }`}
                                >
                                  {product.is_active ? 'Active' : 'Inactive'}
                                </span>
                              </td>
                            )}
                            {(canEdit || canDelete) && (
                              <td className="py-3 text-right">
                                <div className="flex items-center justify-end gap-1">
                                  {viewMode === 'active' ? (
                                    <>
                                      {canEdit && (
                                        <Button
                                          variant="ghost"
                                          size="icon"
                                          onClick={() => openEditDialog(product)}
                                        >
                                          <Pencil className="h-4 w-4" />
                                        </Button>
                                      )}
                                      {canDelete && (
                                        <Button
                                          variant="ghost"
                                          size="icon"
                                          onClick={() => {
                                            setDeletingProduct(product);
                                            setDeleteDialogOpen(true);
                                          }}
                                        >
                                          <Trash2 className="h-4 w-4 text-destructive" />
                                        </Button>
                                      )}
                                    </>
                                  ) : (
                                    <>
                                      {canEdit && (
                                        <Button
                                          variant="ghost"
                                          size="sm"
                                          onClick={() => handleRestore(product)}
                                          disabled={isSubmitting}
                                        >
                                          <RotateCcw className="h-4 w-4 mr-1" />
                                          Restore
                                        </Button>
                                      )}
                                      {canDelete && (
                                        <Button
                                          variant="ghost"
                                          size="sm"
                                          className="text-destructive hover:text-destructive"
                                          onClick={() => {
                                            setDeletingProduct(product);
                                            setPermanentDeleteDialogOpen(true);
                                          }}
                                        >
                                          <Trash2 className="h-4 w-4 mr-1" />
                                          Delete
                                        </Button>
                                      )}
                                    </>
                                  )}
                                </div>
                              </td>
                            )}
                          </tr>
                        );
                      })}
                    </tbody>
                  </table>
                </div>
              </CardContent>
            </Card>
          </>
        )}
      </div>

      {/* Create/Edit Dialog */}
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>
              {editingProduct ? 'Edit Product' : 'Add Product'}
            </DialogTitle>
            <DialogDescription>
              {editingProduct
                ? 'Update the product details below.'
                : 'Fill in the details for the new product.'}
            </DialogDescription>
          </DialogHeader>

          <div className="grid gap-4 py-4">
            {formErrors.general && (
              <div className="p-3 rounded-md bg-destructive/10 border border-destructive/20">
                <p className="text-sm text-destructive">{formErrors.general}</p>
              </div>
            )}
            <div className="grid gap-2">
              <Label htmlFor="name">Name *</Label>
              <Input
                id="name"
                value={formData.name}
                onChange={(e) => {
                  setFormData({ ...formData, name: e.target.value });
                  if (formErrors.name) setFormErrors({ ...formErrors, name: undefined });
                }}
                placeholder="Product name"
                className={formErrors.name ? 'border-destructive' : ''}
              />
              {formErrors.name && (
                <p className="text-sm text-destructive">{formErrors.name}</p>
              )}
            </div>

            <div className="grid gap-2">
              <Label htmlFor="description">Description</Label>
              <Input
                id="description"
                value={formData.description}
                onChange={(e) =>
                  setFormData({ ...formData, description: e.target.value })
                }
                placeholder="Product description"
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="grid gap-2">
                <Label htmlFor="sku">SKU</Label>
                <Input
                  id="sku"
                  value={formData.sku}
                  onChange={(e) => {
                    setFormData({ ...formData, sku: e.target.value });
                    if (formErrors.sku) setFormErrors({ ...formErrors, sku: undefined });
                  }}
                  placeholder="SKU-001"
                  className={formErrors.sku ? 'border-destructive' : ''}
                />
                {formErrors.sku && (
                  <p className="text-sm text-destructive">{formErrors.sku}</p>
                )}
              </div>
              <div className="grid gap-2">
                <Label htmlFor="barcode">Barcode</Label>
                <Input
                  id="barcode"
                  value={formData.barcode}
                  onChange={(e) => {
                    setFormData({ ...formData, barcode: e.target.value });
                    if (formErrors.barcode) setFormErrors({ ...formErrors, barcode: undefined });
                  }}
                  placeholder="8901234567890"
                  className={formErrors.barcode ? 'border-destructive' : ''}
                />
                {formErrors.barcode && (
                  <p className="text-sm text-destructive">{formErrors.barcode}</p>
                )}
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="grid gap-2">
                <Label htmlFor="price">Price (IDR) *</Label>
                <Input
                  id="price"
                  type="number"
                  value={formData.price}
                  onChange={(e) => {
                    setFormData({ ...formData, price: e.target.value });
                    if (formErrors.price) setFormErrors({ ...formErrors, price: undefined });
                  }}
                  placeholder="10000"
                  className={formErrors.price ? 'border-destructive' : ''}
                />
                {formErrors.price && (
                  <p className="text-sm text-destructive">{formErrors.price}</p>
                )}
              </div>
              <div className="grid gap-2">
                <Label htmlFor="cost">Cost (IDR)</Label>
                <Input
                  id="cost"
                  type="number"
                  value={formData.cost}
                  onChange={(e) =>
                    setFormData({ ...formData, cost: e.target.value })
                  }
                  placeholder="8000"
                />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="tax_rate">Tax Rate (%)</Label>
                <Input
                  id="tax_rate"
                  type="number"
                  step="0.01"
                  value={formData.tax_rate}
                  onChange={(e) => setFormData({ ...formData, tax_rate: e.target.value })}
                  placeholder="11"
                />
              </div>
            </div>

            {!editingProduct && (
              <div className="grid grid-cols-2 gap-4">
                <div className="grid gap-2">
                  <Label htmlFor="initial_quantity">Initial Stock</Label>
                  <Input
                    id="initial_quantity"
                    type="number"
                    value={formData.initial_quantity}
                    onChange={(e) =>
                      setFormData({ ...formData, initial_quantity: e.target.value })
                    }
                    placeholder="100"
                  />
                </div>
                <div className="grid gap-2">
                  <Label htmlFor="low_stock_threshold">Min Stock Alert</Label>
                  <Input
                    id="low_stock_threshold"
                    type="number"
                    value={formData.low_stock_threshold}
                    onChange={(e) =>
                      setFormData({ ...formData, low_stock_threshold: e.target.value })
                    }
                    placeholder="5"
                  />
                </div>
              </div>
            )}

            <div className="grid gap-2">
              <Label htmlFor="category">Category</Label>
              <Select
                value={formData.category_id || 'none'}
                onValueChange={(value) =>
                  setFormData({ ...formData, category_id: value === 'none' ? '' : value })
                }
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
              <Label>Product Image</Label>
              <ImageUpload
                value={formData.image_url}
                onChange={(url) => setFormData({ ...formData, image_url: url })}
              />
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setDialogOpen(false)}>
              Cancel
            </Button>
            <Button onClick={handleSubmit} disabled={isSubmitting}>
              {isSubmitting ? (
                <>
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  Saving...
                </>
              ) : editingProduct ? (
                'Update Product'
              ) : (
                'Create Product'
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation Dialog (Soft Delete / Archive) */}
      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Archive Product</DialogTitle>
            <DialogDescription>
              Are you sure you want to archive &quot;{deletingProduct?.name}&quot;?
              The product will be moved to the Archived tab and can be restored later.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteDialogOpen(false)}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={handleDelete} disabled={isSubmitting}>
              {isSubmitting ? (
                <>
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  Archiving...
                </>
              ) : (
                'Archive'
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Permanent Delete Confirmation Dialog */}
      <Dialog open={permanentDeleteDialogOpen} onOpenChange={(open) => {
        setPermanentDeleteDialogOpen(open);
        if (!open) {
          setDeleteError(null);
        }
      }}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Permanently Delete Product</DialogTitle>
            <DialogDescription>
              Are you sure you want to permanently delete &quot;{deletingProduct?.name}&quot;?
              This action cannot be undone. All data associated with this product will be lost.
            </DialogDescription>
          </DialogHeader>
          {deleteError && (
            <div className="p-3 rounded-md bg-destructive/10 border border-destructive/20">
              <p className="text-sm text-destructive">{deleteError}</p>
            </div>
          )}
          <DialogFooter>
            <Button variant="outline" onClick={() => {
              setPermanentDeleteDialogOpen(false);
              setDeleteError(null);
            }}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={handlePermanentDelete} disabled={isSubmitting || !!deleteError}>
              {isSubmitting ? (
                <>
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  Deleting...
                </>
              ) : (
                'Delete Permanently'
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

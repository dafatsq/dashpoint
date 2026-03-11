'use client';

import { useEffect, useState } from 'react';
import { Header } from '@/components/layout/header';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
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
  Minus,
  Loader2,
  Boxes,
  AlertTriangle,
  TrendingDown,
  Package,
  Settings2,
  ImageIcon,
} from 'lucide-react';
import api from '@/lib/api';
import { Product, LowStockItem, AdjustmentType } from '@/types';
import { useAuth, PERMISSIONS } from '@/contexts/auth-context';

// Map UI selections to backend adjustment types
const ADJUSTMENT_TYPES = {
  add: [
    { value: 'purchase', label: 'Restock / Purchase' },
    { value: 'adjustment', label: 'Inventory Correction' },
  ],
  remove: [
    { value: 'damage', label: 'Damaged / Expired' },
    { value: 'loss', label: 'Lost / Stolen' },
    { value: 'adjustment', label: 'Inventory Correction' },
  ],
  count: [
    { value: 'count', label: 'Stock Count' },
  ],
};

// Helper to get quantity from product (handles nested inventory object)
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

function getProductPrice(product: Product): number {
  return parseFloat(product.price) || 0;
}

// Helper to get full image URL
function getImageUrl(path: string | null | undefined): string {
  if (!path) return '';
  if (path.startsWith('http://') || path.startsWith('https://')) {
    return path;
  }
  // Derive the backend base URL from the API URL env var
  // e.g. "http://localhost:8080/api/v1" -> "http://localhost:8080"
  const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';
  const baseUrl = apiUrl.replace(/\/api\/v1\/?$/, '');
  return `${baseUrl}${path}`;
}

export default function InventoryPage() {
  const { hasPermission } = useAuth();
  const canModifyStock = hasPermission(PERMISSIONS.INVENTORY_EDIT);

  const [products, setProducts] = useState<Product[]>([]);
  const [lowStockItems, setLowStockItems] = useState<LowStockItem[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [activeTab, setActiveTab] = useState<'all' | 'low-stock'>('all');
  const [lastRefreshed, setLastRefreshed] = useState<Date>(new Date());

  // Adjustment dialog
  const [adjustDialogOpen, setAdjustDialogOpen] = useState(false);
  const [selectedProduct, setSelectedProduct] = useState<Product | null>(null);
  const [adjustmentType, setAdjustmentType] = useState<'add' | 'remove' | 'count'>('add');
  const [adjustmentQuantity, setAdjustmentQuantity] = useState('');
  const [adjustmentTypeValue, setAdjustmentTypeValue] = useState<AdjustmentType>('purchase');
  const [adjustmentNotes, setAdjustmentNotes] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Fetch data
  useEffect(() => {
    const fetchData = async () => {
      try {
        const [productsResult, lowStockResult] = await Promise.all([
          api.getProducts(),
          api.getLowStock(),
        ]);

        if (productsResult.data) setProducts(productsResult.data);
        if (lowStockResult.data) setLowStockItems(lowStockResult.data);
        setLastRefreshed(new Date());
      } catch (error) {
        console.error('Failed to fetch data:', error);
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();

    // Auto-refresh every 10 seconds
    const interval = setInterval(fetchData, 10000);

    return () => clearInterval(interval);
  }, []);

  // Filter products
  const filteredProducts = products.filter((product) =>
    product.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    product.sku?.toLowerCase().includes(searchQuery.toLowerCase())
  );

  // Open adjustment dialog
  const openAdjustDialog = (product: Product) => {
    setSelectedProduct(product);
    setAdjustmentType('add');
    setAdjustmentQuantity('');
    setAdjustmentTypeValue('purchase');
    setAdjustmentNotes('');
    setAdjustDialogOpen(true);
  };

  // Update selected product when products array changes (for real-time updates)
  useEffect(() => {
    if (selectedProduct && adjustDialogOpen) {
      const updated = products.find(p => p.id === selectedProduct.id);
      if (updated) {
        setSelectedProduct(updated);
      }
    }
  }, [products, selectedProduct, adjustDialogOpen]);

  // Handle adjustment
  const handleAdjustment = async () => {
    if (!selectedProduct || !adjustmentQuantity) return;

    const inputQuantity = parseInt(adjustmentQuantity);
    const currentStock = getProductQuantity(selectedProduct);

    console.log('[Adjustment Debug]', {
      adjustmentType,
      adjustmentTypeValue,
      inputQuantity,
      currentStock,
    });

    let finalQuantity: number;

    if (adjustmentTypeValue === 'count') {
      // Stock count: input is the ABSOLUTE final quantity desired
      // User enters what they want the stock to BE, not how much to add/remove
      finalQuantity = inputQuantity;
      console.log('[Adjustment] Count type - setting to absolute quantity:', finalQuantity);

      // Validate that the new quantity is not negative
      if (finalQuantity < 0) {
        alert('Stock quantity cannot be negative');
        return;
      }
    } else if (adjustmentTypeValue === 'adjustment') {
      // Inventory correction: send negative for remove, positive for add
      finalQuantity = adjustmentType === 'remove' ? -inputQuantity : inputQuantity;
      console.log('[Adjustment] Adjustment type - delta:', finalQuantity);

      // Validate that we won't go negative
      if (currentStock + finalQuantity < 0) {
        alert(`Cannot remove ${inputQuantity} items. Only ${currentStock} available.`);
        return;
      }
    } else {
      // damage, loss, purchase: send positive (backend handles negation for damage/loss)
      finalQuantity = inputQuantity;
      console.log('[Adjustment] Other type - quantity:', finalQuantity);

      // For damage/loss, validate we have enough stock
      if ((adjustmentTypeValue === 'damage' || adjustmentTypeValue === 'loss') && finalQuantity > currentStock) {
        alert(`Cannot remove ${inputQuantity} items. Only ${currentStock} available.`);
        return;
      }
    }

    const adjustment = {
      product_id: selectedProduct.id,
      adjustment_type: adjustmentTypeValue,
      quantity: finalQuantity.toString(),
      reason: adjustmentNotes || undefined,
    };

    console.log('[Adjustment] Sending:', adjustment);

    setIsSubmitting(true);

    try {
      const result = await api.adjustInventory(adjustment);
      if (!result.error) {
        // Refresh products
        const productsResult = await api.getProducts();
        if (productsResult.data) setProducts(productsResult.data);

        // Refresh low stock items
        const lowStockResult = await api.getLowStock();
        if (lowStockResult.data) setLowStockItems(lowStockResult.data);

        setAdjustDialogOpen(false);
      } else {
        alert(result.error);
      }
    } catch (error) {
      console.error('Failed to adjust inventory:', error);
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

  // Calculate stats
  const totalProducts = products.length;
  const totalStock = products.reduce((sum, p) => sum + getProductQuantity(p), 0);
  const totalValue = products.reduce((sum, p) => sum + getProductPrice(p) * getProductQuantity(p), 0);
  const lowStockCount = lowStockItems.length;

  return (
    <div className="flex flex-col h-full">
      <Header title="Inventory" />

      <div className="flex-1 p-6 overflow-auto">
        {/* Stats cards */}
        {/* Stats cards */}
        <div className="grid gap-4 grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 mb-6">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium">Total Products</CardTitle>
              <Package className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{totalProducts}</div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium">Total Stock</CardTitle>
              <Boxes className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{totalStock.toLocaleString()}</div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium">Inventory Value</CardTitle>
              <TrendingDown className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-xl sm:text-2xl font-bold truncate" title={formatCurrency(totalValue)}>
                {formatCurrency(totalValue)}
              </div>
            </CardContent>
          </Card>

          <Card className={lowStockCount > 0 ? 'border-yellow-500' : ''}>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium">Low Stock</CardTitle>
              <AlertTriangle className={`h-4 w-4 ${lowStockCount > 0 ? 'text-yellow-500' : 'text-muted-foreground'}`} />
            </CardHeader>
            <CardContent>
              <div className={`text-2xl font-bold ${lowStockCount > 0 ? 'text-yellow-600' : ''}`}>
                {lowStockCount}
              </div>
            </CardContent>
          </Card>
        </div>


        {/* Filters */}
        <Card className="mb-4">
          <CardHeader className="pb-4">
            <CardTitle className="text-base font-semibold">Filters</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex flex-col gap-4">
              <div className="flex flex-col md:flex-row gap-4">
                {/* Search */}
                <div className="relative flex-1 w-full">
                  <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                  <Input
                    placeholder="Search products..."
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    className="pl-9 w-full"
                  />
                </div>

                {/* Tabs */}
                <div className="flex gap-2">
                  <Button
                    variant={activeTab === 'all' ? 'default' : 'outline'}
                    onClick={() => setActiveTab('all')}
                  >
                    All Products
                  </Button>
                  <Button
                    variant={activeTab === 'low-stock' ? 'default' : 'outline'}
                    onClick={() => setActiveTab('low-stock')}
                  >
                    Low Stock ({lowStockCount})
                  </Button>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>

        <div className="flex justify-end mb-4">
          <div className="text-xs text-muted-foreground whitespace-nowrap">
            Last updated: {lastRefreshed.toLocaleTimeString()} • Auto-refreshes every 10s
          </div>
        </div>

        {/* Content */}
        {isLoading ? (
          <div className="flex items-center justify-center h-64">
            <Loader2 className="h-8 w-8 animate-spin text-primary" />
          </div>
        ) : activeTab === 'low-stock' ? (
          // Low stock view
          lowStockItems.length === 0 ? (
            <Card>
              <CardContent className="flex flex-col items-center justify-center py-12">
                <Boxes className="h-12 w-12 text-muted-foreground mb-4" />
                <p className="text-muted-foreground">All products are well-stocked!</p>
              </CardContent>
            </Card>
          ) : (
            <Card>
              <CardHeader>
                <CardTitle>Low Stock Items</CardTitle>
                <CardDescription>
                  Products that are below their minimum stock level
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  {lowStockItems.map((item) => {
                    const product = products.find((p) => p.id === item.id);
                    return (
                      <div
                        key={item.id}
                        className="flex items-center justify-between rounded-lg border p-4"
                      >
                        <div>
                          <p className="font-medium">{item.name}</p>
                          <p className="text-sm text-muted-foreground">
                            SKU: {item.sku || 'N/A'}
                          </p>
                        </div>
                        <div className="flex items-center gap-4">
                          <div className="text-right">
                            <p className="font-bold text-destructive">
                              {parseFloat(item.quantity)} left
                            </p>
                            <p className="text-xs text-muted-foreground">
                              Available: {parseFloat(item.available_quantity)}
                            </p>
                          </div>
                          {product && canModifyStock && (
                            <Button
                              size="sm"
                              variant="outline"
                              onClick={() => openAdjustDialog(product)}
                            >
                              <Settings2 className="h-4 w-4 mr-1" />
                              Adjust
                            </Button>
                          )}
                        </div>
                      </div>
                    );
                  })}
                </div>
              </CardContent>
            </Card>
          )
        ) : (
          // All products view
          <>
            {/* Mobile Card View */}
            <div className="lg:hidden space-y-4">
              {filteredProducts.map((product) => {
                const quantity = getProductQuantity(product);
                const minQuantity = getProductMinQuantity(product);
                const isLowStock = quantity <= minQuantity;
                const isOutOfStock = quantity === 0;

                return (
                  <div key={product.id} className="border rounded-lg p-4 bg-card text-card-foreground shadow-sm">
                    <div className="flex items-start gap-4">
                      {/* Product Image */}
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
                        <div className="flex justify-between items-start mb-2">
                          <div>
                            <h3 className="font-semibold line-clamp-1">{product.name}</h3>
                            <p className="text-sm text-muted-foreground">{product.sku || '-'}</p>
                          </div>
                          <span
                            className={`inline-flex items-center px-2 py-1 rounded-full text-[10px] font-bold uppercase tracking-wider ${isOutOfStock
                              ? 'bg-red-600 text-white'
                              : isLowStock
                                ? 'bg-yellow-600 text-white'
                                : 'bg-green-600 text-white'
                              }`}
                          >
                            {isOutOfStock ? 'Out' : isLowStock ? 'Low' : 'In'}
                          </span>
                        </div>

                        <div className="grid grid-cols-2 gap-2 text-sm mb-3">
                          <div>
                            <span className="text-muted-foreground">Stock: </span>
                            <span className={`font-bold ${isLowStock ? 'text-destructive' : ''}`}>{quantity}</span>
                          </div>
                          <div>
                            <span className="text-muted-foreground">Min: </span>
                            <span>{minQuantity}</span>
                          </div>
                        </div>

                        {canModifyStock && (
                          <div className="flex items-center justify-end gap-2 border-t pt-3 mt-1">
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() => openAdjustDialog(product)}
                              className="h-8"
                            >
                              <Settings2 className="h-3.5 w-3.5 mr-1" />
                              Adjust
                            </Button>
                          </div>
                        )}
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>

            {/* Desktop Table View */}
            <Card className="hidden lg:block">
              <CardHeader>
                <CardTitle>Inventory ({filteredProducts.length})</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="overflow-x-auto">
                  <table className="w-full">
                    <thead>
                      <tr className="border-b text-left text-sm text-muted-foreground">
                        <th className="pb-3 font-medium">Product</th>
                        <th className="pb-3 font-medium hidden lg:table-cell">SKU</th>
                        <th className="pb-3 font-medium text-right">Stock</th>
                        <th className="pb-3 font-medium text-right hidden lg:table-cell">Min</th>
                        <th className="pb-3 font-medium text-center">Status</th>
                        {canModifyStock && (
                          <th className="pb-3 font-medium text-right">Actions</th>
                        )}
                      </tr>
                    </thead>
                    <tbody>
                      {filteredProducts.map((product) => {
                        const quantity = getProductQuantity(product);
                        const minQuantity = getProductMinQuantity(product);
                        const isLowStock = quantity <= minQuantity;
                        const isOutOfStock = quantity === 0;
                        return (
                          <tr key={product.id} className="border-b last:border-0">
                            <td className="py-3">
                              <p className="font-medium">{product.name}</p>
                              <p className="text-xs text-muted-foreground">
                                {product.category_name || 'No category'}
                              </p>
                            </td>
                            <td className="py-3 text-sm text-muted-foreground hidden lg:table-cell">
                              {product.sku || '-'}
                            </td>
                            <td className={`py-3 text-right font-medium ${isLowStock ? 'text-destructive' : ''}`}>
                              {quantity}
                            </td>
                            <td className="py-3 text-right text-muted-foreground hidden lg:table-cell">
                              {minQuantity}
                            </td>
                            <td className="py-3 text-center">
                              <span
                                className={`inline-flex items-center px-2.5 py-1 rounded-full text-xs font-semibold ${isOutOfStock
                                  ? 'bg-red-600 text-white dark:bg-red-600/90 dark:text-white'
                                  : isLowStock
                                    ? 'bg-yellow-600 text-white dark:bg-yellow-600/90 dark:text-white'
                                    : 'bg-green-600 text-white dark:bg-green-600/90 dark:text-white'
                                  }`}
                              >
                                {isOutOfStock ? 'Out of Stock' : isLowStock ? 'Low Stock' : 'In Stock'}
                              </span>
                            </td>
                            {canModifyStock && (
                              <td className="py-3 text-right">
                                <div className="flex items-center justify-end gap-1">
                                  <Button
                                    variant="outline"
                                    size="sm"
                                    onClick={() => openAdjustDialog(product)}
                                  >
                                    <Settings2 className="h-3 w-3 mr-1" />
                                    Adjust
                                  </Button>
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

      {/* Adjustment Dialog */}
      <Dialog open={adjustDialogOpen} onOpenChange={setAdjustDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {adjustmentType === 'add' ? 'Add Stock' : adjustmentType === 'remove' ? 'Remove Stock' : 'Stock Count'}
            </DialogTitle>
            <DialogDescription asChild>
              <div className="space-y-2">
                <div>
                  {selectedProduct?.name} - Current stock: {selectedProduct ? getProductQuantity(selectedProduct) : 0}
                </div>
                {adjustmentType === 'count' ? (
                  <div className="flex items-center gap-2 text-blue-600 dark:text-blue-400 text-xs">
                    <svg className="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                    <span>Enter the exact quantity counted. This will update inventory to match your physical count.</span>
                  </div>
                ) : (
                  <div className="flex items-center gap-2 text-yellow-600 dark:text-yellow-500 text-xs">
                    <AlertTriangle className="h-3 w-3" />
                    <span>Stock updates in real-time. Current value may change if others adjust inventory.</span>
                  </div>
                )}
              </div>
            </DialogDescription>
          </DialogHeader>

          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label htmlFor="actionType">Action</Label>
              <Select
                value={adjustmentType}
                onValueChange={(value: 'add' | 'remove' | 'count') => {
                  setAdjustmentType(value);
                  setAdjustmentTypeValue(
                    value === 'add' ? 'purchase' : value === 'remove' ? 'damage' : 'count'
                  );
                }}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="add">Add Stock</SelectItem>
                  <SelectItem value="remove">Remove Stock</SelectItem>
                  <SelectItem value="count">Stock Count</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="grid gap-2">
              <Label htmlFor="quantity">Quantity</Label>
              <Input
                id="quantity"
                type="number"
                min="1"
                value={adjustmentQuantity}
                onChange={(e) => setAdjustmentQuantity(e.target.value)}
                placeholder="Enter quantity"
              />
            </div>

            {adjustmentType !== 'count' && (
              <div className="grid gap-2">
                <Label htmlFor="adjustmentType">Adjustment Type</Label>
                <Select
                  value={adjustmentTypeValue}
                  onValueChange={(value: AdjustmentType) => setAdjustmentTypeValue(value)}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {ADJUSTMENT_TYPES[adjustmentType].map((type) => (
                      <SelectItem key={type.value} value={type.value}>
                        {type.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            )}

            <div className="grid gap-2">
              <Label htmlFor="notes">Notes (optional)</Label>
              <Input
                id="notes"
                value={adjustmentNotes}
                onChange={(e) => setAdjustmentNotes(e.target.value)}
                placeholder="Additional notes..."
              />
            </div>

            {adjustmentQuantity && (
              <div className="rounded-lg bg-muted p-3">
                <p className="text-sm">
                  {adjustmentTypeValue === 'count' ? (
                    <>
                      Setting stock to:{' '}
                      <span className="font-bold">{parseInt(adjustmentQuantity || '0')}</span>
                      <span className="text-muted-foreground ml-2">
                        (Current: {selectedProduct ? getProductQuantity(selectedProduct) : 0})
                      </span>
                    </>
                  ) : (
                    <>
                      New stock level:{' '}
                      <span className="font-bold">
                        {selectedProduct
                          ? getProductQuantity(selectedProduct) +
                          (adjustmentType === 'add' ? 1 : -1) * parseInt(adjustmentQuantity || '0')
                          : 0}
                      </span>
                    </>
                  )}
                </p>
              </div>
            )}
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setAdjustDialogOpen(false)}>
              Cancel
            </Button>
            <Button
              onClick={handleAdjustment}
              disabled={isSubmitting || !adjustmentQuantity}
              variant={adjustmentType === 'remove' ? 'destructive' : 'default'}
            >
              {isSubmitting ? (
                <>
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  Saving...
                </>
              ) : adjustmentType === 'add' ? (
                'Add Stock'
              ) : adjustmentType === 'remove' ? (
                'Remove Stock'
              ) : (
                'Update Count'
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

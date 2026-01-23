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
} from 'lucide-react';
import api from '@/lib/api';
import { Product, LowStockItem, AdjustmentType } from '@/types';
import { useAuth, PERMISSIONS } from '@/contexts/auth-context';

// Map UI selections to backend adjustment types
const ADJUSTMENT_TYPES = {
  add: [
    { value: 'purchase', label: 'Restock / Purchase' },
    { value: 'adjustment', label: 'Inventory Correction' },
    { value: 'count', label: 'Stock Count' },
  ],
  remove: [
    { value: 'damage', label: 'Damaged / Expired' },
    { value: 'loss', label: 'Lost / Stolen' },
    { value: 'adjustment', label: 'Inventory Correction' },
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

export default function InventoryPage() {
  const { hasPermission } = useAuth();
  const canAdjust = hasPermission(PERMISSIONS.INVENTORY_ADJUST);
  const canReceive = hasPermission(PERMISSIONS.INVENTORY_RECEIVE);
  const canModifyStock = canAdjust || canReceive;

  const [products, setProducts] = useState<Product[]>([]);
  const [lowStockItems, setLowStockItems] = useState<LowStockItem[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [activeTab, setActiveTab] = useState<'all' | 'low-stock'>('all');
  const [lastRefreshed, setLastRefreshed] = useState<Date>(new Date());

  // Adjustment dialog
  const [adjustDialogOpen, setAdjustDialogOpen] = useState(false);
  const [selectedProduct, setSelectedProduct] = useState<Product | null>(null);
  const [adjustmentType, setAdjustmentType] = useState<'add' | 'remove'>('add');
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
  const openAdjustDialog = (product: Product, type: 'add' | 'remove') => {
    setSelectedProduct(product);
    setAdjustmentType(type);
    setAdjustmentQuantity('');
    setAdjustmentTypeValue(type === 'add' ? 'purchase' : 'damage');
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

    setIsSubmitting(true);

    const quantity = parseInt(adjustmentQuantity);
    
    // Backend expects positive quantity for purchase/adjustment/count
    // and will auto-negate for damage/loss
    const adjustment = {
      product_id: selectedProduct.id,
      adjustment_type: adjustmentTypeValue,
      quantity: quantity.toString(),
      reason: adjustmentNotes || undefined,
    };

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
        <div className="grid gap-4 md:grid-cols-4 mb-6">
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
              <div className="text-2xl font-bold">{formatCurrency(totalValue)}</div>
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

        {/* Tabs */}
        <div className="flex gap-2 mb-4">
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

        {/* Search */}
        <div className="flex items-center justify-between mb-4">
          <div className="relative flex-1 max-w-sm">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Search products..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-9"
            />
          </div>
          <div className="text-xs text-muted-foreground">
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
                                onClick={() => openAdjustDialog(product, 'add')}
                              >
                                <Plus className="h-4 w-4 mr-1" />
                                Restock
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
          <Card>
            <CardHeader>
              <CardTitle>Inventory ({filteredProducts.length})</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead>
                    <tr className="border-b text-left text-sm text-muted-foreground">
                      <th className="pb-3 font-medium">Product</th>
                      <th className="pb-3 font-medium">SKU</th>
                      <th className="pb-3 font-medium text-right">Current Stock</th>
                      <th className="pb-3 font-medium text-right">Min Stock</th>
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
                          <td className="py-3 text-sm text-muted-foreground">
                            {product.sku || '-'}
                          </td>
                          <td className={`py-3 text-right font-medium ${isLowStock ? 'text-destructive' : ''}`}>
                            {quantity}
                          </td>
                          <td className="py-3 text-right text-muted-foreground">
                            {minQuantity}
                          </td>
                          <td className="py-3 text-center">
                            <span
                              className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${
                                isOutOfStock
                                  ? 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
                                  : isLowStock
                                  ? 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400'
                                  : 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
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
                                  onClick={() => openAdjustDialog(product, 'add')}
                                >
                                  <Plus className="h-3 w-3" />
                                </Button>
                                <Button
                                  variant="outline"
                                  size="sm"
                                  onClick={() => openAdjustDialog(product, 'remove')}
                                  disabled={quantity === 0}
                                >
                                  <Minus className="h-3 w-3" />
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
        )}
      </div>

      {/* Adjustment Dialog */}
      <Dialog open={adjustDialogOpen} onOpenChange={setAdjustDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {adjustmentType === 'add' ? 'Add Stock' : 'Remove Stock'}
            </DialogTitle>
            <DialogDescription asChild>
              <div className="space-y-2">
                <div>
                  {selectedProduct?.name} - Current stock: {selectedProduct ? getProductQuantity(selectedProduct) : 0}
                </div>
                <div className="flex items-center gap-2 text-yellow-600 dark:text-yellow-500 text-xs">
                  <AlertTriangle className="h-3 w-3" />
                  <span>Stock updates in real-time. Current value may change if others adjust inventory.</span>
                </div>
              </div>
            </DialogDescription>
          </DialogHeader>

          <div className="grid gap-4 py-4">
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
                  New stock level:{' '}
                  <span className="font-bold">
                    {selectedProduct
                      ? getProductQuantity(selectedProduct) +
                        (adjustmentType === 'add' ? 1 : -1) * parseInt(adjustmentQuantity || '0')
                      : 0}
                  </span>
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
              ) : (
                'Remove Stock'
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

'use client';

import { useEffect, useState, useCallback, useRef } from 'react';
import { Header } from '@/components/layout/header';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Separator } from '@/components/ui/separator';
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
  Trash2,
  CreditCard,
  Banknote,
  QrCode,
  Building2,
  Loader2,
  ShoppingCart,
  AlertCircle,
  Clock,
  CheckCircle,
} from 'lucide-react';
import api from '@/lib/api';
import { Product, CartItem, Category, Shift, PaymentMethod, CreateSaleRequest } from '@/types';
import { useAuth, PERMISSIONS } from '@/contexts/auth-context';

// Helper functions to parse Product string values
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

export default function POSPage() {
  const { hasPermission } = useAuth();
  const canApplyDiscount = hasPermission(PERMISSIONS.SALES_DISCOUNT);

  // Product state
  const [products, setProducts] = useState<Product[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [selectedCategory, setSelectedCategory] = useState<string>('all');
  const [searchQuery, setSearchQuery] = useState('');
  const [isLoadingProducts, setIsLoadingProducts] = useState(true);

  // Cart state
  const [cartItems, setCartItems] = useState<CartItem[]>([]);
  const [discount, setDiscount] = useState(0);

  // Shift state
  const [currentShift, setCurrentShift] = useState<Shift | null>(null);
  const [shiftDialogOpen, setShiftDialogOpen] = useState(false);
  const [startingCash, setStartingCash] = useState('');

  // Checkout state
  const [checkoutDialogOpen, setCheckoutDialogOpen] = useState(false);
  const [paymentMethod, setPaymentMethod] = useState<PaymentMethod>('cash');
  const [amountPaid, setAmountPaid] = useState('');
  const [isProcessing, setIsProcessing] = useState(false);
  const [saleComplete, setSaleComplete] = useState(false);
  const [lastInvoice, setLastInvoice] = useState('');
  const [lastChange, setLastChange] = useState(0);

  const searchInputRef = useRef<HTMLInputElement>(null);

  // Calculate totals
  const subtotal = cartItems.reduce(
    (sum, item) => sum + getProductPrice(item.product) * item.quantity,
    0
  );
  const discountAmount = discount;
  const total = subtotal - discountAmount;
  const change = parseFloat(amountPaid || '0') - total;

  // Fetch initial data
  useEffect(() => {
    const fetchData = async () => {
      try {
        const [productsResult, categoriesResult, shiftResult] = await Promise.all([
          api.getProducts({ active: true }),
          api.getCategories(),
          api.getCurrentShift(),
        ]);

        if (productsResult.data) setProducts(productsResult.data);
        if (categoriesResult.data) setCategories(categoriesResult.data);
        if (shiftResult.data && !shiftResult.error) {
          setCurrentShift(shiftResult.data);
        }
      } catch (error) {
        console.error('Failed to fetch data:', error);
      } finally {
        setIsLoadingProducts(false);
      }
    };

    fetchData();
  }, []);

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

  // Cart functions
  const addToCart = useCallback((product: Product) => {
    setCartItems((items) => {
      const existingItem = items.find((item) => item.product.id === product.id);
      if (existingItem) {
        return items.map((item) =>
          item.product.id === product.id
            ? { ...item, quantity: item.quantity + 1 }
            : item
        );
      }
      return [...items, { product, quantity: 1 }];
    });
  }, []);

  const updateQuantity = useCallback((productId: string, delta: number) => {
    setCartItems((items) =>
      items
        .map((item) =>
          item.product.id === productId
            ? { ...item, quantity: Math.max(0, item.quantity + delta) }
            : item
        )
        .filter((item) => item.quantity > 0)
    );
  }, []);

  const removeFromCart = useCallback((productId: string) => {
    setCartItems((items) => items.filter((item) => item.product.id !== productId));
  }, []);

  const clearCart = useCallback(() => {
    setCartItems([]);
    setDiscount(0);
  }, []);

  // Shift functions
  const handleStartShift = async () => {
    if (!startingCash) return;

    try {
      const result = await api.startShift(parseFloat(startingCash));
      if (result.data) {
        setCurrentShift(result.data);
        setShiftDialogOpen(false);
        setStartingCash('');
      }
    } catch (error) {
      console.error('Failed to start shift:', error);
    }
  };

  // Checkout functions
  const handleCheckout = async () => {
    if (!currentShift || cartItems.length === 0) return;

    setIsProcessing(true);

    // Build request matching backend format
    const saleRequest: CreateSaleRequest = {
      items: cartItems.map((item) => ({
        product_id: item.product.id,
        quantity: item.quantity.toString(),
        unit_price: item.product.price, // Already a string
      })),
      payments: [
        {
          payment_method: paymentMethod,
          amount: total.toString(),
          amount_tendered: paymentMethod === 'cash' ? amountPaid : undefined,
          change_given: paymentMethod === 'cash' && change > 0 ? change.toString() : undefined,
        },
      ],
      discount_value: discountAmount > 0 ? discountAmount.toString() : undefined,
      discount_type: discountAmount > 0 ? 'fixed' : undefined,
    };

    try {
      const result = await api.createSale(saleRequest);
      if (result.data) {
        setLastInvoice(result.data.invoice_no);
        // Store the change before clearing cart (change becomes 0 after clear)
        setLastChange(change > 0 ? change : 0);
        setSaleComplete(true);
        clearCart();
        setAmountPaid('');
      } else if (result.error) {
        alert(result.error);
      }
    } catch (error) {
      console.error('Failed to process sale:', error);
      alert('Failed to process sale');
    } finally {
      setIsProcessing(false);
    }
  };

  const closeCheckoutDialog = () => {
    setCheckoutDialogOpen(false);
    setSaleComplete(false);
    setAmountPaid('');
    setLastChange(0);
  };

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: 'IDR',
      minimumFractionDigits: 0,
    }).format(amount);
  };

  // Quick amount buttons for cash payment
  const quickAmounts = [10000, 20000, 50000, 100000];

  return (
    <div className="flex flex-col h-full">
      <Header title="Point of Sale" />

      {/* Shift status bar */}
      {!currentShift ? (
        <div className="bg-yellow-50 dark:bg-yellow-900/20 border-b border-yellow-200 dark:border-yellow-800 px-6 py-3 flex items-center justify-between">
          <div className="flex items-center gap-2 text-yellow-700 dark:text-yellow-400">
            <AlertCircle className="h-4 w-4" />
            <span className="text-sm">No active shift. Start a shift to begin selling.</span>
          </div>
          <Button size="sm" onClick={() => setShiftDialogOpen(true)}>
            <Clock className="h-4 w-4 mr-2" />
            Start Shift
          </Button>
        </div>
      ) : (
        <div className="bg-green-50 dark:bg-green-900/20 border-b border-green-200 dark:border-green-800 px-6 py-3 flex items-center justify-between">
          <div className="flex items-center gap-2 text-green-700 dark:text-green-400">
            <CheckCircle className="h-4 w-4" />
            <span className="text-sm">
              Shift active | Started: {new Date(currentShift.started_at).toLocaleTimeString()} |
              Transactions: {currentShift.total_transactions}
            </span>
          </div>
        </div>
      )}

      <div className="flex-1 flex overflow-hidden">
        {/* Left side - Product grid */}
        <div className="flex-1 flex flex-col p-4 overflow-hidden">
          {/* Search and filter */}
          <div className="flex gap-4 mb-4">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <Input
                ref={searchInputRef}
                placeholder="Search products or scan barcode..."
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
          </div>

          {/* Product grid */}
          <div className="flex-1 overflow-y-auto">
            {isLoadingProducts ? (
              <div className="flex items-center justify-center h-full">
                <Loader2 className="h-8 w-8 animate-spin text-primary" />
              </div>
            ) : filteredProducts.length === 0 ? (
              <div className="flex flex-col items-center justify-center h-full text-muted-foreground">
                <ShoppingCart className="h-12 w-12 mb-2" />
                <p>No products found</p>
              </div>
            ) : (
              <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-3">
                {filteredProducts.map((product) => {
                  const quantity = getProductQuantity(product);
                  const minQuantity = getProductMinQuantity(product);
                  const price = getProductPrice(product);
                  const isLowStock = quantity <= minQuantity;
                  return (
                    <button
                      key={product.id}
                      onClick={() => currentShift && addToCart(product)}
                      disabled={!currentShift || quantity <= 0}
                      className="bg-card border rounded-lg p-3 text-left hover:border-primary hover:shadow-md transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                      <div className="aspect-square bg-muted rounded-md mb-2 flex items-center justify-center">
                        <ShoppingCart className="h-8 w-8 text-muted-foreground" />
                      </div>
                      <p className="font-medium text-sm truncate">{product.name}</p>
                      <p className="text-xs text-muted-foreground">{product.category_name}</p>
                      <div className="flex items-center justify-between mt-2">
                        <span className="font-bold text-primary">
                          {formatCurrency(price)}
                        </span>
                        <span className={`text-xs ${isLowStock ? 'text-destructive' : 'text-muted-foreground'}`}>
                          Stock: {quantity}
                        </span>
                      </div>
                    </button>
                  );
                })}
              </div>
            )}
          </div>
        </div>

        {/* Right side - Cart */}
        <Card className="w-96 rounded-none border-l border-t-0 border-b-0 border-r-0 flex flex-col">
          <CardHeader className="py-4">
            <CardTitle className="flex items-center justify-between">
              <span>Current Order</span>
              {cartItems.length > 0 && (
                <Button variant="ghost" size="sm" onClick={clearCart}>
                  Clear
                </Button>
              )}
            </CardTitle>
          </CardHeader>

          <CardContent className="flex-1 flex flex-col p-4 pt-0 overflow-hidden">
            {/* Cart items */}
            <div className="flex-1 overflow-y-auto space-y-2">
              {cartItems.length === 0 ? (
                <div className="flex flex-col items-center justify-center h-full text-muted-foreground">
                  <ShoppingCart className="h-12 w-12 mb-2" />
                  <p className="text-sm">Cart is empty</p>
                </div>
              ) : (
                cartItems.map((item) => (
                  <div
                    key={item.product.id}
                    className="flex items-center gap-3 p-2 rounded-lg border bg-card"
                  >
                    <div className="flex-1 min-w-0">
                      <p className="font-medium text-sm truncate">{item.product.name}</p>
                      <p className="text-sm text-muted-foreground">
                        {formatCurrency(getProductPrice(item.product))}
                      </p>
                    </div>
                    <div className="flex items-center gap-1">
                      <Button
                        variant="outline"
                        size="icon"
                        className="h-7 w-7"
                        onClick={() => updateQuantity(item.product.id, -1)}
                      >
                        <Minus className="h-3 w-3" />
                      </Button>
                      <span className="w-8 text-center text-sm font-medium">
                        {item.quantity}
                      </span>
                      <Button
                        variant="outline"
                        size="icon"
                        className="h-7 w-7"
                        onClick={() => updateQuantity(item.product.id, 1)}
                      >
                        <Plus className="h-3 w-3" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-7 w-7 text-destructive"
                        onClick={() => removeFromCart(item.product.id)}
                      >
                        <Trash2 className="h-3 w-3" />
                      </Button>
                    </div>
                  </div>
                ))
              )}
            </div>

            <Separator className="my-4" />

            {/* Totals */}
            <div className="space-y-2 text-sm">
              <div className="flex justify-between">
                <span className="text-muted-foreground">Subtotal</span>
                <span>{formatCurrency(subtotal)}</span>
              </div>
              {canApplyDiscount && (
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Discount</span>
                  <Input
                    type="number"
                    value={discount || ''}
                    onChange={(e) => setDiscount(parseFloat(e.target.value) || 0)}
                    className="w-28 h-7 text-right"
                    placeholder="0"
                  />
                </div>
              )}
              <Separator />
              <div className="flex justify-between text-lg font-bold">
                <span>Total</span>
                <span className="text-primary">{formatCurrency(total)}</span>
              </div>
            </div>

            {/* Checkout button */}
            <Button
              size="lg"
              className="w-full mt-4"
              disabled={!currentShift || cartItems.length === 0}
              onClick={() => {
                setAmountPaid(total.toString());
                setCheckoutDialogOpen(true);
              }}
            >
              <CreditCard className="h-5 w-5 mr-2" />
              Checkout
            </Button>
          </CardContent>
        </Card>
      </div>

      {/* Start Shift Dialog */}
      <Dialog open={shiftDialogOpen} onOpenChange={setShiftDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Start New Shift</DialogTitle>
            <DialogDescription>
              Enter the starting cash amount in your drawer to begin your shift.
            </DialogDescription>
          </DialogHeader>
          <div className="py-4">
            <label className="text-sm font-medium">Starting Cash (IDR)</label>
            <Input
              type="number"
              value={startingCash}
              onChange={(e) => setStartingCash(e.target.value)}
              placeholder="Enter amount..."
              className="mt-2"
            />
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShiftDialogOpen(false)}>
              Cancel
            </Button>
            <Button onClick={handleStartShift} disabled={!startingCash}>
              Start Shift
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Checkout Dialog */}
      <Dialog open={checkoutDialogOpen} onOpenChange={closeCheckoutDialog}>
        <DialogContent className="sm:max-w-md">
          {saleComplete ? (
            <>
              <DialogHeader>
                <DialogTitle className="text-center">
                  <CheckCircle className="h-12 w-12 text-green-500 mx-auto mb-2" />
                  Sale Complete!
                </DialogTitle>
                <DialogDescription className="text-center">
                  Invoice: {lastInvoice}
                </DialogDescription>
              </DialogHeader>
              <div className="text-center py-4">
                <p className="text-2xl font-bold text-primary">
                  Change: {formatCurrency(lastChange)}
                </p>
              </div>
              <DialogFooter>
                <Button className="w-full" onClick={closeCheckoutDialog}>
                  New Sale
                </Button>
              </DialogFooter>
            </>
          ) : (
            <>
              <DialogHeader>
                <DialogTitle>Checkout</DialogTitle>
                <DialogDescription>
                  Total: <span className="font-bold text-primary">{formatCurrency(total)}</span>
                </DialogDescription>
              </DialogHeader>

              <div className="space-y-4 py-4">
                {/* Payment method */}
                <div>
                  <label className="text-sm font-medium mb-2 block">Payment Method</label>
                  <div className="grid grid-cols-2 gap-2">
                    <Button
                      variant={paymentMethod === 'cash' ? 'default' : 'outline'}
                      onClick={() => setPaymentMethod('cash')}
                      className="flex items-center gap-2"
                    >
                      <Banknote className="h-4 w-4" />
                      Cash
                    </Button>
                    <Button
                      variant={paymentMethod === 'card' ? 'default' : 'outline'}
                      onClick={() => setPaymentMethod('card')}
                      className="flex items-center gap-2"
                    >
                      <CreditCard className="h-4 w-4" />
                      Card
                    </Button>
                    <Button
                      variant={paymentMethod === 'qris' ? 'default' : 'outline'}
                      onClick={() => setPaymentMethod('qris')}
                      className="flex items-center gap-2"
                    >
                      <QrCode className="h-4 w-4" />
                      QRIS
                    </Button>
                    <Button
                      variant={paymentMethod === 'transfer' ? 'default' : 'outline'}
                      onClick={() => setPaymentMethod('transfer')}
                      className="flex items-center gap-2"
                    >
                      <Building2 className="h-4 w-4" />
                      Transfer
                    </Button>
                  </div>
                </div>

                {/* Amount paid (for cash) */}
                {paymentMethod === 'cash' && (
                  <div>
                    <label className="text-sm font-medium mb-2 block">Amount Received</label>
                    <Input
                      type="number"
                      value={amountPaid}
                      onChange={(e) => setAmountPaid(e.target.value)}
                      placeholder="Enter amount..."
                    />
                    <div className="flex gap-2 mt-2">
                      {quickAmounts.map((amount) => (
                        <Button
                          key={amount}
                          variant="outline"
                          size="sm"
                          onClick={() => setAmountPaid(amount.toString())}
                        >
                          {amount / 1000}K
                        </Button>
                      ))}
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => setAmountPaid(total.toString())}
                      >
                        Exact
                      </Button>
                    </div>
                    {parseFloat(amountPaid || '0') >= total && (
                      <p className="mt-2 text-sm">
                        Change: <span className="font-bold text-primary">{formatCurrency(change)}</span>
                      </p>
                    )}
                  </div>
                )}
              </div>

              <DialogFooter>
                <Button variant="outline" onClick={closeCheckoutDialog}>
                  Cancel
                </Button>
                <Button
                  onClick={handleCheckout}
                  disabled={
                    isProcessing ||
                    (paymentMethod === 'cash' && parseFloat(amountPaid || '0') < total)
                  }
                >
                  {isProcessing ? (
                    <>
                      <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                      Processing...
                    </>
                  ) : (
                    'Complete Sale'
                  )}
                </Button>
              </DialogFooter>
            </>
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}

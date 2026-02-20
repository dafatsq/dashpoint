'use client';

import { useEffect, useState, useCallback, useRef } from 'react';
import { useRouter } from 'next/navigation';
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
  ArrowDownCircle,
  ArrowUpCircle,
} from 'lucide-react';
import api from '@/lib/api';
import { Product, CartItem, Category, Shift, PaymentMethod, CreateSaleRequest, CashDrawerOperation, CashDrawerOperationsResponse } from '@/types';
import { useAuth, PERMISSIONS } from '@/contexts/auth-context';
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@/components/ui/sheet';

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

// Helper to format currency
const formatCurrency = (amount: number) => {
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
  }).format(amount);
};

// Extracted Cart View Component
interface CartViewProps {
  cartItems: CartItem[];
  onClear: () => void;
  onUpdateQuantity: (id: string, delta: number) => void;
  onRemove: (id: string) => void;
  user: any;
  subtotal: number;
  totalTax: number;
  discount: number;
  discountAmount: number;
  setDiscount: (val: number) => void;
  total: number;
  canApplyDiscount: boolean;
  currentShift: Shift | null;
  onCheckout: () => void;
}

function CartView({
  cartItems,
  onClear,
  onUpdateQuantity,
  onRemove,
  user,
  subtotal,
  totalTax,
  discount,
  discountAmount,
  setDiscount,
  total,
  canApplyDiscount,
  currentShift,
  onCheckout,
}: CartViewProps) {
  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center justify-between p-4 border-b">
        <h2 className="font-semibold text-lg">Current Order</h2>
        {cartItems.length > 0 && (
          <Button variant="ghost" size="sm" onClick={onClear}>
            Clear
          </Button>
        )}
      </div>

      <div className="flex-1 flex flex-col p-4 overflow-hidden">
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
                    onClick={() => onUpdateQuantity(item.product.id, -1)}
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
                    onClick={() => onUpdateQuantity(item.product.id, 1)}
                  >
                    <Plus className="h-3 w-3" />
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-7 w-7 text-destructive"
                    onClick={() => onRemove(item.product.id)}
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
          <div className="flex justify-between items-center pb-2 border-b">
            <span className="text-muted-foreground">Cashier</span>
            <span className="font-medium">{user?.name || 'Unknown'}</span>
          </div>
          <div className="flex justify-between">
            <span className="text-muted-foreground">Subtotal</span>
            <span>{formatCurrency(subtotal)}</span>
          </div>
          {totalTax > 0 && (
            <div className="flex justify-between">
              <span className="text-muted-foreground">Tax</span>
              <span>{formatCurrency(totalTax)}</span>
            </div>
          )}
          {canApplyDiscount && (
            <div className="space-y-1">
              <div className="flex justify-between items-center">
                <span className="text-muted-foreground">Discount (%)</span>
                <div className="flex items-center gap-1">
                  <Input
                    type="number"
                    value={discount || ''}
                    onChange={(e) => {
                      const val = parseFloat(e.target.value) || 0;
                      setDiscount(Math.min(100, Math.max(0, val)));
                    }}
                    className="w-16 h-7 text-right"
                    placeholder="0"
                    max="100"
                    min="0"
                  />
                  <span className="text-sm font-medium w-4">%</span>
                </div>
              </div>
              {discountAmount > 0 && (
                <div className="flex justify-between text-destructive">
                  <span className="text-xs">Discount Amount</span>
                  <span>-{formatCurrency(discountAmount)}</span>
                </div>
              )}
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
          onClick={onCheckout}
        >
          <CreditCard className="h-5 w-5 mr-2" />
          Checkout
        </Button>
      </div>
    </div>
  );
}

export default function POSPage() {
  const router = useRouter();
  const { hasPermission, user } = useAuth();
  const canApplyDiscount = hasPermission(PERMISSIONS.SALES_CREATE);

  // Redirect if no permission
  useEffect(() => {
    if (!hasPermission(PERMISSIONS.POS_VIEW)) {
      router.replace('/');
    }
  }, [hasPermission, router]);

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

  // Shift details and end shift state
  const [shiftDetailsOpen, setShiftDetailsOpen] = useState(false);
  const [endingCash, setEndingCash] = useState('');
  const [closingNotes, setClosingNotes] = useState('');

  // Cash drawer operations state
  const [cashDrawerDialogOpen, setCashDrawerDialogOpen] = useState(false);
  const [cashDrawerOpType, setCashDrawerOpType] = useState<'pay_in' | 'pay_out'>('pay_in');
  const [cashDrawerAmount, setCashDrawerAmount] = useState('');
  const [cashDrawerReason, setCashDrawerReason] = useState('');
  const [cashDrawerOps, setCashDrawerOps] = useState<CashDrawerOperation[]>([]);
  const [cashDrawerTotals, setCashDrawerTotals] = useState<{ pay_in_total: string; pay_out_total: string }>({ pay_in_total: '0', pay_out_total: '0' });
  const [isSubmittingOp, setIsSubmittingOp] = useState(false);

  // Blind close state
  const [blindClose, setBlindClose] = useState(true);
  const [shiftClosed, setShiftClosed] = useState(false);
  const [closedShiftData, setClosedShiftData] = useState<Shift | null>(null);

  // Checkout state
  const [checkoutDialogOpen, setCheckoutDialogOpen] = useState(false);
  const [paymentMethod, setPaymentMethod] = useState<PaymentMethod>('cash');
  const [amountPaid, setAmountPaid] = useState('');
  const [isProcessing, setIsProcessing] = useState(false);
  const [saleComplete, setSaleComplete] = useState(false);
  const [lastInvoice, setLastInvoice] = useState('');
  const [lastChange, setLastChange] = useState(0);

  // Error dialog state
  const [errorDialogOpen, setErrorDialogOpen] = useState(false);
  const [errorTitle, setErrorTitle] = useState('');
  const [errorMessage, setErrorMessage] = useState('');

  const showError = (title: string, message: string) => {
    setErrorTitle(title);
    setErrorMessage(message);
    setErrorDialogOpen(true);
  };

  const searchInputRef = useRef<HTMLInputElement>(null);

  // Calculate totals
  const subtotal = cartItems.reduce(
    (sum, item) => sum + getProductPrice(item.product) * item.quantity,
    0
  );

  const totalTax = cartItems.reduce((sum, item) => {
    const itemSubtotal = getProductPrice(item.product) * item.quantity;
    const taxRate = item.product.tax_rate ? parseFloat(item.product.tax_rate.toString()) : 0;
    return sum + (itemSubtotal * taxRate) / 100;
  }, 0);

  const discountAmount = (subtotal * discount) / 100;
  const total = subtotal + totalTax - discountAmount;
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

  // Poll for shift status AND products every 30 seconds
  useEffect(() => {
    const pollData = async () => {
      await Promise.all([
        refreshShift(),
        refreshProducts() // Keep inventory in sync
      ]);
    };

    const interval = setInterval(pollData, 30000);
    return () => clearInterval(interval);
  }, [currentShift]);

  const refreshProducts = async () => {
    try {
      const result = await api.getProducts({ active: true });
      if (result.data) {
        setProducts(result.data);
      }
    } catch (error) {
      console.error('Failed to refresh products:', error);
    }
  };

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

  const refreshShift = async () => {
    try {
      const result = await api.getCurrentShift();

      // Handle case where shift was closed remotely
      if (!result.data && currentShift) {
        setCurrentShift(null);
        showError('Shift Closed', 'The shift has been closed by another user. Sales are disabled.');
        return;
      }

      if (result.data) {
        // Update local state
        setCurrentShift(result.data);
      }
    } catch (error) {
      console.error('Failed to refresh shift:', error);
    }
  };

  const handleEndShift = async () => {
    if (!endingCash) return;

    // Validate input
    const cashAmount = parseFloat(endingCash);
    if (isNaN(cashAmount) || cashAmount < 0) {
      showError('Invalid Amount', 'Please enter a valid positive cash amount');
      return;
    }

    setIsProcessing(true);
    try {
      const result = await api.closeShift(endingCash, closingNotes);
      if (result.data) {
        // Show reconciliation result before closing dialog
        setClosedShiftData(result.data);
        setShiftClosed(true);
        setCurrentShift(null);
      } else if (result.error) {
        showError('End Shift Failed', result.error);
      }
    } catch (error) {
      console.error('Failed to end shift:', error);
      showError('End Shift Failed', 'An unexpected error occurred while ending the shift.');
    } finally {
      setIsProcessing(false);
    }
  };

  const openShiftDetails = async () => {
    await refreshShift(); // Get latest data before opening
    // Load cash drawer operations
    if (currentShift) {
      const opsResult = await api.getShiftOperations(currentShift.id);
      if (opsResult.data) {
        setCashDrawerOps(opsResult.data.operations || []);
        setCashDrawerTotals({
          pay_in_total: opsResult.data.pay_in_total || '0',
          pay_out_total: opsResult.data.pay_out_total || '0',
        });
      }
    }
    setShiftClosed(false);
    setClosedShiftData(null);
    setEndingCash('');
    setClosingNotes('');
    setShiftDetailsOpen(true);
  };

  // Cash drawer operations
  const openCashDrawerDialog = (type: 'pay_in' | 'pay_out') => {
    setCashDrawerOpType(type);
    setCashDrawerAmount('');
    setCashDrawerReason('');
    setCashDrawerDialogOpen(true);
  };

  const handleCashDrawerOp = async () => {
    if (!cashDrawerAmount || !cashDrawerReason) return;
    setIsSubmittingOp(true);
    try {
      const result = cashDrawerOpType === 'pay_in'
        ? await api.payIn(cashDrawerAmount, cashDrawerReason)
        : await api.payOut(cashDrawerAmount, cashDrawerReason);
      if (result.error) {
        showError('Operation Failed', result.error);
      } else {
        setCashDrawerDialogOpen(false);
        await refreshShift();
      }
    } catch (error) {
      console.error('Failed cash drawer operation:', error);
      showError('Operation Failed', 'Expected error occurred during cash drawer operation.');
    } finally {
      setIsSubmittingOp(false);
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
      discount_value: discount > 0 ? discount.toString() : undefined,
      discount_type: discount > 0 ? 'percentage' : undefined,
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
        refreshShift(); // Refresh shift totals after sale
      } else if (result.error) {
        showError('Transaction Failed', result.error);
      }
    } catch (error) {
      console.error('Failed to process sale:', error);
      showError('Transaction Failed', 'An unexpected error occurred while processing the sale.');
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

  // Quick amount buttons for cash payment
  const quickAmounts = [10000, 20000, 50000, 100000];

  const cartProps = {
    cartItems,
    onClear: clearCart,
    onUpdateQuantity: updateQuantity,
    onRemove: removeFromCart,
    user,
    subtotal,
    totalTax,
    discount,
    discountAmount,
    setDiscount,
    total,
    canApplyDiscount,
    currentShift,
    onCheckout: async () => {
      // Final check for active shift before opening checkout
      const result = await api.getCurrentShift();
      if (!result.data) {
        setCurrentShift(null);
        showError('Shift Closed', 'Shift has been closed. Cannot process sale.');
        return;
      }
      setCurrentShift(result.data);

      if (cartItems.length === 0) return;
      setAmountPaid(total.toString());
      setCheckoutDialogOpen(true);
    },
  };

  return (
    <div className="flex flex-col h-full">
      <Header title="Point of Sale" />

      {/* Shift status bar */}
      {!currentShift ? (
        <div className="bg-card border-b border-l-4 border-l-yellow-500 px-6 py-4 flex items-center justify-between flex-shrink-0 shadow-sm">
          <div className="flex items-center gap-2">
            <AlertCircle className="h-5 w-5 text-yellow-600" />
            <span className="text-sm font-medium">No active shift. Start a shift to begin selling.</span>
          </div>
          {hasPermission(PERMISSIONS.POS_SHIFT_START) && (
            <Button size="sm" onClick={() => setShiftDialogOpen(true)}>
              <Clock className="h-4 w-4 mr-2" />
              Start Shift
            </Button>
          )}
        </div>
      ) : (
        <div className="bg-card border-b border-l-4 border-l-green-500 px-6 py-4 flex items-center justify-between flex-shrink-0 shadow-sm">
          <div className="flex items-center gap-2">
            <CheckCircle className="h-5 w-5 text-green-600" />
            <span className="text-sm font-medium">
              Shift active <span className="text-muted-foreground mx-2">|</span>
              Started: {new Date(currentShift.started_at).toLocaleTimeString()}
            </span>
          </div>
          <div className="flex items-center gap-2">
            <Button size="sm" variant="outline" onClick={() => openCashDrawerDialog('pay_in')} className="gap-1.5">
              <ArrowDownCircle className="h-4 w-4 text-green-600" />
              <span className="hidden sm:inline">Pay In</span>
            </Button>
            <Button size="sm" variant="outline" onClick={() => openCashDrawerDialog('pay_out')} className="gap-1.5">
              <ArrowUpCircle className="h-4 w-4 text-red-600" />
              <span className="hidden sm:inline">Pay Out</span>
            </Button>
            {hasPermission(PERMISSIONS.POS_SHIFT_END) && (
              <Button size="sm" variant="outline" onClick={openShiftDetails}>
                End Shift
              </Button>
            )}
          </div>
        </div>
      )}

      <div className="flex-1 flex overflow-hidden relative">
        {/* Left side - Product grid */}
        <div className="flex-1 flex flex-col p-4 overflow-hidden pb-16 lg:pb-4">
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
              <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5 gap-3">
                {filteredProducts.map((product) => {
                  const quantity = getProductQuantity(product);
                  const minQuantity = getProductMinQuantity(product);
                  const price = getProductPrice(product);
                  const isLowStock = quantity <= minQuantity && quantity > 0;
                  const isOutOfStock = quantity <= 0;

                  return (
                    <button
                      key={product.id}
                      onClick={() => currentShift && addToCart(product)}
                      disabled={!currentShift || isOutOfStock}
                      className={`bg-card border rounded-lg p-3 text-left hover:border-primary hover:shadow-md transition-all disabled:opacity-50 disabled:cursor-not-allowed relative overflow-hidden ${isOutOfStock ? 'opacity-60 bg-muted/50' : ''}`}
                    >
                      {/* Stock Status Badge */}
                      {(isLowStock || isOutOfStock) && (
                        <div className={`absolute top-2 right-2 px-2 py-0.5 text-[10px] font-bold uppercase tracking-wider rounded-full shadow-sm z-10 ${isOutOfStock
                          ? 'bg-red-600 text-white'
                          : 'bg-yellow-500 text-white'
                          }`}>
                          {isOutOfStock ? 'Out of Stock' : 'Low Stock'}
                        </div>
                      )}

                      <div className="aspect-square bg-muted rounded-md mb-2 flex items-center justify-center overflow-hidden">
                        {product.image_url ? (
                          <img
                            src={getImageUrl(product.image_url)}
                            alt={product.name}
                            className={`w-full h-full object-cover ${isOutOfStock ? 'grayscale' : ''}`}
                            onError={(e) => {
                              (e.target as HTMLImageElement).style.display = 'none';
                              ((e.target as HTMLImageElement).nextSibling as HTMLElement).style.display = 'flex';
                            }}
                          />
                        ) : null}
                        <ShoppingCart className="h-8 w-8 text-muted-foreground" style={{ display: product.image_url ? 'none' : 'block' }} />
                      </div>
                      <p className="font-medium text-sm truncate">{product.name}</p>
                      <p className="text-xs text-muted-foreground">{product.category_name}</p>
                      <div className="flex flex-col gap-1 mt-2">
                        <span className="font-bold text-primary">
                          {formatCurrency(price)}
                        </span>
                        <span className={`text-xs font-medium ${isOutOfStock ? 'text-red-600' : isLowStock ? 'text-yellow-600' : 'text-muted-foreground'}`}>
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

        {/* Desktop Right side - Cart Sidebar */}
        <div className="hidden lg:flex w-96 bg-card border-l flex-col h-full">
          <CartView {...cartProps} />
        </div>
      </div>

      {/* Mobile Bottom Bar */}
      <div className="lg:hidden absolute bottom-0 left-0 right-0 p-4 bg-background border-t">
        <Sheet>
          <SheetTrigger asChild>
            <Button className="w-full flex justify-between" size="lg">
              <div className="flex items-center gap-2">
                <ShoppingCart className="h-5 w-5" />
                <span>{cartItems.length} items</span>
              </div>
              <span className="font-bold">{formatCurrency(total)}</span>
            </Button>
          </SheetTrigger>
          <SheetContent className="w-full sm:max-w-md p-0 bg-card gap-0" side="right">
            <SheetHeader>
              <SheetTitle className="sr-only">Current Order</SheetTitle>
            </SheetHeader>
            <div className="h-full pt-6">
              <CartView {...cartProps} />
            </div>
          </SheetContent>
        </Sheet>
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
              placeholder="0"
              className="mt-1"
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

      {/* Cash Drawer Operation Dialog */}
      <Dialog open={cashDrawerDialogOpen} onOpenChange={setCashDrawerDialogOpen}>
        <DialogContent className="sm:max-w-sm">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              {cashDrawerOpType === 'pay_in' ? (
                <><ArrowDownCircle className="h-5 w-5 text-green-600" /> Pay In</>
              ) : (
                <><ArrowUpCircle className="h-5 w-5 text-red-600" /> Pay Out</>
              )}
            </DialogTitle>
            <DialogDescription>
              {cashDrawerOpType === 'pay_in'
                ? 'Add cash to the drawer (e.g., change float, deposit).'
                : 'Remove cash from the drawer (e.g., petty cash, withdrawal).'}
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div>
              <label className="text-sm font-medium">Amount (IDR)</label>
              <Input
                type="number"
                value={cashDrawerAmount}
                onChange={(e) => setCashDrawerAmount(e.target.value)}
                placeholder="Enter amount..."
                className="mt-1.5"
              />
            </div>
            <div>
              <label className="text-sm font-medium">Reason</label>
              <Input
                value={cashDrawerReason}
                onChange={(e) => setCashDrawerReason(e.target.value)}
                placeholder="e.g., Petty cash withdrawal"
                className="mt-1.5"
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCashDrawerDialogOpen(false)}>Cancel</Button>
            <Button
              onClick={handleCashDrawerOp}
              disabled={!cashDrawerAmount || !cashDrawerReason || isSubmittingOp}
              className={cashDrawerOpType === 'pay_in' ? '' : 'bg-red-600 hover:bg-red-700'}
            >
              {isSubmittingOp ? 'Processing...' : (cashDrawerOpType === 'pay_in' ? 'Confirm Pay In' : 'Confirm Pay Out')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Shift Details / End Shift Dialog */}
      <Dialog open={shiftDetailsOpen} onOpenChange={(open) => {
        if (!open) {
          setShiftDetailsOpen(false);
          if (shiftClosed) {
            router.refresh();
          }
        }
      }}>
        <DialogContent className="sm:max-w-md">
          {shiftClosed && closedShiftData ? (
            <>
              <DialogHeader>
                <DialogTitle className="text-center">
                  <CheckCircle className="h-12 w-12 text-green-500 mx-auto mb-2" />
                  Shift Closed
                </DialogTitle>
                <DialogDescription className="text-center">
                  Here is your shift reconciliation summary.
                </DialogDescription>
              </DialogHeader>
              <div className="space-y-3 py-4">
                <div className="grid grid-cols-2 gap-3 text-sm">
                  <div>
                    <span className="text-muted-foreground block text-xs">Opening Cash</span>
                    <span className="font-medium">{formatCurrency(parseFloat(closedShiftData.opening_cash) || 0)}</span>
                  </div>
                  <div>
                    <span className="text-muted-foreground block text-xs">Closing Cash</span>
                    <span className="font-medium">{formatCurrency(parseFloat(closedShiftData.closing_cash || '0'))}</span>
                  </div>
                  <div>
                    <span className="text-muted-foreground block text-xs">Total Sales ({closedShiftData.transaction_count} txn)</span>
                    <span className="font-medium text-green-600">+{formatCurrency(parseFloat(closedShiftData.total_sales) || 0)}</span>
                  </div>
                  <div>
                    <span className="text-muted-foreground block text-xs">Total Cash Sales</span>
                    <span className="font-medium text-blue-600">+{formatCurrency(parseFloat(closedShiftData.total_cash_sales) || 0)}</span>
                  </div>
                  <div>
                    <span className="text-muted-foreground block text-xs">Total Refunds ({closedShiftData.refund_count})</span>
                    <span className="font-medium text-red-600">-{formatCurrency(parseFloat(closedShiftData.total_refunds) || 0)}</span>
                  </div>
                  <div>
                    <span className="text-muted-foreground block text-xs">Expected Cash</span>
                    <span className="font-bold">{formatCurrency(parseFloat(closedShiftData.expected_cash || '0'))}</span>
                  </div>
                  <div>
                    <span className="text-muted-foreground block text-xs">Difference</span>
                    <span className={`font-bold ${parseFloat(closedShiftData.cash_difference || '0') >= 0 ? 'text-green-600' : 'text-red-600'}`}>
                      {formatCurrency(parseFloat(closedShiftData.cash_difference || '0'))}
                    </span>
                  </div>
                </div>
              </div>
              <DialogFooter>
                <Button className="w-full" onClick={() => {
                  setShiftDetailsOpen(false);
                  router.refresh();
                }}>
                  Done
                </Button>
              </DialogFooter>
            </>
          ) : currentShift ? (
            <>
              <DialogHeader>
                <DialogTitle>End Shift</DialogTitle>
                <DialogDescription>
                  Count your cash and enter the total below. Shift details will be shown after closing.
                </DialogDescription>
              </DialogHeader>

              <div className="space-y-4 py-4">
                <div className="grid grid-cols-2 gap-4 text-sm">
                  <div>
                    <span className="text-muted-foreground block text-xs">Started At</span>
                    <span className="font-medium">{new Date(currentShift.started_at).toLocaleString()}</span>
                  </div>
                  <div>
                    <span className="text-muted-foreground block text-xs">Opening Cash</span>
                    <span className="font-medium">{formatCurrency(parseFloat(currentShift.opening_cash) || 0)}</span>
                  </div>
                </div>

                {/* Cash Drawer Operations Summary */}
                {cashDrawerOps.length > 0 && (
                  <>
                    <Separator />
                    <div>
                      <h4 className="text-sm font-medium mb-2">Cash Drawer Activity</h4>
                      <div className="space-y-1.5 max-h-32 overflow-y-auto">
                        {cashDrawerOps.map((op) => (
                          <div key={op.id} className="flex items-center justify-between text-xs">
                            <div className="flex items-center gap-1.5">
                              {op.type === 'pay_in' ? (
                                <ArrowDownCircle className="h-3.5 w-3.5 text-green-600" />
                              ) : (
                                <ArrowUpCircle className="h-3.5 w-3.5 text-red-600" />
                              )}
                              <span className="text-muted-foreground">{op.reason}</span>
                            </div>
                            <span className={op.type === 'pay_in' ? 'text-green-600 font-medium' : 'text-red-600 font-medium'}>
                              {op.type === 'pay_in' ? '+' : '-'}{formatCurrency(parseFloat(op.amount))}
                            </span>
                          </div>
                        ))}
                      </div>
                      <div className="flex justify-between text-xs mt-2 pt-2 border-t">
                        <span className="text-green-600">Pay-In: +{formatCurrency(parseFloat(cashDrawerTotals.pay_in_total))}</span>
                        <span className="text-red-600">Pay-Out: -{formatCurrency(parseFloat(cashDrawerTotals.pay_out_total))}</span>
                      </div>
                    </div>
                  </>
                )}

                <Separator />

                <div>
                  <label className="text-sm font-medium">Ending Cash (Counted)</label>
                  <Input
                    type="number"
                    value={endingCash}
                    onChange={(e) => setEndingCash(e.target.value)}
                    placeholder="Count the cash in your drawer..."
                    className="mt-1.5"
                  />
                  <p className="text-xs text-muted-foreground mt-1">
                    Enter the actual amount of cash you have in the drawer.
                  </p>
                </div>

                <div>
                  <label className="text-sm font-medium">Notes (Optional)</label>
                  <Input
                    value={closingNotes}
                    onChange={(e) => setClosingNotes(e.target.value)}
                    placeholder="Any discrepancies or comments..."
                    className="mt-1.5"
                  />
                </div>
              </div>

              <DialogFooter>
                <Button variant="outline" onClick={() => setShiftDetailsOpen(false)}>
                  Cancel
                </Button>
                <Button
                  variant="destructive"
                  onClick={handleEndShift}
                  disabled={!endingCash || isProcessing}
                >
                  {isProcessing ? 'Ending Shift...' : 'End Shift'}
                </Button>
              </DialogFooter>
            </>
          ) : (
            <>
              <DialogHeader>
                <DialogTitle>No Active Shift</DialogTitle>
                <DialogDescription>There is no active shift to end.</DialogDescription>
              </DialogHeader>
              <DialogFooter>
                <Button onClick={() => setShiftDetailsOpen(false)}>Close</Button>
              </DialogFooter>
            </>
          )}
        </DialogContent>
      </Dialog>

      {/* Checkout Dialog */}
      <Dialog open={checkoutDialogOpen} onOpenChange={closeCheckoutDialog}>
        <DialogContent className="sm:max-w-md w-full max-w-[95vw]">
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
                    <div className="grid grid-cols-3 gap-2 mt-2">
                      {quickAmounts.map((amount) => (
                        <Button
                          key={amount}
                          variant="outline"
                          size="sm"
                          onClick={() => setAmountPaid(amount.toString())}
                          className="w-full"
                        >
                          {amount / 1000}K
                        </Button>
                      ))}
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => setAmountPaid(total.toString())}
                        className="col-span-2 w-full"
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

      {/* Error / Alert Dialog */}
      <Dialog open={errorDialogOpen} onOpenChange={setErrorDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle className="text-destructive flex items-center gap-2">
              <AlertCircle className="h-5 w-5" />
              {errorTitle}
            </DialogTitle>
            <DialogDescription className="pt-2 text-base text-foreground">
              {errorMessage}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button onClick={() => setErrorDialogOpen(false)}>OK</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

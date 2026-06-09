"use client";

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { Header } from "@/components/layout/header";
import { Button } from "@/components/ui/button";
import { Sheet, SheetClose, SheetContent, SheetHeader, SheetTitle, SheetTrigger } from "@/components/ui/sheet";
import { ShoppingCart } from "lucide-react";
import api from "@/lib/api";
import { PERMISSIONS, useAuth } from "@/contexts/auth-context";
import type {
  CartItem,
  CashDrawerOperation,
  CashDrawerOperationsResponse,
  Category,
  PaymentMethod,
  Product,
  Shift,
} from "@/types";

import { PosCartView } from "./pos-cart-view";
import { PosCashDrawerDialog } from "./pos-cash-drawer-dialog";
import { PosCheckoutDialog } from "./pos-checkout-dialog";
import { useGlobalError } from "@/contexts/error-context";
import {
  addCartItem,
  buildSaleCartValidationRequest,
  buildSaleRequest,
  canSubmitEndShift,
  canSubmitStartShift,
  calculateCartTotals,
  formatCurrency,
  parseNumericInput,
  removeCartItem,
  updateCartItemQuantity,
} from "./pos-helpers";
import { PosProductGrid } from "./pos-product-grid";
import { PosShiftDetailsDialog } from "./pos-shift-details-dialog";
import { PosShiftStatusBar } from "./pos-shift-status-bar";
import { PosStartShiftDialog } from "./pos-start-shift-dialog";

function getApiErrorMessage(error: string | undefined, fallback: string): string {
  return error || fallback;
}

function createEmptyCashDrawerTotals() {
  return {
    pay_in_total: "0",
    pay_out_total: "0",
  };
}

export function POSScreen() {
  const router = useRouter();
  const { hasPermission, user, isLoading: isAuthLoading } = useAuth();
  const canApplyDiscount = hasPermission(PERMISSIONS.SALES_CREATE);
  const canCreateSale = hasPermission(PERMISSIONS.SALES_CREATE);
  const canStartShift = hasPermission(PERMISSIONS.POS_SHIFT_START);
  const canEndShift = hasPermission(PERMISSIONS.POS_SHIFT_END);

  const [products, setProducts] = useState<Product[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [selectedCategory, setSelectedCategory] = useState<string>("all");
  const [searchQuery, setSearchQuery] = useState("");
  const [debouncedSearchQuery, setDebouncedSearchQuery] = useState("");
  const [isLoadingProducts, setIsLoadingProducts] = useState(true);
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);
  const [isFetchingMore, setIsFetchingMore] = useState(false);

  const [cartItems, setCartItems] = useState<CartItem[]>([]);
  const [discount, setDiscount] = useState(0);

  const [currentShift, setCurrentShift] = useState<Shift | null>(null);
  const [shiftDialogOpen, setShiftDialogOpen] = useState(false);
  const [startingCash, setStartingCash] = useState("");

  const [shiftDetailsOpen, setShiftDetailsOpen] = useState(false);
  const [endingCash, setEndingCash] = useState("");
  const [closingNotes, setClosingNotes] = useState("");
  const [shiftClosed, setShiftClosed] = useState(false);
  const [closedShiftData, setClosedShiftData] = useState<Shift | null>(null);

  const [cashDrawerDialogOpen, setCashDrawerDialogOpen] = useState(false);
  const [cashDrawerOpType, setCashDrawerOpType] = useState<"pay_in" | "pay_out">("pay_in");
  const [cashDrawerAmount, setCashDrawerAmount] = useState("");
  const [cashDrawerReason, setCashDrawerReason] = useState("");
  const [cashDrawerOps, setCashDrawerOps] = useState<CashDrawerOperation[]>([]);
  const [cashDrawerTotals, setCashDrawerTotals] = useState(createEmptyCashDrawerTotals);
  const [isSubmittingOp, setIsSubmittingOp] = useState(false);

  const [checkoutDialogOpen, setCheckoutDialogOpen] = useState(false);
  const [paymentMethod, setPaymentMethod] = useState<PaymentMethod>("cash");
  const [amountPaid, setAmountPaid] = useState("");
  const [isProcessing, setIsProcessing] = useState(false);
  const [saleComplete, setSaleComplete] = useState(false);
  const [lastInvoice, setLastInvoice] = useState("");
  const [lastChange, setLastChange] = useState(0);

  const { showError } = useGlobalError();

  const loadMoreRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    if (!isAuthLoading && !hasPermission(PERMISSIONS.POS_VIEW)) {
      router.replace("/");
    }
  }, [hasPermission, isAuthLoading, router]);

  useEffect(() => {
    const timeout = setTimeout(() => {
      setDebouncedSearchQuery(searchQuery);
    }, 300);

    return () => clearTimeout(timeout);
  }, [searchQuery]);

  const totals = useMemo(
    () => calculateCartTotals(cartItems, discount, parseNumericInput(amountPaid)),
    [amountPaid, cartItems, discount],
  );

  const refreshShift = useCallback(async () => {
    const result = await api.getCurrentShift();

    if (result.error) {
      showError("Shift Refresh Failed", getApiErrorMessage(result.error, "Failed to refresh shift."));
      return;
    }

    if (!result.data && currentShift) {
      setCurrentShift(null);
      showError("Shift Closed", "The shift has been closed by another user. Sales are disabled.");
      return;
    }

    setCurrentShift(result.data || null);
  }, [currentShift, showError]);

  useEffect(() => {
    const fetchInitialData = async () => {
      const [categoriesResult, shiftResult] = await Promise.all([
        api.getCategories(),
        api.getCurrentShift(),
      ]);

      if (categoriesResult.error) {
        showError("Categories Load Failed", getApiErrorMessage(categoriesResult.error, "Failed to load categories."));
      } else if (categoriesResult.data) {
        setCategories(categoriesResult.data);
      }

      if (shiftResult.error) {
        showError("Shift Load Failed", getApiErrorMessage(shiftResult.error, "Failed to load shift state."));
      } else {
        setCurrentShift(shiftResult.data || null);
      }
    };

    void fetchInitialData();
  }, [showError]);

  const fetchProductsPage = useCallback(
    async (pageToLoad: number, replace: boolean, silent = false) => {
      if (replace && !silent) {
        setIsLoadingProducts(true);
      } else if (!replace) {
        setIsFetchingMore(true);
      }

      try {
        const result = await api.getProductsPage({
          active: true,
          category_id: selectedCategory !== "all" ? selectedCategory : undefined,
          search: debouncedSearchQuery || undefined,
          page: pageToLoad,
          per_page: 24,
        });

        if (result.error) {
          showError("Products Load Failed", getApiErrorMessage(result.error, "Failed to fetch products."));
          return;
        }

        if (result.data) {
          setProducts((current) => (replace ? result.data! : [...current, ...result.data!]));
          if (result.total_pages !== undefined) {
            setHasMore(pageToLoad < result.total_pages);
          } else {
            setHasMore(result.data.length === 24);
          }
          setPage(pageToLoad);
        }
      } finally {
        if (replace && !silent) {
          setIsLoadingProducts(false);
        } else if (!replace) {
          setIsFetchingMore(false);
        }
      }
    },
    [debouncedSearchQuery, selectedCategory, showError],
  );

  const refreshProducts = useCallback(
    async (silent = true) => {
      if (silent) {
        setHasMore(true);
      }
      await fetchProductsPage(1, true, silent);
    },
    [fetchProductsPage],
  );

  const resetCashDrawerSummary = useCallback(() => {
    setCashDrawerOps([]);
    setCashDrawerTotals(createEmptyCashDrawerTotals());
  }, []);

  useEffect(() => {
    setProducts([]);
    setPage(1);
    setHasMore(true);
    void fetchProductsPage(1, true);
  }, [fetchProductsPage]);

  useEffect(() => {
    const node = loadMoreRef.current;
    if (!node || !hasMore || isLoadingProducts || isFetchingMore) {
      return;
    }

    const observer = new IntersectionObserver(
      (entries) => {
        const firstEntry = entries[0];
        if (firstEntry.isIntersecting) {
          void fetchProductsPage(page + 1, false);
        }
      },
      { rootMargin: "200px" },
    );

    observer.observe(node);
    return () => observer.disconnect();
  }, [fetchProductsPage, hasMore, isFetchingMore, isLoadingProducts, page]);

  useEffect(() => {
    const pollData = async () => {
      await Promise.all([refreshShift(), refreshProducts()]);
    };

    const interval = setInterval(() => {
      void pollData();
    }, 30000);

    return () => clearInterval(interval);
  }, [refreshProducts, refreshShift]);

  const addToCart = useCallback((product: Product) => {
    if (!currentShift) {
      return;
    }
    setCartItems((items) => addCartItem(items, product));
  }, [currentShift]);

  const updateQuantity = useCallback((productId: string, delta: number) => {
    setCartItems((items) => updateCartItemQuantity(items, productId, delta));
  }, []);

  const removeFromCart = useCallback((productId: string) => {
    setCartItems((items) => removeCartItem(items, productId));
  }, []);

  const clearCart = useCallback(() => {
    setCartItems([]);
    setDiscount(0);
  }, []);

  const handleStartShift = useCallback(async () => {
    if (!startingCash) {
      showError("Missing Amount", "Enter a starting cash amount before starting a shift.");
      return;
    }

    if (
      !canSubmitStartShift({
        canStartShift,
        startingCash,
      })
    ) {
      if (!canStartShift) {
        showError("Permission Required", "You no longer have permission to start a shift.");
      }
      return;
    }

    const result = await api.startShift(parseNumericInput(startingCash));
    if (result.error) {
      showError("Start Shift Failed", getApiErrorMessage(result.error, "Failed to start shift."));
      return;
    }

    if (result.data) {
      setCurrentShift(result.data);
      setShiftDialogOpen(false);
      setStartingCash("");
    }
  }, [canStartShift, showError, startingCash]);

  const handleEndShift = useCallback(async () => {
    if (
      !canSubmitEndShift({
        canEndShift,
        endingCash,
        isProcessing,
      })
    ) {
      if (!canEndShift) {
        showError("Permission Required", "You no longer have permission to end a shift.");
      }
      return;
    }

    const cashAmount = parseNumericInput(endingCash);
    if (Number.isNaN(cashAmount) || cashAmount < 0) {
      showError("Invalid Amount", "Please enter a valid positive cash amount");
      return;
    }

    setIsProcessing(true);
    try {
      const result = await api.closeShift(endingCash, closingNotes, currentShift?.id);
      if (result.error) {
        showError("End Shift Failed", getApiErrorMessage(result.error, "Failed to end shift."));
        return;
      }

      if (result.data) {
        setClosedShiftData(result.data);
        setShiftClosed(true);
        setCurrentShift(null);
      }
    } finally {
      setIsProcessing(false);
    }
  }, [canEndShift, closingNotes, currentShift?.id, endingCash, isProcessing, showError]);

  const openShiftDetails = useCallback(async () => {
    const shiftResult = await api.getCurrentShift();
    if (shiftResult.error) {
      showError("Shift Load Failed", getApiErrorMessage(shiftResult.error, "Failed to load shift details."));
      return;
    }

    const nextShift = shiftResult.data || null;
    setCurrentShift(nextShift);
    resetCashDrawerSummary();

    if (nextShift) {
      const operationsResult = await api.getShiftOperations(nextShift.id);
      if (operationsResult.error) {
        showError(
          "Cash Drawer Load Failed",
          getApiErrorMessage(operationsResult.error, "Failed to load cash drawer operations."),
        );
      } else if (operationsResult.data) {
        const data = operationsResult.data as CashDrawerOperationsResponse;
        setCashDrawerOps(data.operations || []);
        setCashDrawerTotals({
          pay_in_total: data.pay_in_total || "0",
          pay_out_total: data.pay_out_total || "0",
        });
      }
    }

    setShiftClosed(false);
    setClosedShiftData(null);
    setEndingCash("");
    setClosingNotes("");
    setShiftDetailsOpen(true);
  }, [resetCashDrawerSummary, showError]);

  const openCashDrawerDialog = useCallback((type: "pay_in" | "pay_out") => {
    setCashDrawerOpType(type);
    setCashDrawerAmount("");
    setCashDrawerReason("");
    setCashDrawerDialogOpen(true);
  }, []);

  const handleCashDrawerOp = useCallback(async () => {
    if (!cashDrawerAmount) {
      showError("Missing Amount", "Enter an amount before submitting this cash drawer operation.");
      return;
    }

    if (!cashDrawerReason) {
      showError("Missing Reason", "Enter a reason before submitting this cash drawer operation.");
      return;
    }

    setIsSubmittingOp(true);
    try {
      const result =
        cashDrawerOpType === "pay_in"
          ? await api.payIn(cashDrawerAmount, cashDrawerReason, currentShift?.id)
          : await api.payOut(cashDrawerAmount, cashDrawerReason, currentShift?.id);

      if (result.error) {
        showError("Operation Failed", getApiErrorMessage(result.error, "Cash drawer operation failed."));
        return;
      }

      setCashDrawerDialogOpen(false);
      await refreshShift();
    } finally {
      setIsSubmittingOp(false);
    }
  }, [cashDrawerAmount, cashDrawerOpType, cashDrawerReason, currentShift?.id, refreshShift, showError]);

  const handleCompleteSale = useCallback(async () => {
    if (!currentShift || cartItems.length === 0) {
      return;
    }

    if (paymentMethod === "cash" && parseNumericInput(amountPaid) < totals.total) {
      showError("Insufficient Payment", "Amount received must cover the total before completing the sale.");
      return;
    }

    setIsProcessing(true);
    try {
      const saleRequest = buildSaleRequest(
        cartItems,
        paymentMethod,
        discount,
        totals.total,
        parseNumericInput(amountPaid),
        currentShift.id,
      );

      const result = await api.createSale(saleRequest);
      if (result.error) {
        showError("Transaction Failed", getApiErrorMessage(result.error, "Failed to process sale."));
        return;
      }

      if (result.data) {
        setLastInvoice(result.data.invoice_no);
        setLastChange(totals.change > 0 ? totals.change : 0);
        setSaleComplete(true);
        clearCart();
        setAmountPaid("");
        await refreshShift();
      }
    } finally {
      setIsProcessing(false);
    }
  }, [amountPaid, cartItems, clearCart, currentShift, discount, paymentMethod, refreshShift, showError, totals]);

  const openCheckout = useCallback(async () => {
    const result = await api.getCurrentShift();
    if (result.error) {
      showError("Shift Check Failed", getApiErrorMessage(result.error, "Failed to validate active shift."));
      return;
    }

    if (!result.data) {
      setCurrentShift(null);
      showError("Shift Closed", "Shift has been closed. Cannot process sale.");
      return;
    }

    setCurrentShift(result.data);
    if (cartItems.length === 0) {
      return;
    }

    const validationResult = await api.validateSaleCart(
      buildSaleCartValidationRequest(cartItems, result.data.id),
    );
    if (validationResult.error) {
      showError("Checkout Failed", getApiErrorMessage(validationResult.error, "Cart is no longer valid."));
      return;
    }

    setAmountPaid(totals.total.toString());
    setCheckoutDialogOpen(true);
  }, [cartItems, showError, totals.total]);

  const closeCheckoutDialog = useCallback(() => {
    setCheckoutDialogOpen(false);
    setSaleComplete(false);
    setAmountPaid("");
    setLastChange(0);
  }, []);

  const cartProps = useMemo(
    () => ({
      cartItems,
      subtotal: totals.subtotal,
      totalTax: totals.totalTax,
      discount,
      discountAmount: totals.discountAmount,
      total: totals.total,
      currentShift,
      canApplyDiscount,
      canCreateSale,
      cashierName: user?.name,
      onClear: clearCart,
      onUpdateQuantity: updateQuantity,
      onRemove: removeFromCart,
      onDiscountChange: setDiscount,
      onCheckout: openCheckout,
    }),
    [
      canApplyDiscount,
      canCreateSale,
      cartItems,
      clearCart,
      currentShift,
      discount,
      openCheckout,
      removeFromCart,
      totals,
      updateQuantity,
      user?.name,
    ],
  );

  return (
    <div className="flex flex-col h-full">
      <Header title="Point of Sale" />

      <PosShiftStatusBar
        currentShift={currentShift}
        canStartShift={canStartShift}
        canEndShift={canEndShift}
        onStartShift={() => setShiftDialogOpen(true)}
        onOpenShiftDetails={openShiftDetails}
        onOpenCashDrawerDialog={openCashDrawerDialog}
      />

      <div className="flex-1 flex overflow-hidden relative">
        <PosProductGrid
          products={products}
          categories={categories.map((category) => ({ id: category.id, name: category.name }))}
          currentShift={currentShift}
          isLoadingProducts={isLoadingProducts}
          isFetchingMore={isFetchingMore}
          hasMore={hasMore}
          searchQuery={searchQuery}
          selectedCategory={selectedCategory}
          onSearchChange={setSearchQuery}
          onCategoryChange={setSelectedCategory}
          onAddToCart={addToCart}
          loadMoreRef={loadMoreRef}
        />

        <div className="hidden lg:flex w-96 bg-card border-l flex-col h-full">
          <PosCartView {...cartProps} />
        </div>
      </div>

      <div className="lg:hidden absolute bottom-0 left-0 right-0 p-4 bg-background border-t">
        <Sheet>
          <SheetTrigger asChild>
            <Button className="w-full flex justify-between" size="lg">
              <div className="flex items-center gap-2">
                <ShoppingCart className="h-5 w-5" />
                <span>{cartItems.length} items</span>
              </div>
              <span className="font-bold font-mono">{formatCurrency(totals.total)}</span>
            </Button>
          </SheetTrigger>
          <SheetContent className="w-full sm:max-w-md p-0 bg-card gap-0" side="right">
            <SheetHeader>
              <SheetTitle className="sr-only">Current Order</SheetTitle>
            </SheetHeader>
            <div className="flex h-full flex-col pt-6">
              <div className="min-h-0 flex-1">
                <PosCartView {...cartProps} />
              </div>
              <div className="border-t p-4">
                <SheetClose asChild>
                  <Button variant="outline" className="w-full">
                    Close
                  </Button>
                </SheetClose>
              </div>
            </div>
          </SheetContent>
        </Sheet>
      </div>

      <PosStartShiftDialog
        open={shiftDialogOpen}
        startingCash={startingCash}
        onOpenChange={setShiftDialogOpen}
        onStartingCashChange={setStartingCash}
        onSubmit={handleStartShift}
      />

      <PosCashDrawerDialog
        open={cashDrawerDialogOpen}
        operationType={cashDrawerOpType}
        amount={cashDrawerAmount}
        reason={cashDrawerReason}
        isSubmitting={isSubmittingOp}
        onOpenChange={setCashDrawerDialogOpen}
        onAmountChange={setCashDrawerAmount}
        onReasonChange={setCashDrawerReason}
        onSubmit={handleCashDrawerOp}
      />

      <PosShiftDetailsDialog
        open={shiftDetailsOpen}
        currentShift={currentShift}
        shiftClosed={shiftClosed}
        closedShiftData={closedShiftData}
        endingCash={endingCash}
        closingNotes={closingNotes}
        cashDrawerOps={cashDrawerOps}
        cashDrawerTotals={cashDrawerTotals}
        isProcessing={isProcessing}
        onOpenChange={(open) => {
          setShiftDetailsOpen(open);
          if (!open && shiftClosed) {
            router.refresh();
          }
        }}
        onEndingCashChange={setEndingCash}
        onClosingNotesChange={setClosingNotes}
        onSubmit={handleEndShift}
        onDone={() => {
          setShiftDetailsOpen(false);
          router.refresh();
        }}
      />

      <PosCheckoutDialog
        open={checkoutDialogOpen}
        total={totals.total}
        paymentMethod={paymentMethod}
        amountPaid={amountPaid}
        isProcessing={isProcessing}
        saleComplete={saleComplete}
        lastInvoice={lastInvoice}
        lastChange={lastChange}
        onOpenChange={closeCheckoutDialog}
        onPaymentMethodChange={setPaymentMethod}
        onAmountPaidChange={setAmountPaid}
        onSubmit={handleCompleteSale}
      />


    </div>
  );
}

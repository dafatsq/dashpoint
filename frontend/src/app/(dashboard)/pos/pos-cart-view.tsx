"use client";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Separator } from "@/components/ui/separator";
import { CreditCard, Minus, Plus, ShoppingCart, Trash2 } from "lucide-react";

import { formatCurrency, getProductPrice } from "./pos-helpers";
import type { PosCartViewProps } from "./pos-types";

export function PosCartView({
  cartItems,
  subtotal,
  totalTax,
  discount,
  discountAmount,
  total,
  currentShift,
  canApplyDiscount,
  canCreateSale,
  cashierName,
  onClear,
  onUpdateQuantity,
  onRemove,
  onDiscountChange,
  onCheckout,
}: PosCartViewProps) {
  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center justify-between p-4 border-b">
        <h2 className="font-semibold text-lg">Current Order</h2>
        {cartItems.length > 0 ? (
          <Button variant="ghost" size="sm" onClick={onClear}>
            Clear
          </Button>
        ) : null}
      </div>

      <div className="flex-1 flex flex-col p-4 overflow-hidden">
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
                  <p className="font-medium text-sm break-words">{item.product.name}</p>
                  <p className="text-sm text-muted-foreground font-mono">
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
                  <span className="w-8 text-center text-sm font-medium">{item.quantity}</span>
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

        <div className="space-y-2 text-sm">
          <div className="flex justify-between items-center pb-2 border-b">
            <span className="text-muted-foreground">Cashier</span>
            <span className="font-medium">{cashierName || "Unknown"}</span>
          </div>
          <div className="flex justify-between">
            <span className="text-muted-foreground">Subtotal</span>
            <span className="font-mono">{formatCurrency(subtotal)}</span>
          </div>
          {totalTax > 0 ? (
            <div className="flex justify-between">
              <span className="text-muted-foreground">Tax</span>
              <span className="font-mono">{formatCurrency(totalTax)}</span>
            </div>
          ) : null}
          {canApplyDiscount ? (
            <div className="space-y-1">
              <div className="flex justify-between items-center">
                <span className="text-muted-foreground">Discount (%)</span>
                <div className="flex items-center gap-1">
                  <Input
                    type="number"
                    value={discount || ""}
                    onChange={(event) => {
                      const value = parseFloat(event.target.value) || 0;
                      onDiscountChange(Math.min(100, Math.max(0, value)));
                    }}
                    className="w-16 h-7 text-right"
                    placeholder="0"
                    max="100"
                    min="0"
                  />
                  <span className="text-sm font-medium w-4">%</span>
                </div>
              </div>
              {discountAmount > 0 ? (
                <div className="flex justify-between text-destructive">
                  <span className="text-xs">Discount Amount</span>
                  <span className="font-mono">-{formatCurrency(discountAmount)}</span>
                </div>
              ) : null}
            </div>
          ) : null}
          <Separator />
          <div className="flex justify-between text-lg font-bold">
            <span>Total</span>
            <span className="text-primary font-mono">{formatCurrency(total)}</span>
          </div>
        </div>

        <Button
          size="lg"
          className="w-full mt-4"
          disabled={!currentShift || cartItems.length === 0 || !canCreateSale}
          onClick={() => void onCheckout()}
        >
          <CreditCard className="h-5 w-5 mr-2" />
          Checkout
        </Button>
      </div>
    </div>
  );
}

"use client";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Banknote, Building2, CheckCircle, CreditCard, Loader2, QrCode } from "lucide-react";
import type { PaymentMethod } from "@/types";

import {
  formatCurrencyInputValue,
  formatCurrency,
  parseNumericInput,
  QUICK_CASH_AMOUNTS,
  roundCurrencyAmount,
} from "./pos-helpers";

interface PosCheckoutDialogProps {
  open: boolean;
  total: number;
  paymentMethod: PaymentMethod;
  amountPaid: string;
  isProcessing: boolean;
  saleComplete: boolean;
  lastInvoice: string;
  lastChange: number;
  onOpenChange: () => void;
  onPaymentMethodChange: (method: PaymentMethod) => void;
  onAmountPaidChange: (value: string) => void;
  onSubmit: () => Promise<void>;
}

export function PosCheckoutDialog({
  open,
  total,
  paymentMethod,
  amountPaid,
  isProcessing,
  saleComplete,
  lastInvoice,
  lastChange,
  onOpenChange,
  onPaymentMethodChange,
  onAmountPaidChange,
  onSubmit,
}: PosCheckoutDialogProps) {
  const amountPaidNumber = parseNumericInput(amountPaid);
  const change = roundCurrencyAmount(amountPaidNumber - total);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md w-full max-w-[95vw]">
        {saleComplete ? (
          <>
            <DialogHeader>
              <DialogTitle className="text-center">
                <CheckCircle className="h-12 w-12 text-green-500 mx-auto mb-2" />
                Sale Complete!
              </DialogTitle>
              <DialogDescription className="text-center">Invoice: {lastInvoice}</DialogDescription>
            </DialogHeader>
            <div className="text-center py-4">
              <p className="text-2xl font-bold text-primary">Change: <span className="font-mono">{formatCurrency(lastChange)}</span></p>
            </div>
            <DialogFooter>
              <Button className="w-full" onClick={onOpenChange}>
                New Sale
              </Button>
            </DialogFooter>
          </>
        ) : (
          <>
            <DialogHeader>
              <DialogTitle>Checkout</DialogTitle>
              <DialogDescription>
                Total: <span className="font-bold text-primary font-mono">{formatCurrency(total)}</span>
              </DialogDescription>
            </DialogHeader>

            <div className="space-y-4 py-4">
              <div>
                <label className="text-sm font-medium mb-2 block">Payment Method</label>
                <div className="grid grid-cols-2 gap-2">
                  <Button
                    variant={paymentMethod === "cash" ? "default" : "outline"}
                    onClick={() => onPaymentMethodChange("cash")}
                    className="flex items-center gap-2"
                  >
                    <Banknote className="h-4 w-4" />
                    Cash
                  </Button>
                  <Button
                    variant={paymentMethod === "card" ? "default" : "outline"}
                    onClick={() => onPaymentMethodChange("card")}
                    className="flex items-center gap-2"
                  >
                    <CreditCard className="h-4 w-4" />
                    Card
                  </Button>
                  <Button
                    variant={paymentMethod === "qris" ? "default" : "outline"}
                    onClick={() => onPaymentMethodChange("qris")}
                    className="flex items-center gap-2"
                  >
                    <QrCode className="h-4 w-4" />
                    QRIS
                  </Button>
                  <Button
                    variant={paymentMethod === "transfer" ? "default" : "outline"}
                    onClick={() => onPaymentMethodChange("transfer")}
                    className="flex items-center gap-2"
                  >
                    <Building2 className="h-4 w-4" />
                    Transfer
                  </Button>
                </div>
              </div>

              {paymentMethod === "cash" ? (
                <div>
                  <label className="text-sm font-medium mb-2 block">Amount Received</label>
                  <Input
                    type="number"
                    value={amountPaid}
                    onChange={(event) => onAmountPaidChange(event.target.value)}
                    placeholder="Enter amount..."
                  />
                  <div className="grid grid-cols-3 gap-2 mt-2">
                    {QUICK_CASH_AMOUNTS.map((amount) => (
                      <Button
                        key={amount}
                        variant="outline"
                        size="sm"
                        onClick={() => onAmountPaidChange(amount.toString())}
                        className="w-full"
                      >
                        {amount / 1000}K
                      </Button>
                    ))}
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => onAmountPaidChange(formatCurrencyInputValue(total))}
                      className="col-span-2 w-full"
                    >
                      Exact
                    </Button>
                  </div>
                  {amountPaidNumber >= total ? (
                    <p className="mt-2 text-sm">
                      Change: <span className="font-bold text-primary font-mono">{formatCurrency(change)}</span>
                    </p>
                  ) : null}
                </div>
              ) : null}
            </div>

            <DialogFooter>
              <Button variant="outline" onClick={onOpenChange}>
                Cancel
              </Button>
              <Button
                onClick={() => void onSubmit()}
                disabled={isProcessing}
              >
                {isProcessing ? (
                  <>
                    <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                    Processing...
                  </>
                ) : (
                  "Complete Sale"
                )}
              </Button>
            </DialogFooter>
          </>
        )}
      </DialogContent>
    </Dialog>
  );
}

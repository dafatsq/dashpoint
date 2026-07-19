'use client';

import { XCircle } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import type { PaymentMethod, Sale } from "@/types";
import { useAuth, PERMISSIONS } from "@/contexts/auth-context";

import { formatSalesCurrency, formatSalesDate, getPrimarySalePaymentMethod, getSalesStatusBadge } from "./sales-helpers";

interface SalesDetailDialogProps {
  open: boolean;
  sale: Sale | null;
  paymentIcons: Record<PaymentMethod, React.ReactNode>;
  onOpenChange: (open: boolean) => void;
  onVoidRequest: () => void;
}

export function SalesDetailDialog({ open, sale, paymentIcons, onOpenChange, onVoidRequest }: SalesDetailDialogProps) {
  const { hasPermission } = useAuth();
  const paymentMethod = sale ? getPrimarySalePaymentMethod(sale) : "cash";

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>Sale Details</DialogTitle>
          <DialogDescription>Invoice: {sale?.invoice_no}</DialogDescription>
        </DialogHeader>

        {sale ? (
          <div className="space-y-4">
            <div className="grid grid-cols-2 gap-4 text-sm">
              <div>
                <p className="text-muted-foreground">Date</p>
                <p className="font-medium">{formatSalesDate(sale.created_at)}</p>
              </div>
              <div>
                <p className="text-muted-foreground">Cashier</p>
                <p className="font-medium">{sale.employee_name || "-"}</p>
              </div>
              <div>
                <p className="text-muted-foreground">Payment Method</p>
                <div className="flex items-center gap-1 font-medium">
                  {paymentIcons[paymentMethod]}
                  <span className="capitalize">{paymentMethod}</span>
                </div>
              </div>
              <div>
                <p className="text-muted-foreground">Status</p>
                <span className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium capitalize ${getSalesStatusBadge(sale.status)}`}>
                  {sale.status}
                </span>
              </div>
            </div>

            <div>
              <p className="text-sm font-medium mb-2">Items</p>
              <div className="rounded-lg border divide-y">
                {sale.items?.map((item) => (
                  <div key={item.id} className="flex items-center justify-between p-3">
                    <div>
                      <p className="font-medium">{item.product_name}</p>
                      <p className="text-sm text-muted-foreground font-mono">
                        {formatSalesCurrency(item.unit_price)} x {item.quantity}
                      </p>
                    </div>
                    <p className="font-medium font-mono">{formatSalesCurrency(item.subtotal)}</p>
                  </div>
                ))}
              </div>
            </div>

            <div className="rounded-lg bg-muted p-4 space-y-2">
              <div className="flex justify-between text-sm">
                <span>Subtotal</span>
                <span className="font-mono">{formatSalesCurrency(sale.subtotal)}</span>
              </div>
              {Number.parseFloat(sale.discount_amount) > 0 ? (
                <div className="flex justify-between text-sm text-destructive">
                  <span>Discount</span>
                  <span className="font-mono">-{formatSalesCurrency(sale.discount_amount)}</span>
                </div>
              ) : null}
              <div className="flex justify-between font-bold pt-2 border-t">
                <span>Total</span>
                <span className="font-mono">{formatSalesCurrency(sale.total_amount)}</span>
              </div>
              {sale.payments && sale.payments.length > 0 ? (
                <>
                  <div className="flex justify-between text-sm">
                    <span>Amount Paid</span>
                    <span className="font-mono">{formatSalesCurrency(sale.amount_paid)}</span>
                  </div>
                  {Number.parseFloat(sale.change_amount) > 0 ? (
                    <div className="flex justify-between text-sm">
                      <span>Change</span>
                      <span className="font-mono">{formatSalesCurrency(sale.change_amount)}</span>
                    </div>
                  ) : null}
                </>
              ) : null}
            </div>

            {sale.status === "voided" ? (
              <div className="rounded-lg bg-destructive/10 p-4">
                <p className="text-sm font-medium text-destructive">Voided</p>
                <p className="text-sm text-muted-foreground">Reason: {sale.void_reason}</p>
              </div>
            ) : null}
          </div>
        ) : null}

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Close
          </Button>
          {sale?.status === "completed" && hasPermission(PERMISSIONS.SALES_VOID) ? (
            <Button variant="destructive" onClick={onVoidRequest}>
              <XCircle className="h-4 w-4 mr-2" />
              Void Sale
            </Button>
          ) : null}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

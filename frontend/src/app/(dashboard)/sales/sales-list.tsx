"use client";

import {
  Building2,
  ChevronLeft,
  ChevronRight,
  Eye,
  Loader2,
  Receipt,
} from "lucide-react";

import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import type { PaymentMethod, Sale } from "@/types";

import {
  formatSalesCurrency,
  formatSalesDate,
  getPrimarySalePaymentMethod,
  getSalesStatusBadge,
} from "./sales-helpers";

interface SalesListProps {
  sales: Sale[];
  isLoading: boolean;
  page: number;
  hasMore: boolean;
  paymentIcons: Record<PaymentMethod, React.ReactNode>;
  onPageChange: (page: number) => void;
  onViewSale: (sale: Sale) => void;
}

export function SalesList({
  sales,
  isLoading,
  page,
  hasMore,
  paymentIcons,
  onPageChange,
  onViewSale,
}: SalesListProps) {
  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  if (sales.length === 0) {
    return (
      <Card>
        <CardContent className="flex flex-col items-center justify-center py-12">
          <Receipt className="h-12 w-12 text-muted-foreground mb-4" />
          <p className="text-muted-foreground">No sales found</p>
        </CardContent>
      </Card>
    );
  }

  return (
    <>
      <div className="hidden lg:block overflow-x-auto">
        <table className="w-full min-w-max">
          <thead>
            <tr className="border-b text-left text-sm text-muted-foreground">
              <th className="pb-3 font-medium">Invoice</th>
              <th className="pb-3 font-medium">Date</th>
              <th className="pb-3 font-medium">Cashier</th>
              <th className="pb-3 font-medium">Items</th>
              <th className="pb-3 font-medium text-right">Total</th>
              <th className="pb-3 font-medium text-center">Status</th>
              <th className="pb-3 font-medium text-right">Actions</th>
            </tr>
          </thead>
          <tbody>
            {sales.map((sale) => (
              <tr key={sale.id} className="border-b last:border-0">
                <td className="py-3">
                  <p className="font-mono text-sm">{sale.invoice_no}</p>
                </td>
                <td className="py-3 text-sm text-muted-foreground">
                  {formatSalesDate(sale.created_at)}
                </td>
                <td className="py-3 text-sm">{sale.employee_name || "-"}</td>
                <td className="py-3 text-sm">{sale.item_count} items</td>
                <td className="py-3 text-right font-medium font-mono">
                  {formatSalesCurrency(sale.total_amount)}
                </td>
                <td className="py-3 text-center">
                  <span
                    className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium capitalize ${getSalesStatusBadge(sale.status)}`}
                  >
                    {sale.status}
                  </span>
                </td>
                <td className="py-3 text-right">
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => onViewSale(sale)}
                  >
                    <Eye className="h-4 w-4" />
                  </Button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <div className="lg:hidden space-y-4">
        {sales.map((sale) => {
          const paymentMethod = getPrimarySalePaymentMethod(sale);
          return (
            <div
              key={sale.id}
              className="border rounded-lg p-4 bg-card text-card-foreground shadow-sm"
            >
              <div className="flex justify-between items-start mb-3">
                <div>
                  <p className="font-mono font-bold text-sm">
                    {sale.invoice_no}
                  </p>
                  <p className="text-xs text-muted-foreground mt-1">
                    {formatSalesDate(sale.created_at)}
                  </p>
                </div>
                <span
                  className={`inline-flex items-center px-2 py-1 rounded-full text-[10px] uppercase font-bold tracking-wider ${getSalesStatusBadge(sale.status)}`}
                >
                  {sale.status}
                </span>
              </div>

              <div className="grid grid-cols-2 gap-y-2 text-sm mb-4">
                <div className="flex items-center gap-2 text-muted-foreground">
                  <Building2 className="h-3.5 w-3.5" />
                  <span>{sale.employee_name || "Unknown"}</span>
                </div>
                <div className="flex items-center gap-2 text-muted-foreground justify-self-end">
                  <Receipt className="h-3.5 w-3.5" />
                  <span>{sale.item_count} items</span>
                </div>
                <div className="flex items-center gap-2 text-muted-foreground col-span-2">
                  {paymentIcons[paymentMethod]}
                  <span className="capitalize">{paymentMethod}</span>
                </div>
              </div>

              <div className="flex items-center justify-between border-t pt-3">
                <div>
                  <p className="text-xs text-muted-foreground">Total</p>
                  <p className="text-lg font-bold font-mono">
                    {formatSalesCurrency(sale.total_amount)}
                  </p>
                </div>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => onViewSale(sale)}
                >
                  <Eye className="h-3.5 w-3.5 mr-1" />
                  View
                </Button>
              </div>
            </div>
          );
        })}
      </div>

      <div className="flex items-center justify-between mt-6 pt-4 border-t">
        <div className="flex gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => onPageChange(Math.max(1, page - 1))}
            disabled={page === 1}
          >
            <ChevronLeft className="h-4 w-4 mr-1" />
            Previous
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => onPageChange(page + 1)}
            disabled={!hasMore}
          >
            Next
            <ChevronRight className="h-4 w-4 ml-1" />
          </Button>
        </div>
      </div>
    </>
  );
}

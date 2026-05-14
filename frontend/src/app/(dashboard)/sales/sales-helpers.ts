import type { PaymentMethod, Sale } from "@/types";

export const SALES_PAYMENT_LABELS: Record<PaymentMethod, string> = {
  cash: "cash",
  card: "card",
  qris: "qris",
  transfer: "transfer",
};

export function getSalesStatusBadge(status: string): string {
  switch (status) {
    case "completed":
      return "bg-green-600 text-white dark:bg-green-600/90 dark:text-white";
    case "voided":
      return "bg-red-600 text-white dark:bg-red-600/90 dark:text-white";
    case "pending":
      return "bg-yellow-600 text-white dark:bg-yellow-600/90 dark:text-white";
    default:
      return "bg-gray-600 text-white dark:bg-gray-600/90 dark:text-white";
  }
}

export function getPrimarySalePaymentMethod(sale: Sale): PaymentMethod {
  if (sale.payments && sale.payments.length > 0) {
    return sale.payments[0].payment_method;
  }
  return "cash";
}

export function filterSalesBySearch(sales: Sale[], searchQuery: string): Sale[] {
  const normalizedSearch = searchQuery.trim().toLowerCase();
  if (!normalizedSearch) {
    return sales;
  }

  return sales.filter((sale) => {
    return (
      sale.invoice_no?.toLowerCase().includes(normalizedSearch) ||
      sale.employee_name?.toLowerCase().includes(normalizedSearch)
    );
  });
}

export function formatSalesCurrency(amount: string | number): string {
  const value = typeof amount === "string" ? Number.parseFloat(amount) : amount;
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    minimumFractionDigits: 0,
  }).format(value || 0);
}

export function formatSalesDate(date: string): string {
  return new Date(date).toLocaleString("id-ID", {
    day: "2-digit",
    month: "short",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

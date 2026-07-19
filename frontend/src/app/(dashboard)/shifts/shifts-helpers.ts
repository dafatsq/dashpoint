import type { CashDrawerOperation, Shift } from "@/types";

export interface ShiftQueryParams {
  limit: number;
  offset: number;
  from?: string;
  to?: string;
  opened_by_id?: string;
}

export function buildShiftQueryParams(input: {
  page: number;
  limit: number;
  dateRange: { start: string; end: string };
  selectedUser: string;
}): ShiftQueryParams {
  const params: ShiftQueryParams = {
    limit: input.limit,
    offset: (input.page - 1) * input.limit,
  };

  if (input.dateRange.start) params.from = input.dateRange.start;
  if (input.dateRange.end) params.to = input.dateRange.end;
  if (input.selectedUser && input.selectedUser !== "all") params.opened_by_id = input.selectedUser;

  return params;
}

export function formatShiftDateTime(dateString: string): string {
  return new Date(dateString).toLocaleString("id-ID", {
    day: "2-digit",
    month: "short",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export function formatShiftCurrency(amount: number): string {
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(amount);
}

export function getShiftSummary(shift: Shift) {
  const openingCash = Number.parseFloat(shift.opening_cash || "0");
  const closingCash = shift.closing_cash ? Number.parseFloat(shift.closing_cash) : null;
  const cashDifference = shift.cash_difference ? Number.parseFloat(shift.cash_difference) : null;
  return {
    isOpen: shift.status === "open",
    openingCash,
    closingCash,
    cashDifference,
  };
}

export function getShiftCashDifferenceTone(cashDifference: number | null): string {
  if (cashDifference === null || cashDifference === 0) return "text-muted-foreground";
  return cashDifference > 0 ? "text-green-600" : "text-red-600";
}

export function getShiftOperationLabel(operation: CashDrawerOperation): string {
  return operation.type === "pay_in" ? "+ PAY IN" : "- PAY OUT";
}

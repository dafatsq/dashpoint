"use client";

import { cn } from "@/lib/utils";

interface PriceProps {
  amount: number | string;
  className?: string;
}

export function Price({ amount, className }: PriceProps) {
  const value = typeof amount === "string" ? parseFloat(amount) : amount;
  const formatted = new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    minimumFractionDigits: 0,
  }).format(value || 0);

  return <span className={cn("font-mono", className)}>{formatted}</span>;
}

import { Calendar, Receipt, TrendingDown } from "lucide-react";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { ExpenseSummary } from "@/types";

interface ExpensesSummaryCardsProps {
  summary: ExpenseSummary | null;
  dateRange: { start: string; end: string };
  formatCurrency: (amount: string | number) => string;
  formatDate: (value: string) => string;
}

export function ExpensesSummaryCards({
  summary,
  dateRange,
  formatCurrency,
  formatDate,
}: ExpensesSummaryCardsProps) {
  return (
    <div className="mb-6 grid grid-cols-1 gap-4 md:grid-cols-3">
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Total Expenses</CardTitle>
          <TrendingDown className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold text-destructive">{summary ? formatCurrency(summary.total_amount) : "Rp 0"}</div>
          <p className="text-xs text-muted-foreground">{summary?.expense_count || 0} transactions</p>
        </CardContent>
      </Card>

      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Top Category</CardTitle>
          <Receipt className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">{summary?.by_category?.[0]?.category_name || "-"}</div>
          <p className="text-xs text-muted-foreground">
            {summary?.by_category?.[0] ? formatCurrency(summary.by_category[0].total_amount) : "No expenses yet"}
          </p>
        </CardContent>
      </Card>

      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Period</CardTitle>
          <Calendar className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-lg font-medium">
            {dateRange.start && dateRange.end
              ? `${formatDate(dateRange.start)} - ${formatDate(dateRange.end)}`
              : "No period selected"}
          </div>
          <p className="text-xs text-muted-foreground">{summary?.by_category?.length || 0} categories</p>
        </CardContent>
      </Card>
    </div>
  );
}

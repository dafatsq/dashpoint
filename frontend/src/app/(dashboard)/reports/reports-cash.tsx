'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { ArrowDownRight, ArrowUpRight, Banknote, Minus } from 'lucide-react';
import type { CashReport } from '@/types';

import { formatCurrency } from './reports-helpers';
import { ReportsEmptyState, ReportsLoadingState } from './reports-feedback';

interface ReportsCashProps {
  isLoading: boolean;
  cashReport: CashReport | null;
}

export function ReportsCash({ isLoading, cashReport }: ReportsCashProps) {
  if (isLoading) {
    return <ReportsLoadingState />;
  }

  if (!cashReport) {
    return (
      <ReportsEmptyState
        icon={Banknote}
        message="No cash data available"
        details="Shifts must be closed to appear here"
      />
    );
  }

  const difference = parseFloat(cashReport.difference) || 0;

  return (
    <div className="space-y-6">
      <div className="grid gap-4 grid-cols-1 sm:grid-cols-3 lg:grid-cols-5">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Opening Cash</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold font-mono">{formatCurrency(cashReport.opening_cash)}</div>
            <p className="text-xs text-muted-foreground mt-1">From {cashReport.shift_count} shift(s)</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Cash Sales</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-600 font-mono">+{formatCurrency(cashReport.cash_sales)}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Voided Cash Sales</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-red-600 font-mono">-{formatCurrency(cashReport.cash_voided_sales)}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Pay In</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-blue-600 font-mono">+{formatCurrency(cashReport.pay_in_total ?? '0')}</div>
            <p className="text-xs text-muted-foreground mt-1">Cash added to drawer</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Pay Out</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-orange-600 font-mono">-{formatCurrency(cashReport.pay_out_total ?? '0')}</div>
            <p className="text-xs text-muted-foreground mt-1">Cash removed from drawer</p>
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-4 grid-cols-1 sm:grid-cols-2">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Expected Cash</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold font-mono">{formatCurrency(cashReport.expected_cash)}</div>
            <p className="text-xs text-muted-foreground mt-1">
              Opening + Sales + Pay In - Pay Out
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Actual Cash</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold font-mono">{formatCurrency(cashReport.actual_cash)}</div>
            <p className="text-xs text-muted-foreground mt-1">Closing cash from shifts</p>
          </CardContent>
        </Card>
      </div>

      <Card
        className={`border-l-4 shadow-sm ${
          difference === 0 ? 'border-l-green-500' : difference > 0 ? 'border-l-yellow-500' : 'border-l-red-500'
        }`}
      >
        <CardHeader>
          <CardTitle className="text-base">Cash Difference</CardTitle>
          <CardDescription>Actual minus Expected</CardDescription>
        </CardHeader>
        <CardContent>
          <div
            className={`text-3xl font-bold flex items-center gap-2 ${
              difference >= 0 ? 'text-green-600' : 'text-red-600'
            }`}
          >
            {difference > 0 ? <ArrowUpRight className="h-8 w-8" /> : null}
            {difference < 0 ? <ArrowDownRight className="h-8 w-8" /> : null}
            {difference === 0 ? <Minus className="h-8 w-8" /> : null}
            <span className="font-mono">{formatCurrency(Math.abs(difference))}</span>
            {difference > 0 ? ' over' : null}
            {difference < 0 ? ' short' : null}
            {difference === 0 ? ' (balanced)' : null}
          </div>
          <p className="text-xs text-muted-foreground mt-2">
            Expected: <span className="font-mono">{formatCurrency(cashReport.expected_cash)}</span> • Actual: <span className="font-mono">{formatCurrency(cashReport.actual_cash)}</span>
          </p>
        </CardContent>
      </Card>
    </div>
  );
}

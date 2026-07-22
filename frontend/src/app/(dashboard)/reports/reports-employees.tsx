'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Users } from 'lucide-react';
import type { EmployeeSales } from '@/types';
import { DataSortSelect } from '@/components/shared/data-sort-select';
import { useMemo, useState } from 'react';

import { formatCurrency, formatNumber } from './reports-helpers';
import { ReportsEmptyState, ReportsLoadingState } from './reports-feedback';
import type { ReportDateRange } from './reports-types';

interface ReportsEmployeesProps {
  isLoading: boolean;
  employeeSales: EmployeeSales[];
  dateRange: ReportDateRange;
}

export function ReportsEmployees({ isLoading, employeeSales, dateRange }: ReportsEmployeesProps) {
  const [sort, setSort] = useState('sales_desc');
  const sortedEmployeeSales = useMemo(
    () => [...employeeSales].sort((left, right) => {
      const direction = sort.endsWith('_desc') ? -1 : 1;
      const sortBy = sort.replace(/_(asc|desc)$/, '');
      if (sortBy === 'name') return left.employee_name.localeCompare(right.employee_name) * direction;
      if (sortBy === 'transactions') return (left.transaction_count - right.transaction_count) * direction;
      if (sortBy === 'items') return (left.item_count - right.item_count) * direction;
      return (Number.parseFloat(left.total_sales) - Number.parseFloat(right.total_sales)) * direction;
    }),
    [employeeSales, sort],
  );

  if (isLoading) {
    return <ReportsLoadingState />;
  }

  if (employeeSales.length === 0) {
    return <ReportsEmptyState icon={Users} message="No employee sales data available" />;
  }

  return (
    <div className="space-y-6">
      <Card className="hidden lg:block">
        <CardHeader>
          <div className="flex flex-wrap items-center justify-between gap-4">
            <div>
              <CardTitle>Sales by Employee</CardTitle>
              <CardDescription>
                {dateRange.start} to {dateRange.end}
              </CardDescription>
            </div>
            <DataSortSelect value={sort} options={[{ value: 'sales_desc', label: 'Sales (high-low)' }, { value: 'sales_asc', label: 'Sales (low-high)' }, { value: 'transactions_desc', label: 'Transactions (high-low)' }, { value: 'items_desc', label: 'Items sold (high-low)' }, { value: 'name_asc', label: 'Employee (A-Z)' }]} onChange={setSort} />
          </div>
        </CardHeader>
        <CardContent>
          <div className="overflow-x-auto">
            <table className="w-full min-w-max">
              <thead>
                <tr className="border-b text-left text-sm text-muted-foreground">
                  <th className="pb-3 font-medium">#</th>
                  <th className="pb-3 font-medium">Employee</th>
                  <th className="pb-3 font-medium text-right">Transactions</th>
                  <th className="pb-3 font-medium text-right">Items Sold</th>
                  <th className="pb-3 font-medium text-right">Total Sales</th>
                  <th className="pb-3 font-medium text-right">Avg / Transaction</th>
                </tr>
              </thead>
              <tbody>
                {sortedEmployeeSales.map((employee, index) => (
                  <tr key={employee.employee_id} className="border-b last:border-0 hover:bg-muted/50">
                    <td className="py-3">
                      <span
                        className={`flex h-7 w-7 items-center justify-center rounded-full text-xs font-bold ${
                          index < 3 ? 'bg-primary text-primary-foreground' : 'bg-muted'
                        }`}
                      >
                        {index + 1}
                      </span>
                    </td>
                    <td className="py-3 font-medium">{employee.employee_name}</td>
                    <td className="py-3 text-right">{formatNumber(employee.transaction_count)}</td>
                    <td className="py-3 text-right">{formatNumber(employee.item_count)}</td>
                    <td className="py-3 text-right font-bold font-mono">{formatCurrency(employee.total_sales)}</td>
                    <td className="py-3 text-right font-mono">{formatCurrency(employee.avg_transaction)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </CardContent>
      </Card>

      <div className="lg:hidden space-y-4">
        <div className="flex flex-wrap items-center justify-between gap-4">
          <h3 className="font-semibold text-lg">Sales by Employee</h3>
          <DataSortSelect value={sort} options={[{ value: 'sales_desc', label: 'Sales (high-low)' }, { value: 'sales_asc', label: 'Sales (low-high)' }, { value: 'transactions_desc', label: 'Transactions (high-low)' }, { value: 'items_desc', label: 'Items sold (high-low)' }, { value: 'name_asc', label: 'Employee (A-Z)' }]} onChange={setSort} />
        </div>
        {sortedEmployeeSales.map((employee, index) => (
          <Card key={employee.employee_id}>
            <CardContent className="p-4 space-y-3">
              <div className="flex items-center gap-3 border-b pb-3">
                <span
                  className={`flex h-8 w-8 items-center justify-center rounded-full text-sm font-bold ${
                    index < 3 ? 'bg-primary text-primary-foreground' : 'bg-muted'
                  }`}
                >
                  {index + 1}
                </span>
                <p className="font-bold">{employee.employee_name}</p>
              </div>
              <div className="grid grid-cols-2 gap-3 text-sm">
                <div>
                  <span className="text-xs text-muted-foreground block">Total Sales</span>
                  <span className="font-bold font-mono">{formatCurrency(employee.total_sales)}</span>
                </div>
                <div className="text-right">
                  <span className="text-xs text-muted-foreground block">Transactions</span>
                  <span className="font-medium">{formatNumber(employee.transaction_count)}</span>
                </div>
                <div>
                  <span className="text-xs text-muted-foreground block">Items Sold</span>
                  <span className="font-medium">{formatNumber(employee.item_count)}</span>
                </div>
                <div className="text-right">
                  <span className="text-xs text-muted-foreground block">Avg / Txn</span>
                  <span className="font-medium font-mono">{formatCurrency(employee.avg_transaction)}</span>
                </div>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}

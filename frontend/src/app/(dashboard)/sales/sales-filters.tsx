"use client";

import { Search } from "lucide-react";

import { DateRangePicker } from "@/components/ui/date-range-picker";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { FilterCard } from "@/components/shared/filter-card";
import { DataSortSelect } from "@/components/shared/data-sort-select";

export const SALES_SORT_OPTIONS = [
  { value: "date_desc", label: "Date (newest)" },
  { value: "date_asc", label: "Date (oldest)" },
  { value: "invoice_no_asc", label: "Invoice (A-Z)" },
  { value: "total_desc", label: "Total (high-low)" },
  { value: "total_asc", label: "Total (low-high)" },
  { value: "employee_asc", label: "Cashier (A-Z)" },
  { value: "status_asc", label: "Status (A-Z)" },
] as const;

interface SalesFiltersProps {
  searchQuery: string;
  statusFilter: string;
  employeeFilter: string;
  dateRange: { start: string; end: string };
  employees: { id: string; name: string; role_name?: string }[];
  sort: string;
  onSearchChange: (value: string) => void;
  onStatusChange: (value: string) => void;
  onEmployeeChange: (value: string) => void;
  onDateRangeChange: (value: { start: string; end: string }) => void;
  onSortChange: (value: string) => void;
}

export function SalesFilters(props: SalesFiltersProps) {
  const {
    searchQuery,
    statusFilter,
    employeeFilter,
    dateRange,
    employees,
    sort,
    onSearchChange,
    onStatusChange,
    onEmployeeChange,
    onDateRangeChange,
    onSortChange,
  } = props;

  return (
    <FilterCard>
      <div className="flex flex-wrap items-center justify-center gap-4">
        <div className="relative w-full flex-1 min-w-[200px] sm:min-w-[220px]">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search invoice..."
            value={searchQuery}
            onChange={(event) => onSearchChange(event.target.value)}
            className="pl-9 w-full"
          />
        </div>
        <DateRangePicker
          value={dateRange}
          onChange={onDateRangeChange}
          placeholder="Select date range"
          className="w-full flex-1 min-w-[200px] sm:min-w-[220px]"
        />
        <div className="w-full flex-1 min-w-[200px] sm:min-w-[220px]">
          <Select value={statusFilter} onValueChange={onStatusChange}>
            <SelectTrigger className="w-full">
              <SelectValue placeholder="All Statuses" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Statuses</SelectItem>
              <SelectItem value="completed">Completed</SelectItem>
              <SelectItem value="voided">Voided</SelectItem>
            </SelectContent>
          </Select>
        </div>
        <div className="w-full flex-1 min-w-[200px] sm:min-w-[220px]">
          <Select value={employeeFilter} onValueChange={onEmployeeChange}>
            <SelectTrigger className="w-full">
              <SelectValue placeholder="All Employees" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Employees</SelectItem>
              {employees.map((employee) => (
                <SelectItem key={employee.id} value={employee.id}>
                  {employee.name}
                  {employee.role_name ? ` (${employee.role_name})` : ""}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        <DataSortSelect value={sort} options={SALES_SORT_OPTIONS} onChange={onSortChange} className="w-full flex-1 min-w-[200px] sm:min-w-[220px]" />
      </div>
    </FilterCard>
  );
}

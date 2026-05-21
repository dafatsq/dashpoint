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

interface SalesFiltersProps {
  searchQuery: string;
  statusFilter: string;
  employeeFilter: string;
  dateRange: { start: string; end: string };
  employees: { id: string; name: string; role_name?: string }[];
  onSearchChange: (value: string) => void;
  onStatusChange: (value: string) => void;
  onEmployeeChange: (value: string) => void;
  onDateRangeChange: (value: { start: string; end: string }) => void;
}

export function SalesFilters(props: SalesFiltersProps) {
  const {
    searchQuery,
    statusFilter,
    employeeFilter,
    dateRange,
    employees,
    onSearchChange,
    onStatusChange,
    onEmployeeChange,
    onDateRangeChange,
  } = props;

  return (
    <FilterCard title="Filters">
      <div className="flex flex-col md:grid md:grid-cols-4 gap-4">
        <div className="relative">
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
          className="w-full"
        />
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
    </FilterCard>
  );
}

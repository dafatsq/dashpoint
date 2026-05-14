'use client';

import { Search } from "lucide-react";

import { DateRangePicker } from "@/components/ui/date-range-picker";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import type { User } from "@/types";

interface SalesFiltersProps {
  searchQuery: string;
  statusFilter: string;
  employeeFilter: string;
  dateRange: { start: string; end: string };
  employees: User[];
  onSearchChange: (value: string) => void;
  onStatusChange: (value: string) => void;
  onEmployeeChange: (value: string) => void;
  onDateRangeChange: (value: { start: string; end: string }) => void;
}

export function SalesFilters(props: SalesFiltersProps) {
  return (
    <Card className="mb-6">
      <CardHeader>
        <CardTitle className="text-base">Filters</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="flex flex-col md:grid md:grid-cols-4 gap-4">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Search invoice..."
              value={props.searchQuery}
              onChange={(event) => props.onSearchChange(event.target.value)}
              className="pl-9 w-full"
            />
          </div>
          <DateRangePicker value={props.dateRange} onChange={props.onDateRangeChange} placeholder="Select date range" className="w-full" />
          <Select value={props.statusFilter} onValueChange={props.onStatusChange}>
            <SelectTrigger className="w-full">
              <SelectValue placeholder="All Statuses" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Statuses</SelectItem>
              <SelectItem value="completed">Completed</SelectItem>
              <SelectItem value="voided">Voided</SelectItem>
            </SelectContent>
          </Select>
          <Select value={props.employeeFilter} onValueChange={props.onEmployeeChange}>
            <SelectTrigger className="w-full">
              <SelectValue placeholder="All Cashiers" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Cashiers</SelectItem>
              {props.employees.map((employee) => (
                <SelectItem key={employee.id} value={employee.id}>
                  {employee.name} ({employee.role_name})
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </CardContent>
    </Card>
  );
}

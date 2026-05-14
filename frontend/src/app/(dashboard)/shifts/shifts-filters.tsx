'use client';

import { DateRangePicker } from "@/components/ui/date-range-picker";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import type { User } from "@/types";

interface ShiftsFiltersProps {
  dateRange: { start: string; end: string };
  selectedUser: string;
  users: User[];
  onDateRangeChange: (value: { start: string; end: string }) => void;
  onSelectedUserChange: (value: string) => void;
}

export function ShiftsFilters({ dateRange, selectedUser, users, onDateRangeChange, onSelectedUserChange }: ShiftsFiltersProps) {
  return (
    <Card className="mb-6">
      <CardHeader>
        <CardTitle className="text-base">Filters</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="space-y-1.5">
            <label className="text-xs font-medium text-muted-foreground">Date Range</label>
            <DateRangePicker value={dateRange} onChange={onDateRangeChange} placeholder="Filter by date..." className="w-full" />
          </div>
          <div className="space-y-1.5">
            <label className="text-xs font-medium text-muted-foreground">Employee</label>
            <Select value={selectedUser} onValueChange={onSelectedUserChange}>
              <SelectTrigger className="w-full">
                <SelectValue placeholder="All Employees" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Employees</SelectItem>
                {users.map((user) => (
                  <SelectItem key={user.id} value={user.id}>
                    {user.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

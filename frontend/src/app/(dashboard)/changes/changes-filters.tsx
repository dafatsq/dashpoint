"use client";

import { FilterCard } from "@/components/shared/filter-card";
import { DateRangePicker } from "@/components/ui/date-range-picker";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

import type { ActivityDateRange } from "../activity-helpers";

interface ChangesFiltersProps {
  dateRange: ActivityDateRange;
  selectedUser: string;
  users: { id: string; name: string }[];
  onDateRangeChange: (value: ActivityDateRange) => void;
  onSelectedUserChange: (value: string) => void;
}

export function ChangesFilters({
  dateRange,
  selectedUser,
  users,
  onDateRangeChange,
  onSelectedUserChange,
}: ChangesFiltersProps) {
  return (
    <FilterCard>
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <DateRangePicker
          value={dateRange}
          onChange={onDateRangeChange}
          placeholder="Filter by date..."
          className="w-full"
        />
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
    </FilterCard>
  );
}

"use client";

import { DateRangePicker } from "@/components/ui/date-range-picker";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { FilterCard } from "@/components/shared/filter-card";
import { DataSortSelect } from "@/components/shared/data-sort-select";

export const SHIFT_SORT_OPTIONS = [
  { value: "date_desc", label: "Date (newest)" },
  { value: "date_asc", label: "Date (oldest)" },
  { value: "opened_by_asc", label: "Opened by (A-Z)" },
  { value: "status_asc", label: "Status (A-Z)" },
  { value: "total_sales_desc", label: "Sales (high-low)" },
  { value: "total_sales_asc", label: "Sales (low-high)" },
] as const;

interface ShiftsFiltersProps {
  dateRange: { start: string; end: string };
  selectedUser: string;
  users: { id: string; name: string }[];
  sort: string;
  onDateRangeChange: (value: { start: string; end: string }) => void;
  onSelectedUserChange: (value: string) => void;
  onSortChange: (value: string) => void;
}

export function ShiftsFilters({
  dateRange,
  selectedUser,
  users,
  sort,
  onDateRangeChange,
  onSelectedUserChange,
  onSortChange,
}: ShiftsFiltersProps) {
  return (
    <FilterCard>
      <div className="flex flex-wrap items-center justify-center gap-4">
        <DateRangePicker
          value={dateRange}
          onChange={onDateRangeChange}
          placeholder="Filter by date..."
          className="w-full flex-1 min-w-[200px] sm:min-w-[220px]"
        />
        <div className="w-full flex-1 min-w-[200px] sm:min-w-[220px]">
          <Select value={selectedUser} onValueChange={onSelectedUserChange}>
            <SelectTrigger className="w-full">
              <SelectValue placeholder="All Users" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Users</SelectItem>
              {users.map((user) => (
                <SelectItem key={user.id} value={user.id}>
                  {user.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        <DataSortSelect value={sort} options={SHIFT_SORT_OPTIONS} onChange={onSortChange} className="w-full flex-1 min-w-[200px] sm:min-w-[220px]" />
      </div>
    </FilterCard>
  );
}

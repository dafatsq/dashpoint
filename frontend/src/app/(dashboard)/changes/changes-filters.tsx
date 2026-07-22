"use client";

import { DataSortSelect } from "@/components/shared/data-sort-select";
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

export const CHANGES_SORT_OPTIONS = [
  { value: "date_desc", label: "Date (newest)" },
  { value: "date_asc", label: "Date (oldest)" },
  { value: "user_asc", label: "User (A-Z)" },
  { value: "action_asc", label: "Action (A-Z)" },
  { value: "entity_asc", label: "Entity (A-Z)" },
] as const;

interface ChangesFiltersProps {
  dateRange: ActivityDateRange;
  selectedUser: string;
  users: { id: string; name: string }[];
  sort: string;
  onDateRangeChange: (value: ActivityDateRange) => void;
  onSelectedUserChange: (value: string) => void;
  onSortChange: (value: string) => void;
}

export function ChangesFilters({
  dateRange,
  selectedUser,
  users,
  sort,
  onDateRangeChange,
  onSelectedUserChange,
  onSortChange,
}: ChangesFiltersProps) {
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
        <DataSortSelect value={sort} options={CHANGES_SORT_OPTIONS} onChange={onSortChange} className="w-full flex-1 min-w-[200px] sm:min-w-[220px]" />
      </div>
    </FilterCard>
  );
}

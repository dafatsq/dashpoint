'use client';

import { Search } from "lucide-react";

import { DateRangePicker } from "@/components/ui/date-range-picker";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { FilterCard } from "@/components/shared/filter-card";
import { DataSortSelect } from "@/components/shared/data-sort-select";

export const AUDIT_SORT_OPTIONS = [
  { value: "date_desc", label: "Date (newest)" },
  { value: "date_asc", label: "Date (oldest)" },
  { value: "user_asc", label: "User (A-Z)" },
  { value: "action_asc", label: "Action (A-Z)" },
  { value: "entity_asc", label: "Entity (A-Z)" },
] as const;

import type { ActivityDateRange } from "../activity-helpers";

interface AuditFiltersProps {
  searchQuery: string;
  selectedAction: string;
  selectedEntity: string;
  dateRange: ActivityDateRange;
  sort: string;
  onSearchChange: (value: string) => void;
  onActionChange: (value: string) => void;
  onEntityChange: (value: string) => void;
  onDateRangeChange: (value: ActivityDateRange) => void;
  onSortChange: (value: string) => void;
}

export function AuditFilters({
  searchQuery,
  selectedAction,
  selectedEntity,
  dateRange,
  sort,
  onSearchChange,
  onActionChange,
  onEntityChange,
  onDateRangeChange,
  onSortChange,
}: AuditFiltersProps) {
  return (
    <FilterCard>
      <div className="flex flex-wrap items-center justify-center gap-4">
        <div className="relative w-full flex-1 min-w-[200px] sm:min-w-[220px]">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Search logs..."
            value={searchQuery}
            onChange={(event) => onSearchChange(event.target.value)}
            className="w-full pl-9"
          />
        </div>
        <DateRangePicker
          value={dateRange}
          onChange={onDateRangeChange}
          placeholder="Filter by date..."
          className="w-full flex-1 min-w-[200px] sm:min-w-[220px]"
        />
        <div className="w-full flex-1 min-w-[200px] sm:min-w-[220px]">
          <Select value={selectedAction} onValueChange={onActionChange}>
            <SelectTrigger className="w-full">
              <SelectValue placeholder="All Actions" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Actions</SelectItem>
              <SelectItem value="login">Login</SelectItem>
              <SelectItem value="login_failed">Login Failed</SelectItem>
              <SelectItem value="create">Create</SelectItem>
              <SelectItem value="update">Update</SelectItem>
              <SelectItem value="archive">Archive</SelectItem>
              <SelectItem value="restore">Restore</SelectItem>
              <SelectItem value="delete">Delete</SelectItem>
              <SelectItem value="void">Void</SelectItem>
            </SelectContent>
          </Select>
        </div>
        <div className="w-full flex-1 min-w-[200px] sm:min-w-[220px]">
          <Select value={selectedEntity} onValueChange={onEntityChange}>
            <SelectTrigger className="w-full">
              <SelectValue placeholder="All Entities" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Entities</SelectItem>
              <SelectItem value="auth">Authentication</SelectItem>
              <SelectItem value="user">User</SelectItem>
              <SelectItem value="product">Product</SelectItem>
              <SelectItem value="category">Category</SelectItem>
              <SelectItem value="sale">Sale</SelectItem>
              <SelectItem value="shift">Shift</SelectItem>
              <SelectItem value="inventory">Inventory</SelectItem>
              <SelectItem value="expense">Expense</SelectItem>
            </SelectContent>
          </Select>
        </div>
        <DataSortSelect value={sort} options={AUDIT_SORT_OPTIONS} onChange={onSortChange} className="w-full flex-1 min-w-[200px] sm:min-w-[220px]" />
      </div>
    </FilterCard>
  );
}

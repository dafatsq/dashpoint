import { Plus, Search } from "lucide-react";

import { Button } from "@/components/ui/button";
import { DateRangePicker } from "@/components/ui/date-range-picker";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { ExpenseCategory } from "@/types";
import { FilterCard } from "@/components/shared/filter-card";
import { DataSortSelect } from "@/components/shared/data-sort-select";

export const EXPENSE_SORT_OPTIONS = [
  { value: "date_desc", label: "Date (newest)" },
  { value: "date_asc", label: "Date (oldest)" },
  { value: "amount_desc", label: "Amount (high-low)" },
  { value: "amount_asc", label: "Amount (low-high)" },
  { value: "category_asc", label: "Category (A-Z)" },
  { value: "description_asc", label: "Description (A-Z)" },
] as const;

interface ExpensesToolbarProps {
  categories: ExpenseCategory[];
  searchQuery: string;
  selectedCategory: string;
  dateRange: { start: string; end: string };
  sort: string;
  canCreateExpense: boolean;
  onCreate: () => void;
  onSearchChange: (value: string) => void;
  onCategoryChange: (value: string) => void;
  onDateRangeChange: (value: { start: string; end: string }) => void;
  onSortChange: (value: string) => void;
}

export function ExpensesToolbar({
  categories,
  searchQuery,
  selectedCategory,
  dateRange,
  sort,
  canCreateExpense,
  onCreate,
  onSearchChange,
  onCategoryChange,
  onDateRangeChange,
  onSortChange,
}: ExpensesToolbarProps) {
  return (
    <FilterCard>
      <div className="flex flex-wrap items-center justify-center gap-4">
        <div className="relative w-full flex-1 min-w-[200px] sm:min-w-[220px]">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Search expenses..."
            value={searchQuery}
            onChange={(event) => onSearchChange(event.target.value)}
            className="w-full pl-9"
          />
        </div>
        <div className="w-full flex-1 min-w-[200px] sm:min-w-[220px]">
          <Select value={selectedCategory} onValueChange={onCategoryChange}>
            <SelectTrigger className="w-full">
              <SelectValue placeholder="All Categories" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Categories</SelectItem>
              {categories.map((category) => (
                <SelectItem key={category.id} value={category.id}>
                  {category.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        <DateRangePicker
          value={dateRange}
          onChange={onDateRangeChange}
          placeholder="Select date range"
          className="w-full flex-1 min-w-[200px] sm:min-w-[220px]"
        />
        <DataSortSelect value={sort} options={EXPENSE_SORT_OPTIONS} onChange={onSortChange} className="w-full flex-1 min-w-[200px] sm:min-w-[220px]" />
        {canCreateExpense && (
          <Button onClick={onCreate} className="w-full flex-1 min-w-[200px] sm:min-w-[220px]">
            <Plus className="mr-2 h-4 w-4" />
            Add Expense
          </Button>
        )}
      </div>
    </FilterCard>
  );
}

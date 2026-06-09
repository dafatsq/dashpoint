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

interface ExpensesToolbarProps {
  categories: ExpenseCategory[];
  searchQuery: string;
  selectedCategory: string;
  dateRange: { start: string; end: string };
  canCreateExpense: boolean;
  onCreate: () => void;
  onSearchChange: (value: string) => void;
  onCategoryChange: (value: string) => void;
  onDateRangeChange: (value: { start: string; end: string }) => void;
}

export function ExpensesToolbar({
  categories,
  searchQuery,
  selectedCategory,
  dateRange,
  canCreateExpense,
  onCreate,
  onSearchChange,
  onCategoryChange,
  onDateRangeChange,
}: ExpensesToolbarProps) {
  return (
    <FilterCard>
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="relative w-full">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Search expenses..."
            value={searchQuery}
            onChange={(event) => onSearchChange(event.target.value)}
            className="w-full pl-9"
          />
        </div>
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
        <DateRangePicker
          value={dateRange}
          onChange={onDateRangeChange}
          placeholder="Select date range"
          className="w-full"
        />
        {canCreateExpense && (
          <Button onClick={onCreate} className="w-full">
            <Plus className="mr-2 h-4 w-4" />
            Add Expense
          </Button>
        )}
      </div>
    </FilterCard>
  );
}

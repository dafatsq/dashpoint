import { Plus, Search } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
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
    <Card className="mb-6">
      <CardContent className="p-6">
        <div className="flex flex-col gap-4 sm:flex-row">
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              placeholder="Search expenses..."
              value={searchQuery}
              onChange={(event) => onSearchChange(event.target.value)}
              className="w-full pl-9"
            />
          </div>
          <Select value={selectedCategory} onValueChange={onCategoryChange}>
            <SelectTrigger className="w-full sm:w-48">
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
            className="w-full sm:w-[280px]"
          />
          {canCreateExpense && (
            <Button onClick={onCreate} className="w-full sm:w-auto">
              <Plus className="mr-2 h-4 w-4" />
              Add Expense
            </Button>
          )}
        </div>
      </CardContent>
    </Card>
  );
}

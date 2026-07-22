'use client';

import { Search } from "lucide-react";

import { FilterCard } from "@/components/shared/filter-card";
import { DataSortSelect } from "@/components/shared/data-sort-select";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import type { Category } from "@/types";

export const INVENTORY_SORT_OPTIONS = [
  { value: "name_asc", label: "Product (A-Z)" },
  { value: "name_desc", label: "Product (Z-A)" },
  { value: "stock_asc", label: "Stock (low-high)" },
  { value: "stock_desc", label: "Stock (high-low)" },
  { value: "price_asc", label: "Price (low-high)" },
  { value: "price_desc", label: "Price (high-low)" },
] as const;

interface InventoryControlsProps {
  searchQuery: string;
  activeTab: "all" | "low-stock";
  lowStockCount: number;
  categories: Category[];
  selectedCategory: string;
  sort: string;
  onSearchChange: (value: string) => void;
  onTabChange: (value: "all" | "low-stock") => void;
  onCategoryChange: (value: string) => void;
  onSortChange: (value: string) => void;
}

export function InventoryControls({
  searchQuery,
  activeTab,
  lowStockCount,
  categories,
  selectedCategory,
  sort,
  onSearchChange,
  onTabChange,
  onCategoryChange,
  onSortChange,
}: InventoryControlsProps) {
  return (
    <FilterCard>
      <div className="flex flex-wrap items-center justify-center gap-4">
        <div className="relative w-full flex-1 min-w-[200px] sm:min-w-[220px]">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Search products..."
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

        <DataSortSelect value={sort} options={INVENTORY_SORT_OPTIONS} onChange={onSortChange} className="w-full flex-1 min-w-[200px] sm:min-w-[220px]" />

        <div className="grid grid-cols-2 gap-2 w-full flex-1 min-w-[200px] sm:min-w-[220px]">
          <Button variant={activeTab === "all" ? "default" : "outline"} onClick={() => onTabChange("all")} className="w-full">
            All Products
          </Button>
          <Button
            variant={activeTab === "low-stock" ? "default" : "outline"}
            onClick={() => onTabChange("low-stock")}
            className="w-full text-xs sm:text-sm px-2 truncate"
          >
            Low Stock ({lowStockCount})
          </Button>
        </div>
      </div>
    </FilterCard>
  );
}

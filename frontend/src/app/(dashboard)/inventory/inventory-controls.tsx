'use client';

import { Search } from "lucide-react";

import { FilterCard } from "@/components/shared/filter-card";
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

interface InventoryControlsProps {
  searchQuery: string;
  activeTab: "all" | "low-stock";
  lowStockCount: number;
  categories: Category[];
  selectedCategory: string;
  onSearchChange: (value: string) => void;
  onTabChange: (value: "all" | "low-stock") => void;
  onCategoryChange: (value: string) => void;
}

export function InventoryControls({
  searchQuery,
  activeTab,
  lowStockCount,
  categories,
  selectedCategory,
  onSearchChange,
  onTabChange,
  onCategoryChange,
}: InventoryControlsProps) {
  return (
    <FilterCard>
      <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4">
        <div className="relative w-full">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Search products..."
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

        <div className="grid grid-cols-2 gap-2 w-full">
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

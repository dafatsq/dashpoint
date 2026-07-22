import { Plus, Search } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Category } from "@/types";
import { FilterCard } from "@/components/shared/filter-card";
import { ActiveArchivedToggle } from "@/components/shared/active-archived-toggle";
import { DataSortSelect } from "@/components/shared/data-sort-select";

export const PRODUCT_SORT_OPTIONS = [
  { value: "name_asc", label: "Name (A-Z)" },
  { value: "name_desc", label: "Name (Z-A)" },
  { value: "category_asc", label: "Category (A-Z)" },
  { value: "price_asc", label: "Price (low-high)" },
  { value: "price_desc", label: "Price (high-low)" },
  { value: "stock_asc", label: "Stock (low-high)" },
  { value: "stock_desc", label: "Stock (high-low)" },
] as const;

interface ProductsControlsProps {
  canCreate: boolean;
  categories: Category[];
  searchQuery: string;
  selectedCategory: string;
  sort: string;
  viewMode: "active" | "archived";
  onCreate: () => void;
  onSearchChange: (value: string) => void;
  onCategoryChange: (value: string) => void;
  onSortChange: (value: string) => void;
  onViewModeChange: (value: "active" | "archived") => void;
}

export function ProductsControls({
  canCreate,
  categories,
  searchQuery,
  selectedCategory,
  sort,
  viewMode,
  onCreate,
  onSearchChange,
  onCategoryChange,
  onSortChange,
  onViewModeChange,
}: ProductsControlsProps) {
  return (
    <>
      <ActiveArchivedToggle value={viewMode} onChange={onViewModeChange} className="mb-4" />
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
          <DataSortSelect value={sort} options={PRODUCT_SORT_OPTIONS} onChange={onSortChange} className="w-full flex-1 min-w-[200px] sm:min-w-[220px]" />
          {canCreate && viewMode === "active" && (
            <Button onClick={onCreate} className="w-full flex-1 min-w-[200px] sm:min-w-[220px]">
              <Plus className="mr-2 h-4 w-4" />
              Add Product
            </Button>
          )}
        </div>
      </FilterCard>
    </>
  );
}

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

interface ProductsControlsProps {
  canCreate: boolean;
  categories: Category[];
  searchQuery: string;
  selectedCategory: string;
  viewMode: "active" | "archived";
  onCreate: () => void;
  onSearchChange: (value: string) => void;
  onCategoryChange: (value: string) => void;
  onViewModeChange: (value: "active" | "archived") => void;
}

export function ProductsControls({
  canCreate,
  categories,
  searchQuery,
  selectedCategory,
  viewMode,
  onCreate,
  onSearchChange,
  onCategoryChange,
  onViewModeChange,
}: ProductsControlsProps) {
  return (
    <>
      <ActiveArchivedToggle value={viewMode} onChange={onViewModeChange} className="mb-4" />
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
          {canCreate && viewMode === "active" && (
            <Button onClick={onCreate} className="w-full">
              <Plus className="mr-2 h-4 w-4" />
              Add Product
            </Button>
          )}
        </div>
      </FilterCard>
    </>
  );
}

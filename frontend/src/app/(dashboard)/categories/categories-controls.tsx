'use client';

import { Boxes, Plus, Search, Wallet } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { FilterCard } from "@/components/shared/filter-card";
import { ActiveArchivedToggle } from "@/components/shared/active-archived-toggle";
import { DataSortSelect, type DataSortOption } from "@/components/shared/data-sort-select";

import type { CategoryType } from "./categories-helpers";

interface CategoriesControlsProps {
  activeTab: CategoryType;
  searchQuery: string;
  canCreateCategories: boolean;
  viewMode: "active" | "archived";
  sort: string;
  sortOptions: readonly DataSortOption[];
  onActiveTabChange: (value: CategoryType) => void;
  onSearchChange: (value: string) => void;
  onViewModeChange: (value: "active" | "archived") => void;
  onSortChange: (value: string) => void;
  onCreate: () => void;
}

export function CategoriesControls({
  activeTab,
  searchQuery,
  canCreateCategories,
  viewMode,
  sort,
  sortOptions,
  onActiveTabChange,
  onSearchChange,
  onViewModeChange,
  onSortChange,
  onCreate,
}: CategoriesControlsProps) {
  return (
    <div className="space-y-4">
      {/* Product / Expense Categories switcher */}
      <Tabs value={activeTab} onValueChange={(value) => onActiveTabChange(value as CategoryType)} className="w-full xl:w-auto">
        <TabsList className="grid grid-cols-2 w-full xl:min-w-[400px]">
          <TabsTrigger value="product" className="flex items-center gap-2">
            <Boxes className="h-4 w-4 shrink-0" />
            Product Categories
          </TabsTrigger>
          <TabsTrigger value="expense" className="flex items-center gap-2">
            <Wallet className="h-4 w-4 shrink-0" />
            Expense Categories
          </TabsTrigger>
        </TabsList>
      </Tabs>

      {/* Active / Archived categories switcher */}
      <ActiveArchivedToggle value={viewMode} onChange={onViewModeChange} />

      {/* Filters (Search & Add) */}
      <FilterCard>
        <div className="flex flex-wrap items-center justify-center gap-4">
          <div className="relative w-full flex-1 min-w-[200px] sm:min-w-[220px]">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Search categories..."
              value={searchQuery}
              onChange={(event) => onSearchChange(event.target.value)}
              className="pl-9 w-full"
            />
          </div>
          <DataSortSelect value={sort} options={sortOptions} onChange={onSortChange} className="w-full flex-1 min-w-[200px] sm:min-w-[220px]" />
          {canCreateCategories && viewMode === "active" && (
            <Button onClick={onCreate} className="w-full flex-1 min-w-[200px] sm:min-w-[220px]">
              <Plus className="h-4 w-4 mr-2" />
              Add Category
            </Button>
          )}
        </div>
      </FilterCard>
    </div>
  );
}

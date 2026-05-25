'use client';

import { Boxes, Plus, Search, Wallet } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";

import type { CategoryType } from "./categories-helpers";

interface CategoriesControlsProps {
  activeTab: CategoryType;
  searchQuery: string;
  canCreateCategories: boolean;
  onActiveTabChange: (value: CategoryType) => void;
  onSearchChange: (value: string) => void;
  onCreate: () => void;
}

export function CategoriesControls({
  activeTab,
  searchQuery,
  canCreateCategories,
  onActiveTabChange,
  onSearchChange,
  onCreate,
}: CategoriesControlsProps) {
  return (
    <div className="flex flex-col xl:flex-row xl:items-center justify-between gap-4">
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

      <div className="flex items-center gap-3 w-full xl:w-auto flex-wrap sm:flex-nowrap">
        <div className="relative flex-1 md:min-w-[240px]">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search categories..."
            value={searchQuery}
            onChange={(event) => onSearchChange(event.target.value)}
            className="pl-9 w-full"
          />
        </div>
        {canCreateCategories ? (
          <Button onClick={onCreate}>
            <Plus className="h-4 w-4 mr-2" />
            Add Category
          </Button>
        ) : null}
      </div>
    </div>
  );
}

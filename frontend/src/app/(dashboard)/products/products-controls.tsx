import { Archive, Plus, Search } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Category } from "@/types";

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
    <Card className="mb-6">
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-4">
        <CardTitle className="text-base font-semibold">Filters</CardTitle>
        {canCreate && viewMode === "active" && (
          <Button onClick={onCreate} size="sm">
            <Plus className="mr-2 h-4 w-4" />
            Add Product
          </Button>
        )}
      </CardHeader>
      <CardContent>
        <div className="flex flex-col gap-4">
          <div className="flex gap-1 rounded-lg bg-muted p-1 w-fit">
            <button
              onClick={() => onViewModeChange("active")}
              className={`px-4 py-2 rounded-md text-sm font-medium transition-colors ${
                viewMode === "active"
                  ? "bg-background text-foreground shadow-sm"
                  : "text-muted-foreground hover:text-foreground"
              }`}
            >
              Active
            </button>
            <button
              onClick={() => onViewModeChange("archived")}
              className={`px-4 py-2 rounded-md text-sm font-medium transition-colors flex items-center gap-2 ${
                viewMode === "archived"
                  ? "bg-background text-foreground shadow-sm"
                  : "text-muted-foreground hover:text-foreground"
              }`}
            >
              <Archive className="h-4 w-4" />
              Archived
            </button>
          </div>

          <div className="flex flex-col gap-4 md:flex-row">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                placeholder="Search products..."
                value={searchQuery}
                onChange={(event) => onSearchChange(event.target.value)}
                className="w-full pl-9"
              />
            </div>
            <Select value={selectedCategory} onValueChange={onCategoryChange}>
              <SelectTrigger className="w-full md:w-48">
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
        </div>
      </CardContent>
    </Card>
  );
}

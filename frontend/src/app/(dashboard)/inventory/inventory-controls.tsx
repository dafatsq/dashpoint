'use client';

import { Search } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";

interface InventoryControlsProps {
  searchQuery: string;
  activeTab: "all" | "low-stock";
  lowStockCount: number;
  onSearchChange: (value: string) => void;
  onTabChange: (value: "all" | "low-stock") => void;
}

export function InventoryControls({
  searchQuery,
  activeTab,
  lowStockCount,
  onSearchChange,
  onTabChange,
}: InventoryControlsProps) {
  return (
    <Card className="mb-4">
      <CardHeader className="pb-4">
        <CardTitle className="text-base font-semibold">Filters</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="flex flex-col gap-4">
          <div className="flex flex-col md:flex-row gap-4">
            <div className="relative flex-1 w-full">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder="Search products..."
                value={searchQuery}
                onChange={(event) => onSearchChange(event.target.value)}
                className="pl-9 w-full"
              />
            </div>

            <div className="flex gap-2">
              <Button variant={activeTab === "all" ? "default" : "outline"} onClick={() => onTabChange("all")}>
                All Products
              </Button>
              <Button
                variant={activeTab === "low-stock" ? "default" : "outline"}
                onClick={() => onTabChange("low-stock")}
              >
                Low Stock ({lowStockCount})
              </Button>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

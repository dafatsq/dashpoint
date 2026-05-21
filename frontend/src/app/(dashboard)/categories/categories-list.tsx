'use client';

import { Lock, Pencil, Plus, RotateCcw, Trash2 } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";

import {
  isSpecialExpenseCategory,
  resolveCategoryActionLabels,
  type CategoryType,
  type CategoryViewMode,
  type ManagedCategory,
} from "./categories-helpers";

interface CategoriesListProps {
  categories: ManagedCategory[];
  isLoading: boolean;
  type: CategoryType;
  viewMode: CategoryViewMode;
  canEditCategories: boolean;
  canDeleteCategories: boolean;
  onEdit: (category: ManagedCategory) => void;
  onArchive: (category: ManagedCategory) => void;
  onRestore: (category: ManagedCategory) => void;
  onPermanentDelete: (category: ManagedCategory) => void;
}

export function CategoriesList({
  categories,
  isLoading,
  type,
  viewMode,
  canEditCategories,
  canDeleteCategories,
  onEdit,
  onArchive,
  onRestore,
  onPermanentDelete,
}: CategoriesListProps) {
  const labels = resolveCategoryActionLabels(viewMode);

  if (isLoading) {
    return (
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {[1, 2, 3].map((item) => (
          <Card key={item} className="animate-pulse">
            <div className="h-32 bg-muted/50 rounded-lg" />
          </Card>
        ))}
      </div>
    );
  }

  if (categories.length === 0) {
    return (
      <Card className="border-dashed">
        <CardContent className="flex flex-col items-center justify-center py-12 text-center">
          <div className="h-12 w-12 rounded-full bg-muted flex items-center justify-center mb-4">
            {viewMode === "active" ? (
              <Plus className="h-6 w-6 text-muted-foreground" />
            ) : (
              <Trash2 className="h-6 w-6 text-muted-foreground" />
            )}
          </div>
          <h3 className="font-medium text-lg">{labels.emptyTitle}</h3>
          <p className="text-muted-foreground max-w-sm">{labels.emptyDescription}</p>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      {categories.map((category) => {
        const isSpecial = isSpecialExpenseCategory(category, type);
        const canShowActions = !isSpecial && (canEditCategories || canDeleteCategories);
        return (
          <Card key={category.id} className="group hover:border-primary/50 transition-all duration-200 shadow-sm hover:shadow-md">
            <CardHeader className="pb-2">
              <div className="flex items-start justify-between">
                <div className="space-y-1">
                  <div className="flex items-center gap-2">
                    <CardTitle className="text-lg">{category.name}</CardTitle>
                    {isSpecial ? (
                      <span className="inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-xs font-medium bg-primary/10 text-primary">
                        <Lock className="h-3 w-3" />
                        Special
                      </span>
                    ) : null}
                  </div>
                  <CardDescription className="line-clamp-1 h-5">
                    {category.description || "No description"}
                  </CardDescription>
                </div>
                {canShowActions ? (
                  <div className="flex items-center gap-1">
                    {viewMode === "active" ? (
                      <>
                        {canEditCategories ? (
                          <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => onEdit(category)}>
                            <Pencil className="h-4 w-4" />
                          </Button>
                        ) : null}
                        {canDeleteCategories ? (
                          <Button variant="ghost" size="icon" className="h-8 w-8 text-destructive" onClick={() => onArchive(category)}>
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        ) : null}
                      </>
                    ) : (
                      <>
                        {canDeleteCategories ? (
                          <Button variant="ghost" size="icon" className="h-8 w-8 text-blue-600" onClick={() => onRestore(category)}>
                            <RotateCcw className="h-4 w-4" />
                          </Button>
                        ) : null}
                        {canDeleteCategories ? (
                          <Button variant="ghost" size="icon" className="h-8 w-8 text-destructive" onClick={() => onPermanentDelete(category)}>
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        ) : null}
                      </>
                    )}
                  </div>
                ) : null}
              </div>
            </CardHeader>
            <CardContent>
              <div className="flex items-center justify-between text-xs text-muted-foreground mt-2 pt-2 border-t">
                <div className="flex items-center gap-1.5">
                  {type === "product" && "product_count" in category ? (
                    <span className="px-1.5 py-0.5 rounded bg-primary/10 text-primary font-medium">
                      {category.product_count || 0} Products
                    </span>
                  ) : null}
                  {viewMode === "archived" ? (
                    <span className="px-1.5 py-0.5 rounded bg-amber-100 text-amber-700 font-medium">
                      Archived
                    </span>
                  ) : null}
                </div>
                <span>Updated {new Date(category.updated_at).toLocaleDateString()}</span>
              </div>
            </CardContent>
          </Card>
        );
      })}
    </div>
  );
}

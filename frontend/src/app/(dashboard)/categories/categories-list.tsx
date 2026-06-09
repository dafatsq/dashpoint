'use client';

import { Archive, Loader2, Lock, Pencil, Plus, RotateCcw, Trash2 } from "lucide-react";

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
      <div className="flex h-64 items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
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
              <Archive className="h-6 w-6 text-muted-foreground" />
            )}
          </div>
          <h3 className="font-medium text-lg">{labels.emptyTitle}</h3>
          <p className="text-muted-foreground max-w-sm">{labels.emptyDescription}</p>
        </CardContent>
      </Card>
    );
  }

  const canShowActions =
    viewMode === "active"
      ? canEditCategories || canDeleteCategories
      : canDeleteCategories;

  return (
    <>
      {/* Mobile/Tablet Card View */}
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 lg:hidden">
        {categories.map((category) => {
          const isSpecial = isSpecialExpenseCategory(category, type);
          return (
            <Card key={category.id} className="@container">
              <CardHeader className="pb-2">
                <div className="flex items-start justify-between">
                  <div className="space-y-1 flex-1 pr-4">
                    <div className="flex items-center gap-2">
                      <CardTitle className="text-lg">{category.name}</CardTitle>
                      {isSpecial ? (
                        <span className="inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-xs font-medium bg-primary/10 text-primary">
                          <Lock className="h-3 w-3" />
                          Special
                        </span>
                      ) : null}
                    </div>
                    <CardDescription className="break-words">
                      {category.description || "No description"}
                    </CardDescription>
                  </div>
                  <span className="text-xs text-muted-foreground whitespace-nowrap self-start pt-1">
                    Updated {new Date(category.updated_at).toLocaleDateString()}
                  </span>
                </div>
              </CardHeader>
              <CardContent>
                <div className="flex items-center justify-between mt-2 pt-2 border-t">
                  <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
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
                  {canShowActions ? (
                    <div className="flex items-center gap-2">
                      {viewMode === "active" ? (
                        <>
                          {canEditCategories ? (
                            <Button
                              variant="outline"
                              size="sm"
                              className="h-8 w-8 p-0 @[250px]:w-auto @[250px]:px-3"
                              onClick={() => onEdit(category)}
                              title="Edit"
                            >
                              <Pencil className="h-3.5 w-3.5 @[250px]:mr-1" />
                              <span className="hidden @[250px]:inline">Edit</span>
                            </Button>
                          ) : null}
                          {canDeleteCategories ? (
                            <Button
                              variant="outline"
                              size="sm"
                              className="h-8 w-8 p-0 @[250px]:w-auto @[250px]:px-3 text-amber-600 hover:bg-amber-50 hover:text-amber-700 dark:text-amber-500 dark:hover:bg-amber-950/20"
                              onClick={() => onArchive(category)}
                              title="Archive"
                            >
                              <Archive className="h-3.5 w-3.5 @[250px]:mr-1" />
                              <span className="hidden @[250px]:inline">Archive</span>
                            </Button>
                          ) : null}
                        </>
                      ) : (
                        <>
                          {canDeleteCategories ? (
                            <Button
                              variant="outline"
                              size="sm"
                              className="h-8 w-8 p-0 @[250px]:w-auto @[250px]:px-3"
                              onClick={() => onRestore(category)}
                              title="Restore"
                            >
                              <RotateCcw className="h-3.5 w-3.5 @[250px]:mr-1" />
                              <span className="hidden @[250px]:inline">Restore</span>
                            </Button>
                          ) : null}
                          {canDeleteCategories && !isSpecial ? (
                            <Button
                              variant="outline"
                              size="sm"
                              className="h-8 w-8 p-0 @[250px]:w-auto @[250px]:px-3 text-destructive hover:bg-destructive/10 hover:text-destructive"
                              onClick={() => onPermanentDelete(category)}
                              title="Delete"
                            >
                              <Trash2 className="h-3.5 w-3.5 @[250px]:mr-1" />
                              <span className="hidden @[250px]:inline">Delete</span>
                            </Button>
                          ) : null}
                        </>
                      )}
                    </div>
                  ) : null}
                </div>
              </CardContent>
            </Card>
          );
        })}
      </div>

      {/* Desktop Table View */}
      <Card className="hidden lg:block">
        <CardContent className="p-6">
          <div className="overflow-x-auto">
            <table className="w-full min-w-max">
              <thead>
                <tr className="border-b text-left text-sm text-muted-foreground">
                  <th className="pb-3 font-medium">Category</th>
                  {type === "product" && <th className="pb-3 font-medium text-right">Products</th>}
                  <th className="pb-3 font-medium">Last Updated</th>
                  <th className="pb-3 font-medium text-center">Status</th>
                  {canShowActions && <th className="pb-3 font-medium text-right">Actions</th>}
                </tr>
              </thead>
              <tbody>
                {categories.map((category) => {
                  const isSpecial = isSpecialExpenseCategory(category, type);
                  return (
                    <tr key={category.id} className="border-b last:border-0 hover:bg-muted/50">
                      <td className="py-3">
                        <div className="flex items-center gap-2">
                          <span className="font-medium">{category.name}</span>
                          {isSpecial && (
                            <span className="inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-[10px] font-bold bg-primary/10 text-primary">
                              <Lock className="h-2.5 w-2.5" />
                              Special
                            </span>
                          )}
                        </div>
                        {category.description && (
                          <p className="text-xs text-muted-foreground max-w-xs truncate">{category.description}</p>
                        )}
                      </td>
                      {type === "product" && (
                        <td className="py-3 text-right text-sm">
                          {"product_count" in category ? category.product_count || 0 : 0}
                        </td>
                      )}
                      <td className="py-3 text-sm text-muted-foreground">
                        {new Date(category.updated_at).toLocaleDateString()}
                      </td>
                      <td className="py-3 text-center">
                        <span
                          className={`inline-flex items-center px-2.5 py-1 rounded-full text-xs font-semibold ${
                            viewMode === "active"
                              ? "bg-green-600 text-white dark:bg-green-600/90 dark:text-white"
                              : "bg-amber-100 text-amber-700"
                          }`}
                        >
                          {viewMode === "active" ? "Active" : "Archived"}
                        </span>
                      </td>
                      {canShowActions && (
                        <td className="py-3 text-right">
                          <div className="flex items-center justify-end gap-1">
                            {viewMode === "active" ? (
                              <>
                                {canEditCategories && (
                                  <Button variant="ghost" size="icon" onClick={() => onEdit(category)} title="Edit category">
                                    <Pencil className="h-4 w-4" />
                                  </Button>
                                )}
                                {canDeleteCategories && (
                                  <Button
                                    variant="ghost"
                                    size="icon"
                                    className="text-amber-600 hover:text-amber-700"
                                    onClick={() => onArchive(category)}
                                    title="Archive category"
                                  >
                                    <Archive className="h-4 w-4" />
                                  </Button>
                                )}
                              </>
                            ) : (
                              <>
                                {canDeleteCategories && (
                                  <Button variant="ghost" size="icon" onClick={() => onRestore(category)} title="Restore category">
                                    <RotateCcw className="h-4 w-4 text-blue-600" />
                                  </Button>
                                )}
                                {canDeleteCategories && !isSpecial && (
                                  <Button
                                    variant="ghost"
                                    size="icon"
                                    className="text-destructive hover:text-destructive"
                                    onClick={() => onPermanentDelete(category)}
                                    title="Delete category permanently"
                                  >
                                    <Trash2 className="h-4 w-4" />
                                  </Button>
                                )}
                              </>
                            )}
                          </div>
                        </td>
                      )}
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        </CardContent>
      </Card>
    </>
  );
}

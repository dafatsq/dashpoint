'use client';

import { AlertCircle, Loader2 } from "lucide-react";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

import type { CategoryFormData, CategoryType } from "./categories-helpers";

interface CategoriesFormDialogProps {
  open: boolean;
  activeTab: CategoryType;
  editing: boolean;
  formData: CategoryFormData;
  error: string | null;
  isSubmitting: boolean;
  hasChanges: boolean;
  onOpenChange: (open: boolean) => void;
  onFormDataChange: (formData: CategoryFormData) => void;
  onSubmit: () => void;
}

export function CategoriesFormDialog({
  open,
  activeTab,
  editing,
  formData,
  error,
  isSubmitting,
  hasChanges,
  onOpenChange,
  onFormDataChange,
  onSubmit,
}: CategoriesFormDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{editing ? "Edit Category" : "Add New Category"}</DialogTitle>
          <DialogDescription>
            {activeTab === "product"
              ? "Product categories help organize your inventory."
              : "Expense categories track your business spending."}
          </DialogDescription>
        </DialogHeader>

        <div className="grid gap-4 py-4">
          {error ? (
            <div className="flex items-center gap-2 p-3 rounded-md bg-destructive/10 text-destructive text-sm border border-destructive/20">
              <AlertCircle className="h-4 w-4" />
              {error}
            </div>
          ) : null}
          <div className="grid gap-2">
            <Label htmlFor="name">
              Name <span className="text-destructive">*</span>
            </Label>
            <Input
              id="name"
              value={formData.name}
              onChange={(event) => onFormDataChange({ ...formData, name: event.target.value })}
              placeholder="Category name"
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="description">Description (Optional)</Label>
            <Input
              id="description"
              value={formData.description}
              onChange={(event) => onFormDataChange({ ...formData, description: event.target.value })}
              placeholder="Brief description"
            />
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={isSubmitting}>
            Cancel
          </Button>
          <Button onClick={onSubmit} disabled={isSubmitting || !hasChanges}>
            {isSubmitting ? <Loader2 className="h-4 w-4 mr-2 animate-spin" /> : null}
            {editing ? "Update" : "Create"} Category
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

import { Loader2 } from "lucide-react";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { ImageUpload } from "@/components/ui/image-upload";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Category, Product } from "@/types";

import { ProductFormData } from "./products-helpers";

interface ProductsFormDialogProps {
  open: boolean;
  editingProduct: Product | null;
  categories: Category[];
  formData: ProductFormData;
  formErrors: { name?: string; price?: string; sku?: string; barcode?: string; general?: string };
  isSubmitting: boolean;
  hasChanges: boolean;
  onOpenChange: (open: boolean) => void;
  onFormDataChange: (value: ProductFormData) => void;
  onFormErrorsChange: (
    value: { name?: string; price?: string; sku?: string; barcode?: string; general?: string },
  ) => void;
  onSubmit: () => void;
}

export function ProductsFormDialog({
  open,
  editingProduct,
  categories,
  formData,
  formErrors,
  isSubmitting,
  hasChanges,
  onOpenChange,
  onFormDataChange,
  onFormErrorsChange,
  onSubmit,
}: ProductsFormDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{editingProduct ? "Edit Product" : "Add Product"}</DialogTitle>
          <DialogDescription>
            {editingProduct ? "Update the product details below." : "Fill in the details for the new product."}
          </DialogDescription>
        </DialogHeader>

        <div className="grid gap-4 py-4">
          {formErrors.general && (
            <div className="rounded-md border border-destructive/20 bg-destructive/10 p-3">
              <p className="text-sm text-destructive">{formErrors.general}</p>
            </div>
          )}

          <div className="grid gap-2">
            <Label htmlFor="name">Name *</Label>
            <Input
              id="name"
              value={formData.name}
              onChange={(event) => {
                onFormDataChange({ ...formData, name: event.target.value });
                if (formErrors.name) {
                  onFormErrorsChange({ ...formErrors, name: undefined });
                }
              }}
              placeholder="Product name"
              className={formErrors.name ? "border-destructive" : ""}
            />
            {formErrors.name && <p className="text-sm text-destructive">{formErrors.name}</p>}
          </div>

          <div className="grid gap-2">
            <Label htmlFor="description">Description</Label>
            <Input
              id="description"
              value={formData.description}
              onChange={(event) => onFormDataChange({ ...formData, description: event.target.value })}
              placeholder="Product description"
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="grid gap-2">
              <Label htmlFor="sku">SKU</Label>
              <Input
                id="sku"
                value={formData.sku}
                onChange={(event) => {
                  onFormDataChange({ ...formData, sku: event.target.value });
                  if (formErrors.sku) {
                    onFormErrorsChange({ ...formErrors, sku: undefined });
                  }
                }}
                placeholder="SKU-001"
                className={formErrors.sku ? "border-destructive" : ""}
              />
              {formErrors.sku && <p className="text-sm text-destructive">{formErrors.sku}</p>}
            </div>
            <div className="grid gap-2">
              <Label htmlFor="barcode">Barcode</Label>
              <Input
                id="barcode"
                value={formData.barcode}
                onChange={(event) => {
                  onFormDataChange({ ...formData, barcode: event.target.value });
                  if (formErrors.barcode) {
                    onFormErrorsChange({ ...formErrors, barcode: undefined });
                  }
                }}
                placeholder="8901234567890"
                className={formErrors.barcode ? "border-destructive" : ""}
              />
              {formErrors.barcode && <p className="text-sm text-destructive">{formErrors.barcode}</p>}
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="grid gap-2">
              <Label htmlFor="price">Price (IDR) *</Label>
              <Input
                id="price"
                type="number"
                value={formData.price}
                onChange={(event) => {
                  onFormDataChange({ ...formData, price: event.target.value });
                  if (formErrors.price) {
                    onFormErrorsChange({ ...formErrors, price: undefined });
                  }
                }}
                placeholder="10000"
                className={formErrors.price ? "border-destructive" : ""}
              />
              {formErrors.price && <p className="text-sm text-destructive">{formErrors.price}</p>}
            </div>
            <div className="grid gap-2">
              <Label htmlFor="cost">Cost (IDR)</Label>
              <Input
                id="cost"
                type="number"
                value={formData.cost}
                onChange={(event) => onFormDataChange({ ...formData, cost: event.target.value })}
                placeholder="8000"
              />
            </div>
            <div className="grid gap-2">
              <Label htmlFor="tax_rate">Tax Rate (%)</Label>
              <Input
                id="tax_rate"
                type="number"
                step="0.01"
                value={formData.tax_rate}
                onChange={(event) => onFormDataChange({ ...formData, tax_rate: event.target.value })}
                placeholder="11"
              />
            </div>
          </div>

          {!editingProduct && (
            <div className="grid grid-cols-2 gap-4">
              <div className="grid gap-2">
                <Label htmlFor="initial_quantity">Initial Stock</Label>
                <Input
                  id="initial_quantity"
                  type="number"
                  value={formData.initial_quantity}
                  onChange={(event) => onFormDataChange({ ...formData, initial_quantity: event.target.value })}
                  placeholder="100"
                />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="low_stock_threshold">Min Stock Alert</Label>
                <Input
                  id="low_stock_threshold"
                  type="number"
                  value={formData.low_stock_threshold}
                  onChange={(event) => onFormDataChange({ ...formData, low_stock_threshold: event.target.value })}
                  placeholder="5"
                />
              </div>
            </div>
          )}

          <div className="grid gap-2">
            <Label htmlFor="category">Category</Label>
            <Select
              value={formData.category_id || "none"}
              onValueChange={(value) => onFormDataChange({ ...formData, category_id: value === "none" ? "" : value })}
            >
              <SelectTrigger>
                <SelectValue placeholder="Select category" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="none">No Category</SelectItem>
                {categories.map((category) => (
                  <SelectItem key={category.id} value={category.id}>
                    {category.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="grid gap-2">
            <Label>Product Image</Label>
            <ImageUpload value={formData.image_url} onChange={(url) => onFormDataChange({ ...formData, image_url: url })} />
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={onSubmit} disabled={isSubmitting || (editingProduct ? !hasChanges : false)}>
            {isSubmitting ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Saving...
              </>
            ) : editingProduct ? (
              "Update Product"
            ) : (
              "Create Product"
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

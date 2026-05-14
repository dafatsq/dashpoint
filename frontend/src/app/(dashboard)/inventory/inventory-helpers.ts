import { buildBackendUrl } from "@/lib/config";
import type { AdjustmentType, InventoryAdjustment, Product } from "@/types";

export type InventoryAction = "add" | "remove" | "count";

export const ADJUSTMENT_TYPE_OPTIONS: Record<
  InventoryAction,
  { value: AdjustmentType; label: string }[]
> = {
  add: [
    { value: "purchase", label: "Restock / Purchase" },
    { value: "adjustment", label: "Inventory Correction" },
  ],
  remove: [
    { value: "damage", label: "Damaged / Expired" },
    { value: "loss", label: "Lost / Stolen" },
    { value: "adjustment", label: "Inventory Correction" },
  ],
  count: [{ value: "count", label: "Stock Count" }],
};

export interface AdjustmentFormState {
  action: InventoryAction;
  quantity: string;
  adjustmentType: AdjustmentType;
  notes: string;
}

export function getInventoryProductQuantity(product: Product): number {
  return Number.parseFloat(product.inventory?.quantity || "") || 0;
}

export function getInventoryProductMinQuantity(product: Product): number {
  return Number.parseFloat(product.inventory?.low_stock_threshold || "") || 0;
}

export function getInventoryProductPrice(product: Product): number {
  return Number.parseFloat(product.price || "") || 0;
}

export function classifyInventoryStock(product: Product) {
  const quantity = getInventoryProductQuantity(product);
  const minQuantity = getInventoryProductMinQuantity(product);
  return {
    isLowStock: quantity > 0 && quantity <= minQuantity,
    isOutOfStock: quantity === 0,
  };
}

export function getPermittedInventoryActions(permissions: {
  canAddStock: boolean;
  canRemoveStock: boolean;
  canAdjustStock: boolean;
}): InventoryAction[] {
  const actions: InventoryAction[] = [];
  if (permissions.canAddStock) {
    actions.push("add");
  }
  if (permissions.canRemoveStock) {
    actions.push("remove");
  }
  if (permissions.canAdjustStock) {
    actions.push("count");
  }
  return actions;
}

export function getDefaultAdjustmentType(action: InventoryAction): AdjustmentType {
  if (action === "add") {
    return "purchase";
  }
  if (action === "remove") {
    return "damage";
  }
  return "count";
}

export function createEmptyAdjustmentFormState(action: InventoryAction): AdjustmentFormState {
  return {
    action,
    quantity: "",
    adjustmentType: getDefaultAdjustmentType(action),
    notes: "",
  };
}

export function buildInventoryAdjustmentRequest(input: {
  productId: string;
  action: InventoryAction;
  adjustmentType: AdjustmentType;
  quantity: string;
  currentStock: number;
  notes?: string;
}): InventoryAdjustment {
  const inputQuantity = Number.parseInt(input.quantity, 10) || 0;
  let finalQuantity = inputQuantity;

  if (input.adjustmentType === "count") {
    finalQuantity = inputQuantity;
  } else if (input.adjustmentType === "adjustment") {
    finalQuantity = input.action === "remove" ? -inputQuantity : inputQuantity;
  }

  return {
    product_id: input.productId,
    adjustment_type: input.adjustmentType,
    quantity: String(finalQuantity),
    reason: input.notes || undefined,
  };
}

export function getInventoryProductImageUrl(
  path: string | null | undefined,
  resolver: (path: string) => string = buildBackendUrl,
): string {
  return path ? resolver(path) : "";
}

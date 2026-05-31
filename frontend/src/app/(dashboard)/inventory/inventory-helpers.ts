import { buildBackendUrl } from "@/lib/config";
import type {
  AdjustmentType,
  InventoryAdjustment,
  InventoryAdjustmentRecord,
  Product,
} from "@/types";

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

export function isInventoryActionAllowed(
  action: InventoryAction,
  allowedActions: InventoryAction[],
): boolean {
  return allowedActions.includes(action);
}

export function canSubmitInventoryAdjustment(input: {
  allowedActions: InventoryAction[];
  action: InventoryAction;
  quantity: string;
  isSubmitting: boolean;
}): boolean {
  if (input.isSubmitting) {
    return false;
  }

  if (!input.quantity) {
    return false;
  }

  if (input.allowedActions.length === 0) {
    return false;
  }

  return isInventoryActionAllowed(input.action, input.allowedActions);
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

export function getInventoryAdjustmentTypeLabel(type: AdjustmentType): string {
  switch (type) {
    case "initial":
      return "Initial Stock";
    case "purchase":
      return "Restock";
    case "sale":
      return "Sale";
    case "return":
      return "Return";
    case "adjustment":
      return "Correction";
    case "damage":
      return "Damaged";
    case "loss":
      return "Lost";
    case "transfer":
      return "Transfer";
    case "count":
      return "Stock Count";
    default:
      return type;
  }
}

export function getInventoryAdjustmentChangeLabel(
  adjustment: Pick<InventoryAdjustmentRecord, "adjustment_type" | "quantity_change" | "quantity_after">,
): string {
  const quantityChange = Number.parseFloat(adjustment.quantity_change || "") || 0;
  const quantityAfter = Number.parseFloat(adjustment.quantity_after || "") || 0;

  if (adjustment.adjustment_type === "count") {
    return `Set to ${quantityAfter}`;
  }

  const sign = quantityChange > 0 ? "+" : "";
  return `${sign}${quantityChange}`;
}

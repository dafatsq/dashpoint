export interface LowStockItem {
  id: string;
  name: string;
  sku: string | null;
  barcode: string | null;
  price: string;
  category_id: string | null;
  is_active: boolean;
  quantity: string;
  available_quantity: string;
  is_low_stock: boolean;
}

export type AdjustmentType =
  | "initial"
  | "purchase"
  | "sale"
  | "return"
  | "adjustment"
  | "damage"
  | "loss"
  | "transfer"
  | "count";

export interface InventoryAdjustment {
  product_id: string;
  adjustment_type: AdjustmentType;
  quantity: string;
  reason?: string;
  expected_updated_at?: string;
}

export interface InventoryAdjustmentRecord {
  id: string;
  product_id: string;
  adjustment_type: AdjustmentType;
  quantity_before: string;
  quantity_change: string;
  quantity_after: string;
  reason?: string;
  reference_type?: string;
  reference_id?: string;
  adjusted_by: string;
  created_at: string;
  adjusted_by_user?: {
    name: string;
  };
}

export interface InventoryDetail {
  product_id: string;
  quantity: string;
  available_quantity: string;
  low_stock_threshold: string;
  is_low_stock: boolean;
  updated_at: string;
}

export interface ProductInventoryDetails {
  inventory: InventoryDetail;
  recent_adjustments: InventoryAdjustmentRecord[];
  total_adjustments: number;
}

export interface InventoryValuation {
  total_products: number;
  total_quantity: string;
  total_cost_value: string;
  total_retail_value: string;
  potential_profit: string;
  items?: InventoryValuationItem[];
}

export interface InventoryValuationItem {
  product_id: string;
  product_name: string;
  product_sku?: string;
  category_name?: string;
  quantity: string;
  cost_price: string;
  sell_price: string;
  cost_value: string;
  retail_value: string;
}

export type InventoryReport = InventoryValuation;

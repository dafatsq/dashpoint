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

export type AdjustmentType = "purchase" | "adjustment" | "damage" | "loss" | "count";

export interface InventoryAdjustment {
  product_id: string;
  adjustment_type: AdjustmentType;
  quantity: string;
  reason?: string;
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

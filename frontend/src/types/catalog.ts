export interface Category {
  id: string;
  name: string;
  description: string;
  product_count: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface ProductInventory {
  quantity: string;
  available_quantity: string;
  low_stock_threshold: string;
  is_low_stock: boolean;
}

export interface Product {
  id: string;
  name: string;
  description?: string;
  sku?: string;
  barcode?: string;
  price: string;
  cost: string;
  tax_rate?: string;
  category_id?: string;
  category_name?: string;
  is_active: boolean;
  image_url?: string;
  inventory?: ProductInventory;
  created_at: string;
  updated_at: string;
}

export interface CreateProductRequest {
  name: string;
  description?: string;
  sku?: string;
  barcode?: string;
  price: string;
  cost?: string;
  tax_rate?: string;
  initial_quantity?: string;
  low_stock_threshold?: string;
  category_id?: string;
  image_url?: string;
}

export interface UpdateProductRequest {
  name?: string;
  description?: string;
  sku?: string;
  barcode?: string;
  price?: string;
  cost?: string;
  tax_rate?: string;
  category_id?: string;
  is_active?: boolean;
  image_url?: string;
}

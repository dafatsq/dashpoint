import { CreateProductRequest, Product, UpdateProductRequest } from "@/types";

export interface ProductFormData {
  name: string;
  description: string;
  sku: string;
  barcode: string;
  price: string;
  cost: string;
  tax_rate: string;
  initial_quantity: string;
  low_stock_threshold: string;
  category_id: string;
  image_url: string;
}

export function createEmptyProductFormData(): ProductFormData {
  return {
    name: "",
    description: "",
    sku: "",
    barcode: "",
    price: "",
    cost: "",
    tax_rate: "0",
    initial_quantity: "",
    low_stock_threshold: "5",
    category_id: "",
    image_url: "",
  };
}

export function getProductDisplayImageUrl(
  path: string | null | undefined,
  buildUrl: (path: string) => string,
): string {
  return path ? buildUrl(path) : "";
}

export function getProductPriceValue(product: Product): number {
  return parseFloat(product.price) || 0;
}

export function getProductQuantityValue(product: Product): number {
  return parseFloat(product.inventory?.quantity || "") || 0;
}

export function getProductLowStockThreshold(product: Product): number {
  return parseFloat(product.inventory?.low_stock_threshold || "") || 0;
}

export function mapProductToFormData(product: Product): ProductFormData {
  return {
    name: product.name,
    description: product.description || "",
    sku: product.sku || "",
    barcode: product.barcode || "",
    price: product.price,
    cost: product.cost || "",
    tax_rate: product.tax_rate?.toString() || "0",
    initial_quantity: "",
    low_stock_threshold: product.inventory?.low_stock_threshold || "5",
    category_id: product.category_id || "",
    image_url: product.image_url || "",
  };
}

export function hasProductFormChanges(formData: ProductFormData, editingProduct: Product | null): boolean {
  if (!editingProduct) {
    return true;
  }

  return (
    formData.name !== editingProduct.name ||
    formData.description !== (editingProduct.description || "") ||
    formData.sku !== (editingProduct.sku || "") ||
    formData.barcode !== (editingProduct.barcode || "") ||
    formData.price !== editingProduct.price ||
    formData.cost !== (editingProduct.cost || "") ||
    formData.tax_rate !== (editingProduct.tax_rate?.toString() || "0") ||
    formData.low_stock_threshold !== (editingProduct.inventory?.low_stock_threshold || "5") ||
    formData.category_id !== (editingProduct.category_id || "") ||
    formData.image_url !== (editingProduct.image_url || "")
  );
}

export function buildProductCreateRequest(formData: ProductFormData): CreateProductRequest {
  return {
    name: formData.name,
    description: formData.description || undefined,
    sku: formData.sku || undefined,
    barcode: formData.barcode || undefined,
    price: formData.price,
    cost: formData.cost || undefined,
    tax_rate: formData.tax_rate === "" ? "0" : formData.tax_rate,
    initial_quantity: formData.initial_quantity || undefined,
    low_stock_threshold: formData.low_stock_threshold || undefined,
    category_id: formData.category_id || undefined,
    image_url: formData.image_url || undefined,
  };
}

export function buildProductUpdateRequest(formData: ProductFormData): UpdateProductRequest {
  return {
    name: formData.name,
    description: formData.description || undefined,
    sku: formData.sku || undefined,
    barcode: formData.barcode || undefined,
    price: formData.price,
    cost: formData.cost || undefined,
    tax_rate: formData.tax_rate === "" ? "0" : formData.tax_rate,
    category_id: formData.category_id || undefined,
    image_url: formData.image_url,
  };
}

import type { Category, Product, ProductInventoryDetails } from "@/types";

import type { ApiTransport } from "./transport";
import type {
  CatalogApi,
  CategoriesResponse,
  LowStockResponse,
  ProductsResponse,
} from "./types";

export function createCatalogApi(transport: ApiTransport): CatalogApi {
  return {
    async getProductsPage(params) {
      const searchParams = new URLSearchParams();
      if (params?.category_id) searchParams.set("category_id", params.category_id);
      if (params?.active !== undefined) {
        searchParams.set("active_only", String(params.active));
      }
      if (params?.search) searchParams.set("search", params.search);
      if (params?.page !== undefined) searchParams.set("page", String(params.page));
      if (params?.per_page !== undefined) {
        searchParams.set("per_page", String(params.per_page));
      }
      const query = searchParams.toString();
      const result = await transport.request<ProductsResponse>(
        `/products${query ? `?${query}` : ""}`,
      );
      if (result.error) return { error: result.error };
      return {
        data: result.data?.products || [],
        total: result.data?.total || 0,
        page: result.data?.page || 1,
        per_page: result.data?.per_page || (params?.per_page ?? 20),
        total_pages: result.data?.total_pages || 0,
      };
    },
    async getProducts(params) {
      const result = await this.getProductsPage({
        ...params,
        page: 1,
        per_page: 100,
      });
      if (result.error) return { error: result.error };
      return { data: result.data || [] };
    },
    async getProduct(id) {
      const result = await transport.request<{ product: Product }>(`/products/${id}`);
      if (result.error) return { error: result.error };
      return { data: result.data?.product };
    },
    async getProductInventory(id, params) {
      const searchParams = new URLSearchParams();
      if (params?.limit !== undefined) {
        searchParams.set("limit", String(params.limit));
      }
      if (params?.offset !== undefined) {
        searchParams.set("offset", String(params.offset));
      }
      if (params?.adjustment_type) {
        searchParams.set("adjustment_type", params.adjustment_type);
      }
      const query = searchParams.toString();
      const result = await transport.request<ProductInventoryDetails>(`/products/${id}/inventory${query ? `?${query}` : ""}`);
      if (result.error) return { error: result.error };
      return { data: result.data };
    },
    async updateProductInventoryThreshold(id, payload) {
      const result = await transport.request<{ inventory: ProductInventoryDetails["inventory"] }>(`/products/${id}/inventory`, {
        method: "PATCH",
        body: payload,
      });
      if (result.error) return { error: result.error };
      return { data: result.data?.inventory };
    },
    async lookupProduct(code) {
      const result = await transport.request<{ product: Product }>(
        `/products/lookup?code=${encodeURIComponent(code)}`,
      );
      if (result.error) return { error: result.error };
      return { data: result.data?.product };
    },
    async createProduct(product) {
      const result = await transport.request<{ product: Product }>("/products", {
        method: "POST",
        body: product,
      });
      if (result.error) return { error: result.error };
      return { data: result.data?.product };
    },
    async updateProduct(id, product) {
      const result = await transport.request<{ product: Product }>(`/products/${id}`, {
        method: "PATCH",
        body: product,
      });
      if (result.error) return { error: result.error };
      return { data: result.data?.product };
    },
    deleteProduct(id) {
      return transport.request(`/products/${id}`, { method: "DELETE" });
    },
    permanentDeleteProduct(id) {
      return transport.request(`/products/${id}/permanent`, { method: "DELETE" });
    },
    async getCategories(status = "active") {
      const result = await transport.request<CategoriesResponse>(`/categories?status=${status}`);
      if (result.error) return { error: result.error };
      return { data: result.data?.categories || [] };
    },
    async createCategory(category) {
      const result = await transport.request<{ category: Category }>("/categories", {
        method: "POST",
        body: category,
      });
      if (result.error) return { error: result.error };
      return { data: result.data?.category };
    },
    async updateCategory(id, category) {
      const result = await transport.request<{ category: Category }>(`/categories/${id}`, {
        method: "PATCH",
        body: category,
      });
      if (result.error) return { error: result.error };
      return { data: result.data?.category };
    },
    deleteCategory(id) {
      return transport.request(`/categories/${id}`, { method: "DELETE" });
    },
    permanentDeleteCategory(id) {
      return transport.request(`/categories/${id}/permanent`, { method: "DELETE" });
    },
    async getLowStock(threshold) {
      const query = threshold ? `?threshold=${threshold}` : "";
      const result = await transport.request<LowStockResponse>(
        `/inventory/low-stock${query}`,
      );
      if (result.error) return { error: result.error };
      return { data: result.data?.products || [] };
    },
    adjustInventory(adjustment) {
      return transport.request("/inventory/adjust", {
        method: "POST",
        body: adjustment,
      });
    },
  };
}

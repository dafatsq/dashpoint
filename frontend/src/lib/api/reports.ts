import type { CategorySales, DailyReport, EmployeeSales } from "@/types";

import type { ApiTransport } from "./transport";
import type { InventoryValuationResponse, ReportsApi, TopSellersResponse } from "./types";

export function createReportsApi(transport: ApiTransport): ReportsApi {
  async function exportBlobUrl(endpoint: string): Promise<string> {
    const result = await transport.requestBlob(endpoint);
    if (result.error || !result.data) {
      throw new Error(result.error || "Export failed");
    }

    return URL.createObjectURL(result.data);
  }

  return {
    async getDailyReport(date) {
      const query = date ? `?date=${date}` : "";
      const result = await transport.request<{ report: DailyReport }>(
        `/reports/daily${query}`,
      );
      if (result.error) return { error: result.error };
      return { data: result.data?.report };
    },
    async getSalesRangeReport(params) {
      const searchParams = new URLSearchParams();
      searchParams.set("start_date", params.start_date);
      searchParams.set("end_date", params.end_date);
      return transport.request(`/reports/sales?${searchParams.toString()}`);
    },
    async getTopSellers(params) {
      const searchParams = new URLSearchParams();
      if (params?.from) searchParams.set("start_date", params.from);
      if (params?.to) searchParams.set("end_date", params.to);
      if (params?.limit) searchParams.set("limit", String(params.limit));
      const query = searchParams.toString();
      const result = await transport.request<TopSellersResponse>(
        `/reports/top-sellers${query ? `?${query}` : ""}`,
      );
      if (result.error) return { error: result.error };
      return { data: result.data?.top_sellers || [] };
    },
    async getInventoryReport(includeItems) {
      const query = includeItems ? "?include_items=true" : "";
      const result = await transport.request<InventoryValuationResponse>(
        `/reports/inventory${query}`,
      );
      if (result.error) return { error: result.error };
      return { data: result.data?.valuation };
    },
    async getCashReport(params) {
      const searchParams = new URLSearchParams();
      if (params?.start_date) searchParams.set("start_date", params.start_date);
      if (params?.end_date) searchParams.set("end_date", params.end_date);
      const query = searchParams.toString();
      const result = await transport.request<{ report: import("@/types").CashReport }>(
        `/reports/cash${query ? `?${query}` : ""}`,
      );
      if (result.error) return { error: result.error };
      return { data: result.data?.report };
    },
    async getEmployeeSalesReport(params) {
      const searchParams = new URLSearchParams();
      if (params?.start_date) searchParams.set("start_date", params.start_date);
      if (params?.end_date) searchParams.set("end_date", params.end_date);
      const query = searchParams.toString();
      const result = await transport.request<{ employees: EmployeeSales[] }>(
        `/reports/by-employee${query ? `?${query}` : ""}`,
      );
      if (result.error) return { error: result.error };
      return { data: result.data?.employees || [] };
    },
    async getCategorySalesReport(params) {
      const searchParams = new URLSearchParams();
      if (params?.start_date) searchParams.set("start_date", params.start_date);
      if (params?.end_date) searchParams.set("end_date", params.end_date);
      const query = searchParams.toString();
      const result = await transport.request<{ categories: CategorySales[] }>(
        `/reports/by-category${query ? `?${query}` : ""}`,
      );
      if (result.error) return { error: result.error };
      return { data: result.data?.categories || [] };
    },
    exportSalesCSV(params) {
      const searchParams = new URLSearchParams();
      searchParams.set("start_date", params.start_date);
      searchParams.set("end_date", params.end_date);
      return exportBlobUrl(
        `/reports/export/sales?${searchParams.toString()}`,
      );
    },
    exportInventoryCSV() {
      return exportBlobUrl(
        `/reports/export/inventory`,
      );
    },
    exportTopSellersCSV(params) {
      const searchParams = new URLSearchParams();
      if (params?.start_date) searchParams.set("start_date", params.start_date);
      if (params?.end_date) searchParams.set("end_date", params.end_date);
      if (params?.limit) searchParams.set("limit", String(params.limit));
      const query = searchParams.toString();
      return exportBlobUrl(
        `/reports/export/top-sellers${query ? `?${query}` : ""}`,
      );
    },
    exportComprehensiveReportCSV(params) {
      const searchParams = new URLSearchParams();
      searchParams.set("start_date", params.start_date);
      searchParams.set("end_date", params.end_date);
      return exportBlobUrl(
        `/reports/export/comprehensive?${searchParams.toString()}`,
      );
    },
  };
}

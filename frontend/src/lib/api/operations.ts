import type { CashDrawerOperation, CashDrawerOperationsResponse, DailySummary, Sale, Shift } from "@/types";

import type { ApiTransport } from "./transport";
import type { OperationsApi, SalesResponse, ShiftsResponse } from "./types";

export function createOperationsApi(transport: ApiTransport): OperationsApi {
  const buildQuery = (params: Record<string, string | number | undefined>) => {
    const searchParams = new URLSearchParams();

    for (const [key, value] of Object.entries(params)) {
      if (value !== undefined) {
        searchParams.set(key, String(value));
      }
    }

    const query = searchParams.toString();
    return query ? `?${query}` : "";
  };

  return {
    async getCurrentShift() {
      const result = await transport.request<{ shift: Shift }>("/shifts/current");
      if (result.error) {
        if (
          result.error.includes("not found") ||
          result.error.includes("NO_ACTIVE_SHIFT")
        ) {
          return { data: null };
        }
        return { error: result.error };
      }
      return { data: result.data?.shift };
    },
    async startShift(startingCash) {
      const result = await transport.request<{ shift: Shift }>("/shifts/start", {
        method: "POST",
        body: { opening_cash: String(startingCash) },
      });
      if (result.error) return { error: result.error };
      return { data: result.data?.shift };
    },
    async closeShift(closingCash, notes) {
      const result = await transport.request<{ shift: Shift }>("/shifts/close", {
        method: "POST",
        body: { closing_cash: String(closingCash), notes },
      });
      if (result.error) return { error: result.error };
      return { data: result.data?.shift };
    },
    async getShifts(params) {
      const result = await transport.request<ShiftsResponse>(`/shifts${buildQuery({
        user_id: params?.user_id,
        from: params?.from,
        to: params?.to,
        limit: params?.limit,
        offset: params?.offset,
      })}`);
      if (result.error) return { error: result.error };
      return { data: result.data?.shifts || [], total: result.data?.total || 0 };
    },
    async payIn(amount, reason) {
      const result = await transport.request<{ operation: CashDrawerOperation }>(
        "/shifts/pay-in",
        {
          method: "POST",
          body: { amount, reason },
        },
      );
      if (result.error) return { error: result.error };
      return { data: result.data?.operation };
    },
    async payOut(amount, reason) {
      const result = await transport.request<{ operation: CashDrawerOperation }>(
        "/shifts/pay-out",
        {
          method: "POST",
          body: { amount, reason },
        },
      );
      if (result.error) return { error: result.error };
      return { data: result.data?.operation };
    },
    async getShiftOperations(shiftId) {
      const result = await transport.request<CashDrawerOperationsResponse>(
        `/shifts/${shiftId}/operations`,
      );
      if (result.error) return { error: result.error };
      return { data: result.data };
    },
    async createSale(sale) {
      const result = await transport.request<{ sale: Sale }>("/sales", {
        method: "POST",
        body: sale,
      });
      if (result.error) return { error: result.error };
      return { data: result.data?.sale };
    },
    async getSalesPage(params) {
      const result = await transport.request<SalesResponse>(`/sales${buildQuery({
        from: params?.from,
        to: params?.to,
        employee_id: params?.user_id,
        status: params?.status,
        invoice_no: params?.invoice_no,
        limit: params?.limit,
        offset: params?.offset,
      })}`);
      if (result.error) return { error: result.error };
      return {
        data: result.data?.sales || [],
        total: result.data?.total || 0,
        limit: result.data?.limit || (params?.limit ?? 20),
        offset: result.data?.offset || (params?.offset ?? 0),
      };
    },
    async getSales(params) {
      const result = await this.getSalesPage({
        ...params,
        limit: 100,
        offset: 0,
      });
      if (result.error) return { error: result.error };
      return { data: result.data || [] };
    },
    async getSale(id) {
      const result = await transport.request<{ sale: Sale }>(`/sales/${id}`);
      if (result.error) return { error: result.error };
      return { data: result.data?.sale };
    },
    voidSale(id, reason) {
      return transport.request(`/sales/${id}/void`, {
        method: "POST",
        body: { reason },
      });
    },
    async getDailySummary(date) {
      const query = date ? `?date=${date}` : "";
      const result = await transport.request<{ summary: DailySummary }>(
        `/sales/summary/daily${query}`,
      );
      if (result.error) return { error: result.error };
      return { data: result.data?.summary };
    },
  };
}

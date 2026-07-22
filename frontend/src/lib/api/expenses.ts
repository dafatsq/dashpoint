import type { Expense, ExpenseCategory, ExpenseSummary } from "@/types";

import type { ApiTransport } from "./transport";
import type { ExpensesApi } from "./types";

function withExpectedUpdatedAt(path: string, expectedUpdatedAt?: string): string {
  if (!expectedUpdatedAt) return path;
  const searchParams = new URLSearchParams({ expected_updated_at: expectedUpdatedAt });
  return `${path}?${searchParams.toString()}`;
}

export function createExpensesApi(transport: ApiTransport): ExpensesApi {
  return {
    async getExpenseCategories(status = "active") {
      const result = await transport.request<{ data: ExpenseCategory[] }>(
        `/expenses/categories?status=${status}`,
      );
      if (result.error) return { error: result.error };
      return { data: result.data?.data || [] };
    },
    async createExpenseCategory(name, description) {
      const result = await transport.request<{ data: ExpenseCategory }>(
        "/expenses/categories",
        {
          method: "POST",
          body: { name, description },
        },
      );
      if (result.error) return { error: result.error };
      return { data: result.data?.data };
    },
    async updateExpenseCategory(id, category) {
      const result = await transport.request<{ data: ExpenseCategory }>(
        `/expenses/categories/${id}`,
        {
          method: "PATCH",
          body: category,
        },
      );
      if (result.error) return { error: result.error };
      return { data: result.data?.data };
    },
    deleteExpenseCategory(id, expectedUpdatedAt) {
      return transport.request(
        withExpectedUpdatedAt(`/expenses/categories/${id}`, expectedUpdatedAt),
        { method: "DELETE" },
      );
    },
    permanentDeleteExpenseCategory(id, expectedUpdatedAt) {
      return transport.request(
        withExpectedUpdatedAt(
          `/expenses/categories/${id}/permanent`,
          expectedUpdatedAt,
        ),
        { method: "DELETE" },
      );
    },
    async getExpenses(params) {
      const searchParams = new URLSearchParams();
      if (params?.category_id) searchParams.set("category_id", params.category_id);
      if (params?.start_date) searchParams.set("start_date", params.start_date);
      if (params?.end_date) searchParams.set("end_date", params.end_date);
      if (params?.limit) searchParams.set("limit", String(params.limit));
      if (params?.offset) searchParams.set("offset", String(params.offset));
      if (params?.sort_by) searchParams.set("sort_by", params.sort_by);
      if (params?.sort_direction) searchParams.set("sort_direction", params.sort_direction);
      const query = searchParams.toString();
      const result = await transport.request<{ data: Expense[]; total: number }>(
        `/expenses${query ? `?${query}` : ""}`,
      );
      if (result.error) return { error: result.error };
      return { data: { expenses: result.data?.data || [], total: result.data?.total || 0 } };
    },
    async getExpense(id) {
      const result = await transport.request<{ data: Expense }>(`/expenses/${id}`);
      if (result.error) return { error: result.error };
      return { data: result.data?.data };
    },
    async createExpense(expense) {
      const result = await transport.request<{ data: Expense }>("/expenses", {
        method: "POST",
        body: expense,
      });
      if (result.error) return { error: result.error };
      return { data: result.data?.data };
    },
    async updateExpense(id, expense) {
      const result = await transport.request<{ data: Expense }>(`/expenses/${id}`, {
        method: "PATCH",
        body: expense,
      });
      if (result.error) return { error: result.error };
      return { data: result.data?.data };
    },
    deleteExpense(id, expectedUpdatedAt) {
      return transport.request(withExpectedUpdatedAt(`/expenses/${id}`, expectedUpdatedAt), {
        method: "DELETE",
      });
    },
    async getExpenseSummary(params) {
      const searchParams = new URLSearchParams();
      if (params?.start_date) searchParams.set("start_date", params.start_date);
      if (params?.end_date) searchParams.set("end_date", params.end_date);
      const query = searchParams.toString();
      const result = await transport.request<{ data: ExpenseSummary }>(
        `/expenses/summary${query ? `?${query}` : ""}`,
      );
      if (result.error) return { error: result.error };
      return { data: result.data?.data };
    },
    async getMonthlyExpenses(months) {
      const query = months ? `?months=${months}` : "";
      const result = await transport.request<{ data: { month: string; total: string }[] }>(
        `/expenses/monthly${query}`,
      );
      if (result.error) return { error: result.error };
      return { data: result.data?.data || [] };
    },
  };
}

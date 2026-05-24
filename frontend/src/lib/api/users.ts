import type { User } from "@/types";

import type { ApiTransport } from "./transport";
import type { UserApi, UsersResponse } from "./types";

function buildUsersPageQuery(params: {
  role?: string;
  active?: boolean;
  search?: string;
  page?: number;
  per_page?: number;
}): string {
  const searchParams = new URLSearchParams();

  if (params.role && params.role !== "all") searchParams.set("role", params.role);
  if (params.active !== undefined) {
    searchParams.set("active_only", String(params.active));
  }
  if (params.search) searchParams.set("search", params.search);
  if (params.page !== undefined) searchParams.set("page", String(params.page));
  if (params.per_page !== undefined) {
    searchParams.set("per_page", String(params.per_page));
  }

  const query = searchParams.toString();
  return query ? `?${query}` : "";
}

export function createUserApi(transport: ApiTransport): UserApi {
  return {
    async getUsersPage(params) {
      const result = await transport.request<UsersResponse>(
        `/users${buildUsersPageQuery(params || {})}`,
      );
      if (result.error) return { error: result.error };

      return {
        data: result.data?.users || [],
        total: result.data?.total || 0,
        page: result.data?.page || 1,
        per_page: result.data?.per_page || (params?.per_page ?? 20),
        total_pages: result.data?.total_pages || 0,
      };
    },
    async getUsers(params) {
      const result = await this.getUsersPage({
        ...params,
        page: 1,
        per_page: 100,
      });
      if (result.error) return { error: result.error };
      return { data: result.data || [] };
    },
    async getBasicUsers() {
      const result = await transport.request<{ data: { id: string; name: string }[] }>("/users/basic");
      if (result.error) return { error: result.error };
      return { data: result.data?.data || [] };
    },
    async getUser(id) {
      const result = await transport.request<{ user: User }>(`/users/${id}`);
      if (result.error) return { error: result.error };
      return { data: result.data?.user };
    },
    async createUser(user) {
      const result = await transport.request<{ user: User }>("/users", {
        method: "POST",
        body: user,
      });
      if (result.error) return { error: result.error };
      return { data: result.data?.user };
    },
    async updateUser(id, user) {
      const result = await transport.request<{ user: User }>(`/users/${id}`, {
        method: "PATCH",
        body: user,
      });
      if (result.error) return { error: result.error };
      return { data: result.data?.user };
    },
    deleteUser(id) {
      return transport.request(`/users/${id}`, { method: "DELETE" });
    },
    permanentDeleteUser(id) {
      return transport.request(`/users/${id}/permanent`, { method: "DELETE" });
    },
    async getRoles() {
      const result = await transport.request<{
        roles: { id: string; name: string; description: string }[];
      }>("/roles");
      if (result.error) return { error: result.error };
      return { data: result.data?.roles || [] };
    },
  };
}

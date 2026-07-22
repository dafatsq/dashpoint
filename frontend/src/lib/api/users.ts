import type { Role, User } from "@/types";

import type { ApiTransport } from "./transport";
import type { UserApi, UsersResponse } from "./types";

function buildUsersPageQuery(params: {
  role?: string;
  active?: boolean;
  search?: string;
  page?: number;
  per_page?: number;
  sort_by?: string;
  sort_direction?: "asc" | "desc";
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
  if (params.sort_by) searchParams.set("sort_by", params.sort_by);
  if (params.sort_direction) searchParams.set("sort_direction", params.sort_direction);

  const query = searchParams.toString();
  return query ? `?${query}` : "";
}

function withExpectedUpdatedAt(path: string, expectedUpdatedAt?: string): string {
  if (!expectedUpdatedAt) return path;
  const searchParams = new URLSearchParams({ expected_updated_at: expectedUpdatedAt });
  return `${path}?${searchParams.toString()}`;
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
      const { role, permissions, ...payload } = user;
      void role;
      void permissions;
      const result = await transport.request<{ user: User }>("/users", {
        method: "POST",
        body: payload,
      });
      if (result.error) return { error: result.error };
      return { data: result.data?.user };
    },
    async updateUser(id, user) {
      const { role, permissions, ...payload } = user;
      void role;
      void permissions;
      const result = await transport.request<{ user: User }>(`/users/${id}`, {
        method: "PATCH",
        body: payload,
      });
      if (result.error) return { error: result.error };
      return { data: result.data?.user };
    },
    deleteUser(id, expectedUpdatedAt) {
      return transport.request(withExpectedUpdatedAt(`/users/${id}`, expectedUpdatedAt), {
        method: "DELETE",
      });
    },
    permanentDeleteUser(id, expectedUpdatedAt) {
      return transport.request(
        withExpectedUpdatedAt(`/users/${id}/permanent`, expectedUpdatedAt),
        { method: "DELETE" },
      );
    },
    async getRoles() {
      const result = await transport.request<{ roles: Role[] }>("/roles");
      if (result.error) return { error: result.error };
      return { data: result.data?.roles || [] };
    },
    async updateRolePermissions(id, permissions, expectedPermissions) {
      const result = await transport.request<{ role: Role }>(`/roles/${id}/permissions`, {
        method: "PATCH",
        body: { permissions, expected_permissions: expectedPermissions },
      });
      if (result.error) return { error: result.error };
      return { data: result.data?.role };
    },
  };
}

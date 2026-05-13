import type { AuditApi, AuditLogsResponse } from "./types";
import type { ApiTransport } from "./transport";

export function createAuditApi(transport: ApiTransport): AuditApi {
  return {
    async getAuditLogs(params) {
      const searchParams = new URLSearchParams();
      if (params?.user_id) searchParams.set("user_id", params.user_id);
      if (params?.action) searchParams.set("action", params.action);
      if (params?.entity_type) searchParams.set("entity_type", params.entity_type);
      if (params?.from) searchParams.set("start_date", params.from);
      if (params?.to) searchParams.set("end_date", params.to);
      if (params?.limit) searchParams.set("limit", String(params.limit));
      if (params?.offset) searchParams.set("offset", String(params.offset));
      const query = searchParams.toString();
      const result = await transport.request<AuditLogsResponse>(
        `/logs${query ? `?${query}` : ""}`,
      );
      if (result.error) return { error: result.error };
      return { data: result.data?.logs || [] };
    },
    async getDashboardChanges(params) {
      const searchParams = new URLSearchParams();
      if (params?.entity_type) searchParams.set("entity_type", params.entity_type);
      if (params?.limit !== undefined) searchParams.set("limit", String(params.limit));
      if (params?.offset !== undefined) {
        searchParams.set("offset", String(params.offset));
      }
      if (params?.user_id) searchParams.set("user_id", params.user_id);
      if (params?.from) searchParams.set("start_date", params.from);
      if (params?.to) searchParams.set("end_date", params.to);
      const query = searchParams.toString();
      const result = await transport.request<AuditLogsResponse>(
        `/dashboard/changes${query ? `?${query}` : ""}`,
      );
      if (result.error) return { error: result.error };
      return { data: result.data?.logs || [], total: result.data?.total || 0 };
    },
  };
}

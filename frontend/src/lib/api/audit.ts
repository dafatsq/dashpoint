import type { AuditApi, AuditLogsResponse } from "./types";
import type { ApiTransport } from "./transport";

function appendIfPresent(
  searchParams: URLSearchParams,
  key: string,
  value: string | number | undefined,
) {
  if (value !== undefined && value !== "") {
    searchParams.set(key, String(value));
  }
}

function buildAuditQuery(params?: {
  user_id?: string;
  action?: string;
  entity_type?: string;
  from?: string;
  to?: string;
  search?: string;
  limit?: number;
  offset?: number;
}) {
  const searchParams = new URLSearchParams();
  appendIfPresent(searchParams, "user_id", params?.user_id);
  appendIfPresent(searchParams, "action", params?.action);
  appendIfPresent(searchParams, "entity_type", params?.entity_type);
  appendIfPresent(searchParams, "start_date", params?.from);
  appendIfPresent(searchParams, "end_date", params?.to);
  appendIfPresent(searchParams, "search", params?.search);
  appendIfPresent(searchParams, "limit", params?.limit);
  appendIfPresent(searchParams, "offset", params?.offset);
  const query = searchParams.toString();
  return query ? `?${query}` : "";
}

export function createAuditApi(transport: ApiTransport): AuditApi {
  return {
    async getAuditLogs(params) {
      const result = await transport.request<AuditLogsResponse>(
        `/logs${buildAuditQuery(params)}`,
      );
      if (result.error) return { error: result.error };
      return { data: result.data?.logs || [], total: result.data?.total || 0 };
    },
    async getDashboardChanges(params) {
      const result = await transport.request<AuditLogsResponse>(
        `/dashboard/changes${buildAuditQuery(params)}`,
      );
      if (result.error) return { error: result.error };
      return { data: result.data?.logs || [], total: result.data?.total || 0 };
    },
  };
}

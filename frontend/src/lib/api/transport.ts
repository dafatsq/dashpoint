import { clearAuthSession, getAccessToken, persistAuthPayload, refreshSessionTokens } from "@/lib/auth-session";

import type { ApiResponse, RequestOptions } from "./types";

export class ApiTransport {
  constructor(private readonly baseUrl: string) {}

  getBaseUrl(): string {
    return this.baseUrl;
  }

  getAccessToken(): string | null {
    return getAccessToken();
  }

  setTokens(accessToken: string, refreshToken: string): void {
    persistAuthPayload({ access_token: accessToken, refresh_token: refreshToken });
  }

  clearTokens(): void {
    clearAuthSession();
  }

  async refreshTokens(): Promise<boolean> {
    return refreshSessionTokens();
  }

  async request<T>(
    endpoint: string,
    options: RequestOptions = {},
  ): Promise<ApiResponse<T>> {
    const { method = "GET", body, headers = {} } = options;

    const requestHeaders: Record<string, string> = {
      "Content-Type": "application/json",
      ...headers,
    };

    const accessToken = getAccessToken();
    if (accessToken) {
      requestHeaders.Authorization = `Bearer ${accessToken}`;
    }

    try {
      let response = await fetch(`${this.baseUrl}${endpoint}`, {
        method,
        headers: requestHeaders,
        body: body ? JSON.stringify(body) : undefined,
        cache: "no-store",
      });

      if (response.status === 401 && accessToken) {
        const refreshed = await this.refreshTokens();
        if (refreshed) {
          const nextToken = getAccessToken();
          if (nextToken) {
            requestHeaders.Authorization = `Bearer ${nextToken}`;
          }

          response = await fetch(`${this.baseUrl}${endpoint}`, {
            method,
            headers: requestHeaders,
            body: body ? JSON.stringify(body) : undefined,
            cache: "no-store",
          });
        } else {
          if (typeof window !== "undefined") {
            window.location.href = "/login";
          }
          return { error: "Session expired. Please login again." };
        }
      }

      const text = await response.text();
      let data: Record<string, unknown> = {};

      try {
        data = text ? (JSON.parse(text) as Record<string, unknown>) : {};
      } catch {
        data = {};
      }

      if (!response.ok) {
        const errorMsg =
          (data.error as string | undefined) ||
          (data.message as string | undefined) ||
          `Request failed with status ${response.status}`;

        if (response.status === 401 && data.code === "ACCOUNT_INACTIVE") {
          this.clearTokens();
          if (typeof window !== "undefined") {
            window.location.href = "/login?message=account_inactive";
          }
          return { error: "Your account has been deactivated" };
        }

        if (response.status >= 400 && response.status < 500) {
          console.warn("API Client Error:", {
            status: response.status,
            message: errorMsg,
            data,
          });
        } else {
          console.error("API Server Error:", {
            status: response.status,
            message: errorMsg,
            data,
          });
        }
        return { error: errorMsg };
      }

      return { data: data as T };
    } catch (error) {
      console.error("API request failed:", error);
      return { error: "Network error. Please check your connection." };
    }
  }
}

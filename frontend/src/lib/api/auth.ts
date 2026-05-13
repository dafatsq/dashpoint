import type { ApiTransport } from "./transport";
import type { ApiResponse, AuthApi } from "./types";
import type { AuthPayload } from "@/lib/auth-user";
import type { User } from "@/types";

export function createAuthApi(transport: ApiTransport): AuthApi & {
  request: ApiTransport["request"];
  setTokens: ApiTransport["setTokens"];
  clearTokens: ApiTransport["clearTokens"];
  refreshTokens: ApiTransport["refreshTokens"];
} {
  return {
    request: transport.request.bind(transport),
    setTokens: transport.setTokens.bind(transport),
    clearTokens: transport.clearTokens.bind(transport),
    refreshTokens: transport.refreshTokens.bind(transport),
    async getMe(): Promise<ApiResponse<User>> {
      const result = await transport.request<{ user: User }>("/me");
      if (result.error) return { error: result.error };
      return { data: result.data?.user };
    },
    login(email: string, password: string) {
      return transport.request<AuthPayload>("/auth/login", {
        method: "POST",
        body: { email, password },
      });
    },
    pinLogin(userId: string, pin: string) {
      return transport.request<AuthPayload>("/auth/pin-login", {
        method: "POST",
        body: { user_id: userId, pin },
      });
    },
    async logout() {
      const result = await transport.request("/auth/logout", { method: "POST" });
      transport.clearTokens();
      return result;
    },
  };
}

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
  const request = transport.request.bind(transport);
  const setTokens = transport.setTokens.bind(transport);
  const clearTokens = transport.clearTokens.bind(transport);
  const refreshTokens = transport.refreshTokens.bind(transport);

  return {
    request,
    setTokens,
    clearTokens,
    refreshTokens,
    async getMe(): Promise<ApiResponse<User>> {
      const result = await request<{ user: User }>("/me");
      if (result.error) return { error: result.error };
      return { data: result.data?.user };
    },
    login(email: string, password: string, rememberMe?: boolean) {
      return request<AuthPayload>("/auth/login", {
        method: "POST",
        body:
          rememberMe === undefined
            ? { email, password }
            : { email, password, remember_me: rememberMe },
        credentials: "include",
        skipAuth: true,
      });
    },
    pinLogin(userId: string, pin: string) {
      return request<AuthPayload>("/auth/pin-login", {
        method: "POST",
        body: { user_id: userId, pin },
        credentials: "include",
        skipAuth: true,
      });
    },
    async logout() {
      const result = await request("/auth/logout", {
        method: "POST",
        credentials: "include",
        skipAuth: true,
      });
      clearTokens();
      return result;
    },
  };
}

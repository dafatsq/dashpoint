import type { ApiTransport } from "./transport";
import type { ApiResponse } from "./types";

export interface SetupStatus {
  setup_required: boolean;
}

export interface CreateInitialOwnerInput {
  name: string;
  email: string;
  password: string;
  pin: string;
}

export function createSetupApi(transport: ApiTransport) {
  const request = transport.request.bind(transport);

  return {
    async getSetupStatus(): Promise<ApiResponse<SetupStatus>> {
      return request<SetupStatus>("/setup/status", { skipAuth: true });
    },
    async createInitialOwner(
      input: CreateInitialOwnerInput,
    ): Promise<ApiResponse<{ message: string }>> {
      return request<{ message: string }>("/setup/owner", {
        method: "POST",
        body: {
          name: input.name,
          email: input.email,
          password: input.password,
          pin: input.pin,
        },
        skipAuth: true,
      });
    },
  };
}

export type SetupApi = ReturnType<typeof createSetupApi>;

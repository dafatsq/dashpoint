import { API_BASE_URL } from "@/lib/config";

import { createAuditApi } from "./audit";
import { createAuthApi } from "./auth";
import { createCatalogApi } from "./catalog";
import { createExpensesApi } from "./expenses";
import { createOperationsApi } from "./operations";
import { createReportsApi } from "./reports";
import { ApiTransport } from "./transport";
import type {
  AuditApi,
  AuthApi,
  CatalogApi,
  ExpensesApi,
  OperationsApi,
  ReportsApi,
  UserApi,
} from "./types";
import { createUserApi } from "./users";

export type ApiClient = AuthApi &
  UserApi &
  CatalogApi &
  OperationsApi &
  ReportsApi &
  AuditApi &
  ExpensesApi & {
    request: ApiTransport["request"];
    setTokens: ApiTransport["setTokens"];
    clearTokens: ApiTransport["clearTokens"];
    refreshTokens: ApiTransport["refreshTokens"];
  };

const transport = new ApiTransport(API_BASE_URL);

export const api: ApiClient = {
  ...createAuthApi(transport),
  ...createUserApi(transport),
  ...createCatalogApi(transport),
  ...createOperationsApi(transport),
  ...createReportsApi(transport),
  ...createAuditApi(transport),
  ...createExpensesApi(transport),
};

export * from "./types";
export default api;

import type {
  AuditLog,
  CashDrawerOperation,
  CashDrawerOperationsResponse,
  Category,
  CategorySales,
  CreateExpenseRequest,
  CreateProductRequest,
  CreateSaleRequest,
  CreateUserRequest,
  DailyReport,
  DailySummary,
  EmployeeSales,
  Expense,
  ExpenseCategory,
  ExpenseSummary,
  InventoryAdjustment,
  ProductInventoryDetails,
  InventoryValuation,
  LowStockItem,
  Product,
  Sale,
  SalesRangeReport,
  Shift,
  TopSeller,
  UpdateExpenseRequest,
  UpdateProductRequest,
  UpdateUserRequest,
  User,
} from "@/types";

export interface RequestOptions {
  method?: "GET" | "POST" | "PATCH" | "PUT" | "DELETE";
  body?: unknown;
  headers?: Record<string, string>;
}

export interface ApiResponse<T> {
  data?: T;
  error?: string;
  message?: string;
  total?: number;
  page?: number;
  per_page?: number;
  total_pages?: number;
  limit?: number;
  offset?: number;
}

export interface OffsetPaginationParams {
  limit?: number;
  offset?: number;
}

export interface DateRangeFilterParams {
  from?: string;
  to?: string;
}

export interface UserScopedFilterParams {
  user_id?: string;
  opened_by_id?: string;
}

export interface ProductsResponse {
  products: Product[];
  total: number;
  page: number;
  per_page: number;
  total_pages: number;
}

export interface CategoriesResponse {
  categories: Category[];
}

export interface LowStockResponse {
  products: LowStockItem[];
  count: number;
}

export interface UsersResponse {
  users: User[];
  total: number;
  page: number;
  per_page: number;
  total_pages: number;
}

export interface SalesResponse {
  sales: Sale[];
  total: number;
  limit: number;
  offset: number;
}

export interface ShiftsResponse {
  shifts: Shift[];
  total: number;
}

export interface TopSellersResponse {
  top_sellers: TopSeller[];
  start_date: string;
  end_date: string;
  limit: number;
}

export interface InventoryValuationResponse {
  valuation: InventoryValuation;
}

export interface AuditLogsResponse {
  logs: AuditLog[];
  total: number;
}

export interface AuthApi {
  getMe(): Promise<ApiResponse<User>>;
  login(
    email: string,
    password: string,
  ): Promise<ApiResponse<import("@/lib/auth-user").AuthPayload>>;
  pinLogin(
    userId: string,
    pin: string,
  ): Promise<ApiResponse<import("@/lib/auth-user").AuthPayload>>;
  logout(): Promise<ApiResponse<unknown>>;
}

export interface UserApi {
  getUsersPage(params?: {
    role?: string;
    active?: boolean;
    search?: string;
    page?: number;
    per_page?: number;
  }): Promise<ApiResponse<User[]>>;
  getUsers(params?: {
    role?: string;
    active?: boolean;
  }): Promise<ApiResponse<User[]>>;
  getBasicUsers(): Promise<ApiResponse<{ id: string; name: string }[]>>;
  getUser(id: string): Promise<ApiResponse<User>>;
  createUser(user: CreateUserRequest): Promise<ApiResponse<User>>;
  updateUser(id: string, user: UpdateUserRequest): Promise<ApiResponse<User>>;
  deleteUser(id: string): Promise<ApiResponse<unknown>>;
  permanentDeleteUser(id: string): Promise<ApiResponse<unknown>>;
  getRoles(): Promise<
    ApiResponse<{ id: string; name: string; description: string }[]>
  >;
}

export interface CatalogApi {
  getProductsPage(params?: {
    category_id?: string;
    active?: boolean;
    search?: string;
    page?: number;
    per_page?: number;
  }): Promise<ApiResponse<Product[]>>;
  getProducts(params?: {
    category_id?: string;
    active?: boolean;
    search?: string;
  }): Promise<ApiResponse<Product[]>>;
  getProduct(id: string): Promise<ApiResponse<Product>>;
  getProductInventory(id: string): Promise<ApiResponse<ProductInventoryDetails>>;
  lookupProduct(code: string): Promise<ApiResponse<Product>>;
  createProduct(product: CreateProductRequest): Promise<ApiResponse<Product>>;
  updateProduct(
    id: string,
    product: UpdateProductRequest,
  ): Promise<ApiResponse<Product>>;
  deleteProduct(id: string): Promise<ApiResponse<unknown>>;
  permanentDeleteProduct(id: string): Promise<ApiResponse<unknown>>;
  getCategories(status?: string): Promise<ApiResponse<Category[]>>;
  createCategory(category: {
    name: string;
    description?: string;
  }): Promise<ApiResponse<Category>>;
  updateCategory(
    id: string,
    category: { name?: string; description?: string; is_active?: boolean },
  ): Promise<ApiResponse<Category>>;
  deleteCategory(id: string): Promise<ApiResponse<unknown>>;
  permanentDeleteCategory(id: string): Promise<ApiResponse<unknown>>;
  getLowStock(threshold?: number): Promise<ApiResponse<LowStockItem[]>>;
  adjustInventory(
    adjustment: InventoryAdjustment,
  ): Promise<ApiResponse<unknown>>;
}

export interface OperationsApi {
  getCurrentShift(): Promise<ApiResponse<Shift | null>>;
  startShift(startingCash: number | string): Promise<ApiResponse<Shift>>;
  closeShift(
    closingCash: number | string,
    notes?: string,
  ): Promise<ApiResponse<Shift>>;
  getShifts(
    params?: UserScopedFilterParams &
      DateRangeFilterParams &
      OffsetPaginationParams,
  ): Promise<ApiResponse<Shift[]>>;
  payIn(
    amount: string,
    reason: string,
  ): Promise<ApiResponse<CashDrawerOperation>>;
  payOut(
    amount: string,
    reason: string,
  ): Promise<ApiResponse<CashDrawerOperation>>;
  getShiftOperations(
    shiftId: string,
  ): Promise<ApiResponse<CashDrawerOperationsResponse>>;
  createSale(sale: CreateSaleRequest): Promise<ApiResponse<Sale>>;
  getSalesPage(
    params?: DateRangeFilterParams &
      UserScopedFilterParams &
      OffsetPaginationParams & {
        status?: string;
        invoice_no?: string;
      },
  ): Promise<ApiResponse<Sale[]>>;
  getSales(
    params?: DateRangeFilterParams &
      UserScopedFilterParams & {
        status?: string;
      },
  ): Promise<ApiResponse<Sale[]>>;
  getSale(id: string): Promise<ApiResponse<Sale>>;
  voidSale(id: string, reason: string): Promise<ApiResponse<unknown>>;
  getDailySummary(date?: string): Promise<ApiResponse<DailySummary>>;
}

export interface ReportsApi {
  getDailyReport(date?: string): Promise<ApiResponse<DailyReport>>;
  getSalesRangeReport(params: {
    start_date: string;
    end_date: string;
  }): Promise<ApiResponse<SalesRangeReport>>;
  getTopSellers(params?: {
    from?: string;
    to?: string;
    limit?: number;
  }): Promise<ApiResponse<TopSeller[]>>;
  getInventoryReport(
    includeItems?: boolean,
  ): Promise<ApiResponse<InventoryValuation>>;
  getCashReport(params?: {
    start_date?: string;
    end_date?: string;
  }): Promise<ApiResponse<import("@/types").CashReport>>;
  getEmployeeSalesReport(params?: {
    start_date?: string;
    end_date?: string;
  }): Promise<ApiResponse<EmployeeSales[]>>;
  getCategorySalesReport(params?: {
    start_date?: string;
    end_date?: string;
  }): Promise<ApiResponse<CategorySales[]>>;
  exportSalesCSV(params: {
    start_date: string;
    end_date: string;
  }): Promise<string>;
  exportInventoryCSV(): Promise<string>;
  exportTopSellersCSV(params?: {
    start_date?: string;
    end_date?: string;
    limit?: number;
  }): Promise<string>;
  exportComprehensiveReportCSV(params: {
    start_date: string;
    end_date: string;
  }): Promise<string>;
}

export interface AuditApi {
  getAuditLogs(
    params?: UserScopedFilterParams &
      DateRangeFilterParams &
      OffsetPaginationParams & {
        action?: string;
        entity_type?: string;
        search?: string;
      },
  ): Promise<ApiResponse<AuditLog[]>>;
  getDashboardChanges(
    params?: UserScopedFilterParams &
      DateRangeFilterParams &
      OffsetPaginationParams & {
        entity_type?: string;
      },
  ): Promise<ApiResponse<AuditLog[]>>;
}

export interface ExpensesApi {
  getExpenseCategories(
    status?: string,
  ): Promise<ApiResponse<ExpenseCategory[]>>;
  createExpenseCategory(
    name: string,
    description?: string,
  ): Promise<ApiResponse<ExpenseCategory>>;
  updateExpenseCategory(
    id: string,
    category: { name?: string; description?: string; is_active?: boolean },
  ): Promise<ApiResponse<ExpenseCategory>>;
  deleteExpenseCategory(id: string): Promise<ApiResponse<void>>;
  permanentDeleteExpenseCategory(id: string): Promise<ApiResponse<void>>;
  getExpenses(
    params?: OffsetPaginationParams & {
      category_id?: string;
      start_date?: string;
      end_date?: string;
    },
  ): Promise<ApiResponse<{ expenses: Expense[]; total: number }>>;
  getExpense(id: string): Promise<ApiResponse<Expense>>;
  createExpense(expense: CreateExpenseRequest): Promise<ApiResponse<Expense>>;
  updateExpense(
    id: string,
    expense: UpdateExpenseRequest,
  ): Promise<ApiResponse<Expense>>;
  deleteExpense(id: string): Promise<ApiResponse<void>>;
  getExpenseSummary(params?: {
    start_date?: string;
    end_date?: string;
  }): Promise<ApiResponse<ExpenseSummary>>;
  getMonthlyExpenses(
    months?: number,
  ): Promise<ApiResponse<{ month: string; total: string }[]>>;
}

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

import { Permission, PermissionOverride } from '@/types';

interface RequestOptions {
  method?: 'GET' | 'POST' | 'PATCH' | 'PUT' | 'DELETE';
  body?: unknown;
  headers?: Record<string, string>;
}

export interface ApiResponse<T> {
  data?: T;
  error?: string;
  message?: string;
}

// Response wrapper types matching backend
interface ProductsResponse {
  products: import('@/types').Product[];
  total: number;
  page: number;
  per_page: number;
  total_pages: number;
}

interface CategoriesResponse {
  categories: import('@/types').Category[];
}

interface LowStockResponse {
  products: import('@/types').LowStockItem[];
  count: number;
}

interface UsersResponse {
  users: import('@/types').User[];
  total: number;
}

interface SalesResponse {
  sales: import('@/types').Sale[];
  total: number;
}

interface ShiftsResponse {
  shifts: import('@/types').Shift[];
}

interface TopSellersResponse {
  top_sellers: import('@/types').TopSeller[];
  start_date: string;
  end_date: string;
  limit: number;
}

interface InventoryValuationResponse {
  valuation: import('@/types').InventoryValuation;
}

interface AuditLogsResponse {
  logs: import('@/types').AuditLog[];
  total: number;
}

class ApiClient {
  private baseUrl: string;

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl;
  }

  private getAccessToken(): string | null {
    if (typeof window === 'undefined') return null;
    return localStorage.getItem('access_token');
  }

  private getRefreshToken(): string | null {
    if (typeof window === 'undefined') return null;
    return localStorage.getItem('refresh_token');
  }

  setTokens(accessToken: string, refreshToken: string): void {
    if (typeof window === 'undefined') return;
    localStorage.setItem('access_token', accessToken);
    localStorage.setItem('refresh_token', refreshToken);
  }

  clearTokens(): void {
    if (typeof window === 'undefined') return;
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    localStorage.removeItem('user');
  }

  async refreshTokens(): Promise<boolean> {
    const refreshToken = this.getRefreshToken();
    if (!refreshToken) return false;

    try {
      const response = await fetch(`${this.baseUrl}/auth/refresh`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refresh_token: refreshToken }),
      });

      if (!response.ok) {
        this.clearTokens();
        return false;
      }

      const data = await response.json();
      this.setTokens(data.access_token, data.refresh_token);

      // Also update user data in localStorage if returned
      // The backend returns updated user info with new role/permissions
      if (data.user) {
        console.log('[API] Updating user data from refresh response:', data.user.role_name);
        const userData = {
          id: data.user.id,
          email: data.user.email || undefined,
          name: data.user.name,
          role_id: data.user.role_id,
          role_name: data.user.role_name,
          is_active: data.user.is_active,
          has_pin: data.user.has_pin || false,
          permissions: data.user.permissions || [],
          created_at: data.user.created_at || '',
          updated_at: data.user.updated_at || '',
        };
        localStorage.setItem('user', JSON.stringify(userData));
      }

      return true;
    } catch {
      this.clearTokens();
      return false;
    }
  }

  private async refreshAccessToken(): Promise<boolean> {
    return this.refreshTokens();
  }

  async request<T>(endpoint: string, options: RequestOptions = {}): Promise<ApiResponse<T>> {
    const { method = 'GET', body, headers = {} } = options;

    const requestHeaders: Record<string, string> = {
      'Content-Type': 'application/json',
      ...headers,
    };

    const accessToken = this.getAccessToken();
    if (accessToken) {
      requestHeaders['Authorization'] = `Bearer ${accessToken}`;
    }

    try {
      let response = await fetch(`${this.baseUrl}${endpoint}`, {
        method,
        headers: requestHeaders,
        body: body ? JSON.stringify(body) : undefined,
        cache: 'no-store', // Prevent caching of API responses
      });

      // If unauthorized, try to refresh token
      if (response.status === 401 && accessToken) {
        const refreshed = await this.refreshAccessToken();
        if (refreshed) {
          requestHeaders['Authorization'] = `Bearer ${this.getAccessToken()}`;
          response = await fetch(`${this.baseUrl}${endpoint}`, {
            method,
            headers: requestHeaders,
            body: body ? JSON.stringify(body) : undefined,
            cache: 'no-store',
          });
        } else {
          // Redirect to login
          if (typeof window !== 'undefined') {
            window.location.href = '/login';
          }
          return { error: 'Session expired. Please login again.' };
        }
      }

      const text = await response.text();
      let data;
      try {
        data = text ? JSON.parse(text) : {};
      } catch (e) {
        console.error('Failed to parse JSON response:', text);
        data = {};
      }

      if (!response.ok) {
        const errorMsg = data.error || data.message || `Request failed with status ${response.status}`;

        // Handle account deactivation - log user out immediately
        if (response.status === 401 && data.code === 'ACCOUNT_INACTIVE') {
          this.clearTokens();
          if (typeof window !== 'undefined') {
            window.location.href = '/login?message=account_inactive';
          }
          return { error: 'Your account has been deactivated' };
        }

        if (response.status >= 400 && response.status < 500) {
          console.warn('API Client Error:', { status: response.status, message: errorMsg, data });
        } else {
          console.error('API Server Error:', { status: response.status, message: errorMsg, data });
        }
        return { error: errorMsg };
      }

      return { data };
    } catch (error) {
      console.error('API request failed:', error);
      return { error: 'Network error. Please check your connection.' };
    }
  }

  async getMe(): Promise<ApiResponse<import('@/types').User>> {
    const result = await this.request<{ user: import('@/types').User }>('/me');
    if (result.error) return { error: result.error };
    return { data: result.data?.user };
  }

  // Auth endpoints
  async login(email: string, password: string) {
    return this.request<{
      access_token: string;
      refresh_token: string;
      user: {
        id: string;
        email?: string;
        name: string;
        role_id: string;
        role_name: string;
        is_active: boolean;
        has_pin?: boolean;
        permissions?: string[];
        created_at?: string;
        updated_at?: string;
      };
    }>('/auth/login', {
      method: 'POST',
      body: { email, password },
    });
  }

  async pinLogin(userId: string, pin: string) {
    return this.request<{
      access_token: string;
      refresh_token: string;
      user: {
        id: string;
        email?: string;
        name: string;
        role_id: string;
        role_name: string;
        is_active: boolean;
        has_pin?: boolean;
        permissions?: string[];
        created_at?: string;
        updated_at?: string;
      };
    }>('/auth/pin-login', {
      method: 'POST',
      body: { user_id: userId, pin },
    });
  }

  async logout() {
    const result = await this.request('/auth/logout', { method: 'POST' });
    this.clearTokens();
    return result;
  }

  // User endpoints
  async getUsers(params?: { role?: string; active?: boolean }): Promise<ApiResponse<import('@/types').User[]>> {
    const searchParams = new URLSearchParams();
    if (params?.role) searchParams.set('role', params.role);
    if (params?.active !== undefined) searchParams.set('active_only', String(params.active));
    const query = searchParams.toString();
    const result = await this.request<UsersResponse>(`/users${query ? `?${query}` : ''}`);
    if (result.error) return { error: result.error };
    return { data: result.data?.users || [] };
  }

  async getUser(id: string): Promise<ApiResponse<import('@/types').User>> {
    const result = await this.request<{ user: import('@/types').User }>(`/users/${id}`);
    if (result.error) return { error: result.error };
    return { data: result.data?.user };
  }

  async createUser(user: import('@/types').CreateUserRequest): Promise<ApiResponse<import('@/types').User>> {
    const result = await this.request<{ user: import('@/types').User }>('/users', { method: 'POST', body: user });
    if (result.error) return { error: result.error };
    return { data: result.data?.user };
  }

  async updateUser(id: string, user: import('@/types').UpdateUserRequest): Promise<ApiResponse<import('@/types').User>> {
    const result = await this.request<{ user: import('@/types').User }>(`/users/${id}`, { method: 'PATCH', body: user });
    if (result.error) return { error: result.error };
    return { data: result.data?.user };
  }

  async deleteUser(id: string) {
    return this.request(`/users/${id}`, { method: 'DELETE' });
  }

  async permanentDeleteUser(id: string) {
    return this.request(`/users/${id}/permanent`, { method: 'DELETE' });
  }

  async getRoles(): Promise<ApiResponse<{ id: string; name: string; description: string }[]>> {
    const result = await this.request<{ roles: { id: string; name: string; description: string }[] }>('/roles');
    if (result.error) return { error: result.error };
    return { data: result.data?.roles || [] };
  }

  // Permission endpoints
  async getPermissions(grouped?: boolean): Promise<ApiResponse<Permission[] | Record<string, Permission[]>>> {
    const result = await this.request<{ permissions: Permission[] | Record<string, Permission[]> }>(`/permissions${grouped ? '?grouped=true' : ''}`);
    if (result.error) return { error: result.error };
    return { data: result.data?.permissions };
  }

  async getUserPermissions(userId: string): Promise<ApiResponse<{ effective_permissions: string[]; overrides: PermissionOverride[] }>> {
    const result = await this.request<{ effective_permissions: string[]; overrides: PermissionOverride[] }>(`/users/${userId}/permissions`);
    if (result.error) return { error: result.error };
    return { data: result.data };
  }

  async setUserPermissions(userId: string, permissions: { permission_id: string; allowed: boolean }[]): Promise<ApiResponse<{ message: string; effective_permissions: string[]; overrides: number }>> {
    const result = await this.request<{ message: string; effective_permissions: string[]; overrides: number }>(`/users/${userId}/permissions`, {
      method: 'PATCH',
      body: { permissions },
    });
    if (result.error) return { error: result.error };
    return { data: result.data };
  }

  // Product endpoints
  async getProducts(params?: { category_id?: string; active?: boolean; search?: string }): Promise<ApiResponse<import('@/types').Product[]>> {
    const searchParams = new URLSearchParams();
    if (params?.category_id) searchParams.set('category_id', params.category_id);
    if (params?.active !== undefined) searchParams.set('active_only', String(params.active));
    if (params?.search) searchParams.set('search', params.search);
    const query = searchParams.toString();
    const result = await this.request<ProductsResponse>(`/products${query ? `?${query}` : ''}`);
    if (result.error) return { error: result.error };
    return { data: result.data?.products || [] };
  }

  async getProduct(id: string): Promise<ApiResponse<import('@/types').Product>> {
    const result = await this.request<{ product: import('@/types').Product }>(`/products/${id}`);
    if (result.error) return { error: result.error };
    return { data: result.data?.product };
  }

  async lookupProduct(code: string): Promise<ApiResponse<import('@/types').Product>> {
    const result = await this.request<{ product: import('@/types').Product }>(`/products/lookup?code=${encodeURIComponent(code)}`);
    if (result.error) return { error: result.error };
    return { data: result.data?.product };
  }

  async createProduct(product: import('@/types').CreateProductRequest): Promise<ApiResponse<import('@/types').Product>> {
    const result = await this.request<{ product: import('@/types').Product }>('/products', { method: 'POST', body: product });
    if (result.error) return { error: result.error };
    return { data: result.data?.product };
  }

  async updateProduct(id: string, product: import('@/types').UpdateProductRequest): Promise<ApiResponse<import('@/types').Product>> {
    const result = await this.request<{ product: import('@/types').Product }>(`/products/${id}`, { method: 'PATCH', body: product });
    if (result.error) return { error: result.error };
    return { data: result.data?.product };
  }

  async deleteProduct(id: string) {
    return this.request(`/products/${id}`, { method: 'DELETE' });
  }

  async permanentDeleteProduct(id: string) {
    return this.request(`/products/${id}/permanent`, { method: 'DELETE' });
  }

  // Category endpoints
  async getCategories(): Promise<ApiResponse<import('@/types').Category[]>> {
    const result = await this.request<CategoriesResponse>('/categories');
    if (result.error) return { error: result.error };
    return { data: result.data?.categories || [] };
  }

  async createCategory(category: { name: string; description?: string }): Promise<ApiResponse<import('@/types').Category>> {
    const result = await this.request<{ category: import('@/types').Category }>('/categories', { method: 'POST', body: category });
    if (result.error) return { error: result.error };
    return { data: result.data?.category };
  }

  async updateCategory(id: string, category: { name?: string; description?: string }): Promise<ApiResponse<import('@/types').Category>> {
    const result = await this.request<{ category: import('@/types').Category }>(`/categories/${id}`, { method: 'PATCH', body: category });
    if (result.error) return { error: result.error };
    return { data: result.data?.category };
  }

  async deleteCategory(id: string) {
    return this.request(`/categories/${id}`, { method: 'DELETE' });
  }

  // Inventory endpoints
  async getLowStock(threshold?: number): Promise<ApiResponse<import('@/types').LowStockItem[]>> {
    const query = threshold ? `?threshold=${threshold}` : '';
    const result = await this.request<LowStockResponse>(`/inventory/low-stock${query}`);
    if (result.error) return { error: result.error };
    return { data: result.data?.products || [] };
  }

  async adjustInventory(adjustment: import('@/types').InventoryAdjustment) {
    return this.request('/inventory/adjust', { method: 'POST', body: adjustment });
  }

  // Shift endpoints
  async getCurrentShift(): Promise<ApiResponse<import('@/types').Shift | null>> {
    const result = await this.request<{ shift: import('@/types').Shift }>('/shifts/current');
    if (result.error) {
      if (result.error.includes('not found') || result.error.includes('NO_ACTIVE_SHIFT')) {
        return { data: null };
      }
      return { error: result.error };
    }
    return { data: result.data?.shift };
  }

  async startShift(startingCash: number | string): Promise<ApiResponse<import('@/types').Shift>> {
    const result = await this.request<{ shift: import('@/types').Shift }>('/shifts/start', {
      method: 'POST',
      body: { opening_cash: String(startingCash) },
    });
    if (result.error) return { error: result.error };
    return { data: result.data?.shift };
  }

  async closeShift(closingCash: number | string, notes?: string): Promise<ApiResponse<import('@/types').Shift>> {
    const result = await this.request<{ shift: import('@/types').Shift }>('/shifts/close', {
      method: 'POST',
      body: { closing_cash: String(closingCash), notes },
    });
    if (result.error) return { error: result.error };
    return { data: result.data?.shift };
  }

  async getShifts(params?: { user_id?: string; from?: string; to?: string }): Promise<ApiResponse<import('@/types').Shift[]>> {
    const searchParams = new URLSearchParams();
    if (params?.user_id) searchParams.set('user_id', params.user_id);
    if (params?.from) searchParams.set('from', params.from);
    if (params?.to) searchParams.set('to', params.to);
    const query = searchParams.toString();
    const result = await this.request<ShiftsResponse>(`/shifts${query ? `?${query}` : ''}`);
    if (result.error) return { error: result.error };
    return { data: result.data?.shifts || [] };
  }

  // Sales endpoints
  async createSale(sale: import('@/types').CreateSaleRequest): Promise<ApiResponse<import('@/types').Sale>> {
    const result = await this.request<{ sale: import('@/types').Sale }>('/sales', { method: 'POST', body: sale });
    if (result.error) return { error: result.error };
    return { data: result.data?.sale };
  }

  async getSales(params?: { from?: string; to?: string; user_id?: string; status?: string }): Promise<ApiResponse<import('@/types').Sale[]>> {
    const searchParams = new URLSearchParams();
    if (params?.from) searchParams.set('from', params.from);
    if (params?.to) searchParams.set('to', params.to);
    if (params?.user_id) searchParams.set('user_id', params.user_id);
    if (params?.status) searchParams.set('status', params.status);
    const query = searchParams.toString();
    const result = await this.request<SalesResponse>(`/sales${query ? `?${query}` : ''}`);
    if (result.error) return { error: result.error };
    return { data: result.data?.sales || [] };
  }

  async getSale(id: string): Promise<ApiResponse<import('@/types').Sale>> {
    const result = await this.request<{ sale: import('@/types').Sale }>(`/sales/${id}`);
    if (result.error) return { error: result.error };
    return { data: result.data?.sale };
  }

  async voidSale(id: string, reason: string) {
    return this.request(`/sales/${id}/void`, { method: 'POST', body: { reason } });
  }

  async getDailySummary(date?: string): Promise<ApiResponse<import('@/types').DailySummary>> {
    const query = date ? `?date=${date}` : '';
    const result = await this.request<{ summary: import('@/types').DailySummary }>(`/sales/summary/daily${query}`);
    if (result.error) return { error: result.error };
    return { data: result.data?.summary };
  }

  // Report endpoints
  async getDailyReport(date?: string): Promise<ApiResponse<import('@/types').DailyReport>> {
    const query = date ? `?date=${date}` : '';
    const result = await this.request<{ report: import('@/types').DailyReport }>(`/reports/daily${query}`);
    if (result.error) return { error: result.error };
    return { data: result.data?.report };
  }

  async getSalesRangeReport(params: { start_date: string; end_date: string }): Promise<ApiResponse<import('@/types').SalesRangeReport>> {
    const searchParams = new URLSearchParams();
    searchParams.set('start_date', params.start_date);
    searchParams.set('end_date', params.end_date);
    const result = await this.request<import('@/types').SalesRangeReport>(`/reports/sales?${searchParams.toString()}`);
    if (result.error) return { error: result.error };
    return { data: result.data };
  }

  async getTopSellers(params?: { from?: string; to?: string; limit?: number }): Promise<ApiResponse<import('@/types').TopSeller[]>> {
    const searchParams = new URLSearchParams();
    if (params?.from) searchParams.set('start_date', params.from);
    if (params?.to) searchParams.set('end_date', params.to);
    if (params?.limit) searchParams.set('limit', String(params.limit));
    const query = searchParams.toString();
    const result = await this.request<TopSellersResponse>(`/reports/top-sellers${query ? `?${query}` : ''}`);
    if (result.error) return { error: result.error };
    return { data: result.data?.top_sellers || [] };
  }

  async getInventoryReport(includeItems?: boolean): Promise<ApiResponse<import('@/types').InventoryValuation>> {
    const query = includeItems ? '?include_items=true' : '';
    const result = await this.request<InventoryValuationResponse>(`/reports/inventory${query}`);
    if (result.error) return { error: result.error };
    return { data: result.data?.valuation };
  }

  async getCashReport(date?: string): Promise<ApiResponse<import('@/types').CashReport>> {
    const query = date ? `?date=${date}` : '';
    const result = await this.request<{ report: import('@/types').CashReport }>(`/reports/cash${query}`);
    if (result.error) return { error: result.error };
    return { data: result.data?.report };
  }

  async getEmployeeSalesReport(params?: { start_date?: string; end_date?: string }): Promise<ApiResponse<import('@/types').EmployeeSales[]>> {
    const searchParams = new URLSearchParams();
    if (params?.start_date) searchParams.set('start_date', params.start_date);
    if (params?.end_date) searchParams.set('end_date', params.end_date);
    const query = searchParams.toString();
    const result = await this.request<{ employees: import('@/types').EmployeeSales[] }>(`/reports/by-employee${query ? `?${query}` : ''}`);
    if (result.error) return { error: result.error };
    return { data: result.data?.employees || [] };
  }

  async getCategorySalesReport(params?: { start_date?: string; end_date?: string }): Promise<ApiResponse<import('@/types').CategorySales[]>> {
    const searchParams = new URLSearchParams();
    if (params?.start_date) searchParams.set('start_date', params.start_date);
    if (params?.end_date) searchParams.set('end_date', params.end_date);
    const query = searchParams.toString();
    const result = await this.request<{ categories: import('@/types').CategorySales[] }>(`/reports/by-category${query ? `?${query}` : ''}`);
    if (result.error) return { error: result.error };
    return { data: result.data?.categories || [] };
  }

  // Export endpoints - returns blob URL for download
  async exportSalesCSV(params: { start_date: string; end_date: string }): Promise<string> {
    const token = this.getAccessToken();
    const searchParams = new URLSearchParams();
    searchParams.set('start_date', params.start_date);
    searchParams.set('end_date', params.end_date);
    const response = await fetch(`${this.baseUrl}/reports/export/sales?${searchParams.toString()}`, {
      headers: { Authorization: `Bearer ${token}` }
    });
    if (!response.ok) {
      throw new Error(`Export failed: ${response.statusText}`);
    }
    const blob = await response.blob();
    return URL.createObjectURL(blob);
  }

  async exportInventoryCSV(): Promise<string> {
    const token = this.getAccessToken();
    const response = await fetch(`${this.baseUrl}/reports/export/inventory`, {
      headers: { Authorization: `Bearer ${token}` }
    });
    if (!response.ok) {
      throw new Error(`Export failed: ${response.statusText}`);
    }
    const blob = await response.blob();
    return URL.createObjectURL(blob);
  }

  async exportTopSellersCSV(params?: { start_date?: string; end_date?: string; limit?: number }): Promise<string> {
    const token = this.getAccessToken();
    const searchParams = new URLSearchParams();
    if (params?.start_date) searchParams.set('start_date', params.start_date);
    if (params?.end_date) searchParams.set('end_date', params.end_date);
    if (params?.limit) searchParams.set('limit', String(params.limit));
    const query = searchParams.toString();
    const response = await fetch(`${this.baseUrl}/reports/export/top-sellers${query ? `?${query}` : ''}`, {
      headers: { Authorization: `Bearer ${token}` }
    });
    if (!response.ok) {
      throw new Error(`Export failed: ${response.statusText}`);
    }
    const blob = await response.blob();
    return URL.createObjectURL(blob);
  }

  async exportComprehensiveReportCSV(params: { start_date: string; end_date: string }): Promise<string> {
    const token = this.getAccessToken();
    const searchParams = new URLSearchParams();
    searchParams.set('start_date', params.start_date);
    searchParams.set('end_date', params.end_date);
    const response = await fetch(`${this.baseUrl}/reports/export/comprehensive?${searchParams.toString()}`, {
      headers: { Authorization: `Bearer ${token}` }
    });
    if (!response.ok) {
      throw new Error(`Export failed: ${response.statusText}`);
    }
    const blob = await response.blob();
    return URL.createObjectURL(blob);
  }

  // Audit log endpoints
  async getAuditLogs(params?: {
    user_id?: string;
    action?: string;
    entity_type?: string;
    from?: string;
    to?: string;
    limit?: number;
    offset?: number;
  }): Promise<ApiResponse<import('@/types').AuditLog[]>> {
    const searchParams = new URLSearchParams();
    if (params?.user_id) searchParams.set('user_id', params.user_id);
    if (params?.action) searchParams.set('action', params.action);
    if (params?.entity_type) searchParams.set('entity_type', params.entity_type);
    if (params?.from) searchParams.set('from', params.from);
    if (params?.to) searchParams.set('to', params.to);
    if (params?.limit) searchParams.set('limit', String(params.limit));
    if (params?.offset) searchParams.set('offset', String(params.offset));
    const query = searchParams.toString();
    const result = await this.request<AuditLogsResponse>(`/logs${query ? `?${query}` : ''}`);
    if (result.error) return { error: result.error };
    return { data: result.data?.logs || [] };
  }

  // Expense endpoints
  async getExpenseCategories(): Promise<ApiResponse<import('@/types').ExpenseCategory[]>> {
    const result = await this.request<{ data: import('@/types').ExpenseCategory[] }>('/expenses/categories');
    if (result.error) return { error: result.error };
    return { data: result.data?.data || [] };
  }

  async createExpenseCategory(name: string, description?: string): Promise<ApiResponse<import('@/types').ExpenseCategory>> {
    const result = await this.request<{ data: import('@/types').ExpenseCategory }>('/expenses/categories', {
      method: 'POST',
      body: { name, description },
    });
    if (result.error) return { error: result.error };
    return { data: result.data?.data };
  }

  async getExpenses(params?: {
    category_id?: string;
    start_date?: string;
    end_date?: string;
    limit?: number;
    offset?: number;
  }): Promise<ApiResponse<{ expenses: import('@/types').Expense[]; total: number }>> {
    const searchParams = new URLSearchParams();
    if (params?.category_id) searchParams.set('category_id', params.category_id);
    if (params?.start_date) searchParams.set('start_date', params.start_date);
    if (params?.end_date) searchParams.set('end_date', params.end_date);
    if (params?.limit) searchParams.set('limit', String(params.limit));
    if (params?.offset) searchParams.set('offset', String(params.offset));
    const query = searchParams.toString();
    const result = await this.request<{ data: import('@/types').Expense[]; total: number }>(`/expenses${query ? `?${query}` : ''}`);
    if (result.error) return { error: result.error };
    return { data: { expenses: result.data?.data || [], total: result.data?.total || 0 } };
  }

  async getExpense(id: string): Promise<ApiResponse<import('@/types').Expense>> {
    const result = await this.request<{ data: import('@/types').Expense }>(`/expenses/${id}`);
    if (result.error) return { error: result.error };
    return { data: result.data?.data };
  }

  async createExpense(expense: import('@/types').CreateExpenseRequest): Promise<ApiResponse<import('@/types').Expense>> {
    const result = await this.request<{ data: import('@/types').Expense }>('/expenses', {
      method: 'POST',
      body: expense,
    });
    if (result.error) return { error: result.error };
    return { data: result.data?.data };
  }

  async updateExpense(id: string, expense: import('@/types').UpdateExpenseRequest): Promise<ApiResponse<import('@/types').Expense>> {
    const result = await this.request<{ data: import('@/types').Expense }>(`/expenses/${id}`, {
      method: 'PATCH',
      body: expense,
    });
    if (result.error) return { error: result.error };
    return { data: result.data?.data };
  }

  async deleteExpense(id: string): Promise<ApiResponse<void>> {
    return this.request(`/expenses/${id}`, { method: 'DELETE' });
  }

  async getExpenseSummary(params?: { start_date?: string; end_date?: string }): Promise<ApiResponse<import('@/types').ExpenseSummary>> {
    const searchParams = new URLSearchParams();
    if (params?.start_date) searchParams.set('start_date', params.start_date);
    if (params?.end_date) searchParams.set('end_date', params.end_date);
    const query = searchParams.toString();
    const result = await this.request<{ data: import('@/types').ExpenseSummary }>(`/expenses/summary${query ? `?${query}` : ''}`);
    if (result.error) return { error: result.error };
    return { data: result.data?.data };
  }

  async getMonthlyExpenses(months?: number): Promise<ApiResponse<{ month: string; total: string }[]>> {
    const query = months ? `?months=${months}` : '';
    const result = await this.request<{ data: { month: string; total: string }[] }>(`/expenses/monthly${query}`);
    if (result.error) return { error: result.error };
    return { data: result.data?.data || [] };
  }
}

export const api = new ApiClient(API_BASE_URL);
export default api;

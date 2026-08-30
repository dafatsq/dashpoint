export type UserRole = "owner" | "manager" | "cashier";

export interface Permission {
  id: string;
  key: string;
  name: string;
  description?: string;
  category: string;
}

export interface PermissionOverride {
  permission_id: string;
  permission_key: string;
  permission_name: string;
  allowed: boolean;
  granted_by?: string;
  granted_by_name?: string;
  created_at?: string;
}

export interface Role {
  id: string;
  name: string;
  description?: string;
  permissions?: string[];
}

export interface User {
  id: string;
  email?: string;
  name: string;
  role_id: string;
  role_name: UserRole;
  role?: Role;
  is_active: boolean;
  has_pin: boolean;
  permissions?: string[];
  last_login_at?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateUserRequest {
  email: string;
  password: string;
  name: string;
  role?: UserRole;
  role_id?: string;
  pin?: string;
  permissions?: string[];
}

export interface UpdateUserRequest {
  email?: string;
  name?: string;
  role?: UserRole;
  role_id?: string;
  is_active?: boolean;
  password?: string;
  pin?: string;
  current_password?: string;
  current_pin?: string;
  permissions?: string[];
  expected_updated_at?: string;
}

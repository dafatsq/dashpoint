import type { Role, User, UserRole } from "@/types";

type StoredUser = Partial<User> & {
  role?: UserRole | string;
  role_name?: UserRole | string;
};

export interface ApiUserPayload {
  id: string;
  email?: string;
  name: string;
  role_id?: string;
  role_name?: UserRole | string;
  role?: UserRole | string | Role;
  is_active: boolean;
  has_pin?: boolean;
  permissions?: string[];
  created_at?: string;
  updated_at?: string;
}

export interface AuthPayload {
  access_token: string;
  refresh_token: string;
  user?: ApiUserPayload;
}

type NormalizableUser = ApiUserPayload | User;

function resolveRoleName(user: ApiUserPayload | StoredUser | User): UserRole {
  const roleName = user.role_name ?? user.role;
  return (typeof roleName === "string" ? roleName : "cashier") as UserRole;
}

export function normalizeUser(user: NormalizableUser): User {
  return {
    id: user.id,
    email: user.email || undefined,
    name: user.name,
    role_id: user.role_id || "",
    role_name: resolveRoleName(user),
    is_active: user.is_active,
    has_pin: user.has_pin || false,
    permissions: user.permissions || [],
    created_at: user.created_at || "",
    updated_at: user.updated_at || "",
  };
}

export function hydrateStoredUser(storedUser: string): User | null {
  try {
    const parsed = JSON.parse(storedUser) as StoredUser;

    if (!parsed.id || !parsed.name) {
      return null;
    }

    return {
      id: parsed.id,
      email: parsed.email || undefined,
      name: parsed.name,
      role_id: parsed.role_id || "",
      role_name: resolveRoleName(parsed),
      role:
        parsed.role && typeof parsed.role !== "string" ? parsed.role : undefined,
      is_active: parsed.is_active ?? true,
      has_pin: parsed.has_pin || false,
      permissions: parsed.permissions || [],
      last_login_at: parsed.last_login_at,
      created_at: parsed.created_at || "",
      updated_at: parsed.updated_at || "",
    };
  } catch {
    return null;
  }
}

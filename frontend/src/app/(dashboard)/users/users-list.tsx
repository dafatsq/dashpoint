"use client";

import {
  Archive,
  ChevronLeft,
  ChevronRight,
  Loader2,
  Pencil,
  Plus,
  RotateCcw,
  Search,
  Shield,
  ShieldAlert,
  ShieldCheck,
  Trash2,
  Users,
} from "lucide-react";

import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import type { User, UserRole } from "@/types";
import { ActiveArchivedToggle } from "@/components/shared/active-archived-toggle";
import { DataTableContainer } from "@/components/shared/data-table-container";
import { FilterCard } from "@/components/shared/filter-card";
import { DataSortSelect } from "@/components/shared/data-sort-select";

export const USER_SORT_OPTIONS = [
  { value: "created_at_desc", label: "Newest added" },
  { value: "created_at_asc", label: "Oldest added" },
  { value: "name_asc", label: "Name (A-Z)" },
  { value: "name_desc", label: "Name (Z-A)" },
  { value: "email_asc", label: "Email (A-Z)" },
  { value: "role_asc", label: "Role (A-Z)" },
] as const;

interface UsersListProps {
  pageError?: string | null;
  users: User[];
  viewMode: "active" | "archived";
  isLoading: boolean;
  canCreateUser: boolean;
  canEditUserAny: boolean;
  canDeleteUserAny: boolean;
  searchQuery: string;
  selectedRole: string;
  sort: string;
  page: number;
  limit: number;
  total: number;
  hasMore: boolean;
  isSubmitting: boolean;
  onCreate: () => void;
  onSearchChange: (value: string) => void;
  onRoleChange: (value: string) => void;
  onSortChange: (value: string) => void;
  onViewModeChange: (value: "active" | "archived") => void;
  onLimitChange: (value: number) => void;
  onPageChange: (value: number) => void;
  canEditUser: (user: User) => boolean;
  canDeleteUser: (user: User) => boolean;
  onEdit: (user: User) => void;
  onArchive: (user: User) => void;
  onPermanentDelete: (user: User) => void;
  onRestore: (user: User) => void;
}

function getRoleIcon(role: UserRole) {
  switch (role) {
    case "owner":
      return <ShieldAlert className="h-4 w-4" />;
    case "manager":
      return <ShieldCheck className="h-4 w-4" />;
    case "cashier":
      return <Shield className="h-4 w-4" />;
  }
}

function getRoleBadgeColor(role: UserRole) {
  switch (role) {
    case "owner":
      return "bg-purple-600 text-white dark:bg-purple-600/90 dark:text-white";
    case "manager":
      return "bg-blue-600 text-white dark:bg-blue-600/90 dark:text-white";
    case "cashier":
      return "bg-green-600 text-white dark:bg-green-600/90 dark:text-white";
  }
}

function getBooleanBadgeClasses(enabled: boolean) {
  return enabled
    ? "bg-green-600 text-white dark:bg-green-600/90 dark:text-white"
    : "bg-gray-600 text-white dark:bg-gray-600/90 dark:text-white";
}

export function UsersList({
  pageError,
  users,
  viewMode,
  isLoading,
  canCreateUser,
  canEditUserAny,
  canDeleteUserAny,
  searchQuery,
  selectedRole,
  sort,
  page,
  limit,
  total,
  hasMore,
  isSubmitting,
  onCreate,
  onSearchChange,
  onRoleChange,
  onSortChange,
  onViewModeChange,
  onLimitChange,
  onPageChange,
  canEditUser,
  canDeleteUser,
  onEdit,
  onArchive,
  onPermanentDelete,
  onRestore,
}: UsersListProps) {
  return (
    <div className="p-6 pt-4">
      <ActiveArchivedToggle value={viewMode} onChange={onViewModeChange} className="mb-4" />
      <FilterCard>
        {pageError && (
          <div className="mb-4 rounded-md border border-destructive/20 bg-destructive/10 p-3">
            <p className="text-sm text-destructive">{pageError}</p>
          </div>
        )}
        <div className="flex flex-wrap items-center justify-center gap-4">
          <div className="relative w-full flex-1 min-w-[200px] sm:min-w-[220px]">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Search users..."
              value={searchQuery}
              onChange={(event) => onSearchChange(event.target.value)}
              className="pl-9 w-full"
            />
          </div>
          <div className="w-full flex-1 min-w-[200px] sm:min-w-[220px]">
            <Select value={selectedRole} onValueChange={onRoleChange}>
              <SelectTrigger className="w-full">
                <SelectValue placeholder="All Roles" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Roles</SelectItem>
                <SelectItem value="owner">Owner</SelectItem>
                <SelectItem value="manager">Manager</SelectItem>
                <SelectItem value="cashier">Cashier</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <DataSortSelect value={sort} options={USER_SORT_OPTIONS} onChange={onSortChange} className="w-full flex-1 min-w-[200px] sm:min-w-[220px]" />
          {canCreateUser && viewMode === "active" && (
            <Button onClick={onCreate} className="w-full flex-1 min-w-[200px] sm:min-w-[220px]">
              <Plus className="h-4 w-4 mr-2" />
              Add User
            </Button>
          )}
        </div>
      </FilterCard>

      <DataTableContainer limit={limit} onLimitChange={onLimitChange} total={total} currentCount={users.length}>

          {isLoading ? (
            <div className="flex items-center justify-center h-64">
              <Loader2 className="h-8 w-8 animate-spin text-primary" />
            </div>
          ) : users.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12">
              <Users className="h-12 w-12 text-muted-foreground mb-4" />
              <p className="text-muted-foreground">
                {viewMode === "active" ? "No users found" : "No archived users"}
              </p>
              {canCreateUser && viewMode === "active" && (
                <Button variant="link" onClick={onCreate}>
                  Add your first user
                </Button>
              )}
            </div>
          ) : (
            <>
              <div className="hidden lg:block overflow-x-auto">
                <table className="w-full min-w-max">
                  <thead>
                    <tr className="border-b text-left text-sm text-muted-foreground">
                      <th className="pb-3 font-medium">Name</th>
                      <th className="pb-3 font-medium">Email</th>
                      <th className="pb-3 font-medium">Role</th>
                      <th className="pb-3 font-medium text-center">PIN Set</th>
                      <th className="pb-3 font-medium text-center">Status</th>
                      {(canCreateUser ||
                        canEditUserAny ||
                        canDeleteUserAny) && (
                        <th className="pb-3 font-medium text-right">Actions</th>
                      )}
                    </tr>
                  </thead>
                  <tbody>
                    {users.map((user) => (
                      <tr key={user.id} className="border-b last:border-0">
                        <td className="py-3">
                          <p className="font-medium">{user.name}</p>
                        </td>
                        <td className="py-3 text-sm text-muted-foreground">
                          {user.email || "-"}
                        </td>
                        <td className="py-3">
                          <span
                            className={`inline-flex items-center gap-1 px-2 py-1 rounded-full text-xs font-medium ${getRoleBadgeColor(
                              user.role_name,
                            )}`}
                          >
                            {getRoleIcon(user.role_name)}
                            <span className="capitalize">{user.role_name}</span>
                          </span>
                        </td>
                        <td className="py-3 text-center">
                          <span
                            className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${getBooleanBadgeClasses(
                              user.has_pin,
                            )}`}
                          >
                            {user.has_pin ? "Yes" : "No"}
                          </span>
                        </td>
                        <td className="py-3 text-center">
                          <span
                            className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${getBooleanBadgeClasses(
                              user.is_active,
                            )}`}
                          >
                            {user.is_active ? "Active" : "Archived"}
                          </span>
                        </td>
                        {(canCreateUser ||
                          canEditUserAny ||
                          canDeleteUserAny) && (
                          <td className="py-3 text-right">
                            <div className="flex items-center justify-end gap-1">
                              {viewMode === "active" ? (
                                <>
                                  {canEditUser(user) && (
                                    <Button
                                      variant="ghost"
                                      size="icon"
                                      onClick={() => onEdit(user)}
                                      title="Edit user"
                                    >
                                      <Pencil className="h-4 w-4" />
                                    </Button>
                                  )}
                                  {canDeleteUser(user) && (
                                    <Button
                                      variant="ghost"
                                      size="icon"
                                      onClick={() => onArchive(user)}
                                      title="Archive user"
                                    >
                                      <Archive className="h-4 w-4 text-amber-600 hover:text-amber-700 dark:text-amber-500" />
                                    </Button>
                                  )}
                                </>
                              ) : (
                                <>
                                  {canDeleteUser(user) && (
                                    <Button
                                      variant="ghost"
                                      size="sm"
                                      onClick={() => onRestore(user)}
                                      disabled={isSubmitting}
                                    >
                                      <RotateCcw className="h-4 w-4 mr-1" />
                                      Restore
                                    </Button>
                                  )}
                                  {canDeleteUser(user) && (
                                    <Button
                                      variant="ghost"
                                      size="sm"
                                      className="text-destructive hover:text-destructive"
                                      onClick={() => onPermanentDelete(user)}
                                    >
                                      <Trash2 className="h-4 w-4 mr-1" />
                                      Delete
                                    </Button>
                                  )}
                                </>
                              )}
                            </div>
                          </td>
                        )}
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>

              <div className="lg:hidden space-y-4">
            {users.map((user) => (
              <Card key={user.id} className="@container">
                <CardContent className="p-4 space-y-3">
                  <div className="flex justify-between items-start border-b pb-2">
                    <div>
                      <p className="font-bold">{user.name}</p>
                      <p className="text-sm text-muted-foreground">
                        {user.email || "-"}
                      </p>
                    </div>
                    <span
                      className={`inline-flex items-center gap-1 px-2 py-1 rounded-full text-xs font-medium ${getRoleBadgeColor(
                        user.role_name,
                      )}`}
                    >
                      {getRoleIcon(user.role_name)}
                      <span className="capitalize">{user.role_name}</span>
                    </span>
                  </div>

                  <div className="grid grid-cols-2 gap-2 text-sm text-muted-foreground">
                    <div>
                      <span className="text-xs block">PIN Set</span>
                      <span
                        className={
                          user.has_pin ? "text-green-600 font-medium" : "text-neutral-500"
                        }
                      >
                        {user.has_pin ? "Yes" : "No"}
                      </span>
                    </div>
                    <div>
                      <span className="text-xs block">Status</span>
                      <span
                        className={
                          user.is_active ? "text-green-600 font-medium" : "text-neutral-500"
                        }
                      >
                        {user.is_active ? "Active" : "Archived"}
                      </span>
                    </div>
                  </div>

                  {(canCreateUser ||
                    canEditUserAny ||
                    canDeleteUserAny) && (
                    <div className="flex justify-end gap-2 pt-2 border-t">
                      {viewMode === "active" ? (
                        <>
                          {canEditUser(user) && (
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() => onEdit(user)}
                              className="h-8 w-8 p-0 @[250px]:w-auto @[250px]:px-3"
                              title="Edit"
                            >
                              <Pencil className="h-3.5 w-3.5 @[250px]:mr-1" />
                              <span className="hidden @[250px]:inline">Edit</span>
                            </Button>
                          )}
                          {canDeleteUser(user) && (
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() => onArchive(user)}
                              className="h-8 w-8 p-0 @[250px]:w-auto @[250px]:px-3 text-amber-600 hover:bg-amber-50 hover:text-amber-700 dark:text-amber-500 dark:hover:bg-amber-950/20"
                              title="Archive"
                            >
                              <Archive className="h-3.5 w-3.5 @[250px]:mr-1" />
                              <span className="hidden @[250px]:inline">Archive</span>
                            </Button>
                          )}
                        </>
                      ) : (
                        <>
                          {canDeleteUser(user) && (
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() => onRestore(user)}
                              disabled={isSubmitting}
                              className="h-8 w-8 p-0 @[250px]:w-auto @[250px]:px-3"
                              title="Restore"
                            >
                              <RotateCcw className="h-3.5 w-3.5 @[250px]:mr-1" />
                              <span className="hidden @[250px]:inline">Restore</span>
                            </Button>
                          )}
                          {canDeleteUser(user) && (
                            <Button
                              variant="outline"
                              size="sm"
                              className="h-8 w-8 p-0 @[250px]:w-auto @[250px]:px-3 text-destructive hover:bg-destructive/10 hover:text-destructive"
                              onClick={() => onPermanentDelete(user)}
                              title="Delete"
                            >
                              <Trash2 className="h-3.5 w-3.5 @[250px]:mr-1" />
                              <span className="hidden @[250px]:inline">Delete</span>
                            </Button>
                          )}
                        </>
                      )}
                    </div>
                  )}
                </CardContent>
              </Card>
            ))}
              </div>
            </>
          )}

          {!isLoading && users.length > 0 && (
            <div className="flex items-center justify-between mt-6 pt-4 border-t">
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => onPageChange(Math.max(1, page - 1))}
                  disabled={page === 1}
                >
                  <ChevronLeft className="h-4 w-4 mr-1" />
                  Previous
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => onPageChange(page + 1)}
                  disabled={!hasMore}
                >
                  Next
                  <ChevronRight className="h-4 w-4 ml-1" />
                </Button>
              </div>
            </div>
          )}
      </DataTableContainer>
    </div>
  );
}

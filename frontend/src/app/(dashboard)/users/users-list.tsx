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
  Settings2,
  Shield,
  ShieldAlert,
  ShieldCheck,
  Trash2,
  Users,
} from "lucide-react";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import type { User, UserRole } from "@/types";

interface UsersListProps {
  pageError?: string | null;
  users: User[];
  viewMode: "active" | "archived";
  isLoading: boolean;
  canCreateUser: boolean;
  canEditUserAny: boolean;
  canDeleteUserAny: boolean;
  canManagePermissions: boolean;
  searchQuery: string;
  selectedRole: string;
  page: number;
  limit: number;
  total: number;
  hasMore: boolean;
  isSubmitting: boolean;
  onCreate: () => void;
  onSearchChange: (value: string) => void;
  onRoleChange: (value: string) => void;
  onViewModeChange: (value: "active" | "archived") => void;
  onLimitChange: (value: number) => void;
  onPageChange: (value: number) => void;
  canEditUser: (user: User) => boolean;
  canDeleteUser: (user: User) => boolean;
  canManageUserPermissions: (user: User) => boolean;
  onEdit: (user: User) => void;
  onManagePermissions: (user: User) => void;
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

export function UsersList({
  pageError,
  users,
  viewMode,
  isLoading,
  canCreateUser,
  canEditUserAny,
  canDeleteUserAny,
  canManagePermissions,
  searchQuery,
  selectedRole,
  page,
  limit,
  total,
  hasMore,
  isSubmitting,
  onCreate,
  onSearchChange,
  onRoleChange,
  onViewModeChange,
  onLimitChange,
  onPageChange,
  canEditUser,
  canDeleteUser,
  canManageUserPermissions,
  onEdit,
  onManagePermissions,
  onArchive,
  onPermanentDelete,
  onRestore,
}: UsersListProps) {
  return (
    <div className="flex-1 p-6 overflow-auto">
      <div className="flex gap-1 mb-4 p-1 bg-muted rounded-lg w-fit">
        <button
          onClick={() => onViewModeChange("active")}
          className={`px-4 py-2 rounded-md text-sm font-medium transition-colors ${
            viewMode === "active"
              ? "bg-background text-foreground shadow-sm"
              : "text-muted-foreground hover:text-foreground"
          }`}
        >
          Active
        </button>
        <button
          onClick={() => onViewModeChange("archived")}
          className={`px-4 py-2 rounded-md text-sm font-medium transition-colors flex items-center gap-2 ${
            viewMode === "archived"
              ? "bg-background text-foreground shadow-sm"
              : "text-muted-foreground hover:text-foreground"
          }`}
        >
          <Archive className="h-4 w-4" />
          Archived
        </button>
      </div>

      <Card className="mb-6">
        <CardContent className="p-6">
          {pageError && (
            <div className="mb-4 rounded-md border border-destructive/20 bg-destructive/10 p-3">
              <p className="text-sm text-destructive">{pageError}</p>
            </div>
          )}
          <div className="flex flex-col sm:flex-row gap-4">
            <div className="relative flex-1 w-full">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder="Search users..."
                value={searchQuery}
                onChange={(event) => onSearchChange(event.target.value)}
                className="pl-9 w-full"
              />
            </div>
            <Select value={selectedRole} onValueChange={onRoleChange}>
              <SelectTrigger className="w-full sm:w-40">
                <SelectValue placeholder="All Roles" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Roles</SelectItem>
                <SelectItem value="owner">Owner</SelectItem>
                <SelectItem value="manager">Manager</SelectItem>
                <SelectItem value="cashier">Cashier</SelectItem>
              </SelectContent>
            </Select>
            {canCreateUser && viewMode === "active" && (
              <Button onClick={onCreate} className="w-full sm:w-auto">
                <Plus className="h-4 w-4 mr-2" />
                Add User
              </Button>
            )}
          </div>
        </CardContent>
      </Card>

      {isLoading ? (
        <div className="flex items-center justify-center h-64">
          <Loader2 className="h-8 w-8 animate-spin text-primary" />
        </div>
      ) : users.length === 0 ? (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <Users className="h-12 w-12 text-muted-foreground mb-4" />
            <p className="text-muted-foreground">
              {viewMode === "active" ? "No users found" : "No archived users"}
            </p>
            {canCreateUser && viewMode === "active" && (
              <Button variant="link" onClick={onCreate}>
                Add your first user
              </Button>
            )}
          </CardContent>
        </Card>
      ) : (
        <>
          <Card className="hidden lg:block">
            <CardHeader>
              <CardTitle>
                {viewMode === "active" ? "Users" : "Archived Users"} ({users.length})
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead>
                    <tr className="border-b text-left text-sm text-muted-foreground">
                      <th className="pb-3 font-medium">Name</th>
                      <th className="pb-3 font-medium">Email</th>
                      <th className="pb-3 font-medium">Role</th>
                      <th className="pb-3 font-medium text-center">PIN Set</th>
                      <th className="pb-3 font-medium text-center">Status</th>
                      {(canCreateUser ||
                        canEditUserAny ||
                        canDeleteUserAny ||
                        canManagePermissions) && (
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
                            className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${
                              user.has_pin
                                ? "bg-green-600 text-white dark:bg-green-600/90 dark:text-white"
                                : "bg-gray-600 text-white dark:bg-gray-600/90 dark:text-white"
                            }`}
                          >
                            {user.has_pin ? "Yes" : "No"}
                          </span>
                        </td>
                        <td className="py-3 text-center">
                          <span
                            className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${
                              user.is_active
                                ? "bg-green-600 text-white dark:bg-green-600/90 dark:text-white"
                                : "bg-gray-600 text-white dark:bg-gray-600/90 dark:text-white"
                            }`}
                          >
                            {user.is_active ? "Active" : "Archived"}
                          </span>
                        </td>
                        {(canCreateUser ||
                          canEditUserAny ||
                          canDeleteUserAny ||
                          canManagePermissions) && (
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
                                  {canManageUserPermissions(user) && (
                                    <Button
                                      variant="ghost"
                                      size="icon"
                                      onClick={() => onManagePermissions(user)}
                                      title="Manage permissions"
                                    >
                                      <Settings2 className="h-4 w-4" />
                                    </Button>
                                  )}
                                  {canDeleteUser(user) && (
                                    <Button
                                      variant="ghost"
                                      size="icon"
                                      onClick={() => onArchive(user)}
                                      title="Archive user"
                                    >
                                      <Trash2 className="h-4 w-4 text-destructive" />
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
            </CardContent>
          </Card>

          <div className="lg:hidden space-y-4">
            <h3 className="font-semibold text-lg">
              {viewMode === "active" ? "Users" : "Archived Users"} ({users.length})
            </h3>
            {users.map((user) => (
              <Card key={user.id}>
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
                    canDeleteUserAny ||
                    canManagePermissions) && (
                    <div className="flex justify-end gap-2 pt-2 border-t">
                      {viewMode === "active" ? (
                        <>
                          {canEditUser(user) && (
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() => onEdit(user)}
                              className="h-8"
                            >
                              <Pencil className="h-3.5 w-3.5 mr-1" />
                              Edit
                            </Button>
                          )}
                          {canManageUserPermissions(user) && (
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() => onManagePermissions(user)}
                              className="h-8"
                            >
                              <Settings2 className="h-3.5 w-3.5 mr-1" />
                              Perms
                            </Button>
                          )}
                          {canDeleteUser(user) && (
                            <Button
                              variant="outline"
                              size="sm"
                              onClick={() => onArchive(user)}
                              className="h-8 text-destructive hover:text-destructive"
                            >
                              <Trash2 className="h-3.5 w-3.5 mr-1" />
                              Archive
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
                              className="h-8"
                            >
                              <RotateCcw className="h-3.5 w-3.5 mr-1" />
                              Restore
                            </Button>
                          )}
                          {canDeleteUser(user) && (
                            <Button
                              variant="outline"
                              size="sm"
                              className="h-8 text-destructive hover:text-destructive"
                              onClick={() => onPermanentDelete(user)}
                            >
                              <Trash2 className="h-3.5 w-3.5 mr-1" />
                              Delete
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

          <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between mt-6 pt-4 border-t">
            <div className="flex items-center gap-2">
              <span className="text-sm text-muted-foreground">Show</span>
              <Select value={String(limit)} onValueChange={(value) => onLimitChange(Number(value))}>
                <SelectTrigger className="w-[80px]">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="10">10</SelectItem>
                  <SelectItem value="20">20</SelectItem>
                  <SelectItem value="50">50</SelectItem>
                  <SelectItem value="100">100</SelectItem>
                </SelectContent>
              </Select>
              <span className="text-sm text-muted-foreground">entries</span>
            </div>

            <div className="text-sm text-muted-foreground">
              Page {page} • {Math.min(users.length, limit)} entries of {total}
            </div>

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
        </>
      )}
    </div>
  );
}

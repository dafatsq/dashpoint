"use client";

import { useEffect, useMemo, useState } from "react";
import { ShieldCheck, Users as UsersIcon } from "lucide-react";

import { Header } from "@/components/layout/header";
import api from "@/lib/api";
import { PERMISSIONS, useAuth } from "@/contexts/auth-context";
import { useGlobalError } from "@/contexts/error-context";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import type {
  CreateUserRequest,
  Role,
  UpdateUserRequest,
  User,
} from "@/types";

import {
  canDeleteUser,
  canEditUser,
  findRoleIdByName,
  getAssignableUserRoles,
  hasUserFormChanges,
} from "./users-helpers";
import { ConfirmationDialog } from "@/components/shared/confirmation-dialog";
import { UsersFormDialog } from "./users-form-dialog";
import { UsersList } from "./users-list";
import { UsersRolePermissionsDialog } from "./users-role-permissions-dialog";
import { normalizeRolePermissionKeys } from "./users-role-permissions";
import { UsersRolesList } from "./users-roles-list";

export default function UsersScreen() {
  const { user: currentUser, hasPermission } = useAuth();
  const canEditUserAny = hasPermission(PERMISSIONS.USERS_EDIT);
  const canDeleteUserAny = hasPermission(PERMISSIONS.USERS_DELETE);
  const isOwner = currentUser?.role_name === "owner";
  const assignableCreateRoles = useMemo(
    () =>
      getAssignableUserRoles({
        currentUser,
        isOwner,
        hasPermission,
      }),
    [currentUser, hasPermission, isOwner],
  );
  const canCreateUser = assignableCreateRoles.length > 0;

  const [users, setUsers] = useState<User[]>([]);
  const [activeTab, setActiveTab] = useState<"users" | "roles">("users");
  const [roles, setRoles] = useState<Role[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [pageError, setPageError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedRole, setSelectedRole] = useState<string>("all");
  const [sort, setSort] = useState("created_at_desc");
  const [viewMode, setViewMode] = useState<"active" | "archived">("active");
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(20);
  const [total, setTotal] = useState(0);
  const [hasMore, setHasMore] = useState(true);

  const [dialogOpen, setDialogOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [permanentDeleteDialogOpen, setPermanentDeleteDialogOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [editingUser, setEditingUser] = useState<User | null>(null);
  const [deletingUser, setDeletingUser] = useState<User | null>(null);
  const [editingRole, setEditingRole] = useState<Role | null>(null);
  const [rolePermissionsDialogOpen, setRolePermissionsDialogOpen] = useState(false);
  const [rolePermissionValues, setRolePermissionValues] = useState<string[]>([]);
  const [isSubmittingRolePermissions, setIsSubmittingRolePermissions] = useState(false);
  const { showError } = useGlobalError();

  const [formData, setFormData] = useState<CreateUserRequest>({
    email: "",
    password: "",
    name: "",
    role: undefined,
    pin: "",
  });
  const [formErrors, setFormErrors] = useState<{ general?: string }>({});

  const resetToFirstPage = () => setPage(1);

  useEffect(() => {
    const fetchUsers = async () => {
      setIsLoading(true);
      setPageError(null);

      try {
        const result = await api.getUsersPage({
          active: viewMode === "active",
          search: searchQuery || undefined,
          role: selectedRole !== "all" ? selectedRole : undefined,
          page,
          per_page: limit,
          sort_by: sort.replace(/_(asc|desc)$/, ""),
          sort_direction: sort.endsWith("_desc") ? "desc" : "asc",
        });

        if (result.error) {
          setPageError(result.error);
          setUsers([]);
          setTotal(0);
          setHasMore(false);
          return;
        }

        if (result.data) {
          setUsers(result.data);
          setTotal(result.total || 0);
          if (result.total_pages !== undefined) {
            setHasMore(page < result.total_pages);
          } else {
            setHasMore(result.data.length === limit);
          }
        }
      } catch {
        setPageError("Failed to load users. Please try again.");
      } finally {
        setIsLoading(false);
      }
    };

    fetchUsers();
  }, [viewMode, page, limit, searchQuery, selectedRole, sort]);

  useEffect(() => {
    const fetchRoles = async () => {
      try {
        const result = await api.getRoles();
        if (result.error) {
          setPageError(
            (current) => current || result.error || "Failed to load roles.",
          );
          return;
        }
        setRoles(result.data || []);
      } catch {
        setPageError((current) => current || "Failed to load roles.");
      }
    };

    fetchRoles();
  }, []);

  const refreshRoles = async () => {
    const result = await api.getRoles();
    if (result.error) {
      showError("Load Failed", result.error);
      return;
    }
    setRoles(result.data || []);
  };

  const resetForm = () => {
    setFormData({
      email: "",
      password: "",
      name: "",
      role: assignableCreateRoles[0],
      pin: "",
    });
    setEditingUser(null);
    setFormErrors({});
  };

  const openCreateDialog = () => {
    resetForm();
    setDialogOpen(true);
  };

  const openEditDialog = (user: User) => {
    setEditingUser(user);
    setFormData({
      email: user.email || "",
      password: "",
      name: user.name,
      role: user.role_name,
      pin: "",
    });
    setFormErrors({});
    setDialogOpen(true);
  };

  const hasChanges = useMemo(
    () => hasUserFormChanges(formData, editingUser),
    [formData, editingUser],
  );
  const availableDialogRoles = useMemo(
    () =>
      editingUser
        ? getAssignableUserRoles({
            currentUser,
            isOwner,
            hasPermission,
            editingUser,
          })
        : assignableCreateRoles,
    [assignableCreateRoles, currentUser, editingUser, hasPermission, isOwner],
  );
  const hasRolePermissionChanges = useMemo(() => {
    if (!editingRole) {
      return false;
    }

    const savedPermissions = normalizeRolePermissionKeys(editingRole.permissions || []);
    const nextPermissions = normalizeRolePermissionKeys(rolePermissionValues);
    if (savedPermissions.length !== nextPermissions.length) {
      return true;
    }

    return savedPermissions.some((permission, index) => permission !== nextPermissions[index]);
  }, [editingRole, rolePermissionValues]);

  const submitUserForm = async () => {
    if (!hasChanges) {
      showError("No Changes", "Make a change before saving.");
      return;
    }

    if (
      !formData.name ||
      (!editingUser && (!formData.email || !formData.password || !formData.pin))
    ) {
      showError("Incomplete Form", "Complete all required fields before saving.");
      return;
    }

    setIsSubmitting(true);
    setFormErrors({});

    try {
      if (!editingUser) {
        if (availableDialogRoles.length === 0) {
          showError("Permission Denied", "You do not have permission to create any user roles.");
          return;
        }

        if (!formData.role || !availableDialogRoles.includes(formData.role)) {
          showError("Invalid Role", "Select an allowed role");
          return;
        }
      }

      const roleID = findRoleIdByName(roles, formData.role);
      if (!roleID) {
        showError("Invalid Role", "Invalid role selected");
        return;
      }

      if (editingUser) {
        const updateData: UpdateUserRequest = {
          name: formData.name,
          role_id: roleID,
          email: formData.email,
          expected_updated_at: editingUser.updated_at,
        };
        if (formData.password) updateData.password = formData.password;
        if (formData.pin) updateData.pin = formData.pin;

        const result = await api.updateUser(editingUser.id, updateData);
        if (result.error) {
          showError("Save Failed", result.error);
          return;
        }
        if (result.data) {
          setUsers((current) =>
            current.map((entry) => (entry.id === editingUser.id ? result.data! : entry)),
          );
        }
      } else {
        const result = await api.createUser({
          ...formData,
          role_id: roleID,
        });
        if (result.error) {
          showError("Save Failed", result.error);
          return;
        }
        if (result.data) {
          setUsers((current) => [...current, result.data!]);
        }
      }

      setDialogOpen(false);
      resetForm();
    } finally {
      setIsSubmitting(false);
    }
  };

  const archiveUser = async () => {
    if (!deletingUser) return;

    setIsSubmitting(true);
    try {
      const result = await api.deleteUser(deletingUser.id, deletingUser.updated_at);
      if (result.error) {
        showError("Archive Failed", result.error);
        return;
      }
      setUsers((current) => current.filter((entry) => entry.id !== deletingUser.id));
      setDeleteDialogOpen(false);
      setDeletingUser(null);
    } finally {
      setIsSubmitting(false);
    }
  };

  const restoreUser = async (user: User) => {
    setIsSubmitting(true);
    setPageError(null);
    try {
      const result = await api.updateUser(user.id, {
        is_active: true,
        expected_updated_at: user.updated_at,
      });
      if (result.error) {
        showError("Restore Failed", result.error);
        return;
      }
      if (result.data) {
        setUsers((current) => current.filter((entry) => entry.id !== user.id));
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  const permanentlyDeleteUser = async () => {
    if (!deletingUser) return;

    setIsSubmitting(true);
    try {
      const result = await api.permanentDeleteUser(
        deletingUser.id,
        deletingUser.updated_at,
      );
      if (result.error) {
        showError("Delete Failed", result.error);
        return;
      }
      setUsers((current) => current.filter((entry) => entry.id !== deletingUser.id));
      setPermanentDeleteDialogOpen(false);
      setDeletingUser(null);
    } finally {
      setIsSubmitting(false);
    }
  };

  const openRolePermissionsDialog = (role: Role) => {
    if (!isOwner || role.name === "owner") {
      return;
    }

    setEditingRole(role);
    setRolePermissionValues(normalizeRolePermissionKeys(role.permissions || []));
    setRolePermissionsDialogOpen(true);
  };

  const submitRolePermissions = async () => {
    if (!editingRole || !isOwner) {
      return;
    }
    if (!hasRolePermissionChanges) {
      return;
    }

    setIsSubmittingRolePermissions(true);
    try {
      const result = await api.updateRolePermissions(
        editingRole.id,
        normalizeRolePermissionKeys(rolePermissionValues),
        normalizeRolePermissionKeys(editingRole.permissions || []),
      );
      if (result.error) {
        showError("Save Failed", result.error);
        return;
      }
      if (result.data) {
        const savedRole = result.data;
        setRoles((current) =>
          current.map((role) => (role.id === savedRole.id ? savedRole : role)),
        );
        setEditingRole(savedRole);
        setRolePermissionValues(normalizeRolePermissionKeys(savedRole.permissions || []));
      } else {
        await refreshRoles();
      }
    } finally {
      setIsSubmittingRolePermissions(false);
    }
  };

  const userCanEdit = (targetUser: User) =>
    canEditUser({
      currentUser,
      targetUser,
      canEditUserAny,
      hasPermission,
    });

  const userCanDelete = (targetUser: User) =>
    canDeleteUser({
      currentUser,
      targetUser,
      canDeleteUserAny,
      hasPermission,
    });

  const handleSearchChange = (value: string) => {
    setSearchQuery(value);
    resetToFirstPage();
  };

  const handleRoleChange = (value: string) => {
    setSelectedRole(value);
    resetToFirstPage();
  };

  const handleViewModeChange = (value: "active" | "archived") => {
    setViewMode(value);
    resetToFirstPage();
  };

  const handleLimitChange = (value: number) => {
    setLimit(value);
    resetToFirstPage();
  };

  return (
    <div className="flex flex-col h-full">
      <Header title="Users" />
      <div className="flex-1 overflow-auto">
        <div className="px-6 pt-6">
          <Tabs value={activeTab} onValueChange={(value) => setActiveTab(value as "users" | "roles")} className="w-full xl:w-auto">
            <TabsList className="grid w-full grid-cols-2 xl:min-w-[400px]">
              <TabsTrigger value="users" className="flex items-center gap-2">
                <UsersIcon className="h-4 w-4 shrink-0" />
                Users
              </TabsTrigger>
              <TabsTrigger value="roles" className="flex items-center gap-2">
                <ShieldCheck className="h-4 w-4 shrink-0" />
                Roles
              </TabsTrigger>
            </TabsList>
          </Tabs>
        </div>

        {activeTab === "users" ? (
          <UsersList
            pageError={pageError}
            users={users}
            viewMode={viewMode}
            isLoading={isLoading}
            canCreateUser={canCreateUser}
            canEditUserAny={canEditUserAny}
            canDeleteUserAny={canDeleteUserAny}
            searchQuery={searchQuery}
            selectedRole={selectedRole}
            sort={sort}
            page={page}
            limit={limit}
            total={total}
            hasMore={hasMore}
            isSubmitting={isSubmitting}
            onCreate={openCreateDialog}
            onSearchChange={handleSearchChange}
            onRoleChange={handleRoleChange}
            onSortChange={(value) => {
              setSort(value);
              resetToFirstPage();
            }}
            onViewModeChange={handleViewModeChange}
            onLimitChange={handleLimitChange}
            onPageChange={setPage}
            canEditUser={userCanEdit}
            canDeleteUser={userCanDelete}
            onEdit={openEditDialog}
            onArchive={(user) => {
              setDeletingUser(user);
              setDeleteDialogOpen(true);
            }}
            onPermanentDelete={(user) => {
              setDeletingUser(user);
              setPermanentDeleteDialogOpen(true);
            }}
            onRestore={restoreUser}
          />
        ) : (
          <div className="p-6 pt-4">
            <UsersRolesList
              roles={roles}
              isOwner={isOwner}
              onEditPermissions={openRolePermissionsDialog}
            />
          </div>
        )}
      </div>

      <UsersFormDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        editingUser={editingUser}
        currentUser={currentUser}
        formData={formData}
        setFormData={setFormData}
        formErrors={formErrors}
        isSubmitting={isSubmitting}
        hasChanges={hasChanges}
        availableRoles={availableDialogRoles}
        onSubmit={submitUserForm}
      />

      <ConfirmationDialog
        open={deleteDialogOpen}
        onOpenChange={setDeleteDialogOpen}
        title="Archive User"
        description={`Are you sure you want to archive "${deletingUser?.name}"? The user will be moved to the Archived tab and can be restored later.`}
        confirmText="Archive"
        isSubmitting={isSubmitting}
        loadingText="Archiving..."
        onConfirm={archiveUser}
      />

      <ConfirmationDialog
        open={permanentDeleteDialogOpen}
        onOpenChange={setPermanentDeleteDialogOpen}
        title="Permanently Delete User"
        description={`Are you sure you want to permanently delete "${deletingUser?.name}"? This action cannot be undone. All data associated with this user will be lost.`}
        confirmText="Delete Permanently"
        confirmVariant="destructive"
        isSubmitting={isSubmitting}
        loadingText="Deleting..."
        onConfirm={permanentlyDeleteUser}
      />

      <UsersRolePermissionsDialog
        open={rolePermissionsDialogOpen}
        role={editingRole}
        selectedPermissions={rolePermissionValues}
        isSubmitting={isSubmittingRolePermissions}
        hasChanges={hasRolePermissionChanges}
        onOpenChange={(open) => {
          setRolePermissionsDialogOpen(open);
          if (!open) {
            setEditingRole(null);
            setRolePermissionValues([]);
            setIsSubmittingRolePermissions(false);
          }
        }}
        onPermissionsChange={setRolePermissionValues}
        onSubmit={() => void submitRolePermissions()}
      />
    </div>
  );
}

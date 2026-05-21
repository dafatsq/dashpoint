"use client";

import { useEffect, useMemo, useState } from "react";

import { Header } from "@/components/layout/header";
import api from "@/lib/api";
import { PERMISSIONS, useAuth } from "@/contexts/auth-context";
import { useGlobalError } from "@/contexts/error-context";
import type {
  CreateUserRequest,
  Permission,
  PermissionOverride,
  UpdateUserRequest,
  User,
} from "@/types";

import {
  canDeleteUser,
  canEditUser,
  canManageUserPermissions,
  createPermissionChangeSet,
  findRoleIdByName,
  getAssignableUserRoles,
  hasUserFormChanges,
} from "./users-helpers";
import { ConfirmationDialog } from "@/components/shared/confirmation-dialog";
import { UsersFormDialog } from "./users-form-dialog";
import { UsersList } from "./users-list";
import { UsersPermissionsDialog } from "./users-permissions-dialog";

export default function UsersScreen() {
  const { user: currentUser, hasPermission } = useAuth();
  const canEditUserAny =
    hasPermission(PERMISSIONS.USERS_EDIT) ||
    hasPermission("can_edit_manager_users") ||
    hasPermission("can_edit_cashier_users");
  const canDeleteUserAny =
    hasPermission(PERMISSIONS.USERS_DELETE) ||
    hasPermission("can_delete_manager_users") ||
    hasPermission("can_delete_cashier_users");
  const canManagePermissions =
    hasPermission(PERMISSIONS.USERS_PERMISSIONS) ||
    hasPermission("can_manage_manager_permissions") ||
    hasPermission("can_manage_cashier_permissions");
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
  const [roles, setRoles] = useState<Array<{ id: string; name: string; description?: string }>>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [pageError, setPageError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedRole, setSelectedRole] = useState<string>("all");
  const [viewMode, setViewMode] = useState<"active" | "archived">("active");
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(20);
  const [total, setTotal] = useState(0);
  const [hasMore, setHasMore] = useState(true);

  const [dialogOpen, setDialogOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [permanentDeleteDialogOpen, setPermanentDeleteDialogOpen] = useState(false);
  const [permissionsDialogOpen, setPermissionsDialogOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [editingUser, setEditingUser] = useState<User | null>(null);
  const [deletingUser, setDeletingUser] = useState<User | null>(null);
  const [permissionsUser, setPermissionsUser] = useState<User | null>(null);
  const { showError } = useGlobalError();
  const [saveSuccess, setSaveSuccess] = useState(false);

  const [allPermissions, setAllPermissions] = useState<Record<string, Permission[]>>({});
  const [userEffectivePermissions, setUserEffectivePermissions] = useState<string[]>([]);
  const [userOverrides, setUserOverrides] = useState<PermissionOverride[]>([]);
  const [permissionChanges, setPermissionChanges] = useState<Record<string, boolean | null>>({});
  const [isLoadingPermissions, setIsLoadingPermissions] = useState(false);

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
  }, [viewMode, page, limit, searchQuery, selectedRole]);

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

  const closePermissionsDialog = () => {
    setPermissionsDialogOpen(false);
    setPermissionsUser(null);
    setPermissionChanges({});
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
      const result = await api.deleteUser(deletingUser.id);
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
      const result = await api.updateUser(user.id, { is_active: true });
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
      const result = await api.permanentDeleteUser(deletingUser.id);
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

  const openPermissionsDialog = async (user: User) => {
    setPermissionsUser(user);
    setPermissionsDialogOpen(true);
    setIsLoadingPermissions(true);
    setPermissionChanges({});
    setSaveSuccess(false);

    try {
      const [permissionsResult, userPermissionsResult] = await Promise.all([
        api.getPermissions(true),
        api.getUserPermissions(user.id),
      ]);

      if (permissionsResult.error) {
        showError("Load Failed", permissionsResult.error);
        return;
      }
      if (userPermissionsResult.error) {
        showError("Load Failed", userPermissionsResult.error);
        return;
      }

      setAllPermissions((permissionsResult.data || {}) as Record<string, Permission[]>);
      setUserEffectivePermissions(
        userPermissionsResult.data?.effective_permissions || [],
      );
      setUserOverrides(userPermissionsResult.data?.overrides || []);
    } catch {
      showError("Load Error", "Failed to load permissions. Please try again.");
    } finally {
      setIsLoadingPermissions(false);
    }
  };

  const togglePermission = (permission: Permission, enabled: boolean) => {
    setPermissionChanges((current) =>
      createPermissionChangeSet({
        permission,
        enabled,
        allPermissions,
        permissionChanges: current,
        userOverrides,
        userEffectivePermissions,
        currentUser,
        targetUser: permissionsUser,
      }),
    );
  };

  const savePermissionChanges = async () => {
    if (!permissionsUser) return;
    if (Object.keys(permissionChanges).length === 0) {
      showError("No Changes", "Make at least one permission change before saving.");
      return;
    }

    setIsSubmitting(true);
    try {
      const permissions = Object.entries(permissionChanges).map(([id, allowed]) => ({
        permission_id: id,
        allowed: allowed === true,
      }));

      const result = await api.setUserPermissions(permissionsUser.id, permissions);
      if (result.error) {
        showError("Save Failed", result.error);
        return;
      }

      const refreshed = await api.getUserPermissions(permissionsUser.id);
      if (refreshed.error) {
        showError("Save Failed", refreshed.error);
        return;
      }

      setUserEffectivePermissions(refreshed.data?.effective_permissions || []);
      setUserOverrides(refreshed.data?.overrides || []);
      setPermissionChanges({});
      setSaveSuccess(true);
      setTimeout(() => setSaveSuccess(false), 3000);
    } catch (error) {
      showError("Save Error", error instanceof Error ? error.message : "An unexpected error occurred");
    } finally {
      setIsSubmitting(false);
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

  const userCanManagePermissions = (targetUser: User) =>
    canManageUserPermissions({
      currentUser,
      targetUser,
      canManagePermissions,
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
      <UsersList
        pageError={pageError}
        users={users}
        viewMode={viewMode}
        isLoading={isLoading}
        canCreateUser={canCreateUser}
        canEditUserAny={canEditUserAny}
        canDeleteUserAny={canDeleteUserAny}
        canManagePermissions={canManagePermissions}
        searchQuery={searchQuery}
        selectedRole={selectedRole}
        page={page}
        limit={limit}
        total={total}
        hasMore={hasMore}
        isSubmitting={isSubmitting}
        onCreate={openCreateDialog}
        onSearchChange={handleSearchChange}
        onRoleChange={handleRoleChange}
        onViewModeChange={handleViewModeChange}
        onLimitChange={handleLimitChange}
        onPageChange={setPage}
        canEditUser={userCanEdit}
        canDeleteUser={userCanDelete}
        canManageUserPermissions={userCanManagePermissions}
        onEdit={openEditDialog}
        onManagePermissions={openPermissionsDialog}
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

      <UsersFormDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        editingUser={editingUser}
        currentUser={currentUser}
        formData={formData}
        setFormData={setFormData}
        formErrors={formErrors}
        isSubmitting={isSubmitting}
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

      <UsersPermissionsDialog
        open={permissionsDialogOpen}
        onOpenChange={(open) => {
          setPermissionsDialogOpen(open);
          if (!open) {
            closePermissionsDialog();
          }
        }}
        permissionsUser={permissionsUser}
        currentUser={currentUser}
        allPermissions={allPermissions}
        userEffectivePermissions={userEffectivePermissions}
        userOverrides={userOverrides}
        permissionChanges={permissionChanges}
        isLoadingPermissions={isLoadingPermissions}
        isSubmitting={isSubmitting}
        saveSuccess={saveSuccess}
        onTogglePermission={togglePermission}
        onSaveChanges={savePermissionChanges}
        onCancel={closePermissionsDialog}
      />
    </div>
  );
}

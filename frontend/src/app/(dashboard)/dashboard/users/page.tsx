'use client';

import { useEffect, useState } from 'react';
import { Header } from '@/components/layout/header';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Search,
  Plus,
  Pencil,
  Trash2,
  Loader2,
  Users,
  Shield,
  ShieldCheck,
  ShieldAlert,
  RotateCcw,
  Archive,
  Settings2,
} from 'lucide-react';
import api from '@/lib/api';
import { User, CreateUserRequest, UpdateUserRequest, UserRole, Permission, PermissionOverride } from '@/types';
import { useAuth, PERMISSIONS } from '@/contexts/auth-context';
import { AccountManager } from '@/lib/account-manager';
import { Switch } from '@/components/ui/switch';

// Role hierarchy for permission management
const roleHierarchy: Record<string, number> = {
  owner: 3,
  manager: 2,
  cashier: 1,
};

export default function UsersPage() {
  const { user: currentUser, hasPermission } = useAuth();
  const canManageUsers = hasPermission(PERMISSIONS.USERS_MANAGE);
  const canManagePermissions = hasPermission(PERMISSIONS.USERS_PERMISSIONS);
  const isOwner = currentUser?.role_name === 'owner';

  // Helper to check if current user can edit a specific user
  const canEditUser = (targetUser: User) => {
    if (!canManageUsers) return false;
    // Owners can edit anyone
    if (isOwner) return true;
    // Non-owners cannot edit owners
    if (targetUser.role_name === 'owner') return false;
    return true;
  };

  // Helper to check if current user can manage permissions of a specific user
  const canManageUserPermissions = (targetUser: User) => {
    if (!canManagePermissions) return false;
    const currentLevel = roleHierarchy[currentUser?.role_name || ''] || 0;
    const targetLevel = roleHierarchy[targetUser.role_name] || 0;
    // Can only manage permissions of users with strictly lower role level
    return targetLevel < currentLevel;
  };

  const [users, setUsers] = useState<User[]>([]);
  const [roles, setRoles] = useState<{ id: string; name: string; description: string }[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedRole, setSelectedRole] = useState<string>('all');
  const [viewMode, setViewMode] = useState<'active' | 'archived'>('active');

  // Dialog states
  const [dialogOpen, setDialogOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [permanentDeleteDialogOpen, setPermanentDeleteDialogOpen] = useState(false);
  const [permissionsDialogOpen, setPermissionsDialogOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [editingUser, setEditingUser] = useState<User | null>(null);
  const [deletingUser, setDeletingUser] = useState<User | null>(null);
  const [permissionsUser, setPermissionsUser] = useState<User | null>(null);
  const [deleteError, setDeleteError] = useState<string | null>(null);

  // Permission management state
  const [allPermissions, setAllPermissions] = useState<Record<string, Permission[]>>({});
  const [userEffectivePermissions, setUserEffectivePermissions] = useState<string[]>([]);
  const [userOverrides, setUserOverrides] = useState<PermissionOverride[]>([]);
  const [permissionChanges, setPermissionChanges] = useState<Record<string, boolean | null>>({});
  const [isLoadingPermissions, setIsLoadingPermissions] = useState(false);

  // Form state
  const [formData, setFormData] = useState<CreateUserRequest>({
    email: '',
    password: '',
    name: '',
    role: 'cashier',
    pin: '',
  });
  const [formErrors, setFormErrors] = useState<{ general?: string }>({});

  // Fetch users
  useEffect(() => {
    const fetchUsers = async () => {
      setIsLoading(true);
      try {
        const result = await api.getUsers({ active: viewMode === 'active' });
        if (result.data) setUsers(result.data);
      } catch (error) {
        console.error('Failed to fetch users:', error);
      } finally {
        setIsLoading(false);
      }
    };

    fetchUsers();
  }, [viewMode]);

  // Fetch roles
  useEffect(() => {
    const fetchRoles = async () => {
      try {
        const result = await api.getRoles();
        if (result.data) setRoles(result.data);
      } catch (error) {
        console.error('Failed to fetch roles:', error);
      }
    };

    fetchRoles();
  }, []);

  // Filter users
  const filteredUsers = users.filter((user) => {
    const matchesSearch =
      user.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      user.email?.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesRole = selectedRole === 'all' || user.role_name === selectedRole;
    return matchesSearch && matchesRole;
  });

  // Reset form
  const resetForm = () => {
    setFormData({
      email: '',
      password: '',
      name: '',
      role: 'cashier',
      pin: '',
    });
    setEditingUser(null);
    setFormErrors({});
  };

  // Open create dialog
  const openCreateDialog = () => {
    resetForm();
    setDialogOpen(true);
  };

  // Open edit dialog
  const openEditDialog = (user: User) => {
    setEditingUser(user);
    setFormData({
      email: user.email || '',
      password: '', // Don't pre-fill password
      name: user.name,
      role: user.role_name,
      pin: '',
    });
    setDialogOpen(true);
  };

  // Handle submit
  const handleSubmit = async () => {
    if (!formData.name || (!editingUser && (!formData.email || !formData.password))) return;

    setIsSubmitting(true);
    setFormErrors({});

    try {
      // Find role_id from role name
      const roleObj = roles.find(r => r.name === formData.role);
      if (!roleObj) {
        setFormErrors({ general: 'Invalid role selected' });
        setIsSubmitting(false);
        return;
      }

      if (editingUser) {
        const updateData: UpdateUserRequest = {
          name: formData.name,
          role_id: roleObj.id,
          email: formData.email,
        };
        // Only include password/pin if provided
        if (formData.password) updateData.password = formData.password;
        if (formData.pin) updateData.pin = formData.pin;

        const result = await api.updateUser(editingUser.id, updateData);
        if (result.error) {
          console.warn('Update user failed:', result.error);
          setFormErrors({ general: result.error });
          return;
        }
        if (result.data) {
          setUsers((prev) =>
            prev.map((u) => (u.id === editingUser.id ? result.data! : u))
          );
        }
      } else {
        const createData = {
          ...formData,
          role_id: roleObj.id,
        };
        const result = await api.createUser(createData);
        if (result.error) {
          console.warn('Create user failed:', result.error);
          setFormErrors({ general: result.error });
          return;
        }
        if (result.data) {
          setUsers((prev) => [...prev, result.data!]);
        }
      }
      setDialogOpen(false);
      resetForm();
    } catch (error) {
      console.error('Failed to save user:', error);
      setFormErrors({ general: 'Failed to save user. Please try again.' });
    } finally {
      setIsSubmitting(false);
    }
  };

  // Handle delete (soft delete / archive)
  const handleDelete = async () => {
    if (!deletingUser) return;

    setIsSubmitting(true);

    try {
      await api.deleteUser(deletingUser.id);
      setUsers((prev) => prev.filter((u) => u.id !== deletingUser.id));
      setDeleteDialogOpen(false);
      setDeletingUser(null);
    } catch (error) {
      console.error('Failed to delete user:', error);
    } finally {
      setIsSubmitting(false);
    }
  };

  // Handle restore
  const handleRestore = async (user: User) => {
    setIsSubmitting(true);
    try {
      const result = await api.updateUser(user.id, { is_active: true });
      if (result.data) {
        setUsers((prev) => prev.filter((u) => u.id !== user.id));
      }
    } catch (error) {
      console.error('Failed to restore user:', error);
    } finally {
      setIsSubmitting(false);
    }
  };

  // Handle permanent delete
  const handlePermanentDelete = async () => {
    if (!deletingUser) return;

    setIsSubmitting(true);
    setDeleteError(null);
    try {
      const result = await api.permanentDeleteUser(deletingUser.id);
      if (result.error) {
        setDeleteError(result.error);
      } else {
        setUsers((prev) => prev.filter((u) => u.id !== deletingUser.id));
        setPermanentDeleteDialogOpen(false);
        setDeletingUser(null);
      }
    } catch (error) {
      console.error('Failed to permanently delete user:', error);
      setDeleteError('Failed to delete user. Please try again.');
    } finally {
      setIsSubmitting(false);
    }
  };

  // Open permissions dialog
  const openPermissionsDialog = async (user: User) => {
    setPermissionsUser(user);
    setPermissionsDialogOpen(true);
    setIsLoadingPermissions(true);
    setPermissionChanges({});

    try {
      // Fetch all permissions grouped by category
      const permResult = await api.getPermissions(true);
      if (permResult.data) {
        setAllPermissions(permResult.data as Record<string, Permission[]>);
      }

      // Fetch user's current permissions
      const userPermResult = await api.getUserPermissions(user.id);
      if (userPermResult.data) {
        setUserEffectivePermissions(userPermResult.data.effective_permissions || []);
        setUserOverrides(userPermResult.data.overrides || []);
      }
    } catch (error) {
      console.error('Failed to fetch permissions:', error);
    } finally {
      setIsLoadingPermissions(false);
    }
  };

  // Check if a permission has an override
  const getPermissionOverride = (permissionId: string): PermissionOverride | undefined => {
    return userOverrides.find(o => o.permission_id === permissionId);
  };

  // Check if a permission is currently enabled (either via role or override)
  const isPermissionEnabled = (permission: Permission): boolean => {
    // First check if there's a pending change
    if (permissionChanges[permission.id] !== undefined) {
      return permissionChanges[permission.id] === true;
    }
    // Then check if there's an override
    const override = getPermissionOverride(permission.id);
    if (override) {
      return override.allowed;
    }
    // Finally check effective permissions
    return userEffectivePermissions.includes(permission.key);
  };

  // Handle permission toggle
  const handlePermissionToggle = (permission: Permission, enabled: boolean) => {
    // Check if this is reverting to the role default
    const override = getPermissionOverride(permission.id);
    const roleDefault = userEffectivePermissions.includes(permission.key) && !override;
    
    if (enabled === roleDefault && !override) {
      // Remove from changes if reverting to role default
      const newChanges = { ...permissionChanges };
      delete newChanges[permission.id];
      setPermissionChanges(newChanges);
    } else {
      setPermissionChanges(prev => ({
        ...prev,
        [permission.id]: enabled,
      }));
    }
  };

  // Save permission changes
  const savePermissionChanges = async () => {
    if (!permissionsUser || Object.keys(permissionChanges).length === 0) return;

    setIsSubmitting(true);
    try {
      const permissions = Object.entries(permissionChanges).map(([id, allowed]) => ({
        permission_id: id,
        allowed: allowed === true,
      }));

      const result = await api.setUserPermissions(permissionsUser.id, permissions);
      if (result.error) {
        console.error('Failed to save permissions:', result.error);
        return;
      }

      // Refresh the permissions data
      const userPermResult = await api.getUserPermissions(permissionsUser.id);
      if (userPermResult.data) {
        setUserEffectivePermissions(userPermResult.data.effective_permissions || []);
        setUserOverrides(userPermResult.data.overrides || []);
      }
      setPermissionChanges({});
    } catch (error) {
      console.error('Failed to save permissions:', error);
    } finally {
      setIsSubmitting(false);
    }
  };

  // Get the status text for a permission
  const getPermissionStatus = (permission: Permission): { text: string; color: string } => {
    const override = getPermissionOverride(permission.id);
    const hasChange = permissionChanges[permission.id] !== undefined;
    
    if (hasChange) {
      return { text: 'Modified', color: 'text-yellow-600' };
    }
    if (override) {
      return { 
        text: override.allowed ? 'Override: Granted' : 'Override: Denied', 
        color: override.allowed ? 'text-green-600' : 'text-red-600' 
      };
    }
    return { text: 'From Role', color: 'text-muted-foreground' };
  };

  // Check if a permission is a "view/access" permission (controls sidebar access)
  // can_view_sales is the parent for can_void_sale (void is in sales history page)
  // can_create_sale is independent (POS page)
  const isViewPermission = (permission: Permission): boolean => {
    // can_create_sale is independent - POS page
    if (permission.key === 'can_create_sale') {
      return false;
    }
    // can_view_sales is the access permission for sales history
    if (permission.key === 'can_view_sales') {
      return true;
    }
    // can_void_sale is a child of can_view_sales
    if (permission.key === 'can_void_sale') {
      return false;
    }
    return permission.key.startsWith('can_view_');
  };

  // Get the view permission for a category
  const getViewPermissionForCategory = (category: string, permissions: Permission[]): Permission | undefined => {
    // For sales category, can_view_sales is the parent for void permission
    if (category === 'sales') {
      return permissions.find(p => p.key === 'can_view_sales');
    }
    return permissions.find(p => isViewPermission(p));
  };

  // Check if a permission should be disabled based on parent permission
  const isPermissionDisabledByParent = (permission: Permission, category: string): boolean => {
    // can_void_sale depends on can_view_sales (void is in sales history page)
    if (permission.key === 'can_void_sale') {
      const permissions = allPermissions[category] || [];
      const viewPerm = permissions.find(p => p.key === 'can_view_sales');
      if (viewPerm && !isPermissionEnabled(viewPerm)) {
        return true;
      }
    }
    // can_create_sale is independent - never disabled by parent
    if (permission.key === 'can_create_sale') {
      return false;
    }
    // For other categories, use the standard view permission check
    if (category !== 'sales') {
      const viewPerm = getViewPermissionForCategory(category, allPermissions[category] || []);
      if (viewPerm && !isViewPermission(permission) && !isPermissionEnabled(viewPerm)) {
        return true;
      }
    }
    return false;
  };

  // Check if view permission for a category is enabled (for backward compatibility)
  const isViewPermissionEnabled = (category: string): boolean => {
    const permissions = allPermissions[category] || [];
    const viewPerm = getViewPermissionForCategory(category, permissions);
    if (!viewPerm) return true;
    return isPermissionEnabled(viewPerm);
  };

  // Sort permissions: view permissions first, then others
  const sortPermissions = (permissions: Permission[]): Permission[] => {
    return [...permissions].sort((a, b) => {
      const aIsView = isViewPermission(a);
      const bIsView = isViewPermission(b);
      if (aIsView && !bIsView) return -1;
      if (!aIsView && bIsView) return 1;
      return 0;
    });
  };

  // Get display name for permission (rename view permissions to "Access to...")
  const getPermissionDisplayName = (permission: Permission, category: string): string => {
    // Sales permissions - special naming
    if (permission.key === 'can_view_sales') {
      return 'Sales History Access';
    }
    if (permission.key === 'can_create_sale') {
      return 'Create Sales (POS)';
    }
    if (isViewPermission(permission)) {
      // Special case: system category shows as "Audit"
      if (category === 'system') {
        return 'Access to Audit';
      }
      // Capitalize category name
      const categoryName = category.charAt(0).toUpperCase() + category.slice(1);
      return `Access to ${categoryName}`;
    }
    return permission.name;
  };

  const getRoleIcon = (role: UserRole) => {
    switch (role) {
      case 'owner':
        return <ShieldAlert className="h-4 w-4" />;
      case 'manager':
        return <ShieldCheck className="h-4 w-4" />;
      case 'cashier':
        return <Shield className="h-4 w-4" />;
    }
  };

  const getRoleBadgeColor = (role: UserRole) => {
    switch (role) {
      case 'owner':
        return 'bg-purple-600 text-white dark:bg-purple-600/90 dark:text-white';
      case 'manager':
        return 'bg-blue-600 text-white dark:bg-blue-600/90 dark:text-white';
      case 'cashier':
        return 'bg-green-600 text-white dark:bg-green-600/90 dark:text-white';
    }
  };

  return (
    <div className="flex flex-col h-full">
      <Header title="User Management" />

      <div className="flex-1 p-6 overflow-auto">
        {/* Tab Toggle */}
        <div className="flex gap-1 mb-4 p-1 bg-muted rounded-lg w-fit">
          <button
            onClick={() => setViewMode('active')}
            className={`px-4 py-2 rounded-md text-sm font-medium transition-colors ${
              viewMode === 'active'
                ? 'bg-background text-foreground shadow-sm'
                : 'text-muted-foreground hover:text-foreground'
            }`}
          >
            Active
          </button>
          <button
            onClick={() => setViewMode('archived')}
            className={`px-4 py-2 rounded-md text-sm font-medium transition-colors flex items-center gap-2 ${
              viewMode === 'archived'
                ? 'bg-background text-foreground shadow-sm'
                : 'text-muted-foreground hover:text-foreground'
            }`}
          >
            <Archive className="h-4 w-4" />
            Archived
          </button>
        </div>

        {/* Toolbar */}
        <div className="flex flex-col sm:flex-row gap-4 mb-6">
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Search users..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-9"
            />
          </div>
          <Select value={selectedRole} onValueChange={setSelectedRole}>
            <SelectTrigger className="w-40">
              <SelectValue placeholder="All Roles" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Roles</SelectItem>
              <SelectItem value="owner">Owner</SelectItem>
              <SelectItem value="manager">Manager</SelectItem>
              <SelectItem value="cashier">Cashier</SelectItem>
            </SelectContent>
          </Select>
          {canManageUsers && viewMode === 'active' && (
            <Button onClick={openCreateDialog}>
              <Plus className="h-4 w-4 mr-2" />
              Add User
            </Button>
          )}
        </div>

        {/* Users table */}
        {isLoading ? (
          <div className="flex items-center justify-center h-64">
            <Loader2 className="h-8 w-8 animate-spin text-primary" />
          </div>
        ) : filteredUsers.length === 0 ? (
          <Card>
            <CardContent className="flex flex-col items-center justify-center py-12">
              <Users className="h-12 w-12 text-muted-foreground mb-4" />
              <p className="text-muted-foreground">
                {viewMode === 'active' ? 'No users found' : 'No archived users'}
              </p>
              {canManageUsers && viewMode === 'active' && (
                <Button variant="link" onClick={openCreateDialog}>
                  Add your first user
                </Button>
              )}
            </CardContent>
          </Card>
        ) : (
          <Card>
            <CardHeader>
              <CardTitle>
                {viewMode === 'active' ? 'Users' : 'Archived Users'} ({filteredUsers.length})
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
                      {canManageUsers && (
                        <th className="pb-3 font-medium text-right">Actions</th>
                      )}
                    </tr>
                  </thead>
                  <tbody>
                    {filteredUsers.map((user) => (
                      <tr key={user.id} className="border-b last:border-0">
                        <td className="py-3">
                          <p className="font-medium">{user.name}</p>
                        </td>
                        <td className="py-3 text-sm text-muted-foreground">
                          {user.email || '-'}
                        </td>
                        <td className="py-3">
                          <span
                            className={`inline-flex items-center gap-1 px-2 py-1 rounded-full text-xs font-medium ${getRoleBadgeColor(
                              user.role_name
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
                                ? 'bg-green-600 text-white dark:bg-green-600/90 dark:text-white'
                                : 'bg-gray-600 text-white dark:bg-gray-600/90 dark:text-white'
                            }`}
                          >
                            {user.has_pin ? 'Yes' : 'No'}
                          </span>
                        </td>
                        <td className="py-3 text-center">
                          <span
                            className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${
                              user.is_active
                                ? 'bg-green-600 text-white dark:bg-green-600/90 dark:text-white'
                                : 'bg-gray-600 text-white dark:bg-gray-600/90 dark:text-white'
                            }`}
                          >
                            {user.is_active ? 'Active' : 'Archived'}
                          </span>
                        </td>
                        {canManageUsers && (
                          <td className="py-3 text-right">
                            <div className="flex items-center justify-end gap-1">
                              {viewMode === 'active' ? (
                                <>
                                  {canEditUser(user) && (
                                    <Button
                                      variant="ghost"
                                      size="icon"
                                      onClick={() => openEditDialog(user)}
                                      title="Edit user"
                                    >
                                      <Pencil className="h-4 w-4" />
                                    </Button>
                                  )}
                                  {canManageUserPermissions(user) && (
                                    <Button
                                      variant="ghost"
                                      size="icon"
                                      onClick={() => openPermissionsDialog(user)}
                                      title="Manage permissions"
                                    >
                                      <Settings2 className="h-4 w-4" />
                                    </Button>
                                  )}
                                  {canEditUser(user) && user.role_name !== 'owner' && (
                                    <Button
                                      variant="ghost"
                                      size="icon"
                                      onClick={() => {
                                        setDeletingUser(user);
                                        setDeleteDialogOpen(true);
                                      }}
                                      title="Archive user"
                                    >
                                      <Trash2 className="h-4 w-4 text-destructive" />
                                    </Button>
                                  )}
                                </>
                              ) : (
                                <>
                                  {canEditUser(user) && (
                                    <Button
                                      variant="ghost"
                                      size="sm"
                                      onClick={() => handleRestore(user)}
                                      disabled={isSubmitting}
                                    >
                                      <RotateCcw className="h-4 w-4 mr-1" />
                                      Restore
                                    </Button>
                                  )}
                                  {canEditUser(user) && user.role_name !== 'owner' && (
                                    <Button
                                      variant="ghost"
                                      size="sm"
                                      className="text-destructive hover:text-destructive"
                                      onClick={() => {
                                        setDeletingUser(user);
                                        setPermanentDeleteDialogOpen(true);
                                      }}
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
        )}
      </div>

      {/* Create/Edit Dialog */}
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>
              {editingUser ? 'Edit User' : 'Add User'}
            </DialogTitle>
            <DialogDescription>
              {editingUser
                ? 'Update the user details below.'
                : 'Fill in the details for the new user.'}
            </DialogDescription>
          </DialogHeader>

          <div className="grid gap-4 py-4">
            {formErrors.general && (
              <div className="p-3 rounded-md bg-destructive/10 border border-destructive/20">
                <p className="text-sm text-destructive">{formErrors.general}</p>
              </div>
            )}

            <div className="grid gap-2">
              <Label htmlFor="name">Name *</Label>
              <Input
                id="name"
                value={formData.name}
                onChange={(e) =>
                  setFormData({ ...formData, name: e.target.value })
                }
                placeholder="Full name"
              />
            </div>

            <div className="grid gap-2">
              <Label htmlFor="email">Email {!editingUser && '*'}</Label>
              <Input
                id="email"
                type="email"
                value={formData.email}
                onChange={(e) =>
                  setFormData({ ...formData, email: e.target.value })
                }
                placeholder="user@example.com"
              />
            </div>

            <div className="grid gap-2">
              <Label htmlFor="password">
                Password {!editingUser ? '*' : '(leave blank to keep current)'}
              </Label>
              <Input
                id="password"
                type="password"
                value={formData.password}
                onChange={(e) =>
                  setFormData({ ...formData, password: e.target.value })
                }
                placeholder={editingUser ? 'Leave blank to keep current' : 'Enter password'}
              />
            </div>

            <div className="grid gap-2">
              <Label htmlFor="pin">PIN (4-6 digits, for quick login)</Label>
              <Input
                id="pin"
                type="password"
                value={formData.pin}
                onChange={(e) =>
                  setFormData({ ...formData, pin: e.target.value.replace(/\D/g, '').slice(0, 6) })
                }
                placeholder={editingUser ? 'Leave blank to keep current' : 'Optional 4-6 digit PIN'}
                maxLength={6}
              />
            </div>

            <div className="grid gap-2">
              <Label htmlFor="role">Role</Label>
              <Select
                value={formData.role}
                onValueChange={(value: UserRole) =>
                  setFormData({ ...formData, role: value })
                }
              >
                <SelectTrigger>
                  <SelectValue placeholder="Select role" />
                </SelectTrigger>
                <SelectContent>
                  {isOwner && (
                    <SelectItem value="owner">
                      <div className="flex items-center gap-2">
                        <ShieldAlert className="h-4 w-4" />
                        Owner - Full access
                      </div>
                    </SelectItem>
                  )}
                  <SelectItem value="manager">
                    <div className="flex items-center gap-2">
                      <ShieldCheck className="h-4 w-4" />
                      Manager - Most access
                    </div>
                  </SelectItem>
                  <SelectItem value="cashier">
                    <div className="flex items-center gap-2">
                      <Shield className="h-4 w-4" />
                      Cashier - Sales only
                    </div>
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setDialogOpen(false)}>
              Cancel
            </Button>
            <Button onClick={handleSubmit} disabled={isSubmitting}>
              {isSubmitting ? (
                <>
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  Saving...
                </>
              ) : editingUser ? (
                'Update User'
              ) : (
                'Create User'
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation Dialog (Soft Delete / Archive) */}
      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Archive User</DialogTitle>
            <DialogDescription>
              Are you sure you want to archive &quot;{deletingUser?.name}&quot;? 
              The user will be moved to the Archived tab and can be restored later.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteDialogOpen(false)}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={handleDelete} disabled={isSubmitting}>
              {isSubmitting ? (
                <>
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  Archiving...
                </>
              ) : (
                'Archive'
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Permanent Delete Confirmation Dialog */}
      <Dialog open={permanentDeleteDialogOpen} onOpenChange={(open) => {
        setPermanentDeleteDialogOpen(open);
        if (!open) {
          setDeleteError(null);
        }
      }}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Permanently Delete User</DialogTitle>
            <DialogDescription>
              Are you sure you want to permanently delete &quot;{deletingUser?.name}&quot;? 
              This action cannot be undone. All data associated with this user will be lost.
            </DialogDescription>
          </DialogHeader>
          {deleteError && (
            <div className="p-3 rounded-md bg-destructive/10 border border-destructive/20">
              <p className="text-sm text-destructive">{deleteError}</p>
            </div>
          )}
          <DialogFooter>
            <Button variant="outline" onClick={() => {
              setPermanentDeleteDialogOpen(false);
              setDeleteError(null);
            }}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={handlePermanentDelete} disabled={isSubmitting || !!deleteError}>
              {isSubmitting ? (
                <>
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  Deleting...
                </>
              ) : (
                'Delete Permanently'
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Permissions Management Dialog */}
      <Dialog open={permissionsDialogOpen} onOpenChange={(open) => {
        setPermissionsDialogOpen(open);
        if (!open) {
          setPermissionsUser(null);
          setPermissionChanges({});
        }
      }}>
        <DialogContent className="max-w-2xl max-h-[80vh] flex flex-col overflow-hidden">
          <DialogHeader className="flex-shrink-0">
            <DialogTitle className="flex items-center gap-2">
              <Settings2 className="h-5 w-5" />
              Manage Permissions
            </DialogTitle>
            <DialogDescription>
              Configure individual permissions for <span className="font-semibold">{permissionsUser?.name}</span> ({permissionsUser?.role_name}).
              These overrides will grant or deny access regardless of the user&apos;s role.
            </DialogDescription>
          </DialogHeader>

          <div className="flex-1 overflow-y-auto pr-2">
          {isLoadingPermissions ? (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
          ) : (
            <div className="space-y-6">
              {Object.entries(allPermissions).map(([category, permissions]) => {
                const sortedPermissions = sortPermissions(permissions);
                
                return (
                  <div key={category} className="space-y-3">
                    <h4 className="font-semibold capitalize text-sm border-b pb-1">{category}</h4>
                    <div className="space-y-2">
                      {sortedPermissions.map((permission) => {
                        const isViewPerm = isViewPermission(permission);
                        const isEnabled = isPermissionEnabled(permission);
                        const status = getPermissionStatus(permission);
                        
                        // Use the new per-permission disabled check
                        const isDisabled = isPermissionDisabledByParent(permission, category);
                        
                        // Get the appropriate message for disabled state
                        const getDisabledMessage = () => {
                          if (permission.key === 'can_void_sale') {
                            return 'Enable "Sales History Access" first';
                          }
                          return `Enable "Access to ${category.charAt(0).toUpperCase() + category.slice(1)}" first`;
                        };
                        
                        return (
                          <div 
                            key={permission.id} 
                            className={`flex items-center justify-between py-2 px-3 rounded-md transition-colors ${
                              isDisabled 
                                ? 'bg-muted/30 opacity-50' 
                                : 'bg-muted/50 hover:bg-muted'
                            }`}
                          >
                            <div className="flex-1">
                              <div className="flex items-center gap-2">
                                <span className={`font-medium text-sm ${isDisabled ? 'text-muted-foreground' : ''}`}>
                                  {getPermissionDisplayName(permission, category)}
                                </span>
                                {isViewPerm && (
                                  <span className="text-xs text-blue-600 font-medium">(Controls Access)</span>
                                )}
                                {!isDisabled && (
                                  <span className={`text-xs ${status.color}`}>({status.text})</span>
                                )}
                              </div>
                              {permission.description && (
                                <p className="text-xs text-muted-foreground mt-0.5">{permission.description}</p>
                              )}
                              {isDisabled && (
                                <p className="text-xs text-orange-600 mt-0.5">
                                  {getDisabledMessage()}
                                </p>
                              )}
                            </div>
                            <Switch
                              checked={isDisabled ? false : isEnabled}
                              onCheckedChange={(checked) => handlePermissionToggle(permission, checked)}
                              disabled={isDisabled}
                            />
                          </div>
                        );
                      })}
                    </div>
                  </div>
                );
              })}
            </div>
          )}
          </div>

          <DialogFooter className="flex-shrink-0 flex items-center justify-between gap-2 pt-4 border-t bg-background">
            <div className="text-sm text-muted-foreground">
              {Object.keys(permissionChanges).length > 0 && (
                <span className="text-yellow-600">
                  {Object.keys(permissionChanges).length} unsaved change(s)
                </span>
              )}
            </div>
            <div className="flex gap-2">
              <Button 
                variant="outline" 
                onClick={() => {
                  setPermissionsDialogOpen(false);
                  setPermissionsUser(null);
                  setPermissionChanges({});
                }}
              >
                Cancel
              </Button>
              <Button 
                onClick={savePermissionChanges} 
                disabled={isSubmitting || Object.keys(permissionChanges).length === 0}
              >
                {isSubmitting ? (
                  <>
                    <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                    Saving...
                  </>
                ) : (
                  'Save Changes'
                )}
              </Button>
            </div>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

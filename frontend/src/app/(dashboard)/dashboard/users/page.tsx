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
} from 'lucide-react';
import api from '@/lib/api';
import { User, CreateUserRequest, UpdateUserRequest, UserRole } from '@/types';
import { useAuth, PERMISSIONS } from '@/contexts/auth-context';
import { AccountManager } from '@/lib/account-manager';

export default function UsersPage() {
  const { user: currentUser, hasPermission } = useAuth();
  const canManageUsers = hasPermission(PERMISSIONS.USERS_MANAGE);
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
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [editingUser, setEditingUser] = useState<User | null>(null);
  const [deletingUser, setDeletingUser] = useState<User | null>(null);
  const [deleteError, setDeleteError] = useState<string | null>(null);

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
        return 'bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-400';
      case 'manager':
        return 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400';
      case 'cashier':
        return 'bg-gray-100 text-gray-700 dark:bg-gray-900/30 dark:text-gray-400';
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
                                ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
                                : 'bg-gray-100 text-gray-700 dark:bg-gray-900/30 dark:text-gray-400'
                            }`}
                          >
                            {user.has_pin ? 'Yes' : 'No'}
                          </span>
                        </td>
                        <td className="py-3 text-center">
                          <span
                            className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${
                              user.is_active
                                ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
                                : 'bg-gray-100 text-gray-700 dark:bg-gray-900/30 dark:text-gray-400'
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
                                    >
                                      <Pencil className="h-4 w-4" />
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
    </div>
  );
}

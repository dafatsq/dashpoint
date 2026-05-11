'use client';

import { useState, useEffect, useMemo } from 'react';
import { Header } from '@/components/layout/header';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Switch } from '@/components/ui/switch';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  ShieldCheck,
  Save,
  Loader2,
  User as UserIcon,
  Pencil,
  AlertCircle
} from 'lucide-react';
import { getRememberMeKey, migrateSession } from '@/lib/session';
import { useAuth } from '@/contexts/auth-context';
import { AccountManager } from '@/lib/account-manager';
import api from '@/lib/api';

export default function SettingsPage() {
  const { user, login, refreshUser } = useAuth();
  const [isSaving, setIsSaving] = useState(false);
  const [rememberMe, setRememberMe] = useState(true);
  const [quickAccess, setQuickAccess] = useState(false);
  const [initialRememberMe, setInitialRememberMe] = useState(true);
  const [initialQuickAccess, setInitialQuickAccess] = useState(false);

  // Profile Edit State
  const [verifyPasswordOpen, setVerifyPasswordOpen] = useState(false);
  const [editProfileOpen, setEditProfileOpen] = useState(false);
  const [passwordEntry, setPasswordEntry] = useState('');
  const [passwordError, setPasswordError] = useState('');
  const [isVerifyingPassword, setIsVerifyingPassword] = useState(false);

  const [editData, setEditData] = useState({
    name: '',
    email: '',
    password: '',
    pin: ''
  });
  const [editError, setEditError] = useState('');
  const [isUpdatingProfile, setIsUpdatingProfile] = useState(false);

  // Load saved preference on mount
  useEffect(() => {
    if (!user) return;
    
    const prefKey = getRememberMeKey(user.id);
    const savedPref = localStorage.getItem(prefKey);
    
    // Default is true if not explicitly set to false
    const isRememberMe = savedPref !== 'false';
    setRememberMe(isRememberMe);
    setInitialRememberMe(isRememberMe);

    const isSaved = AccountManager.getAccount(user.id) !== null;
    setQuickAccess(isSaved);
    setInitialQuickAccess(isSaved);
  }, [user]);

  const hasChanges = rememberMe !== initialRememberMe || quickAccess !== initialQuickAccess;

  const handleSave = async () => {
    if (!user) return;
    
    setIsSaving(true);

    const prefKey = getRememberMeKey(user.id);
    // Save the auth preference per user
    localStorage.setItem(prefKey, rememberMe ? 'true' : 'false');

    // Migrate existing tokens to the correct storage backend
    migrateSession(rememberMe);

    // Manage quick access
    const shouldSaveQuickAccess = rememberMe || quickAccess;
    if (shouldSaveQuickAccess) {
      AccountManager.saveAccount({
        id: user.id,
        name: user.name,
        email: user.email,
        role_name: user.role_name,
        has_pin: user.has_pin
      });
    } else {
      AccountManager.removeAccount(user.id);
    }
    
    setInitialRememberMe(rememberMe);
    setInitialQuickAccess(quickAccess);

    await new Promise((resolve) => setTimeout(resolve, 500));
    setIsSaving(false);
  };

  const handleOpenEditProfile = () => {
    setPasswordEntry('');
    setPasswordError('');
    setVerifyPasswordOpen(true);
  };

  const handleVerifyPassword = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!user) return;
    
    setPasswordError('');
    setIsVerifyingPassword(true);

    try {
      const result = await login(user.email || '', passwordEntry, false);
      if (result.success) {
        setVerifyPasswordOpen(false);
        setEditData({
          name: user.name,
          email: user.email || '',
          password: '',
          pin: ''
        });
        setEditError('');
        setEditProfileOpen(true);
      } else {
        setPasswordError(result.error || 'Invalid Password');
      }
    } catch {
      setPasswordError('An unexpected error occurred');
    } finally {
      setIsVerifyingPassword(false);
    }
  };

  const hasProfileChanges = useMemo(() => {
    if (!user) return false;
    return (
      editData.name !== user.name ||
      editData.email !== (user.email || '') ||
      editData.password !== '' ||
      editData.pin !== ''
    );
  }, [editData, user]);

  const handleUpdateProfile = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!user || !editData.name) return;

    setEditError('');
    setIsUpdatingProfile(true);

    try {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const updateData: any = {
        name: editData.name,
        email: editData.email,
      };
      if (editData.password) updateData.password = editData.password;
      if (editData.pin) updateData.pin = editData.pin;

      const result = await api.updateUser(user.id, updateData);
      
      if (result.error) {
        setEditError(result.error);
      } else {
        await refreshUser();
        setEditProfileOpen(false);
      }
    } catch (error) {
      setEditError('Failed to update profile. Please try again.');
    } finally {
      setIsUpdatingProfile(false);
    }
  };

  return (
    <div className="flex flex-col h-full">
      <Header title="Settings" />

      <div className="flex-1 p-6 overflow-auto">
        <div className="max-w-3xl mx-auto space-y-6">

          {/* User Profile */}
          <Card>
            <CardHeader className="flex flex-row items-start justify-between space-y-0 pb-4">
              <div className="flex items-center gap-2">
                <UserIcon className="h-5 w-5 text-primary" />
                <div>
                  <CardTitle>My Profile</CardTitle>
                  <CardDescription>Manage your personal details</CardDescription>
                </div>
              </div>
              <Button variant="outline" size="sm" onClick={handleOpenEditProfile}>
                <Pencil className="h-4 w-4 mr-2" />
                Edit Profile
              </Button>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div>
                  <Label className="text-muted-foreground text-xs uppercase">Name</Label>
                  <p className="font-medium">{user?.name}</p>
                </div>
                <div>
                  <Label className="text-muted-foreground text-xs uppercase">Email</Label>
                  <p className="font-medium">{user?.email || 'Not provided'}</p>
                </div>
                <div>
                  <Label className="text-muted-foreground text-xs uppercase">Role</Label>
                  <p className="font-medium capitalize">{user?.role_name}</p>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Authentication Settings */}
          <Card>
            <CardHeader>
              <div className="flex items-center gap-2">
                <ShieldCheck className="h-5 w-5 text-primary" />
                <div>
                  <CardTitle>Authentication</CardTitle>
                  <CardDescription>Manage security and sign-in preferences</CardDescription>
                </div>
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex items-center justify-between gap-4">
                <div className="flex-1 space-y-0.5">
                  <Label>Automatic Sign-In (Remember Me)</Label>
                  <p className="text-sm text-muted-foreground">
                    Keep this account signed in across browser restarts. Turning this off means this specific account logs out when the browser closes.
                  </p>
                </div>
                <Switch
                  checked={rememberMe}
                  onCheckedChange={(checked) => {
                    setRememberMe(checked);
                    if (checked) setQuickAccess(true);
                  }}
                />
              </div>

              <div className="flex items-center justify-between gap-4 pt-4 border-t">
                <div className="flex-1 space-y-0.5">
                  <Label>Quick Access (Save Login)</Label>
                  <p className="text-sm text-muted-foreground">
                    Save this account in the browser so you can log back in instantly using only your Quick PIN.
                  </p>
                </div>
                <Switch
                  checked={rememberMe ? true : quickAccess}
                  onCheckedChange={setQuickAccess}
                  disabled={rememberMe}
                />
              </div>
            </CardContent>
          </Card>

          {/* Save Button */}
          <div className="flex justify-end pt-2">
            <Button 
              onClick={handleSave} 
              disabled={isSaving || !hasChanges} 
              className="w-full sm:w-auto"
            >
              {isSaving ? (
                <>
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  Saving...
                </>
              ) : (
                <>
                  <Save className="h-4 w-4 mr-2" />
                  Save Settings
                </>
              )}
            </Button>
          </div>
        </div>
      </div>

      {/* Verify Password Dialog */}
      <Dialog open={verifyPasswordOpen} onOpenChange={setVerifyPasswordOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Verify Password</DialogTitle>
            <DialogDescription>
              Please enter your password to edit your profile.
            </DialogDescription>
          </DialogHeader>
          <form onSubmit={handleVerifyPassword} className="space-y-4 mt-4" autoComplete="off">
            {passwordError && (
              <div className="flex items-center gap-2 rounded-md bg-destructive/10 p-3 text-sm text-destructive">
                <AlertCircle className="h-4 w-4 flex-shrink-0" />
                <span>{passwordError}</span>
              </div>
            )}
            <div className="space-y-2">
              <Input
                type="password"
                placeholder="Enter your password"
                value={passwordEntry}
                onChange={(e) => setPasswordEntry(e.target.value)}
                readOnly
                onFocus={(e) => e.target.removeAttribute('readonly')}
                required
                disabled={isVerifyingPassword}
                autoFocus
              />
            </div>
            <div className="flex justify-end gap-2">
              <Button type="button" variant="outline" onClick={() => setVerifyPasswordOpen(false)} disabled={isVerifyingPassword}>
                Cancel
              </Button>
              <Button type="submit" disabled={isVerifyingPassword || !passwordEntry}>
                {isVerifyingPassword ? (
                  <>
                    <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                    Verifying...
                  </>
                ) : (
                  'Verify'
                )}
              </Button>
            </div>
          </form>
        </DialogContent>
      </Dialog>

      {/* Edit Profile Dialog */}
      <Dialog open={editProfileOpen} onOpenChange={setEditProfileOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Edit Profile</DialogTitle>
            <DialogDescription>
              Update your personal details below.
            </DialogDescription>
          </DialogHeader>
          <form onSubmit={handleUpdateProfile} className="space-y-4 mt-4" autoComplete="off">
            {editError && (
              <div className="flex items-center gap-2 rounded-md bg-destructive/10 p-3 text-sm text-destructive">
                <AlertCircle className="h-4 w-4 flex-shrink-0" />
                <span>{editError}</span>
              </div>
            )}
            <div className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="edit-name">Name <span className="text-destructive">*</span></Label>
                <Input
                  id="edit-name"
                  value={editData.name}
                  onChange={(e) => setEditData({ ...editData, name: e.target.value })}
                  required
                  disabled={isUpdatingProfile}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="edit-email">Email</Label>
                <Input
                  id="edit-email"
                  type="email"
                  value={editData.email}
                  onChange={(e) => setEditData({ ...editData, email: e.target.value })}
                  disabled={isUpdatingProfile}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="edit-password">Password (leave blank to keep current)</Label>
                <Input
                  id="edit-password"
                  type="password"
                  placeholder="Leave blank to keep current"
                  value={editData.password}
                  onChange={(e) => setEditData({ ...editData, password: e.target.value })}
                  disabled={isUpdatingProfile}
                  readOnly
                  onFocus={(e) => e.target.removeAttribute('readonly')}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="edit-pin">PIN (leave blank to keep current)</Label>
                <Input
                  id="edit-pin"
                  type="password"
                  placeholder="Leave blank to keep current"
                  value={editData.pin}
                  onChange={(e) => setEditData({ ...editData, pin: e.target.value })}
                  maxLength={6}
                  pattern="\d*"
                  inputMode="numeric"
                  disabled={isUpdatingProfile}
                  readOnly
                  onFocus={(e) => e.target.removeAttribute('readonly')}
                />
              </div>
            </div>
            <div className="flex justify-end gap-2 pt-2">
              <Button type="button" variant="outline" onClick={() => setEditProfileOpen(false)} disabled={isUpdatingProfile}>
                Cancel
              </Button>
              <Button type="submit" disabled={isUpdatingProfile || !hasProfileChanges || !editData.name}>
                {isUpdatingProfile ? (
                  <>
                    <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                    Saving...
                  </>
                ) : (
                  'Save Changes'
                )}
              </Button>
            </div>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
}

'use client';

import { useEffect, useMemo, useState } from "react";
import { Loader2, Save } from "lucide-react";

import { Header } from "@/components/layout/header";
import { Button } from "@/components/ui/button";
import { getRememberMeKey, migrateSession } from "@/lib/session";
import { useAuth } from "@/contexts/auth-context";
import { AccountManager } from "@/lib/account-manager";
import api from "@/lib/api";

import { SettingsAuthCard } from "./settings-auth-card";
import { SettingsEditProfileDialog } from "./settings-edit-profile-dialog";
import {
  buildProfileUpdatePayload,
  buildSettingsPreferences,
  hasSettingsPreferenceChanges,
  profileHasChanges,
  type SettingsPreferences,
  type SettingsProfileForm,
} from "./settings-helpers";
import { SettingsProfileCard } from "./settings-profile-card";
import { SettingsVerifyPasswordDialog } from "./settings-verify-password-dialog";

export function SettingsScreen() {
  const { user, login, refreshUser } = useAuth();
  const [isSaving, setIsSaving] = useState(false);
  const [preferences, setPreferences] = useState<SettingsPreferences>({
    rememberMe: true,
    quickAccess: false,
  });
  const [initialPreferences, setInitialPreferences] = useState<SettingsPreferences>({
    rememberMe: true,
    quickAccess: false,
  });
  const [verifyPasswordOpen, setVerifyPasswordOpen] = useState(false);
  const [editProfileOpen, setEditProfileOpen] = useState(false);
  const [passwordEntry, setPasswordEntry] = useState("");
  const [passwordError, setPasswordError] = useState("");
  const [isVerifyingPassword, setIsVerifyingPassword] = useState(false);
  const [editForm, setEditForm] = useState<SettingsProfileForm>({
    name: "",
    email: "",
    password: "",
    pin: "",
  });
  const [editError, setEditError] = useState("");
  const [isUpdatingProfile, setIsUpdatingProfile] = useState(false);

  useEffect(() => {
    if (!user) return;

    const timer = window.setTimeout(() => {
      const preferenceKey = getRememberMeKey(user.id);
      const nextPreferences = buildSettingsPreferences(
        localStorage.getItem(preferenceKey),
        AccountManager.getAccount(user.id) !== null,
      );
      setPreferences(nextPreferences);
      setInitialPreferences(nextPreferences);
    }, 0);

    return () => window.clearTimeout(timer);
  }, [user]);

  const hasPreferenceChanges = hasSettingsPreferenceChanges(preferences, initialPreferences);

  const hasProfileEditChanges = useMemo(() => {
    if (!user) return false;
    return profileHasChanges(user, editForm);
  }, [editForm, user]);

  const handleSave = async () => {
    if (!user) return;

    setIsSaving(true);

    const preferenceKey = getRememberMeKey(user.id);
    localStorage.setItem(preferenceKey, preferences.rememberMe ? "true" : "false");
    migrateSession(preferences.rememberMe);

    const shouldSaveQuickAccess = preferences.rememberMe || preferences.quickAccess;
    if (shouldSaveQuickAccess) {
      AccountManager.saveAccount({
        id: user.id,
        name: user.name,
        email: user.email,
        role_name: user.role_name,
        has_pin: user.has_pin,
      });
    } else {
      AccountManager.removeAccount(user.id);
    }

    setInitialPreferences(preferences);

    await new Promise((resolve) => setTimeout(resolve, 500));
    setIsSaving(false);
  };

  const handleOpenEditProfile = () => {
    setPasswordEntry("");
    setPasswordError("");
    setVerifyPasswordOpen(true);
  };

  const handleVerifyPassword = async (event: React.FormEvent) => {
    event.preventDefault();
    if (!user) return;

    setPasswordError("");
    setIsVerifyingPassword(true);

    try {
      const result = await login(user.email || "", passwordEntry, false);
      if (result.success) {
        setVerifyPasswordOpen(false);
        setEditForm({
          name: user.name,
          email: user.email || "",
          password: "",
          pin: "",
        });
        setEditError("");
        setEditProfileOpen(true);
      } else {
        setPasswordError(result.error || "Invalid Password");
      }
    } catch {
      setPasswordError("An unexpected error occurred");
    } finally {
      setIsVerifyingPassword(false);
    }
  };

  const handleUpdateProfile = async (event: React.FormEvent) => {
    event.preventDefault();
    if (!user || !editForm.name) return;

    setEditError("");
    setIsUpdatingProfile(true);

    try {
      const result = await api.updateUser(user.id, buildProfileUpdatePayload(user, editForm));
      if (result.error) {
        setEditError(result.error);
      } else {
        await refreshUser();
        setEditProfileOpen(false);
      }
    } catch {
      setEditError("Failed to update profile. Please try again.");
    } finally {
      setIsUpdatingProfile(false);
    }
  };

  return (
    <div className="flex h-full flex-col">
      <Header title="Settings" />

      <div className="flex-1 overflow-auto p-6">
        <div className="mx-auto max-w-3xl space-y-6">
          <SettingsProfileCard user={user} onEdit={handleOpenEditProfile} />

          <SettingsAuthCard
            preferences={preferences}
            onPreferencesChange={setPreferences}
          />

          <div className="flex justify-end pt-2">
            <Button onClick={handleSave} disabled={isSaving || !hasPreferenceChanges} className="w-full sm:w-auto">
              {isSaving ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Saving...
                </>
              ) : (
                <>
                  <Save className="mr-2 h-4 w-4" />
                  Save Settings
                </>
              )}
            </Button>
          </div>
        </div>
      </div>

      <SettingsVerifyPasswordDialog
        open={verifyPasswordOpen}
        password={passwordEntry}
        error={passwordError}
        isSubmitting={isVerifyingPassword}
        onOpenChange={setVerifyPasswordOpen}
        onPasswordChange={setPasswordEntry}
        onSubmit={handleVerifyPassword}
      />

      <SettingsEditProfileDialog
        open={editProfileOpen}
        form={editForm}
        error={editError}
        isSubmitting={isUpdatingProfile}
        hasChanges={hasProfileEditChanges}
        onOpenChange={setEditProfileOpen}
        onFormChange={setEditForm}
        onSubmit={handleUpdateProfile}
      />
    </div>
  );
}

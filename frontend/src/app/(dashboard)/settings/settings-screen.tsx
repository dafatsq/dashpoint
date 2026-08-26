"use client";

import { useEffect, useMemo, useState } from "react";
import { Loader2, Save } from "lucide-react";

import { Header } from "@/components/layout/header";
import { Button } from "@/components/ui/button";
import { getRememberMeKey } from "@/lib/session";
import { useAuth } from "@/contexts/auth-context";
import { useGlobalError } from "@/contexts/error-context";
import { AccountManager } from "@/lib/account-manager";
import api from "@/lib/api";

import { SettingsAuthCard } from "./settings-auth-card";
import { SettingsEditProfileDialog } from "./settings-edit-profile-dialog";
import {
  buildProfileUpdatePayload,
  buildSettingsPreferences,
  hasSettingsPreferenceChanges,
  normalizeSettingsPreferences,
  profileHasChanges,
  type SettingsPreferences,
  type SettingsProfileForm,
} from "./settings-helpers";
import { SettingsProfileCard } from "./settings-profile-card";
import { SettingsVerifyPasswordDialog } from "./settings-verify-password-dialog";
import {
  captureDeviceSessionState,
  restoreDeviceSessionState,
} from "./verify-session-guard";

function preferencesAreEqual(
  left: SettingsPreferences,
  right: SettingsPreferences,
) {
  return (
    left.rememberMe === right.rememberMe &&
    left.quickAccess === right.quickAccess
  );
}

export function SettingsScreen() {
  const { user, login, refreshUser } = useAuth();
  const [isSaving, setIsSaving] = useState(false);
  const [preferences, setPreferences] = useState<SettingsPreferences>({
    rememberMe: true,
    quickAccess: false,
  });
  const [initialPreferences, setInitialPreferences] =
    useState<SettingsPreferences>({
      rememberMe: true,
      quickAccess: false,
    });
  const [verifyPasswordOpen, setVerifyPasswordOpen] = useState(false);
  const [editProfileOpen, setEditProfileOpen] = useState(false);
  const [passwordEntry, setPasswordEntry] = useState("");
  const { showError } = useGlobalError();
  const [isVerifyingPassword, setIsVerifyingPassword] = useState(false);
  const [editForm, setEditForm] = useState<SettingsProfileForm>({
    name: "",
    email: "",
    password: "",
    pin: "",
  });
  const [isUpdatingProfile, setIsUpdatingProfile] = useState(false);

  useEffect(() => {
    if (!user) return;

    const timer = window.setTimeout(() => {
      const preferenceKey = getRememberMeKey(user.id);
      const nextPreferences = buildSettingsPreferences(
        localStorage.getItem(preferenceKey),
        AccountManager.getAccount(user.id) !== null,
      );
      setPreferences((current) =>
        preferencesAreEqual(current, nextPreferences)
          ? current
          : nextPreferences,
      );
      setInitialPreferences((current) =>
        preferencesAreEqual(current, nextPreferences)
          ? current
          : nextPreferences,
      );
    }, 0);

    return () => window.clearTimeout(timer);
  }, [user]);

  const hasPreferenceChanges = hasSettingsPreferenceChanges(
    preferences,
    initialPreferences,
  );

  const hasProfileEditChanges = useMemo(() => {
    if (!user) return false;
    return profileHasChanges(user, editForm);
  }, [editForm, user]);

  const handleSave = async () => {
    if (!user) return;

    setIsSaving(true);

    const nextPreferences = normalizeSettingsPreferences(preferences);

    const preferenceKey = getRememberMeKey(user.id);
    localStorage.setItem(
      preferenceKey,
      nextPreferences.rememberMe ? "true" : "false",
    );
    // Remember-me now takes effect on the refresh cookie issued at the next
    // login; there are no stored tokens to migrate between web storages.

    if (nextPreferences.quickAccess) {
      AccountManager.saveAccount({
        id: user.id,
        name: user.name,
        email: user.email,
        role_name: user.role_name,
        has_pin: user.has_pin,
      });
      localStorage.setItem("dashpoint_device_trusted", "true");
    } else {
      AccountManager.removeAccount(user.id);
      localStorage.removeItem("dashpoint_device_trusted");
    }

    setPreferences(nextPreferences);
    setInitialPreferences(nextPreferences);

    await new Promise((resolve) => setTimeout(resolve, 500));
    setIsSaving(false);
  };

  const handleOpenEditProfile = () => {
    setPasswordEntry("");
    setVerifyPasswordOpen(true);
  };

  const handleVerifyPassword = async (event: React.FormEvent) => {
    event.preventDefault();
    if (!user) return;
    if (!passwordEntry) {
      showError("Password Required", "Enter your password before verifying.");
      return;
    }

    setIsVerifyingPassword(true);

    try {
      // This login is only an identity check: omitting saveAccount keeps the
      // refresh cookie and remembered-device state exactly as they were.
      const deviceSessionState = captureDeviceSessionState(user.id);
      const result = await login(user.email || "", passwordEntry);
      if (result.success) {
        restoreDeviceSessionState(deviceSessionState, user.id);
        setVerifyPasswordOpen(false);
        setEditForm({
          name: user.name,
          email: user.email || "",
          password: "",
          pin: "",
        });
        setEditProfileOpen(true);
      } else {
        showError("Verification Failed", result.error || "Invalid Password");
      }
    } catch {
      showError("Verification Error", "An unexpected error occurred");
    } finally {
      setIsVerifyingPassword(false);
    }
  };

  const handleUpdateProfile = async (event: React.FormEvent) => {
    event.preventDefault();
    if (!user) return;
    if (!editForm.name) {
      showError("Name Required", "Name is required.");
      return;
    }
    if (!hasProfileEditChanges) {
      showError("No Changes", "Make a change before saving.");
      return;
    }

    setIsUpdatingProfile(true);

    try {
      const payload = buildProfileUpdatePayload(user, editForm);
      // The server requires proof of the current credential when a user
      // changes their own password/PIN; reuse the password that was just
      // verified to open this dialog.
      const updatesCredentials = Boolean(payload.password || payload.pin);
      const result = await api.updateUser(
        user.id,
        {
          ...payload,
          ...(updatesCredentials ? { current_password: passwordEntry } : {}),
          expected_updated_at: user.updated_at,
        },
      );
      if (result.error) {
        showError("Update Failed", result.error);
      } else {
        await refreshUser();
        setEditProfileOpen(false);
      }
    } catch {
      showError("Update Error", "Failed to update profile. Please try again.");
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
            <Button
              onClick={handleSave}
              disabled={isSaving || !hasPreferenceChanges}
              className="w-full sm:w-auto"
            >
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
        isSubmitting={isVerifyingPassword}
        onOpenChange={setVerifyPasswordOpen}
        onPasswordChange={setPasswordEntry}
        onSubmit={handleVerifyPassword}
      />

      <SettingsEditProfileDialog
        open={editProfileOpen}
        form={editForm}
        isSubmitting={isUpdatingProfile}
        hasChanges={hasProfileEditChanges}
        onOpenChange={setEditProfileOpen}
        onFormChange={setEditForm}
        onSubmit={handleUpdateProfile}
      />
    </div>
  );
}

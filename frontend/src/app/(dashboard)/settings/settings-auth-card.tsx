'use client';

import { ShieldCheck } from "lucide-react";

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";

import {
  updateQuickAccessPreference,
  updateRememberMePreference,
  type SettingsPreferences,
} from "./settings-helpers";

interface SettingsAuthCardProps {
  preferences: SettingsPreferences;
  onPreferencesChange: (nextPreferences: SettingsPreferences) => void;
}

export function SettingsAuthCard({ preferences, onPreferencesChange }: SettingsAuthCardProps) {
  return (
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
              Keep this account signed in across browser restarts. Enabling this automatically enables Quick Access.
            </p>
          </div>
          <Switch
            checked={preferences.rememberMe}
            onCheckedChange={(checked) => {
              onPreferencesChange(updateRememberMePreference(preferences, checked));
            }}
          />
        </div>

        <div className="flex items-center justify-between gap-4 border-t pt-4">
          <div className="flex-1 space-y-0.5">
            <Label>Quick Access (Save Login)</Label>
            <p className="text-sm text-muted-foreground">
              Save this account on this device so you can sign back in instantly using your Quick PIN.
            </p>
          </div>
          <Switch
            checked={preferences.quickAccess}
            onCheckedChange={(checked) => {
              onPreferencesChange(updateQuickAccessPreference(preferences, checked));
            }}
          />
        </div>
      </CardContent>
    </Card>
  );
}

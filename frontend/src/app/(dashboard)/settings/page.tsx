'use client';

import { useState, useEffect } from 'react';
import { Header } from '@/components/layout/header';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Switch } from '@/components/ui/switch';
import {
  ShieldCheck,
  Save,
  Loader2,
} from 'lucide-react';
import { getRememberMeKey, migrateSession } from '@/lib/session';
import { useAuth } from '@/contexts/auth-context';
import { AccountManager } from '@/lib/account-manager';

export default function SettingsPage() {
  const { user } = useAuth();
  const [isSaving, setIsSaving] = useState(false);
  const [rememberMe, setRememberMe] = useState(true);
  const [quickAccess, setQuickAccess] = useState(false);
  const [initialRememberMe, setInitialRememberMe] = useState(true);
  const [initialQuickAccess, setInitialQuickAccess] = useState(false);

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

  return (
    <div className="flex flex-col h-full">
      <Header title="Settings" />

      <div className="flex-1 p-6 overflow-auto">
        <div className="max-w-3xl mx-auto space-y-6">

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
    </div>
  );
}

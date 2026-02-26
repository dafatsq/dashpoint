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
import { REMEMBER_ME_KEY } from '@/lib/session';

export default function SettingsPage() {
  const [isSaving, setIsSaving] = useState(false);
  const [rememberMe, setRememberMe] = useState(true);

  // Load saved preference on mount
  useEffect(() => {
    const savedPref = localStorage.getItem(REMEMBER_ME_KEY);
    // Default is true if not explicitly set to false
    if (savedPref === 'false') {
      // eslint-disable-next-line react-hooks/exhaustive-deps
      setRememberMe(false);
    } else {
      // eslint-disable-next-line react-hooks/exhaustive-deps
      setRememberMe(true);
    }
  }, []);

  const handleSave = async () => {
    setIsSaving(true);

    // Save the auth preference globally
    localStorage.setItem(REMEMBER_ME_KEY, rememberMe ? 'true' : 'false');

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
              <div className="flex items-center justify-between">
                <div>
                  <Label>Automatic Sign-In (Remember Me)</Label>
                  <p className="text-sm text-muted-foreground">
                    Keep me signed in across browser restarts. Turning this off is recommended for shared devices.
                  </p>
                </div>
                <Switch
                  checked={rememberMe}
                  onCheckedChange={setRememberMe}
                />
              </div>
            </CardContent>
          </Card>

          {/* Save Button */}
          <div className="flex justify-end">
            <Button onClick={handleSave} disabled={isSaving} size="lg">
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

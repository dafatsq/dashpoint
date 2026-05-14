'use client';

import { ShieldCheck } from "lucide-react";

import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";

interface LoginSettingsPanelProps {
  isDeviceTrusted: boolean;
  showDemoAccess: boolean;
  onDeviceTrustedChange: (checked: boolean) => void;
  onDemoAccessChange: (checked: boolean) => void;
}

export function LoginSettingsPanel({
  isDeviceTrusted,
  showDemoAccess,
  onDeviceTrustedChange,
  onDemoAccessChange,
}: LoginSettingsPanelProps) {
  return (
    <div className="mb-6 space-y-4 rounded-lg border bg-card p-4 text-card-foreground shadow-sm">
      <div className="flex items-center justify-between">
        <div className="space-y-0.5">
          <Label className="flex items-center gap-2 text-base font-semibold">
            <ShieldCheck className="h-4 w-4 text-primary" /> Auto Login (Quick
            Access)
          </Label>
          <p className="text-xs text-muted-foreground">
            Save accounts on this browser
          </p>
        </div>
        <Switch
          checked={isDeviceTrusted}
          onCheckedChange={onDeviceTrustedChange}
        />
      </div>

      <div className="flex items-center justify-between">
        <div className="space-y-0.5">
          <Label className="text-base font-semibold">Quick Demo Login</Label>
          <p className="text-xs text-muted-foreground">
            Show demo auto-fill buttons
          </p>
        </div>
        <Switch checked={showDemoAccess} onCheckedChange={onDemoAccessChange} />
      </div>
    </div>
  );
}

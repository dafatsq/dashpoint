"use client";

import { Loader2, ShieldCheck } from "lucide-react";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import type { Role } from "@/types";

import {
  ROLE_PERMISSION_GROUPS,
  toggleRolePermissionKey,
} from "./users-role-permissions";

interface UsersRolePermissionsDialogProps {
  open: boolean;
  role: Role | null;
  selectedPermissions: string[];
  isSubmitting: boolean;
  hasChanges: boolean;
  onOpenChange: (open: boolean) => void;
  onPermissionsChange: (permissions: string[]) => void;
  onSubmit: () => void;
}

export function UsersRolePermissionsDialog({
  open,
  role,
  selectedPermissions,
  isSubmitting,
  hasChanges,
  onOpenChange,
  onPermissionsChange,
  onSubmit,
}: UsersRolePermissionsDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <ShieldCheck className="h-5 w-5" />
            Edit Role Permissions
          </DialogTitle>
          <DialogDescription>
            {role ? `Configure page access for ${role.name}.` : "Configure role permissions."}
          </DialogDescription>
        </DialogHeader>

        <div className="divide-y divide-border">
          {ROLE_PERMISSION_GROUPS.map((group) => {
            const accessEnabled = selectedPermissions.includes(group.accessKey);
            const manageEnabled = group.manageKey ? selectedPermissions.includes(group.manageKey) : false;

            return (
              <div key={group.category} className="py-4 space-y-3">
                <div>
                  <div className="font-medium">{group.label}</div>
                  <div className="text-sm text-muted-foreground">
                    {group.description}
                  </div>
                </div>
                <div className="flex flex-col gap-2.5 pl-3">
                  <div className="flex items-center justify-between gap-3">
                    <Label htmlFor={`${group.category}-access`} className="text-sm font-normal">
                      Access
                    </Label>
                    <Switch
                      id={`${group.category}-access`}
                      checked={accessEnabled}
                      onCheckedChange={(checked) =>
                        onPermissionsChange(toggleRolePermissionKey(selectedPermissions, group.accessKey, checked))
                      }
                    />
                  </div>
                  {group.manageKey ? (
                    <div className="flex items-center justify-between gap-3">
                      <Label htmlFor={`${group.category}-manage`} className="text-sm font-normal">
                        {group.manageLabel ?? "Manage"}
                      </Label>
                      <Switch
                        id={`${group.category}-manage`}
                        checked={manageEnabled}
                        onCheckedChange={(checked) =>
                          onPermissionsChange(toggleRolePermissionKey(selectedPermissions, group.manageKey!, checked))
                        }
                      />
                    </div>
                  ) : null}
                </div>
              </div>
            );
          })}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={isSubmitting}>
            Cancel
          </Button>
          <Button onClick={onSubmit} disabled={isSubmitting || !role || !hasChanges}>
            {isSubmitting ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Saving...
              </>
            ) : (
              "Save Permissions"
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

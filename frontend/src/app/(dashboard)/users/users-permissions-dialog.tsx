"use client";

import {
  BarChart3,
  CheckCircle,
  Layers,
  Loader2,
  Lock,
  Package,
  Receipt,
  Settings2,
  ShieldAlert,
  ShoppingCart,
  Users,
  Wallet,
} from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Switch } from "@/components/ui/switch";
import type { Permission, PermissionOverride, User } from "@/types";

import {
  CATEGORY_ORDER,
  currentUserCanGrantPermission,
  getPermissionDisplayName,
  getPermissionOverride,
  getVisibleChangesCount,
  HIDDEN_PERMS,
  isPermissionDisabledByParent,
  isPermissionEnabled,
  isViewPermission,
  normalizeDeleteArchiveWording,
  REPLACEMENT_PARENT_TOGGLES,
  sortPermissions,
} from "./users-helpers";

interface UsersPermissionsDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  permissionsUser: User | null;
  currentUser: User | null;
  allPermissions: Record<string, Permission[]>;
  userEffectivePermissions: string[];
  userOverrides: PermissionOverride[];
  permissionChanges: Record<string, boolean | null>;
  isLoadingPermissions: boolean;
  isSubmitting: boolean;

  saveSuccess: boolean;
  onTogglePermission: (permission: Permission, enabled: boolean) => void;
  onSaveChanges: () => void;
  onCancel: () => void;
}

const nestedPermissionKeysByParent: Record<string, string[]> = {
  can_create_user: [
    "can_create_manager_users",
    "can_create_cashier_users",
  ],
  can_edit_user: [
    "can_edit_manager_users",
    "can_edit_cashier_users",
  ],
  can_delete_user: [
    "can_delete_manager_users",
    "can_delete_cashier_users",
  ],
  can_manage_permissions: [
    "can_manage_manager_permissions",
    "can_manage_cashier_permissions",
  ],
};

function getCategoryIcon(category: string) {
  switch (category) {
    case "pos":
      return <ShoppingCart className="h-4 w-4" />;
    case "sales":
      return <Receipt className="h-4 w-4" />;
    case "categories":
      return <Layers className="h-4 w-4" />;
    case "inventory":
      return <Package className="h-4 w-4" />;
    case "reports":
      return <BarChart3 className="h-4 w-4" />;
    case "expenses":
      return <Wallet className="h-4 w-4" />;
    case "users":
      return <Users className="h-4 w-4" />;
    case "system":
      return <Settings2 className="h-4 w-4" />;
    default:
      return <Lock className="h-4 w-4" />;
  }
}

export function UsersPermissionsDialog({
  open,
  onOpenChange,
  permissionsUser,
  currentUser,
  allPermissions,
  userEffectivePermissions,
  userOverrides,
  permissionChanges,
  isLoadingPermissions,
  isSubmitting,

  saveSuccess,
  onTogglePermission,
  onSaveChanges,
  onCancel,
}: UsersPermissionsDialogProps) {
  const visibleChangesCount = getVisibleChangesCount(
    permissionChanges,
    allPermissions,
    permissionsUser,
  );

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="w-[90%] sm:w-full max-w-md max-h-[80vh] flex flex-col overflow-hidden">
        <DialogHeader className="flex-shrink-0">
          <DialogTitle className="flex items-center gap-2">
            <Settings2 className="h-5 w-5" />
            Manage Permissions
          </DialogTitle>
          <DialogDescription>
            Configure individual permissions for{" "}
            <span className="font-semibold">{permissionsUser?.name}</span> (
            {permissionsUser?.role_name}). These overrides will grant or deny access
            regardless of the user&apos;s role.
          </DialogDescription>
        </DialogHeader>

        <div className="flex-1 overflow-y-auto">
          {isLoadingPermissions ? (
            <div className="flex items-center justify-center py-12">
              <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
          ) : (
            <div className="h-full overflow-y-auto custom-scrollbar">
              <div className="space-y-6 pb-6">
                {Object.entries(allPermissions)
                  .sort(([left], [right]) => {
                    const leftIndex = CATEGORY_ORDER.indexOf(left as (typeof CATEGORY_ORDER)[number]);
                    const rightIndex = CATEGORY_ORDER.indexOf(right as (typeof CATEGORY_ORDER)[number]);
                    return (leftIndex === -1 ? 99 : leftIndex) - (rightIndex === -1 ? 99 : rightIndex);
                  })
                  .map(([category, permissions]) => {
                    const sortedPermissions = sortPermissions(permissions);

                    return (
                      <Card
                        key={category}
                        className="overflow-hidden border-none shadow-sm bg-muted/20"
                      >
                        <CardHeader className="pb-3 pt-4 px-4 border-b bg-muted/30">
                          <CardTitle className="text-base flex items-center gap-2 capitalize">
                            {getCategoryIcon(category)}
                            {category === "pos" ? "POS" : category} Permissions
                          </CardTitle>
                        </CardHeader>
                        <CardContent className="p-0">
                          <div className="divide-y divide-border/50">
                            {sortedPermissions.map((permission) => {
                              if (HIDDEN_PERMS.includes(permission.key)) {
                                return null;
                              }

                              const enabled = isPermissionEnabled(
                                permission,
                                permissionChanges,
                                userOverrides,
                                userEffectivePermissions,
                              );
                              const viewPermission = isViewPermission(permission);
                              const disabledByParent = isPermissionDisabledByParent(
                                permission,
                                category,
                                allPermissions,
                                permissionChanges,
                                userOverrides,
                                userEffectivePermissions,
                              );
                              const replacementParentToggles =
                                REPLACEMENT_PARENT_TOGGLES[permission.key] || [];
                              const cannotGrant = !currentUserCanGrantPermission(
                                permission,
                                currentUser,
                                permissionsUser,
                              );
                              const switchDisabled = disabledByParent || cannotGrant;
                              const override = getPermissionOverride(
                                permission.id,
                                userOverrides,
                              );

                              const statusBadge =
                                permissionChanges[permission.id] !== undefined ? (
                                  <Badge
                                    variant="outline"
                                    className="text-yellow-600 border-yellow-200 bg-yellow-50"
                                  >
                                    Modified
                                  </Badge>
                                ) : override ? (
                                  <Badge
                                    variant="outline"
                                    className={
                                      override.allowed
                                        ? "text-green-600 border-green-200 bg-green-50"
                                        : "text-red-600 border-red-200 bg-red-50"
                                    }
                                  >
                                    {override.allowed ? "Granted" : "Denied"}
                                  </Badge>
                                ) : (
                                  <Badge
                                    variant="secondary"
                                    className="text-muted-foreground font-normal"
                                  >
                                    Default
                                  </Badge>
                                );

                              const nestedPermissions =
                                permissionsUser?.role_name === "manager"
                                  ? permissions.filter((entry) =>
                                      (nestedPermissionKeysByParent[permission.key] || []).includes(
                                        entry.key,
                                      ),
                                    )
                                  : [];

                              return (
                                <div key={permission.id} className="flex flex-col">
                                  {replacementParentToggles.length === 0 && (
                                    <div
                                      className={`flex items-start justify-between p-4 transition-colors ${
                                        disabledByParent
                                          ? "opacity-50 bg-muted/10"
                                          : cannotGrant
                                            ? "opacity-60 bg-muted/10"
                                            : "hover:bg-muted/30"
                                      }`}
                                    >
                                      <div className="flex-1 mr-4">
                                        <div className="flex items-center gap-2 mb-1">
                                          <span
                                            className={`font-medium text-sm ${
                                              disabledByParent ? "text-muted-foreground" : ""
                                            }`}
                                          >
                                            {getPermissionDisplayName(permission, category)}
                                          </span>
                                          {viewPermission && (
                                            <Badge
                                              variant="outline"
                                              className="text-[10px] h-5 px-1.5 text-blue-600 border-blue-200 bg-blue-50"
                                            >
                                              Access
                                            </Badge>
                                          )}
                                          {!disabledByParent && statusBadge}
                                        </div>

                                        {permission.description && (
                                          <p className="text-sm text-muted-foreground leading-relaxed">
                                            {normalizeDeleteArchiveWording(
                                              permission.description,
                                            )}
                                          </p>
                                        )}

                                        {disabledByParent && (
                                          <p className="text-xs text-orange-600 mt-1.5 flex items-center gap-1.5">
                                            <ShieldAlert className="h-3 w-3" />
                                            {permission.key === "can_void_sale"
                                              ? "Requires Sales History Access"
                                              : `Requires ${category.charAt(0).toUpperCase() + category.slice(1)} Access`}
                                          </p>
                                        )}
                                        {!disabledByParent && cannotGrant && (
                                          <p className="text-xs text-muted-foreground mt-1.5 flex items-center gap-1.5">
                                            <Lock className="h-3 w-3" />
                                            You don&apos;t have this permission
                                          </p>
                                        )}
                                      </div>

                                      <div className="flex items-center h-6 mt-1">
                                        <Switch
                                          checked={disabledByParent ? false : enabled}
                                          onCheckedChange={(checked) =>
                                            onTogglePermission(permission, checked)
                                          }
                                          disabled={switchDisabled}
                                        />
                                      </div>
                                    </div>
                                  )}

                                  {replacementParentToggles.map((actionItem) => (
                                    <div
                                      key={`${permission.id}-${actionItem.label}`}
                                      className={`flex items-start justify-between p-4 transition-colors ${
                                        disabledByParent
                                          ? "opacity-50 bg-muted/10"
                                          : cannotGrant
                                            ? "opacity-60 bg-muted/10"
                                            : "hover:bg-muted/30"
                                      }`}
                                    >
                                      <div className="flex-1 mr-4">
                                        <div className="flex items-center gap-2 mb-1">
                                          <span
                                            className={`font-medium text-sm ${
                                              disabledByParent ? "text-muted-foreground" : ""
                                            }`}
                                          >
                                            {actionItem.label}
                                          </span>
                                          {!disabledByParent && statusBadge}
                                        </div>
                                        <p className="text-sm text-muted-foreground leading-relaxed">
                                          {actionItem.description}
                                        </p>
                                      </div>
                                      <div className="flex items-center h-6 mt-1">
                                        <Switch
                                          checked={disabledByParent ? false : enabled}
                                          onCheckedChange={(checked) =>
                                            onTogglePermission(permission, checked)
                                          }
                                          disabled={switchDisabled}
                                        />
                                      </div>
                                    </div>
                                  ))}

                                  {nestedPermissions.length > 0 && (
                                    <div className="pl-12 pr-4 pb-4 bg-muted/5 border-t border-border/50 pt-3">
                                      <div className="space-y-3">
                                        {nestedPermissions.map((nestedPermission) => {
                                          const nestedEnabled = isPermissionEnabled(
                                            nestedPermission,
                                            permissionChanges,
                                            userOverrides,
                                            userEffectivePermissions,
                                          );
                                          const nestedCannotGrant =
                                            !currentUserCanGrantPermission(
                                              nestedPermission,
                                              currentUser,
                                              permissionsUser,
                                            );
                                          const parentDisabled = !enabled;

                                          return (
                                            <div
                                              key={nestedPermission.id}
                                              className={`flex items-center justify-between ${
                                                parentDisabled || switchDisabled
                                                  ? "opacity-50"
                                                  : nestedCannotGrant
                                                    ? "opacity-60"
                                                    : ""
                                              }`}
                                            >
                                              <div className="flex-1 mr-4">
                                                <div className="text-sm font-medium text-foreground/90">
                                                  {getPermissionDisplayName(
                                                    nestedPermission,
                                                    category,
                                                  )}
                                                </div>
                                                {nestedPermission.description && (
                                                  <div className="text-xs text-muted-foreground mt-0.5">
                                                    {normalizeDeleteArchiveWording(
                                                      nestedPermission.description,
                                                    )}
                                                  </div>
                                                )}
                                              </div>
                                              <Switch
                                                checked={parentDisabled ? false : nestedEnabled}
                                                onCheckedChange={(checked) =>
                                                  onTogglePermission(
                                                    nestedPermission,
                                                    checked,
                                                  )
                                                }
                                                disabled={
                                                  parentDisabled ||
                                                  nestedCannotGrant ||
                                                  switchDisabled
                                                }
                                              />
                                            </div>
                                          );
                                        })}
                                      </div>
                                    </div>
                                  )}
                                </div>
                              );
                            })}
                          </div>
                        </CardContent>
                      </Card>
                    );
                  })}
              </div>
            </div>
          )}
        </div>

        <DialogFooter className="flex-shrink-0 flex items-center justify-between gap-2 pt-4 border-t bg-background">
          <div className="text-sm">
            {saveSuccess ? (
              <span className="text-green-600 flex items-center gap-1.5 font-medium animate-in fade-in zoom-in duration-300">
                <CheckCircle className="h-4 w-4 flex-shrink-0" />
                Permissions saved successfully!
              </span>
            ) : visibleChangesCount > 0 ? (
              <span className="text-yellow-600">
                {visibleChangesCount} unsaved change(s)
              </span>
            ) : null}
          </div>
          <div className="flex gap-2">
            <Button variant="outline" onClick={onCancel}>
              Cancel
            </Button>
            <Button
              onClick={onSaveChanges}
              disabled={isSubmitting}
            >
              {isSubmitting ? (
                <>
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  Saving...
                </>
              ) : (
                "Save Changes"
              )}
            </Button>
          </div>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

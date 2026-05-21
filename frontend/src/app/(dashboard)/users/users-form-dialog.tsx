"use client";

import { Loader2, Shield, ShieldAlert, ShieldCheck } from "lucide-react";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import type { CreateUserRequest, User, UserRole } from "@/types";

interface UsersFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  editingUser: User | null;
  currentUser: User | null;
  availableRoles: UserRole[];
  formData: CreateUserRequest;
  setFormData: React.Dispatch<React.SetStateAction<CreateUserRequest>>;
  formErrors: { general?: string };
  isSubmitting: boolean;
  onSubmit: () => void;
}

export function UsersFormDialog({
  open,
  onOpenChange,
  editingUser,
  currentUser,
  availableRoles,
  formData,
  setFormData,
  formErrors,
  isSubmitting,
  onSubmit,
}: UsersFormDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{editingUser ? "Edit User" : "Add User"}</DialogTitle>
          <DialogDescription>
            {editingUser
              ? "Update the user details below."
              : "Fill in the details for the new user."}
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
              onChange={(event) =>
                setFormData((current) => ({ ...current, name: event.target.value }))
              }
              placeholder="Full name"
            />
          </div>

          <div className="grid gap-2">
            <Label htmlFor="email">Email {!editingUser && "*"}</Label>
            <Input
              id="email"
              type="email"
              value={formData.email}
              onChange={(event) =>
                setFormData((current) => ({ ...current, email: event.target.value }))
              }
              placeholder="user@example.com"
            />
          </div>

          <div className="grid gap-2">
            <Label htmlFor="password">
              Password {!editingUser ? "*" : "(leave blank to keep current)"}
            </Label>
            <Input
              id="password"
              type="password"
              value={formData.password}
              onChange={(event) =>
                setFormData((current) => ({ ...current, password: event.target.value }))
              }
              placeholder={
                editingUser ? "Leave blank to keep current" : "Enter password"
              }
            />
          </div>

          <div className="grid gap-2">
            <Label htmlFor="pin">PIN (4-6 digits, for quick login) {!editingUser && "*"}</Label>
            <Input
              id="pin"
              type="password"
              value={formData.pin}
              onChange={(event) =>
                setFormData((current) => ({
                  ...current,
                  pin: event.target.value.replace(/\D/g, "").slice(0, 6),
                }))
              }
              placeholder={
                editingUser ? "Leave blank to keep current" : "Optional 4-6 digit PIN"
              }
              maxLength={6}
            />
          </div>

          <div className="grid gap-2">
            <Label htmlFor="role">Role</Label>
            {editingUser?.id === currentUser?.id && (
              <p className="text-xs text-muted-foreground -mt-1">
                You cannot change your own role.
              </p>
            )}
            {!editingUser && availableRoles.length === 0 && (
              <p className="text-xs text-muted-foreground -mt-1">
                No user roles are currently assignable with your permissions.
              </p>
            )}
            <Select
              value={formData.role}
              disabled={editingUser?.id === currentUser?.id || availableRoles.length === 0}
              onValueChange={(value: UserRole) =>
                setFormData((current) => ({ ...current, role: value }))
              }
            >
              <SelectTrigger>
                <SelectValue placeholder="Select role" />
              </SelectTrigger>
              <SelectContent>
                {availableRoles.includes("owner") && (
                  <SelectItem value="owner">
                    <div className="flex items-center gap-2">
                      <ShieldAlert className="h-4 w-4" />
                      Owner - Full access
                    </div>
                  </SelectItem>
                )}
                {availableRoles.includes("manager") && (
                  <SelectItem value="manager">
                    <div className="flex items-center gap-2">
                      <ShieldCheck className="h-4 w-4" />
                      Manager - Most access
                    </div>
                  </SelectItem>
                )}
                {availableRoles.includes("cashier") && (
                  <SelectItem value="cashier">
                    <div className="flex items-center gap-2">
                      <Shield className="h-4 w-4" />
                      Cashier - Sales only
                    </div>
                  </SelectItem>
                )}
              </SelectContent>
            </Select>
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button
            onClick={onSubmit}
            disabled={isSubmitting}
          >
            {isSubmitting ? (
              <>
                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                Saving...
              </>
            ) : editingUser ? (
              "Update User"
            ) : (
              "Create User"
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

'use client';

import { AlertCircle, Loader2 } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

import type { SettingsProfileForm } from "./settings-helpers";

interface SettingsEditProfileDialogProps {
  open: boolean;
  form: SettingsProfileForm;
  error: string;
  isSubmitting: boolean;
  hasChanges: boolean;
  onOpenChange: (open: boolean) => void;
  onFormChange: (form: SettingsProfileForm) => void;
  onSubmit: (event: React.FormEvent) => void;
}

export function SettingsEditProfileDialog({
  open,
  form,
  error,
  isSubmitting,
  hasChanges,
  onOpenChange,
  onFormChange,
  onSubmit,
}: SettingsEditProfileDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Edit Profile</DialogTitle>
          <DialogDescription>Update your personal details below.</DialogDescription>
        </DialogHeader>
        <form onSubmit={onSubmit} className="mt-4 space-y-4" autoComplete="off">
          {error ? (
            <div className="flex items-center gap-2 rounded-md bg-destructive/10 p-3 text-sm text-destructive">
              <AlertCircle className="h-4 w-4 flex-shrink-0" />
              <span>{error}</span>
            </div>
          ) : null}
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="edit-name">Name <span className="text-destructive">*</span></Label>
              <Input
                id="edit-name"
                value={form.name}
                onChange={(event) => onFormChange({ ...form, name: event.target.value })}
                required
                disabled={isSubmitting}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="edit-email">Email</Label>
              <Input
                id="edit-email"
                type="email"
                value={form.email}
                onChange={(event) => onFormChange({ ...form, email: event.target.value })}
                disabled={isSubmitting}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="edit-password">Password (leave blank to keep current)</Label>
              <Input
                id="edit-password"
                type="password"
                placeholder="Leave blank to keep current"
                value={form.password}
                onChange={(event) => onFormChange({ ...form, password: event.target.value })}
                disabled={isSubmitting}
                readOnly
                onFocus={(event) => event.target.removeAttribute("readonly")}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="edit-pin">PIN (leave blank to keep current)</Label>
              <Input
                id="edit-pin"
                type="password"
                placeholder="Leave blank to keep current"
                value={form.pin}
                onChange={(event) => onFormChange({ ...form, pin: event.target.value })}
                maxLength={6}
                pattern="\\d*"
                inputMode="numeric"
                disabled={isSubmitting}
                readOnly
                onFocus={(event) => event.target.removeAttribute("readonly")}
              />
            </div>
          </div>
          <div className="flex justify-end gap-2 pt-2">
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)} disabled={isSubmitting}>
              Cancel
            </Button>
            <Button type="submit" disabled={isSubmitting || !hasChanges || !form.name}>
              {isSubmitting ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Saving...
                </>
              ) : (
                "Save Changes"
              )}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}

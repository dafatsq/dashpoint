"use client";

import { Loader2 } from "lucide-react";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";

interface SettingsVerifyPasswordDialogProps {
  open: boolean;
  password: string;

  isSubmitting: boolean;
  onOpenChange: (open: boolean) => void;
  onPasswordChange: (value: string) => void;
  onSubmit: (event: React.FormEvent) => void;
}

export function SettingsVerifyPasswordDialog({
  open,
  password,
  isSubmitting,
  onOpenChange,
  onPasswordChange,
  onSubmit,
}: SettingsVerifyPasswordDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Verify Password</DialogTitle>
          <DialogDescription>
            Please enter your password to edit your profile.
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={onSubmit} className="mt-4 space-y-4" autoComplete="off">
          <Input
            type="password"
            placeholder="Enter your password"
            value={password}
            onChange={(event) => onPasswordChange(event.target.value)}
            readOnly
            onFocus={(event) => event.target.removeAttribute("readonly")}
            required
            disabled={isSubmitting}
            autoFocus
          />
          <div className="flex flex-col-reverse gap-2">
            <Button
              type="button"
              variant="outline"
              className="w-full"
              onClick={() => onOpenChange(false)}
              disabled={isSubmitting}
            >
              Cancel
            </Button>
            <Button type="submit" className="w-full" disabled={isSubmitting}>
              {isSubmitting ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Verifying...
                </>
              ) : (
                "Verify"
              )}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}

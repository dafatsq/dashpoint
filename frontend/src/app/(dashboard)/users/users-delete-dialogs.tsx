"use client";

import { Loader2 } from "lucide-react";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import type { User } from "@/types";

interface UsersArchiveDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  deletingUser: User | null;
  isSubmitting: boolean;
  onConfirm: () => void;
}

export function UsersArchiveDialog({
  open,
  onOpenChange,
  deletingUser,
  isSubmitting,
  onConfirm,
}: UsersArchiveDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Archive User</DialogTitle>
          <DialogDescription>
            Are you sure you want to archive &quot;{deletingUser?.name}&quot;? The user
            will be moved to the Archived tab and can be restored later.
          </DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button variant="destructive" onClick={onConfirm} disabled={isSubmitting}>
            {isSubmitting ? (
              <>
                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                Archiving...
              </>
            ) : (
              "Archive"
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

interface UsersPermanentDeleteDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  deletingUser: User | null;
  isSubmitting: boolean;
  deleteError: string | null;
  onConfirm: () => void;
  onResetError: () => void;
}

export function UsersPermanentDeleteDialog({
  open,
  onOpenChange,
  deletingUser,
  isSubmitting,
  deleteError,
  onConfirm,
  onResetError,
}: UsersPermanentDeleteDialogProps) {
  return (
    <Dialog
      open={open}
      onOpenChange={(nextOpen) => {
        onOpenChange(nextOpen);
        if (!nextOpen) {
          onResetError();
        }
      }}
    >
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Permanently Delete User</DialogTitle>
          <DialogDescription>
            Are you sure you want to permanently delete &quot;{deletingUser?.name}
            &quot;? This action cannot be undone. All data associated with this user
            will be lost.
          </DialogDescription>
        </DialogHeader>
        {deleteError && (
          <div className="p-3 rounded-md bg-destructive/10 border border-destructive/20">
            <p className="text-sm text-destructive">{deleteError}</p>
          </div>
        )}
        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => {
              onOpenChange(false);
              onResetError();
            }}
          >
            Cancel
          </Button>
          <Button
            variant="destructive"
            onClick={onConfirm}
            disabled={isSubmitting || !!deleteError}
          >
            {isSubmitting ? (
              <>
                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                Deleting...
              </>
            ) : (
              "Delete Permanently"
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

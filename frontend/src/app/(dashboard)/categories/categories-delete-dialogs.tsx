'use client';

import { AlertCircle, Loader2, ShieldAlert } from "lucide-react";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";

interface CategoriesDeleteDialogsProps {
  archiveOpen: boolean;
  permanentDeleteOpen: boolean;
  categoryName: string;
  error: string | null;
  isSubmitting: boolean;
  onArchiveOpenChange: (open: boolean) => void;
  onPermanentDeleteOpenChange: (open: boolean) => void;
  onArchive: () => void;
  onPermanentDelete: () => void;
}

function ErrorBanner({ error }: { error: string | null }) {
  if (!error) {
    return null;
  }

  return (
    <div className="flex items-center gap-2 p-3 rounded-md bg-destructive/10 text-destructive text-sm border border-destructive/20">
      <AlertCircle className="h-4 w-4" />
      {error}
    </div>
  );
}

export function CategoriesDeleteDialogs({
  archiveOpen,
  permanentDeleteOpen,
  categoryName,
  error,
  isSubmitting,
  onArchiveOpenChange,
  onPermanentDeleteOpenChange,
  onArchive,
  onPermanentDelete,
}: CategoriesDeleteDialogsProps) {
  return (
    <>
      <Dialog open={archiveOpen} onOpenChange={onArchiveOpenChange}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Archive Category</DialogTitle>
            <DialogDescription>
              Are you sure you want to archive <span className="font-semibold text-foreground">{categoryName}</span>?
              It will be moved to the Archived tab.
            </DialogDescription>
          </DialogHeader>

          <ErrorBanner error={error} />

          <DialogFooter className="gap-2 sm:gap-0">
            <Button variant="outline" onClick={() => onArchiveOpenChange(false)} disabled={isSubmitting}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={onArchive} disabled={isSubmitting}>
              {isSubmitting ? <Loader2 className="h-4 w-4 mr-2 animate-spin" /> : null}
              Archive Category
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={permanentDeleteOpen} onOpenChange={onPermanentDeleteOpenChange}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle className="text-destructive flex items-center gap-2">
              <ShieldAlert className="h-5 w-5" />
              Permanent Delete
            </DialogTitle>
            <DialogDescription>
              This will permanently delete <span className="font-semibold text-foreground">{categoryName}</span>.
              This action <span className="font-bold text-destructive">cannot be undone</span>.
            </DialogDescription>
          </DialogHeader>

          <ErrorBanner error={error} />

          <DialogFooter className="gap-2 sm:gap-0">
            <Button variant="outline" onClick={() => onPermanentDeleteOpenChange(false)} disabled={isSubmitting}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={onPermanentDelete} disabled={isSubmitting}>
              {isSubmitting ? <Loader2 className="h-4 w-4 mr-2 animate-spin" /> : null}
              Delete Permanently
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}

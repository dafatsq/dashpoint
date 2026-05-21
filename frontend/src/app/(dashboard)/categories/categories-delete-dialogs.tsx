'use client';

import { Loader2, ShieldAlert } from "lucide-react";

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

  isSubmitting: boolean;
  onArchiveOpenChange: (open: boolean) => void;
  onPermanentDeleteOpenChange: (open: boolean) => void;
  onArchive: () => void;
  onPermanentDelete: () => void;
}

function SubmittingIcon({ isSubmitting }: { isSubmitting: boolean }) {
  return isSubmitting ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null;
}

export function CategoriesDeleteDialogs({
  archiveOpen,
  permanentDeleteOpen,
  categoryName,
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
              Are you sure you want to archive{" "}
              <span className="font-semibold text-foreground">{categoryName}</span>? It
              will be moved to the Archived tab.
            </DialogDescription>
          </DialogHeader>

          <DialogFooter>
            <Button variant="outline" onClick={() => onArchiveOpenChange(false)} disabled={isSubmitting}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={onArchive} disabled={isSubmitting}>
              <SubmittingIcon isSubmitting={isSubmitting} />
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
              This will permanently delete{" "}
              <span className="font-semibold text-foreground">{categoryName}</span>.
              This action{" "}
              <span className="font-bold text-destructive">cannot be undone</span>.
            </DialogDescription>
          </DialogHeader>

          <DialogFooter>
            <Button variant="outline" onClick={() => onPermanentDeleteOpenChange(false)} disabled={isSubmitting}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={onPermanentDelete} disabled={isSubmitting}>
              <SubmittingIcon isSubmitting={isSubmitting} />
              Delete Permanently
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}

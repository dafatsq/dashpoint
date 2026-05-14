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

interface ExpensesDeleteDialogProps {
  open: boolean;
  deleteError: string | null;
  isSubmitting: boolean;
  onOpenChange: (open: boolean) => void;
  onDelete: () => void;
}

export function ExpensesDeleteDialog({
  open,
  deleteError,
  isSubmitting,
  onOpenChange,
  onDelete,
}: ExpensesDeleteDialogProps) {
  return (
    <Dialog
      open={open}
      onOpenChange={(nextOpen) => {
        onOpenChange(nextOpen);
      }}
    >
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Delete Expense</DialogTitle>
          <DialogDescription>Are you sure you want to delete this expense? This action cannot be undone.</DialogDescription>
        </DialogHeader>

        {deleteError && (
          <div className="rounded-md border border-destructive/20 bg-destructive/10 p-3">
            <p className="text-sm text-destructive">{deleteError}</p>
          </div>
        )}

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button variant="destructive" onClick={onDelete} disabled={isSubmitting || !!deleteError}>
            {isSubmitting ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Deleting...
              </>
            ) : (
              "Delete"
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

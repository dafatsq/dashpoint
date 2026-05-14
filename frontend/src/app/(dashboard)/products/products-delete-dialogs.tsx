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
import { Product } from "@/types";

interface ProductsDeleteDialogsProps {
  deletingProduct: Product | null;
  deleteDialogOpen: boolean;
  permanentDeleteDialogOpen: boolean;
  isSubmitting: boolean;
  deleteError: string | null;
  onDeleteDialogOpenChange: (open: boolean) => void;
  onPermanentDeleteDialogOpenChange: (open: boolean) => void;
  onArchive: () => void;
  onPermanentDelete: () => void;
  onClearDeleteError: () => void;
}

export function ProductsDeleteDialogs({
  deletingProduct,
  deleteDialogOpen,
  permanentDeleteDialogOpen,
  isSubmitting,
  deleteError,
  onDeleteDialogOpenChange,
  onPermanentDeleteDialogOpenChange,
  onArchive,
  onPermanentDelete,
  onClearDeleteError,
}: ProductsDeleteDialogsProps) {
  return (
    <>
      <Dialog open={deleteDialogOpen} onOpenChange={onDeleteDialogOpenChange}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Archive Product</DialogTitle>
            <DialogDescription>
              Are you sure you want to archive &quot;{deletingProduct?.name}&quot;? The product will be moved to the
              Archived tab and can be restored later.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => onDeleteDialogOpenChange(false)}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={onArchive} disabled={isSubmitting}>
              {isSubmitting ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Archiving...
                </>
              ) : (
                "Archive"
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog
        open={permanentDeleteDialogOpen}
        onOpenChange={(open) => {
          onPermanentDeleteDialogOpenChange(open);
          if (!open) {
            onClearDeleteError();
          }
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Permanently Delete Product</DialogTitle>
            <DialogDescription>
              Are you sure you want to permanently delete &quot;{deletingProduct?.name}&quot;? This action cannot be
              undone. All data associated with this product will be lost.
            </DialogDescription>
          </DialogHeader>
          {deleteError && (
            <div className="rounded-md border border-destructive/20 bg-destructive/10 p-3">
              <p className="text-sm text-destructive">{deleteError}</p>
            </div>
          )}
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => {
                onPermanentDeleteDialogOpenChange(false);
                onClearDeleteError();
              }}
            >
              Cancel
            </Button>
            <Button variant="destructive" onClick={onPermanentDelete} disabled={isSubmitting || !!deleteError}>
              {isSubmitting ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Deleting...
                </>
              ) : (
                "Delete Permanently"
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}

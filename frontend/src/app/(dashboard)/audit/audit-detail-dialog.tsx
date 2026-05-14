'use client';

import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import type { AuditLog } from "@/types";

import { ActivityFieldChanges } from "../activity-field-changes";
import {
  formatActivityDateTime,
  getActivityActionLabel,
  getActivityEntityLabel,
} from "../activity-helpers";

interface AuditDetailDialogProps {
  open: boolean;
  log: AuditLog | null;
  onOpenChange: (open: boolean) => void;
}

export function AuditDetailDialog({ open, log, onOpenChange }: AuditDetailDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>Log Details</DialogTitle>
          <DialogDescription>{log ? formatActivityDateTime(log.created_at) : ""}</DialogDescription>
        </DialogHeader>

        {log ? (
          <div className="space-y-4">
            <div className="grid grid-cols-2 gap-4 text-sm">
              <div>
                <span className="text-muted-foreground">User:</span>
                <p className="font-medium">{log.user_name || "System"}</p>
              </div>
              <div>
                <span className="text-muted-foreground">Action:</span>
                <p className="font-medium">{getActivityActionLabel(log.action)}</p>
              </div>
              <div>
                <span className="text-muted-foreground">Entity Type:</span>
                <p className="font-medium">{getActivityEntityLabel(log.entity_type)}</p>
              </div>
              <div>
                <span className="text-muted-foreground">Entity ID:</span>
                <p className="font-mono text-xs font-medium">{log.entity_id || "-"}</p>
              </div>
              <div>
                <span className="text-muted-foreground">IP Address:</span>
                <p className="font-medium">{log.ip_address || "-"}</p>
              </div>
              <div>
                <span className="text-muted-foreground">User Agent:</span>
                <p className="truncate text-xs font-medium" title={log.user_agent}>
                  {log.user_agent || "-"}
                </p>
              </div>
            </div>

            <div className="border-t pt-4">
              <h4 className="mb-2 font-medium">Changes</h4>
              <ActivityFieldChanges
                log={log}
                emptyMessage="No detailed changes recorded"
              />
            </div>
          </div>
        ) : null}
      </DialogContent>
    </Dialog>
  );
}

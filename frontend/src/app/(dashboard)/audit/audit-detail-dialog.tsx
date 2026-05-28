"use client";

import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import type { AuditLog } from "@/types";

import { ActivityFieldChanges } from "../activity-field-changes";
import {
  formatActivityDateTime,
  getActivityActionLabel,
  getActivityEntityLabel,
  buildActivityDescription,
} from "../activity-helpers";

interface AuditDetailDialogProps {
  open: boolean;
  log: AuditLog | null;
  onOpenChange: (open: boolean) => void;
}

export function AuditDetailDialog({
  open,
  log,
  onOpenChange,
}: AuditDetailDialogProps) {
  const detailItems = log
    ? [
        { label: "User", value: log.user_name || "System" },
        { label: "Action", value: getActivityActionLabel(log.action) },
        {
          label: "Entity Type",
          value: getActivityEntityLabel(log.entity_type),
        },
        {
          label: "Entity ID",
          value: log.entity_id || "-",
          valueClassName: "font-mono text-xs font-medium",
        },
        {
          label: "Description",
          value: buildActivityDescription(log),
          wrapperClassName: "col-span-2",
        },
      ]
    : [];

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>Log Details</DialogTitle>
          <DialogDescription>
            {log ? formatActivityDateTime(log.created_at) : ""}
          </DialogDescription>
        </DialogHeader>

        {log ? (
          <div className="space-y-4">
            <div className="grid grid-cols-2 gap-4 text-sm">
              {detailItems.map((item) => (
                <div key={item.label} className={item.wrapperClassName}>
                  <span className="text-muted-foreground">{item.label}:</span>
                  <p className={item.valueClassName ?? "font-medium"}>
                    {item.value}
                  </p>
                </div>
              ))}
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

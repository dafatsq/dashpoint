'use client';

import type { AuditLog } from "@/types";

import {
  buildActivityFieldChanges,
  formatActivityFieldName,
  formatActivityFieldValue,
  isActivityImageField,
} from "./activity-helpers";

interface ActivityFieldChangesProps {
  log: AuditLog;
  emptyMessage?: string | null;
}

export function ActivityFieldChanges({ log, emptyMessage = null }: ActivityFieldChangesProps) {
  const changes = buildActivityFieldChanges(log);

  if (changes.length === 0) {
    return emptyMessage ? (
      <p className="text-sm text-muted-foreground">{emptyMessage}</p>
    ) : null;
  }

  return (
    <div className="space-y-2">
      {changes.map(({ key, oldVal, newVal, label }) => (
        <div key={`${key}-${label ?? "default"}`} className="text-sm border-b pb-2 last:border-b-0">
          <div className="font-medium capitalize">
            {label ?? formatActivityFieldName(key, log.action)}
          </div>
          <div className="mt-1 flex items-start gap-2 text-xs text-muted-foreground">
            {isActivityImageField(key) ? (
              <div className="flex items-center gap-2">
                {oldVal !== undefined ? (
                  <div className="flex flex-col items-center gap-1">
                    {/* eslint-disable-next-line @next/next/no-img-element */}
                    <img src={String(oldVal)} alt="old" className="h-8 w-8 rounded border object-cover opacity-60" />
                    <span className="text-[9px] text-red-500">before</span>
                  </div>
                ) : null}
                {oldVal !== undefined && newVal !== undefined ? <span>→</span> : null}
                {newVal !== undefined ? (
                  <div className="flex flex-col items-center gap-1">
                    {/* eslint-disable-next-line @next/next/no-img-element */}
                    <img src={String(newVal)} alt="new" className="h-8 w-8 rounded border object-cover" />
                    <span className="text-[9px] text-green-600">after</span>
                  </div>
                ) : null}
                {newVal === undefined && oldVal !== undefined ? (
                  <span className="text-[10px] text-red-500">(removed)</span>
                ) : null}
              </div>
            ) : oldVal !== undefined && newVal !== undefined ? (
              <>
                <span className="text-red-500 line-through">{formatActivityFieldValue(key, oldVal)}</span>
                <span>→</span>
                <span className="text-green-600">{formatActivityFieldValue(key, newVal)}</span>
              </>
            ) : newVal !== undefined ? (
              <span>{formatActivityFieldValue(key, newVal)}</span>
            ) : (
              <span className="text-red-500 line-through">{formatActivityFieldValue(key, oldVal)}</span>
            )}
          </div>
        </div>
      ))}
    </div>
  );
}

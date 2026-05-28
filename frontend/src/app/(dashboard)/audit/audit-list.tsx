"use client";

import {
  ChevronLeft,
  ChevronRight,
  Eye,
  Loader2,
  ScrollText,
} from "lucide-react";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import type { AuditLog } from "@/types";

import {
  formatActivityDate,
  getActivityActionLabel,
  getActivityBadgeColor,
  getActivityEntityLabel,
} from "../activity-helpers";

interface AuditListProps {
  logs: AuditLog[];
  isLoading: boolean;
  error: string | null;
  page: number;
  hasMore: boolean;
  onRetry: () => void;
  onPageChange: (page: number) => void;
  onViewLog: (log: AuditLog) => void;
}

export function AuditList({
  logs,
  isLoading,
  error,
  page,
  hasMore,
  onRetry,
  onPageChange,
  onViewLog,
}: AuditListProps) {
  const renderDetailsButton = (log: AuditLog) => (
    <Button
      variant="outline"
      size="sm"
      className="h-8"
      onClick={() => onViewLog(log)}
    >
      <Eye className="mr-1 h-3.5 w-3.5" />
      Details
    </Button>
  );

  if (isLoading) {
    return (
      <div className="flex h-64 items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  if (error) {
    return (
      <Card>
        <CardContent className="flex flex-col items-center justify-center gap-3 py-12">
          <ScrollText className="h-12 w-12 text-muted-foreground" />
          <p className="text-sm text-muted-foreground">{error}</p>
          <Button variant="outline" size="sm" onClick={onRetry}>
            Retry
          </Button>
        </CardContent>
      </Card>
    );
  }

  if (logs.length === 0) {
    return (
      <Card>
        <CardContent className="flex flex-col items-center justify-center py-12">
          <ScrollText className="mb-4 h-12 w-12 text-muted-foreground" />
          <p className="text-muted-foreground">No audit logs found</p>
        </CardContent>
      </Card>
    );
  }

  return (
    <>
      <Card className="hidden lg:block">
        <CardHeader>
          <CardTitle>Activity Log</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b text-left text-sm text-muted-foreground">
                  <th className="pb-3 font-medium">Timestamp</th>
                  <th className="pb-3 font-medium">User</th>
                  <th className="pb-3 font-medium">Action</th>
                  <th className="pb-3 font-medium">Entity</th>
                  <th className="pb-3 text-right font-medium">Details</th>
                </tr>
              </thead>
              <tbody>
                {logs.map((log) => (
                  <tr key={log.id} className="border-b last:border-0">
                    <td className="py-3 text-sm">
                      {formatActivityDate(log.created_at)}
                    </td>
                    <td className="py-3 text-sm font-medium">
                      {log.user_name || "System"}
                    </td>
                    <td className="py-3">
                      <span
                        className={`inline-flex rounded-full px-2 py-1 text-xs font-medium ${getActivityBadgeColor(log.action)}`}
                      >
                        {getActivityActionLabel(log.action)}
                      </span>
                    </td>
                    <td className="py-3 text-sm capitalize">
                      {getActivityEntityLabel(log.entity_type)}
                      {log.entity_id ? (
                        <span className="ml-1 text-xs text-muted-foreground">
                          ({log.entity_id.slice(0, 8)}...)
                        </span>
                      ) : null}
                    </td>
                    <td className="py-3 text-right">
                      {renderDetailsButton(log)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </CardContent>
      </Card>

      <div className="space-y-4 lg:hidden">
        {logs.map((log) => (
          <Card key={log.id}>
            <CardContent className="p-4">
              <div className="mb-3 flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <span
                    className={`inline-flex rounded-full px-2 py-1 text-xs font-medium ${getActivityBadgeColor(log.action)}`}
                  >
                    {getActivityActionLabel(log.action)}
                  </span>
                  <span className="text-xs text-muted-foreground">
                    {formatActivityDate(log.created_at)}
                  </span>
                </div>
                {renderDetailsButton(log)}
              </div>

              <div className="grid gap-1">
                <div className="flex justify-between text-sm">
                  <span className="text-muted-foreground">User:</span>
                  <span className="font-medium">
                    {log.user_name || "System"}
                  </span>
                </div>
                <div className="flex justify-between text-sm">
                  <span className="text-muted-foreground">Entity:</span>
                  <span className="font-medium">
                    {getActivityEntityLabel(log.entity_type)}
                  </span>
                </div>
                {log.entity_id ? (
                  <div className="flex justify-between text-sm">
                    <span className="text-muted-foreground">ID:</span>
                    <span className="font-mono text-xs">
                      {log.entity_id.slice(0, 8)}...
                    </span>
                  </div>
                ) : null}
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      <div className="flex items-center justify-between pt-4">
        <p className="text-sm text-muted-foreground">Page {page}</p>
        <div className="flex gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => onPageChange(Math.max(1, page - 1))}
            disabled={page === 1}
          >
            <ChevronLeft className="h-4 w-4" />
            Previous
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => onPageChange(page + 1)}
            disabled={!hasMore}
          >
            Next
            <ChevronRight className="h-4 w-4" />
          </Button>
        </div>
      </div>
    </>
  );
}

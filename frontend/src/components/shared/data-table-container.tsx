"use client";

import { Card, CardContent } from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { cn } from "@/lib/utils";

interface DataTableContainerProps {
  limit: number;
  onLimitChange: (value: number) => void;
  total: number;
  currentCount: number;
  children: React.ReactNode;
  className?: string;
  hideLimitSelector?: boolean;
}

const pageSizeOptions = ["10", "20", "50", "100"] as const;

export function DataTableContainer({
  limit,
  onLimitChange,
  total,
  currentCount,
  children,
  className = "",
  hideLimitSelector = false,
}: DataTableContainerProps) {
  return (
    <Card
      className={cn(
        "flex flex-col border-0 bg-transparent shadow-none md:border md:bg-card md:shadow",
        className,
      )}
    >
      <CardContent className="flex-1 px-0 py-0 md:p-6">
        {!hideLimitSelector && (
          <div className="flex flex-col md:flex-row gap-4 items-center justify-between mb-4">
            <div className="flex items-center gap-2">
              <span className="text-sm text-muted-foreground">Show</span>
              <Select
                value={String(limit)}
                onValueChange={(value) => {
                  onLimitChange(Number(value));
                }}
              >
                <SelectTrigger className="w-[80px]">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {pageSizeOptions.map((option) => (
                    <SelectItem key={option} value={option}>
                      {option}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <span className="text-sm text-muted-foreground">entries</span>
            </div>
            {total > 0 && (
              <span className="text-sm text-muted-foreground">
                {Math.min(currentCount, limit)} entries of {total}
              </span>
            )}
          </div>
        )}
        {children}
      </CardContent>
    </Card>
  );
}

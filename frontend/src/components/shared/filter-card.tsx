"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { cn } from "@/lib/utils";

interface FilterCardProps {
  title?: string;
  action?: React.ReactNode;
  children: React.ReactNode;
  className?: string;
}

export function FilterCard({
  title,
  action,
  children,
  className = "",
}: FilterCardProps) {
  const hasHeader = !!title || !!action;
  return (
    <Card className={cn("mb-6", className)}>
      {hasHeader && (
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-4">
          {title && <CardTitle className="text-base font-semibold">{title}</CardTitle>}
          {action}
        </CardHeader>
      )}
      <CardContent className={cn(!hasHeader && "pt-6")}>{children}</CardContent>
    </Card>
  );
}

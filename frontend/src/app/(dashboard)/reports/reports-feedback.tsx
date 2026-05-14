'use client';

import type { ComponentType } from 'react';
import { Card, CardContent } from '@/components/ui/card';
import { Loader2 } from 'lucide-react';

interface ReportsLoadingStateProps {
  className?: string;
}

export function ReportsLoadingState({ className = 'h-64' }: ReportsLoadingStateProps) {
  return (
    <div className={`flex items-center justify-center ${className}`}>
      <Loader2 className="h-8 w-8 animate-spin text-primary" />
    </div>
  );
}

interface ReportsEmptyStateProps {
  icon: ComponentType<{ className?: string }>;
  message: string;
  details?: string;
}

export function ReportsEmptyState({ icon: Icon, message, details }: ReportsEmptyStateProps) {
  return (
    <Card>
      <CardContent className="flex flex-col items-center justify-center py-12">
        <Icon className="h-12 w-12 text-muted-foreground mb-4" />
        <p className="text-muted-foreground">{message}</p>
        {details ? <p className="text-xs text-muted-foreground mt-1">{details}</p> : null}
      </CardContent>
    </Card>
  );
}

interface ReportsErrorBannerProps {
  message: string | null;
}

export function ReportsErrorBanner({ message }: ReportsErrorBannerProps) {
  if (!message) return null;

  return (
    <Card className="mb-6 border-red-300 bg-red-50">
      <CardContent className="p-4 text-sm text-red-700">{message}</CardContent>
    </Card>
  );
}

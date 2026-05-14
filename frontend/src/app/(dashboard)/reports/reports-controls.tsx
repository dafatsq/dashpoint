'use client';

import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { DateRangePicker } from '@/components/ui/date-range-picker';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Download, RefreshCw } from 'lucide-react';

import { DATE_PRESETS } from './reports-helpers';
import type { DatePresetKey, ReportDateRange } from './reports-types';

interface ReportsControlsProps {
  datePreset: DatePresetKey;
  onPresetChange: (preset: DatePresetKey) => void;
  dateRange: ReportDateRange;
  onDateRangeChange: (value: ReportDateRange) => void;
  onRefresh?: () => void;
  onExport?: () => void;
  refreshLabel?: string;
  exportLabel?: string;
  canExport?: boolean;
  isLoading?: boolean;
  showDateRange?: boolean;
}

export function ReportsControls({
  datePreset,
  onPresetChange,
  dateRange,
  onDateRangeChange,
  onRefresh,
  onExport,
  refreshLabel = 'Refresh',
  exportLabel = 'Export CSV',
  canExport = false,
  isLoading = false,
  showDateRange = true,
}: ReportsControlsProps) {
  return (
    <Card className="mb-6">
      <CardContent className="p-6">
        <div className="flex flex-col sm:flex-row items-stretch sm:items-center gap-4">
          {showDateRange && (
            <>
              <Select value={datePreset} onValueChange={(value) => onPresetChange(value as DatePresetKey)}>
                <SelectTrigger className="w-full sm:w-40">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {Object.entries(DATE_PRESETS).map(([key, { label }]) => (
                    <SelectItem key={key} value={key}>
                      {label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <DateRangePicker
                value={dateRange}
                onChange={onDateRangeChange}
                placeholder="Select date range"
                className="w-full sm:w-[280px]"
              />
            </>
          )}
          {(onRefresh || (canExport && onExport)) && (
            <div className="flex gap-2 w-full sm:w-auto">
              {onRefresh && (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={onRefresh}
                  disabled={isLoading}
                  className="flex-1 sm:flex-none"
                >
                  <RefreshCw className={`h-4 w-4 mr-2 ${isLoading ? 'animate-spin' : ''}`} />
                  {refreshLabel}
                </Button>
              )}
              {canExport && onExport && (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={onExport}
                  className="flex-1 sm:flex-none"
                >
                  <Download className="h-4 w-4 mr-2" />
                  {exportLabel}
                </Button>
              )}
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
}

'use client';

import { Button } from '@/components/ui/button';
import { DateRangePicker } from '@/components/ui/date-range-picker';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Download, RefreshCw } from 'lucide-react';
import { FilterCard } from '@/components/shared/filter-card';

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
    <FilterCard>
      <div className="flex flex-wrap items-center justify-center gap-4">
        {showDateRange && (
          <>
            <div className="w-full flex-1 min-w-[200px] sm:min-w-[220px]">
              <Select value={datePreset} onValueChange={(value) => onPresetChange(value as DatePresetKey)}>
                <SelectTrigger className="w-full">
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
            </div>
            <DateRangePicker
              value={dateRange}
              onChange={onDateRangeChange}
              placeholder="Select date range"
              className="w-full flex-1 min-w-[200px] sm:min-w-[220px]"
            />
          </>
        )}
        {onRefresh && (
          <Button
            variant="outline"
            onClick={onRefresh}
            disabled={isLoading}
            className="w-full flex-1 min-w-[200px] sm:min-w-[220px]"
          >
            <RefreshCw className={`h-4 w-4 mr-2 ${isLoading ? 'animate-spin' : ''}`} />
            {refreshLabel}
          </Button>
        )}
        {canExport && onExport && (
          <Button
            variant="outline"
            onClick={onExport}
            className="w-full flex-1 min-w-[200px] sm:min-w-[220px]"
          >
            <Download className="h-4 w-4 mr-2" />
            {exportLabel}
          </Button>
        )}
      </div>
    </FilterCard>);
}

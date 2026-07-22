"use client";

import { ArrowDownUp } from "lucide-react";

import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

export interface DataSortOption {
  value: string;
  label: string;
}

interface DataSortSelectProps {
  value: string;
  options: readonly DataSortOption[];
  onChange: (value: string) => void;
  className?: string;
}

export function DataSortSelect({
  value,
  options,
  onChange,
  className,
}: DataSortSelectProps) {
  return (
    <div className={`flex items-center gap-2 ${className || ""}`}>
      <Select value={value} onValueChange={onChange}>
        <SelectTrigger className="min-w-0 flex-1">
          <span className="flex min-w-0 items-center gap-2">
            <ArrowDownUp className="h-4 w-4 shrink-0 text-muted-foreground" />
            <span className="text-muted-foreground">Sort by:</span>
            <SelectValue />
          </span>
        </SelectTrigger>
        <SelectContent>
          {options.map((option) => (
            <SelectItem key={option.value} value={option.value}>
              {option.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </div>
  );
}

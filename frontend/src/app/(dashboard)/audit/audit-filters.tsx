'use client';

import { Search } from "lucide-react";

import { Card, CardContent } from "@/components/ui/card";
import { DateRangePicker } from "@/components/ui/date-range-picker";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

import type { ActivityDateRange } from "../activity-helpers";

interface AuditFiltersProps {
  searchQuery: string;
  selectedAction: string;
  selectedEntity: string;
  dateRange: ActivityDateRange;
  onSearchChange: (value: string) => void;
  onActionChange: (value: string) => void;
  onEntityChange: (value: string) => void;
  onDateRangeChange: (value: ActivityDateRange) => void;
}

export function AuditFilters({
  searchQuery,
  selectedAction,
  selectedEntity,
  dateRange,
  onSearchChange,
  onActionChange,
  onEntityChange,
  onDateRangeChange,
}: AuditFiltersProps) {
  return (
    <Card className="mb-6">
      <CardContent className="p-4">
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-4">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              placeholder="Search logs..."
              value={searchQuery}
              onChange={(event) => onSearchChange(event.target.value)}
              className="w-full pl-9"
            />
          </div>
          <DateRangePicker
            value={dateRange}
            onChange={onDateRangeChange}
            placeholder="Filter by date..."
            className="w-full"
          />
          <Select value={selectedAction} onValueChange={onActionChange}>
            <SelectTrigger className="w-full">
              <SelectValue placeholder="All Actions" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Actions</SelectItem>
              <SelectItem value="login">Login</SelectItem>
              <SelectItem value="login_failed">Login Failed</SelectItem>
              <SelectItem value="create">Create</SelectItem>
              <SelectItem value="update">Update</SelectItem>
              <SelectItem value="archive">Archive</SelectItem>
              <SelectItem value="restore">Restore</SelectItem>
              <SelectItem value="delete">Delete</SelectItem>
              <SelectItem value="void">Void</SelectItem>
            </SelectContent>
          </Select>
          <Select value={selectedEntity} onValueChange={onEntityChange}>
            <SelectTrigger className="w-full">
              <SelectValue placeholder="All Entities" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Entities</SelectItem>
              <SelectItem value="auth">Authentication</SelectItem>
              <SelectItem value="user">User</SelectItem>
              <SelectItem value="product">Product</SelectItem>
              <SelectItem value="category">Category</SelectItem>
              <SelectItem value="sale">Sale</SelectItem>
              <SelectItem value="shift">Shift</SelectItem>
              <SelectItem value="inventory">Inventory</SelectItem>
              <SelectItem value="expense">Expense</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </CardContent>
    </Card>
  );
}

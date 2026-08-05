import type { ReactNode } from "react";

import {
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  DataSortSelect,
  type DataSortOption,
} from "@/components/shared/data-sort-select";

interface DashboardSectionHeaderProps {
  icon: ReactNode;
  title: string;
  description: string;
  sort: string;
  sortOptions: readonly DataSortOption[];
  onSortChange: (value: string) => void;
}

export function DashboardSectionHeader({
  icon,
  title,
  description,
  sort,
  sortOptions,
  onSortChange,
}: DashboardSectionHeaderProps) {
  return (
    <CardHeader className="px-0 pt-0 pb-4 md:p-6">
      <div
        data-slot="dashboard-section-header-layout"
        className="flex flex-wrap items-start justify-between gap-x-4 gap-y-3"
      >
        <div className="min-w-0 flex-1 sm:min-w-64">
          <div className="flex items-center gap-2">
            {icon}
            <CardTitle>{title}</CardTitle>
          </div>
          <CardDescription className="mt-1.5">{description}</CardDescription>
        </div>
        <DataSortSelect
          value={sort}
          options={sortOptions}
          onChange={onSortChange}
          className="w-full shrink-0 sm:w-auto sm:min-w-[220px]"
        />
      </div>
    </CardHeader>
  );
}

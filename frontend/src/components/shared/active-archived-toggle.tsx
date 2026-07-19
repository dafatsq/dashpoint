"use client";

import { Archive, Check } from "lucide-react";
import { cn } from "@/lib/utils";

type ActiveArchivedValue = "active" | "archived";

interface ActiveArchivedToggleProps {
  value: ActiveArchivedValue;
  onChange: (value: ActiveArchivedValue) => void;
  className?: string;
}

const activeArchivedOptions: Array<{
  value: ActiveArchivedValue;
  label: string;
  icon?: typeof Archive;
}> = [
  { value: "active", label: "Active", icon: Check },
  { value: "archived", label: "Archived", icon: Archive },
];

export function ActiveArchivedToggle({
  value,
  onChange,
  className = "",
}: ActiveArchivedToggleProps) {
  return (
    <div className={cn("flex w-fit gap-1 rounded-lg bg-muted p-1", className)}>
      {activeArchivedOptions.map((option) => {
        const Icon = option.icon;
        const isSelected = value === option.value;

        return (
          <button
            key={option.value}
            onClick={() => onChange(option.value)}
            className={cn(
              "flex-1 flex items-center justify-center gap-2 rounded-md px-4 py-2 text-sm font-medium transition-colors",
              isSelected
                ? "bg-background text-foreground shadow-sm"
                : "text-muted-foreground hover:text-foreground",
            )}
          >
            {Icon ? <Icon className="h-4 w-4" /> : null}
            {option.label}
          </button>
        );
      })}
    </div>
  );
}

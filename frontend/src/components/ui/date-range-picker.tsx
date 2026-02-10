"use client"

import * as React from "react"
import { format } from "date-fns"
import { CalendarIcon, X } from "lucide-react"
import type { DateRange } from "react-day-picker"

import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import { Calendar } from "@/components/ui/calendar"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"

interface DateRangePickerProps {
  value?: { start: string; end: string }
  onChange?: (value: { start: string; end: string }) => void
  placeholder?: string
  className?: string
  disabled?: boolean
}

export function DateRangePicker({
  value,
  onChange,
  placeholder = "Select date range",
  className,
  disabled = false,
}: DateRangePickerProps) {
  const [open, setOpen] = React.useState(false)
  const [month, setMonth] = React.useState<Date>(
    value?.start ? new Date(value.start) : new Date()
  )

  // Convert string dates to Date objects for the calendar
  const dateRange: DateRange | undefined = React.useMemo(() => {
    if (!value?.start && !value?.end) return undefined
    return {
      from: value?.start ? new Date(value.start) : undefined,
      to: value?.end ? new Date(value.end) : undefined,
    }
  }, [value])

  // Handle calendar selection
  const handleSelect = (range: DateRange | undefined) => {
    if (onChange) {
      onChange({
        start: range?.from ? format(range.from, "yyyy-MM-dd") : "",
        end: range?.to ? format(range.to, "yyyy-MM-dd") : "",
      })
    }
  }

  // Clear the selection
  const handleClear = (e: React.MouseEvent) => {
    e.stopPropagation()
    if (onChange) {
      onChange({ start: "", end: "" })
    }
  }

  // Format display text
  const displayText = React.useMemo(() => {
    if (!value?.start && !value?.end) return placeholder
    if (value?.start && value?.end) {
      return `${format(new Date(value.start), "MMM d, yyyy")} - ${format(new Date(value.end), "MMM d, yyyy")}`
    }
    if (value?.start) {
      return `${format(new Date(value.start), "MMM d, yyyy")} - Select end date`
    }
    return placeholder
  }, [value, placeholder])

  const hasValue = value?.start || value?.end

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          disabled={disabled}
          className={cn(
            "justify-start text-left font-normal",
            !hasValue && "text-muted-foreground",
            className
          )}
        >
          <CalendarIcon className="mr-2 h-4 w-4" />
          <span className="flex-1 truncate">{displayText}</span>
          {hasValue && (
            <X
              className="ml-2 h-4 w-4 opacity-50 hover:opacity-100"
              onClick={handleClear}
            />
          )}
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-auto p-0" align="start">
        <Calendar
          mode="range"
          month={month}
          onMonthChange={setMonth}
          selected={dateRange}
          onSelect={handleSelect}
          numberOfMonths={1}
          captionLayout="dropdown"
          startMonth={new Date(2020, 0)}
          endMonth={new Date(2030, 11)}
        />
        <div className="border-t p-3 flex justify-end gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => {
              handleSelect(undefined)
            }}
          >
            Clear
          </Button>
          <Button
            size="sm"
            onClick={() => setOpen(false)}
          >
            Done
          </Button>
        </div>
      </PopoverContent>
    </Popover>
  )
}

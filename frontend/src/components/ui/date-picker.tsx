"use client"

import * as React from "react"
import { format } from "date-fns"
import { CalendarIcon } from "lucide-react"

import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import { Calendar } from "@/components/ui/calendar"
import {
    Popover,
    PopoverContent,
    PopoverTrigger,
} from "@/components/ui/popover"

interface DatePickerProps {
    date?: string; // Expecting YYYY-MM-DD
    onSelect?: (date: string) => void;
    disabled?: boolean;
}

export function DatePicker({ date, onSelect, disabled }: DatePickerProps) {
    const [open, setOpen] = React.useState(false)

    // Convert string to Date
    const parsedDate = date ? new Date(date) : undefined

    // Handle selection from calendar
    const handleSelect = (selectedDate: Date | undefined) => {
        if (selectedDate && onSelect) {
            // Format to YYYY-MM-DD string to maintain local timezone
            const formatted = [
                selectedDate.getFullYear(),
                String(selectedDate.getMonth() + 1).padStart(2, '0'),
                String(selectedDate.getDate()).padStart(2, '0')
            ].join('-');
            onSelect(formatted);
        }
        setOpen(false)
    }

    return (
        <Popover open={open} onOpenChange={setOpen}>
            <PopoverTrigger asChild>
                <Button
                    variant={"outline"}
                    disabled={disabled}
                    className={cn(
                        "w-full justify-start text-left font-normal px-3 py-2",
                        !date && "text-muted-foreground"
                    )}
                >
                    <CalendarIcon className="mr-2 h-4 w-4" />
                    {date ? format(new Date(date), "PPP") : <span>Pick a date</span>}
                </Button>
            </PopoverTrigger>
            <PopoverContent className="w-auto p-0" align="start">
                <Calendar
                    mode="single"
                    selected={parsedDate}
                    onSelect={handleSelect}
                    initialFocus
                />
            </PopoverContent>
        </Popover>
    )
}

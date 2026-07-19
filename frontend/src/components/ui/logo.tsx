import { cn } from "@/lib/utils";

interface LogoProps {
  className?: string;
}

export function Logo({ className }: LogoProps) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox="51 52 153 153"
      className={cn("h-6 w-6 text-primary", className)}
      aria-hidden="true"
    >
      <path
        d="M 65,191 L 145,111"
        fill="none"
        stroke="currentColor"
        strokeWidth={28}
        strokeLinecap="round"
      />
      <path
        d="M 110,190 L 160,140"
        fill="none"
        stroke="currentColor"
        strokeWidth={18}
        strokeLinecap="round"
      />
      <circle cx="186" cy="70" r="18" fill="currentColor" />
    </svg>
  );
}

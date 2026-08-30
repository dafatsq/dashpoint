'use client';

import { Button } from "@/components/ui/button";

import { getDemoLoginCredentials } from "./login-demo-credentials";

interface LoginDemoAccessProps {
  onSelectCredentials: (email: string, password: string) => void;
}

export function LoginDemoAccess({
  onSelectCredentials,
}: LoginDemoAccessProps) {
  // Empty unless this build opted into demo mode via
  // NEXT_PUBLIC_ENABLE_QUICK_DEMO_ACCESS=true plus
  // NEXT_PUBLIC_DEMO_CREDENTIALS_JSON at build time.
  const demoCredentials = getDemoLoginCredentials();
  if (demoCredentials.length === 0) {
    return null;
  }

  return (
    <div className="mt-8 border-t pt-6">
      <p className="mb-4 text-center text-xs font-semibold uppercase tracking-wider text-muted-foreground">
        Quick Demo Access
      </p>
      <div className="grid grid-cols-1 gap-2">
        {demoCredentials.map((demo) => (
          <Button
            key={demo.role}
            variant="outline"
            size="sm"
            className="group flex h-auto flex-col items-start px-3 py-2 hover:border-primary/50 hover:bg-primary/5"
            onClick={() => onSelectCredentials(demo.email, demo.pass)}
            type="button"
          >
            <div className="flex w-full items-center justify-between">
              <span className="text-[10px] font-bold uppercase text-primary/70">
                {demo.role}
              </span>
              <span className="text-[10px] text-muted-foreground transition-colors group-hover:text-primary">
                Click to fill
              </span>
            </div>
            <div className="mt-0.5 text-xs font-mono">
              {demo.email}
              <span className="mx-1 text-muted-foreground">/</span>
              {demo.pass}
            </div>
          </Button>
        ))}
      </div>
    </div>
  );
}

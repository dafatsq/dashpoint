'use client';

import { Suspense } from "react";
import { Loader2 } from "lucide-react";

import { LoginScreen } from "./login-screen";

export default function LoginPage() {
  return (
    <Suspense
      fallback={
        <div className="min-h-screen flex items-center justify-center bg-background">
          <Loader2 className="h-8 w-8 animate-spin text-primary" />
        </div>
      }
    >
      <LoginScreen />
    </Suspense>
  );
}

"use client";

import { useEffect } from "react";
import { useRouter, usePathname } from "next/navigation";
import { useAuth } from "@/contexts/auth-context";
import { Loader2 } from "lucide-react";
import { hasRouteAccess } from "@/lib/nav-config";

interface RouteGuardProps {
  children: React.ReactNode;
}

export function RouteGuard({ children }: RouteGuardProps) {
  const router = useRouter();
  const pathname = usePathname();
  const { user, hasPermission, hasAnyPermission, isLoading, isAuthenticated } = useAuth();

  useEffect(() => {
    if (isLoading) return;
    if (!isAuthenticated || !user) return;

    if (
      !hasRouteAccess(pathname, {
        hasPermission,
        hasAnyPermission,
      })
    ) {
      router.replace("/?access_denied=true");
    }
  }, [pathname, user, hasPermission, hasAnyPermission, isLoading, isAuthenticated, router]);

  // Show loading while checking auth
  if (isLoading) {
    return (
      <div className="flex h-full items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  // If not authenticated, render nothing (will be redirected by layout)
  if (!isAuthenticated || !user) {
    return null;
  }

  if (
    !hasRouteAccess(pathname, {
      hasPermission,
      hasAnyPermission,
    })
  ) {
    return (
      <div className="flex h-full items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  return <>{children}</>;
}

"use client";

import { useEffect } from "react";
import { useRouter, usePathname } from "next/navigation";
import { useAuth, PERMISSIONS } from "@/contexts/auth-context";
import { Loader2 } from "lucide-react";

// Map routes to required permissions
const routePermissions: Record<string, string | string[] | undefined> = {
  "/": undefined, // Everyone can access dashboard
  "/pos": PERMISSIONS.POS_VIEW,
  "/products": PERMISSIONS.PRODUCTS_VIEW,
  "/inventory": PERMISSIONS.INVENTORY_VIEW,
  "/sales": PERMISSIONS.SALES_VIEW,
  "/reports": PERMISSIONS.REPORTS_VIEW,
  "/expenses": PERMISSIONS.REPORTS_VIEW,
  "/users": PERMISSIONS.USERS_VIEW,
  "/audit": PERMISSIONS.AUDIT_VIEW,
  "/shifts": [PERMISSIONS.POS_VIEW, PERMISSIONS.SALES_VIEW, PERMISSIONS.REPORTS_VIEW],
  "/categories": PERMISSIONS.CATEGORIES_VIEW,
  // '/changes' has no permission requirement so everyone can access it
  // '/settings' explicitly has no permission requirement so everyone can access it
};

interface RouteGuardProps {
  children: React.ReactNode;
}

export function RouteGuard({ children }: RouteGuardProps) {
  const router = useRouter();
  const pathname = usePathname();
  const { user, hasPermission, hasAnyPermission, isLoading, isAuthenticated } = useAuth();

  useEffect(() => {
    // Wait for auth to load
    if (isLoading) return;

    // Not authenticated - will be handled by dashboard layout
    if (!isAuthenticated || !user) return;

    // Find the matching route permission
    // Check exact match first, then check parent routes
    let requiredPermission: string | string[] | undefined;

    // Exact match
    if (routePermissions[pathname] !== undefined) {
      requiredPermission = routePermissions[pathname];
    } else {
      // Check if it's a sub-route (e.g., /users/123)
      const pathParts = pathname.split("/");
      for (let i = pathParts.length; i > 0; i--) {
        const parentPath = pathParts.slice(0, i).join("/") || "/";
        if (parentPath in routePermissions) {
          requiredPermission = routePermissions[parentPath];
          break;
        }
      }
    }

    // If route requires permission and user doesn't have it, redirect
    if (requiredPermission) {
      const hasAccess = Array.isArray(requiredPermission)
        ? hasAnyPermission(requiredPermission)
        : hasPermission(requiredPermission);
        
      if (!hasAccess) {
        console.log(
          `[RouteGuard] Access denied to ${pathname} - missing permission: ${requiredPermission}`,
        );
        router.replace("/?access_denied=true");
      }
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

  // Check permission for current route
  let requiredPermission: string | string[] | undefined;
  if (routePermissions[pathname] !== undefined) {
    requiredPermission = routePermissions[pathname];
  } else {
    const pathParts = pathname.split("/");
    for (let i = pathParts.length; i > 0; i--) {
      const parentPath = pathParts.slice(0, i).join("/") || "/";
      if (parentPath in routePermissions) {
        requiredPermission = routePermissions[parentPath];
        break;
      }
    }
  }

  // If permission required but not granted, show nothing while redirecting
  if (requiredPermission) {
    const hasAccess = Array.isArray(requiredPermission)
      ? hasAnyPermission(requiredPermission)
      : hasPermission(requiredPermission);
      
    if (!hasAccess) {
      return (
        <div className="flex h-full items-center justify-center">
          <Loader2 className="h-8 w-8 animate-spin text-primary" />
        </div>
      );
    }
  }

  return <>{children}</>;
}

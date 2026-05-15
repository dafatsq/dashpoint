"use client";

import { useCallback, useEffect, useState } from "react";
import { WifiOff } from "lucide-react";

import { BACKEND_BASE_URL } from "@/lib/config";

export function OfflineBanner() {
  const [isUnavailable, setIsUnavailable] = useState(false);

  const checkHealth = useCallback(async () => {
    try {
      const response = await fetch(`${BACKEND_BASE_URL}/api/v1/health`, {
        cache: "no-store",
      });
      const data = (await response.json()) as {
        status?: string;
        services?: {
          database?: string;
        };
      };

      const backendOk = response.ok && data.status === "ok";
      const databaseOk = data.services?.database === "connected";
      setIsUnavailable(!(backendOk && databaseOk));
    } catch {
      setIsUnavailable(true);
    }
  }, []);

  useEffect(() => {
    const initialCheckId = window.setTimeout(() => {
      void checkHealth();
    }, 0);

    const handleConnectivityEvent = () => {
      void checkHealth();
    };
    const intervalId = window.setInterval(() => {
      void checkHealth();
    }, 30000);

    window.addEventListener("online", handleConnectivityEvent);
    window.addEventListener("offline", handleConnectivityEvent);

    return () => {
      window.clearTimeout(initialCheckId);
      window.clearInterval(intervalId);
      window.removeEventListener("online", handleConnectivityEvent);
      window.removeEventListener("offline", handleConnectivityEvent);
    };
  }, [checkHealth]);

  if (!isUnavailable) return null;

  return (
    <div className="absolute top-0 left-0 right-0 z-[100] bg-red-500 text-white px-4 py-2.5 flex items-center justify-center gap-2 text-sm font-medium shadow-md transition-all animate-in slide-in-from-top">
      <WifiOff className="h-4 w-4" />
      <span>⚠️ Connection to backend or database lost. Unsaved changes may be lost or fail to execute.</span>
    </div>
  );
}

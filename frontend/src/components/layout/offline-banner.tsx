"use client";

import { useEffect, useState } from "react";
import { WifiOff } from "lucide-react";

export function OfflineBanner() {
  const [isOffline, setIsOffline] = useState(false);

  useEffect(() => {
    // Set initial state
    if (typeof navigator !== "undefined") {
      setIsOffline(!navigator.onLine);
    }

    const handleOnline = () => setIsOffline(false);
    const handleOffline = () => setIsOffline(true);

    window.addEventListener("online", handleOnline);
    window.addEventListener("offline", handleOffline);

    return () => {
      window.removeEventListener("online", handleOnline);
      window.removeEventListener("offline", handleOffline);
    };
  }, []);

  if (!isOffline) return null;

  return (
    <div className="absolute top-0 left-0 right-0 z-[100] bg-red-500 text-white px-4 py-2.5 flex items-center justify-center gap-2 text-sm font-medium shadow-md transition-all animate-in slide-in-from-top">
      <WifiOff className="h-4 w-4" />
      <span>⚠️ No internet connection. Unsaved changes may be lost or fail to execute.</span>
    </div>
  );
}

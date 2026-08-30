import { dirname } from "node:path";
import { fileURLToPath } from "node:url";

import type { NextConfig } from "next";

const frontendRoot = dirname(fileURLToPath(import.meta.url));
const isDesktopBuild = process.env.NEXT_PUBLIC_DESKTOP_BUILD === "true";

const nextConfig: NextConfig = {
  poweredByHeader: false,
  turbopack: {
    root: frontendRoot,
  },
  ...(isDesktopBuild
    ? {
        output: "export" as const,
        distDir: "dist",
        trailingSlash: true,
        images: {
          unoptimized: true,
        },
      }
    : {}),
};

export default nextConfig;

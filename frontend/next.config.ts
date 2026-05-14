import { dirname } from "node:path";
import { fileURLToPath } from "node:url";

import type { NextConfig } from "next";

const frontendRoot = dirname(fileURLToPath(import.meta.url));

const nextConfig: NextConfig = {
  turbopack: {
    root: frontendRoot,
  },
};

export default nextConfig;

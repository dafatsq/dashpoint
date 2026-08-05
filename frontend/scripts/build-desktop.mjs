import { spawnSync } from "node:child_process";

import { resolveDesktopBuildEnvironment } from "./desktop-build-config.mjs";
import { createDesktopBuildCommand } from "./desktop-build-process.mjs";

const { command: npmCommand, args, shell } = createDesktopBuildCommand();
const result = spawnSync(npmCommand, args, {
	env: {
		...resolveDesktopBuildEnvironment(process.env),
	},
	shell,
	stdio: "inherit",
});

if (result.error) {
  console.error(`Failed to start ${npmCommand}: ${result.error.message}`);
  process.exit(1);
}

process.exit(result.status ?? 1);

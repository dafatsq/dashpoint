export function createDesktopBuildCommand(platform = process.platform) {
	const isWindows = platform === "win32";

	return {
		command: isWindows ? "npm.cmd" : "npm",
		args: ["run", "build:desktop:next"],
		shell: isWindows,
	};
}

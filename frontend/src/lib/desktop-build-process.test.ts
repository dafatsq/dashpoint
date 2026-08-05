import { describe, expect, it } from "vitest";

import { createDesktopBuildCommand } from "../../scripts/desktop-build-process.mjs";

describe("createDesktopBuildCommand", () => {
	it("runs npm.cmd through a Windows shell", () => {
		expect(createDesktopBuildCommand("win32")).toEqual({
			command: "npm.cmd",
			args: ["run", "build:desktop:next"],
			shell: true,
		});
	});

	it("runs npm directly on Unix platforms", () => {
		expect(createDesktopBuildCommand("darwin")).toEqual({
			command: "npm",
			args: ["run", "build:desktop:next"],
			shell: false,
		});
	});
});

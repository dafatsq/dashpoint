// @vitest-environment jsdom

import { act, createElement } from "react";
import { createRoot, type Root } from "react-dom/client";
import { afterEach, beforeEach, describe, expect, test } from "vitest";

import { DashboardSectionHeader } from "./dashboard-section-header";

describe("DashboardSectionHeader", () => {
  let container: HTMLDivElement;
  let root: Root;

  beforeEach(() => {
    globalThis.IS_REACT_ACT_ENVIRONMENT = true;
    container = document.createElement("div");
    document.body.appendChild(container);
    root = createRoot(container);
  });

  afterEach(() => {
    act(() => {
      root.unmount();
    });
    container.remove();
  });

  test("wraps the sort selector below the title content when the header is narrow", async () => {
    await act(async () => {
      root.render(
        createElement(DashboardSectionHeader, {
          icon: createElement("span", null, "icon"),
          title: "Shift History",
          description: "Recent and active shared store shifts",
          sort: "date_desc",
          sortOptions: [{ value: "date_desc", label: "Date (newest)" }],
          onSortChange: () => undefined,
        }),
      );
    });

    const layout = container.querySelector(
      '[data-slot="dashboard-section-header-layout"]',
    );
    const sortControl = container.querySelector("button")?.parentElement;

    expect(layout?.className).toContain("flex-wrap");
    expect(layout?.querySelector("div")?.className).toContain("sm:min-w-64");
    expect(sortControl?.className).toContain("w-full");
    expect(sortControl?.className).toContain("sm:w-auto");
  });
});

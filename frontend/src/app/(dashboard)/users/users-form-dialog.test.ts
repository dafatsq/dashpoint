// @vitest-environment jsdom

import { act, createElement } from "react";
import { createRoot, type Root } from "react-dom/client";
import {
  afterEach,
  beforeEach,
  describe,
  expect,
  test,
} from "vitest";

import { UsersFormDialog } from "./users-form-dialog";

const selfUser = {
  role_id: "role-owner",
  id: "user-self",
  name: "Self User",
  email: "self@dashpoint.local",
  role_name: "owner" as const,
  is_active: true,
  has_pin: true,
  permissions: [],
  created_at: "",
  updated_at: "2026-06-05T10:00:00Z",
};

const otherUser = {
  ...selfUser,
  id: "user-other",
  name: "Other User",
};

function buildProps(editingUser: typeof selfUser | null) {
  return {
    open: true,
    onOpenChange: () => {},
    editingUser,
    currentUser: selfUser,
    availableRoles: ["cashier" as const],
    formData: {
      email: "",
      password: "",
      name: editingUser?.name ?? "",
      role: undefined,
      pin: "",
    },
    setFormData: () => {},
    formErrors: {},
    isSubmitting: false,
    hasChanges: true,
    onSubmit: () => {},
  };
}

describe("UsersFormDialog current-password proof field", () => {
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

  test("asks for the current password when editing your own record", async () => {
    const { UsersFormDialog: DialogComponent } = await import("./users-form-dialog");

    await act(async () => {
      root.render(createElement(DialogComponent, buildProps(selfUser)));
    });

    expect(document.querySelector("#current_password")).not.toBeNull();
  });

  test("does not ask for the current password when editing someone else", async () => {
    const { UsersFormDialog: DialogComponent } = await import("./users-form-dialog");

    await act(async () => {
      root.render(createElement(DialogComponent, buildProps(otherUser)));
    });

    expect(document.querySelector("#current_password")).toBeNull();
  });
});

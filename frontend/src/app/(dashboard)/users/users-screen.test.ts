// @vitest-environment jsdom

import { act, createElement } from "react";
import { createRoot, type Root } from "react-dom/client";
import {
  afterEach,
  beforeEach,
  describe,
  expect,
  test,
  vi,
} from "vitest";

const usersPageMock = vi.fn();
const rolesMock = vi.fn();
const permanentDeleteMock = vi.fn();
const showErrorMock = vi.fn();

const archivedUser = {
  id: "user-archived",
  name: "Archived User",
  email: "archived@dashpoint.local",
  role_name: "cashier",
  is_active: false,
  has_pin: true,
  permissions: [],
  created_at: "",
  updated_at: "",
};

vi.mock("@/contexts/auth-context", () => ({
  PERMISSIONS: {
    USERS_EDIT: "can_edit_user",
    USERS_DELETE: "can_delete_user",
    USERS_PERMISSIONS: "can_manage_permissions",
  },
  useAuth: () => ({
    user: {
      id: "owner-1",
      role_name: "owner",
    },
    hasPermission: () => true,
  }),
}));

vi.mock("@/contexts/error-context", () => ({
  useGlobalError: () => ({
    showError: showErrorMock,
  }),
}));

vi.mock("@/lib/api", () => ({
  default: {
    getUsersPage: usersPageMock,
    getRoles: rolesMock,
    deleteUser: vi.fn(),
    permanentDeleteUser: permanentDeleteMock,
    updateUser: vi.fn(),
    createUser: vi.fn(),
    getPermissions: vi.fn(),
    getUserPermissions: vi.fn(),
    setUserPermissions: vi.fn(),
  },
}));

vi.mock("@/components/layout/header", () => ({
  Header: ({ title }: { title?: string }) => createElement("div", {}, title),
}));

vi.mock("./users-list", () => ({
  UsersList: ({
    onPermanentDelete,
  }: {
    onPermanentDelete: (user: typeof archivedUser) => void;
  }) =>
    createElement(
      "button",
      {
        type: "button",
        onClick: () => onPermanentDelete(archivedUser),
      },
      "Open Permanent Delete",
    ),
}));

vi.mock("@/components/shared/confirmation-dialog", () => ({
  ConfirmationDialog: ({
    open,
    onConfirm,
    title,
  }: {
    open: boolean;
    onConfirm: () => void;
    title: string;
  }) =>
    open
      ? createElement(
          "button",
          {
            type: "button",
            onClick: onConfirm,
          },
          title === "Permanently Delete User" ? "Confirm Permanent Delete" : "Confirm Archive"
        )
      : null,
}));

vi.mock("./users-form-dialog", () => ({
  UsersFormDialog: () => null,
}));

vi.mock("./users-permissions-dialog", () => ({
  UsersPermissionsDialog: () => null,
}));

describe("UsersScreen", () => {
  let container: HTMLDivElement;
  let root: Root;

  beforeEach(() => {
    globalThis.IS_REACT_ACT_ENVIRONMENT = true;
    container = document.createElement("div");
    document.body.appendChild(container);
    root = createRoot(container);
    usersPageMock.mockResolvedValue({
      data: [],
      total: 0,
      page: 1,
      per_page: 20,
      total_pages: 0,
    });
    rolesMock.mockResolvedValue({ data: [] });
    permanentDeleteMock.mockResolvedValue({ data: { message: "ok" } });
    showErrorMock.mockReset();
  });

  afterEach(() => {
    act(() => {
      root.unmount();
    });
    container.remove();
    vi.clearAllMocks();
  });

  test("does not show a delete error after a successful permanent delete", async () => {
    const { default: UsersScreen } = await import("./users-screen");

    await act(async () => {
      root.render(createElement(UsersScreen));
      await Promise.resolve();
    });

    const openButton = Array.from(container.querySelectorAll("button")).find(
      (button) => button.textContent === "Open Permanent Delete",
    );

    await act(async () => {
      openButton?.dispatchEvent(
        new MouseEvent("click", { bubbles: true, cancelable: true }),
      );
      await Promise.resolve();
    });

    const confirmButton = Array.from(container.querySelectorAll("button")).find(
      (button) => button.textContent === "Confirm Permanent Delete",
    );

    await act(async () => {
      confirmButton?.dispatchEvent(
        new MouseEvent("click", { bubbles: true, cancelable: true }),
      );
      await Promise.resolve();
    });

    expect(permanentDeleteMock).toHaveBeenCalledWith("user-archived");
    expect(showErrorMock).not.toHaveBeenCalledWith(
      "Delete Error",
      "Failed to delete user. Please try again.",
    );
  });
});

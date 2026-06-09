// @vitest-environment jsdom

import { act, createContext, createElement, useContext, type ReactNode } from "react";
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
const updateRolePermissionsMock = vi.fn();
const permanentDeleteMock = vi.fn();
const showErrorMock = vi.fn();
const managerRole = {
  id: "role-manager",
  name: "manager",
  description: "Manager",
  permissions: ["access_users_page"],
};

const archivedUser = {
  id: "user-archived",
  name: "Archived User",
  email: "archived@dashpoint.local",
  role_name: "cashier",
  is_active: false,
  has_pin: true,
  permissions: [],
  created_at: "",
  updated_at: "2026-06-05T10:00:00Z",
};

vi.mock("@/contexts/auth-context", () => ({
  PERMISSIONS: {
    USERS_EDIT: "manage_users_page",
    USERS_DELETE: "manage_users_page",
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
    updateRolePermissions: updateRolePermissionsMock,
  },
}));

vi.mock("@/components/layout/header", () => ({
  Header: ({ title }: { title?: string }) => createElement("div", {}, title),
}));

const TabsContext = createContext<(value: string) => void>(() => undefined);

vi.mock("@/components/ui/tabs", () => ({
  Tabs: ({
    children,
    onValueChange,
  }: {
    children: ReactNode;
    onValueChange?: (value: string) => void;
  }) =>
    createElement(
      TabsContext.Provider,
      { value: onValueChange || (() => undefined) },
      children,
    ),
  TabsList: ({ children }: { children: ReactNode }) =>
    createElement("div", {}, children),
  TabsTrigger: ({
    children,
    value,
  }: {
    children: ReactNode;
    value: string;
  }) => {
    const onValueChange = useContext(TabsContext);
    return createElement(
      "button",
      {
        type: "button",
        onClick: () => onValueChange(value),
      },
      children,
    );
  },
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

vi.mock("./users-roles-list", () => ({
  UsersRolesList: ({
    roles,
    onEditPermissions,
  }: {
    roles: typeof managerRole[];
    onEditPermissions: (role: typeof managerRole) => void;
  }) =>
    createElement(
      "button",
      {
        type: "button",
        onClick: () => onEditPermissions(roles[0]),
      },
      "Open Role Permissions",
    ),
}));

vi.mock("./users-role-permissions-dialog", () => ({
  UsersRolePermissionsDialog: ({
    hasChanges,
    open,
    onPermissionsChange,
    onSubmit,
  }: {
    hasChanges: boolean;
    open: boolean;
    onPermissionsChange: (permissions: string[]) => void;
    onSubmit: () => void;
  }) =>
    open
      ? createElement(
          "div",
          {},
          createElement(
            "button",
            {
              type: "button",
              onClick: () =>
                onPermissionsChange([
                  "access_users_page",
                  "manage_users_page",
                ]),
            },
            "Enable Manage Users",
          ),
          createElement(
            "button",
            {
              disabled: !hasChanges,
              type: "button",
              onClick: onSubmit,
            },
            "Save Role Permissions",
          ),
        )
      : null,
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
    rolesMock.mockResolvedValue({ data: [managerRole] });
    updateRolePermissionsMock.mockResolvedValue({
      data: {
        ...managerRole,
        permissions: ["access_users_page", "manage_users_page"],
      },
    });
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

    expect(permanentDeleteMock).toHaveBeenCalledWith(
      "user-archived",
      "2026-06-05T10:00:00Z",
    );
    expect(showErrorMock).not.toHaveBeenCalledWith(
      "Delete Error",
      "Failed to delete user. Please try again.",
    );
  });

  test("keeps role permissions dialog open after successful save", async () => {
    const { default: UsersScreen } = await import("./users-screen");

    await act(async () => {
      root.render(createElement(UsersScreen));
      await Promise.resolve();
    });

    const rolesTab = Array.from(container.querySelectorAll("button")).find(
      (button) => button.textContent === "Roles",
    );

    await act(async () => {
      expect(rolesTab).toBeTruthy();
      rolesTab?.click();
      await Promise.resolve();
      await Promise.resolve();
    });

    const openButton = Array.from(container.querySelectorAll("button")).find(
      (button) => button.textContent === "Open Role Permissions",
    );

    await act(async () => {
      expect(openButton).toBeTruthy();
      openButton?.click();
      await Promise.resolve();
      await Promise.resolve();
    });

    const initialSaveButton = Array.from(container.querySelectorAll("button")).find(
      (button) => button.textContent === "Save Role Permissions",
    ) as HTMLButtonElement | undefined;

    expect(initialSaveButton?.disabled).toBe(true);

    updateRolePermissionsMock.mockClear();
    await act(async () => {
      initialSaveButton?.click();
      await Promise.resolve();
    });
    expect(updateRolePermissionsMock).not.toHaveBeenCalled();

    const changeButton = Array.from(container.querySelectorAll("button")).find(
      (button) => button.textContent === "Enable Manage Users",
    );

    await act(async () => {
      expect(changeButton).toBeTruthy();
      changeButton?.click();
      await Promise.resolve();
      await Promise.resolve();
    });

    const changedSaveButton = Array.from(container.querySelectorAll("button")).find(
      (button) => button.textContent === "Save Role Permissions",
    );

    await act(async () => {
      expect(changedSaveButton).toBeTruthy();
      changedSaveButton?.click();
      await Promise.resolve();
      await Promise.resolve();
    });

    expect(updateRolePermissionsMock).toHaveBeenCalledWith("role-manager", [
      "access_users_page",
      "manage_users_page",
    ], [
      "access_users_page",
    ]);
    expect(
      Array.from(container.querySelectorAll("button")).some(
        (button) => button.textContent === "Save Role Permissions",
      ),
    ).toBe(true);
  });
});

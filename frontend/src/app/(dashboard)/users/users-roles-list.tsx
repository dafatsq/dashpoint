"use client";

import { Pencil, Shield, ShieldAlert, ShieldCheck } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { DataSortSelect } from "@/components/shared/data-sort-select";
import type { Role } from "@/types";
import { useMemo, useState } from "react";

interface UsersRolesListProps {
  roles: Role[];
  isOwner: boolean;
  onEditPermissions: (role: Role) => void;
}

function getRoleIcon(roleName: string) {
  switch (roleName) {
    case "owner":
      return <ShieldAlert className="h-4 w-4" />;
    case "manager":
      return <ShieldCheck className="h-4 w-4" />;
    default:
      return <Shield className="h-4 w-4" />;
  }
}

export function UsersRolesList({ roles, isOwner, onEditPermissions }: UsersRolesListProps) {
  const [sort, setSort] = useState("priority");
  const sortedRoles = useMemo(
    () =>
      [...roles].sort((left, right) => {
        if (sort === "permissions_desc" || sort === "permissions_asc") {
          const direction = sort.endsWith("_desc") ? -1 : 1;
          return ((left.permissions?.length || 0) - (right.permissions?.length || 0)) * direction;
        }
        if (sort === "name_asc" || sort === "name_desc") {
          const direction = sort.endsWith("_desc") ? -1 : 1;
          return left.name.localeCompare(right.name) * direction;
        }
        const priority: Record<string, number> = { owner: 0, manager: 1, cashier: 2 };
        return (priority[left.name] ?? 3) - (priority[right.name] ?? 3);
      }),
    [roles, sort],
  );

  return (
    <Card>
      <CardHeader>
        <div className="flex flex-wrap items-center justify-between gap-4">
          <CardTitle>Roles</CardTitle>
          <DataSortSelect
            value={sort}
            options={[
              { value: "priority", label: "Role priority" },
              { value: "name_asc", label: "Name (A-Z)" },
              { value: "name_desc", label: "Name (Z-A)" },
              { value: "permissions_desc", label: "Permissions (high-low)" },
              { value: "permissions_asc", label: "Permissions (low-high)" },
            ]}
            onChange={setSort}
          />
        </div>
      </CardHeader>
      <CardContent>
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b text-left text-sm text-muted-foreground">
                <th className="pb-3 font-medium">Role</th>
                <th className="pb-3 font-medium">Description</th>
                <th className="pb-3 font-medium text-right">Permissions</th>
                {isOwner ? <th className="pb-3 font-medium text-right">Actions</th> : null}
              </tr>
            </thead>
            <tbody>
              {sortedRoles.map((role) => {
                const isOwnerRole = role.name === "owner";
                return (
                  <tr key={role.id} className="border-b last:border-0">
                    <td className="py-3">
                      <div className="flex items-center gap-2 font-medium capitalize">
                        {getRoleIcon(role.name)}
                        {role.name}
                      </div>
                    </td>
                    <td className="py-3 text-sm text-muted-foreground">
                      {role.description || "-"}
                    </td>
                    <td className="py-3 text-right text-sm text-muted-foreground">
                      {role.permissions?.length || 0}
                    </td>
                    {isOwner ? (
                      <td className="py-3 text-right">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => onEditPermissions(role)}
                          disabled={isOwnerRole}
                        >
                          <Pencil className="mr-2 h-4 w-4" />
                          Edit Permissions
                        </Button>
                      </td>
                    ) : null}
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      </CardContent>
    </Card>
  );
}

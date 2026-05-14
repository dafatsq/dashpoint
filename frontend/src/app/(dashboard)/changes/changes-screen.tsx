'use client';

import { useCallback, useEffect, useState, type ReactNode } from "react";
import { Boxes, Package, Receipt, Wallet } from "lucide-react";

import { Header } from "@/components/layout/header";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import api from "@/lib/api";
import type { User } from "@/types";

import type { ActivityChangeTab, ActivityDateRange } from "../activity-helpers";
import { buildChangesListResetKey } from "../activity-helpers";
import { ChangesFilters } from "./changes-filters";
import { ChangesList } from "./changes-list";

const CHANGE_TABS: { value: ActivityChangeTab; label: string; icon: ReactNode }[] = [
  { value: "product", label: "Products", icon: <Package className="h-4 w-4" /> },
  { value: "inventory", label: "Inventory", icon: <Boxes className="h-4 w-4" /> },
  { value: "sale", label: "Sales", icon: <Receipt className="h-4 w-4" /> },
  { value: "expense", label: "Expenses", icon: <Wallet className="h-4 w-4" /> },
];

export function ChangesScreen() {
  const [dateRange, setDateRange] = useState<ActivityDateRange>({ start: "", end: "" });
  const [selectedUser, setSelectedUser] = useState("all");
  const [activeTab, setActiveTab] = useState<ActivityChangeTab>("product");
  const [users, setUsers] = useState<User[]>([]);
  const [usersError, setUsersError] = useState<string | null>(null);

  const fetchUsers = useCallback(async () => {
    setUsersError(null);
    const result = await api.getUsers();
    if (result.error) {
      setUsersError(result.error);
      return;
    }
    if (result.data) {
      setUsers(result.data);
    }
  }, []);

  useEffect(() => {
    const timer = window.setTimeout(() => {
      void fetchUsers();
    }, 0);

    return () => window.clearTimeout(timer);
  }, [fetchUsers]);

  return (
    <div className="flex h-full flex-col">
      <Header title="Changes" />

      <div className="flex-1 overflow-auto p-6">
        <div className="mx-auto w-full">
          {usersError ? (
            <div className="mb-4 rounded-md border border-destructive/20 bg-destructive/10 p-3 text-sm text-destructive">
              <div className="flex items-center justify-between gap-3">
                <span>{usersError}</span>
                <Button variant="outline" size="sm" onClick={() => void fetchUsers()}>
                  Retry
                </Button>
              </div>
            </div>
          ) : null}

          <ChangesFilters
            dateRange={dateRange}
            selectedUser={selectedUser}
            users={users}
            onDateRangeChange={setDateRange}
            onSelectedUserChange={setSelectedUser}
          />

          <Card className="flex flex-col border-0 bg-transparent shadow-none md:border md:bg-card md:shadow">
            <CardContent className="flex-1 px-0 py-0 md:p-6">
              <Tabs value={activeTab} onValueChange={(value) => setActiveTab(value as ActivityChangeTab)} className="w-full">
                <TabsList className="mb-6 flex w-full justify-start overflow-x-auto pb-1 no-scrollbar lg:inline-flex lg:w-auto">
                  {CHANGE_TABS.map((tab) => (
                    <TabsTrigger key={tab.value} value={tab.value} className="gap-1.5 px-4">
                      {tab.icon}
                      {tab.label}
                    </TabsTrigger>
                  ))}
                </TabsList>

                {CHANGE_TABS.map((tab) => (
                  <TabsContent key={tab.value} value={tab.value}>
                    <ChangesList
                      key={buildChangesListResetKey({
                        entityType: tab.value,
                        dateRange,
                        selectedUser,
                      })}
                      entityType={tab.value}
                      dateRange={dateRange}
                      selectedUser={selectedUser}
                    />
                  </TabsContent>
                ))}
              </Tabs>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}

"use client";

import { useCallback, useEffect, useState, type ReactNode } from "react";
import { Boxes, Package, Receipt, Wallet } from "lucide-react";

import { Header } from "@/components/layout/header";
import { Card, CardContent } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import api from "@/lib/api";

import type { ActivityChangeTab, ActivityDateRange } from "../activity-helpers";
import { buildChangesListResetKey } from "../activity-helpers";
import { ChangesFilters } from "./changes-filters";
import { ChangesList } from "./changes-list";

const CHANGE_TABS: {
  value: ActivityChangeTab;
  label: string;
  icon: ReactNode;
}[] = [
  {
    value: "product",
    label: "Products",
    icon: <Package className="h-4 w-4" />,
  },
  {
    value: "inventory",
    label: "Inventory",
    icon: <Boxes className="h-4 w-4" />,
  },
  { value: "sale", label: "Sales", icon: <Receipt className="h-4 w-4" /> },
  { value: "expense", label: "Expenses", icon: <Wallet className="h-4 w-4" /> },
];

export function ChangesScreen() {
  const [dateRange, setDateRange] = useState<ActivityDateRange>({
    start: "",
    end: "",
  });
  const [selectedUser, setSelectedUser] = useState("all");
  const [activeTab, setActiveTab] = useState<ActivityChangeTab>("product");
  const [sort, setSort] = useState("date_desc");
  const [users, setUsers] = useState<{ id: string; name: string }[]>([]);

  const loadUsers = useCallback(async () => {
    const result = await api.getBasicUsers();
    if (!result.error && result.data) {
      setUsers([...result.data].sort((a, b) => a.name.localeCompare(b.name)));
    }
  }, []);

  // Fetch active employees for filter (no permission gate, active users only)
  useEffect(() => {
    const timer = window.setTimeout(() => {
      void loadUsers();
    }, 0);

    return () => window.clearTimeout(timer);
  }, [loadUsers]);

  return (
    <div className="flex h-full flex-col">
      <Header title="Changes" />

      <div className="flex-1 overflow-auto p-6">
        <div className="mx-auto w-full">
          <ChangesFilters
            dateRange={dateRange}
            selectedUser={selectedUser}
            users={users}
            sort={sort}
            onDateRangeChange={setDateRange}
            onSelectedUserChange={setSelectedUser}
            onSortChange={setSort}
          />

          <Card className="flex flex-col border-0 bg-transparent shadow-none md:border md:bg-card md:shadow">
            <CardContent className="flex-1 px-0 py-0 md:p-6">
              <Tabs
                value={activeTab}
                onValueChange={(value) =>
                  setActiveTab(value as ActivityChangeTab)
                }
                className="w-full"
              >
                <TabsList className="mb-6 flex w-full justify-start overflow-x-auto pb-1 no-scrollbar lg:inline-flex lg:w-auto">
                  {CHANGE_TABS.map((tab) => (
                    <TabsTrigger
                      key={tab.value}
                      value={tab.value}
                      className="gap-1.5 px-4"
                    >
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
                      }) + `|${sort}`}
                      entityType={tab.value}
                      dateRange={dateRange}
                      selectedUser={selectedUser}
                      sort={sort}
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

"use client";

import { useCallback, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { Banknote, Building2, CreditCard, QrCode } from "lucide-react";

import { Header } from "@/components/layout/header";
import { useAuth, PERMISSIONS } from "@/contexts/auth-context";
import { DataTableContainer } from "@/components/shared/data-table-container";
import { useGlobalError } from "@/contexts/error-context";
import api from "@/lib/api";
import type { PaymentMethod, Sale } from "@/types";

import { SalesDetailDialog } from "./sales-detail-dialog";
import { SalesFilters } from "./sales-filters";
import { SalesList } from "./sales-list";
import { SalesVoidDialog } from "./sales-void-dialog";

const PAYMENT_ICONS: Record<PaymentMethod, React.ReactNode> = {
  cash: <Banknote className="h-4 w-4" />,
  card: <CreditCard className="h-4 w-4" />,
  qris: <QrCode className="h-4 w-4" />,
  transfer: <Building2 className="h-4 w-4" />,
};

export function SalesScreen() {
  const router = useRouter();
  const { hasPermission } = useAuth();
  const [sales, setSales] = useState<Sale[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [pageError, setPageError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [employeeFilter, setEmployeeFilter] = useState<string>("all");
  const [dateRange, setDateRange] = useState({ start: "", end: "" });
  const [employees, setEmployees] = useState<
    { id: string; name: string; role_name?: string }[]
  >([]);
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(20);
  const [hasMore, setHasMore] = useState(true);
  const [total, setTotal] = useState(0);
  const [viewDialogOpen, setViewDialogOpen] = useState(false);
  const [selectedSale, setSelectedSale] = useState<Sale | null>(null);
  const [voidDialogOpen, setVoidDialogOpen] = useState(false);
  const [voidReason, setVoidReason] = useState("");
  const { showError } = useGlobalError();
  const [isVoiding, setIsVoiding] = useState(false);

  const resetToFirstPage = useCallback(() => {
    setPage(1);
  }, []);

  const loadEmployees = useCallback(async () => {
    const result = await api.getBasicUsers();
    if (!result.error && result.data) {
      setEmployees(
        [...result.data].sort((a, b) => a.name.localeCompare(b.name)),
      );
    }
  }, []);

  const loadSales = useCallback(async () => {
    setIsLoading(true);
    setPageError(null);

    const params: {
      from?: string;
      to?: string;
      status?: string;
      user_id?: string;
      invoice_no?: string;
      limit?: number;
      offset?: number;
    } = {
      limit,
      offset: (page - 1) * limit,
    };

    if (dateRange.start) params.from = dateRange.start;
    if (dateRange.end) params.to = dateRange.end;
    if (statusFilter !== "all") params.status = statusFilter;
    if (employeeFilter !== "all") params.user_id = employeeFilter;
    if (debouncedSearch.trim()) params.invoice_no = debouncedSearch.trim();

    const result = await api.getSalesPage(params);
    if (result.error) {
      setPageError(result.error);
    } else {
      const data = result.data || [];
      setSales(data);
      setTotal(result.total || 0);
      setHasMore(data.length === limit);
    }

    setIsLoading(false);
  }, [
    dateRange.end,
    dateRange.start,
    debouncedSearch,
    employeeFilter,
    limit,
    page,
    statusFilter,
  ]);

  // Debounce search input — reset to page 1 and wait 300ms before fetching
  useEffect(() => {
    const timer = window.setTimeout(() => {
      setDebouncedSearch(searchQuery);
      resetToFirstPage();
    }, 300);
    return () => window.clearTimeout(timer);
  }, [resetToFirstPage, searchQuery]);

  useEffect(() => {
    if (!hasPermission(PERMISSIONS.SALES_VIEW)) {
      router.replace("/");
    }
  }, [hasPermission, router]);

  // Fetch active employees for filter (no permission gate, active users only)
  useEffect(() => {
    const timer = window.setTimeout(() => {
      void loadEmployees();
    }, 0);

    return () => window.clearTimeout(timer);
  }, [loadEmployees]);

  useEffect(() => {
    const timer = window.setTimeout(() => {
      void loadSales();
    }, 0);

    return () => window.clearTimeout(timer);
  }, [loadSales]);

  const viewSaleDetails = async (sale: Sale) => {
    const result = await api.getSale(sale.id);
    if (result.error) {
      setPageError(result.error);
      return;
    }
    if (result.data) {
      setSelectedSale(result.data);
      setViewDialogOpen(true);
    }
  };

  const handleVoidSale = async () => {
    if (!selectedSale) return;
    if (!voidReason.trim()) {
      showError("Reason Required", "Enter a reason before voiding this sale.");
      return;
    }

    setIsVoiding(true);
    const result = await api.voidSale(selectedSale.id, voidReason);
    if (result.error) {
      showError("Void Failed", result.error);
      setIsVoiding(false);
      return;
    }

    setSales((prev) =>
      prev.map((sale) =>
        sale.id === selectedSale.id
          ? { ...sale, status: "voided" as const }
          : sale,
      ),
    );
    setSelectedSale((prev) =>
      prev ? { ...prev, status: "voided", void_reason: voidReason } : prev,
    );
    setVoidDialogOpen(false);
    setViewDialogOpen(false);
    setVoidReason("");
    setIsVoiding(false);
  };

  return (
    <div className="flex flex-col h-full">
      <Header title="Sales History" />

      <div className="flex-1 p-6 overflow-auto">
        <SalesFilters
          searchQuery={searchQuery}
          statusFilter={statusFilter}
          employeeFilter={employeeFilter}
          dateRange={dateRange}
          employees={employees}
          onSearchChange={(value) => {
            setSearchQuery(value);
            resetToFirstPage();
          }}
          onStatusChange={(value) => {
            setStatusFilter(value);
            resetToFirstPage();
          }}
          onEmployeeChange={(value) => {
            setEmployeeFilter(value);
            resetToFirstPage();
          }}
          onDateRangeChange={(range) => {
            setDateRange(range);
            resetToFirstPage();
          }}
        />

        {pageError ? (
          <div className="mb-4 rounded-md border border-destructive/20 bg-destructive/10 p-3 text-sm text-destructive">
            {pageError}
          </div>
        ) : null}

        <DataTableContainer
          limit={limit}
          onLimitChange={(value) => {
            setLimit(value);
            setPage(1);
          }}
          total={total}
          currentCount={sales.length}
        >
          <SalesList
            sales={sales}
            isLoading={isLoading}
            page={page}
            hasMore={hasMore}
            paymentIcons={PAYMENT_ICONS}
            onPageChange={setPage}
            onViewSale={(sale) => void viewSaleDetails(sale)}
          />
        </DataTableContainer>
      </div>

      <SalesDetailDialog
        open={viewDialogOpen}
        sale={selectedSale}
        paymentIcons={PAYMENT_ICONS}
        onOpenChange={setViewDialogOpen}
        onVoidRequest={() => {
          setVoidDialogOpen(true);
        }}
      />

      <SalesVoidDialog
        open={voidDialogOpen}
        reason={voidReason}
        isVoiding={isVoiding}
        onOpenChange={setVoidDialogOpen}
        onReasonChange={setVoidReason}
        onConfirm={() => void handleVoidSale()}
      />
    </div>
  );
}

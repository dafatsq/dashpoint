'use client';

import { useEffect, useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import { Banknote, Building2, CreditCard, QrCode } from "lucide-react";

import { Header } from "@/components/layout/header";
import { useAuth, PERMISSIONS } from "@/contexts/auth-context";
import api from "@/lib/api";
import type { PaymentMethod, Sale, User } from "@/types";

import { SalesDetailDialog } from "./sales-detail-dialog";
import { SalesFilters } from "./sales-filters";
import { filterSalesBySearch } from "./sales-helpers";
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
  const [employeeError, setEmployeeError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [employeeFilter, setEmployeeFilter] = useState<string>("all");
  const [dateRange, setDateRange] = useState({ start: "", end: "" });
  const [employees, setEmployees] = useState<User[]>([]);
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(20);
  const [hasMore, setHasMore] = useState(true);
  const [total, setTotal] = useState(0);
  const [viewDialogOpen, setViewDialogOpen] = useState(false);
  const [selectedSale, setSelectedSale] = useState<Sale | null>(null);
  const [voidDialogOpen, setVoidDialogOpen] = useState(false);
  const [voidReason, setVoidReason] = useState("");
  const [voidError, setVoidError] = useState<string | null>(null);
  const [isVoiding, setIsVoiding] = useState(false);

  useEffect(() => {
    if (!hasPermission(PERMISSIONS.SALES_VIEW)) {
      router.replace("/");
    }
  }, [hasPermission, router]);

  useEffect(() => {
    const fetchEmployees = async () => {
      setEmployeeError(null);
      const result = await api.getUsers({ active: true });
      if (result.error) {
        setEmployeeError(result.error);
        return;
      }
      if (result.data) {
        setEmployees([...result.data].sort((a, b) => a.name.localeCompare(b.name)));
      }
    };
    void fetchEmployees();
  }, []);

  useEffect(() => {
    const fetchSales = async () => {
      setIsLoading(true);
      setPageError(null);

      const params: {
        from?: string;
        to?: string;
        status?: string;
        user_id?: string;
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

      const result = await api.getSalesPage(params);
      if (result.error) {
        setPageError(result.error);
      } else {
        setSales(result.data || []);
        setTotal(result.total || 0);
        setHasMore((result.data || []).length === limit);
      }

      setIsLoading(false);
    };

    void fetchSales();
  }, [dateRange, employeeFilter, limit, page, statusFilter]);

  const filteredSales = useMemo(() => filterSalesBySearch(sales, searchQuery), [sales, searchQuery]);

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
    if (!selectedSale || !voidReason) return;

    setIsVoiding(true);
    setVoidError(null);
    const result = await api.voidSale(selectedSale.id, voidReason);
    if (result.error) {
      setVoidError(result.error);
      setIsVoiding(false);
      return;
    }

    setSales((prev) => prev.map((sale) => (sale.id === selectedSale.id ? { ...sale, status: "voided" as const } : sale)));
    setSelectedSale((prev) => (prev ? { ...prev, status: "voided", void_reason: voidReason } : prev));
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
            setPage(1);
          }}
          onStatusChange={(value) => {
            setStatusFilter(value);
            setPage(1);
          }}
          onEmployeeChange={(value) => {
            setEmployeeFilter(value);
            setPage(1);
          }}
          onDateRangeChange={(range) => {
            setDateRange(range);
            setPage(1);
          }}
        />

        {pageError ? <div className="mb-4 rounded-md border border-destructive/20 bg-destructive/10 p-3 text-sm text-destructive">{pageError}</div> : null}
        {employeeError ? <div className="mb-4 rounded-md border border-destructive/20 bg-destructive/10 p-3 text-sm text-destructive">{employeeError}</div> : null}

        <SalesList
          sales={filteredSales}
          isLoading={isLoading}
          page={page}
          limit={limit}
          total={total}
          hasMore={hasMore}
          paymentIcons={PAYMENT_ICONS}
          onPageChange={setPage}
          onLimitChange={(value) => {
            setLimit(value);
            setPage(1);
          }}
          onViewSale={(sale) => void viewSaleDetails(sale)}
        />
      </div>

      <SalesDetailDialog
        open={viewDialogOpen}
        sale={selectedSale}
        paymentIcons={PAYMENT_ICONS}
        onOpenChange={setViewDialogOpen}
        onVoidRequest={() => {
          setVoidError(null);
          setVoidDialogOpen(true);
        }}
      />

      <SalesVoidDialog
        open={voidDialogOpen}
        reason={voidReason}
        error={voidError}
        isVoiding={isVoiding}
        onOpenChange={setVoidDialogOpen}
        onReasonChange={setVoidReason}
        onConfirm={() => void handleVoidSale()}
      />
    </div>
  );
}

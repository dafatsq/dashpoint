'use client';

import { useEffect, useState } from 'react';
import { Header } from '@/components/layout/header';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Search,
  Loader2,
  Receipt,
  Eye,
  XCircle,
  Calendar,
  CreditCard,
  Banknote,
  QrCode,
  Building2,
} from 'lucide-react';
import api from '@/lib/api';
import { Sale, PaymentMethod } from '@/types';
import { useAuth, PERMISSIONS } from '@/contexts/auth-context';

const PAYMENT_ICONS: Record<PaymentMethod, React.ReactNode> = {
  cash: <Banknote className="h-4 w-4" />,
  card: <CreditCard className="h-4 w-4" />,
  qris: <QrCode className="h-4 w-4" />,
  transfer: <Building2 className="h-4 w-4" />,
};

export default function SalesHistoryPage() {
  const { hasPermission } = useAuth();
  const [sales, setSales] = useState<Sale[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('all');
  const [dateFrom, setDateFrom] = useState('');
  const [dateTo, setDateTo] = useState('');

  // View dialog
  const [viewDialogOpen, setViewDialogOpen] = useState(false);
  const [selectedSale, setSelectedSale] = useState<Sale | null>(null);

  // Void dialog
  const [voidDialogOpen, setVoidDialogOpen] = useState(false);
  const [voidReason, setVoidReason] = useState('');
  const [isVoiding, setIsVoiding] = useState(false);

  // Fetch sales
  useEffect(() => {
    const fetchSales = async () => {
      setIsLoading(true);
      try {
        const params: { from?: string; to?: string; status?: string } = {};
        if (dateFrom) params.from = dateFrom;
        if (dateTo) params.to = dateTo;
        if (statusFilter !== 'all') params.status = statusFilter;

        const result = await api.getSales(params);
        if (result.data) setSales(result.data);
      } catch (error) {
        console.error('Failed to fetch sales:', error);
      } finally {
        setIsLoading(false);
      }
    };

    fetchSales();
  }, [dateFrom, dateTo, statusFilter]);

  // Filter sales
  const filteredSales = sales.filter((sale) =>
    sale.invoice_no?.toLowerCase().includes(searchQuery.toLowerCase()) ||
    sale.employee_name?.toLowerCase().includes(searchQuery.toLowerCase())
  );

  // View sale details
  const viewSaleDetails = async (sale: Sale) => {
    try {
      const result = await api.getSale(sale.id);
      if (result.data) {
        setSelectedSale(result.data);
        setViewDialogOpen(true);
      }
    } catch (error) {
      console.error('Failed to fetch sale details:', error);
    }
  };

  // Void sale
  const handleVoidSale = async () => {
    if (!selectedSale || !voidReason) return;

    setIsVoiding(true);
    try {
      const result = await api.voidSale(selectedSale.id, voidReason);
      if (!result.error) {
        setSales((prev) =>
          prev.map((s) =>
            s.id === selectedSale.id ? { ...s, status: 'voided' as const } : s
          )
        );
        setVoidDialogOpen(false);
        setViewDialogOpen(false);
        setVoidReason('');
      } else {
        alert(result.error);
      }
    } catch (error) {
      console.error('Failed to void sale:', error);
    } finally {
      setIsVoiding(false);
    }
  };

  const formatCurrency = (amount: string | number) => {
    const num = typeof amount === 'string' ? parseFloat(amount) : amount;
    return new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: 'IDR',
      minimumFractionDigits: 0,
    }).format(num || 0);
  };

  const formatDate = (date: string) => {
    return new Date(date).toLocaleString('id-ID', {
      day: '2-digit',
      month: 'short',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'completed':
        return 'bg-green-600 text-white dark:bg-green-600/90 dark:text-white';
      case 'voided':
        return 'bg-red-600 text-white dark:bg-red-600/90 dark:text-white';
      case 'pending':
        return 'bg-yellow-600 text-white dark:bg-yellow-600/90 dark:text-white';
      default:
        return 'bg-gray-600 text-white dark:bg-gray-600/90 dark:text-white';
    }
  };

  // Get primary payment method from sale
  const getPrimaryPaymentMethod = (sale: Sale): PaymentMethod => {
    if (sale.payments && sale.payments.length > 0) {
      return sale.payments[0].payment_method;
    }
    return 'cash';
  };

  return (
    <div className="flex flex-col h-full">
      <Header title="Sales History" />

      <div className="flex-1 p-6 overflow-auto">
        {/* Filters */}
        <Card className="mb-6">
          <CardHeader>
            <CardTitle className="text-base">Filters</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid gap-4 md:grid-cols-4">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                <Input
                  placeholder="Search invoice..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="pl-9"
                />
              </div>
              <div className="flex items-center gap-2">
                <Calendar className="h-4 w-4 text-muted-foreground" />
                <Input
                  type="date"
                  value={dateFrom}
                  onChange={(e) => setDateFrom(e.target.value)}
                  placeholder="From"
                />
              </div>
              <div className="flex items-center gap-2">
                <Calendar className="h-4 w-4 text-muted-foreground" />
                <Input
                  type="date"
                  value={dateTo}
                  onChange={(e) => setDateTo(e.target.value)}
                  placeholder="To"
                />
              </div>
              <Select value={statusFilter} onValueChange={setStatusFilter}>
                <SelectTrigger>
                  <SelectValue placeholder="All Statuses" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Statuses</SelectItem>
                  <SelectItem value="completed">Completed</SelectItem>
                  <SelectItem value="voided">Voided</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </CardContent>
        </Card>

        {/* Sales table */}
        {isLoading ? (
          <div className="flex items-center justify-center h-64">
            <Loader2 className="h-8 w-8 animate-spin text-primary" />
          </div>
        ) : filteredSales.length === 0 ? (
          <Card>
            <CardContent className="flex flex-col items-center justify-center py-12">
              <Receipt className="h-12 w-12 text-muted-foreground mb-4" />
              <p className="text-muted-foreground">No sales found</p>
            </CardContent>
          </Card>
        ) : (
          <Card>
            <CardHeader>
              <CardTitle>Sales ({filteredSales.length})</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead>
                    <tr className="border-b text-left text-sm text-muted-foreground">
                      <th className="pb-3 font-medium">Invoice</th>
                      <th className="pb-3 font-medium">Date</th>
                      <th className="pb-3 font-medium">Cashier</th>
                      <th className="pb-3 font-medium">Items</th>
                      <th className="pb-3 font-medium text-right">Total</th>
                      <th className="pb-3 font-medium text-center">Status</th>
                      <th className="pb-3 font-medium text-right">Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {filteredSales.map((sale) => (
                      <tr key={sale.id} className="border-b last:border-0">
                        <td className="py-3">
                          <p className="font-mono text-sm">{sale.invoice_no}</p>
                        </td>
                        <td className="py-3 text-sm text-muted-foreground">
                          {formatDate(sale.created_at)}
                        </td>
                        <td className="py-3 text-sm">{sale.employee_name || '-'}</td>
                        <td className="py-3 text-sm">{sale.item_count} items</td>
                        <td className="py-3 text-right font-medium">
                          {formatCurrency(sale.total_amount)}
                        </td>
                        <td className="py-3 text-center">
                          <span
                            className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium capitalize ${getStatusBadge(
                              sale.status
                            )}`}
                          >
                            {sale.status}
                          </span>
                        </td>
                        <td className="py-3 text-right">
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => viewSaleDetails(sale)}
                          >
                            <Eye className="h-4 w-4" />
                          </Button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </CardContent>
          </Card>
        )}
      </div>

      {/* View Sale Dialog */}
      <Dialog open={viewDialogOpen} onOpenChange={setViewDialogOpen}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>Sale Details</DialogTitle>
            <DialogDescription>
              Invoice: {selectedSale?.invoice_no}
            </DialogDescription>
          </DialogHeader>

          {selectedSale && (
            <div className="space-y-4">
              {/* Sale info */}
              <div className="grid grid-cols-2 gap-4 text-sm">
                <div>
                  <p className="text-muted-foreground">Date</p>
                  <p className="font-medium">{formatDate(selectedSale.created_at)}</p>
                </div>
                <div>
                  <p className="text-muted-foreground">Cashier</p>
                  <p className="font-medium">{selectedSale.employee_name || '-'}</p>
                </div>
                <div>
                  <p className="text-muted-foreground">Payment Method</p>
                  <div className="flex items-center gap-1 font-medium">
                    {PAYMENT_ICONS[getPrimaryPaymentMethod(selectedSale)]}
                    <span className="capitalize">{getPrimaryPaymentMethod(selectedSale)}</span>
                  </div>
                </div>
                <div>
                  <p className="text-muted-foreground">Status</p>
                  <span
                    className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium capitalize ${getStatusBadge(
                      selectedSale.status
                    )}`}
                  >
                    {selectedSale.status}
                  </span>
                </div>
              </div>

              {/* Items */}
              <div>
                <p className="text-sm font-medium mb-2">Items</p>
                <div className="rounded-lg border divide-y">
                  {selectedSale.items?.map((item) => (
                    <div key={item.id} className="flex items-center justify-between p-3">
                      <div>
                        <p className="font-medium">{item.product_name}</p>
                        <p className="text-sm text-muted-foreground">
                          {formatCurrency(item.unit_price)} x {item.quantity}
                        </p>
                      </div>
                      <p className="font-medium">{formatCurrency(item.subtotal)}</p>
                    </div>
                  ))}
                </div>
              </div>

              {/* Totals */}
              <div className="rounded-lg bg-muted p-4 space-y-2">
                <div className="flex justify-between text-sm">
                  <span>Subtotal</span>
                  <span>{formatCurrency(selectedSale.subtotal)}</span>
                </div>
                {parseFloat(selectedSale.discount_amount) > 0 && (
                  <div className="flex justify-between text-sm text-destructive">
                    <span>Discount</span>
                    <span>-{formatCurrency(selectedSale.discount_amount)}</span>
                  </div>
                )}
                <div className="flex justify-between font-bold pt-2 border-t">
                  <span>Total</span>
                  <span>{formatCurrency(selectedSale.total_amount)}</span>
                </div>
                {selectedSale.payments && selectedSale.payments.length > 0 && (
                  <>
                    <div className="flex justify-between text-sm">
                      <span>Amount Paid</span>
                      <span>{formatCurrency(selectedSale.amount_paid)}</span>
                    </div>
                    {parseFloat(selectedSale.change_amount) > 0 && (
                      <div className="flex justify-between text-sm">
                        <span>Change</span>
                        <span>{formatCurrency(selectedSale.change_amount)}</span>
                      </div>
                    )}
                  </>
                )}
              </div>

              {/* Void info */}
              {selectedSale.status === 'voided' && (
                <div className="rounded-lg bg-destructive/10 p-4">
                  <p className="text-sm font-medium text-destructive">Voided</p>
                  <p className="text-sm text-muted-foreground">
                    Reason: {selectedSale.void_reason}
                  </p>
                </div>
              )}
            </div>
          )}

          <DialogFooter>
            {selectedSale?.status === 'completed' && hasPermission(PERMISSIONS.SALES_VOID) && (
              <Button
                variant="destructive"
                onClick={() => {
                  setVoidDialogOpen(true);
                }}
              >
                <XCircle className="h-4 w-4 mr-2" />
                Void Sale
              </Button>
            )}
            <Button variant="outline" onClick={() => setViewDialogOpen(false)}>
              Close
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Void Confirmation Dialog */}
      <Dialog open={voidDialogOpen} onOpenChange={setVoidDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Void Sale</DialogTitle>
            <DialogDescription>
              This will void the sale and restore inventory. This action cannot be undone.
            </DialogDescription>
          </DialogHeader>

          <div className="py-4">
            <Label htmlFor="reason">Reason for voiding *</Label>
            <Input
              id="reason"
              value={voidReason}
              onChange={(e) => setVoidReason(e.target.value)}
              placeholder="Enter reason..."
              className="mt-2"
            />
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setVoidDialogOpen(false)}>
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={handleVoidSale}
              disabled={isVoiding || !voidReason}
            >
              {isVoiding ? (
                <>
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  Voiding...
                </>
              ) : (
                'Confirm Void'
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

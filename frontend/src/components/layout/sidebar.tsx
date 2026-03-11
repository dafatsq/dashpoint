'use client';

import Link from 'next/link';
import { usePathname, useRouter } from 'next/navigation';
import { useAuth, PERMISSIONS } from '@/contexts/auth-context';
import { AccountManager, SavedAccount } from '@/lib/account-manager';
import { AccountSwitcher } from '@/components/account-switcher';
import { cn } from '@/lib/utils';
import {
  Store,
  UserCog,
  ChevronLeft,
  ChevronRight,
  LogOut,
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { useState, useEffect } from 'react';

interface NavItem {
  href: string;
  label: string;
  icon: React.ReactNode;
  permission?: string;
}

import { navItems } from '@/lib/nav-config';

interface SidebarProps {
  onNavigate?: () => void;
}

export function Sidebar({ onNavigate }: SidebarProps) {
  const pathname = usePathname();
  const router = useRouter();
  const { user, hasPermission, hasAnyPermission, logout, pinLogin } = useAuth();
  const [collapsed, setCollapsed] = useState(false);
  const [savedAccounts, setSavedAccounts] = useState<SavedAccount[]>(AccountManager.getSavedAccounts());
  const [showSwitchDialog, setShowSwitchDialog] = useState(false);
  const [selectedAccount, setSelectedAccount] = useState<SavedAccount | null>(null);
  const [pin, setPin] = useState('');
  const [pinError, setPinError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [refreshTrigger, setRefreshTrigger] = useState(0);

  useEffect(() => {
    // Update saved accounts when dialog is opened/closed
    if (showSwitchDialog) {
      setSavedAccounts(AccountManager.getSavedAccounts());
      setRefreshTrigger(prev => prev + 1); // Trigger refresh in AccountSwitcher
    }
  }, [showSwitchDialog]);

  useEffect(() => {
    // Refresh saved accounts list
    setSavedAccounts(AccountManager.getSavedAccounts());
  }, []);

  const filteredNavItems = navItems.filter(
    (item) => {
      if (item.permissions) {
        return hasAnyPermission(item.permissions);
      }
      return !item.permission || hasPermission(item.permission);
    }
  );

  const handleLogout = async () => {
    await logout();
    window.location.href = '/login';
  };

  const handleAccountClick = (account: SavedAccount) => {
    setSelectedAccount(account);
    setPin('');
    setPinError('');
    setShowSwitchDialog(true);
  };

  const handlePinSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedAccount) return;

    setPinError('');
    setIsSubmitting(true);

    try {
      const result = await pinLogin(selectedAccount.id, pin);

      if (result.success) {
        setShowSwitchDialog(false);
        setSelectedAccount(null);
        setPin('');
        router.push('/');
      } else {
        setPinError(result.error || 'Invalid PIN');
      }
    } catch {
      setPinError('An unexpected error occurred');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <aside
      className={cn(
        'flex h-full flex-col bg-card transition-all duration-300',
        !onNavigate && 'border-r',
        onNavigate ? 'w-full' : (collapsed ? 'w-16' : 'w-64'),
        !onNavigate && 'hidden lg:flex h-screen'
      )}
    >
      {/* Header - Hidden when in mobile drawer */}
      {!onNavigate && (
        <div className="flex h-16 items-center justify-between border-b px-4">
          {!collapsed && (
            <div className="flex items-center gap-2">
              <Store className="h-6 w-6 text-primary" />
              <span className="font-bold text-lg">DashPoint</span>
            </div>
          )}
          <Button
            variant="ghost"
            size="icon"
            onClick={() => setCollapsed(!collapsed)}
            className={cn(collapsed && 'mx-auto')}
          >
            {collapsed ? (
              <ChevronRight className="h-4 w-4" />
            ) : (
              <ChevronLeft className="h-4 w-4" />
            )}
          </Button>
        </div>
      )}

      {/* Navigation */}
      <nav className="flex-1 overflow-y-auto p-2">
        <ul className="space-y-1">
          {filteredNavItems.map((item) => {
            const isActive = pathname === item.href;
            return (
              <li key={item.href}>
                <Link
                  href={item.href}
                  onClick={onNavigate}
                  className={cn(
                    'flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors',
                    isActive
                      ? 'bg-primary text-primary-foreground'
                      : 'text-muted-foreground hover:bg-muted hover:text-foreground',
                    collapsed && 'justify-center px-2'
                  )}
                  title={collapsed ? item.label : undefined}
                >
                  {item.icon}
                  {!collapsed && <span>{item.label}</span>}
                </Link>
              </li>
            );
          })}
        </ul>
      </nav>

      {/* User info & Account Actions */}
      <div className="border-t p-4">
        {!collapsed && user && (
          <div className="mb-3">
            <p className="font-medium text-sm truncate">{user.name}</p>
            <p className="text-xs text-muted-foreground capitalize">{user.role_name}</p>
          </div>
        )}
        <div className={cn('flex gap-2', collapsed && 'flex-col')}>
          {savedAccounts.filter(account => account.id !== user?.id).length > 0 && (
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button
                  variant="ghost"
                  size={collapsed ? 'icon' : 'default'}
                  className={cn('flex-1', collapsed && 'w-full')}
                  title={collapsed ? 'Switch Account' : undefined}
                >
                  <UserCog className="h-4 w-4" />
                  {!collapsed && <span className="ml-2">Switch</span>}
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="start" className="w-56 mb-2 ml-1">
                <DropdownMenuLabel>Saved Accounts ({savedAccounts.filter(account => account.id !== user?.id).length})</DropdownMenuLabel>
                <DropdownMenuSeparator />
                {savedAccounts.filter(account => account.id !== user?.id).map((account) => (
                  <DropdownMenuItem
                    key={account.id}
                    onClick={() => handleAccountClick(account)}
                    className="cursor-pointer"
                  >
                    <div className="flex flex-col">
                      <span className="font-medium">{account.name}</span>
                      <span className="text-xs text-muted-foreground capitalize">{account.role_name}</span>
                    </div>
                  </DropdownMenuItem>
                ))}
              </DropdownMenuContent>
            </DropdownMenu>
          )}
          <Button
            variant="ghost"
            size={collapsed ? 'icon' : 'default'}
            onClick={handleLogout}
            className={cn('flex-1', collapsed && 'w-full')}
            title={collapsed ? 'Logout' : undefined}
          >
            <LogOut className="h-4 w-4" />
            {!collapsed && <span className="ml-2">Logout</span>}
          </Button>
        </div>
      </div>

      {/* Switch Account Dialog */}
      <Dialog open={showSwitchDialog} onOpenChange={setShowSwitchDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Enter PIN</DialogTitle>
            <DialogDescription>
              Enter your PIN to access {selectedAccount?.name}&apos;s account
            </DialogDescription>
          </DialogHeader>

          <form onSubmit={handlePinSubmit} className="space-y-4 mt-4" autoComplete="off">
            <input type="text" name="username" autoComplete="username" style={{ display: 'none' }} />

            {pinError && (
              <div className="flex items-center gap-2 rounded-md bg-destructive/10 p-3 text-sm text-destructive">
                <UserCog className="h-4 w-4 flex-shrink-0" />
                <span>{pinError}</span>
              </div>
            )}

            <div className="space-y-2">
              <Input
                type="password"
                name="pin-entry"
                placeholder="Enter your PIN"
                value={pin}
                onChange={(e) => setPin(e.target.value)}
                onFocus={(e) => e.target.removeAttribute('readonly')}
                readOnly
                required
                disabled={isSubmitting}
                autoFocus
                maxLength={6}
                pattern="\d*"
                inputMode="numeric"
                autoComplete="new-password"
                data-form-type="other"
                data-lpignore="true"
                data-1p-ignore="true"
              />
            </div>

            <div className="flex gap-2">
              <Button
                type="button"
                variant="outline"
                className="flex-1"
                onClick={() => {
                  setShowSwitchDialog(false);
                  setSelectedAccount(null);
                  setPin('');
                  setPinError('');
                }}
                disabled={isSubmitting}
              >
                Cancel
              </Button>
              <Button
                type="submit"
                className="flex-1"
                disabled={isSubmitting || !pin}
              >
                {isSubmitting ? (
                  <>
                    <UserCog className="mr-2 h-4 w-4 animate-spin" />
                    Signing in...
                  </>
                ) : (
                  'Sign In'
                )}
              </Button>
            </div>
          </form>
        </DialogContent>
      </Dialog>
    </aside>
  );
}

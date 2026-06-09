'use client';

import Link from 'next/link';
import { usePathname, useRouter } from 'next/navigation';
import { useAuth } from '@/contexts/auth-context';
import { AccountManager } from '@/lib/account-manager';
import { AccountSwitcher } from '@/components/account-switcher';
import { cn } from '@/lib/utils';
import {
  Store,
  ChevronLeft,
  ChevronRight,
  LogOut,
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { useMemo, useState } from 'react';

import { filterSwitchableAccounts, } from '@/components/account-switcher-utils';
import { filterVisibleNavItems, navItems } from '@/lib/nav-config';

interface SidebarProps {
  onNavigate?: () => void;
}

export function Sidebar({ onNavigate }: SidebarProps) {
  const pathname = usePathname();
  const router = useRouter();
  const { user, hasPermission, hasAnyPermission, logout } = useAuth();
  const [collapsed, setCollapsed] = useState(false);
  const [, setSavedAccountsVersion] = useState(0);
  const [showSwitchDialog, setShowSwitchDialog] = useState(false);

  const filteredNavItems = useMemo(
    () =>
      filterVisibleNavItems(navItems, {
        hasPermission,
        hasAnyPermission,
      }),
    [hasAnyPermission, hasPermission],
  );

  const switchableAccountsCount = filterSwitchableAccounts(
    AccountManager.getSavedAccounts(),
    user?.id,
  ).length;

  const refreshSavedAccounts = () => {
    setSavedAccountsVersion((version) => version + 1);
  };

  const handleLogout = async () => {
    await logout();
    router.replace('/login');
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
          {switchableAccountsCount > 0 && (
            <Button
              variant="ghost"
              size={collapsed ? 'icon' : 'default'}
              onClick={() => setShowSwitchDialog(true)}
              className={cn('flex-1', collapsed && 'w-full')}
              title={collapsed ? 'Switch Account' : undefined}
            >
              <Store className="h-4 w-4" />
              {!collapsed && <span className="ml-2">Switch</span>}
            </Button>
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
      <Dialog
        open={showSwitchDialog}
        onOpenChange={(open) => {
          setShowSwitchDialog(open);
          if (open) {
            refreshSavedAccounts();
          }
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Switch Account</DialogTitle>
          </DialogHeader>
          <AccountSwitcher
            excludeUserId={user?.id}
            refreshTrigger={showSwitchDialog ? 1 : 0}
            onAccountsChange={refreshSavedAccounts}
            onAccountSelect={() => {
              refreshSavedAccounts();
              setShowSwitchDialog(false);
            }}
          />
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowSwitchDialog(false)}>
              Cancel
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </aside>
  );
}

'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/contexts/auth-context';
import { AccountManager, SavedAccount } from '@/lib/account-manager';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { User, Shield, ShieldCheck, ShieldAlert, X, AlertCircle, Loader2, KeyRound } from 'lucide-react';
import { cn } from '@/lib/utils';

interface AccountSwitcherProps {
  onAccountSelect?: () => void;
  refreshTrigger?: number;
}

export function AccountSwitcher({ onAccountSelect, refreshTrigger }: AccountSwitcherProps) {
  const router = useRouter();
  const { pinLogin } = useAuth();
  const [accounts, setAccounts] = useState<SavedAccount[]>(AccountManager.getSavedAccounts());
  const [selectedAccount, setSelectedAccount] = useState<SavedAccount | null>(null);
  const [pin, setPin] = useState('');
  const [error, setError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  useEffect(() => {
    // Refresh accounts when component mounts or refreshTrigger changes
    setAccounts(AccountManager.getSavedAccounts());
  }, [refreshTrigger]);

  const getRoleIcon = (role: string) => {
    switch (role) {
      case 'owner':
        return <ShieldAlert className="h-5 w-5 text-purple-500" />;
      case 'manager':
        return <ShieldCheck className="h-5 w-5 text-blue-500" />;
      case 'cashier':
        return <Shield className="h-5 w-5 text-green-500" />;
      default:
        return <User className="h-5 w-5 text-muted-foreground" />;
    }
  };

  const getRoleBadgeColor = (role: string) => {
    switch (role) {
      case 'owner':
        return 'bg-purple-600 text-white dark:bg-purple-600/90 dark:text-white';
      case 'manager':
        return 'bg-blue-600 text-white dark:bg-blue-600/90 dark:text-white';
      case 'cashier':
        return 'bg-green-600 text-white dark:bg-green-600/90 dark:text-white';
      default:
        return 'bg-gray-600 text-white dark:bg-gray-600/90 dark:text-white';
    }
  };

  const handleAccountClick = (account: SavedAccount) => {
    if (!account.has_pin) {
      setError('This account does not have a PIN configured');
      return;
    }
    setSelectedAccount(account);
    setPin('');
    setError('');
  };

  const handlePinSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedAccount) return;

    setError('');
    setIsSubmitting(true);

    try {
      const result = await pinLogin(selectedAccount.id, pin);

      if (result.success) {
        setSelectedAccount(null);
        setPin('');
        onAccountSelect?.();
        router.push('/');
      } else {
        setError(result.error || 'Invalid PIN');
      }
    } catch {
      setError('An unexpected error occurred');
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleRemoveAccount = (accountId: string, e: React.MouseEvent) => {
    e.stopPropagation();
    AccountManager.removeAccount(accountId);
    setAccounts(AccountManager.getSavedAccounts());
  };

  if (accounts.length === 0) {
    return (
      <div className="text-center py-12">
        <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-muted">
          <User className="h-8 w-8 text-muted-foreground" />
        </div>
        <h3 className="text-lg font-medium mb-2">No Saved Accounts</h3>
        <p className="text-sm text-muted-foreground mb-6">
          Accounts with PIN will be saved here for quick switching
        </p>
      </div>
    );
  }

  return (
    <>
      <div className="space-y-3">
        <div className="flex items-center justify-between mb-4">
          <div>
            <h3 className="text-lg font-semibold">Saved Accounts</h3>
            <p className="text-sm text-muted-foreground">
              Select an account and enter your PIN
            </p>
          </div>
        </div>

        <div className="grid gap-3">
          {accounts.map((account) => (
            <Card
              key={account.id}
              className="cursor-pointer hover:bg-accent transition-colors overflow-hidden min-w-0"
              onClick={() => handleAccountClick(account)}
            >
              <CardContent className="p-4">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3 flex-1 min-w-0">
                    <div className="flex h-10 w-10 sm:h-12 sm:w-12 items-center justify-center rounded-full bg-primary/10 shrink-0">
                      {getRoleIcon(account.role_name)}
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2">
                        <h4 className="font-medium truncate">{account.name}</h4>
                      </div>
                      <div className="flex items-center gap-2 text-sm text-muted-foreground">
                        <span className={cn(
                          'px-1.5 py-0.5 rounded-full text-[10px] sm:text-xs font-medium capitalize shrink-0',
                          getRoleBadgeColor(account.role_name)
                        )}>
                          {account.role_name}
                        </span>
                        {account.email && (
                          <span className="truncate">{account.email}</span>
                        )}
                      </div>
                    </div>
                  </div>
                  <div className="flex items-center gap-2 shrink-0 pl-1">
                    {account.has_pin && (
                      <div className="flex items-center gap-1 text-xs text-muted-foreground mr-1">
                        <KeyRound className="h-3 w-3" />
                        <span className="hidden sm:inline">PIN</span>
                      </div>
                    )}
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-8 w-8 p-0"
                      onClick={(e) => handleRemoveAccount(account.id, e)}
                    >
                      <X className="h-4 w-4" />
                    </Button>
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      </div>

      {/* PIN Dialog */}
      <Dialog open={!!selectedAccount} onOpenChange={(open) => !open && setSelectedAccount(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Enter PIN</DialogTitle>
            <DialogDescription>
              Enter your PIN to access {selectedAccount?.name}&apos;s account
            </DialogDescription>
          </DialogHeader>

          <form onSubmit={handlePinSubmit} className="space-y-4 mt-4" autoComplete="off">
            <input type="text" name="username" autoComplete="username" style={{ display: 'none' }} />

            {error && (
              <div className="flex items-center gap-2 rounded-md bg-destructive/10 p-3 text-sm text-destructive">
                <AlertCircle className="h-4 w-4 flex-shrink-0" />
                <span>{error}</span>
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
                onClick={() => setSelectedAccount(null)}
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
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
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
    </>
  );
}

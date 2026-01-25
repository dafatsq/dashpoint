'use client';

import { useState, useEffect, Suspense } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { useAuth } from '@/contexts/auth-context';
import { AccountManager } from '@/lib/account-manager';
import { AccountSwitcher } from '@/components/account-switcher';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Store, Loader2, AlertCircle, Mail, Info } from 'lucide-react';

// Message mapping for different logout reasons
const LOGOUT_MESSAGES: Record<string, string> = {
  account_inactive: 'Your account has been deactivated. Please contact an administrator.',
  account_deactivated: 'Your account has been deactivated by an administrator.',
  account_deleted: 'Your account has been deleted.',
  force_logout: 'You have been logged out by an administrator.',
  permissions_changed: 'Your permissions have been updated. Please log in again.',
  role_changed: 'Your role has been changed. Please log in again.',
};

function LoginPageContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { login, isLoading: authLoading } = useAuth();
  
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [infoMessage, setInfoMessage] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [hasSavedAccounts, setHasSavedAccounts] = useState(false);
  const [activeTab, setActiveTab] = useState<'saved' | 'email'>('saved');

  // Check for logout message from URL params
  useEffect(() => {
    const message = searchParams.get('message');
    if (message && LOGOUT_MESSAGES[message]) {
      setInfoMessage(LOGOUT_MESSAGES[message]);
      // Clear the URL parameter without causing a navigation
      window.history.replaceState({}, '', '/login');
    }
  }, [searchParams]);

  // Check for saved accounts on mount and when page becomes visible
  useEffect(() => {
    const checkSavedAccounts = () => {
      const accounts = AccountManager.getSavedAccounts();
      const hasAccounts = accounts.length > 0;
      setHasSavedAccounts(hasAccounts);
      setActiveTab(hasAccounts ? 'saved' : 'email');
    };

    checkSavedAccounts();

    // Re-check when page becomes visible (e.g., after logout redirect)
    const handleVisibilityChange = () => {
      if (document.visibilityState === 'visible') {
        checkSavedAccounts();
      }
    };

    document.addEventListener('visibilitychange', handleVisibilityChange);
    window.addEventListener('focus', checkSavedAccounts);

    return () => {
      document.removeEventListener('visibilitychange', handleVisibilityChange);
      window.removeEventListener('focus', checkSavedAccounts);
    };
  }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setIsSubmitting(true);

    try {
      const result = await login(email, password);
      
      if (result.success) {
        router.push('/dashboard');
      } else {
        setError(result.error || 'Login failed');
      }
    } catch {
      setError('An unexpected error occurred');
    } finally {
      setIsSubmitting(false);
    }
  };

  if (authLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-background to-muted p-4">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-primary">
            <Store className="h-8 w-8 text-primary-foreground" />
          </div>
          <CardTitle className="text-2xl font-bold">DashPoint POS</CardTitle>
          <CardDescription>
            Sign in to access your point of sale system
          </CardDescription>
        </CardHeader>
        <CardContent>
          {infoMessage && (
            <div className="flex items-center gap-2 rounded-md bg-blue-50 dark:bg-blue-950 p-3 text-sm text-blue-800 dark:text-blue-200 mb-4">
              <Info className="h-4 w-4 flex-shrink-0" />
              <span>{infoMessage}</span>
            </div>
          )}
          
          <Tabs value={activeTab} onValueChange={(v) => setActiveTab(v as 'saved' | 'email')}>
            <TabsList className="grid w-full grid-cols-2 mb-6">
              <TabsTrigger value="saved" disabled={!hasSavedAccounts}>
                Quick Access
              </TabsTrigger>
              <TabsTrigger value="email">
                <Mail className="h-4 w-4 mr-2" />
                Email Login
              </TabsTrigger>
            </TabsList>

            <TabsContent value="saved" className="mt-0">
              <AccountSwitcher />
            </TabsContent>

            <TabsContent value="email" className="mt-0">
              <form onSubmit={handleSubmit} className="space-y-4">
                {error && (
                  <div className="flex items-center gap-2 rounded-md bg-destructive/10 p-3 text-sm text-destructive">
                    <AlertCircle className="h-4 w-4" />
                    <span>{error}</span>
                  </div>
                )}
                
                <div className="space-y-2">
                  <Label htmlFor="email">Email</Label>
                  <Input
                    id="email"
                    type="email"
                    placeholder="Enter your email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    required
                    disabled={isSubmitting}
                  />
                </div>
                
                <div className="space-y-2">
                  <Label htmlFor="password">Password</Label>
                  <Input
                    id="password"
                    type="password"
                    placeholder="Enter your password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    required
                    disabled={isSubmitting}
                  />
                </div>
                
                <Button
                  type="submit"
                  className="w-full"
                  disabled={isSubmitting}
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
              </form>

              <div className="mt-6 text-center text-sm text-muted-foreground">
                <p>Demo Credentials:</p>
                <p className="mt-1 font-mono text-xs">
                  owner@dashpoint.local / owner123
                </p>
              </div>
            </TabsContent>
          </Tabs>
        </CardContent>
      </Card>
    </div>
  );
}

export default function LoginPage() {
  return (
    <Suspense fallback={
      <div className="min-h-screen flex items-center justify-center bg-background">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    }>
      <LoginPageContent />
    </Suspense>
  );
}
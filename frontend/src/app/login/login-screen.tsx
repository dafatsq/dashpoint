'use client';

import { useCallback, useEffect, useState, type FormEvent } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import {
  Info,
  Loader2,
  Mail,
} from "lucide-react";

import { AccountSwitcher } from "@/components/account-switcher";
import { Logo } from "@/components/ui/logo";

import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useAuth } from "@/contexts/auth-context";
import { AccountManager } from "@/lib/account-manager";
import { api } from "@/lib/api";
import { LoginEmailForm } from "./login-email-form";
import { LoginOwnerSetup } from "./login-owner-setup";
import {
  getDefaultLoginTab,
  getEffectiveSaveAccountDecision,
  getHasSavedAccounts,
  getLoginInfoMessage,
  getSaveLoginControlState,
  isStoredPreferenceEnabled,
  type LoginTab,
} from "./login-helpers";

export function LoginScreen() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { login, isLoading: authLoading } = useAuth();

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [infoMessage, setInfoMessage] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [saveLogin, setSaveLogin] = useState(false);
  const [isDeviceTrusted, setIsDeviceTrusted] = useState(false);
  const [isClient, setIsClient] = useState(false);
  const [setupRequired, setSetupRequired] = useState(false);
  const [hasSavedAccounts, setHasSavedAccounts] = useState(false);
  const [activeTab, setActiveTab] = useState<LoginTab>("saved");
  const saveLoginControlState = getSaveLoginControlState(
    isDeviceTrusted,
    saveLogin,
  );

  const refreshSavedAccounts = useCallback((forceDefaultTab = false) => {
    const hasAccounts = getHasSavedAccounts(AccountManager.getSavedAccounts());

    setHasSavedAccounts(hasAccounts);
    setActiveTab((currentTab) => {
      if (forceDefaultTab) {
        return getDefaultLoginTab(hasAccounts);
      }

      if (!hasAccounts && currentTab === "saved") {
        return "email";
      }

      return currentTab;
    });
  }, []);

  useEffect(() => {
    const nextInfoMessage = getLoginInfoMessage(searchParams.get("message"));
    if (!nextInfoMessage) {
      return;
    }

    setInfoMessage(nextInfoMessage);
    window.history.replaceState({}, "", "/login");
  }, [searchParams]);

  useEffect(() => {
    setIsClient(true);
    let cancelled = false;
    api
      .getSetupStatus()
      .then((result) => {
        if (!cancelled && !result.error && result.data?.setup_required) {
          setSetupRequired(true);
        }
      })
      .catch(() => {
        // Setup detection is best-effort; the login form remains usable.
      });
    return () => {
      cancelled = true;
    };
  }, []);

  useEffect(() => {
    const trustedDeviceEnabled = isStoredPreferenceEnabled(
      localStorage.getItem("dashpoint_device_trusted"),
    );
    setIsDeviceTrusted(trustedDeviceEnabled);
    setSaveLogin(trustedDeviceEnabled);
    refreshSavedAccounts(true);

    const handleVisibilityChange = () => {
      if (document.visibilityState === "visible") {
        refreshSavedAccounts(true);
      }
    };
    const handleWindowFocus = () => {
      refreshSavedAccounts(true);
    };

    window.addEventListener("focus", handleWindowFocus);
    document.addEventListener("visibilitychange", handleVisibilityChange);

    return () => {
      window.removeEventListener("focus", handleWindowFocus);
      document.removeEventListener("visibilitychange", handleVisibilityChange);
    };
  }, [refreshSavedAccounts]);

  const handleSubmit = useCallback(
    async (event: FormEvent<HTMLFormElement>) => {
      event.preventDefault();
      setError("");
      setIsSubmitting(true);

      try {
        const result = await login(
          email,
          password,
          getEffectiveSaveAccountDecision(isDeviceTrusted, saveLogin),
        );

        if (result.success) {
          router.push("/");
        } else {
          setError(result.error || "Login failed");
        }
      } catch {
        setError("An unexpected error occurred");
      } finally {
        setIsSubmitting(false);
      }
    },
    [email, isDeviceTrusted, login, password, router, saveLogin],
  );

  if (authLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-background to-muted p-4">
      <div className="w-full max-w-md">
        <div className="relative text-center flex flex-col space-y-1.5 p-6 pb-4">
          <Logo className="mx-auto mb-4 h-16 w-16 text-primary" />
          <h2 className="text-2xl font-bold tracking-tight text-foreground">DashPoint POS</h2>
          <p className="text-sm text-muted-foreground">
            {isClient && setupRequired
              ? "Create your owner account to get started"
              : "Sign in to access your point of sale system"}
          </p>
        </div>
        <div className="p-4 sm:p-6">
          {infoMessage && (
            <div className="mb-4 flex items-center gap-2 rounded-md bg-blue-50 p-3 text-sm text-blue-800 dark:bg-blue-950 dark:text-blue-200">
              <Info className="h-4 w-4 flex-shrink-0" />
              <span>{infoMessage}</span>
            </div>
          )}

          {isClient && setupRequired ? (
            <LoginOwnerSetup
              onSuccess={() => {
                setSetupRequired(false);
                router.push("/login?message=owner_created");
              }}
            />
          ) : (
          <Tabs
            value={activeTab}
            onValueChange={(value) => setActiveTab(value as LoginTab)}
          >
            <TabsList className="mb-4 grid w-full grid-cols-2 sm:mb-6">
              <TabsTrigger value="saved" disabled={!hasSavedAccounts}>
                Quick Access
              </TabsTrigger>
              <TabsTrigger value="email">
                <Mail className="mr-2 h-4 w-4" />
                Email Login
              </TabsTrigger>
            </TabsList>

            <TabsContent value="saved" className="mt-0">
              <AccountSwitcher onAccountsChange={() => refreshSavedAccounts()} />
            </TabsContent>

            <TabsContent value="email" className="mt-0">
              <LoginEmailForm
                email={email}
                password={password}
                error={error}
                isSubmitting={isSubmitting}
                saveLoginChecked={saveLoginControlState.checked}
                saveLoginDisabled={saveLoginControlState.disabled}
                onEmailChange={setEmail}
                onPasswordChange={setPassword}
                onSaveLoginChange={setSaveLogin}
                onSubmit={handleSubmit}
              />
            </TabsContent>
          </Tabs>
          )}
        </div>
      </div>
    </div>
  );
}

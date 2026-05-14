'use client';

import { useCallback, useEffect, useState, type FormEvent } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import {
  Info,
  Loader2,
  Mail,
  Settings,
  Store,
} from "lucide-react";

import { AccountSwitcher } from "@/components/account-switcher";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useAuth } from "@/contexts/auth-context";
import { AccountManager } from "@/lib/account-manager";

import { LoginDemoAccess } from "./login-demo-access";
import { LoginEmailForm } from "./login-email-form";
import {
  getDefaultLoginTab,
  getEffectiveSaveAccountDecision,
  getHasSavedAccounts,
  getLoginInfoMessage,
  getSaveLoginControlState,
  getShowDemoAccessPreference,
  isStoredPreferenceEnabled,
  type LoginTab,
} from "./login-helpers";
import { LoginSettingsPanel } from "./login-settings-panel";

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
  const [showDemoAccess, setShowDemoAccess] = useState(false);
  const [showSettings, setShowSettings] = useState(false);
  const [isClient, setIsClient] = useState(false);
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

  const handleDeviceTrustedChange = useCallback(
    (checked: boolean) => {
      setIsDeviceTrusted(checked);

      if (checked) {
        localStorage.setItem("dashpoint_device_trusted", "true");
        return;
      }

      localStorage.removeItem("dashpoint_device_trusted");
      AccountManager.clearAll();
      refreshSavedAccounts(true);
    },
    [refreshSavedAccounts],
  );

  const handleDemoAccessChange = useCallback((checked: boolean) => {
    setShowDemoAccess(checked);

    if (checked) {
      localStorage.setItem("dashpoint_demo_access", "true");
      return;
    }

    localStorage.removeItem("dashpoint_demo_access");
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
    setIsDeviceTrusted(
      isStoredPreferenceEnabled(localStorage.getItem("dashpoint_device_trusted")),
    );
    setShowDemoAccess(
      getShowDemoAccessPreference(
        process.env.NEXT_PUBLIC_ENABLE_QUICK_DEMO_ACCESS === "true",
        localStorage.getItem("dashpoint_demo_access"),
      ),
    );
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
      <Card className="w-full max-w-md">
        <CardHeader className="relative text-center">
          <Button
            variant="ghost"
            size="icon"
            className="absolute right-4 top-4 text-muted-foreground hover:text-primary"
            onClick={() => setShowSettings((current) => !current)}
            type="button"
          >
            <Settings className="h-5 w-5" />
          </Button>
          <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-primary">
            <Store className="h-8 w-8 text-primary-foreground" />
          </div>
          <CardTitle className="text-2xl font-bold">DashPoint POS</CardTitle>
          <CardDescription>
            Sign in to access your point of sale system
          </CardDescription>
        </CardHeader>
        <CardContent className="p-4 sm:p-6">
          {showSettings && isClient && (
            <LoginSettingsPanel
              isDeviceTrusted={isDeviceTrusted}
              showDemoAccess={showDemoAccess}
              onDeviceTrustedChange={handleDeviceTrustedChange}
              onDemoAccessChange={handleDemoAccessChange}
            />
          )}

          {infoMessage && (
            <div className="mb-4 flex items-center gap-2 rounded-md bg-blue-50 p-3 text-sm text-blue-800 dark:bg-blue-950 dark:text-blue-200">
              <Info className="h-4 w-4 flex-shrink-0" />
              <span>{infoMessage}</span>
            </div>
          )}

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

              {showDemoAccess && isClient && (
                <LoginDemoAccess
                  onSelectCredentials={(nextEmail, nextPassword) => {
                    setEmail(nextEmail);
                    setPassword(nextPassword);
                  }}
                />
              )}
            </TabsContent>
          </Tabs>
        </CardContent>
      </Card>
    </div>
  );
}

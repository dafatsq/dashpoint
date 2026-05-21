"use client";

import {
  createContext,
  useContext,
  useState,
  ReactNode,
  useCallback,
} from "react";
import { ErrorDialog, GlobalErrorState } from "@/components/ui/error-dialog";

interface ErrorContextType {
  showError: (title: string, message: unknown) => void;
  clearError: () => void;
}

const ErrorContext = createContext<ErrorContextType | null>(null);

function normalizeErrorMessage(message: unknown): string {
  if (typeof message === "string") {
    return message;
  }

  if (message && typeof message === "object" && "message" in message) {
    const candidate = (message as { message?: unknown }).message;
    if (typeof candidate === "string") {
      return candidate;
    }
  }

  try {
    return JSON.stringify(message);
  } catch {
    return "An unexpected error occurred.";
  }
}

export function ErrorProvider({ children }: { children: ReactNode }) {
  const [errorState, setErrorState] = useState<GlobalErrorState | null>(null);

  const showError = useCallback((title: string, message: unknown) => {
    setErrorState({ title, message: normalizeErrorMessage(message) });
  }, []);

  const clearError = useCallback(() => {
    setErrorState(null);
  }, []);

  return (
    <ErrorContext.Provider value={{ showError, clearError }}>
      {children}
      <ErrorDialog error={errorState} onClose={clearError} />
    </ErrorContext.Provider>
  );
}

export function useGlobalError() {
  const context = useContext(ErrorContext);
  if (!context) {
    throw new Error("useGlobalError must be used within an ErrorProvider");
  }
  return context;
}

'use client';

import type { FormEvent } from "react";
import { AlertCircle, Loader2 } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

interface LoginEmailFormProps {
  email: string;
  password: string;
  error: string;
  isSubmitting: boolean;
  saveLoginChecked: boolean;
  saveLoginDisabled: boolean;
  onEmailChange: (value: string) => void;
  onPasswordChange: (value: string) => void;
  onSaveLoginChange: (checked: boolean) => void;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
}

export function LoginEmailForm({
  email,
  password,
  error,
  isSubmitting,
  saveLoginChecked,
  saveLoginDisabled,
  onEmailChange,
  onPasswordChange,
  onSaveLoginChange,
  onSubmit,
}: LoginEmailFormProps) {
  return (
    <form onSubmit={onSubmit} className="space-y-4">
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
          onChange={(event) => onEmailChange(event.target.value)}
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
          onChange={(event) => onPasswordChange(event.target.value)}
          required
          disabled={isSubmitting}
        />
      </div>

      <div className="flex items-start space-x-3 rounded-md border border-border/60 bg-muted/30 px-3 py-2.5">
        <Checkbox
          id="save-login"
          checked={saveLoginChecked}
          onCheckedChange={(checked) => onSaveLoginChange(checked === true)}
          disabled={isSubmitting || saveLoginDisabled}
        />
        <div className="grid gap-1.5 leading-none">
          <Label
            htmlFor="save-login"
            className="cursor-pointer text-sm font-medium"
          >
            Save login on this device
          </Label>
          <p className="text-xs text-muted-foreground">
            Save this account for Quick Access instead of adding it
            automatically after sign-in.
          </p>
        </div>
      </div>

      <Button type="submit" className="w-full" disabled={isSubmitting}>
        {isSubmitting ? (
          <>
            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            Signing in...
          </>
        ) : (
          "Sign In"
        )}
      </Button>
    </form>
  );
}

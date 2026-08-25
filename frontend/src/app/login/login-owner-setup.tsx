'use client';

import { useState, type FormEvent } from "react";
import { AlertCircle, Loader2 } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { api } from "@/lib/api";

import {
  normalizeOwnerEmail,
  validateOwnerSetupInput,
} from "./setup-helpers";

interface LoginOwnerSetupProps {
  onSuccess: () => void;
}

export function LoginOwnerSetup({ onSuccess }: LoginOwnerSetupProps) {
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [pin, setPin] = useState("");
  const [error, setError] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setError("");

    const validationError = validateOwnerSetupInput({
      name,
      email,
      password,
      pin,
    });
    if (validationError) {
      setError(validationError);
      return;
    }

    setIsSubmitting(true);
    try {
      const result = await api.createInitialOwner({
        name: name.trim(),
        email: normalizeOwnerEmail(email),
        password,
        pin: pin.trim(),
      });

      if (result.error) {
        setError(result.error);
        return;
      }

      onSuccess();
    } catch {
      setError("An unexpected error occurred");
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="rounded-md border border-primary/20 bg-primary/5 p-3 text-sm text-muted-foreground">
        This instance has no user accounts yet. Create the first owner account
        to start using DashPoint.
      </div>

      {error && (
        <div className="flex items-center gap-2 rounded-md bg-destructive/10 p-3 text-sm text-destructive">
          <AlertCircle className="h-4 w-4" />
          <span>{error}</span>
        </div>
      )}

      <div className="space-y-2">
        <Label htmlFor="setup-name">Your name</Label>
        <Input
          id="setup-name"
          type="text"
          placeholder="Enter the owner's name"
          value={name}
          onChange={(event) => setName(event.target.value)}
          required
          disabled={isSubmitting}
          maxLength={100}
        />
      </div>

      <div className="space-y-2">
        <Label htmlFor="setup-email">Email</Label>
        <Input
          id="setup-email"
          type="email"
          placeholder="Enter your email"
          value={email}
          onChange={(event) => setEmail(event.target.value)}
          required
          disabled={isSubmitting}
        />
      </div>

      <div className="space-y-2">
        <Label htmlFor="setup-password">Password</Label>
        <Input
          id="setup-password"
          type="password"
          placeholder="Choose a strong password"
          value={password}
          onChange={(event) => setPassword(event.target.value)}
          required
          disabled={isSubmitting}
        />
        <p className="text-xs text-muted-foreground">
          At least 8 characters. This protects all business data.
        </p>
      </div>

      <div className="space-y-2">
        <Label htmlFor="setup-pin">POS PIN</Label>
        <Input
          id="setup-pin"
          type="password"
          inputMode="numeric"
          placeholder="4 to 6 digits"
          value={pin}
          onChange={(event) => setPin(event.target.value)}
          required
          disabled={isSubmitting}
          maxLength={6}
        />
        <p className="text-xs text-muted-foreground">
          Used for quick sign-in at the register.
        </p>
      </div>

      <Button type="submit" className="w-full" disabled={isSubmitting}>
        {isSubmitting ? (
          <>
            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            Creating owner account...
          </>
        ) : (
          "Create Owner Account"
        )}
      </Button>
    </form>
  );
}

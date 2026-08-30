import { describe, expect, it } from "vitest";

import {
  MIN_OWNER_PASSWORD_LENGTH,
  MIN_OWNER_PIN_LENGTH,
  MAX_OWNER_PIN_LENGTH,
  normalizeOwnerEmail,
  validateOwnerSetupInput,
} from "./setup-helpers";

describe("normalizeOwnerEmail", () => {
  it("trims and lowercases the email", () => {
    expect(normalizeOwnerEmail("  Owner@Example.COM  ")).toBe(
      "owner@example.com",
    );
  });
});

describe("validateOwnerSetupInput", () => {
  const validInput = {
    name: "Store Owner",
    email: "owner@example.com",
    password: "long-enough-password",
    pin: "1234",
  };

  it("accepts a fully valid input", () => {
    expect(validateOwnerSetupInput(validInput)).toBe("");
  });

  it("rejects a missing name", () => {
    expect(validateOwnerSetupInput({ ...validInput, name: "   " })).toBe(
      "Name is required",
    );
  });

  it("rejects an invalid email", () => {
    expect(validateOwnerSetupInput({ ...validInput, email: "not-an-email" })).toBe(
      "Invalid email format",
    );
  });

  it("rejects a short password", () => {
    expect(
      validateOwnerSetupInput({
        ...validInput,
        password: "a".repeat(MIN_OWNER_PASSWORD_LENGTH - 1),
      }),
    ).toBe(
      `Password must be at least ${MIN_OWNER_PASSWORD_LENGTH} characters long`,
    );
  });

  it("rejects a PIN below the minimum length", () => {
    expect(
      validateOwnerSetupInput({
        ...validInput,
        pin: "1".repeat(MIN_OWNER_PIN_LENGTH - 1),
      }),
    ).toBe(`PIN must be 4 to 6 digits`);
  });

  it("rejects a PIN above the maximum length", () => {
    expect(
      validateOwnerSetupInput({
        ...validInput,
        pin: "1".repeat(MAX_OWNER_PIN_LENGTH + 1),
      }),
    ).toBe(`PIN must be 4 to 6 digits`);
  });

  it("rejects a PIN containing non-digits", () => {
    expect(validateOwnerSetupInput({ ...validInput, pin: "12ab" })).toBe(
      "PIN must be 4 to 6 digits",
    );
  });
});

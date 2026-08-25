export interface OwnerSetupInput {
  name: string;
  email: string;
  password: string;
  pin: string;
}

const MIN_NAME_LENGTH = 1;
const MAX_NAME_LENGTH = 100;
const MAX_EMAIL_LENGTH = 254;
export const MIN_OWNER_PASSWORD_LENGTH = 8;
const MAX_PASSWORD_LENGTH = 128;
export const MIN_OWNER_PIN_LENGTH = 4;
export const MAX_OWNER_PIN_LENGTH = 6;

export function normalizeOwnerEmail(email: string): string {
  return email.trim().toLowerCase();
}

function isValidPin(pin: string): boolean {
  if (
    pin.length < MIN_OWNER_PIN_LENGTH ||
    pin.length > MAX_OWNER_PIN_LENGTH
  ) {
    return false;
  }
  return /^\d+$/.test(pin);
}

/**
 * Validates the initial owner setup form client-side, mirroring the backend
 * validation of POST /api/v1/setup/owner. Returns an empty string when valid,
 * or a user-facing error message otherwise.
 */
export function validateOwnerSetupInput(input: OwnerSetupInput): string {
  const name = input.name.trim();
  if (name.length < MIN_NAME_LENGTH) {
    return "Name is required";
  }
  if (name.length > MAX_NAME_LENGTH) {
    return "Name is too long";
  }

  const email = normalizeOwnerEmail(input.email);
  if (!email) {
    return "Email is required";
  }
  if (email.length > MAX_EMAIL_LENGTH || !email.includes("@")) {
    return "Invalid email format";
  }

  if (input.password.length < MIN_OWNER_PASSWORD_LENGTH) {
    return `Password must be at least ${MIN_OWNER_PASSWORD_LENGTH} characters long`;
  }
  if (input.password.length > MAX_PASSWORD_LENGTH) {
    return "Password is too long";
  }

  const pin = input.pin.trim();
  if (!isValidPin(pin)) {
    return "PIN must be 4 to 6 digits";
  }

  return "";
}

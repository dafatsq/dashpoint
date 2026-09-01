export interface DemoCredential {
  role: string;
  email: string;
  pass: string;
}

// Demo deployment only (dashpoint-demo branch): the login buttons always
// exist and fill these accounts. The backend ensures these accounts exist
// and stay active on every start (backend/internal/database/demo_accounts.go),
// and the passwords ship in this public bundle by design — never put real
// accounts here.
const DEMO_CREDENTIALS: readonly DemoCredential[] = [
  { role: "Owner", email: "demo-owner@dashpoint.local", pass: "demo1234" },
  { role: "Manager", email: "demo-manager@dashpoint.local", pass: "demo1234" },
  { role: "Cashier", email: "demo-cashier@dashpoint.local", pass: "demo1234" },
];

export function getDemoLoginCredentials(): readonly DemoCredential[] {
  return DEMO_CREDENTIALS;
}

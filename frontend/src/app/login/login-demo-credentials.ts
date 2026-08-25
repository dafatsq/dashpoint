// Demo credentials live in their own module so they can be dynamically
// imported behind the NEXT_PUBLIC_ENABLE_QUICK_DEMO_ACCESS build flag. When
// the flag is false at build time the import is dead-code eliminated and the
// credentials never ship in the production bundle.
export interface DemoCredential {
  role: string;
  email: string;
  pass: string;
}

export const DEMO_LOGIN_CREDENTIALS: readonly DemoCredential[] = [
  { role: "Owner", email: "owner@dashpoint.local", pass: "owner123" },
  { role: "Manager", email: "manager@dashpoint.local", pass: "manager123" },
  { role: "Cashier", email: "cashier@dashpoint.local", pass: "cashier123" },
];

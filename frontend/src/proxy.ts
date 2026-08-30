import { NextRequest, NextResponse } from "next/server";

// Defense-in-depth (audit L6): page routes are gated client-side by the auth
// context, and the API re-checks permissions on every call — this proxy
// additionally bounces requests that carry no refresh cookie at the server
// edge, so logged-out visitors never receive a dashboard shell.
//
// The refresh cookie is httpOnly, so presence (not contents) is the signal.
const PUBLIC_PATHS = ["/login"];

export function proxy(req: NextRequest) {
  const { pathname } = req.nextUrl;
  if (PUBLIC_PATHS.some((p) => pathname === p || pathname.startsWith(`${p}/`))) {
    return NextResponse.next();
  }
  if (req.cookies.has("refresh_token")) {
    return NextResponse.next();
  }
  const login = new URL("/login", req.url);
  return NextResponse.redirect(login);
}

export const config = {
  matcher: [
    // Everything except Next internals and static assets.
    "/((?!_next/static|_next/image|favicon.ico|.*\\.(?:png|jpg|jpeg|gif|svg|ico|webp|txt|xml)$).*)",
  ],
};

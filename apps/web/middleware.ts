import { NextResponse, type NextRequest } from "next/server";

export function middleware(request: NextRequest) {
  const { pathname, search } = request.nextUrl;
  const hasAccess = !!request.cookies.get("halomail_access")?.value;
  const hasRefresh = !!request.cookies.get("halomail_refresh")?.value;
  const isAuthenticated = hasAccess || hasRefresh;

  // Protect dashboard routes: require active session
  if (pathname.startsWith("/dashboard")) {
    if (!isAuthenticated) {
      const loginUrl = new URL("/login", request.url);
      loginUrl.searchParams.set("redirect", `${pathname}${search}`);
      return NextResponse.redirect(loginUrl);
    }
  }

  // Auth pages: redirect to dashboard if user already has an active session
  const authRoutes = ["/login", "/register", "/forgot-password", "/reset-password"];
  if (authRoutes.includes(pathname)) {
    if (isAuthenticated) {
      const redirectParam = request.nextUrl.searchParams.get("redirect");
      const targetUrl =
        redirectParam && redirectParam.startsWith("/") && !redirectParam.startsWith("//")
          ? new URL(redirectParam, request.url)
          : new URL("/dashboard", request.url);
      return NextResponse.redirect(targetUrl);
    }
  }

  return NextResponse.next();
}

export const config = {
  matcher: [
    "/dashboard/:path*",
    "/login",
    "/register",
    "/forgot-password",
    "/reset-password",
  ],
};

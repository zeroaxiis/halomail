import { NextRequest, NextResponse } from "next/server";

export const runtime = "nodejs";

const isProd = process.env.NODE_ENV === "production";
let backend = process.env.API_BASE_URL || process.env.NEXT_PUBLIC_API_URL;
if (!backend || (isProd && backend.includes("localhost"))) {
  backend = isProd ? "https://halomailapi.zeroaxiis.tech" : "http://localhost:8080";
}
backend = backend.replace(/\/$/, "");

const cookieOptions = {
  httpOnly: true,
  secure: process.env.NODE_ENV === "production",
  sameSite: "lax" as const,
  path: "/",
};

function getRedirectURI(req: NextRequest): string {
  const custom = process.env.GOOGLE_AUTH_REDIRECT_URL;
  if (custom && custom.startsWith("http")) return custom;
  const origin = process.env.PUBLIC_WEB_URL || req.nextUrl.origin;
  return `${origin.replace(/\/$/, "")}/api/auth/google/callback`;
}

export async function GET(req: NextRequest) {
  const searchParams = req.nextUrl.searchParams;
  const code = searchParams.get("code");
  const state = searchParams.get("state");
  const error = searchParams.get("error");

  const loginUrl = new URL("/login", req.nextUrl.origin);

  if (error) {
    loginUrl.searchParams.set("error", error === "access_denied" ? "google_cancelled" : "google_error");
    return NextResponse.redirect(loginUrl);
  }

  if (!code || !state) {
    loginUrl.searchParams.set("error", "missing_code");
    return NextResponse.redirect(loginUrl);
  }

  // Verify state from cookie
  const stateCookie = req.cookies.get("halomail_oauth_state")?.value;
  let savedState = "";
  let returnTo = "/dashboard";
  if (stateCookie) {
    try {
      const parsed = JSON.parse(stateCookie);
      savedState = parsed.state;
      returnTo = parsed.returnTo || "/dashboard";
    } catch {
      savedState = stateCookie;
    }
  }

  if (!savedState || savedState !== state) {
    loginUrl.searchParams.set("error", "invalid_state");
    return NextResponse.redirect(loginUrl);
  }

  const clientId = process.env.GOOGLE_CLIENT_ID;
  const clientSecret = process.env.GOOGLE_CLIENT_SECRET;
  if (!clientId || !clientSecret) {
    loginUrl.searchParams.set("error", "google_not_configured");
    return NextResponse.redirect(loginUrl);
  }

  const redirectUri = getRedirectURI(req);

  try {
    // 1. Exchange authorization code with Google for tokens
    const tokenRes = await fetch("https://oauth2.googleapis.com/token", {
      method: "POST",
      headers: { "Content-Type": "application/x-www-form-urlencoded" },
      body: new URLSearchParams({
        code,
        client_id: clientId.trim(),
        client_secret: clientSecret.trim(),
        redirect_uri: redirectUri,
        grant_type: "authorization_code",
      }),
      signal: AbortSignal.timeout(15000),
    });

    if (!tokenRes.ok) {
      console.error("Google token exchange failed", await tokenRes.text().catch(() => ""));
      loginUrl.searchParams.set("error", "token_exchange_failed");
      return NextResponse.redirect(loginUrl);
    }

    const tokenData = await tokenRes.json();
    const accessToken = tokenData.access_token;
    if (!accessToken) {
      loginUrl.searchParams.set("error", "no_access_token");
      return NextResponse.redirect(loginUrl);
    }

    // 2. Fetch user profile from Google UserInfo endpoint
    const profileRes = await fetch("https://www.googleapis.com/oauth2/v3/userinfo", {
      headers: { Authorization: `Bearer ${accessToken}` },
      signal: AbortSignal.timeout(10000),
    });

    if (!profileRes.ok) {
      loginUrl.searchParams.set("error", "profile_fetch_failed");
      return NextResponse.redirect(loginUrl);
    }

    const profile = await profileRes.json();
    if (!profile.email || !profile.email_verified) {
      loginUrl.searchParams.set("error", "unverified_email");
      return NextResponse.redirect(loginUrl);
    }

    // 3. Authenticate / register user in Go backend identity service
    const deliverySecret = process.env.OTP_DELIVERY_SECRET || "";
    const backendRes = await fetch(`${backend}/v1/auth/google`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "X-OTP-Delivery-Secret": deliverySecret,
      },
      body: JSON.stringify({
        email: profile.email,
        name: profile.name || "",
        avatarUrl: profile.picture || "",
        googleId: profile.sub || "",
      }),
      signal: AbortSignal.timeout(15000),
    });

    if (!backendRes.ok) {
      const errData = await backendRes.json().catch(() => ({}));
      console.error("Backend Google auth failed", errData);
      loginUrl.searchParams.set("error", errData.message || "login_failed");
      return NextResponse.redirect(loginUrl);
    }

    const data = await backendRes.json();
    const session = data.session;
    if (!session?.accessToken || !session?.refreshToken) {
      loginUrl.searchParams.set("error", "session_error");
      return NextResponse.redirect(loginUrl);
    }

    // 4. Success: Set session cookies and redirect to target page
    const destination = new URL(returnTo.startsWith("/") ? returnTo : "/dashboard", req.nextUrl.origin);
    const response = NextResponse.redirect(destination);

    response.cookies.set("halomail_access", session.accessToken, { ...cookieOptions, maxAge: 900 });
    response.cookies.set("halomail_refresh", session.refreshToken, { ...cookieOptions, maxAge: 30 * 86400 });
    response.cookies.delete("halomail_oauth_state");

    return response;
  } catch (err) {
    console.error("Google auth callback unexpected failure", err);
    loginUrl.searchParams.set("error", "auth_failed");
    return NextResponse.redirect(loginUrl);
  }
}

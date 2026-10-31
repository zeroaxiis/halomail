import { NextRequest, NextResponse } from "next/server";

export async function GET(request: NextRequest) {
  const target = new URL("/dashboard/meetings", process.env.PUBLIC_WEB_URL || request.url);
  const state = request.nextUrl.searchParams.get("state");
  const code = request.nextUrl.searchParams.get("code");
  const errorParam = request.nextUrl.searchParams.get("error");

  if (errorParam) {
    target.searchParams.set("calendar", "failed");
    target.searchParams.set("error", errorParam);
    return NextResponse.redirect(target);
  }

  if (!state || !code) {
    target.searchParams.set("calendar", "failed");
    target.searchParams.set("error", "Missing authorization code or state from Google");
    return NextResponse.redirect(target);
  }

  const isProd = process.env.NODE_ENV === "production";
  let backend = process.env.API_BASE_URL || process.env.NEXT_PUBLIC_API_URL;
  if (!backend || (isProd && backend.includes("localhost"))) {
    backend = isProd ? "https://halomailapi.zeroaxiis.tech" : "http://localhost:8080";
  }
  backend = backend.replace(/\/$/, "");

  let access = request.cookies.get("halomail_access")?.value;
  const refresh = request.cookies.get("halomail_refresh")?.value;
  let newAccessToken: string | null = null;

  // If access token is missing, attempt to refresh session
  if (!access && refresh) {
    try {
      const refreshRes = await fetch(`${backend}/halomail.identity.v1.AuthService/RefreshSession`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ refreshToken: refresh }),
        cache: "no-store",
      });
      if (refreshRes.ok) {
        const data = await refreshRes.json();
        if (data.session?.accessToken) {
          access = data.session.accessToken;
          newAccessToken = data.session.accessToken;
        }
      }
    } catch {
      // ignore
    }
  }

  if (!access) {
    target.searchParams.set("calendar", "failed");
    target.searchParams.set("error", "Your session expired. Please log in again.");
    return NextResponse.redirect(target);
  }

  let connected = false;
  let errorMessage = "";

  try {
    let response = await fetch(`${backend}/v1/calendar/callback`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${access}`,
      },
      body: JSON.stringify({ state, code }),
      signal: AbortSignal.timeout(30000),
      cache: "no-store",
    });

    // If 401 and refresh token exists, refresh access token and retry once
    if (response.status === 401 && refresh) {
      try {
        const refreshRes = await fetch(`${backend}/halomail.identity.v1.AuthService/RefreshSession`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ refreshToken: refresh }),
          cache: "no-store",
        });
        if (refreshRes.ok) {
          const data = await refreshRes.json();
          if (data.session?.accessToken) {
            access = data.session.accessToken;
            newAccessToken = data.session.accessToken;
            response = await fetch(`${backend}/v1/calendar/callback`, {
              method: "POST",
              headers: {
                "Content-Type": "application/json",
                Authorization: `Bearer ${access}`,
              },
              body: JSON.stringify({ state, code }),
              signal: AbortSignal.timeout(30000),
              cache: "no-store",
            });
          }
        }
      } catch {
        // ignore
      }
    }

    if (response.ok) {
      connected = true;
    } else {
      const body = await response.json().catch(() => ({}));
      errorMessage = body.message || `Backend error (status ${response.status})`;
      console.error("Calendar callback failed:", response.status, errorMessage);
    }
  } catch (err) {
    errorMessage = err instanceof Error ? err.message : "Network error contacting API";
    console.error("Calendar callback exception:", err);
  }

  target.searchParams.set("calendar", connected ? "connected" : "failed");
  if (!connected && errorMessage) {
    target.searchParams.set("error", errorMessage);
  }

  const response = NextResponse.redirect(target);
  response.cookies.delete("halomail_calendar_state");
  if (newAccessToken) {
    response.cookies.set("halomail_access", newAccessToken, {
      httpOnly: true,
      secure: isProd,
      sameSite: "lax",
      path: "/",
      maxAge: 900,
    });
  }
  return response;
}

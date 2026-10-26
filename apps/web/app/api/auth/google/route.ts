import { NextRequest, NextResponse } from "next/server";
import { randomBytes } from "node:crypto";

export const runtime = "nodejs";

function getRedirectURI(req: NextRequest): string {
  const custom = process.env.GOOGLE_AUTH_REDIRECT_URL;
  if (custom && custom.startsWith("http")) return custom;
  const origin = process.env.PUBLIC_WEB_URL || req.nextUrl.origin;
  return `${origin.replace(/\/$/, "")}/api/auth/google/callback`;
}

export async function GET(req: NextRequest) {
  const clientId = process.env.GOOGLE_CLIENT_ID;
  if (!clientId || clientId.trim() === "") {
    const loginUrl = new URL("/login", req.nextUrl.origin);
    loginUrl.searchParams.set("error", "google_not_configured");
    return NextResponse.redirect(loginUrl);
  }

  const returnTo = req.nextUrl.searchParams.get("redirect") || "/dashboard";
  const state = randomBytes(24).toString("hex");
  const redirectUri = getRedirectURI(req);

  const googleAuthUrl = new URL("https://accounts.google.com/o/oauth2/v2/auth");
  googleAuthUrl.searchParams.set("client_id", clientId.trim());
  googleAuthUrl.searchParams.set("redirect_uri", redirectUri);
  googleAuthUrl.searchParams.set("response_type", "code");
  googleAuthUrl.searchParams.set("scope", "openid email profile");
  googleAuthUrl.searchParams.set("state", state);
  googleAuthUrl.searchParams.set("access_type", "offline");
  googleAuthUrl.searchParams.set("prompt", "select_account");

  const response = NextResponse.redirect(googleAuthUrl.toString());
  const isProd = process.env.NODE_ENV === "production";
  response.cookies.set("halomail_oauth_state", JSON.stringify({ state, returnTo }), {
    httpOnly: true,
    secure: isProd,
    sameSite: "lax",
    path: "/",
    maxAge: 600, // 10 minutes
  });

  return response;
}

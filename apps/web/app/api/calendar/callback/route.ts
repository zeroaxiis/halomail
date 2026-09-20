import { NextRequest, NextResponse } from "next/server";

export async function GET(request: NextRequest) {
  const target = new URL("/dashboard/meetings", process.env.PUBLIC_WEB_URL || request.url);
  const state = request.nextUrl.searchParams.get("state");
  const code = request.nextUrl.searchParams.get("code");
  const access = request.cookies.get("halomail_access")?.value;
  let connected = false;
  if (state && code && access && state === request.cookies.get("halomail_calendar_state")?.value) {
    try {
      const backend = process.env.API_BASE_URL || process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
      const response = await fetch(`${backend}/v1/calendar/callback`, {
        method: "POST", headers: { "Content-Type": "application/json", Authorization: `Bearer ${access}` },
        body: JSON.stringify({ state, code }), signal: AbortSignal.timeout(30000), cache: "no-store",
      });
      connected = response.ok;
    } catch { connected = false; }
  }
  target.searchParams.set("calendar", connected ? "connected" : "failed");
  const response = NextResponse.redirect(target);
  response.cookies.delete("halomail_calendar_state");
  return response;
}

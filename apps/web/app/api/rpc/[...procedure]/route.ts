import { NextRequest, NextResponse } from "next/server";

const backend = process.env.API_BASE_URL || process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
const auth = "halomail.identity.v1.AuthService/";
const cookieOptions = { httpOnly: true, secure: process.env.NODE_ENV === "production", sameSite: "lax" as const, path: "/" };

export async function POST(request: NextRequest, context: { params: Promise<{ procedure: string[] }> }) {
  if (request.headers.get("origin") !== (process.env.PUBLIC_WEB_URL || request.nextUrl.origin)) {
    return NextResponse.json({ message: "Invalid request origin" }, { status: 403 });
  }
  const { procedure: parts } = await context.params;
  const procedure = parts.join("/");
  if (!/^halomail\.(identity|scheduling|contact|template)\.v1\.[A-Za-z]+Service\/[A-Za-z]+$/.test(procedure) && !["v1/usage", "v1/meetings/info"].includes(procedure)) {
    return NextResponse.json({ message: "Unknown endpoint" }, { status: 404 });
  }
  let raw = "";
  const reader = request.body?.getReader();
  const decoder = new TextDecoder();
  let size = 0;
  if (reader) {
    for (;;) {
      const { done, value } = await reader.read();
      if (done) break;
      size += value.byteLength;
      if (size > 65536) { await reader.cancel(); return NextResponse.json({ message: "Request too large" }, { status: 413 }); }
      raw += decoder.decode(value, { stream: true });
    }
    raw += decoder.decode();
  }
  let body: Record<string, unknown>;
  try { body = JSON.parse(raw || "{}"); }
  catch { return NextResponse.json({ message: "Invalid JSON" }, { status: 400 }); }
  if (!body || typeof body !== "object" || Array.isArray(body)) return NextResponse.json({ message: "Provide a JSON object" }, { status: 400 });
  const signingIn = [auth + "Login", auth + "Register"].includes(procedure);
  const refreshing = procedure === auth + "RefreshSession";
  const signingOut = procedure === auth + "Logout";
  if (refreshing || signingOut) body = { refreshToken: request.cookies.get("halomail_refresh")?.value || "" };
  try {
    const upstream = await fetch(`${backend}/${procedure}`, {
      method: "POST", cache: "no-store", signal: AbortSignal.timeout(30000),
      headers: {
        "Content-Type": "application/json",
        ...(request.cookies.get("halomail_access")?.value ? { Authorization: `Bearer ${request.cookies.get("halomail_access")!.value}` } : {}),
        ...(request.headers.get("x-halomail-key") ? { "X-HaloMail-Key": request.headers.get("x-halomail-key")! } : {}),
      },
      body: JSON.stringify(body),
    });
    const data = await upstream.json();
    const session = data.session;
    if (signingIn || refreshing) delete data.session;
    const response = NextResponse.json(data, { status: upstream.status, headers: { "Cache-Control": "no-store" } });
    if (upstream.ok && session && (signingIn || refreshing)) {
      response.cookies.set("halomail_access", session.accessToken, { ...cookieOptions, maxAge: 900 });
      response.cookies.set("halomail_refresh", session.refreshToken, { ...cookieOptions, maxAge: 30 * 86400 });
    }
    if (signingOut) {
      response.cookies.set("halomail_access", "", { ...cookieOptions, maxAge: 0 });
      response.cookies.set("halomail_refresh", "", { ...cookieOptions, maxAge: 0 });
    }
    if (upstream.ok && procedure.endsWith("CalendarService/StartConnect")) {
      const state = new URL(data.authorizationUrl).searchParams.get("state");
      if (state) response.cookies.set("halomail_calendar_state", state, { ...cookieOptions, maxAge: 600 });
    }
    return response;
  } catch {
    const response = NextResponse.json({ message: "The API is unavailable. Please try again." }, { status: 502 });
    if (signingOut) {
      response.cookies.set("halomail_access", "", { ...cookieOptions, maxAge: 0 });
      response.cookies.set("halomail_refresh", "", { ...cookieOptions, maxAge: 0 });
    }
    return response;
  }
}

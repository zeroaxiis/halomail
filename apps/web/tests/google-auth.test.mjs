import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { createRequire } from "node:module";
import test from "node:test";
import vm from "node:vm";

const require = createRequire(import.meta.url);
const ts = require("typescript");
const { NextRequest } = require("next/server");

const initSource = readFileSync(new URL("../app/api/auth/google/route.ts", import.meta.url), "utf8");
const callbackSource = readFileSync(new URL("../app/api/auth/google/callback/route.ts", import.meta.url), "utf8");

function compileRoute(source) {
  return ts.transpileModule(source, {
    compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 },
  }).outputText;
}

const compiledInit = compileRoute(initSource);
const compiledCallback = compileRoute(callbackSource);

test("Google Auth init redirects to Google accounts URL and sets state cookie", async () => {
  const exports = {};
  vm.runInNewContext(compiledInit, {
    exports,
    require,
    process: {
      env: {
        GOOGLE_CLIENT_ID: "mock-google-client-id.apps.googleusercontent.com",
        PUBLIC_WEB_URL: "https://halomail.zeroaxiis.tech",
        NODE_ENV: "production",
      },
    },
    URL,
  });

  const req = new NextRequest("https://halomail.zeroaxiis.tech/api/auth/google?redirect=/dashboard/settings");
  const res = await exports.GET(req);

  assert.equal(res.status, 307);
  const location = res.headers.get("location");
  assert.ok(location);
  const targetUrl = new URL(location);
  assert.equal(targetUrl.hostname, "accounts.google.com");
  assert.equal(targetUrl.searchParams.get("client_id"), "mock-google-client-id.apps.googleusercontent.com");
  assert.equal(targetUrl.searchParams.get("redirect_uri"), "https://halomail.zeroaxiis.tech/api/auth/google/callback");
  assert.equal(targetUrl.searchParams.get("response_type"), "code");
  assert.ok(targetUrl.searchParams.get("state"));

  const cookie = res.cookies.get("halomail_oauth_state");
  assert.ok(cookie);
  const stateData = JSON.parse(cookie.value);
  assert.equal(stateData.state, targetUrl.searchParams.get("state"));
  assert.equal(stateData.returnTo, "/dashboard/settings");
  assert.equal(cookie.httpOnly, true);
  assert.equal(cookie.secure, true);
});

test("Google Auth init redirects to error when client ID is missing", async () => {
  const exports = {};
  vm.runInNewContext(compiledInit, {
    exports,
    require,
    process: { env: { GOOGLE_CLIENT_ID: "" } },
    URL,
  });

  const req = new NextRequest("http://localhost:3000/api/auth/google");
  const res = await exports.GET(req);
  assert.equal(res.status, 307);
  const targetUrl = new URL(res.headers.get("location"));
  assert.equal(targetUrl.pathname, "/login");
  assert.equal(targetUrl.searchParams.get("error"), "google_not_configured");
});

test("Google Auth callback exchanges code, authenticates with backend and sets cookies", async () => {
  const fetchCalls = [];
  const exports = {};
  vm.runInNewContext(compiledCallback, {
    exports,
    require,
    process: {
      env: {
        GOOGLE_CLIENT_ID: "mock-client-id",
        GOOGLE_CLIENT_SECRET: "mock-client-secret",
        OTP_DELIVERY_SECRET: "s".repeat(32),
        PUBLIC_WEB_URL: "https://halomail.zeroaxiis.tech",
        API_BASE_URL: "https://halomailapi.zeroaxiis.tech",
        NODE_ENV: "production",
      },
    },
    URL,
    URLSearchParams,
    AbortSignal,
    fetch: async (url, init) => {
      fetchCalls.push({ url, init });
      if (url === "https://oauth2.googleapis.com/token") {
        return Response.json({ access_token: "google-access-token" });
      }
      if (url === "https://www.googleapis.com/oauth2/v3/userinfo") {
        return Response.json({
          sub: "google-12345",
          email: "user@example.com",
          email_verified: true,
          name: "Test User",
          picture: "https://example.com/avatar.jpg",
        });
      }
      if (url.endsWith("/v1/auth/google")) {
        return Response.json({
          user: { id: "usr_123", email: "user@example.com" },
          session: { accessToken: "hl_access_jwt", refreshToken: "hl_refresh_token" },
        });
      }
      throw new Error(`Unexpected URL: ${url}`);
    },
  });

  const stateObj = JSON.stringify({ state: "valid-state-123", returnTo: "/dashboard" });
  const req = new NextRequest("https://halomail.zeroaxiis.tech/api/auth/google/callback?code=mock-code&state=valid-state-123", {
    headers: {
      cookie: `halomail_oauth_state=${encodeURIComponent(stateObj)}`,
    },
  });

  const res = await exports.GET(req);
  assert.equal(res.status, 307);
  const targetUrl = new URL(res.headers.get("location"));
  assert.equal(targetUrl.pathname, "/dashboard");

  // Check auth cookies set
  assert.equal(res.cookies.get("halomail_access").value, "hl_access_jwt");
  assert.equal(res.cookies.get("halomail_refresh").value, "hl_refresh_token");

  // Check backend call verified secret
  const backendCall = fetchCalls.find(c => c.url.endsWith("/v1/auth/google"));
  assert.ok(backendCall);
  assert.equal(backendCall.init.headers["X-OTP-Delivery-Secret"], "s".repeat(32));
  const backendBody = JSON.parse(backendCall.init.body);
  assert.equal(backendBody.email, "user@example.com");
  assert.equal(backendBody.name, "Test User");
  assert.equal(backendBody.googleId, "google-12345");
});

test("Google Auth callback rejects invalid state or access cancellation", async () => {
  const exports = {};
  vm.runInNewContext(compiledCallback, {
    exports,
    require,
    process: { env: { GOOGLE_CLIENT_ID: "id", GOOGLE_CLIENT_SECRET: "secret" } },
    URL,
  });

  // Test cancellation
  const cancelReq = new NextRequest("http://localhost:3000/api/auth/google/callback?error=access_denied");
  const cancelRes = await exports.GET(cancelReq);
  assert.equal(new URL(cancelRes.headers.get("location")).searchParams.get("error"), "google_cancelled");

  // Test state mismatch
  const mismatchReq = new NextRequest("http://localhost:3000/api/auth/google/callback?code=abc&state=wrong-state", {
    headers: { cookie: "halomail_oauth_state=correct-state" },
  });
  const mismatchRes = await exports.GET(mismatchReq);
  assert.equal(new URL(mismatchRes.headers.get("location")).searchParams.get("error"), "invalid_state");
});

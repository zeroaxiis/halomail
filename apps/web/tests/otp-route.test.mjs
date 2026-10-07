import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { createRequire } from "node:module";
import test from "node:test";
import vm from "node:vm";

const require = createRequire(import.meta.url);
const ts = require("typescript");
const { NextRequest } = require("next/server");
const source = readFileSync(new URL("../app/api/rpc/[...procedure]/route.ts", import.meta.url), "utf8");
const compiled = ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } }).outputText;

function route({ upstream, sendMail = async () => {}, secret = "s".repeat(32) }) {
  const calls = [];
  const exports = {};
  vm.runInNewContext(compiled, {
    exports,
    require: name => name === "@/lib/otp-mail" ? { sendLoginCode: sendMail } : require(name),
    process: { env: { OTP_DELIVERY_SECRET: secret, PUBLIC_WEB_URL: "http://localhost:3000" } },
    TextDecoder, AbortSignal,
    fetch: async (url, init) => {
      calls.push({ url, init });
      return upstream(url, init);
    },
  });
  return {
    calls,
    post(procedure, body, origin = "http://localhost:3000") {
      const req = new NextRequest(`http://localhost:3000/api/rpc/${procedure}`, {
        method: "POST", headers: { origin, "Content-Type": "application/json" }, body: JSON.stringify(body),
      });
      return exports.POST(req, { params: Promise.resolve({ procedure: procedure.split("/") }) });
    },
  };
}

test("request emails OTP without exposing code or creating cookies", async () => {
  let mailed;
  const handler = route({
    upstream: async () => Response.json({ challengeId: "challenge", email: "user@example.com", code: "012345", expiresIn: 600 }),
    sendMail: async (email, code) => { mailed = { email, code }; },
  });
  const res = await handler.post("v1/auth/otp/request", { email: "user@example.com", password: "password" });
  assert.equal(res.status, 200);
  assert.deepEqual(await res.json(), { challengeId: "challenge", expiresIn: 600 });
  assert.deepEqual(mailed, { email: "user@example.com", code: "012345" });
  assert.equal(res.headers.get("set-cookie"), null);
  assert.equal(res.headers.get("cache-control"), "no-store");
  assert.equal(handler.calls[0].init.headers["X-OTP-Delivery-Secret"], "s".repeat(32));
});

test("delivery failure cancels the challenge and creates no session", async () => {
  const handler = route({
    upstream: async () => Response.json({ challengeId: "challenge", email: "user@example.com", code: "012345", expiresIn: 600 }),
    sendMail: async () => { throw new Error("Email provider unavailable"); },
  });
  const res = await handler.post("v1/auth/otp/request", { email: "user@example.com", password: "password" });
  assert.equal(res.status, 502);
  assert.equal(res.headers.get("set-cookie"), null);
  assert.equal(handler.calls.length, 2);
  assert.ok(handler.calls[1].url.endsWith("/v1/auth/otp/cancel"));
  assert.deepEqual(JSON.parse(handler.calls[1].init.body), { challengeId: "challenge" });
  assert.ok(!(await res.text()).includes("012345"));
});

test("successful verification stores tokens only in HttpOnly cookies", async () => {
  const handler = route({ upstream: async () => Response.json({ user: { id: "user" }, session: { accessToken: "access", refreshToken: "refresh" } }) });
  const res = await handler.post("v1/auth/otp/verify", { challengeId: "challenge", code: "012345" });
  assert.deepEqual(await res.json(), { user: { id: "user" } });
  assert.equal(res.cookies.get("halomail_access").value, "access");
  assert.equal(res.cookies.get("halomail_refresh").value, "refresh");
  assert.match(res.headers.get("set-cookie"), /HttpOnly/i);
  assert.equal(handler.calls[0].init.headers["X-OTP-Delivery-Secret"], undefined);
});

test("failed verification and legacy auth cannot create cookies", async () => {
  const handler = route({ upstream: async () => Response.json({ message: "invalid code" }, { status: 401 }) });
  const res = await handler.post("v1/auth/otp/verify", { challengeId: "challenge", code: "000000" });
  assert.equal(res.status, 401);
  assert.equal(res.headers.get("set-cookie"), null);
  const legacy = route({ upstream: async () => Response.json({ user: {}, session: { accessToken: "access", refreshToken: "refresh" } }) });
  for (const method of ["Login", "Register"]) {
    const response = await legacy.post(`halomail.identity.v1.AuthService/${method}`, {});
    assert.equal(response.headers.get("set-cookie"), null);
    assert.equal((await response.json()).session, undefined);
  }
});

test("rejects cross-origin requests, missing configuration and public cancellation", async () => {
  const handler = route({ secret: "", upstream: async () => { throw new Error("should not call upstream"); } });
  assert.equal((await handler.post("v1/auth/otp/request", {}, "https://untrusted.example")).status, 403);
  assert.equal((await handler.post("v1/auth/otp/request", {})).status, 503);
  assert.equal((await handler.post("v1/auth/otp/cancel", {})).status, 404);
  assert.equal(handler.calls.length, 0);
});

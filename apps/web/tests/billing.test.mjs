import assert from "node:assert/strict";
import { createHmac } from "node:crypto";
import { readFileSync } from "node:fs";
import { createRequire } from "node:module";
import test from "node:test";
import vm from "node:vm";
const require = createRequire(import.meta.url);
const ts = require("typescript");
const { NextRequest } = require("next/server");
const sign = (body, key = "webhook-secret") => createHmac("sha256", key).update(body).digest("hex");

function setup(overrides = {}) {
  const calls = [], uploads = [], orders = [];
  const env = { API_BASE_URL: "https://backend.test", PUBLIC_WEB_URL: "https://app.test",
    RAZORPAY_KEY_ID: "key", RAZORPAY_KEY_SECRET: "api-secret", RAZORPAY_WEBHOOK_SECRET: "webhook-secret",
    R2_ENDPOINT: "https://r2.test", R2_BUCKET_NAME: "bucket", R2_ACCESS_KEY_ID: "access", R2_SECRET_ACCESS_KEY: "secret", ...overrides.env };
  const payment = { id: "pay_123", order_id: "order_123", amount: 100, currency: "INR", status: "captured", created_at: 1791676800, notes: [], ...overrides.payment };
  const order = { id: "order_123", status: "paid", amount: 100, currency: "INR",
    notes: { user_id: "user", org_id: "org", tier: "pro", product: "halomail", billing_cycle: "yearly", name: "<script>bad</script>", ...overrides.notes }, ...overrides.order };
  class Razorpay {
    payments = { fetch: async () => payment };
    orders = { fetch: async () => order, create: async options => { orders.push(options); return { ...options, id: "order_new" }; } };
  }
  class S3Client { async send(command) { uploads.push(command.input); return overrides.storage ? overrides.storage(command) : {}; } }
  const cache = {};
  function load(path) {
    if (cache[path]) return cache[path];
    const source = readFileSync(new URL("../" + path + ".ts", import.meta.url), "utf8");
    const compiled = ts.transpileModule(source, { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022, esModuleInterop: true } }).outputText;
    const exports = {};
    cache[path] = exports;
    vm.runInNewContext(compiled, { exports, Buffer, AbortSignal, URL, console: { error() {} },
      process: { env }, require: name => {
        if (name === "./billing-mail") return { sendBillingReceipt: async () => {} };
        if (name.startsWith("@/")) return load(name.slice(2));
        if (name === "razorpay") return Razorpay;
        if (name === "@aws-sdk/client-s3") return { S3Client,
          PutObjectCommand: class { constructor(input) { this.input = input; } },
          GetObjectCommand: class { constructor(input) { this.input = input; } },
          ListObjectsV2Command: class { constructor(input) { this.input = input; } } };
        return require(name);
      },
      fetch: async (url, init) => {
        calls.push({ url, init });
        if (url.endsWith("GetCurrentUser")) return Response.json({ user: { id: "user", orgId: "org", email: "user@example.com", name: "User" } });
        return overrides.backend ? overrides.backend(url, init) : Response.json({ success: true, orgId: "org", userId: "user" });
      },
    });
    return exports;
  }
  function request(path, body, headers = {}) {
    return new NextRequest("https://app.test" + path, { method: "POST", headers: { origin: "https://app.test", cookie: "halomail_access=token", "Content-Type": "application/json", ...headers }, body: JSON.stringify(body) });
  }
  return { load, calls, uploads, orders, request, payment, order };
}

test("both webhook events resolve order notes, activate first and archive the same invoice", async () => {
  const app = setup();
  for (const event of ["payment.captured", "order.paid"]) {
    const payload = { event, payload: { payment: { entity: app.payment } } };
    const res = await app.load("app/api/razorpay/webhook/route").POST(app.request("/api/razorpay/webhook", payload, { "x-razorpay-signature": sign(JSON.stringify(payload)) }));
    assert.equal(res.status, 200);
  }
  assert.equal(app.uploads.length, 2);
  assert.equal(app.uploads[0].Key, "halomail/invice/org/user/invoice_pay_123.html");
  assert.equal(app.uploads[1].Key, app.uploads[0].Key);
  assert.ok(app.uploads[0].Body.includes("&lt;script&gt;"));
  const forwarded = JSON.parse(app.calls[0].init.body);
  assert.equal(forwarded.payload.order.entity.notes.billing_cycle, "yearly");
  assert.equal(app.calls[0].url, "https://backend.test/v1/billing/razorpay/webhook");
  assert.equal(app.calls[0].init.headers["X-Razorpay-Signature"], sign(app.calls[0].init.body));
});

test("missing signatures or webhook configuration fail closed", async () => {
  for (const env of [{}, { RAZORPAY_WEBHOOK_SECRET: "" }]) {
    const app = setup({ env });
    const res = await app.load("app/api/razorpay/webhook/route").POST(app.request("/api/razorpay/webhook", {}));
    assert.ok([401, 503].includes(res.status));
    assert.equal(app.calls.length, 0);
  }
});

test("backend rejection and R2 failure return retryable failures", async () => {
  for (const options of [
    { backend: () => Response.json({ error: "down" }, { status: 500 }) },
    { backend: () => Response.json({ ignored: true }) },
    { storage: () => { throw Error("R2 unavailable"); } },
  ]) {
    const app = setup(options);
    const payload = { event: "payment.captured", payload: { payment: { entity: app.payment } } };
    const res = await app.load("app/api/razorpay/webhook/route").POST(app.request("/api/razorpay/webhook", payload, { "x-razorpay-signature": sign(JSON.stringify(payload)) }));
    assert.equal(res.status, 503);
  }
});

test("recovery handles old orders missing org_id using backend-resolved ownership", async () => {
  const app = setup({ notes: { org_id: undefined, billing_cycle: undefined } });
  const res = await app.load("app/api/razorpay/reconcile/route").POST(app.request("/api/razorpay/reconcile", { paymentId: "pay_123" }));
  assert.equal(res.status, 200);
  assert.equal(app.uploads[0].Key, "halomail/invice/org/user/invoice_pay_123.html");
});

test("checkout verification rejects tampering, unpaid payments and account mismatches", async () => {
  for (const options of [
    { signature: "bad" }, { notes: { user_id: "other" } }, { payment: { status: "authorized" } },
    { payment: { amount: 1 } }, { payment: { currency: "USD" } }, { payment: { amount_refunded: 100 } },
    { order: { id: "order_other" } },
  ]) {
    const app = setup(options);
    const body = { razorpay_order_id: "order_123", razorpay_payment_id: "pay_123", razorpay_signature: options.signature || sign("order_123|pay_123", "api-secret") };
    const res = await app.load("app/api/razorpay/verify/route").POST(app.request("/api/razorpay/verify", body));
    assert.ok(res.status >= 400);
    assert.equal(app.uploads.length, 0);
    assert.ok(app.calls.every(call => call.url.endsWith("GetCurrentUser")));
  }
});

test("testing checkout costs 1 in every currency and cannot take identity or amount from the browser", async () => {
  const app = setup();
  for (const plan of ["Pro", "Business"]) for (const billingCycle of ["monthly", "yearly"]) for (const currency of ["INR", "USD", "EUR", "GBP", "CAD", "AUD", "SGD", "AED"]) {
    const res = await app.load("app/api/razorpay/route").POST(app.request("/api/razorpay", { plan, billingCycle, currency, amount: 99999, orgId: "attacker", userId: "attacker" }));
    assert.equal(res.status, 200);
    const order = app.orders.at(-1);
    assert.equal(order.amount, 100);
    assert.equal(order.notes.org_id, "org");
    assert.equal(order.notes.user_id, "user");
    assert.equal(order.notes.billing_cycle, billingCycle);
  }
});

test("invoice downloads reject unauthenticated, foreign and backup paths", async () => {
  for (const key of ["halomail/invice/other/user/invoice_pay_123.html", "halomail/invice/org/other/invoice_pay_123.html", "halomail/production/backup.dump", "invoices/org/../other/invoice_pay_123.html"]) {
    const app = setup();
    const req = new NextRequest("https://app.test/api/invoices/download?key=" + encodeURIComponent(key), { headers: { cookie: "halomail_access=token" } });
    assert.equal((await app.load("app/api/invoices/download/route").GET(req)).status, 404);
    assert.equal(app.uploads.length, 0);
  }
  const app = setup();
  assert.equal((await app.load("app/api/invoices/route").GET(new NextRequest("https://app.test/api/invoices?orgId=org"))).status, 401);
});

test("invoice history follows all pages and deduplicates legacy invoices", async () => {
  const app = setup({ storage: command => {
    const input = command.input;
    if (input.Prefix === "invoices/org/") return { Contents: [{ Key: "invoices/org/invoice_pay_123.html", LastModified: new Date() }] };
    if (input.ContinuationToken) return { Contents: [{ Key: "halomail/invice/org/user/invoice_pay_456.html", LastModified: new Date() }] };
    return { Contents: [{ Key: "halomail/invice/org/user/invoice_pay_123.html", LastModified: new Date() }], IsTruncated: true, NextContinuationToken: "next" };
  } });
  const req = new NextRequest("https://app.test/api/invoices?orgId=attacker", { headers: { cookie: "halomail_access=token" } });
  const res = await app.load("app/api/invoices/route").GET(req);
  assert.equal(res.status, 200);
  assert.equal((await res.json()).invoices.length, 2);
  assert.equal(app.uploads[1].ContinuationToken, "next");
});

test("successful checkout verification returns the activated plan and downloadable invoice", async () => {
  const app = setup();
  const res = await app.load("app/api/razorpay/verify/route").POST(app.request("/api/razorpay/verify", {
    razorpay_payment_id: "pay_123", razorpay_order_id: "order_123", razorpay_signature: sign("order_123|pay_123", "api-secret"),
  }));
  assert.equal(res.status, 200);
  assert.equal((await res.json()).plan, "pro");
  assert.equal(app.uploads[0].IfNoneMatch, "*");
});
test("an existing invoice is a successful idempotent retry", async () => {
  const app = setup({ storage: () => { throw { $metadata: { httpStatusCode: 412 } }; } });
  const res = await app.load("app/api/razorpay/reconcile/route").POST(app.request("/api/razorpay/reconcile", { paymentId: "pay_123" }));
  assert.equal(res.status, 200);
});
test("invoice bytes download as an attachment after ownership verification", async () => {
  const app = setup({ storage: () => ({ Body: { transformToByteArray: async () => new TextEncoder().encode("<html>receipt</html>") } }) });
  const req = new NextRequest("https://app.test/api/invoices/download?key=halomail/invice/org/user/invoice_pay_123.html", { headers: { cookie: "halomail_access=token" } });
  const res = await app.load("app/api/invoices/download/route").GET(req);
  assert.equal(res.status, 200);
  assert.equal(await res.text(), "<html>receipt</html>");
  assert.match(res.headers.get("Content-Disposition"), /attachment/);
});
test("order creation rejects anonymous or cross-origin requests and unsupported plans", async () => {
  for (const [headers, body, status] of [
    [{ cookie: "" }, { plan: "Pro", billingCycle: "monthly", currency: "INR" }, 401],
    [{ origin: "https://other.test" }, { plan: "Pro", billingCycle: "monthly", currency: "INR" }, 403],
    [{}, { plan: "Enterprise", billingCycle: "monthly", currency: "INR" }, 400],
  ]) {
    const app = setup();
    const res = await app.load("app/api/razorpay/route").POST(app.request("/api/razorpay", body, headers));
    assert.equal(res.status, status);
    assert.equal(app.orders.length, 0);
  }
});

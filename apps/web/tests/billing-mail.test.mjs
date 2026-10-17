import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { createRequire } from "node:module";
import test from "node:test";
import vm from "node:vm";

const require = createRequire(import.meta.url);
const ts = require("typescript");
const compiled = ts.transpileModule(readFileSync(new URL("../lib/billing-mail.ts", import.meta.url), "utf8"), {
  compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022, esModuleInterop: true },
}).outputText;
const details = { invoiceNumber: "HM-20261013-AbCd123", customerName: "Customer", total: "INR 299.00", dateLabel: "13 Oct 2026" };
const pdf = Buffer.from("%PDF-1.3 attachment bytes");

function mailer({ env = {}, response = () => Response.json({ id: "email-id" }) } = {}) {
  const requests = [], smtp = [];
  let closed = false;
  const exports = {};
  vm.runInNewContext(compiled, {
    exports, Buffer, AbortSignal, URL, console: { error() {} }, process: { env },
    require: name => {
      assert.equal(name, "nodemailer");
      return { createTransport: () => ({
        sendMail: async message => { smtp.push(message); return { messageId: "smtp-id" }; },
        close: () => { closed = true; },
      }) };
    },
    fetch: async (url, init) => { requests.push({ url, init }); return response(); },
  });
  return { ...exports, requests, smtp, isClosed: () => closed };
}

test("invoice email uses the dedicated key and exact sender with a separate downloadable PDF", async () => {
  const app = mailer({ env: { NTS_RESEND_API_KEY: " nts-key ", RESEND_API_KEY: "wrong-key", EMAIL_FROM: "wrong@example.com" } });
  const downloadURL = "https://halomail.zeroaxiis.tech/api/invoices/download?key=halomail%2Finvice%2Forg%2Fuser%2Finvoice_pay_123.pdf";
  assert.equal(await app.sendBillingReceipt("customer@example.com", "pay_123", pdf, { ...details, downloadURL }), "email-id");
  const { url, init } = app.requests[0];
  const body = JSON.parse(init.body);
  assert.equal(url, "https://api.resend.com/emails");
  assert.equal(init.headers.Authorization, "Bearer nts-key");
  assert.equal(init.headers["Idempotency-Key"], "invoice-v2/pay_123");
  assert.equal(body.from, "noreply@nts.zeroaxiis.tech");
  assert.deepEqual(body.to, ["customer@example.com"]);
  assert.equal(body.attachments.length, 1);
  assert.equal(body.attachments[0].content_type, "application/pdf");
  assert.equal(body.attachments[0].filename, "HaloMail_Invoice_pay_123.pdf");
  assert.deepEqual(Buffer.from(body.attachments[0].content, "base64"), pdf);
  assert.match(body.html, /PDF|pdf/);
  assert.match(body.html, /Download invoice/);
  assert.ok(body.html.includes(downloadURL));
  assert.match(body.text, /Amount paid: INR 299.00/);
  assert.doesNotMatch(body.html, /GSTIN|DIGITALLY|UNIT PRICE|display:\s*(flex|grid)|<style|data:application\/pdf/);
  assert.equal(app.smtp.length, 0);
});

test("email escapes customer content and rejects executable download URLs", () => {
  const app = mailer();
  const body = app.renderBillingEmail({ ...details, customerName: '<img src=x onerror="alert(1)">' });
  assert.ok(body.html.includes("&lt;img"));
  assert.doesNotMatch(body.html, /<img/);
  assert.throws(() => app.renderBillingEmail({ ...details, downloadURL: "javascript:alert(1)" }), /Invalid invoice/);
});

test("sample preview has no payment claim or broken download link", () => {
  const body = mailer().renderBillingEmail({ ...details, sample: true });
  assert.match(body.subject, /Preview/);
  assert.match(body.html, /No payment has been recorded/);
  assert.doesNotMatch(body.html, /Amount paid|href=|Thank you for your payment/);
});

test("SMTP also delivers the compact email and an attachment with the fixed sender", async () => {
  const app = mailer({ env: { SMTP_HOST: "localhost", EMAIL_FROM: "wrong@example.com" } });
  assert.equal(await app.sendBillingReceipt("customer@example.com", "pay_123", pdf, details), "smtp-id");
  assert.equal(app.requests.length, 0);
  assert.equal(app.smtp[0].from, "noreply@nts.zeroaxiis.tech");
  assert.equal(app.smtp[0].attachments[0].contentDisposition, "attachment");
  assert.deepEqual(app.smtp[0].attachments[0].content, pdf);
  assert.match(app.smtp[0].text, /INR 299.00/);
  assert.ok(app.isClosed());
});

test("shared key cannot send invoices, and API failures do not create duplicate SMTP sends", async () => {
  const generic = mailer({ env: { RESEND_API_KEY: "wrong-key" } });
  assert.equal(await generic.sendBillingReceipt("customer@example.com", "pay_123", pdf, details), undefined);
  assert.equal(generic.requests.length, 0);
  for (const response of [
    () => Response.json({ error: "rejected" }, { status: 403 }),
    () => Response.json({}),
    () => { throw new Error("timeout"); },
  ]) {
    const app = mailer({ env: { NTS_RESEND_API_KEY: "nts-key", SMTP_HOST: "localhost" }, response });
    assert.equal(await app.sendBillingReceipt("customer@example.com", "pay_123", pdf, details), undefined);
    assert.equal(app.requests.length, 1);
    assert.equal(app.smtp.length, 0);
  }
});

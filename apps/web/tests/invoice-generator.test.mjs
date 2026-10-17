import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { createRequire } from "node:module";
import test from "node:test";
import vm from "node:vm";

const require = createRequire(new URL("../lib/invoice-generator.ts", import.meta.url));
const ts = require("typescript");
const compiled = ts.transpileModule(readFileSync(new URL("../lib/invoice-generator.ts", import.meta.url), "utf8"), {
  compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022, esModuleInterop: true },
}).outputText;
const exports = {};
vm.runInNewContext(compiled, { exports, require, Buffer, Intl });
const { getInvoiceDetails, generateInvoiceHTML, generateInvoicePDF } = exports;
const invoice = { paymentId: "pay_AbCd123", orderId: "order_123", amountMinor: 29900, currency: "INR",
  date: "2026-10-13T00:00:00+05:30", customerName: "Customer", customerEmail: "customer@example.com", tier: "pro", billingCycle: "monthly" };

test("invoice totals always convert minor units without size-based guessing", () => {
  for (const [amountMinor, total] of [[1, "INR 0.01"], [100, "INR 1.00"], [299, "INR 2.99"],
    [999, "INR 9.99"], [1000, "INR 10.00"], [29900, "INR 299.00"], [100000, "INR 1,000.00"]]) {
    assert.equal(getInvoiceDetails({ ...invoice, amountMinor }).total, total);
  }
  assert.equal(getInvoiceDetails({ ...invoice, currency: "USD" }).total, "USD 299.00");
  for (const amountMinor of [-1, 1.5, NaN, Infinity]) assert.throws(() => getInvoiceDetails({ ...invoice, amountMinor }), /minor units/);
});

test("invoice identity keeps the full case-sensitive payment reference and Indian invoice date", () => {
  assert.equal(getInvoiceDetails(invoice).invoiceNumber, "HM-20261013-AbCd123");
  assert.equal(getInvoiceDetails(invoice).dateLabel, "13 Oct 2026");
  assert.notEqual(getInvoiceDetails(invoice).invoiceNumber, getInvoiceDetails({ ...invoice, paymentId: "pay_abcd123" }).invoiceNumber);
});

test("invoice shows the approved seller and brand without invented tax or statutory claims", () => {
  const html = generateInvoiceHTML({ ...invoice, customerName: "<script>bad</script>" });
  assert.match(html, /Gorakhpur - 273004/);
  assert.match(html, /alt="ZeroAxiis"/);
  assert.match(html, /&lt;script&gt;bad&lt;\/script&gt;/);
  assert.match(html, /INR 299.00/);
  assert.match(html, /Balance due/);
  assert.doesNotMatch(html, /GSTIN|IGST|GST|PAN:|DIGITALLY|Private Limited|Bengaluru|renew automatically|<script>/);
});

test("preview is clearly identified while real invoices use payment receipt wording", () => {
  const html = generateInvoiceHTML({ ...invoice, sample: true });
  assert.match(html, /SAMPLE \/ FOR PREVIEW/);
  assert.match(html, /No payment has been recorded/);
  assert.doesNotMatch(html, /PAID|Amount paid|Balance due/);
});

test("generator produces a PDF attachment for real and preview invoices", async () => {
  for (const sample of [true, false]) {
    const pdf = await generateInvoicePDF({ ...invoice, sample });
    assert.equal(pdf.subarray(0, 5).toString(), "%PDF-");
    assert.ok(pdf.length > 1000);
  }
});

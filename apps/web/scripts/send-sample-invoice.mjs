import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { getInvoiceDetails } from "../lib/invoice-generator.ts";
import { BILLING_EMAIL_FROM, sendBillingReceipt } from "../lib/billing-mail.ts";
import { sampleInvoice } from "./sample-invoice-data.mjs";

process.loadEnvFile(new URL("../../../.env", import.meta.url));
if (!process.env.NTS_RESEND_API_KEY?.trim()) throw new Error("NTS_RESEND_API_KEY is required");
const recipient = process.argv[2] || sampleInvoice.customerEmail;
assert.equal(recipient, sampleInvoice.customerEmail, "The preview must match the intended invoice recipient");
const pdf = readFileSync(new URL("../../../output/pdf/halomail-invoice-preview.pdf", import.meta.url));
assert.equal(pdf.subarray(0, 5).toString(), "%PDF-");
const id = await sendBillingReceipt(recipient, sampleInvoice.paymentId, pdf, getInvoiceDetails(sampleInvoice));
if (!id) throw new Error("Resend did not accept the invoice email");
console.log(JSON.stringify({ accepted: true, id, from: BILLING_EMAIL_FROM, to: recipient, attachmentBytes: pdf.length }));

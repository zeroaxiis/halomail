import { generateInvoicePDF, generateInvoiceHTML, getInvoiceDetails } from "../lib/invoice-generator.ts";
import { renderBillingEmail } from "../lib/billing-mail.ts";
import { sampleInvoice } from "./sample-invoice-data.mjs";
import { mkdirSync, writeFileSync } from "node:fs";

const pdf = await generateInvoicePDF(sampleInvoice);
const html = generateInvoiceHTML(sampleInvoice);
const email = renderBillingEmail(getInvoiceDetails(sampleInvoice));
const output = new URL("../../../output/pdf/", import.meta.url);
mkdirSync(output, { recursive: true });
writeFileSync(new URL("halomail-invoice-preview.pdf", output), pdf);
writeFileSync(new URL("halomail-invoice-preview.html", output), html);
writeFileSync(new URL("halomail-invoice-email.html", output), email.html);
writeFileSync(new URL("../sample_invoice_299.pdf", import.meta.url), pdf);
writeFileSync(new URL("../sample_invoice_299.html", import.meta.url), html);
console.log("Created PDF, printable invoice, and email previews in output/pdf");
console.log("PDF bytes:", pdf.length);

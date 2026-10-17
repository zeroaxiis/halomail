import PDFDocument from "pdfkit";
import logo from "./email-logo.json" with { type: "json" };

export interface InvoiceData {
  paymentId: string;
  orderId: string;
  /** The captured Razorpay amount, in the currency's smallest unit. */
  amountMinor: number;
  currency: string;
  date: string | Date;
  customerName: string;
  customerEmail: string;
  tier?: string;
  billingCycle?: string;
  sample?: boolean;
}

export function escapeInvoiceHTML(value: unknown): string {
  return String(value ?? "").replace(/[&<>"']/g, c => ({
    "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;",
  }[c]!));
}

export function getInvoiceDetails(data: InvoiceData) {
  if (!Number.isSafeInteger(data.amountMinor) || data.amountMinor < 0) {
    throw new Error("Invoice amount must be a non-negative integer in minor units");
  }
  const currency = data.currency.toUpperCase();
  const digits = new Intl.NumberFormat("en", { style: "currency", currency }).resolvedOptions().maximumFractionDigits ?? 2;
  const format = (minor: number) => `${currency} ${new Intl.NumberFormat(currency === "INR" ? "en-IN" : "en-US", {
    minimumFractionDigits: digits, maximumFractionDigits: digits,
  }).format(minor / 10 ** digits)}`;
  const date = new Date(data.date);
  if (Number.isNaN(date.getTime())) throw new Error("Invalid invoice date");
  if (!data.paymentId || !data.orderId) throw new Error("Invoice payment and order references are required");
  const dateParts = new Intl.DateTimeFormat("en-GB", {
    timeZone: "Asia/Kolkata", year: "numeric", month: "2-digit", day: "2-digit",
  }).formatToParts(date);
  const part = (type: string) => dateParts.find(p => p.type === type)!.value;
  return {
    invoiceNumber: `HM-${part("year")}${part("month")}${part("day")}-${data.paymentId.replace(/^pay_/, "")}`,
    dateLabel: new Intl.DateTimeFormat("en-GB", {
      timeZone: "Asia/Kolkata", day: "2-digit", month: "short", year: "numeric",
    }).format(date),
    customerName: data.customerName || "Customer",
    customerEmail: data.customerEmail,
    sellerName: "HaloMail",
    sellerAddress: "Gorakhpur - 273004",
    total: format(data.amountMinor),
    zero: format(0),
    currency,
    plan: `HaloMail ${data.tier === "business" ? "Business" : "Pro"}`,
    term: data.billingCycle === "yearly" ? "Yearly subscription" : "Monthly subscription",
    paymentId: data.paymentId,
    orderId: data.orderId,
    sample: Boolean(data.sample),
  };
}

/** A printable invoice with the same data as the archived HTML version. */
export async function generateInvoicePDF(data: InvoiceData): Promise<Buffer> {
  const invoice = getInvoiceDetails(data);
  return new Promise((resolve, reject) => {
    const doc = new PDFDocument({ size: "A4", margin: 42, bufferPages: true, info: {
      Title: `${invoice.sample ? "Sample " : ""}Invoice ${invoice.invoiceNumber}`,
      Author: "HaloMail", Subject: "HaloMail subscription invoice",
    } });
    const chunks: Buffer[] = [];
    doc.on("data", (chunk: Buffer) => chunks.push(chunk));
    doc.on("end", () => resolve(Buffer.concat(chunks)));
    doc.on("error", reject);

    const left = 42, width = 511.28, right = left + width;
    const ink = "#202124", muted = "#62666b", line = "#ced1d4";
    const text = (value: string, x: number, y: number, w: number, size = 9, bold = false, color = ink, align: "left" | "right" = "left") => {
      doc.font(bold ? "Helvetica-Bold" : "Helvetica").fontSize(size).fillColor(color)
        .text(value, x, y, { width: w, align, lineGap: 3 });
      return doc.y;
    };
    const rule = (y: number, x = left, w = width, color = line) => {
      doc.moveTo(x, y).lineTo(x + w, y).lineWidth(0.65).strokeColor(color).stroke();
    };
    const label = (value: string, x: number, y: number, w: number) => text(value, x, y, w, 8, true, muted);

    doc.image(Buffer.from(logo.content, "base64"), left, 40, { fit: [36, 36] });
    text("HaloMail", left + 46, 40, 230, 23, true);
    text("by ZeroAxiis", left + 47, 70, 200, 8.5, false, muted);
    text("INVOICE", 355, 42, right - 355, 23, true, ink, "right");
    text(invoice.sample ? "SAMPLE / FOR PREVIEW" : "PAID", 355, 74, right - 355, 8, true, muted, "right");
    rule(103, left, width, ink);

    label("ISSUED BY", left, 124, 215);
    let sellerBottom = text(invoice.sellerName, left, 143, 215, 11, true);
    sellerBottom = text(invoice.sellerAddress, left, sellerBottom + 5, 215, 9);
    text("halomail.zeroaxiis.tech", left, sellerBottom + 6, 215, 8.5, false, muted);

    const metaX = 312, metaWidth = right - metaX;
    label("INVOICE NUMBER", metaX, 124, metaWidth);
    let metaY = text(invoice.invoiceNumber, metaX, 143, metaWidth, 9, true);
    label("INVOICE DATE", metaX, metaY + 13, metaWidth);
    metaY = text(invoice.dateLabel, metaX, metaY + 29, metaWidth, 9);

    const partyY = Math.max(215, metaY + 22);
    rule(partyY - 12);
    label("BILL TO", left, partyY, width);
    const nameBottom = text(invoice.customerName, left, partyY + 19, width, 11, true);
    const customerBottom = invoice.customerEmail
      ? text(invoice.customerEmail, left, nameBottom + 5, width, 9, false, muted) : nameBottom;

    const tableY = customerBottom + 28;
    const cols = [left, left + 32, left + 272, left + 314, left + 412, right];
    doc.rect(left, tableY, width, 29).fill("#f1f2f3");
    rule(tableY, left, width, ink);
    text("#", cols[0] + 9, tableY + 9, 18, 8, true);
    text("DESCRIPTION", cols[1] + 7, tableY + 9, 223, 8, true);
    text("QTY", cols[2], tableY + 9, 35, 8, true, ink, "right");
    text("UNIT PRICE", cols[3], tableY + 9, 88, 8, true, ink, "right");
    text("AMOUNT", cols[4], tableY + 9, 91, 8, true, ink, "right");
    rule(tableY + 29);
    const rowY = tableY + 47;
    text("1", cols[0] + 9, rowY, 18, 9);
    text(invoice.plan, cols[1] + 7, rowY, 223, 10, true);
    text(invoice.term, cols[1] + 7, rowY + 19, 223, 9, false, muted);
    text("1", cols[2], rowY, 35, 9, false, ink, "right");
    text(invoice.total, cols[3] - 6, rowY, 94, 9, false, ink, "right");
    text(invoice.total, cols[4] - 6, rowY, 97, 9, true, ink, "right");
    const tableEnd = tableY + 97;
    rule(tableEnd);

    const totalY = tableEnd + 22;
    const summaryX = 326, summaryW = right - summaryX;
    text("Subtotal", summaryX, totalY, 95, 9);
    text(invoice.total, summaryX + 90, totalY, summaryW - 90, 9, false, ink, "right");
    rule(totalY + 23, summaryX, summaryW);
    text("Total", summaryX, totalY + 37, 80, 12, true);
    text(invoice.total, summaryX + 75, totalY + 35, summaryW - 75, 16, true, ink, "right");
    let summaryEnd = totalY + 69;
    if (!invoice.sample) {
      text("Amount paid", summaryX, totalY + 71, 110, 9);
      text(invoice.total, summaryX + 95, totalY + 71, summaryW - 95, 9, false, ink, "right");
      text("Balance due", summaryX, totalY + 91, 110, 9);
      text(invoice.zero, summaryX + 95, totalY + 91, summaryW - 95, 9, true, ink, "right");
      summaryEnd = totalY + 112;
    }

    const paymentY = Math.max(summaryEnd + 34, 558);
    rule(paymentY);
    label("PAYMENT DETAILS", left, paymentY + 19, width);
    const refRow = (name: string, value: string, y: number) => {
      text(name, left, y, 108, 8.5, false, muted);
      return text(value, left + 114, y, width - 114, 8.5);
    };
    let refY = refRow("Payment reference", invoice.paymentId, paymentY + 40);
    refY = refRow("Order reference", invoice.orderId, refY + 9);
    refRow("Payment provider", invoice.sample ? "Razorpay (sample)" : "Razorpay", refY + 9);

    const note = invoice.sample
      ? "Sample invoice for layout review. No payment has been recorded."
      : "Payment received. Keep this invoice for your records.";
    text(note, left, 734, width, 8.5, false, muted);
    rule(777);
    text("HaloMail by ZeroAxiis", left, 790, 250, 8, false, muted);
    text("1 / 1", right - 55, 790, 55, 8, false, muted, "right");
    doc.end();
  });
}

/** Archived browser/print copy. This document is never used as an email body. */
export function generateInvoiceHTML(data: InvoiceData): string {
  const i = getInvoiceDetails(data);
  const e = escapeInvoiceHTML;
  return `<!doctype html>
<html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>Invoice ${e(i.invoiceNumber)}</title>
<style>
*{box-sizing:border-box}body{margin:0;padding:36px 16px;background:#f1f2f3;color:#202124;font:13px/1.6 Arial,Helvetica,sans-serif}
.sheet{max-width:794px;margin:auto;background:#fff;padding:52px 56px;min-height:1060px}.header{display:flex;justify-content:space-between;align-items:start;border-bottom:1px solid #202124;padding-bottom:26px}.brand{display:flex;gap:14px;align-items:center}.brand img{width:46px;height:46px;object-fit:contain}.brand strong{display:block;font-size:28px;line-height:1.2}.muted,.label{color:#62666b}.label{font-size:10px;font-weight:700;letter-spacing:1px}h1{font-size:28px;line-height:1.2;margin:0 0 9px}.heading{text-align:right}.parties{display:grid;grid-template-columns:1fr 1fr;gap:36px;padding:27px 0}.parties p{margin:6px 0 0}.number{overflow-wrap:anywhere}.buyer{padding:20px 0 27px;border-top:1px solid #ced1d4}.buyer p{margin:7px 0 0;overflow-wrap:anywhere}table{width:100%;border-collapse:collapse;font-size:12px}th{background:#f1f2f3;text-align:left;font-size:10px;border-top:1px solid #202124;border-bottom:1px solid #ced1d4;padding:12px 9px}td{padding:19px 9px;border-bottom:1px solid #ced1d4;vertical-align:top}.num{text-align:right;white-space:nowrap}.totals{width:300px;margin:20px 0 40px auto}.totals td{padding:8px 0;border:0}.totals .total td{border-top:1px solid #ced1d4;padding:18px 0;font-weight:700;font-size:20px}.payment{border-top:1px solid #ced1d4;padding-top:20px}.payment dl{display:grid;grid-template-columns:130px 1fr;gap:7px;margin:15px 0 0;font-size:12px}.payment dt{color:#62666b}.payment dd{margin:0;overflow-wrap:anywhere}.note{margin-top:58px;font-size:12px;color:#62666b}footer{border-top:1px solid #ced1d4;padding-top:13px;margin-top:28px;font-size:11px;color:#62666b;display:flex;justify-content:space-between}
@page{size:A4;margin:0}@media print{body{padding:0;background:white}.sheet{width:210mm;max-width:none;min-height:297mm;padding:14mm}}@media(max-width:600px){.sheet{padding:28px 20px}.parties{gap:20px}.brand strong{font-size:22px}.brand img{width:32px;height:32px}.brand{gap:8px}h1{font-size:22px}.totals{width:260px}th,td{padding-left:5px;padding-right:5px;font-size:10px}}
</style></head><body><main class="sheet">
<header class="header"><div class="brand"><img src="data:image/png;base64,${logo.content}" alt="ZeroAxiis"><div><strong>HaloMail</strong><span class="muted">by ZeroAxiis</span></div></div><div class="heading"><h1>INVOICE</h1><span class="label">${i.sample ? "SAMPLE / FOR PREVIEW" : "PAID"}</span></div></header>
<section class="parties"><div><div class="label">ISSUED BY</div><p><strong>${e(i.sellerName)}</strong><br>${e(i.sellerAddress)}<br><span class="muted">halomail.zeroaxiis.tech</span></p></div><div><div class="label">INVOICE NUMBER</div><p class="number"><strong>${e(i.invoiceNumber)}</strong></p><div class="label" style="margin-top:14px">INVOICE DATE</div><p>${e(i.dateLabel)}</p></div></section>
<section class="buyer"><div class="label">BILL TO</div><p><strong>${e(i.customerName)}</strong><br><span class="muted">${e(i.customerEmail)}</span></p></section>
<table aria-label="Invoice items"><thead><tr><th>#</th><th>DESCRIPTION</th><th class="num">QTY</th><th class="num">UNIT PRICE</th><th class="num">AMOUNT</th></tr></thead><tbody><tr><td>1</td><td><strong>${e(i.plan)}</strong><br><span class="muted">${e(i.term)}</span></td><td class="num">1</td><td class="num">${e(i.total)}</td><td class="num"><strong>${e(i.total)}</strong></td></tr></tbody></table>
<table class="totals" aria-label="Invoice totals"><tr><td>Subtotal</td><td class="num">${e(i.total)}</td></tr><tr class="total"><td>Total</td><td class="num">${e(i.total)}</td></tr>${i.sample ? "" : `<tr><td>Amount paid</td><td class="num">${e(i.total)}</td></tr><tr><td>Balance due</td><td class="num"><strong>${e(i.zero)}</strong></td></tr>`}</table>
<section class="payment"><div class="label">PAYMENT DETAILS</div><dl><dt>Payment reference</dt><dd>${e(i.paymentId)}</dd><dt>Order reference</dt><dd>${e(i.orderId)}</dd><dt>Payment provider</dt><dd>Razorpay${i.sample ? " (sample)" : ""}</dd></dl></section>
<p class="note">${i.sample ? "Sample invoice for layout review. No payment has been recorded." : "Payment received. Keep this invoice for your records."}</p><footer><span>HaloMail by ZeroAxiis</span><span>1 / 1</span></footer>
</main></body></html>`;
}

import { sendBillingReceipt } from "./billing-mail";
import { createHmac, timingSafeEqual } from "node:crypto";
import { NextRequest, NextResponse } from "next/server";
import Razorpay from "razorpay";
import { S3Client, PutObjectCommand } from "@aws-sdk/client-s3";

export class BillingError extends Error {
  constructor(message: string, public status = 400) { super(message); }
}
export function billingError(error: unknown) {
  if (error instanceof BillingError) return NextResponse.json({ error: error.message }, { status: error.status });
  console.error("Billing operation failed", error);
  return NextResponse.json({ error: "Billing could not be completed. Please retry; do not pay again." }, { status: 503 });
}
export function backendURL() {
  return (process.env.API_BASE_URL || process.env.NEXT_PUBLIC_API_URL || process.env.PUBLIC_API_URL || "http://localhost:8080").replace(/\/$/, "");
}
export type BillingUser = { id: string; orgId: string; name: string; email: string };
export async function billingUser(req: NextRequest): Promise<BillingUser> {
  if (req.method === "POST" && req.headers.get("origin") !== (process.env.PUBLIC_WEB_URL || req.nextUrl.origin)) throw new BillingError("Invalid request origin", 403);
  const token = req.cookies.get("halomail_access")?.value;
  if (!token) throw new BillingError("Please sign in again.", 401);
  const response = await fetch(`${backendURL()}/halomail.identity.v1.AuthService/GetCurrentUser`, {
    method: "POST", cache: "no-store", signal: AbortSignal.timeout(15000),
    headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` }, body: "{}",
  });
  if (!response.ok) throw new BillingError("Could not verify your session. Please sign in again.", response.status === 401 ? 401 : 503);
  const { user } = await response.json();
  const orgId = user?.orgId || user?.org_id;
  if (!user?.id || !orgId) throw new BillingError("Account information is unavailable.", 503);
  return { id: user.id, orgId, name: user.name || "Customer", email: user.email || "" };
}
export function razorpayKeys() {
  const key_id = process.env.RAZORPAY_KEY_ID || process.env.RAZERPAY_LIVE_API_KEY;
  const key_secret = process.env.RAZORPAY_KEY_SECRET || process.env.RAZERPAY_LIVE_KEY_SECRET;
  if (!key_id || !key_secret) throw new BillingError("Payments are not configured.", 503);
  return { key_id, key_secret };
}
export function razorpay() { return new Razorpay(razorpayKeys()); }
export function validSignature(body: string, signature: string | null, secret: string | undefined) {
  if (!secret || !signature || !/^[a-f0-9]{64}$/i.test(signature)) return false;
  return timingSafeEqual(createHmac("sha256", secret).update(body).digest(), Buffer.from(signature, "hex"));
}
export function invoiceStorage() {
  const { R2_ENDPOINT, R2_ACCESS_KEY_ID, R2_SECRET_ACCESS_KEY, R2_BUCKET_NAME } = process.env;
  if (!R2_ENDPOINT || !R2_ACCESS_KEY_ID || !R2_SECRET_ACCESS_KEY || !R2_BUCKET_NAME) throw new BillingError("Invoice storage is not configured.", 503);
  return { client: new S3Client({ region: process.env.R2_REGION || "auto", endpoint: R2_ENDPOINT,
    credentials: { accessKeyId: R2_ACCESS_KEY_ID, secretAccessKey: R2_SECRET_ACCESS_KEY } }), bucket: R2_BUCKET_NAME };
}
export function invoicePrefix() { return (process.env.R2_INVOICE_PREFIX || "halomail/invice").replace(/^\/+|\/+$/g, ""); }
export function invoicePrefixes(user: BillingUser) { return [`${invoicePrefix()}/${user.orgId}/${user.id}/`, `invoices/${user.orgId}/`]; }
export function ownsInvoice(key: string, user: BillingUser) {
  return invoicePrefixes(user).some(prefix => key.startsWith(prefix) && /^invoice_pay_[A-Za-z0-9]+\.html$/.test(key.slice(prefix.length)));
}
function escapeHTML(value: unknown) {
  return String(value ?? "").replace(/[&<>"']/g, c => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c]!));
}

// Fetch the payment AND order: payment notes need not contain the order notes.
export async function fulfillPayment(paymentId: string, user?: BillingUser, expectedOrderId?: string) {
  if (!/^pay_[A-Za-z0-9]+$/.test(paymentId)) throw new BillingError("Invalid payment ID.");
  const provider = razorpay();
  const payment = await provider.payments.fetch(paymentId);
  if (!payment.order_id) throw new BillingError("Payment has no HaloMail order.");
  const order = await provider.orders.fetch(payment.order_id);
  const notes = (order.notes || {}) as Record<string, string>;
  if (notes.product !== "halomail" || !notes.user_id || !["pro", "business"].includes(notes.tier)) {
    throw new BillingError("This order is missing valid HaloMail account details. Contact support with the payment ID.", 422);
  }
  if (user && ((notes.org_id && notes.org_id !== user.orgId) || notes.user_id !== user.id)) throw new BillingError("This payment belongs to another account.", 403);
  if (expectedOrderId && order.id !== expectedOrderId) throw new BillingError("Payment does not match the checkout order.", 403);
  if (payment.status !== "captured" || order.status !== "paid") throw new BillingError("Payment is awaiting capture. Retry after it is captured; do not pay again.", 409);
  if (Number(payment.amount) !== Number(order.amount) || payment.currency !== order.currency || Number(payment.amount_refunded) > 0) throw new BillingError("Payment amount or refund status does not match the order.", 409);
  const cycle = notes.billing_cycle || "monthly";
  if (!["monthly", "yearly"].includes(cycle)) throw new BillingError("Invalid billing cycle.");
  const secret = process.env.RAZORPAY_WEBHOOK_SECRET;
  if (!secret) throw new BillingError("Payment verification is not configured.", 503);
  const body = JSON.stringify({ event: "order.paid", payload: { order: { entity: { ...order, notes: { ...notes, billing_cycle: cycle } } }, payment: { entity: payment } } });
  const result = await fetch(`${backendURL()}/v1/billing/razorpay/webhook`, {
    method: "POST", cache: "no-store", signal: AbortSignal.timeout(20000),
    headers: { "Content-Type": "application/json", "X-Razorpay-Signature": createHmac("sha256", secret).update(body).digest("hex") }, body,
  });
  const activated = result.ok ? await result.json() : null;
  if (!activated?.success || !activated.orgId || activated.userId !== notes.user_id) throw new BillingError("Payment received, but plan activation is pending. Retry; do not pay again.", 503);
  notes.org_id = activated.orgId;
  const date = new Date(payment.created_at * 1000).toISOString();
  const html = `<!doctype html><html><head><meta charset="utf-8"><title>HaloMail Invoice ${escapeHTML(payment.id)}</title></head>
<body style="font-family:Arial,sans-serif;max-width:720px;margin:40px auto;padding:24px;color:#111827">
<h1>HaloMail — Payment Receipt / Invoice</h1><p>Paid on ${escapeHTML(date)}</p>
<p>Billed to: ${escapeHTML(notes.name)} (${escapeHTML(notes.email || payment.email)})</p>
<table cellpadding="10"><tr><th>Plan</th><td>${escapeHTML(notes.tier)} (${escapeHTML(cycle)})</td></tr>
<tr><th>Amount paid</th><td>${escapeHTML(payment.currency)} ${(Number(payment.amount) / 100).toFixed(2)}</td></tr>
<tr><th>Payment ID</th><td>${escapeHTML(payment.id)}</td></tr><tr><th>Order ID</th><td>${escapeHTML(order.id)}</td></tr>
<tr><th>Account ID</th><td>${escapeHTML(notes.org_id)}</td></tr></table>
<p>Payment received through Razorpay. For assistance: support@zeroaxiis.tech</p></body></html>`;
  const { client, bucket } = invoiceStorage();
  const key = `${invoicePrefix()}/${notes.org_id}/${notes.user_id}/invoice_${payment.id}.html`;
  let created = true;
  try {
    await client.send(new PutObjectCommand({ Bucket: bucket, Key: key, Body: html, ContentType: "text/html; charset=utf-8",
    Metadata: { "payment-id": payment.id, "paid-at": date }, IfNoneMatch: "*" }));
  } catch (error) {
    if ((error as { $metadata?: { httpStatusCode?: number } }).$metadata?.httpStatusCode !== 412) throw error;
    created = false;
  }
  if (created) await sendBillingReceipt(notes.email || payment.email, payment.id, html);
  return { success: true, plan: notes.tier, paymentId: payment.id, invoiceKey: key };
}

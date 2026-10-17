import { sendBillingReceipt } from "./billing-mail";
import { generateInvoiceHTML, generateInvoicePDF, getInvoiceDetails, type InvoiceData } from "./invoice-generator";
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
  return invoicePrefixes(user).some(prefix => key.startsWith(prefix) && /^invoice_pay_[A-Za-z0-9]+\.(html|pdf)$/.test(key.slice(prefix.length)));
}

type RefundPayment = { id: string; order_id: string; status: string; amount: number | string; currency: string };
type RefundOrder = { id: string; status: string; amount: number | string; currency: string; notes?: unknown };
type ProcessedRefund = { id: string; payment_id: string; status: string; amount: number; currency: string; created_at: number };
type ProviderRefund = Omit<ProcessedRefund, "amount"> & { amount?: number };

// Only completed refunds count. Provider payment totals can include refunds still being processed.
async function syncPaymentRefunds(provider: ReturnType<typeof razorpay>, payment: RefundPayment, order: RefundOrder, trigger?: ProviderRefund) {
  const notes = (order.notes || {}) as Record<string, string>;
  const amount = Number(payment.amount);
  if (notes.product !== "halomail" || !notes.user_id || !["pro", "business"].includes(notes.tier) ||
      order.id !== payment.order_id || order.status !== "paid" || !["captured", "refunded"].includes(payment.status) ||
      !Number.isSafeInteger(amount) || amount <= 0 || Number(order.amount) !== amount || order.currency !== payment.currency) {
    throw new BillingError("Refund does not match a HaloMail purchase.", 422);
  }
  const refunds = new Map<string, ProcessedRefund>();
  const add = (refund: ProviderRefund) => {
    if (!/^rfnd_[A-Za-z0-9]+$/.test(refund.id) || refund.payment_id !== payment.id || refund.currency !== payment.currency ||
        typeof refund.amount !== "number" || !Number.isSafeInteger(refund.amount) || refund.amount <= 0 || refund.amount > amount ||
        !Number.isSafeInteger(refund.created_at) || refund.created_at <= 0) {
      throw new BillingError("Invalid refund details returned by the payment provider.", 503);
    }
    if (refund.status === "processed") refunds.set(refund.id, {
      id: refund.id, payment_id: refund.payment_id, amount: refund.amount, currency: refund.currency,
      status: refund.status, created_at: refund.created_at,
    });
  };
  if (trigger) add(trigger);
  const seen = new Set<string>();
  for (let skip = 0; ; skip += 100) {
    const page = await provider.payments.fetchMultipleRefund(payment.id, { count: 100, skip });
    if (!Array.isArray(page.items)) throw new BillingError("Refund history is unavailable.", 503);
    const previousSize = seen.size;
    for (const refund of page.items) { seen.add(refund.id); add(refund); }
    if (page.items.length < 100) break;
    if (seen.size === previousSize) throw new BillingError("Refund history pagination did not advance.", 503);
  }
  if (!refunds.size) return { fullyRefunded: false, downgraded: false };
  const total = [...refunds.values()].reduce((sum, refund) => sum + refund.amount, 0);
  if (!Number.isSafeInteger(total) || total > amount) throw new BillingError("Refund amounts exceed the payment.", 503);
  const secret = process.env.RAZORPAY_WEBHOOK_SECRET;
  if (!secret) throw new BillingError("Refund verification is not configured.", 503);
  const body = JSON.stringify({ event: "refund.processed", payload: {
    order: { entity: order }, payment: { entity: payment }, refunds: { items: [...refunds.values()] },
  } });
  const response = await fetch(`${backendURL()}/v1/billing/razorpay/webhook`, {
    method: "POST", cache: "no-store", signal: AbortSignal.timeout(20000),
    headers: { "Content-Type": "application/json", "X-Razorpay-Signature": createHmac("sha256", secret).update(body).digest("hex") }, body,
  });
  const result = response.ok ? await response.json() : null;
  if (!result?.success || !result.orgId || result.userId !== notes.user_id ||
      typeof result.fullyRefunded !== "boolean" || typeof result.downgraded !== "boolean") {
    throw new BillingError("Refund received, but the account update is pending. Retry the webhook.", 503);
  }
  return { fullyRefunded: result.fullyRefunded as boolean, downgraded: result.downgraded as boolean };
}

export async function reconcileRefund(refundId: string) {
  if (!/^rfnd_[A-Za-z0-9]+$/.test(refundId)) throw new BillingError("Invalid refund ID.", 400);
  const provider = razorpay();
  const refund = await provider.refunds.fetch(refundId);
  if (refund.id !== refundId || refund.status !== "processed" || !/^pay_[A-Za-z0-9]+$/.test(refund.payment_id)) {
    throw new BillingError("Refund processing has not been confirmed. Retry the webhook.", 503);
  }
  const payment = await provider.payments.fetch(refund.payment_id);
  if (payment.id !== refund.payment_id || !payment.order_id) throw new BillingError("Refund payment could not be verified.", 422);
  const order = await provider.orders.fetch(payment.order_id);
  if ((order.notes as Record<string, string> | undefined)?.product !== "halomail") return { ignored: true };
  const result = await syncPaymentRefunds(provider, payment, order, refund);
  return { success: true, paymentId: payment.id, ...result };
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
  if (!["captured", "refunded"].includes(payment.status) || order.status !== "paid") throw new BillingError("Payment is awaiting capture. Retry after it is captured; do not pay again.", 409);
  if (Number(payment.amount) !== Number(order.amount) || payment.currency !== order.currency) throw new BillingError("Payment amount does not match the order.", 409);
  if (payment.status === "refunded" || Number(payment.amount_refunded) > 0) {
    const refund = await syncPaymentRefunds(provider, payment, order);
    if (refund.fullyRefunded) return { success: true, refunded: true, paymentId: payment.id, ...refund };
    if (payment.status !== "captured") throw new BillingError("Refund is still processing. Retry after it is completed.", 409);
  }
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
  const invoiceData: InvoiceData = {
    paymentId: payment.id,
    orderId: order.id,
    amountMinor: Number(payment.amount),
    currency: payment.currency,
    date,
    customerName: notes.name || (user?.name ?? "Customer"),
    customerEmail: notes.email || payment.email || (user?.email ?? ""),
    tier: notes.tier,
    billingCycle: cycle,
  };
  const html = generateInvoiceHTML(invoiceData);
  const pdfBuffer = await generateInvoicePDF(invoiceData);

  const { client, bucket } = invoiceStorage();
  const htmlKey = `${invoicePrefix()}/${notes.org_id}/${notes.user_id}/invoice_${payment.id}.html`;
  const pdfKey = `${invoicePrefix()}/${notes.org_id}/${notes.user_id}/invoice_${payment.id}.pdf`;
  let created = false;
  for (const document of [
    { Key: htmlKey, Body: html, ContentType: "text/html; charset=utf-8" },
    { Key: pdfKey, Body: pdfBuffer, ContentType: "application/pdf" },
  ]) {
    try {
      await client.send(new PutObjectCommand({ Bucket: bucket, ...document,
        Metadata: { "payment-id": payment.id, "paid-at": date }, IfNoneMatch: "*" }));
      if (document.Key === pdfKey) created = true;
    } catch (error) {
      if ((error as { $metadata?: { httpStatusCode?: number } }).$metadata?.httpStatusCode !== 412) throw error;
    }
  }
  if (created) {
    const downloadURL = new URL("/api/invoices/download", process.env.PUBLIC_WEB_URL || "https://halomail.zeroaxiis.tech");
    downloadURL.searchParams.set("key", pdfKey);
    await sendBillingReceipt(invoiceData.customerEmail, payment.id, pdfBuffer, {
      ...getInvoiceDetails(invoiceData), downloadURL: downloadURL.href,
    });
  }
  return { success: true, plan: notes.tier, paymentId: payment.id, invoiceKey: pdfKey };
}

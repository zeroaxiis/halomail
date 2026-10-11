import { NextRequest, NextResponse } from "next/server";
import { billingError, fulfillPayment, validSignature } from "@/lib/billing-server";
export const runtime = "nodejs";
export async function POST(req: NextRequest) {
  try {
    if (!process.env.RAZORPAY_WEBHOOK_SECRET) return NextResponse.json({ error: "Webhook not configured" }, { status: 503 });
    const raw = await req.text();
    if (!validSignature(raw, req.headers.get("x-razorpay-signature"), process.env.RAZORPAY_WEBHOOK_SECRET)) return NextResponse.json({ error: "Invalid signature" }, { status: 401 });
    let event;
    try { event = JSON.parse(raw); } catch { return NextResponse.json({ error: "Invalid JSON" }, { status: 400 }); }
    if (!["payment.captured", "order.paid"].includes(event.event)) return NextResponse.json({ ignored: true });
    const paymentId = event.payload?.payment?.entity?.id;
    if (!paymentId) return NextResponse.json({ error: "Missing payment" }, { status: 400 });
    return NextResponse.json(await fulfillPayment(paymentId));
  } catch (error) { return billingError(error); }
}

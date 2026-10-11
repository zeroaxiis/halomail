import { NextRequest, NextResponse } from "next/server";
import { billingError, billingUser, fulfillPayment } from "@/lib/billing-server";
// Authenticated recovery also fetches the authoritative Razorpay payment and order.
export async function POST(req: NextRequest) {
  try {
    const user = await billingUser(req);
    const { paymentId } = await req.json();
    return NextResponse.json(await fulfillPayment(paymentId, user));
  } catch (error) { return billingError(error); }
}

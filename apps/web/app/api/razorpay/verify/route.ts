import { NextRequest, NextResponse } from "next/server";
import { BillingError, billingError, billingUser, fulfillPayment, razorpayKeys, validSignature } from "@/lib/billing-server";
export async function POST(req: NextRequest) {
  try {
    const user = await billingUser(req);
    const { razorpay_order_id, razorpay_payment_id, razorpay_signature } = await req.json();
    if (!validSignature(`${razorpay_order_id}|${razorpay_payment_id}`, razorpay_signature, razorpayKeys().key_secret)) throw new BillingError("Invalid payment signature.", 403);
    return NextResponse.json(await fulfillPayment(razorpay_payment_id, user, razorpay_order_id));
  } catch (error) { return billingError(error); }
}

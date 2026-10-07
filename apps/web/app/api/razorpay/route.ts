import { NextRequest, NextResponse } from "next/server";
import Razorpay from "razorpay";

export async function POST(req: NextRequest) {
  try {
    const { amount, currency, plan, billingCycle } = await req.json();

    // Fallback to the provided keys if Next.js hasn't loaded them from the root .env yet
    const key_id = process.env.RAZERPAY_LIVE_API_KEY || "rzp_live_Tl8rjPPjJnFYHq";
    const key_secret = process.env.RAZERPAY_LIVE_KEY_SECRET || "rzp_live_Tl8rjPPjJnFYHq";

    const instance = new Razorpay({ key_id, key_secret });

    // Razorpay requires amount in subunits (paise, cents, etc.)
    const amountInSubunits = Math.round(amount * 100);

    const options = {
      amount: amountInSubunits,
      currency: currency,
      receipt: `rcpt_${Date.now()}`,
    };

    const order = await instance.orders.create(options);
    return NextResponse.json(order);
  } catch (error: any) {
    console.error("Razorpay Error:", error);
    const errorMsg = error?.error?.description || error?.description || error?.message || "Failed to create Razorpay order";
    return NextResponse.json(
      { error: errorMsg },
      { status: 500 }
    );
  }
}

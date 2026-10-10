import { NextRequest, NextResponse } from "next/server";
import Razorpay from "razorpay";

export async function POST(req: NextRequest) {
  try {
    const { amount, currency, plan, billingCycle, orgId, userId, name, email } = await req.json();

    const key_id = process.env.RAZERPAY_LIVE_API_KEY;
    const key_secret = process.env.RAZERPAY_LIVE_KEY_SECRET;

    if (!key_id || !key_secret) {
      throw new Error("Razorpay keys are missing from environment variables");
    }

    const instance = new Razorpay({ key_id, key_secret });

    // Razorpay requires amount in subunits (paise, cents, etc.)
    const amountInSubunits = Math.round(amount * 100);

    const options = {
      amount: amountInSubunits,
      currency: currency,
      receipt: `rcpt_${Date.now()}`,
      notes: {
        org_id: orgId,
        tier: plan.toLowerCase(),
        account_id: orgId,
        user_id: userId,
        name: name || "",
        email: email || "",
        product: "halomail"
      }
    };

    const order = await instance.orders.create(options);
    return NextResponse.json(order);
  } catch (error: any) {
    console.error("Razorpay Order Creation Error:", error);
    return NextResponse.json(
      { error: "Payment service is temporarily unavailable. Please try again later." },
      { status: 500 }
    );
  }
}

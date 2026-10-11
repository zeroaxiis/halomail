import { NextRequest, NextResponse } from "next/server";
import { BillingError, billingError, billingUser, invoiceStorage, razorpay, razorpayKeys } from "@/lib/billing-server";
import { PRICING } from "@/lib/billing-pricing";
export async function POST(req: NextRequest) {
  try {
    const user = await billingUser(req);
    const { currency, plan, billingCycle } = await req.json();
    if (!["Pro", "Business"].includes(plan) || !["monthly", "yearly"].includes(billingCycle) || !["INR", "USD", "EUR", "GBP", "CAD", "AUD", "SGD", "AED"].includes(currency)) throw new BillingError("Invalid plan, billing cycle or currency.");
    if (!process.env.RAZORPAY_WEBHOOK_SECRET) throw new BillingError("Payments are not configured.", 503);
    invoiceStorage();
    const prices = PRICING[plan as "Pro" | "Business"][billingCycle as "monthly" | "yearly"];
    const order = await razorpay().orders.create({
      amount: Math.round(prices[currency as keyof typeof prices] * 100), currency, receipt: `rcpt_${Date.now()}`,
      notes: { org_id: user.orgId, account_id: user.orgId, user_id: user.id, tier: plan.toLowerCase(),
        billing_cycle: billingCycle, name: user.name, email: user.email, product: "halomail" },
    });
    return NextResponse.json({ ...order, keyId: razorpayKeys().key_id }, { headers: { "Cache-Control": "no-store" } });
  } catch (error) { return billingError(error); }
}

import { headers } from "next/headers";
import { PricingClient } from "@/app/(marketing)/pricing/pricing-client";

export const metadata = {
  title: "Billing & Upgrade",
};

export default async function BillingPage() {
  const headersList = await headers();
  const countryCode = headersList.get("x-vercel-ip-country") || "US";

  return (
    <div className="max-w-6xl mx-auto">
      <div className="mb-8">
        <h1 className="text-3xl font-semibold tracking-tight">Billing & Upgrade</h1>
        <p className="mt-2 text-muted-foreground">
          Manage your subscription or upgrade to unlock higher limits and more features.
        </p>
      </div>
      
      <div className="border border-border rounded-xl bg-card p-2 md:p-6">
        <PricingClient initialCountryCode={countryCode} />
      </div>
    </div>
  );
}

import { headers } from "next/headers";
import { PricingClient } from "@/app/(marketing)/pricing/pricing-client";
import { CurrentPlanClient, InvoicesClient } from "./invoices-client";

export const metadata = {
  title: "Billing & Upgrade",
};

export default async function BillingPage() {
  const headersList = await headers();
  const countryCode = headersList.get("x-vercel-ip-country") || "US";

  return (
    <div className="max-w-6xl mx-auto pb-12">
      <div className="mb-8">
        <h1 className="text-3xl font-semibold tracking-tight">Billing & Upgrade</h1>
        <p className="mt-2 text-muted-foreground">
          Manage your subscription or upgrade to unlock higher limits and more features.
        </p>
      </div>

      <CurrentPlanClient />
      
      <div className="border border-border rounded-xl bg-card p-2 md:p-6 mb-12">
        <PricingClient initialCountryCode={countryCode} />
      </div>

      <div className="border border-border rounded-xl bg-card p-6">
        <h2 className="text-xl font-semibold tracking-tight border-b border-border pb-4 mb-4">Payment History & Invoices</h2>
        <p className="text-sm text-muted-foreground">Download your past transaction invoices here. Invoices are securely stored in Cloudflare R2.</p>
        <InvoicesClient />
      </div>
    </div>
  );
}

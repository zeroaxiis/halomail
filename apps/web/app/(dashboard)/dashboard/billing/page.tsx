import { headers } from "next/headers";
import { PageHeader } from "@/components/dashboard/page-header";
import { PanelHeading } from "@/components/dashboard/ui";
import { PricingClient } from "@/app/(marketing)/pricing/pricing-client";
import { CurrentPlanClient, InvoicesClient } from "./invoices-client";

export const metadata = { title: "Billing" };

export default async function BillingPage() {
	const headersList = await headers();
	const countryCode = headersList.get("x-vercel-ip-country") || "US";
	return (
		<>
			<PageHeader title="Billing" />
			<CurrentPlanClient />
			<section>
				<PanelHeading title="Available plans" />
				<div className="workspace-billing-pricing">
					<PricingClient initialCountryCode={countryCode} />
				</div>
			</section>
			<section>
				<PanelHeading title="Payment history" />
				<InvoicesClient />
			</section>
		</>
	);
}

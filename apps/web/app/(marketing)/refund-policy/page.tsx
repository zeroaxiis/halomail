import { Metadata } from "next";

export const metadata: Metadata = {
  title: "Refund Policy | HaloMail",
  description: "HaloMail refund and cancellation policy.",
};

export default function RefundPolicyPage() {
  return (
    <div className="mx-auto max-w-3xl px-6 py-24 sm:py-32">
      <h1 className="text-4xl font-bold tracking-tight text-foreground sm:text-5xl mb-12">
        Refund Policy
      </h1>
      <div className="space-y-8 text-muted-foreground">
        <section>
          <h2 className="text-2xl font-semibold text-foreground mb-4">1. Refund Eligibility</h2>
          <p>
            We strive to provide the best service possible. However, if you are not entirely satisfied with your purchase, we're here to help. Refunds are only applicable under specific conditions and must be requested within <strong>7 days</strong> of the original purchase date.
          </p>
        </section>

        <section>
          <h2 className="text-2xl font-semibold text-foreground mb-4">2. Conditions for Refund</h2>
          <p className="mb-4">To be eligible for a refund, one of the following conditions must be met:</p>
          <ul className="list-disc pl-6 space-y-2">
            <li>You were charged incorrectly due to a billing error on our end.</li>
            <li>The service was completely inaccessible or non-functional for an extended period, and our support team was unable to resolve the issue.</li>
            <li>You accidentally purchased a subscription and did not use any paid features (e.g., you did not send any emails or use premium bandwidth).</li>
          </ul>
        </section>

        <section>
          <h2 className="text-2xl font-semibold text-foreground mb-4">3. Non-Refundable Situations</h2>
          <p className="mb-4">Refunds will <strong>not</strong> be granted in the following scenarios:</p>
          <ul className="list-disc pl-6 space-y-2">
            <li>Requests made after 7 days of the transaction date.</li>
            <li>You have actively utilized the service (e.g., sent emails, used API bandwidth) during the billing period.</li>
            <li>Account termination due to a violation of our Terms of Service (e.g., sending spam, phishing).</li>
            <li>Change of mind after extensive use of the platform.</li>
          </ul>
        </section>

        <section>
          <h2 className="text-2xl font-semibold text-foreground mb-4">4. Process for Requesting a Refund</h2>
          <p>
            To request a refund, please contact our support team at <strong>support@zeroaxiis.tech</strong> with your account details, transaction ID, and a detailed explanation of your request. Our team will review your request and typically respond within 48-72 hours.
          </p>
        </section>

        <section>
          <h2 className="text-2xl font-semibold text-foreground mb-4">5. Processing Time</h2>
          <p>
            If your refund is approved, we will initiate a refund to your original method of payment (via Razorpay). You will receive the credit within a certain amount of days, depending on your card issuer's policies (usually 5-7 business days).
          </p>
        </section>
      </div>
    </div>
  );
}

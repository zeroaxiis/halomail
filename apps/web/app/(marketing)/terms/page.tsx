import { Metadata } from "next";

export const metadata: Metadata = {
  title: "Terms and Conditions | HaloMail",
  description: "HaloMail terms of service and usage conditions.",
};

export default function TermsPage() {
  return (
    <div className="mx-auto max-w-3xl px-6 py-24 sm:py-32">
      <h1 className="text-4xl font-bold tracking-tight text-foreground sm:text-5xl mb-12">
        Terms and Conditions
      </h1>
      <div className="space-y-8 text-muted-foreground">
        <section>
          <h2 className="text-2xl font-semibold text-foreground mb-4">1. Acceptance of Terms</h2>
          <p>
            By accessing and using HaloMail (the "Service"), you accept and agree to be bound by the terms and provision of this agreement. If you do not agree to abide by these terms, please do not use this Service.
          </p>
        </section>

        <section>
          <h2 className="text-2xl font-semibold text-foreground mb-4">2. Provision of Services</h2>
          <p>
            HaloMail provides email delivery and API infrastructure. We reserve the right to modify, suspend, or discontinue any aspect of the service at any time without prior notice.
          </p>
        </section>

        <section>
          <h2 className="text-2xl font-semibold text-foreground mb-4">3. User Responsibilities</h2>
          <p className="mb-4">As a user of HaloMail, you agree to:</p>
          <ul className="list-disc pl-6 space-y-2">
            <li>Provide accurate and complete registration information.</li>
            <li>Never use the Service to send unsolicited spam, phishing emails, or malicious content.</li>
            <li>Maintain the security of your account and API keys.</li>
            <li>Comply with all applicable local and international laws regarding email communication.</li>
          </ul>
        </section>

        <section>
          <h2 className="text-2xl font-semibold text-foreground mb-4">4. Account Termination</h2>
          <p>
            We may terminate or suspend access to our Service immediately, without prior notice or liability, for any reason whatsoever, including without limitation if you breach the Terms. Any violation related to spam or abuse will result in immediate permanent suspension without a refund.
          </p>
        </section>

        <section>
          <h2 className="text-2xl font-semibold text-foreground mb-4">5. Payments and Billing</h2>
          <p>
            Payments are processed securely via our payment partners (Razorpay). By subscribing to a paid tier, you agree to pay all applicable fees. Refunds are strictly governed by our Refund Policy, and are generally limited to specific conditions within 7 days of purchase.
          </p>
        </section>

        <section>
          <h2 className="text-2xl font-semibold text-foreground mb-4">6. Limitation of Liability</h2>
          <p>
            In no event shall HaloMail, nor its directors, employees, partners, agents, suppliers, or affiliates, be liable for any indirect, incidental, special, consequential or punitive damages, including without limitation, loss of profits, data, use, goodwill, or other intangible losses, resulting from your access to or use of or inability to access or use the Service.
          </p>
        </section>

        <section>
          <h2 className="text-2xl font-semibold text-foreground mb-4">7. Contact Information</h2>
          <p>
            If you have any questions regarding these Terms, please contact us at <strong>support@zeroaxiis.tech</strong> or <strong>legal@zeroaxiis.tech</strong>.<br /><br />
            <strong>zeroaxiis</strong><br />
            Uttar Pradesh, 273004<br />
            India
          </p>
        </section>
      </div>
    </div>
  );
}

import { Metadata } from "next";

export const metadata: Metadata = {
  title: "Privacy Policy | HaloMail",
  description: "HaloMail global privacy policy, data protection disclosures, and Google API compliance.",
};

export default function PrivacyPage() {
  return (
    <div className="mx-auto max-w-4xl px-6 py-24 sm:py-32">
      <div className="border-b border-border pb-8 mb-10">
        <p className="text-xs font-mono uppercase tracking-widest text-muted-foreground mb-2">Legal Compliance</p>
        <h1 className="text-4xl font-bold tracking-tight text-foreground sm:text-5xl">
          Global Privacy Policy
        </h1>
        <p className="mt-4 text-sm text-muted-foreground">
          Effective Date: October 31, 2026 • Last Updated: October 31, 2026
        </p>
      </div>

      <div className="space-y-10 text-muted-foreground leading-relaxed text-sm sm:text-base">
        <section>
          <h2 className="text-xl font-semibold text-foreground mb-3">1. Overview & Commitments</h2>
          <p>
            HaloMail ("we", "us", or "our", operated by <strong>zeroaxiis</strong>) provides high-performance transactional email delivery, secure embeddable contact forms, and automated calendar scheduling infrastructure. We are committed to safeguarding personal information and complying with global privacy laws, including the European Union and United Kingdom General Data Protection Regulation (<strong>GDPR / UK GDPR</strong>), the California Consumer Privacy Act and California Privacy Rights Act (<strong>CCPA / CPRA</strong>), the Indian Digital Personal Data Protection Act (<strong>DPDP Act 2023</strong>), Brazil&apos;s Lei Geral de Proteção de Dados (<strong>LGPD</strong>), and Canada&apos;s Personal Information Protection and Electronic Documents Act (<strong>PIPEDA</strong>).
          </p>
        </section>

        <section>
          <h2 className="text-xl font-semibold text-foreground mb-3">2. Data Controller vs. Data Processor</h2>
          <p className="mb-3">
            To ensure complete clarity under international privacy frameworks:
          </p>
          <ul className="list-disc pl-6 space-y-2">
            <li>
              <strong>HaloMail as a Data Controller:</strong> We act as the Data Controller with respect to account management details, billing information, and direct communications regarding your HaloMail account.
            </li>
            <li>
              <strong>HaloMail as a Data Processor:</strong> When you embed forms on your website, receive contact messages, or host calendar booking slots, <em>you</em> are the Data Controller for your end-users&apos; data, and HaloMail acts strictly as a Data Processor following your configurations and instructions. You are responsible for ensuring that your collection of data through HaloMail complies with applicable local laws and that proper notices are provided to your visitors.
            </li>
          </ul>
        </section>

        <section>
          <h2 className="text-xl font-semibold text-foreground mb-3">3. Information We Collect</h2>
          <div className="space-y-3">
            <p>We collect only the minimum personal information necessary to deliver and secure our platform:</p>
            <ul className="list-disc pl-6 space-y-2">
              <li><strong>Account Credentials:</strong> Name, verified email address, hashed passwords, and cryptographic session identifiers. One-time login codes (OTP) are generated server-side and never exposed to the client.</li>
              <li><strong>Billing & Payment Information:</strong> Subscription tiers and payment records. Sensitive payment details (e.g., credit card numbers, UPI handles) are tokenized and processed directly by our PCI-DSS compliant payment gateway (Razorpay); HaloMail never stores raw payment card numbers.</li>
              <li><strong>Form Submissions & Booking Records:</strong> Form input fields, submission timestamps, attendee names, attendee email addresses, meeting notes, and selected time slots.</li>
              <li><strong>Technical Diagnostics:</strong> IP addresses, browser user-agent strings, rate-limiting telemetry, and outbox delivery statuses used strictly to prevent spam, detect bot attacks, and ensure network security.</li>
            </ul>
          </div>
        </section>

        <section className="rounded-lg border border-border bg-card/40 p-6">
          <h2 className="text-xl font-semibold text-foreground mb-3">4. Google API Services & Limited Use Policy</h2>
          <p className="mb-3">
            HaloMail allows users to connect their Google Calendar and generate Google Meet video conferencing links. Our use of Google user data is strictly bounded:
          </p>
          <ul className="list-disc pl-6 space-y-2 mb-4">
            <li><strong>Calendar Scopes:</strong> We request access to <code>calendar.events</code> and <code>calendar.readonly</code> solely to calculate real-time scheduling availability, prevent double bookings, and insert confirmed calendar invites.</li>
            <li><strong>Cryptographic Storage:</strong> Google OAuth tokens (access and refresh tokens) are encrypted at rest in our databases using AES-256-GCM authenticated encryption.</li>
            <li><strong>Zero Commercial Exploitation:</strong> We will never sell Google user data to any third party, use Google user data for advertising, or use Google user data to train generalized artificial intelligence or machine learning models.</li>
            <li><strong>Human Review Prohibition:</strong> HaloMail employees and contractors are strictly prohibited from reading or viewing your connected Google Calendar data, except where you have provided explicit consent for debugging a specific support ticket, or where required by law.</li>
          </ul>
          <p className="font-medium text-foreground bg-secondary/50 p-3 rounded text-xs sm:text-sm">
            HaloMail&apos;s use and transfer to any other app of information received from Google APIs will adhere to the <a href="https://developers.google.com/terms/api-services-user-data-policy" target="_blank" rel="noopener noreferrer" className="underline hover:text-foreground">Google API Services User Data Policy</a>, including the Limited Use requirements.
          </p>
        </section>

        <section>
          <h2 className="text-xl font-semibold text-foreground mb-3">5. Legal Bases for Processing (GDPR & International Law)</h2>
          <p className="mb-3">We process personal information under the following lawful grounds:</p>
          <ul className="list-disc pl-6 space-y-2">
            <li><strong>Performance of a Contract:</strong> Processing necessary to provide form routing, deliver notifications, sync appointments, and fulfill our Terms of Service.</li>
            <li><strong>Legitimate Interests:</strong> Securing our infrastructure, enforcing rate limits, preventing automated spam, and diagnosing platform outages.</li>
            <li><strong>Compliance with Legal Obligations:</strong> Retaining tax and transactional records required by applicable financial authorities.</li>
            <li><strong>Consent:</strong> Where you have expressly consented to an integration (such as connecting Google Calendar or subscribing to product updates). Consent may be revoked at any time.</li>
          </ul>
        </section>

        <section>
          <h2 className="text-xl font-semibold text-foreground mb-3">6. "Do Not Sell or Share My Information" (CCPA / CPRA)</h2>
          <p>
            We do <strong>not</strong> sell, rent, release, disclose, disseminate, make available, transfer, or otherwise communicate personal information or recipient contact data to third parties in exchange for monetary or other valuable consideration. Under the California Consumer Privacy Act (CCPA) and California Privacy Rights Act (CPRA), California residents have the right to know what personal information is collected, request deletion, and not receive discriminatory treatment for exercising privacy rights.
          </p>
        </section>

        <section>
          <h2 className="text-xl font-semibold text-foreground mb-3">7. Your Global Privacy Rights & How to Exercise Them</h2>
          <p className="mb-3">Regardless of your jurisdiction, HaloMail grants users comprehensive data rights:</p>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 my-4">
            <div className="rounded border border-border p-4 bg-card/20">
              <h3 className="font-semibold text-foreground text-sm mb-1">Right to Access & Portability</h3>
              <p className="text-xs text-muted-foreground">Request a machine-readable copy (JSON/CSV) of your stored profile, form submissions, and meeting bookings.</p>
            </div>
            <div className="rounded border border-border p-4 bg-card/20">
              <h3 className="font-semibold text-foreground text-sm mb-1">Right to Erasure (Right to be Forgotten)</h3>
              <p className="text-xs text-muted-foreground">Permanently delete your account, forms, API keys, and stored submissions at any time from your dashboard or via API.</p>
            </div>
            <div className="rounded border border-border p-4 bg-card/20">
              <h3 className="font-semibold text-foreground text-sm mb-1">Right to Rectification</h3>
              <p className="text-xs text-muted-foreground">Correct inaccurate or incomplete profile details directly in your account settings.</p>
            </div>
            <div className="rounded border border-border p-4 bg-card/20">
              <h3 className="font-semibold text-foreground text-sm mb-1">Right to Revoke Integrations</h3>
              <p className="text-xs text-muted-foreground">Disconnect Google Calendar with 1-click in your Meetings settings, which immediately purges your stored OAuth tokens.</p>
            </div>
          </div>
          <p>
            To submit an inquiry or exercise any statutory rights, email our team at <strong>privacy@zeroaxiis.tech</strong> or <strong>legal@zeroaxiis.tech</strong>. We respond to all verified requests within thirty (30) days without charge.
          </p>
        </section>

        <section>
          <h2 className="text-xl font-semibold text-foreground mb-3">8. Security Architecture & Safeguards</h2>
          <p>
            We implement industry-grade technical and organizational security controls:
          </p>
          <ul className="list-disc pl-6 space-y-2 mt-2">
            <li>End-to-end transport layer encryption (TLS 1.2+ and TLS 1.3) with HSTS enforcement.</li>
            <li>AES-256-GCM symmetric encryption for all calendar credentials and third-party tokens.</li>
            <li>Cryptographic SHA-256 hashing for all issued API keys (keys are shown once upon creation and cannot be retrieved plaintext).</li>
            <li>Strict Content-Security-Policy (CSP), Clickjacking defense (X-Frame-Options: DENY), and MIME-sniffing protection.</li>
            <li>Continuous automated vulnerability scanning and database isolation.</li>
          </ul>
        </section>

        <section>
          <h2 className="text-xl font-semibold text-foreground mb-3">9. Data Protection Officer & Grievance Redressal</h2>
          <p>
            In compliance with international data regulations and Section 8(9) of the Indian Digital Personal Data Protection Act 2023, you may contact our designated Data Protection & Grievance Officer:
          </p>
          <div className="mt-3 rounded border border-border bg-secondary/20 p-4 font-mono text-xs">
            <p><strong>Entity:</strong> zeroaxiis (HaloMail Platform)</p>
            <p><strong>Designation:</strong> Data Protection Officer / Grievance Redressal Officer</p>
            <p><strong>Email:</strong> legal@zeroaxiis.tech / support@zeroaxiis.tech</p>
            <p><strong>Address:</strong> Uttar Pradesh, 273004, India</p>
          </div>
        </section>
      </div>
    </div>
  );
}

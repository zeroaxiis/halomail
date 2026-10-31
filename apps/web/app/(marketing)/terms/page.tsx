import { Metadata } from "next";

export const metadata: Metadata = {
  title: "Terms of Service | HaloMail",
  description: "HaloMail terms of service, acceptable use conditions, liability limits, and legal protections.",
};

export default function TermsPage() {
  return (
    <div className="mx-auto max-w-4xl px-6 py-24 sm:py-32">
      <div className="border-b border-border pb-8 mb-10">
        <p className="text-xs font-mono uppercase tracking-widest text-muted-foreground mb-2">Legal Terms</p>
        <h1 className="text-4xl font-bold tracking-tight text-foreground sm:text-5xl">
          Terms of Service
        </h1>
        <p className="mt-4 text-sm text-muted-foreground">
          Effective Date: October 31, 2026 • Last Updated: October 31, 2026
        </p>
      </div>

      <div className="space-y-10 text-muted-foreground leading-relaxed text-sm sm:text-base">
        <section>
          <h2 className="text-xl font-semibold text-foreground mb-3">1. Acceptance of Terms & Binding Agreement</h2>
          <p>
            By creating an account, accessing, embedding, or using HaloMail (the "Service", operated by <strong>zeroaxiis</strong>), you enter into a legally binding agreement and accept all terms contained herein. If you are using the Service on behalf of an entity, organization, or company, you represent and warrant that you have full legal authority to bind that entity to these Terms. If you do not agree to these Terms, you must immediately cease all access and use of the Service.
          </p>
        </section>

        <section>
          <h2 className="text-xl font-semibold text-foreground mb-3">2. Description of Services & Evolution</h2>
          <p>
            HaloMail provides developer-first contact form endpoints, transactional email infrastructure, and automated calendar scheduling. We reserve the absolute right to modify, upgrade, deprecate, or suspend any feature or endpoint at any time with or without notice. We shall not be liable to you or any third party for any modification, suspension, or discontinuance of the Service.
          </p>
        </section>

        <section>
          <h2 className="text-xl font-semibold text-foreground mb-3">3. Acceptable Use Policy & Zero-Tolerance Anti-Spam</h2>
          <p className="mb-3">
            You agree to use HaloMail in strict compliance with all applicable domestic and international laws, including CAN-SPAM (US), CASL (Canada), GDPR (EU), and the Indian Information Technology Act. You explicitly represent and warrant that you will NOT:
          </p>
          <ul className="list-disc pl-6 space-y-2">
            <li>Transmit unsolicited bulk email (spam), phishing attempts, deceptive headers, or fraudulent messages.</li>
            <li>Use the Service to transmit malicious code, malware, viruses, or exploit kits.</li>
            <li>Collect, harvest, or solicit passwords, financial credentials, or sensitive health data through forms without explicit, lawful authorization.</li>
            <li>Interfere with, disrupt, or place an unreasonable load on our infrastructure (such as DDoS or bypassing rate limits).</li>
            <li>Attempt to reverse engineer, decompile, or extract the source code of our proprietary routing and anti-abuse systems.</li>
          </ul>
          <p className="mt-3">
            Any violation of this section entitles HaloMail to terminate your account immediately without warning, notice, or refund.
          </p>
        </section>

        <section>
          <h2 className="text-xl font-semibold text-foreground mb-3">4. Customer Data & End-User Consent Responsibility</h2>
          <p className="mb-3">
            When you embed HaloMail forms or calendar booking links on your websites, you act as the independent Data Controller for the personal data of your website visitors and meeting attendees. You agree, represent, and warrant that:
          </p>
          <ul className="list-disc pl-6 space-y-2">
            <li>You have obtained all necessary consents, authorizations, and legal permissions from your visitors before collecting or transmitting their data through HaloMail.</li>
            <li>You maintain a conspicuous, legally compliant Privacy Policy on your own digital properties that accurately informs end-users regarding how their information is collected, processed, and routed.</li>
            <li>HaloMail is merely a technological conduit and processor acting under your direction; HaloMail disclaims all liability for the content, lawfulness, or legitimacy of submissions captured via your access keys.</li>
          </ul>
        </section>

        <section>
          <h2 className="text-xl font-semibold text-foreground mb-3">5. Third-Party Integrations (Google, Resend, Razorpay)</h2>
          <p>
            HaloMail integrates with external third-party services, including Google (for Calendar and Meet conferencing), Resend (for SMTP email transport), and Razorpay (for payment billing). Your use of those integrations is subject to each respective third-party&apos;s terms and policies. HaloMail is not responsible for any downtime, API changes, account suspensions, or service failures originating from third-party platforms.
          </p>
        </section>

        <section className="rounded-lg border border-border bg-card/40 p-6 text-foreground font-mono text-xs sm:text-sm uppercase tracking-wide">
          <h2 className="text-base font-bold text-foreground mb-3 font-sans normal-case tracking-normal">
            6. Conspicuous Disclaimer of Warranties ("AS IS")
          </h2>
          <p className="leading-relaxed">
            TO THE MAXIMUM EXTENT PERMITTED BY APPLICABLE LAW, THE SERVICE, APIS, SOFTWARE, AND DOCUMENTATION ARE PROVIDED ON AN "AS IS" AND "AS AVAILABLE" BASIS, WITHOUT WARRANTIES OF ANY KIND, WHETHER EXPRESS, IMPLIED, STATUTORY, OR OTHERWISE. HALOMAIL AND ZEROAXIIS SPECIFICALLY DISCLAIM ALL IMPLIED WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE, TITLE, QUIET ENJOYMENT, AND NON-INFRINGEMENT. WE DO NOT WARRANT THAT THE SERVICE WILL OPERATE UNINTERRUPTED, SECURE, OR ERROR-FREE, OR THAT ANY EMAILS, NOTIFICATIONS, OR SUBMISSIONS WILL BE DELIVERED WITHOUT LOSS OR DELAY.
          </p>
        </section>

        <section className="rounded-lg border border-border bg-card/40 p-6 text-foreground font-mono text-xs sm:text-sm uppercase tracking-wide">
          <h2 className="text-base font-bold text-foreground mb-3 font-sans normal-case tracking-normal">
            7. Absolute Limitation of Liability
          </h2>
          <p className="leading-relaxed">
            TO THE MAXIMUM EXTENT PERMITTED BY APPLICABLE LAW, IN NO EVENT SHALL HALOMAIL, ZEROAXIIS, ITS FOUNDERS, DIRECTORS, EMPLOYEES, AFFILIATES, OR SUPPLIERS BE LIABLE FOR ANY INDIRECT, INCIDENTAL, SPECIAL, CONSEQUENTIAL, OR PUNITIVE DAMAGES, INCLUDING WITHOUT LIMITATION DAMAGES FOR LOSS OF PROFITS, REVENUE, GOODWILL, DATA, WORK STOPPAGE, OR BUSINESS INTERRUPTION, ARISING OUT OF OR IN CONNECTION WITH YOUR USE OR INABILITY TO USE THE SERVICE.
          </p>
          <p className="mt-4 leading-relaxed">
            UNDER NO CIRCUMSTANCES SHALL HALOMAIL&apos;S AGGREGATE LIABILITY ARISING OUT OF OR RELATING TO THIS AGREEMENT, UNDER ANY LEGAL OR EQUITABLE THEORY (WHETHER IN CONTRACT, TORT, STRICT LIABILITY, OR OTHERWISE), EXCEED THE GREATER OF FIFTY UNITED STATES DOLLARS ($50.00 USD) OR THE TOTAL AMOUNT ACTUALLY PAID BY YOU TO HALOMAIL IN THE TWELVE (12) MONTHS PRECEDING THE CLAIM.
          </p>
        </section>

        <section>
          <h2 className="text-xl font-semibold text-foreground mb-3">8. Comprehensive User Indemnification</h2>
          <p>
            You agree to defend, indemnify, and hold harmless HaloMail, zeroaxiis, and its officers, directors, employees, contractors, and agents from and against any and all claims, liabilities, damages, losses, expenses, judgments, fines, and reasonable attorney&apos;s fees arising out of or related to:
          </p>
          <ul className="list-disc pl-6 space-y-2 mt-2">
            <li>Your access to, use of, or misuse of the Service.</li>
            <li>Your violation of any provision of these Terms or Acceptable Use Policy.</li>
            <li>Any content, message, or personal data submitted, processed, or distributed through your form endpoints or API keys.</li>
            <li>Your violation of any law, regulation, or third-party right (including privacy, publicity, or intellectual property rights).</li>
          </ul>
        </section>

        <section>
          <h2 className="text-xl font-semibold text-foreground mb-3">9. Class Action Waiver & Binding Dispute Resolution</h2>
          <p>
            YOU AND HALOMAIL AGREE THAT ALL DISPUTES MUST BE RESOLVED ON AN INDIVIDUAL BASIS. YOU WAIVE ANY RIGHT TO PARTICIPATE AS A PLAINTIFF OR CLASS MEMBER IN ANY PURPORTED CLASS ACTION, COLLECTIVE ACTION, PRIVATE ATTORNEY GENERAL ACTION, OR REPRESENTATIVE PROCEEDING AGAINST HALOMAIL OR ZEROAXIIS.
          </p>
        </section>

        <section>
          <h2 className="text-xl font-semibold text-foreground mb-3">10. Governing Law & Exclusive Jurisdiction</h2>
          <p>
            These Terms and any dispute or claim arising out of or in connection with them shall be governed by and construed in accordance with the laws of India, without giving effect to any conflict of law principles. You agree that any legal action or proceeding shall be brought exclusively in the competent courts located in Uttar Pradesh, India.
          </p>
        </section>

        <section>
          <h2 className="text-xl font-semibold text-foreground mb-3">11. Legal Inquiries & Notices</h2>
          <p>
            Notices under these Terms should be sent electronically to:
          </p>
          <div className="mt-3 rounded border border-border bg-secondary/20 p-4 font-mono text-xs">
            <p><strong>Legal & Compliance Department:</strong> zeroaxiis (HaloMail)</p>
            <p><strong>Email:</strong> legal@zeroaxiis.tech / support@zeroaxiis.tech</p>
            <p><strong>Location:</strong> Uttar Pradesh, 273004, India</p>
          </div>
        </section>
      </div>
    </div>
  );
}

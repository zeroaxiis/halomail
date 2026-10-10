import { Metadata } from "next";

export const metadata: Metadata = {
  title: "Privacy Policy | HaloMail",
  description: "HaloMail privacy policy and data handling procedures.",
};

export default function PrivacyPage() {
  return (
    <div className="mx-auto max-w-3xl px-6 py-24 sm:py-32">
      <h1 className="text-4xl font-bold tracking-tight text-foreground sm:text-5xl mb-12">
        Privacy Policy
      </h1>
      <div className="space-y-8 text-muted-foreground">
        <section>
          <h2 className="text-2xl font-semibold text-foreground mb-4">1. Information We Collect</h2>
          <p>
            When you use HaloMail, we may collect personal information that you provide to us directly, such as your name, email address, and payment information (processed securely by Razorpay). We also collect diagnostic data regarding the emails you send via our API (such as delivery status) to provide our service.
          </p>
        </section>

        <section>
          <h2 className="text-2xl font-semibold text-foreground mb-4">2. How We Use Your Information</h2>
          <p className="mb-4">We use the collected information for various purposes:</p>
          <ul className="list-disc pl-6 space-y-2">
            <li>To provide and maintain our Service.</li>
            <li>To notify you about changes to our Service.</li>
            <li>To provide customer support.</li>
            <li>To monitor the usage of our Service and prevent abuse (e.g., spam).</li>
          </ul>
        </section>

        <section>
          <h2 className="text-2xl font-semibold text-foreground mb-4">3. Data Security and Privacy</h2>
          <p>
            The security of your data is important to us. We implement strict security measures to protect your personal information. We do not sell your personal information or the email addresses of your recipients to third parties. Email contents transmitted through our API are processed securely and deleted from our active transactional databases in accordance with our retention policies.
          </p>
        </section>

        <section>
          <h2 className="text-2xl font-semibold text-foreground mb-4">4. Third-Party Services</h2>
          <p>
            We may employ third-party companies (such as payment processors like Razorpay and email delivery partners) to facilitate our Service. These third parties have access to your Personal Data only to perform these tasks on our behalf and are obligated not to disclose or use it for any other purpose.
          </p>
        </section>

        <section>
          <h2 className="text-2xl font-semibold text-foreground mb-4">5. Your Rights</h2>
          <p>
            You have the right to access, update, or delete the personal information we hold about you. You can do this from your account dashboard or by contacting our support team at support@zeroaxiis.tech.
          </p>
        </section>

        <section>
          <h2 className="text-2xl font-semibold text-foreground mb-4">6. Changes to This Privacy Policy</h2>
          <p>
            We may update our Privacy Policy from time to time. We will notify you of any changes by posting the new Privacy Policy on this page and updating the "effective date". You are advised to review this Privacy Policy periodically for any changes.
          </p>
        </section>
      </div>
    </div>
  );
}

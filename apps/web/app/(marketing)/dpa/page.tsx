import { Metadata } from "next";

export const metadata: Metadata = {
  title: "Data Processing Agreement | zeroaxiis",
  description: "Data Processing Agreement for zeroaxiis services.",
};

export default function DPAPage() {
  return (
    <div className="mx-auto max-w-3xl px-6 py-24 sm:py-32">
      <h1 className="text-4xl font-bold tracking-tight text-foreground sm:text-5xl mb-12">
        Data Processing Agreement
      </h1>
      <div className="space-y-8 text-muted-foreground">
        <section>
          <h2 className="text-2xl font-semibold text-foreground mb-4">1. Scope and Applicability</h2>
          <p>
            This Data Processing Agreement ("DPA") governs the processing of personal data by zeroaxiis on behalf of our users. We act as a "Data Processor" for the email data and contact forms you submit through our APIs, while you act as the "Data Controller".
          </p>
        </section>

        <section>
          <h2 className="text-2xl font-semibold text-foreground mb-4">2. Processing of Personal Data</h2>
          <p>
            zeroaxiis processes personal data solely for the purpose of providing the Services (e.g., routing emails, scheduling meetings, delivering notifications). We do not use this data for our own marketing purposes, nor do we sell this data to third parties.
          </p>
        </section>

        <section>
          <h2 className="text-2xl font-semibold text-foreground mb-4">3. Data Security</h2>
          <p>
            We implement robust technical and organizational measures to ensure the security of the personal data processed through our platform. This includes encryption in transit, strict access controls, and regular security audits.
          </p>
        </section>

        <section>
          <h2 className="text-2xl font-semibold text-foreground mb-4">4. Sub-processors</h2>
          <p>
            We may use trusted third-party sub-processors (such as cloud hosting providers and payment gateways) to help deliver our Services. All sub-processors are bound by strict data protection obligations equivalent to those in this DPA.
          </p>
        </section>

        <section>
          <h2 className="text-2xl font-semibold text-foreground mb-4">5. Contact</h2>
          <p>
            For privacy and data processing inquiries, please contact our Data Protection Officer at <strong>legal@zeroaxiis.tech</strong>.
          </p>
        </section>
      </div>
    </div>
  );
}

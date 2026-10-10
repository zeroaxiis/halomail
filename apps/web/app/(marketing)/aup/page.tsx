import { Metadata } from "next";

export const metadata: Metadata = {
  title: "Acceptable Use Policy | zeroaxiis",
  description: "Acceptable Use Policy for zeroaxiis services.",
};

export default function AUPPage() {
  return (
    <div className="mx-auto max-w-3xl px-6 py-24 sm:py-32">
      <h1 className="text-4xl font-bold tracking-tight text-foreground sm:text-5xl mb-12">
        Acceptable Use Policy
      </h1>
      <div className="space-y-8 text-muted-foreground">
        <section>
          <h2 className="text-2xl font-semibold text-foreground mb-4">1. Overview</h2>
          <p>
            This Acceptable Use Policy ("AUP") outlines unacceptable uses of zeroaxiis's services (the "Services"). By using our Services, you agree to comply with this AUP.
          </p>
        </section>

        <section>
          <h2 className="text-2xl font-semibold text-foreground mb-4">2. Prohibited Activities</h2>
          <p className="mb-4">You may not use the Services to:</p>
          <ul className="list-disc pl-6 space-y-2">
            <li>Send unsolicited bulk emails (SPAM) or unsolicited promotional materials.</li>
            <li>Send malicious content, including malware, viruses, or phishing emails.</li>
            <li>Promote illegal activities, violence, or hate speech.</li>
            <li>Impersonate any person or entity, or falsely state your affiliation.</li>
            <li>Attempt to bypass rate limits or abuse the API infrastructure.</li>
          </ul>
        </section>

        <section>
          <h2 className="text-2xl font-semibold text-foreground mb-4">3. Enforcement</h2>
          <p>
            We reserve the right to monitor your usage of the Services to ensure compliance with this AUP. Violation of this AUP may result in immediate suspension or termination of your account without prior notice and without a refund.
          </p>
        </section>

        <section>
          <h2 className="text-2xl font-semibold text-foreground mb-4">4. Reporting Abuse</h2>
          <p>
            If you suspect a user is violating this AUP, please report them to <strong>support@zeroaxiis.tech</strong> or <strong>legal@zeroaxiis.tech</strong>.
          </p>
        </section>
      </div>
    </div>
  );
}

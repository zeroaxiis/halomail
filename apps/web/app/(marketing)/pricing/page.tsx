import { headers } from "next/headers";
import { ArrowRight, ArrowUpRight, Check, Minus, Sparkles } from "lucide-react";
import Link from "next/link";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { PricingClient } from "./pricing-client";

export const metadata = { title: "Pricing" };

/* -------------------------------------------------------------------------- */
/* Comparison                                                                 */
/* -------------------------------------------------------------------------- */

const COMPARE: [string, string | boolean, string | boolean, string | boolean, string | boolean][] = [
  ["Users / included seats", "1", "1", "3", "Custom"],
  ["Websites / projects", "1", "5", "25", "Custom"],
  ["Forms", "3", "25", "100", "Custom"],
  ["Form submissions / month", "100", "1,000", "2,500", "Custom"],
  ["Submission history", "30 days", "1 year", "2 years", "Custom"],
  ["Email notifications", true, true, true, true],
  ["Spam protection", true, true, true, true],
  ["Custom redirect", true, true, true, true],
  ["Autoresponders", false, true, true, true],
  ["Custom email templates", false, true, true, true],
  ["Webhook endpoints", false, "3", "20", "Custom"],
  ["File upload storage", false, "1 GB", "10 GB", "Custom"],
  ["CSV / JSON export", false, true, true, true],
  ["Meeting types", "1", "10", "Unlimited", "Unlimited"],
  ["Meeting requests / month", "10", "100", "250", "Custom"],
  ["Calendar connections", "1", "3", "10", "Custom"],
  ["Approval-based booking", true, true, true, true],
  ["Meeting links", true, true, true, true],
  ["Reschedule / cancel", true, true, true, true],
  ["Instant booking", false, true, true, true],
  ["Custom availability", "Basic", true, true, true],
  ["Automated reminders", false, true, true, true],
  ["Team scheduling", false, false, true, true],
  ["Round-robin / assignment", false, false, true, true],
];

/* -------------------------------------------------------------------------- */
/* FAQ                                                                        */
/* -------------------------------------------------------------------------- */

const FAQS = [
  {
    q: "Is there a free tier?",
    a: "Yes. The Free plan is perfect for personal sites, offering 1 website, 3 forms, 100 submissions, and 10 meeting requests every month.",
  },
  {
    q: "What does 'unlimited' meeting types mean?",
    a: "You can create as many different meeting types as you need (e.g. 15-min intro, 60-min deep dive). Monthly meeting requests and other metered features retain their stated limits.",
  },
  {
    q: "Do you charge extra for more users?",
    a: "The Business plan includes 3 users by default. If you need more seats, shared workflows, or custom requirements, reach out for an Enterprise plan.",
  },
  {
    q: "What happens if I go over my monthly limits?",
    a: "We alert you at 80% and 100% of your usage. At the cap, we'll ask you to upgrade rather than surprise you with overage charges.",
  },
  {
    q: "Is storage included?",
    a: "Yes, Pro includes 1 GB and Business includes 10 GB of total stored bytes for file uploads.",
  },
  {
    q: "Is there a contract?",
    a: "Pro and Business plans are billed monthly or yearly, and you can cancel anytime. Enterprise agreements are negotiated separately.",
  },
];

/* -------------------------------------------------------------------------- */

export default async function PricingPage() {
  const headersList = await headers();
  const countryCode = headersList.get("x-vercel-ip-country") || "US";

  return (
    <>
      {/* Hero */}
      <section className="relative overflow-hidden border-b border-border">
        <div aria-hidden className="bg-spotlight pointer-events-none absolute inset-x-0 top-0 h-[420px]" />
        <div
          aria-hidden
          className="bg-grid pointer-events-none absolute inset-0 opacity-40 [mask-image:radial-gradient(ellipse_at_top,black,transparent_65%)]"
        />

        <div className="container relative py-20 md:py-24">
          <div className="mx-auto flex max-w-2xl flex-col items-center text-center">
            <Badge variant="outline" className="mb-6 gap-1.5 bg-card/60 py-1 pl-2 pr-3 backdrop-blur">
              <Sparkles className="size-3.5 text-brand" />
              <span className="font-normal text-muted-foreground">Pricing</span>
            </Badge>
            <h1 className="text-4xl font-semibold tracking-tight md:text-5xl">
              A useful Free tier. A clear upgrade.
            </h1>
            <p className="mt-5 text-balance text-lg text-muted-foreground">
              ZeroAxiis connects forms, enquiries and approval-based meeting requests in one product. Free supports a real small-site workflow; Pro adds automation and capacity.
            </p>
          </div>

          {/* Dynamic Pricing Client */}
          <PricingClient initialCountryCode={countryCode} />

          <p className="mt-12 text-center text-xs text-muted-foreground max-w-2xl mx-auto">
            Yearly is paid upfront. Features and monthly usage limits remain the same for both billing intervals. Prices are per workspace, including the seats listed in the matrix.
          </p>
        </div>
      </section>

      {/* Comparison table */}
      <section className="container py-20 md:py-24">
        <div className="mb-10 flex flex-col items-center gap-4 text-center">
          <Badge variant="outline" className="bg-card/60 px-3 py-1 font-normal text-muted-foreground">
            Feature matrix
          </Badge>
          <h2 className="max-w-xl text-3xl font-semibold tracking-tight md:text-4xl">
            Compare plans side by side
          </h2>
        </div>

        <Card className="mx-auto max-w-5xl overflow-hidden p-0 overflow-x-auto">
          <div className="min-w-[700px] grid grid-cols-[1.6fr_1fr_1fr_1fr_1fr] border-b border-border px-6 py-3.5 text-[11px] uppercase tracking-wide text-muted-foreground">
            <span>Feature</span>
            <span className="text-center">Free</span>
            <span className="text-center">Pro</span>
            <span className="text-center">Business</span>
            <span className="text-center">Enterprise</span>
          </div>

          {COMPARE.map(([label, free, pro, business, enterprise]) => (
            <div
              key={label}
              className="min-w-[700px] grid grid-cols-[1.6fr_1fr_1fr_1fr_1fr] items-center border-b border-border px-6 py-3.5 text-sm last:border-0 hover:bg-muted/30 transition-colors"
            >
              <span className="pr-4 text-muted-foreground">{label}</span>
              <span className="text-center">
                <Cell value={free} />
              </span>
              <span className="text-center">
                <Cell value={pro} />
              </span>
              <span className="text-center">
                <Cell value={business} />
              </span>
              <span className="text-center">
                <Cell value={enterprise} />
              </span>
            </div>
          ))}
        </Card>
      </section>

      {/* FAQ */}
      <section className="container py-20 md:py-24">
        <div className="grid gap-10 lg:grid-cols-2">
          <h2 className="max-w-sm text-3xl font-semibold tracking-tight md:text-4xl">
            Frequently asked questions
          </h2>
          <div className="flex flex-col items-start gap-5">
            <p className="max-w-md text-muted-foreground">
              Have more questions? Feel free to reach out to our team or consult our detailed documentation for integration help.
            </p>
            <Button asChild variant="outline">
              <a href="/docs/index.html">
                Read the docs <ArrowUpRight className="size-4" />
              </a>
            </Button>
          </div>
        </div>

        <div className="mt-14 grid gap-x-10 gap-y-8 border-t border-border pt-10 md:grid-cols-2">
          {FAQS.map((f) => (
            <div key={f.q}>
              <h3 className="font-medium tracking-tight">{f.q}</h3>
              <p className="mt-2 text-sm leading-relaxed text-muted-foreground">{f.a}</p>
            </div>
          ))}
        </div>
      </section>

      {/* CTA */}
      <section className="container pb-24">
        <Card className="relative overflow-hidden p-12 text-center md:p-16">
          <div aria-hidden className="bg-spotlight pointer-events-none absolute inset-0" />
          <div className="relative flex flex-col items-center gap-5">
            <h2 className="max-w-xl text-3xl font-semibold tracking-tight md:text-4xl">
              Start free. Upgrade when you need more.
            </h2>
            <p className="max-w-md text-muted-foreground">
              Experience the platform with our generous free tier, and easily move to a paid plan as you grow.
            </p>
            <div className="mt-2 flex flex-wrap items-center justify-center gap-3">
              <Button asChild size="lg">
                <Link href="/register">
                  Create your account <ArrowRight className="size-4" />
                </Link>
              </Button>
            </div>
          </div>
        </Card>
      </section>
    </>
  );
}

/** Renders a comparison cell: tick, dash, or a short label. */
function Cell({ value }: { value: string | boolean }) {
  if (value === true) return <Check className="mx-auto size-4 text-brand" />;
  if (value === false) return <Minus className="mx-auto size-4 text-muted-foreground/50" />;
  return <span className="text-xs text-muted-foreground font-medium">{value}</span>;
}

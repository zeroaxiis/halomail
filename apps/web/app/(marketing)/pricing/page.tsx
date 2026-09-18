import { ArrowRight, ArrowUpRight, Check, Minus, Server, Sparkles } from "lucide-react";
import Link from "next/link";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";

export const metadata = { title: "Pricing" };

/* -------------------------------------------------------------------------- */
/* Plans                                                                      */
/* -------------------------------------------------------------------------- */

const TIERS = [
  {
    name: "Self-host",
    price: "$0",
    unit: "forever",
    blurb: "The entire platform, running on your own infrastructure. No feature is held back.",
    cta: "Read the deploy guide",
    href: "/docs/integrate.html",
    featured: false,
    features: [
      "Every feature, nothing gated",
      "Unlimited users, forms, and bookings",
      "Google & Outlook calendar sync",
      "Signed webhooks and scoped API keys",
      "Your database, your data, your domain",
      "MIT licensed, fork it if you like",
    ],
  },
  {
    name: "Hosted",
    price: "$5",
    unit: "/month",
    blurb: "The same software, run for you. Upgrades, backups, and uptime are our problem.",
    cta: "Start hosted",
    href: "/register",
    featured: true,
    features: [
      "Everything in Self-host",
      "Managed Postgres with daily backups",
      "Automatic updates and security patches",
      "Email delivery configured out of the box",
      "99.9% uptime target",
      "Email support",
    ],
  },
];

/* -------------------------------------------------------------------------- */
/* Comparison                                                                 */
/* -------------------------------------------------------------------------- */

const COMPARE: [string, string | boolean, string | boolean][] = [
  ["Booking pages", "Unlimited", "Unlimited"],
  ["Contact forms", "Unlimited", "Unlimited"],
  ["Calendar sync (Google, Outlook)", true, true],
  ["Signed webhooks", true, true],
  ["Scoped API keys", true, true],
  ["Email themes + custom HTML", true, true],
  ["Audit log", true, true],
  ["Database", "You run it", "Managed + backups"],
  ["Updates", "git pull", "Automatic"],
  ["Uptime target", "Yours to keep", "99.9%"],
  ["Support", "GitHub issues", "Email"],
  ["Data location", "Anywhere", "Singapore"],
];

/* -------------------------------------------------------------------------- */
/* FAQ                                                                        */
/* -------------------------------------------------------------------------- */

const FAQS = [
  {
    q: "Is the free tier crippled?",
    a: "No. Self-hosting gives you every feature in the codebase. There is no enterprise edition and no license key. You are paying for operations when you choose hosted, not for features.",
  },
  {
    q: "What does $5 actually cover?",
    a: "A managed Postgres with daily backups, automatic deploys of new versions, configured email delivery, and someone to email when something breaks.",
  },
  {
    q: "Can I move between them?",
    a: "In both directions. It is the same schema and the same API, so a database dump moves your data either way. No export fees, no lock-in.",
  },
  {
    q: "What does self-hosting really cost?",
    a: "In monolith mode it fits the free tiers of a container host plus a managed Postgres, genuinely $0 at small scale. An always-on instance with more headroom runs roughly $5–10/month.",
  },
  {
    q: "Do you charge per seat?",
    a: "No. One price covers your whole team, however many people that is.",
  },
  {
    q: "Is there a trial or a refund?",
    a: "Self-hosting is the trial: run the exact same software for as long as you like before paying anything. Cancel hosted whenever; billing stops at the end of the period.",
  },
];

/* -------------------------------------------------------------------------- */

export default function PricingPage() {
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
              Free to run yourself. Five dollars not to.
            </h1>
            <p className="mt-5 text-balance text-lg text-muted-foreground">
              Every feature is in the open-source build. The paid plan buys operations —
              backups, updates, uptime, not unlocked functionality.
            </p>
          </div>

          {/* Plans */}
          <div className="mx-auto mt-14 grid max-w-4xl gap-5 md:grid-cols-2">
            {TIERS.map((t) => (
              <Card
                key={t.name}
                className={`relative p-7 ${
                  t.featured ? "border-brand/40 shadow-xl" : ""
                }`}
              >
                {t.featured ? (
                  <span className="absolute -top-3 left-1/2 -translate-x-1/2">
                    <Badge variant="brand">Recommended</Badge>
                  </span>
                ) : null}

                <div className="flex items-center gap-2">
                  {t.featured ? (
                    <Sparkles className="size-4 text-brand" />
                  ) : (
                    <Server className="size-4 text-muted-foreground" />
                  )}
                  <span className="font-medium">{t.name}</span>
                </div>

                <div className="mt-4 flex items-baseline gap-1.5">
                  <span className="text-5xl font-semibold tracking-tight">{t.price}</span>
                  <span className="text-sm text-muted-foreground">{t.unit}</span>
                </div>

                <p className="mt-3 min-h-[48px] text-sm text-muted-foreground">{t.blurb}</p>

                <Button
                  asChild
                  className="mt-6 w-full"
                  size="lg"
                  variant={t.featured ? "brand" : "outline"}
                >
                  {t.href.startsWith("/docs") ? (
                    <a href={t.href}>
                      {t.cta} <ArrowUpRight className="size-4" />
                    </a>
                  ) : (
                    <Link href={t.href}>
                      {t.cta} <ArrowRight className="size-4" />
                    </Link>
                  )}
                </Button>

                <ul className="mt-7 space-y-3 text-sm">
                  {t.features.map((f) => (
                    <li key={f} className="flex items-start gap-2.5">
                      <Check className="mt-0.5 size-4 shrink-0 text-brand" />
                      <span className="text-muted-foreground">{f}</span>
                    </li>
                  ))}
                </ul>
              </Card>
            ))}
          </div>

          <p className="mt-8 text-center text-xs text-muted-foreground">
            No credit card to self-host · cancel hosted anytime · no per-seat pricing
          </p>
        </div>
      </section>

      {/* Comparison table */}
      <section className="container py-20 md:py-24">
        <div className="mb-10 flex flex-col items-center gap-4 text-center">
          <Badge variant="outline" className="bg-card/60 px-3 py-1 font-normal text-muted-foreground">
            Side by side
          </Badge>
          <h2 className="max-w-xl text-3xl font-semibold tracking-tight md:text-4xl">
            The difference is who runs the server
          </h2>
        </div>

        <Card className="mx-auto max-w-4xl overflow-hidden p-0">
          <div className="grid grid-cols-[1.6fr_1fr_1fr] border-b border-border px-6 py-3.5 text-[11px] uppercase tracking-wide text-muted-foreground">
            <span>Feature</span>
            <span className="text-center">Self-host</span>
            <span className="text-center">Hosted · $5</span>
          </div>

          {COMPARE.map(([label, free, paid]) => (
            <div
              key={label}
              className="grid grid-cols-[1.6fr_1fr_1fr] items-center border-b border-border px-6 py-3.5 text-sm last:border-0"
            >
              <span className="pr-4 text-muted-foreground">{label}</span>
              <span className="text-center">
                <Cell value={free} />
              </span>
              <span className="text-center">
                <Cell value={paid} />
              </span>
            </div>
          ))}
        </Card>
      </section>

      {/* Self-host cost breakdown */}
      <section className="border-y border-border bg-card/30 py-20 md:py-24">
        <div className="container grid items-center gap-12 lg:grid-cols-2">
          <div>
            <h2 className="max-w-md text-3xl font-semibold tracking-tight md:text-4xl">
              What self-hosting actually costs
            </h2>
            <p className="mt-5 max-w-md text-muted-foreground">
              In monolith mode HaloMail is one container plus a Postgres database. At
              portfolio scale that fits inside free tiers. The guide walks through the
              exact providers and settings.
            </p>
            <Button asChild variant="outline" size="lg" className="mt-8">
              <a href="/docs/integrate.html">
                Deploy it yourself <ArrowUpRight className="size-4" />
              </a>
            </Button>
          </div>

          <Card className="overflow-hidden p-0">
            {[
              ["Hobby · free tiers, monolith", "$0"],
              ["Small · always-on 256 MB + managed Postgres", "$5–10"],
              ["Growth · services scaled independently", "per service"],
            ].map(([label, cost], i) => (
              <div
                key={label}
                className={`flex items-center justify-between px-6 py-5 ${
                  i === 0 ? "" : "border-t border-border"
                }`}
              >
                <span className="pr-4 text-sm text-muted-foreground">{label}</span>
                <span className="whitespace-nowrap font-mono text-sm font-medium">{cost}</span>
              </div>
            ))}
          </Card>
        </div>
      </section>

      {/* FAQ */}
      <section className="container py-20 md:py-24">
        <div className="grid gap-10 lg:grid-cols-2">
          <h2 className="max-w-sm text-3xl font-semibold tracking-tight md:text-4xl">
            Questions about the money.
          </h2>
          <div className="flex flex-col items-start gap-5">
            <p className="max-w-md text-muted-foreground">
              Short version: the code is free and always will be. Five dollars a month
              means you never have to think about Postgres upgrades again.
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
              Start free. Upgrade only if you get tired of ops.
            </h2>
            <p className="max-w-md text-muted-foreground">
              Same software either way. Moving between them is a database dump.
            </p>
            <div className="mt-2 flex flex-wrap items-center justify-center gap-3">
              <Button asChild size="lg">
                <Link href="/register">
                  Create your account <ArrowRight className="size-4" />
                </Link>
              </Button>
              <Button asChild size="lg" variant="outline">
                <a href="https://github.com/aashishrajdev/halomail" target="_blank" rel="noreferrer">
                  View on GitHub
                </a>
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
  return <span className="text-xs text-muted-foreground">{value}</span>;
}

import {
  ArrowRight,
  ArrowUpRight,
  CalendarClock,
  CheckCircle2,
  Code2,
  Inbox,
  KeyRound,
  Mail,
  Palette,
  ShieldCheck,
  Webhook,
  Zap,
} from "lucide-react";
import Link from "next/link";
import { FaqList } from "@/components/marketing/faq-list";
import { FormsTableCard } from "@/components/marketing/forms-table-card";
import { HeroPreview } from "@/components/marketing/hero-preview";
import { SchedulingDemo } from "@/components/marketing/scheduling-demo";
import { SpamDemo } from "@/components/marketing/spam-demo";
import { ThemeShowcase } from "@/components/marketing/theme-showcase";
import { WeekCard } from "@/components/marketing/week-card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";

export default function LandingPage() {
  return (
    <>
      <Hero />
      <LogoStrip />
      <Features />
      <Dashboard />
      <FormsTable />
      <HowItWorks />
      <Integrate />
      <FeatureDetail />
      <Stats />
      <Faq />
      <FinalCta />
    </>
  );
}

/* -------------------------------------------------------------------------- */
/* Section shell                                                              */
/* -------------------------------------------------------------------------- */

function SectionLabel({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex justify-center">
      <span className="rounded-full border border-border bg-card/60 px-4 py-1.5 backdrop-blur">
        <span className="section-label text-sm font-medium uppercase tracking-[0.25em]">
          {children}
        </span>
      </span>
    </div>
  );
}

function SectionHead({
  label,
  title,
  desc,
}: {
  label: string;
  title: React.ReactNode;
  desc?: string;
}) {
  return (
    <div className="mb-12 flex flex-col items-center gap-4 text-center">
      <SectionLabel>{label}</SectionLabel>
      <h2 className="max-w-2xl text-3xl font-semibold tracking-tight md:text-4xl">{title}</h2>
      {desc ? <p className="max-w-xl text-balance text-muted-foreground">{desc}</p> : null}
    </div>
  );
}

/* -------------------------------------------------------------------------- */
/* Hero                                                                       */
/* -------------------------------------------------------------------------- */

function Hero() {
  return (
    <section className="relative overflow-hidden border-b border-border">
      <div aria-hidden className="bg-spotlight pointer-events-none absolute inset-x-0 top-0 h-[520px]" />
      <div
        aria-hidden
        className="bg-grid pointer-events-none absolute inset-0 opacity-50 [mask-image:radial-gradient(ellipse_at_top,black,transparent_65%)]"
      />

      <div className="container relative flex flex-col items-center py-24 text-center md:py-32">
        <h1 className="max-w-4xl animate-fade-up text-4xl font-semibold leading-[1.08] tracking-tight md:text-6xl">
          Scheduling and contact forms,
          <br className="hidden md:block" /> behind one clean API.
        </h1>

        <p className="mt-6 max-w-xl animate-fade-up text-balance text-lg text-muted-foreground">
          Give every user a public booking page and an embeddable contact form, backed by
          typed ConnectRPC, a generated SDK, signed webhooks, and a dashboard you don&apos;t
          have to build.
        </p>

        <div className="mt-9 flex animate-fade-up flex-wrap items-center justify-center gap-3">
          <Button asChild size="lg">
            <Link href="/register">
              Get started <ArrowRight className="size-4" />
            </Link>
          </Button>
          <Button asChild size="lg" variant="outline">
            <a href="/docs/integrate.html">Read the guide</a>
          </Button>
        </div>

        <p className="mt-5 text-xs text-muted-foreground">
          No credit card · deploys on free tiers · self-host anywhere
        </p>

        <HeroPreview />
      </div>

      {/* Horizon curve closing the hero. */}
      <div aria-hidden className="bg-arc pointer-events-none absolute -bottom-px left-1/2 h-40 w-[160%] -translate-x-1/2" />
    </section>
  );
}

/* -------------------------------------------------------------------------- */
/* Logo strip                                                                 */
/* -------------------------------------------------------------------------- */

const STACK = [
  "PostgreSQL",
  "Redis",
  "Google Calendar",
  "Google Meet",
  "Resend",
  "ConnectRPC",
  "Docker",
  "OpenTelemetry",
];

function LogoStrip() {
  return (
    <section className="border-b border-border py-16">
      <SectionLabel>Plays well with</SectionLabel>
      <h2 className="mt-5 text-center text-xl font-medium tracking-tight md:text-2xl">
        Built on the tools you already run
      </h2>

      <div className="relative mt-8 overflow-hidden [mask-image:linear-gradient(to_right,transparent,black_12%,black_88%,transparent)]">
        <div className="flex w-max animate-marquee">
          {[...STACK, ...STACK].map((name, i) => (
            <div
              key={`${name}-${i}`}
              className="mr-3 flex h-14 min-w-[180px] items-center justify-center rounded-lg border border-border bg-card px-6 text-sm text-muted-foreground"
            >
              {name}
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

/* -------------------------------------------------------------------------- */
/* Features — bento                                                           */
/* -------------------------------------------------------------------------- */

function Features() {
  return (
    <section id="features" className="container scroll-mt-24 py-20 md:py-28">
      <SectionHead
        label="Features"
        title="Everything a booking link needs, and nothing it doesn't"
        desc="Two products sharing one identity layer, one API, and one deployable binary."
      />

      <div className="grid gap-4 lg:grid-cols-3">
        {/* Scheduling — wide */}
        <Card className="bg-card relative overflow-hidden p-6 lg:col-span-2">
          <FeatureHead
            icon={<CalendarClock />}
            title="Scheduling that respects timezones"
            desc="Weekly availability rules, date overrides, Google Calendar sync, and auto-generated Google Meet links. Slots are computed server-side so two people never book the same minute."
          />
          <SchedulingDemo />
        </Card>

        {/* Spam */}
        <Card className="bg-card p-6">
          <FeatureHead
            icon={<ShieldCheck />}
            title="Spam-proof by default"
            desc="Honeypot fields, per-IP rate limiting, and a spam score on every submission."
          />
          <SpamDemo />
        </Card>

        {/* Developer */}
        <Card className="bg-card overflow-hidden p-6">
          <FeatureHead
            icon={<Code2 />}
            title="Typed end to end"
            desc="Protobuf definitions generate both the Go server and the TypeScript SDK, so a renamed field breaks the build instead of production."
          />
          <div className="mt-6 rounded-lg border border-border bg-background p-4">
            <pre className="whitespace-pre-wrap break-words font-mono text-[12px] leading-relaxed text-muted-foreground">
              <code>{`const halo = new HaloMail({
  apiUrl: process.env.API_URL,
});

await halo.contact.submitMessage({
  formSlug:    "portfolio",
  senderName:  "Jane Visitor",
  senderEmail: "jane@example.com",
  data:        { message: "Let's talk." },
});`}</code>
            </pre>
          </div>
        </Card>

        {/* Email designer — wide so the theme previews get room */}
        <Card className="bg-card p-6 lg:col-span-2">
          <FeatureHead
            icon={<Palette />}
            title="Emails that don't look templated"
            desc="Five built-in themes plus custom HTML, with live preview before you send."
          />
          <ThemeShowcase />
        </Card>
      </div>
    </section>
  );
}

function FeatureHead({
  icon,
  title,
  desc,
}: {
  icon: React.ReactNode;
  title: string;
  desc: string;
}) {
  return (
    <div>
      <div className="mb-4 inline-flex size-9 items-center justify-center rounded-lg border border-border bg-secondary [&_svg]:size-4">
        {icon}
      </div>
      <h3 className="text-lg font-medium tracking-tight">{title}</h3>
      <p className="mt-2 text-sm text-muted-foreground">{desc}</p>
    </div>
  );
}

/* -------------------------------------------------------------------------- */
/* Dashboard split                                                            */
/* -------------------------------------------------------------------------- */

function Dashboard() {
  return (
    <section className="border-y border-border bg-card/30 py-20 md:py-28">
      <div className="container grid items-center gap-12 lg:grid-cols-2">
        <div>
          <h2 className="max-w-md text-3xl font-semibold tracking-tight md:text-4xl">
            One dashboard for meetings and messages.
          </h2>
          <p className="mt-5 max-w-md text-muted-foreground">
            Every booking, every form submission, every API key and webhook delivery in a
            single place, with an audit log that records who changed what.
          </p>
          <ul className="mt-7 space-y-3">
            {[
              "Read, search, and mark submissions without leaving the app",
              "Rotate scoped API keys and replay failed webhooks",
              "Per-form target addresses and redirect URLs",
            ].map((line) => (
              <li key={line} className="flex gap-3 text-sm">
                <CheckCircle2 className="mt-0.5 size-4 shrink-0 text-brand" />
                <span className="text-muted-foreground">{line}</span>
              </li>
            ))}
          </ul>
          <Button asChild className="mt-8" size="lg">
            <Link href="/dashboard">
              Open the dashboard <ArrowRight className="size-4" />
            </Link>
          </Button>
        </div>

        <WeekCard />
      </div>
    </section>
  );
}

/* -------------------------------------------------------------------------- */
/* Forms table split                                                          */
/* -------------------------------------------------------------------------- */

function FormsTable() {
  return (
    <section className="container py-20 md:py-28">
      <div className="grid items-center gap-12 lg:grid-cols-2">
        <FormsTableCard />

        <div className="order-1 lg:order-2">
          <h2 className="max-w-md text-3xl font-semibold tracking-tight md:text-4xl">
            Embed forms anywhere with an access key.
          </h2>
          <p className="mt-5 max-w-md text-muted-foreground">
            Just like Web3Forms, generate a unique access key for your website. When users submit the form, data is stored and email notifications are instantly routed to you.
          </p>
          <Button asChild variant="outline" size="lg" className="mt-8">
            <a href="/docs/integrate.html">
              See the embed guide <ArrowUpRight className="size-4" />
            </a>
          </Button>
        </div>
      </div>
    </section>
  );
}

/* -------------------------------------------------------------------------- */
/* How it works                                                               */
/* -------------------------------------------------------------------------- */

const STEPS = [
  {
    n: "1",
    title: "Create a form or event type",
    desc: "From the dashboard or one API call. You get a slug and a public booking handle.",
    example: "halomail.app/book/you",
  },
  {
    n: "2",
    title: "Paste two lines into your site",
    desc: "A script tag and a data-halomail attribute on the form you already designed.",
    example: '<form data-halomail="portfolio">',
  },
  {
    n: "3",
    title: "Messages and bookings land",
    desc: "Stored, forwarded to your inbox, and pushed to your endpoints as signed webhooks.",
    example: "POST /SubmitMessage · 200 OK",
  },
];

function HowItWorks() {
  return (
    <section className="border-y border-border bg-card/30 py-20 md:py-28">
      <div className="container">
        <SectionHead
          label="How it works"
          title="Live on your site in three steps"
          desc="No SDK required for the basics: a slug, two lines of HTML, and you're receiving."
        />
        <div className="relative grid gap-4 md:grid-cols-3">
          {STEPS.map((s, i) => (
            <div key={s.n} className="relative">
              <Card className="bg-card h-full p-6">
                <div className="flex items-center gap-3">
                  <span className="flex size-7 items-center justify-center rounded-full border border-brand/40 bg-brand/10 font-mono text-sm text-brand">
                    {s.n}
                  </span>
                  <h3 className="font-medium tracking-tight">{s.title}</h3>
                </div>
                <p className="mt-3 text-sm leading-relaxed text-muted-foreground">{s.desc}</p>
                <p className="mt-4 truncate rounded-md border border-border bg-background px-3 py-2 font-mono text-[11px] text-muted-foreground">
                  {s.example}
                </p>
              </Card>
              {i < STEPS.length - 1 ? (
                <ArrowRight
                  aria-hidden
                  className="absolute -right-4 top-1/2 z-10 hidden size-4 -translate-y-1/2 text-muted-foreground md:block"
                />
              ) : null}
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

/* -------------------------------------------------------------------------- */
/* Integrate — code                                                           */
/* -------------------------------------------------------------------------- */

function Integrate() {
  return (
    <section className="container py-20 md:py-28">
      <SectionHead
        label="Integrate"
        title="Two lines of HTML, or the raw API"
        desc="The widget is progressive enhancement over a form you already have. Prefer to own the request? It's plain JSON over POST."
      />

      <div className="grid gap-4 lg:grid-cols-2">
        <CodeCard title="index.html · drop-in widget">
          {`<script src="https://api.halomail.app/widget.js" defer></script>

<form data-halomail="portfolio">
  <input name="name" required>
  <input name="email" type="email" required>
  <textarea name="message" required></textarea>
  <input name="_hl_hp" tabindex="-1" style="display:none">
  <button type="submit">Send</button>
</form>`}
        </CodeCard>

        <CodeCard title="terminal · book a meeting">
          {`curl https://api.halomail.app/halomail.scheduling.v1.BookingService/CreateBooking \\
  -H 'Content-Type: application/json' \\
  -d '{
    "eventTypeId":  "evt_01a01b8f",
    "inviteeName":  "Grace Hopper",
    "inviteeEmail": "grace@example.com",
    "start":        "2026-06-15T09:00:00Z"
  }'`}
        </CodeCard>
      </div>

      <div className="mt-4 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <Mini icon={<Webhook />} label="Signed webhooks" />
        <Mini icon={<KeyRound />} label="Scoped API keys" />
        <Mini icon={<Mail />} label="Resend or SMTP" />
        <Mini icon={<Zap />} label="Single-binary deploy" />
      </div>
    </section>
  );
}

function CodeCard({ title, children }: { title: string; children: string }) {
  return (
    <Card className="bg-card overflow-hidden">
      <div className="flex items-center gap-2 border-b border-border px-4 py-3">
        <span className="size-2.5 rounded-full bg-destructive/60" />
        <span className="size-2.5 rounded-full bg-yellow-500/60" />
        <span className="size-2.5 rounded-full bg-emerald-500/60" />
        <span className="ml-2 font-mono text-[11px] text-muted-foreground">{title}</span>
      </div>
      <pre className="whitespace-pre-wrap break-words p-5 font-mono text-[12.5px] leading-relaxed">
        <code>{children}</code>
      </pre>
    </Card>
  );
}

function Mini({ icon, label }: { icon: React.ReactNode; label: string }) {
  return (
    <div className="flex items-center gap-3 rounded-lg border border-border bg-card px-4 py-3 text-sm [&_svg]:size-4 [&_svg]:text-muted-foreground">
      {icon}
      <span className="font-medium">{label}</span>
    </div>
  );
}

/* -------------------------------------------------------------------------- */
/* Feature detail — the full surface, folded in from the old /features page   */
/* -------------------------------------------------------------------------- */

const DETAIL = [
  {
    icon: <CalendarClock />,
    title: "Scheduling",
    tagline: "A cal.com-style booking experience for every user.",
    points: [
      "Availability rules with per-date overrides",
      "Google Calendar sync & auto-generated Google Meet links",
      "Reschedule and cancel links built in",
    ],
  },
  {
    icon: <Inbox />,
    title: "Contact forms",
    tagline: "Web3Forms-style access keys for your website.",
    points: [
      "Spam scoring plus rate limiting",
      "Stored, searchable submissions",
      "Instant email forwarding to your inbox",
    ],
  },
  {
    icon: <Palette />,
    title: "Email designer",
    tagline: "Notification emails that match your brand.",
    points: [
      "Five built-in themes, custom HTML too",
      "Template variables, safely substituted",
      "Live preview before you send",
    ],
  },
  {
    icon: <Code2 />,
    title: "Developer surface",
    tagline: "Everything is an API before it's a screen.",
    points: [
      "Scoped keys and signed webhooks",
      "Generated TypeScript SDK",
      "Audit logs on sensitive actions",
    ],
  },
];

function FeatureDetail() {
  return (
    <section className="border-t border-border bg-card/30 py-20 md:py-28">
      <div className="container">
        <SectionHead
          label="The full surface"
          title="Everything you need, nothing you don't"
          desc="One typed API powers the dashboard, the SDK, and your own integrations."
        />
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {DETAIL.map((s) => (
            <Card key={s.title} className="bg-card p-6">
              <div className="mb-4 inline-flex size-9 items-center justify-center rounded-lg border border-border bg-secondary [&_svg]:size-4">
                {s.icon}
              </div>
              <h3 className="font-medium tracking-tight">{s.title}</h3>
              <p className="mt-1.5 text-sm text-muted-foreground">{s.tagline}</p>
              <ul className="mt-4 space-y-2 border-t border-border pt-4 text-[13px] text-muted-foreground">
                {s.points.map((p) => (
                  <li key={p} className="flex gap-2">
                    <span className="mt-[7px] size-1 shrink-0 rounded-full bg-brand" />
                    {p}
                  </li>
                ))}
              </ul>
            </Card>
          ))}
        </div>
      </div>
    </section>
  );
}

/* -------------------------------------------------------------------------- */
/* Stats                                                                      */
/* -------------------------------------------------------------------------- */

const STATS = [
  ["6", "services, one binary"],
  ["~15 MB", "container image"],
  ["$0", "to run on free tiers"],
  ["MIT", "licensed, forever"],
];

function Stats() {
  return (
    <section className="border-y border-border">
      <div className="container grid divide-y divide-border sm:grid-cols-2 sm:divide-x lg:grid-cols-4 lg:divide-y-0">
        {STATS.map(([v, k]) => (
          <div key={k} className="px-6 py-10 text-center">
            <p className="text-3xl font-semibold tracking-tight md:text-4xl">{v}</p>
            <p className="mt-2 text-sm text-muted-foreground">{k}</p>
          </div>
        ))}
      </div>
    </section>
  );
}

/* -------------------------------------------------------------------------- */
/* FAQ                                                                        */
/* -------------------------------------------------------------------------- */

const FAQS = [
  {
    q: "What is HaloMail?",
    a: "An open-source platform that gives every user a public booking page and an embeddable contact form, behind one typed API. Run it as a hosted service or self-host the whole thing as a single container.",
  },
  {
    q: "How do I start using HaloMail?",
    a: "Create an account, make a form or event type, then paste a script tag into your site. The integration guide walks through both, with copy-paste snippets for plain HTML and React.",
  },
  {
    q: "Do I need a database?",
    a: "PostgreSQL, yes. A free Neon project is plenty to start. Redis is optional: without it, rate limiting falls back to an in-memory limiter, which is correct for a single instance.",
  },
  {
    q: "Can I use my own domain for emails?",
    a: "Yes. Verify a domain with your email provider, then point EMAIL_FROM at an address on it. Without a verified domain, delivery is limited to your own account address.",
  },
  {
    q: "Is it really free to run?",
    a: "In monolith mode it fits inside the free tiers of a container host and a managed Postgres. Scale a single service out to its own deployment later without touching code.",
  },
  {
    q: "How do I keep the API off the public internet?",
    a: "Every write endpoint except form submission and booking requires a bearer token. Deploy the gateway behind your own proxy if you want the rest locked down further.",
  },
  {
    q: "Is the free tier crippled?",
    a: "No. Self-hosting gives you every feature in the codebase. There is no enterprise edition and no license key. You are paying for operations when you choose hosted, not for features.",
  },
  {
    q: "What does the $5 hosted plan actually cover?",
    a: "A managed Postgres with daily backups, automatic deploys of new versions, configured email delivery, and someone to email when something breaks.",
  },
  {
    q: "Can I move between hosted and self-hosted?",
    a: "In both directions. It is the same schema and the same API, so a database dump moves your data either way. No export fees, no lock-in.",
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

function Faq() {
  return (
    <section className="container py-20 md:py-28">
      <div className="grid gap-10 lg:grid-cols-2">
        <div>
          <h2 className="max-w-sm text-3xl font-semibold tracking-tight md:text-4xl">
            Have a question? We&apos;ve got answers.
          </h2>
        </div>
        <div className="flex flex-col items-start gap-5">
          <p className="max-w-md text-muted-foreground">
            Confused or curious? The docs cover the whole surface, from the first embed to
            self-hosting the stack on your own infrastructure.
          </p>
          <Button asChild variant="outline">
            <a href="/docs/index.html">
              Read the docs <ArrowUpRight className="size-4" />
            </a>
          </Button>
        </div>
      </div>

      <FaqList items={FAQS} />
    </section>
  );
}

/* -------------------------------------------------------------------------- */
/* Final CTA                                                                  */
/* -------------------------------------------------------------------------- */

function FinalCta() {
  return (
    <section className="container pb-24">
      <Card className="bg-card relative overflow-hidden p-12 text-center md:p-16">
        <div aria-hidden className="bg-spotlight pointer-events-none absolute inset-0" />
        <div className="relative flex flex-col items-center gap-5">
          <SectionLabel>Get started</SectionLabel>
          <h2 className="max-w-xl text-3xl font-semibold tracking-tight md:text-4xl">
            Ship scheduling and contact in an afternoon.
          </h2>
          <p className="max-w-md text-muted-foreground">
            Self-host the whole platform as one container, or run each service on its own.
            Same API either way.
          </p>
          <div className="mt-2 flex flex-wrap items-center justify-center gap-3">
            <Button asChild size="lg">
              <Link href="/register">
                Create your account <ArrowRight className="size-4" />
              </Link>
            </Button>
            <Button asChild size="lg" variant="outline">
              <a href="https://github.com/zeroaxiis/halomail" target="_blank" rel="noreferrer">
                View on GitHub
              </a>
            </Button>
          </div>
        </div>
      </Card>
    </section>
  );
}

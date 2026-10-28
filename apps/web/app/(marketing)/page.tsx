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
    <div className="mb-10 sm:mb-12 flex flex-col items-center gap-3 sm:gap-4 text-center px-2">
      <SectionLabel>{label}</SectionLabel>
      <h2 className="max-w-2xl text-2xl font-semibold tracking-tight sm:text-3xl md:text-4xl">{title}</h2>
      {desc ? <p className="max-w-xl text-balance text-sm sm:text-base text-muted-foreground">{desc}</p> : null}
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

      <div className="container relative flex flex-col items-center py-16 text-center sm:py-24 md:py-32">
        <h1 className="max-w-4xl animate-fade-up text-3xl font-semibold leading-[1.12] tracking-tight sm:text-5xl md:text-6xl">
          Scheduling and contact forms,
          <br className="hidden sm:block" /> behind one API.
        </h1>

        <p className="mt-5 max-w-2xl animate-fade-up text-sm sm:text-base text-muted-foreground">
          Accept meeting bookings and form submissions, delivered straight to your inbox.
        </p>

        <div className="mt-8 sm:mt-10 flex w-full max-w-xs sm:max-w-none animate-fade-up flex-col sm:flex-row items-center justify-center gap-3">
          <Button asChild size="lg" className="w-full sm:w-auto">
            <Link href="/register">
              Get started <ArrowRight className="size-4" />
            </Link>
          </Button>
          <Button asChild size="lg" variant="outline" className="w-full sm:w-auto">
            <a href="/docs/index.html">Read the docs</a>
          </Button>
        </div>

        <p className="mt-4 text-xs sm:text-sm text-muted-foreground/80 font-medium">
          Free tier available · No credit card required · Instant setup in minutes.
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

const INTEGRATIONS = [
  {
    name: "React",
    icon: (
      <svg className="size-6 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.6">
        <ellipse cx="12" cy="12" rx="10" ry="4.5" />
        <ellipse cx="12" cy="12" rx="10" ry="4.5" transform="rotate(60 12 12)" />
        <ellipse cx="12" cy="12" rx="10" ry="4.5" transform="rotate(120 12 12)" />
        <circle cx="12" cy="12" r="1.5" fill="currentColor" />
      </svg>
    ),
  },
  {
    name: "Next.js",
    icon: (
      <svg className="size-6 shrink-0" viewBox="0 0 24 24" fill="currentColor">
        <path d="M12 2C6.477 2 2 6.477 2 12s4.477 10 10 10 10-4.477 10-10S17.523 2 12 2zm3.626 14.13L9.67 8.358V15.7H8.33V7.3h1.493l6.096 7.974V7.3h1.344v8.83h-1.637z" />
      </svg>
    ),
  },
  {
    name: "Vue",
    icon: (
      <svg className="size-6 shrink-0" viewBox="0 0 24 24" fill="currentColor">
        <path d="M2 3h3.5L12 14.5 18.5 3H22L12 20.5 2 3z" />
        <path d="M6.5 3h3L12 7.8 14.5 3h3L12 12.5 6.5 3z" opacity="0.6" />
      </svg>
    ),
  },
  {
    name: "Svelte",
    icon: (
      <svg className="size-6 shrink-0" viewBox="0 0 24 24" fill="currentColor">
        <path d="M19.9 8.2c-.6-1.5-1.9-2.7-3.4-3.2-2.3-.8-4.8-.2-6.7 1.4l-2.6 2.2c-.8.7-1.4 1.6-1.6 2.6-.3 1.3 0 2.6.8 3.7.6 1 1.6 1.7 2.7 2.1l2.5.9c.7.3 1.3.7 1.8 1.2.6.7.9 1.6.8 2.5-.1 1-.7 1.9-1.6 2.4-1.3.7-2.9.7-4.2 0-1-.6-1.6-1.5-1.9-2.5H4c.3 1.7 1.2 3.2 2.6 4.1 1.6 1.1 3.6 1.5 5.5 1.1 1.9-.4 3.5-1.6 4.4-3.3.9-1.7.9-3.7 0-5.4-.5-.9-1.2-1.6-2.1-2.1l-2.7-1.1c-.6-.2-1.2-.6-1.7-1-.5-.6-.8-1.4-.7-2.2.1-.8.6-1.6 1.4-2.1 1.2-.7 2.7-.8 3.9-.2.8.4 1.4 1.2 1.7 2.1h3.6z" />
      </svg>
    ),
  },
  {
    name: "Astro",
    icon: (
      <svg className="size-6 shrink-0" viewBox="0 0 24 24" fill="currentColor">
        <path d="M8.5 18.5c-.3.4-.4.9-.3 1.4.1.7.6 1.3 1.3 1.4.6.1 1.2-.2 1.6-.7l.9-1.3c.4-.6 1.1-.9 1.8-.9h.4c.7 0 1.4.3 1.8.9l.9 1.3c.4.5 1 .8 1.6.7.7-.1 1.2-.7 1.3-1.4.1-.5 0-1-.3-1.4L13.8 3.6c-.4-.6-1-1-1.8-1s-1.4.4-1.8 1L4.5 13.9l4 4.6zm3.5-12.7 3.3 7.2H8.7L12 5.8z" />
      </svg>
    ),
  },
  {
    name: "Nuxt",
    icon: (
      <svg className="size-6 shrink-0" viewBox="0 0 24 24" fill="currentColor">
        <path d="M9.5 7.5 3 18.5h7l2-3.5 2 3.5h7L15 7.5z" />
      </svg>
    ),
  },
  {
    name: "Remix",
    icon: (
      <svg className="size-6 shrink-0" viewBox="0 0 24 24" fill="currentColor">
        <path d="M3.5 4h8.3c3.6 0 6.2 2.3 6.2 5.5 0 2.2-1.3 4.1-3.3 4.9L19 20h-4.3l-3.8-5.1H7.5V20H3.5V4zm4 7h4.1c1.4 0 2.4-.8 2.4-2s-1-2-2.4-2H7.5v4z" />
      </svg>
    ),
  },
  {
    name: "Tailwind CSS",
    icon: (
      <svg className="size-6 shrink-0" viewBox="0 0 24 24" fill="currentColor">
        <path d="M12 6c-2.4 0-3.9 1.2-4.5 3.6 1-.8 2.1-1.1 3.3-.8 1 .2 1.7 1 2.5 1.7 1.3 1.3 2.7 2.7 5.7 2.7 2.4 0 3.9-1.2 4.5-3.6-1 .8-2.1 1.1-3.3.8-1-.2-1.7-1-2.5-1.7-1.3-1.3-2.7-2.7-5.7-2.7zM6 12c-2.4 0-3.9 1.2-4.5 3.6 1-.8 2.1-1.1 3.3-.8 1 .2 1.7 1 2.5 1.7 1.3 1.3 2.7 2.7 5.7 2.7 2.4 0 3.9-1.2 4.5-3.6-1 .8-2.1 1.1-3.3.8-1-.2-1.7-1-2.5-1.7-1.3-1.3-2.7-2.7-5.7-2.7z" />
      </svg>
    ),
  },
  {
    name: "HTML5 / JS",
    icon: (
      <svg className="size-6 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <polyline points="16 18 22 12 16 6" />
        <polyline points="8 6 2 12 8 18" />
      </svg>
    ),
  },
  {
    name: "Framer",
    icon: (
      <svg className="size-6 shrink-0" viewBox="0 0 24 24" fill="currentColor">
        <path d="M4 2h16v7h-8l8 7H4v-7h8L4 2z" />
        <path d="M4 16h8v6l-8-6z" />
      </svg>
    ),
  },
  {
    name: "Webflow",
    icon: (
      <svg className="size-6 shrink-0" viewBox="0 0 24 24" fill="currentColor">
        <path d="M17.8 7.3c-.6-.4-1.3-.6-2-.6-1.5 0-2.8 1-3.3 2.4l-.1.4-.1-.4C11.8 7.7 10.5 6.7 9 6.7c-.8 0-1.5.2-2.1.6L2 17.3h4.6l2.1-6.6c.2-.6.7-1 1.3-1s1.1.4 1.3 1l2.1 6.6H18l2.1-6.6c.2-.6.7-1 1.3-1s1.1.4 1.3 1l2.1 6.6H22L17.8 7.3z" />
      </svg>
    ),
  },
  {
    name: "WordPress",
    icon: (
      <svg className="size-6 shrink-0" viewBox="0 0 24 24" fill="currentColor">
        <path d="M12 2a10 10 0 1 0 10 10A10 10 0 0 0 12 2zm0 18.2a8.2 8.2 0 0 1-4.7-1.5l3.8-10.4 2.8 7.8 1.1-3.1.9 2.5a8.2 8.2 0 0 1-3.9 4.7zm-6.8-5.3L8.6 6.5A8.1 8.1 0 0 1 12 3.8a8.2 8.2 0 0 1 5.4 2l-2 5.5zm13-2.7a8.2 8.2 0 0 1-.7 3.3L15.4 9.6c.3-.5.4-1.1.4-1.7 0-.3 0-.6-.1-.9a8.1 8.1 0 0 1 2.5 4.9z" />
      </svg>
    ),
  },
];

function LogoStrip() {
  return (
    <section className="border-b border-border py-14 sm:py-16 overflow-hidden">
      <SectionLabel>Integrations</SectionLabel>
      <h2 className="mt-4 text-center text-xl font-semibold tracking-tight sm:text-2xl">
        Works with any framework or site builder
      </h2>

      <div className="group relative mt-10 overflow-hidden [mask-image:linear-gradient(to_right,transparent,black_10%,black_90%,transparent)]">
        <div className="flex w-max animate-marquee items-center gap-12 sm:gap-16 group-hover:[animation-play-state:paused] pr-12 sm:pr-16">
          {[...INTEGRATIONS, ...INTEGRATIONS].map((item, i) => (
            <div
              key={`${item.name}-${i}`}
              className="flex items-center gap-3 text-muted-foreground/75 transition-colors duration-200 hover:text-foreground cursor-default select-none"
            >
              <span className="shrink-0 transition-transform duration-200 hover:scale-110">
                {item.icon}
              </span>
              <span className="text-base sm:text-lg font-medium tracking-tight whitespace-nowrap">
                {item.name}
              </span>
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
    <section id="features" className="container scroll-mt-24 py-16 sm:py-20 md:py-28">
      <SectionHead
        label="Features"
        title="Everything your customer interactions require, built for scale"
        desc="A unified platform sharing one identity layer, high-throughput APIs, and enterprise-grade security."
      />

      <div className="grid gap-4 lg:grid-cols-3">
        {/* Scheduling — wide */}
        <Card className="bg-card relative overflow-hidden p-4 sm:p-6 lg:col-span-2">
          <FeatureHead
            icon={<CalendarClock />}
            title="Scheduling that respects timezones"
            desc="Weekly availability rules, date overrides, Google Calendar sync, and auto-generated Google Meet links. Slots are computed server-side so two people never book the same minute."
          />
          <SchedulingDemo />
        </Card>

        {/* Spam */}
        <Card className="bg-card p-4 sm:p-6">
          <FeatureHead
            icon={<ShieldCheck />}
            title="Spam-proof by default"
            desc="Honeypot fields, per-IP rate limiting, and a spam score on every submission."
          />
          <SpamDemo />
        </Card>

        {/* Developer */}
        <Card className="bg-card overflow-hidden p-4 sm:p-6">
          <FeatureHead
            icon={<Code2 />}
            title="Typed end to end"
            desc="Protobuf definitions generate both the Go server and the TypeScript SDK, so a renamed field breaks the build instead of production."
          />
          <div className="mt-6 rounded-lg border border-border bg-background p-3 sm:p-4">
            <pre className="overflow-x-auto whitespace-pre-wrap break-words font-mono text-[11px] sm:text-[12px] leading-relaxed text-muted-foreground">
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
        <Card className="bg-card p-4 sm:p-6 lg:col-span-2">
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
    <section className="border-y border-border bg-card/30 py-16 sm:py-20 md:py-28">
      <div className="container grid items-center gap-8 sm:gap-12 lg:grid-cols-2">
        <div>
          <h2 className="max-w-md text-2xl sm:text-3xl font-semibold tracking-tight md:text-4xl">
            One dashboard for meetings and messages.
          </h2>
          <p className="mt-4 sm:mt-5 max-w-md text-sm sm:text-base text-muted-foreground">
            Every booking, every form submission, every API key and webhook delivery in a
            single place, with an audit log that records who changed what.
          </p>
          <ul className="mt-6 sm:mt-7 space-y-3">
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
          <Button asChild className="mt-7 sm:mt-8 w-full sm:w-auto" size="lg">
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
    <section className="container py-16 sm:py-20 md:py-28">
      <div className="grid items-center gap-8 sm:gap-12 lg:grid-cols-2">
        <FormsTableCard />

        <div className="order-1 lg:order-2">
          <h2 className="max-w-md text-2xl sm:text-3xl font-semibold tracking-tight md:text-4xl">
            Embed forms anywhere with secure access keys.
          </h2>
          <p className="mt-4 sm:mt-5 max-w-md text-sm sm:text-base text-muted-foreground">
            Generate scoped access keys to drop custom contact forms directly into any site or app. Submissions are encrypted, stored, and routed to your team with instant spam filtering and deliverability guarantees.
          </p>
          <Button asChild variant="outline" size="lg" className="mt-7 sm:mt-8 w-full sm:w-auto">
            <a href="/docs/index.html#embed">
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
    <section className="border-y border-border bg-card/30 py-16 sm:py-20 md:py-28">
      <div className="container">
        <SectionHead
          label="How it works"
          title="Live on your site in three steps"
          desc="No SDK required for the basics: a slug, two lines of HTML, and you're receiving."
        />
        <div className="relative grid gap-4 md:grid-cols-3">
          {STEPS.map((s, i) => (
            <div key={s.n} className="relative">
              <Card className="bg-card h-full p-4 sm:p-6">
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
        <Mini icon={<Zap />} label="Fast and reliable" />
      </div>
    </section>
  );
}

function CodeCard({ title, children }: { title: string; children: string }) {
  return (
    <Card className="bg-card overflow-hidden">
      <div className="flex items-center gap-2 border-b border-border px-3 sm:px-4 py-3">
        <span className="size-2.5 rounded-full bg-destructive/60" />
        <span className="size-2.5 rounded-full bg-yellow-500/60" />
        <span className="size-2.5 rounded-full bg-emerald-500/60" />
        <span className="ml-2 font-mono text-[11px] text-muted-foreground truncate">{title}</span>
      </div>
      <pre className="overflow-x-auto whitespace-pre-wrap break-words p-4 sm:p-5 font-mono text-[11.5px] sm:text-[12.5px] leading-relaxed">
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
    tagline: "Frictionless calendar booking experience built for individuals and teams.",
    points: [
      "Availability rules with per-date overrides",
      "Google Calendar sync & auto-generated Google Meet links",
      "Reschedule and cancel links built in",
    ],
  },
  {
    icon: <Inbox />,
    title: "Contact forms",
    tagline: "Secure, zero-maintenance embeddable forms for any website.",
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
    <section className="border-t border-border bg-card/30 py-16 sm:py-20 md:py-28">
      <div className="container">
        <SectionHead
          label="The full surface"
          title="Everything you need, nothing you don't"
          desc="One typed API powers the dashboard, the SDK, and your own integrations."
        />
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {DETAIL.map((s) => (
            <Card key={s.title} className="bg-card p-4 sm:p-6">
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
  ["99.9%", "uptime SLA"],
  ["Global", "edge network"],
  ["100%", "secure"],
  ["24/7", "support"],
];

function Stats() {
  return (
    <section className="border-y border-border">
      <div className="container grid grid-cols-2 divide-y divide-border sm:divide-y-0 sm:divide-x lg:grid-cols-4">
        {STATS.map(([v, k], i) => (
          <div key={k} className={`px-4 sm:px-6 py-6 sm:py-10 text-center ${i % 2 === 1 ? "border-l border-border sm:border-l-0" : ""}`}>
            <p className="text-2xl sm:text-3xl font-semibold tracking-tight md:text-4xl">{v}</p>
            <p className="mt-1.5 sm:mt-2 text-xs sm:text-sm text-muted-foreground">{k}</p>
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
    a: "HaloMail is a unified communication and scheduling platform that delivers embeddable contact forms and seamless booking pages, backed by high-performance APIs and centralized management.",
  },
  {
    q: "How do I start using HaloMail?",
    a: "Create an account, make a form or event type, then paste a script tag into your site. The integration guide walks through both, with copy-paste snippets for plain HTML and React.",
  },
  {
    q: "Can I use my own domain for emails?",
    a: "Yes. Verify a domain with your email provider, then point EMAIL_FROM at an address on it. Without a verified domain, delivery is limited to your own account address.",
  },
  {
    q: "Do you charge per seat?",
    a: "No. One price covers your whole team, however many people that is.",
  },
  {
    q: "Is there a trial or a refund?",
    a: "Cancel whenever; billing stops at the end of the period.",
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
            using the raw API.
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
    <section className="container pb-20 sm:pb-24">
      <Card className="bg-card relative overflow-hidden p-6 sm:p-12 md:p-16 text-center">
        <div aria-hidden className="bg-spotlight pointer-events-none absolute inset-0" />
        <div className="relative flex flex-col items-center gap-4 sm:gap-5">
          <SectionLabel>Get started</SectionLabel>
          <h2 className="max-w-xl text-2xl sm:text-3xl font-semibold tracking-tight md:text-4xl">
            Modernize your scheduling and contact flows today.
          </h2>
          <p className="max-w-md text-sm sm:text-base text-muted-foreground">
            Deploy production-ready booking flows and secure forms in minutes.
          </p>
          <div className="mt-2 flex w-full max-w-xs sm:max-w-none flex-col sm:flex-row items-center justify-center gap-3">
            <Button asChild size="lg" className="w-full sm:w-auto">
              <Link href="/register">
                Create your account <ArrowRight className="size-4" />
              </Link>
            </Button>
          </div>
        </div>
      </Card>
    </section>
  );
}

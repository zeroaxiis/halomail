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
      <span className="inline-flex items-center rounded-full border border-border/80 bg-secondary/50 px-3.5 py-1 text-xs font-medium uppercase tracking-[0.2em] text-muted-foreground">
        {children}
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
  desc?: React.ReactNode;
}) {
  return (
    <div className="mb-10 sm:mb-14 flex flex-col items-center gap-3.5 sm:gap-4 text-center px-2">
      <SectionLabel>{label}</SectionLabel>
      <h2 className="max-w-3xl text-2xl font-semibold tracking-tight sm:text-3xl md:text-4xl text-foreground">
        {title}
      </h2>
      {desc ? (
        <p className="max-w-2xl text-sm sm:text-base text-muted-foreground leading-relaxed">
          {desc}
        </p>
      ) : null}
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

/* -------------------------------------------------------------------------- */
function Features() {
  return (
    <section id="features" className="container scroll-mt-24 py-16 sm:py-20 md:py-28">
      <SectionHead
        label="Features"
        title={
          <>
            Everything your customer interactions require,
            <br className="hidden sm:inline" />
            {" "}built for scale.
          </>
        }
        desc={
          <>
            A unified platform sharing one identity layer and high-throughput APIs,
            <br className="hidden sm:inline" />
            {" "}built for personal portfolios, side projects, and scale.
          </>
        }
      />

      <div className="grid gap-4 lg:grid-cols-3">
        {/* Scheduling — wide */}
        <Card className="bg-card relative overflow-hidden p-4 sm:p-6 lg:col-span-2">
          <FeatureHead
            title="Scheduling that respects timezones"
            desc="Weekly availability rules, date overrides, Google Calendar sync, and auto-generated Google Meet links. Slots are computed server-side so two people never book the same minute."
          />
          <SchedulingDemo />
        </Card>

        {/* Spam */}
        <Card className="bg-card p-4 sm:p-6">
          <FeatureHead
            title="Spam-proof by default"
            desc="Honeypot fields, per-IP rate limiting, and an intelligent spam score on every submission."
          />
          <SpamDemo />
        </Card>

        {/* Developer */}
        <Card className="bg-card overflow-hidden p-4 sm:p-6">
          <FeatureHead
            title="Typed end to end"
            desc="Protobuf definitions generate both the Go server and the TypeScript SDK, so a renamed field breaks the build instead of production."
          />
          <div className="mt-6 rounded-lg border border-border bg-background p-3 sm:p-4">
            <pre className="overflow-x-auto whitespace-pre-wrap break-words font-mono text-[11px] sm:text-[12px] leading-relaxed text-muted-foreground">
              <code>{`const halo = new HaloMail({
  apiKey: process.env.HALOMAIL_API_KEY,
});

await halo.contact.submitMessage({
  accessKey:   "hl_live_a81f09c2",
  senderName:  "Sarah Chen",
  senderEmail: "sarah@example.com",
  data: {
    project:  "Website Redesign",
    timeline: "Next month",
    message:  "Loved your portfolio! Are you free for a quick chat?",
  },
});`}</code>
            </pre>
          </div>
        </Card>

        {/* Email designer — wide so the theme previews get room */}
        <Card className="bg-card p-4 sm:p-6 lg:col-span-2">
          <FeatureHead
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
  title,
  desc,
}: {
  title: string;
  desc: string;
}) {
  return (
    <div>
      <h3 className="text-lg font-semibold tracking-tight text-foreground">{title}</h3>
      <p className="mt-1.5 text-sm text-muted-foreground leading-relaxed">{desc}</p>
    </div>
  );
}

/* -------------------------------------------------------------------------- */
/* Dashboard split                                                            */
/* -------------------------------------------------------------------------- */

function Dashboard() {
  return (
    <section className="border-y border-border py-16 sm:py-20 md:py-28">
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
              "Search, filter, and triage incoming submissions with built-in spam analytics",
              "Rotate scoped access keys for frontend embeds and replay signed webhooks",
              "Per-form notification routing, custom redirects, and auto-response templates",
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
            Drop custom forms into your portfolio, Framer, Webflow, or Next.js app in seconds. Submissions are spam-filtered, securely stored, and delivered directly to your inbox with zero server maintenance.
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
    title: "Create your form or booking link",
    desc: "Pick your personal booking handle (halomail.app/book/you) and set availability, or create a contact form in seconds.",
    example: "halomail.app/book/alex",
  },
  {
    n: "2",
    title: "Paste two lines into your site",
    desc: "Add our script tag and data-halomail attribute to any form on your portfolio, blog, or web app.",
    example: '<form data-halomail="contact">',
  },
  {
    n: "3",
    title: "Delivered straight to your inbox",
    desc: "Messages land in your personal inbox with spam blocked automatically, and bookings sync right to Google Calendar.",
    example: "POST /SubmitMessage · 200 OK",
  },
];

function HowItWorks() {
  return (
    <section className="border-y border-border py-16 sm:py-20 md:py-28">
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
        desc="Drop our widget into any portfolio, Webflow site, or React app. Prefer full control? Submit raw JSON over POST or use our SDK."
      />

      <div className="grid gap-4 lg:grid-cols-2">
        <CodeCard title="index.html · drop-in widget">
          {`<script src="https://api.halomail.app/widget.js" defer></script>

<form data-halomail="contact">
  <input name="name" placeholder="Your Name" required>
  <input name="email" type="email" placeholder="Your Email" required>
  <textarea name="message" placeholder="What would you like to discuss?" required></textarea>
  <input name="_hl_hp" tabindex="-1" style="display:none">
  <button type="submit">Send Message</button>
</form>`}
        </CodeCard>

        <CodeCard title="terminal · book a meeting">
          {`curl https://api.halomail.app/halomail.scheduling.v1.BookingService/CreateBooking \\
  -H 'Content-Type: application/json' \\
  -d '{
    "eventTypeId":  "evt_01a01b8f",
    "inviteeName":  "Sarah Chen",
    "inviteeEmail": "sarah@example.com",
    "start":        "2026-06-15T09:00:00Z"
  }'`}
        </CodeCard>
      </div>

      <div className="mt-4 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <Mini icon={<Mail className="size-4 text-muted-foreground" />} label="Direct inbox forwarding" />
        <Mini icon={<CalendarClock className="size-4 text-muted-foreground" />} label="Google Calendar & Meet sync" />
        <Mini icon={<ShieldCheck className="size-4 text-muted-foreground" />} label="Built-in spam protection" />
        <Mini icon={<Webhook />} label="Webhooks & ConnectRPC API" />
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
    title: "Scheduling",
    tagline: "Frictionless calendar booking experience built for individuals and teams.",
    points: [
      "Availability rules with per-date overrides",
      "Google Calendar sync & auto-generated Google Meet links",
      "Reschedule and cancel links built in",
    ],
  },
  {
    title: "Contact forms",
    tagline: "Secure, zero-maintenance embeddable forms for any website.",
    points: [
      "Spam scoring plus rate limiting",
      "Stored, searchable submissions",
      "Instant email forwarding to your inbox",
    ],
  },
  {
    title: "Email designer",
    tagline: "Notification emails that match your brand.",
    points: [
      "Five built-in themes, custom HTML too",
      "Template variables, safely substituted",
      "Live preview before you send",
    ],
  },
  {
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
    <section className="border-t border-border py-16 sm:py-20 md:py-28">
      <div className="container">
        <SectionHead
          label="The full surface"
          title={
            <>
              Everything you need,
              <br className="hidden sm:inline" />
              {" "}nothing you don't.
            </>
          }
          desc={
            <>
              One typed API powering your forms, calendar booking,
              <br className="hidden sm:inline" />
              {" "}and custom website integrations.
            </>
          }
        />
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {DETAIL.map((s) => (
            <Card key={s.title} className="bg-card p-4 sm:p-6">
              <h3 className="font-semibold tracking-tight text-foreground">{s.title}</h3>
              <p className="mt-1.5 text-sm text-muted-foreground">{s.tagline}</p>
              <ul className="mt-4 space-y-2 border-t border-border pt-4 text-[13px] text-muted-foreground">
                {s.points.map((p) => (
                  <li key={p} className="flex gap-2">
                    <span className="mt-[7px] size-1 shrink-0 rounded-full bg-brand/70" />
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
  ["99.99%", "uptime SLA"],
  ["< 50ms", "API response latency"],
  ["AES-256", "at-rest encryption"],
  ["HMAC-SHA256", "signed webhooks"],
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
    a: "HaloMail is a unified platform for contact forms and meeting booking. It gives creators, developers, and businesses a drop-in contact form backend and a personal booking link (halomail.app/book/you) without building a server or paying for separate tools.",
  },
  {
    q: "Can I use it on my portfolio or personal website?",
    a: "Yes! That is one of our primary use cases. You can paste our 2-line snippet into any HTML, Webflow, Framer, Astro, or React portfolio to receive messages directly in your personal inbox, and share your booking link to let people schedule calls with you.",
  },
  {
    q: "Do I need to build a backend or configure SMTP?",
    a: "No. HaloMail handles form storage, spam filtering, Google Calendar availability, meeting link generation, and email forwarding out of the box.",
  },
  {
    q: "Is there a free tier available?",
    a: "Yes. HaloMail includes a generous free tier with zero credit card required. You can set up your booking handle and contact form in under five minutes.",
  },
  {
    q: "How does spam protection work?",
    a: "Every form includes built-in honeypot detection, per-IP rate limiting, and intelligent scoring. Bots get blocked silently so your personal inbox stays clean.",
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
            Upgrade your customer interaction infrastructure.
          </h2>
          <p className="max-w-md text-sm sm:text-base text-muted-foreground">
            Deploy secure forms and calendar scheduling in minutes. Free tier available with instant setup.
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

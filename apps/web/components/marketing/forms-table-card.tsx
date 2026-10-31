"use client";

import { useState } from "react";
import { Card } from "@/components/ui/card";

type FormEndpoint = {
  slug: string;
  name: string;
  submissions: number;
  cleanRate: string;
  lastActive: string;
  key: string;
};

const FORMS: FormEndpoint[] = [
  {
    slug: "portfolio-contact",
    name: "Portfolio Contact",
    submissions: 142,
    cleanRate: "99.8%",
    lastActive: "2m ago",
    key: "hl_forms_9a2b71c",
  },
  {
    slug: "client-inquiry",
    name: "Client Inquiries",
    submissions: 84,
    cleanRate: "100%",
    lastActive: "14m ago",
    key: "hl_forms_4f1c83e",
  },
  {
    slug: "newsletter",
    name: "Newsletter Signup",
    submissions: 320,
    cleanRate: "99.4%",
    lastActive: "5m ago",
    key: "hl_forms_7e8d24a",
  },
  {
    slug: "book-meeting",
    name: "Discovery Bookings",
    submissions: 56,
    cleanRate: "100%",
    lastActive: "32m ago",
    key: "hl_forms_1b3a90f",
  },
  {
    slug: "product-feedback",
    name: "User Feedback",
    submissions: 48,
    cleanRate: "100%",
    lastActive: "1h ago",
    key: "hl_forms_5c9e62b",
  },
];

export function FormsTableCard() {
  const [activeSlug, setActiveSlug] = useState<string>("portfolio-contact");
  const selected = FORMS.find((f) => f.slug === activeSlug) || FORMS[0];

  return (
    <Card className="bg-card order-2 overflow-hidden border border-border p-0 lg:order-1 shadow-sm">
      {/* Top Header Bar */}
      <div className="flex items-center justify-between border-b border-border px-5 py-3 bg-secondary/20">
        <div className="flex items-center gap-2">
          <span className="font-mono text-xs font-semibold text-foreground">form_endpoints</span>
          <span className="text-muted-foreground/60 text-xs">/</span>
          <span className="text-xs text-muted-foreground">production</span>
        </div>
        <div className="flex items-center gap-2">
          <span className="size-1.5 rounded-full bg-emerald-500 animate-pulse" />
          <span className="text-[11px] font-mono text-muted-foreground">5 active keys</span>
        </div>
      </div>

      {/* Table Head */}
      <div className="grid grid-cols-12 border-b border-border px-5 py-2 text-[11px] uppercase tracking-wider text-muted-foreground bg-secondary/10">
        <span className="col-span-5 font-medium">Endpoint</span>
        <span className="col-span-3 text-right font-medium">Submissions</span>
        <span className="col-span-2 text-right font-medium">Clean</span>
        <span className="col-span-2 text-right font-medium">Activity</span>
      </div>

      {/* Table Rows */}
      <div className="divide-y divide-border">
        {FORMS.map((r) => {
          const isSelected = activeSlug === r.slug;
          return (
            <button
              key={r.slug}
              type="button"
              onClick={() => setActiveSlug(r.slug)}
              className={`grid w-full grid-cols-12 items-center px-5 py-3 text-left text-xs transition-colors focus-visible:outline-none ${
                isSelected
                  ? "bg-secondary/60 text-foreground"
                  : "text-muted-foreground hover:bg-secondary/30 hover:text-foreground"
              }`}
            >
              <div className="col-span-5 flex items-center gap-2">
                <span className="font-mono text-xs text-foreground font-medium">/{r.slug}</span>
              </div>
              <span className="col-span-3 text-right font-mono font-medium text-foreground">
                {r.submissions.toLocaleString()}
              </span>
              <span className="col-span-2 text-right font-mono text-[11px] text-muted-foreground">
                {r.cleanRate}
              </span>
              <span className="col-span-2 text-right font-mono text-[11px] text-muted-foreground/80">
                {r.lastActive}
              </span>
            </button>
          );
        })}
      </div>

      {/* Embed snippet footer for the selected form */}
      <div className="border-t border-border bg-secondary/30 p-4">
        <div className="flex items-center justify-between pb-2 text-[11px] text-muted-foreground">
          <span className="font-medium text-foreground">Live HTML Embed</span>
          <span className="font-mono text-[10px] text-muted-foreground">POST /submit</span>
        </div>
        <pre className="overflow-x-auto rounded border border-border bg-background p-2.5 font-mono text-[11px] leading-relaxed text-muted-foreground">
          <code>{`<form action="https://halomailapi.zeroaxiis.tech/submit" method="POST">
  <input type="hidden" name="access_key" value="${selected.key}" />
  <input type="email" name="senderEmail" required placeholder="visitor@example.com" />
  <textarea name="message" required placeholder="Your message..."></textarea>
  <button type="submit">Send</button>
</form>`}</code>
        </pre>
      </div>
    </Card>
  );
}

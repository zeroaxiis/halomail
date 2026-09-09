"use client";

import { useState } from "react";

const FEATURED = 2;

const THEMES: { name: string; body: React.ReactNode }[] = [
  {
    name: "Minimal",
    body: (
      <div className="flex h-full flex-col bg-white p-3.5 text-zinc-900">
        <p className="text-[12px] font-semibold">Thanks for reaching out</p>
        <p className="mt-2 text-[9.5px] leading-relaxed text-zinc-500">
          Hi Jane, got your message about the portfolio. I&apos;ll reply within a day, or grab
          a slot below.
        </p>
        <div className="mt-auto w-max rounded-sm bg-zinc-900 px-2.5 py-1.5 text-[9px] font-medium text-white">
          Book a call
        </div>
      </div>
    ),
  },
  {
    name: "Apple",
    body: (
      <div className="flex h-full items-center justify-center bg-zinc-100 p-3">
        <div className="w-full max-w-[128px] rounded-2xl bg-white p-3 text-center shadow-sm">
          <div className="mx-auto flex size-7 items-center justify-center rounded-full bg-zinc-100 text-[12px]">
            👋
          </div>
          <p className="mt-1.5 text-[10px] font-semibold text-zinc-900">You&apos;re booked</p>
          <p className="mt-0.5 text-[8.5px] text-zinc-400">Tue, 09:00 · Intro Call</p>
          <div className="mx-auto mt-2.5 w-max rounded-full bg-blue-500 px-2.5 py-1 text-[8.5px] font-medium text-white">
            Add to Calendar
          </div>
        </div>
      </div>
    ),
  },
  {
    name: "Notion",
    body: (
      <div className="h-full bg-white p-3.5 text-zinc-800">
        <div className="text-[14px]">✉️</div>
        <p className="mt-1.5 text-[11px] font-semibold">New message · portfolio</p>
        <p className="mt-1.5 text-[9.5px] leading-relaxed text-zinc-500">
          Jane Visitor wrote: &quot;Saw your portfolio, can we talk?&quot;
        </p>
        <div className="mt-2.5 space-y-1.5 text-[9px] text-zinc-600">
          <div className="flex items-center gap-1.5">
            <span className="flex size-2.5 items-center justify-center rounded-sm bg-zinc-800 text-[7px] text-white">
              ✓
            </span>
            <span>Forwarded to inbox</span>
          </div>
          <div className="flex items-center gap-1.5">
            <span className="size-2.5 rounded-sm border border-zinc-300" />
            <span>Reply within 24h</span>
          </div>
        </div>
      </div>
    ),
  },
  {
    name: "Glass",
    body: (
      <div className="flex h-full items-center justify-center bg-gradient-to-br from-zinc-200 via-zinc-300 to-zinc-400 p-3">
        <div className="w-full max-w-[128px] rounded-xl border border-white/60 bg-white/50 p-3 text-zinc-900 shadow-sm backdrop-blur-md">
          <p className="text-[10px] font-semibold">Meeting confirmed</p>
          <p className="mt-1 text-[8.5px] text-zinc-600">Wed 14:30 · Design Review</p>
          <div className="mt-2.5 w-max rounded-md border border-white/70 bg-white/70 px-2.5 py-1 text-[8.5px] font-medium text-zinc-800">
            Join call
          </div>
        </div>
      </div>
    ),
  },
  {
    name: "Terminal",
    body: (
      <div className="h-full whitespace-nowrap bg-zinc-950 p-3 font-mono text-[9px] leading-relaxed text-emerald-400">
        <p>$ halomail send</p>
        <p className="text-emerald-300/70">from: jane@example.com</p>
        <p className="text-emerald-300/70">subject: let&apos;s talk</p>
        <p className="mt-1.5 text-emerald-200">delivered · 200 OK</p>
        <p className="mt-1">▮</p>
      </div>
    ),
  },
];

/**
 * Expanding theme previews. The active panel is wide, the rest are slivers;
 * hovering (or focusing) a panel hands it the focus, and leaving the strip
 * returns it to the featured theme. State lives in React so it can never
 * stick to a stale panel.
 */
export function ThemeShowcase() {
  const [active, setActive] = useState(FEATURED);

  return (
    <div
      className="mt-6 flex h-64 gap-2"
      onMouseLeave={() => setActive(FEATURED)}
    >
      {THEMES.map((t, i) => (
        <div
          key={t.name}
          onMouseEnter={() => setActive(i)}
          onFocus={() => setActive(i)}
          tabIndex={0}
          role="img"
          aria-label={`${t.name} email theme preview`}
          style={{ flexGrow: active === i ? 4 : 1 }}
          className="relative min-w-0 basis-0 overflow-hidden rounded-md border border-border transition-[flex-grow] duration-500 ease-out motion-reduce:transition-none focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
        >
          <div className="h-full w-[200px] min-w-full">{t.body}</div>
          <span className="absolute bottom-1.5 left-1.5 rounded bg-background/75 px-1.5 py-0.5 text-[10px] text-foreground backdrop-blur">
            {t.name}
          </span>
        </div>
      ))}
    </div>
  );
}

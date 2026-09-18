"use client";

import { useState } from "react";

const DAYS = [
  { day: "Mon", booked: 3 },
  { day: "Tue", booked: 5 },
  { day: "Wed", booked: 2 },
  { day: "Thu", booked: 6 },
  { day: "Fri", booked: 4 },
  { day: "Sat", booked: 1 },
  { day: "Sun", booked: 5 },
];
const MAX = 8;

const BOOKINGS = [
  { title: "Intro Call", when: "Tue 09:00", day: 1 },
  { title: "Design Review", when: "Wed 14:30", day: 2 },
  { title: "Pairing", when: "Thu 11:00", day: 3 },
];

const DEFAULT_DAY = 3;

/**
 * Linked scheduling demo: the booked-hours strip and the bookings list share
 * one selected day — hover either side and the other follows.
 */
export function SchedulingDemo() {
  const [active, setActive] = useState(DEFAULT_DAY);

  return (
    <div className="mt-6 grid gap-3 sm:grid-cols-2">
      <div className="rounded-lg border border-border bg-background p-4">
        <p className="text-xs text-muted-foreground">Weekly rules</p>
        <div className="mt-3 space-y-2">
          {[
            ["Mon – Fri", "09:00 – 17:00"],
            ["Saturday", "Unavailable"],
          ].map(([d, t]) => (
            <div key={d} className="flex items-center justify-between text-xs">
              <span className="text-muted-foreground">{d}</span>
              <span className="font-mono">{t}</span>
            </div>
          ))}
        </div>
        <div className="mt-4 flex h-16 items-end gap-1" onMouseLeave={() => setActive(DEFAULT_DAY)}>
          {DAYS.map((d, i) => (
            <button
              key={d.day}
              type="button"
              aria-label={`${d.day}: ${d.booked} bookings`}
              onMouseEnter={() => setActive(i)}
              onFocus={() => setActive(i)}
              style={{ height: `${(d.booked / MAX) * 100}%` }}
              className={`flex-1 rounded-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring ${
                i === active ? "bg-brand/70" : "bg-secondary hover:bg-secondary/70"
              }`}
            />
          ))}
        </div>
        <p className="mt-2 text-[11px] text-muted-foreground">
          {DAYS[active].day} ·{" "}
          <span className="font-medium text-foreground">{DAYS[active].booked} bookings</span>
        </p>
      </div>

      <div className="rounded-lg border border-border bg-background p-4">
        <p className="text-xs text-muted-foreground">Next bookings</p>
        <div className="mt-3 space-y-2">
          {BOOKINGS.map((b) => (
            <div
              key={b.title}
              onMouseEnter={() => setActive(b.day)}
              className={`flex items-center justify-between rounded-md border px-3 py-2 transition-colors ${
                b.day === active
                  ? "border-brand/40 bg-brand/10"
                  : "border-border hover:border-brand/40 hover:bg-brand/5"
              }`}
            >
              <span className="text-xs font-medium">{b.title}</span>
              <span className="font-mono text-[11px] text-muted-foreground">{b.when}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

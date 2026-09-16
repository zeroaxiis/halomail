"use client";

import { useState } from "react";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";

const DAYS = [
  { day: "Mon", slots: ["10:00", "11:30", "15:00"] },
  { day: "Tue", slots: ["09:00", "09:30", "10:00"] },
  { day: "Wed", slots: ["13:00", "14:30", "16:00"] },
  { day: "Thu", slots: ["09:30", "11:00", "15:30"] },
];

type Row = { id: string; n: string; m: string; fresh?: boolean; booking?: boolean };

const INBOX: Row[] = [
  { id: "seed-1", n: "Jane Visitor", m: "Saw your portfolio, can we talk?" },
  { id: "seed-2", n: "Marc Dev", m: "Question about the API pricing" },
  { id: "seed-3", n: "Priya S.", m: "Booking rescheduled to Thursday" },
];

let rowId = 0;

/**
 * Working product shot. Left: pick a day and slot to book. Right: a live
 * inbox — send it a message yourself. Both flows land as rows with the
 * matching API call shown underneath.
 */
export function HeroPreview() {
  const [day, setDay] = useState(1);
  const [slot, setSlot] = useState<number | null>(0);
  const [feed, setFeed] = useState<Row[]>(INBOX);
  const [draft, setDraft] = useState("");
  const [lastCall, setLastCall] = useState("/SubmitMessage");
  const [selected, setSelected] = useState<string[]>([]);

  const toggleSelect = (id: string) =>
    setSelected((s) => (s.includes(id) ? s.filter((x) => x !== id) : [...s, id]));

  const archiveSelected = () => {
    setFeed((f) => f.filter((r) => !selected.includes(r.id)));
    setSelected([]);
    setLastCall("/ArchiveMessages");
  };

  const pickDay = (i: number) => {
    setDay(i);
    setSlot(null);
  };

  const pickSlot = (i: number) => {
    setSlot(i);
    setFeed((f) => [
      {
        id: `row-${rowId++}`,
        n: "Grace Hopper",
        m: `Intro Call confirmed · ${DAYS[day].day} ${DAYS[day].slots[i]}`,
        fresh: true,
        booking: true,
      },
      // A rebooking replaces the previous confirmation instead of stacking up.
      ...f.filter((r) => !r.booking).map((r) => ({ ...r, fresh: false })),
    ]);
    setLastCall("/CreateBooking");
  };

  const sendMessage = (e: React.FormEvent) => {
    e.preventDefault();
    const text = draft.trim();
    if (!text) return;
    setFeed((f) => [
      { id: `row-${rowId++}`, n: "You", m: text, fresh: true },
      ...f.map((r) => ({ ...r, fresh: false })),
    ]);
    setDraft("");
    setLastCall("/SubmitMessage");
  };

  return (
    <div className="relative mt-16 w-full max-w-4xl animate-fade-up">
      <Card className="glass overflow-hidden p-2 shadow-2xl">
        <div className="rounded-md border border-border bg-background">
          <div className="flex items-center gap-2 border-b border-border px-4 py-2.5">
            <span className="size-2.5 rounded-full bg-destructive/60" />
            <span className="size-2.5 rounded-full bg-yellow-500/60" />
            <span className="size-2.5 rounded-full bg-emerald-500/60" />
            <span className="ml-3 font-mono text-[11px] text-muted-foreground">
              halomail.app/book/grace-hopper
            </span>
          </div>

          <div className="grid gap-4 p-4 text-left sm:grid-cols-2">
            <div className="rounded-lg border border-border p-4">
              <p className="text-xs text-muted-foreground">Intro Call · 30 min</p>
              <p className="mt-1 font-medium">Pick a time</p>
              <div className="mt-3 flex gap-1.5">
                {DAYS.map((d, i) => (
                  <button
                    key={d.day}
                    type="button"
                    aria-pressed={i === day}
                    onClick={() => pickDay(i)}
                    className={`flex-1 rounded-md border px-2 py-2 text-center text-[11px] transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring ${
                      i === day
                        ? "border-foreground/30 bg-secondary font-medium"
                        : "border-border text-muted-foreground hover:bg-secondary/50"
                    }`}
                  >
                    {d.day}
                  </button>
                ))}
              </div>
              <div className="mt-3 space-y-1.5">
                {DAYS[day].slots.map((t, i) => (
                  <button
                    key={t}
                    type="button"
                    aria-pressed={i === slot}
                    onClick={() => pickSlot(i)}
                    className={`block w-full rounded-md border px-3 py-2 text-left font-mono text-xs transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring ${
                      i === slot
                        ? "border-brand/40 bg-brand/10 text-brand"
                        : "border-border text-muted-foreground hover:border-brand/30 hover:text-foreground"
                    }`}
                  >
                    {t}
                    {i === slot ? <span className="float-right">✓ booked</span> : null}
                  </button>
                ))}
              </div>
            </div>

            <div className="rounded-lg border border-border p-4">
              <div className="flex items-center justify-between">
                <p className="text-xs text-muted-foreground">Inbox</p>
                <Badge variant="success" className="px-1.5 py-0 text-[10px]">live</Badge>
              </div>
              <div className="mt-3 space-y-2">
                {feed.length === 0 ? (
                  <p className="rounded-md border border-dashed border-border px-3 py-4 text-center text-[11px] text-muted-foreground">
                    Inbox zero 🎉
                  </p>
                ) : null}
                {feed.slice(0, 3).map((row) => {
                  const isSelected = selected.includes(row.id);
                  return (
                    <button
                      key={row.id}
                      type="button"
                      role="checkbox"
                      aria-checked={isSelected}
                      onClick={() => toggleSelect(row.id)}
                      className={`block w-full rounded-md border px-3 py-2 text-left transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring ${
                        isSelected
                          ? "border-brand/60 bg-brand/10"
                          : row.fresh
                            ? "animate-fade-up border-brand/40 bg-brand/5"
                            : "border-border hover:bg-secondary/40"
                      }`}
                    >
                      <p className="flex items-center gap-1.5 text-xs font-medium">
                        <span
                          className={`flex size-3 items-center justify-center rounded-sm border text-[8px] leading-none ${
                            isSelected
                              ? "border-brand bg-brand text-white"
                              : "border-border text-transparent"
                          }`}
                        >
                          ✓
                        </span>
                        {row.n}
                        {row.fresh && !isSelected ? (
                          <span className="size-1.5 rounded-full bg-brand" />
                        ) : null}
                      </p>
                      <p className="truncate pl-[18px] text-[11px] text-muted-foreground">{row.m}</p>
                    </button>
                  );
                })}
              </div>

              {selected.length > 0 ? (
                <div className="mt-2 flex items-center justify-between rounded-md border border-brand/30 bg-brand/5 px-2.5 py-1.5 text-[11px]">
                  <span className="text-muted-foreground">
                    <span className="font-medium text-foreground">{selected.length}</span> selected
                  </span>
                  <span className="flex gap-2">
                    <button
                      type="button"
                      onClick={() => setSelected([])}
                      className="text-muted-foreground transition-colors hover:text-foreground focus-visible:outline-none"
                    >
                      Clear
                    </button>
                    <button
                      type="button"
                      onClick={archiveSelected}
                      className="font-medium text-brand transition-colors hover:text-brand/80 focus-visible:outline-none"
                    >
                      Archive
                    </button>
                  </span>
                </div>
              ) : null}

              <form onSubmit={sendMessage} className="mt-3 flex gap-1.5">
                <input
                  value={draft}
                  onChange={(e) => setDraft(e.target.value)}
                  placeholder="Try it, say hello…"
                  aria-label="Send a test message"
                  maxLength={80}
                  className="h-8 min-w-0 flex-1 rounded-md border border-border bg-background px-2.5 text-xs placeholder:text-muted-foreground/70 focus:outline-none focus:ring-1 focus:ring-ring"
                />
                <button
                  type="submit"
                  className="h-8 shrink-0 rounded-md bg-primary px-3 text-xs font-medium text-primary-foreground transition-colors hover:bg-primary/90 focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                >
                  Send
                </button>
              </form>

              <p className="mt-3 font-mono text-[10px] text-muted-foreground">
                POST {lastCall} · 200 OK
              </p>
            </div>
          </div>
        </div>
      </Card>
    </div>
  );
}

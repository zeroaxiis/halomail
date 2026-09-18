"use client";

import { useEffect, useRef, useState } from "react";
import { Badge } from "@/components/ui/badge";

type Entry = { email: string; score: number; reason: string };

const POOL: Entry[] = [
  { email: "jane@example.com", score: 0.02, reason: "Honeypot untouched · normal typing cadence" },
  { email: "winner@lottery.biz", score: 0.97, reason: "Honeypot filled · burst of 14 requests from one IP" },
  { email: "marc@studio.dev", score: 0.05, reason: "Known sender · passed all heuristics" },
  { email: "free-crypto@pump.io", score: 0.99, reason: "Link-stuffed body · disposable domain" },
  { email: "recruiter@bigco.com", score: 0.11, reason: "First-time sender · clean content" },
  { email: "seo-guru@backlinks.cc", score: 0.93, reason: "Keyword spam · failed rate limit" },
  { email: "priya@design.studio", score: 0.03, reason: "Honeypot untouched · verified domain" },
];

const VISIBLE = 3;

/**
 * Live spam-scoring demo: submissions stream in and get scored; hovering
 * pauses the stream and reveals why each one passed or failed.
 */
export function SpamDemo() {
  const [head, setHead] = useState(VISIBLE);
  const [paused, setPaused] = useState(false);
  const [hovered, setHovered] = useState<number | null>(null);
  const reduced = useRef(false);

  useEffect(() => {
    reduced.current = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
  }, []);

  useEffect(() => {
    if (paused || reduced.current) return;
    const id = setInterval(() => setHead((h) => h + 1), 2600);
    return () => clearInterval(id);
  }, [paused]);

  const rows = Array.from(
    { length: VISIBLE },
    (_, i) => POOL[(head - i + POOL.length * 100) % POOL.length],
  );

  return (
    <div
      className="mt-6 space-y-2"
      onMouseEnter={() => setPaused(true)}
      onMouseLeave={() => setPaused(false)}
    >
      {rows.map((r, i) => (
        <div
          key={`${r.email}-${head - i}`}
          onMouseEnter={() => setHovered(i)}
          onMouseLeave={() => setHovered(null)}
          className={`rounded-md border border-border bg-background px-3 py-2 transition-colors hover:border-foreground/20 ${
            i === 0 ? "animate-fade-up" : ""
          }`}
        >
          <div className="flex items-center justify-between">
            <span className="truncate text-xs text-muted-foreground">{r.email}</span>
            <Badge
              variant={r.score > 0.5 ? "danger" : "success"}
              className="ml-2 px-1.5 py-0 text-[10px]"
            >
              {r.score.toFixed(2)}
            </Badge>
          </div>
          {hovered === i ? (
            <p className="mt-1 text-[10px] text-muted-foreground">{r.reason}</p>
          ) : null}
        </div>
      ))}
      <p className="text-[10px] text-muted-foreground/70">
        {paused ? "Paused · hover a row for the verdict" : "Scoring live submissions…"}
      </p>
    </div>
  );
}

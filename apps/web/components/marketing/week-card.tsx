"use client";

import { useEffect, useRef, useState } from "react";
import { Card } from "@/components/ui/card";

/** Counts from the previous value to the next one whenever it changes. */
function AnimatedNumber({ value }: { value: string }) {
  const target = parseFloat(value);
  const suffix = value.replace(String(target), "");
  const [display, setDisplay] = useState(target);
  const prev = useRef(target);

  useEffect(() => {
    const from = prev.current;
    prev.current = target;
    if (from === target) return;
    if (window.matchMedia("(prefers-reduced-motion: reduce)").matches) {
      setDisplay(target);
      return;
    }
    const start = performance.now();
    const duration = 600;
    let raf: number;
    const tick = (t: number) => {
      const p = Math.min(1, (t - start) / duration);
      const eased = 1 - (1 - p) ** 3;
      setDisplay(Math.round(from + (target - from) * eased));
      if (p < 1) raf = requestAnimationFrame(tick);
    };
    raf = requestAnimationFrame(tick);
    return () => cancelAnimationFrame(raf);
  }, [target]);

  return (
    <>
      {display}
      {suffix}
    </>
  );
}

type RangeKey = "1D" | "7D";

const CHART = { w: 320, h: 90, top: 10, bottom: 82 };

const RANGES: Record<
  RangeKey,
  {
    stats: [string, string][];
    labels: string[];
    current: number[];
    previous: number[];
    currentName: string;
    previousName: string;
  }
> = {
  "1D": {
    stats: [
      ["Bookings", "6"],
      ["Messages", "21"],
      ["Show rate", "89%"],
    ],
    labels: ["9a", "11a", "1p", "3p", "5p", "7p", "9p"],
    current: [2, 5, 4, 8, 6, 9, 7],
    previous: [1, 3, 4, 5, 4, 6, 5],
    currentName: "Today",
    previousName: "Yesterday",
  },
  "7D": {
    stats: [
      ["Bookings", "34"],
      ["Messages", "128"],
      ["Show rate", "92%"],
    ],
    labels: ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"],
    current: [8, 14, 11, 24, 21, 32, 38],
    previous: [5, 9, 7, 14, 11, 18, 22],
    currentName: "This week",
    previousName: "Last week",
  },
};

function toPoints(values: number[], max: number) {
  const step = CHART.w / (values.length - 1);
  return values.map((v, i) => ({
    x: i * step,
    y: CHART.bottom - (v / max) * (CHART.bottom - CHART.top),
  }));
}

export function WeekCard() {
  const [range, setRange] = useState<RangeKey>("7D");
  const [hover, setHover] = useState<number | null>(null);

  const data = RANGES[range];
  const max = Math.max(...data.current, ...data.previous);
  const cur = toPoints(data.current, max);
  const prev = toPoints(data.previous, max);
  const polyline = (pts: { x: number; y: number }[]) =>
    pts.map((p) => `${p.x},${p.y}`).join(" ");

  const onMove = (e: React.PointerEvent<SVGSVGElement>) => {
    const rect = e.currentTarget.getBoundingClientRect();
    const x = ((e.clientX - rect.left) / rect.width) * CHART.w;
    const step = CHART.w / (data.labels.length - 1);
    setHover(Math.min(data.labels.length - 1, Math.max(0, Math.round(x / step))));
  };

  return (
    <Card className="bg-card overflow-hidden p-5">
      <div className="flex items-center justify-between">
        <p className="text-sm font-medium">{data.currentName}</p>
        <div className="flex gap-1">
          {(["1D", "7D"] as const).map((t) => (
            <button
              key={t}
              type="button"
              aria-pressed={range === t}
              onClick={() => {
                setRange(t);
                setHover(null);
              }}
              className={`rounded-md border px-2 py-0.5 text-[11px] transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring ${
                range === t
                  ? "border-foreground/25 bg-secondary"
                  : "border-border text-muted-foreground hover:bg-secondary/60"
              }`}
            >
              {t}
            </button>
          ))}
        </div>
      </div>

      <div className="mt-4 grid grid-cols-3 gap-3">
        {data.stats.map(([k, v]) => (
          <div key={k} className="rounded-lg border border-border bg-background p-3">
            <p className="text-[11px] text-muted-foreground">{k}</p>
            <p className="mt-1 text-xl font-semibold tracking-tight">
              <AnimatedNumber value={v} />
            </p>
          </div>
        ))}
      </div>

      <div className="mt-4 rounded-lg border border-border bg-background p-4">
        <div className="relative">
          <svg
            viewBox={`0 0 ${CHART.w} ${CHART.h}`}
            className="h-24 w-full cursor-crosshair"
            role="img"
            aria-label={`${data.currentName} volume vs ${data.previousName.toLowerCase()}`}
            onPointerMove={onMove}
            onPointerLeave={() => setHover(null)}
          >
            {hover !== null ? (
              <line
                x1={cur[hover].x}
                x2={cur[hover].x}
                y1={CHART.top - 6}
                y2={CHART.bottom + 4}
                stroke="hsl(var(--border))"
                strokeWidth="1"
              />
            ) : null}
            <polyline
              points={polyline(prev)}
              fill="none"
              stroke="hsl(var(--muted-foreground))"
              strokeWidth="1.5"
              strokeDasharray="3 3"
              opacity="0.5"
            />
            <polyline
              points={polyline(cur)}
              fill="none"
              stroke="hsl(var(--brand))"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
            {hover !== null ? (
              <>
                <circle
                  cx={prev[hover].x}
                  cy={prev[hover].y}
                  r="3.5"
                  fill="hsl(var(--muted-foreground))"
                  stroke="hsl(var(--background))"
                  strokeWidth="2"
                />
                <circle
                  cx={cur[hover].x}
                  cy={cur[hover].y}
                  r="4"
                  fill="hsl(var(--brand))"
                  stroke="hsl(var(--background))"
                  strokeWidth="2"
                />
              </>
            ) : null}
          </svg>

          {hover !== null ? (
            <div
              className="pointer-events-none absolute top-0 z-10 -translate-y-1 rounded-md border border-border bg-card px-2.5 py-1.5 text-[11px] shadow-md"
              style={{
                left: `${(cur[hover].x / CHART.w) * 100}%`,
                transform: `translateX(${cur[hover].x > CHART.w * 0.7 ? "-100%" : cur[hover].x < CHART.w * 0.15 ? "0" : "-50%"})`,
              }}
            >
              <p className="font-medium text-foreground">{data.labels[hover]}</p>
              <p className="mt-0.5 whitespace-nowrap text-muted-foreground">
                {data.currentName}: <span className="font-medium text-foreground">{data.current[hover]}</span>
              </p>
              <p className="whitespace-nowrap text-muted-foreground">
                {data.previousName}: <span className="font-medium text-foreground">{data.previous[hover]}</span>
              </p>
            </div>
          ) : null}
        </div>

        <div className="mt-2 flex justify-between font-mono text-[10px] text-muted-foreground">
          {data.labels.map((d) => (
            <span key={d}>{d}</span>
          ))}
        </div>

        <div className="mt-3 flex items-center gap-4 text-[11px] text-muted-foreground">
          <span className="flex items-center gap-1.5">
            <span className="h-0.5 w-4 rounded-full bg-brand" />
            {data.currentName}
          </span>
          <span className="flex items-center gap-1.5">
            <span className="h-px w-4 border-t border-dashed border-muted-foreground" />
            {data.previousName}
          </span>
        </div>
      </div>
    </Card>
  );
}

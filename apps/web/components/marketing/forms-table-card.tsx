"use client";

import { ArrowDown, ArrowUp, Inbox } from "lucide-react";
import { useState } from "react";
import { Card } from "@/components/ui/card";

type Row = { slug: string; messages: number; spam: number; lastMin: number };

const ROWS: Row[] = [
  { slug: "portfolio", messages: 142, spam: 0.4, lastMin: 2 },
  { slug: "hire-me", messages: 38, spam: 1.1, lastMin: 60 },
  { slug: "support", messages: 512, spam: 3.8, lastMin: 4 },
  { slug: "newsletter", messages: 96, spam: 0.9, lastMin: 22 },
  { slug: "feedback", messages: 27, spam: 0.0, lastMin: 180 },
];

const COLUMNS: { key: SortKey; label: string; align: string }[] = [
  { key: "slug", label: "Form", align: "text-left" },
  { key: "messages", label: "Messages", align: "text-right" },
  { key: "spam", label: "Spam", align: "text-right" },
  { key: "lastMin", label: "Last", align: "text-right" },
];

type SortKey = "slug" | "messages" | "spam" | "lastMin";

function ago(min: number) {
  return min < 60 ? `${min}m ago` : `${Math.round(min / 60)}h ago`;
}

/** Demo forms table: sortable columns, hover + selectable rows. */
export function FormsTableCard() {
  const [sort, setSort] = useState<{ key: SortKey; dir: 1 | -1 }>({ key: "messages", dir: -1 });
  const [selected, setSelected] = useState<string | null>(null);

  const rows = [...ROWS].sort((a, b) => {
    const va = a[sort.key];
    const vb = b[sort.key];
    const cmp = typeof va === "string" ? va.localeCompare(vb as string) : (va as number) - (vb as number);
    return cmp * sort.dir;
  });

  const toggleSort = (key: SortKey) =>
    setSort((s) => (s.key === key ? { key, dir: s.dir === 1 ? -1 : 1 } : { key, dir: key === "slug" ? 1 : -1 }));

  return (
    <Card className="bg-card order-2 overflow-hidden p-0 lg:order-1">
      <div className="grid grid-cols-4 border-b border-border px-5 py-2 text-[11px] uppercase tracking-wide text-muted-foreground">
        {COLUMNS.map((c) => (
          <button
            key={c.key}
            type="button"
            onClick={() => toggleSort(c.key)}
            aria-label={`Sort by ${c.label}`}
            className={`flex items-center gap-1 py-1 uppercase tracking-wide transition-colors hover:text-foreground focus-visible:outline-none focus-visible:text-foreground ${
              c.align === "text-right" ? "justify-end" : ""
            }`}
          >
            {c.label}
            {sort.key === c.key ? (
              sort.dir === 1 ? (
                <ArrowUp className="size-3" />
              ) : (
                <ArrowDown className="size-3" />
              )
            ) : null}
          </button>
        ))}
      </div>
      {rows.map((r) => (
        <button
          key={r.slug}
          type="button"
          onClick={() => setSelected(selected === r.slug ? null : r.slug)}
          aria-pressed={selected === r.slug}
          className={`grid w-full grid-cols-4 items-center border-b border-border px-5 py-3.5 text-left text-sm transition-colors last:border-0 focus-visible:outline-none ${
            selected === r.slug ? "bg-brand/10" : "hover:bg-secondary/50"
          }`}
        >
          <span className="flex items-center gap-2 font-mono text-xs">
            <Inbox className="size-3.5 text-muted-foreground" />
            {r.slug}
          </span>
          <span className="text-right font-medium">{r.messages}</span>
          <span className="text-right text-muted-foreground">{r.spam.toFixed(1)}%</span>
          <span className="text-right text-xs text-muted-foreground">{ago(r.lastMin)}</span>
        </button>
      ))}
    </Card>
  );
}

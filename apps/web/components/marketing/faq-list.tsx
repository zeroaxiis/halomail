"use client";

import { ChevronDown } from "lucide-react";
import { useState } from "react";

/** Accordion FAQ — one item open at a time, animated with grid-template-rows. */
export function FaqList({ items }: { items: { q: string; a: string }[] }) {
  const [open, setOpen] = useState<number | null>(0);

  return (
    <div className="mt-14 border-t border-border">
      {items.map((f, i) => {
        const isOpen = open === i;
        return (
          <div key={f.q} className="border-b border-border">
            <button
              type="button"
              aria-expanded={isOpen}
              onClick={() => setOpen(isOpen ? null : i)}
              className="flex w-full items-center justify-between gap-4 py-5 text-left focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
            >
              <span className="font-medium tracking-tight">{f.q}</span>
              <ChevronDown
                className={`size-4 shrink-0 text-muted-foreground transition-transform duration-300 ${
                  isOpen ? "rotate-180" : ""
                }`}
              />
            </button>
            <div
              className={`grid transition-[grid-template-rows] duration-300 ease-out ${
                isOpen ? "grid-rows-[1fr]" : "grid-rows-[0fr]"
              }`}
            >
              <div className="overflow-hidden">
                <p className="max-w-3xl pb-5 text-sm leading-relaxed text-muted-foreground">
                  {f.a}
                </p>
              </div>
            </div>
          </div>
        );
      })}
    </div>
  );
}

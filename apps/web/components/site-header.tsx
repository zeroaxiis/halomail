"use client";

import Link from "next/link";
import { Logo } from "@/components/logo";
import { ThemeToggle } from "@/components/theme-toggle";
import { Button } from "@/components/ui/button";

export function SiteHeader() {
  return (
    <header className="sticky top-0 z-40 border-b border-border bg-background/70 backdrop-blur">
      <div className="container flex h-14 items-center justify-between">
        <div className="flex items-center gap-8">
          <Link href="/" aria-label="HaloMail home">
            <Logo />
          </Link>
          <nav className="hidden gap-6 text-sm text-muted-foreground md:flex">
            <Link href="/pricing" className="transition-colors hover:text-foreground">Pricing</Link>
            {/* Static page copied from apps/docs by scripts/copy-docs.mjs, so it
                is a plain <a>, not a Next route. */}
            <a href="/docs/integrate.html" className="transition-colors hover:text-foreground">
              Docs
            </a>
          </nav>
        </div>
        <div className="flex items-center gap-2">
          <ThemeToggle />
          <Button asChild variant="ghost" size="sm">
            <Link href="/login">Log in</Link>
          </Button>
          <Button asChild size="sm">
            <Link href="/register">Get started</Link>
          </Button>
        </div>
      </div>
    </header>
  );
}

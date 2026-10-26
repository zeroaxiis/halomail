"use client";

import { useState } from "react";
import Link from "next/link";
import { Menu, X } from "lucide-react";
import { Logo } from "@/components/logo";
import { ThemeToggle } from "@/components/theme-toggle";
import { Button } from "@/components/ui/button";

export function SiteHeader() {
  const [mobileOpen, setMobileOpen] = useState(false);

  return (
    <header className="sticky top-0 z-40 border-b border-border bg-background/80 backdrop-blur-md">
      <div className="container flex h-14 items-center justify-between">
        <div className="flex items-center gap-6 lg:gap-8">
          <Link
            href="/"
            aria-label="HaloMail home"
            className="inline-flex items-center transition-opacity hover:opacity-90 shrink-0"
          >
            <Logo />
          </Link>
          <nav className="hidden items-center gap-6 text-sm text-muted-foreground md:flex">
            <Link href="/pricing" className="transition-colors hover:text-foreground">Pricing</Link>
            <a href="/docs/integrate.html" className="transition-colors hover:text-foreground">
              Docs
            </a>
          </nav>
        </div>

        <div className="flex items-center gap-1.5 sm:gap-2">
          <ThemeToggle />
          
          <div className="hidden sm:flex items-center gap-2">
            <Button asChild variant="ghost" size="sm">
              <Link href="/login">Log in</Link>
            </Button>
            <Button asChild size="sm">
              <Link href="/register">Get started</Link>
            </Button>
          </div>

          <Button asChild size="sm" className="sm:hidden px-2.5 h-8 text-xs">
            <Link href="/register">Sign up</Link>
          </Button>

          {/* Mobile hamburger menu toggle */}
          <button
            type="button"
            aria-label={mobileOpen ? "Close menu" : "Open menu"}
            onClick={() => setMobileOpen(!mobileOpen)}
            className="flex size-9 items-center justify-center rounded-md border border-border text-muted-foreground transition-colors hover:bg-secondary/60 hover:text-foreground md:hidden focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
          >
            {mobileOpen ? <X className="size-4" /> : <Menu className="size-4" />}
          </button>
        </div>
      </div>

      {/* Mobile Drawer Dropdown */}
      {mobileOpen && (
        <div className="border-b border-border bg-background/95 px-6 py-5 backdrop-blur-md md:hidden animate-fade-up">
          <nav className="flex flex-col gap-4 text-sm font-medium">
            <Link
              href="/pricing"
              onClick={() => setMobileOpen(false)}
              className="py-1 text-muted-foreground transition-colors hover:text-foreground"
            >
              Pricing
            </Link>
            <a
              href="/docs/integrate.html"
              onClick={() => setMobileOpen(false)}
              className="py-1 text-muted-foreground transition-colors hover:text-foreground"
            >
              Documentation
            </a>
            <Link
              href="/#features"
              onClick={() => setMobileOpen(false)}
              className="py-1 text-muted-foreground transition-colors hover:text-foreground"
            >
              Features
            </Link>
            <div className="my-2 border-t border-border pt-4 flex flex-col gap-2.5">
              <Button asChild variant="outline" className="w-full justify-center">
                <Link href="/login" onClick={() => setMobileOpen(false)}>Log in</Link>
              </Button>
              <Button asChild className="w-full justify-center">
                <Link href="/register" onClick={() => setMobileOpen(false)}>Get started free</Link>
              </Button>
            </div>
          </nav>
        </div>
      )}
    </header>
  );
}

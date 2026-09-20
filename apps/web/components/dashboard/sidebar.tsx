"use client";

import { CalendarClock, Inbox, LayoutDashboard, Settings } from "lucide-react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { Logo } from "@/components/logo";
import { cn } from "@/lib/utils";

const nav = [
  { href: "/dashboard", label: "Overview", icon: LayoutDashboard },
  { href: "/dashboard/meetings", label: "Meetings", icon: CalendarClock },
  { href: "/dashboard/forms", label: "Forms", icon: Inbox },
  { href: "/dashboard/settings", label: "Settings", icon: Settings },
];

export function Sidebar() {
  const pathname = usePathname();
  return (
    <aside className="w-16 shrink-0 border-r border-border md:w-60">
      <div className="flex h-14 items-center border-b border-border px-5">
        <Link href="/dashboard" aria-label="HaloMail dashboard">
          <span className="md:hidden font-bold">H</span><span className="hidden md:inline"><Logo /></span>
        </Link>
      </div>
      <nav className="flex-1 space-y-1 p-3">
        {nav.map((item) => {
          const active = item.href === "/dashboard" ? pathname === item.href : pathname.startsWith(item.href);
          const Icon = item.icon;
          return (
            <Link
              key={item.href}
              href={item.href}
              aria-label={item.label}
              title={item.label}
              className={cn(
                "flex items-center gap-3 rounded-md px-3 py-2 text-sm transition-colors",
                active ? "bg-secondary font-medium text-foreground" : "text-muted-foreground hover:bg-secondary/60 hover:text-foreground",
              )}
            >
              <Icon className="size-4" />
              <span className="hidden md:inline">{item.label}</span>
            </Link>
          );
        })}
      </nav>
    </aside>
  );
}

"use client";

import { LogOut } from "lucide-react";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { ThemeToggle } from "@/components/theme-toggle";
import { Button } from "@/components/ui/button";
import { clearSession, getUser, type SessionUser } from "@/lib/auth";
import { rpc } from "@/lib/api";

export function Topbar() {
  const router = useRouter();
  const [user, setUser] = useState<SessionUser | null>(null);
  useEffect(() => setUser(getUser()), []);

  async function logout() {
    await rpc("halomail.identity.v1.AuthService/Logout").catch(() => {});
    clearSession();
    router.replace("/login");
  }

  return (
    <header className="flex h-14 items-center justify-end gap-3 border-b border-border px-6">
      <span className="hidden text-sm text-muted-foreground sm:inline">{user?.email}</span>
      <ThemeToggle />
      <Button variant="ghost" size="icon" aria-label="Log out" onClick={logout}>
        <LogOut className="size-4" />
      </Button>
    </header>
  );
}

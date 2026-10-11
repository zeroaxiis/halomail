"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { LoginShell } from "@/components/auth/login-shell";
import { Logo } from "@/components/logo";

export default function AuthLayout({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();

  if (pathname === "/login") return <LoginShell>{children}</LoginShell>;

  return (
    <div className="flex min-h-screen flex-col items-center justify-center px-4">
      <Link href="/" className="mb-8"><Logo className="text-lg" /></Link>
      <div className="w-full max-w-sm">{children}</div>
      <p className="mt-8 text-xs text-muted-foreground">© {new Date().getFullYear()} HaloMail</p>
    </div>
  );
}

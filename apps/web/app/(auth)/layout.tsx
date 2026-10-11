"use client";

import { usePathname } from "next/navigation";
import { LoginShell } from "@/components/auth/login-shell";

export default function AuthLayout({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();

  return <LoginShell isRegister={pathname === "/register"}>{children}</LoginShell>;
}

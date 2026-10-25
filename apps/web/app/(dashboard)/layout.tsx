"use client";

import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { Topbar } from "@/components/dashboard/topbar";
import { saveSession, clearSession, type SessionUser } from "@/lib/auth";
import { rpc } from "@/lib/api";
import { WorkspaceProvider } from "@/components/dashboard/workspace-context";
import "./dashboard.css";

export default function DashboardLayout({
  children,
}: { children: React.ReactNode }) {
  const router = useRouter();
  const [user, setUser] = useState<SessionUser | null>(null);

  useEffect(() => {
    rpc<{ user: SessionUser }>(
      "halomail.identity.v1.AuthService/GetCurrentUser",
    )
      .then((result) => {
        saveSession("", result.user);
        setUser(result.user);
      })
      .catch(() => {
        clearSession();
        router.replace("/login");
      });
  }, [router]);

  if (!user)
    return (
      <div className="dashboard-shell workspace-loading" role="status">
        <span className="workspace-wordmark">HaloMail</span>
        <p>Opening your workspace…</p>
      </div>
    );

  return (
    <WorkspaceProvider initialUser={user}>
      <div className="dashboard-shell min-h-screen">
        <a href="#workspace-content" className="dashboard-skip-link">
          Skip to content
        </a>
        <Topbar />
        <main id="workspace-content" tabIndex={-1} className="dashboard-main">
          {children}
        </main>
      </div>
    </WorkspaceProvider>
  );
}

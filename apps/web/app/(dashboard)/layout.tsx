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

  const [cancellingDeletion, setCancellingDeletion] = useState(false);

  async function handleCancelDeletion() {
    if (!user || cancellingDeletion) return;
    setCancellingDeletion(true);
    try {
      await rpc("halomail.identity.v1.UserService/CancelDeleteAccount");
      const updated = { ...user, scheduledDeleteAt: undefined, deletionScheduledAt: undefined };
      setUser(updated);
      saveSession("", updated);
    } catch {
      // ignore
    } finally {
      setCancellingDeletion(false);
    }
  }

  if (!user)
    return (
      <div className="dashboard-shell workspace-loading" role="status">
        <span className="workspace-wordmark">HaloMail</span>
        <p>Opening your workspace…</p>
      </div>
    );

  const isDeletionPending = !!user.scheduledDeleteAt;

  return (
    <WorkspaceProvider initialUser={user}>
      <div className="dashboard-shell min-h-screen">
        <a href="#workspace-content" className="dashboard-skip-link">
          Skip to content
        </a>
        {isDeletionPending && (
          <div className="bg-amber-500/15 border-b border-amber-500/30 px-4 py-2.5 text-xs text-amber-700 dark:text-amber-300 flex flex-wrap items-center justify-between gap-3 sticky top-0 z-50 backdrop-blur-md">
            <div className="flex items-center gap-2">
              <span className="font-semibold uppercase tracking-wider text-[10px] bg-amber-500/20 text-amber-700 dark:text-amber-200 px-2 py-0.5 rounded border border-amber-500/30">
                Notice
              </span>
              <span>
                Your account is scheduled for permanent deletion in 24 hours. Your data will be permanently wiped unless cancelled.
              </span>
            </div>
            <button
              type="button"
              onClick={handleCancelDeletion}
              disabled={cancellingDeletion}
              className="px-3 py-1 text-xs font-semibold rounded bg-amber-500 text-amber-950 hover:bg-amber-400 disabled:opacity-50 transition shadow-sm"
            >
              {cancellingDeletion ? "Cancelling…" : "Cancel Deletion"}
            </button>
          </div>
        )}
        <Topbar />
        <main id="workspace-content" tabIndex={-1} className="dashboard-main">
          {children}
        </main>
      </div>
    </WorkspaceProvider>
  );
}

"use client";

import { createContext, useContext, useEffect, useState } from "react";
import { type SessionUser, saveSession } from "@/lib/auth";
import { useRpc } from "@/lib/use-rpc";

export interface Allowance {
  used: number;
  limit: number;
  remaining: number;
  resetsAt: string | null;
}
export interface WorkspaceUsage {
  plan: string;
  expiresAt?: string;
  forms: Allowance;
  meetings: Allowance;
}

const WorkspaceContext = createContext<{
  user: SessionUser;
  updateUser: (user: SessionUser) => void;
  usage: ReturnType<typeof useRpc<WorkspaceUsage>>;
} | null>(null);

export function WorkspaceProvider({
  initialUser,
  children,
}: { initialUser: SessionUser; children: React.ReactNode }) {
  const [user, setUser] = useState(initialUser);
  const usage = useRpc<WorkspaceUsage>("v1/usage");
  const reload = usage.reload;

  useEffect(() => {
    const refresh = () => {
      void reload();
    };
    window.addEventListener("billing-updated", refresh);
    window.addEventListener("focus", refresh);
    return () => {
      window.removeEventListener("billing-updated", refresh);
      window.removeEventListener("focus", refresh);
    };
  }, [reload]);

  function updateUser(next: SessionUser) {
    saveSession("", next);
    setUser(next);
  }

  return (
    <WorkspaceContext.Provider value={{ user, updateUser, usage }}>
      {children}
    </WorkspaceContext.Provider>
  );
}

export function useWorkspace() {
  const context = useContext(WorkspaceContext);
  if (!context)
    throw new Error("Workspace components must be inside WorkspaceProvider");
  return context;
}

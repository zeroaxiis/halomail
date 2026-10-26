"use client";

const TOKEN_KEY = "halomail_token";
const USER_KEY = "halomail_user";

export interface SessionUser {
  id: string;
  email: string;
  name: string;
  handle: string;
  org_id: string;
  role?: string;
  timezone?: string;
  scheduledDeleteAt?: { seconds: string } | string;
  deletionScheduledAt?: { seconds: string } | string;
}

export function saveSession(token: string, user: SessionUser) {
  void token;
  if (typeof window !== "undefined") {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.setItem(USER_KEY, JSON.stringify(user));
  }
}

export function getToken(): string | null {
  return null;
}

export function getUser(): SessionUser | null {
  if (typeof window === "undefined") return null;
  const raw = localStorage.getItem(USER_KEY);
  try { return raw ? (JSON.parse(raw) as SessionUser) : null; }
  catch { return null; }
}

export function hasStoredSession(): boolean {
  return getUser() !== null;
}

export function clearSession() {
  if (typeof window !== "undefined") {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(USER_KEY);
  }
}


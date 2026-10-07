"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { rpc } from "@/lib/api";
import { saveSession, type SessionUser } from "@/lib/auth";

export default function LoginPage() {
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [challengeId, setChallengeId] = useState("");
  const [code, setCode] = useState("");
  const [expiresAt, setExpiresAt] = useState(0);
  const [resendAt, setResendAt] = useState(0);
  const [now, setNow] = useState(0);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!challengeId) return;
    const timer = setInterval(() => setNow(Date.now()), 1000);
    return () => clearInterval(timer);
  }, [challengeId]);

  const remaining = Math.max(0, Math.ceil((expiresAt - now) / 1000));
  const resendIn = Math.max(0, Math.ceil((resendAt - now) / 1000));

  async function requestCode() {
    setError(null);
    setLoading(true);
    try {
      const startedAt = Date.now();
      const res = await rpc<{ challengeId: string; expiresIn: number }>("v1/auth/otp/request", { email, password });
      setChallengeId(res.challengeId);
      setCode("");
      setNow(Date.now());
      setExpiresAt(startedAt + res.expiresIn * 1000);
      setResendAt(Date.now() + 60000);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not send the code");
    } finally {
      setLoading(false);
    }
  }

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!challengeId) { await requestCode(); return; }
    setError(null);
    setLoading(true);
    try {
      const res = await rpc<{ user: SessionUser }>("v1/auth/otp/verify", { challengeId, code });
      saveSession("", res.user);
      setPassword("");
      router.replace("/dashboard");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Verification failed");
    } finally {
      setLoading(false);
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-xl">{challengeId ? "Check your email" : "Welcome back"}</CardTitle>
        <CardDescription>
          {challengeId ? `Enter the 6-digit code sent to ${email}.` : "Enter your password, then verify your email to log in."}
        </CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={onSubmit} className="space-y-4">
          {challengeId ? (
            <div className="space-y-1.5">
              <Label htmlFor="code">Verification code</Label>
              <Input id="code" type="text" inputMode="numeric" autoComplete="one-time-code" autoFocus required
                pattern="[0-9]{6}" minLength={6} maxLength={6} value={code} disabled={loading || remaining === 0}
                onChange={(e) => setCode(e.target.value.replace(/\D/g, ""))} placeholder="000000" />
              <p className="text-sm text-muted-foreground">
                {remaining > 0 ? `Code expires in ${Math.floor(remaining / 60)}:${String(remaining % 60).padStart(2, "0")}.` : "Your code has expired. Request a new one below."}
              </p>
            </div>
          ) : (
            <>
              <div className="space-y-1.5">
                <Label htmlFor="email">Email</Label>
                <Input id="email" type="email" autoComplete="email" required disabled={loading} value={email} onChange={(e) => setEmail(e.target.value)} placeholder="you@example.com" />
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="password">Password</Label>
                <Input id="password" type="password" autoComplete="current-password" required disabled={loading} value={password} onChange={(e) => setPassword(e.target.value)} placeholder="Your password" />
              </div>
            </>
          )}
          {error && <p role="alert" className="text-sm text-destructive">{error}</p>}
          <Button type="submit" className="w-full" disabled={loading || (!!challengeId && (remaining === 0 || code.length !== 6))}>
            {loading ? "Please wait..." : challengeId ? "Verify and log in" : "Send login code"}
          </Button>
          {challengeId && (
            <div className="flex justify-between gap-2">
              <Button type="button" variant="ghost" disabled={loading || resendIn > 0} onClick={requestCode}>
                {resendIn > 0 ? `Resend in ${resendIn}s` : "Resend code"}
              </Button>
              <Button type="button" variant="ghost" disabled={loading} onClick={() => { setChallengeId(""); setCode(""); setPassword(""); setError(null); }}>
                Change account
              </Button>
            </div>
          )}
        </form>
        <p className="mt-4 text-center text-sm text-muted-foreground">
          No account? <Link href="/register" className="text-foreground underline-offset-4 hover:underline">Create one</Link>
        </p>
      </CardContent>
    </Card>
  );
}

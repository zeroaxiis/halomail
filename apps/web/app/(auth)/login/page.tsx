"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import styles from "@/components/auth/login.module.css";
import { rpc } from "@/lib/api";
import { saveSession, type SessionUser } from "@/lib/auth";

export default function LoginPage() {
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
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
    <div className={styles.form}>
      <h2 className={styles.formTitle}>{challengeId ? "Check your email." : "Welcome back."}</h2>
      <p className={styles.formDescription}>
        {challengeId ? `Enter the 6-digit code we sent to ${email}.` : "Sign in to your HaloMail workspace."}
      </p>
        <form onSubmit={onSubmit} className={styles.fields} aria-busy={loading}>
          {challengeId ? (
            <div className={styles.field}>
              <Label htmlFor="code">Verification code</Label>
              <Input id="code" type="text" inputMode="numeric" autoComplete="one-time-code" autoFocus required
                pattern="[0-9]{6}" minLength={6} maxLength={6} value={code} disabled={loading || remaining === 0}
                className={styles.codeInput} aria-describedby="code-help"
                onChange={(e) => setCode(e.target.value.replace(/\D/g, ""))} placeholder="000000" />
              <p id="code-help" className={styles.codeHelp}>
                {remaining > 0 ? `Code expires in ${Math.floor(remaining / 60)}:${String(remaining % 60).padStart(2, "0")}.` : "Your code has expired. Request a new one below."}
              </p>
            </div>
          ) : (
            <>
              <div className={styles.field}>
                <Label htmlFor="email">Email address</Label>
                <Input id="email" type="email" autoComplete="email" required disabled={loading} value={email} onChange={(e) => setEmail(e.target.value)} placeholder="you@example.com" />
              </div>
              <div className={styles.field}>
                <Label htmlFor="password">Password</Label>
                <div className={styles.passwordWrap}>
                  <Input id="password" type={showPassword ? "text" : "password"} autoComplete="current-password" required disabled={loading} value={password} onChange={(e) => setPassword(e.target.value)} placeholder="Enter your password" />
                  <button type="button" className={styles.passwordToggle} disabled={loading}
                    aria-label={showPassword ? "Hide password" : "Show password"}
                    aria-pressed={showPassword} aria-controls="password"
                    onClick={() => setShowPassword((visible) => !visible)}>
                    {showPassword ? "Hide" : "Show"}
                  </button>
                </div>
              </div>
            </>
          )}
          {error && <p role="alert" className={styles.error}>{error}</p>}
          <Button type="submit" className={styles.submit} disabled={loading || (!!challengeId && (remaining === 0 || code.length !== 6))}>
            {loading ? "Please wait..." : challengeId ? "Verify and log in" : "Send login code"}
          </Button>
          {challengeId && (
            <div className={styles.codeActions}>
              <Button type="button" variant="ghost" disabled={loading || resendIn > 0} onClick={requestCode}>
                {resendIn > 0 ? `Resend in ${resendIn}s` : "Resend code"}
              </Button>
              <Button type="button" variant="ghost" disabled={loading} onClick={() => { setChallengeId(""); setCode(""); setPassword(""); setShowPassword(false); setError(null); }}>
                Change account
              </Button>
            </div>
          )}
        </form>
        {!challengeId && <p className={styles.formNote}>We&apos;ll email you a verification code after this step.</p>}
        <p className={styles.formRegister}>
          New to HaloMail? <Link href="/register">Create your account</Link>
        </p>
    </div>
  );
}

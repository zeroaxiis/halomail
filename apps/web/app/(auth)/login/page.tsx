"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import styles from "@/components/auth/login.module.css";
import { GoogleSignInButton, AuthDivider } from "@/components/auth/google-button";
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
  const [checkingAuth, setCheckingAuth] = useState(true);

  // If already authenticated with a valid session, redirect immediately to dashboard
  useEffect(() => {
    let cancelled = false;
    rpc<{ user: SessionUser }>("halomail.identity.v1.AuthService/GetCurrentUser")
      .then((res) => {
        if (cancelled) return;
        saveSession("", res.user);
        const searchParams = new URLSearchParams(window.location.search);
        const redirect = searchParams.get("redirect") || "/dashboard";
        router.replace(redirect);
      })
      .catch(() => {
        if (!cancelled) {
          setCheckingAuth(false);
        }
      });
    return () => {
      cancelled = true;
    };
  }, [router]);

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
    if (!challengeId) {
      // Pre-check: if a valid session exists, redirect immediately instead of requesting OTP
      setLoading(true);
      try {
        const active = await rpc<{ user: SessionUser }>("halomail.identity.v1.AuthService/GetCurrentUser");
        if (active?.user) {
          saveSession("", active.user);
          const searchParams = new URLSearchParams(window.location.search);
          router.replace(searchParams.get("redirect") || "/dashboard");
          return;
        }
      } catch {
        // Not authenticated, proceed with credentials + OTP
      } finally {
        setLoading(false);
      }
      await requestCode();
      return;
    }
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

  useEffect(() => {
    const searchParams = new URLSearchParams(window.location.search);
    const err = searchParams.get("error");
    if (err) {
      const messages: Record<string, string> = {
        google_not_configured: "Google sign-in is not configured yet. Please configure GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET.",
        google_cancelled: "Google sign-in was cancelled.",
        unverified_email: "Your Google email address is unverified.",
        invalid_state: "Your sign-in session expired. Please try again.",
        missing_code: "Missing authorization code from Google.",
        token_exchange_failed: "Could not exchange authorization code with Google.",
        login_failed: "Could not sign in with Google. Please try again.",
      };
      setError(messages[err] || decodeURIComponent(err));
    }
  }, []);

  if (checkingAuth) {
    return (
      <div className={styles.form}>
        <h2 className={styles.formTitle}>Welcome back.</h2>
        <p className={styles.formDescription}>Verifying your session…</p>
      </div>
    );
  }

  return (
    <div className={styles.form}>
      <h2 className={styles.formTitle}>{challengeId ? "Check your email." : "Welcome back."}</h2>
      <p className={styles.formDescription}>
        {challengeId ? `Enter the 6-digit code we sent to ${email}.` : "Sign in to your HaloMail workspace."}
      </p>
      {!challengeId && (
        <>
          <GoogleSignInButton text="Continue with Google" />
          <AuthDivider text="or continue with email" />
        </>
      )}
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
                <div className={styles.fieldHeader}>
                  <Label htmlFor="password">Password</Label>
                  <Link href="/forgot-password" className={styles.forgotLink}>
                    Forgot password?
                  </Link>
                </div>
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

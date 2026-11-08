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

export default function RegisterPage() {
  const router = useRouter();
  const [name, setName] = useState("");
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

  async function requestResend() {
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
      setError(err instanceof Error ? err.message : "Could not resend the code");
    } finally {
      setLoading(false);
    }
  }

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!challengeId) {
      setError(null);
      setLoading(true);
      try {
        // Step 1: Create the account in database
        await rpc("halomail.identity.v1.AuthService/Register", { name, email, password });

        // Step 2: Immediately trigger OTP email delivery on the same page
        const startedAt = Date.now();
        const otpRes = await rpc<{ challengeId: string; expiresIn: number }>("v1/auth/otp/request", { email, password });
        setChallengeId(otpRes.challengeId);
        setCode("");
        setNow(Date.now());
        setExpiresAt(startedAt + otpRes.expiresIn * 1000);
        setResendAt(Date.now() + 60000);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Registration failed");
      } finally {
        setLoading(false);
      }
      return;
    }

    // Step 3: Verify the 6-digit code and directly log into dashboard
    setError(null);
    setLoading(true);
    try {
      const res = await rpc<{ user: SessionUser }>("v1/auth/otp/verify", { challengeId, code });
      saveSession("", res.user);
      setPassword("");
      router.replace("/dashboard");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Verification failed. Check the code and try again.");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className={`${styles.form} ${styles.registerForm}`}>
      <h2 className={styles.formTitle}>
        {challengeId ? "Check your email." : "Create your account."}
      </h2>
      <p className={styles.formDescription}>
        {challengeId
          ? `Enter the 6-digit code we sent to ${email} to enter your workspace.`
          : "Start your HaloMail workspace."}
      </p>

      {!challengeId && (
        <>
          <GoogleSignInButton text="Sign up with Google" />
          <AuthDivider text="or sign up with email" />
        </>
      )}

      <form onSubmit={onSubmit} className={styles.fields} aria-busy={loading}>
        {challengeId ? (
          <div className={styles.field}>
            <Label htmlFor="code">Verification code</Label>
            <Input
              id="code"
              type="text"
              inputMode="numeric"
              autoComplete="one-time-code"
              autoFocus
              required
              pattern="[0-9]{6}"
              minLength={6}
              maxLength={6}
              value={code}
              disabled={loading || remaining === 0}
              className={styles.codeInput}
              aria-describedby="code-help"
              onChange={(e) => setCode(e.target.value.replace(/\D/g, ""))}
              placeholder="000000"
            />
            <p id="code-help" className={styles.codeHelp}>
              {remaining > 0
                ? `Code expires in ${Math.floor(remaining / 60)}:${String(remaining % 60).padStart(2, "0")}.`
                : "Your code has expired. Request a new one below."}
            </p>
          </div>
        ) : (
          <>
            <div className={styles.field}>
              <Label htmlFor="name">Name</Label>
              <Input
                id="name"
                autoComplete="name"
                required
                disabled={loading}
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="Your name"
              />
            </div>
            <div className={styles.field}>
              <Label htmlFor="email">Email address</Label>
              <Input
                id="email"
                type="email"
                autoComplete="email"
                required
                disabled={loading}
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="you@example.com"
              />
            </div>
            <div className={styles.field}>
              <Label htmlFor="password">Password</Label>
              <div className={styles.passwordWrap}>
                <Input
                  id="password"
                  type={showPassword ? "text" : "password"}
                  autoComplete="new-password"
                  required
                  minLength={8}
                  disabled={loading}
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="At least 8 characters"
                />
                <button
                  type="button"
                  className={styles.passwordToggle}
                  disabled={loading}
                  aria-label={showPassword ? "Hide password" : "Show password"}
                  aria-pressed={showPassword}
                  aria-controls="password"
                  onClick={() => setShowPassword((visible) => !visible)}
                >
                  {showPassword ? "Hide" : "Show"}
                </button>
              </div>
            </div>
          </>
        )}

        {error && (
          <div role="alert" className={`${styles.error} text-left space-y-1`}>
            <p>{error}</p>
            {error.includes("already a member of a team") && (
              <p className="text-[11px] opacity-90 pt-1">
                To access your team, please <Link href="/login" className="underline font-medium">log in here</Link>. If you want to create your own separate account, leave the team from the Team page first.
              </p>
            )}
            {error.includes("already exists") && (
              <p className="text-[11px] opacity-90 pt-1">
                Please <Link href="/login" className="underline font-medium">click here to log in</Link>.
              </p>
            )}
          </div>
        )}

        <Button
          type="submit"
          className={styles.submit}
          disabled={loading || (!!challengeId && (remaining === 0 || code.length !== 6))}
        >
          {loading ? "Please wait..." : challengeId ? "Verify & enter dashboard" : "Create account"}
        </Button>

        {challengeId && (
          <div className={styles.codeActions}>
            <Button
              type="button"
              variant="ghost"
              disabled={loading || resendIn > 0}
              onClick={requestResend}
            >
              {resendIn > 0 ? `Resend in ${resendIn}s` : "Resend code"}
            </Button>
            <Button
              type="button"
              variant="ghost"
              disabled={loading}
              onClick={() => {
                setChallengeId("");
                setCode("");
                setError(null);
              }}
            >
              Back to edit
            </Button>
          </div>
        )}
      </form>

      {!challengeId && (
        <p className={styles.formNote}>We&apos;ll email you a verification code to confirm your email.</p>
      )}

      <p className={styles.formRegister}>
        Already have an account? <Link href="/login">Log in</Link>
      </p>
    </div>
  );
}

"use client";

import Link from "next/link";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import styles from "@/components/auth/login.module.css";
import { rpc } from "@/lib/api";

export default function ForgotPasswordPage() {
  const [email, setEmail] = useState("");
  const [sent, setSent] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    setLoading(true);

    try {
      await rpc<{ success: boolean }>(
        "halomail.identity.v1.AuthService/ForgotPassword",
        { email }
      );
      setSent(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to send reset link");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className={styles.form}>
      <h2 className={styles.formTitle}>
        {sent ? "Check your email" : "Reset password"}
      </h2>
      <p className={styles.formDescription}>
        {sent
          ? `If an account exists for ${email}, we've sent instructions to reset your password.`
          : "Enter the email associated with your account and we'll send you a link to reset your password."}
      </p>

      {sent ? (
        <div className={styles.fields}>
          <div className={styles.success}>
            Instructions sent! Check your inbox or spam folder for your recovery link. The link expires in 1 hour.
          </div>
          <Button asChild className={styles.submit}>
            <Link href="/login">Return to sign in</Link>
          </Button>
          <div className={styles.codeActions}>
            <Button
              type="button"
              variant="ghost"
              disabled={loading}
              onClick={() => {
                setSent(false);
                setError(null);
              }}
            >
              Try another email
            </Button>
          </div>
        </div>
      ) : (
        <form onSubmit={onSubmit} className={styles.fields} aria-busy={loading}>
          <div className={styles.field}>
            <Label htmlFor="email">Email address</Label>
            <Input
              id="email"
              type="email"
              autoComplete="email"
              required
              autoFocus
              disabled={loading}
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="you@example.com"
            />
          </div>

          {error && (
            <p role="alert" className={styles.error}>
              {error}
            </p>
          )}

          <Button type="submit" className={styles.submit} disabled={loading || !email}>
            {loading ? "Sending link..." : "Send reset link"}
          </Button>
        </form>
      )}

      <p className={styles.formRegister}>
        Remember your password? <Link href="/login">Sign in</Link>
      </p>
    </div>
  );
}

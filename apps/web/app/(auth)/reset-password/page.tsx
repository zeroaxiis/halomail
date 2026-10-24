"use client";

import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { Suspense, useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import styles from "@/components/auth/login.module.css";
import { rpc } from "@/lib/api";

function ResetPasswordForm() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const token = searchParams.get("token") || "";

  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  if (!token) {
    return (
      <div className={styles.form}>
        <h2 className={styles.formTitle}>Invalid reset link</h2>
        <p className={styles.formDescription}>
          This password reset link is invalid or incomplete. Please request a new recovery link.
        </p>
        <Button asChild className={styles.submit}>
          <Link href="/forgot-password">Request new link</Link>
        </Button>
      </div>
    );
  }

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);

    if (password.length < 8) {
      setError("Password must be at least 8 characters long.");
      return;
    }

    if (password !== confirmPassword) {
      setError("Passwords do not match. Please verify and try again.");
      return;
    }

    setLoading(true);

    try {
      await rpc<{ success: boolean }>(
        "halomail.identity.v1.AuthService/ResetPassword",
        { token, newPassword: password }
      );
      setSuccess(true);
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : "Failed to reset password. The link may have expired."
      );
    } finally {
      setLoading(false);
    }
  }

  if (success) {
    return (
      <div className={styles.form}>
        <h2 className={styles.formTitle}>Password updated</h2>
        <p className={styles.formDescription}>
          Your password has been reset successfully. You can now sign in with your new credentials.
        </p>
        <div className={styles.fields}>
          <div className={styles.success}>
            All active sessions have been safely terminated. Please log in to continue.
          </div>
          <Button asChild className={styles.submit}>
            <Link href="/login">Sign in with new password</Link>
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className={styles.form}>
      <h2 className={styles.formTitle}>Set new password</h2>
      <p className={styles.formDescription}>
        Choose a strong password containing at least 8 characters.
      </p>

      <form onSubmit={onSubmit} className={styles.fields} aria-busy={loading}>
        <div className={styles.field}>
          <Label htmlFor="new-password">New password</Label>
          <div className={styles.passwordWrap}>
            <Input
              id="new-password"
              type={showPassword ? "text" : "password"}
              autoComplete="new-password"
              required
              autoFocus
              minLength={8}
              disabled={loading}
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="Minimum 8 characters"
            />
            <button
              type="button"
              className={styles.passwordToggle}
              disabled={loading}
              aria-label={showPassword ? "Hide password" : "Show password"}
              onClick={() => setShowPassword((v) => !v)}
            >
              {showPassword ? "Hide" : "Show"}
            </button>
          </div>
        </div>

        <div className={styles.field}>
          <Label htmlFor="confirm-password">Confirm new password</Label>
          <div className={styles.passwordWrap}>
            <Input
              id="confirm-password"
              type={showPassword ? "text" : "password"}
              autoComplete="new-password"
              required
              minLength={8}
              disabled={loading}
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              placeholder="Re-enter your password"
            />
          </div>
        </div>

        {error && (
          <p role="alert" className={styles.error}>
            {error}
          </p>
        )}

        <Button
          type="submit"
          className={styles.submit}
          disabled={loading || password.length < 8 || !confirmPassword}
        >
          {loading ? "Updating password..." : "Update password"}
        </Button>
      </form>

      <p className={styles.formRegister}>
        Remember your password? <Link href="/login">Sign in</Link>
      </p>
    </div>
  );
}

export default function ResetPasswordPage() {
  return (
    <Suspense
      fallback={
        <div className={styles.form}>
          <h2 className={styles.formTitle}>Loading...</h2>
        </div>
      }
    >
      <ResetPasswordForm />
    </Suspense>
  );
}

"use client";

import Link from "next/link";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import styles from "@/components/auth/login.module.css";
import { rpc } from "@/lib/api";

export default function RegisterPage() {
  const [created, setCreated] = useState(false);
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    setLoading(true);
    try {
      await rpc("halomail.identity.v1.AuthService/Register", { name, email, password });
      setPassword("");
      setCreated(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Registration failed");
    } finally {
      setLoading(false);
    }
  }

  if (created) return (
    <div className={styles.form} aria-live="polite">
      <h2 className={styles.formTitle}>Account created.</h2>
      <p className={styles.formDescription}>Log in with your password and verify the code sent to your email to access your dashboard.</p>
      <Button asChild className={`${styles.submit} w-full`}><Link href="/login">Continue to login</Link></Button>
    </div>
  );

  return (
    <div className={`${styles.form} ${styles.registerForm}`}>
      <h2 className={styles.formTitle}>Create your account.</h2>
      <p className={styles.formDescription}>Start your HaloMail workspace.</p>
        <form onSubmit={onSubmit} className={styles.fields} aria-busy={loading}>
          <div className={styles.field}>
            <Label htmlFor="name">Name</Label>
            <Input id="name" autoComplete="name" required disabled={loading} value={name} onChange={(e) => setName(e.target.value)} placeholder="Your name" />
          </div>
          <div className={styles.field}>
            <Label htmlFor="email">Email address</Label>
            <Input id="email" type="email" autoComplete="email" required disabled={loading} value={email} onChange={(e) => setEmail(e.target.value)} placeholder="you@example.com" />
          </div>
          <div className={styles.field}>
            <Label htmlFor="password">Password</Label>
            <div className={styles.passwordWrap}>
              <Input id="password" type={showPassword ? "text" : "password"} autoComplete="new-password" required minLength={8} disabled={loading} value={password} onChange={(e) => setPassword(e.target.value)} placeholder="At least 8 characters" />
              <button type="button" className={styles.passwordToggle} disabled={loading}
                aria-label={showPassword ? "Hide password" : "Show password"}
                aria-pressed={showPassword} aria-controls="password"
                onClick={() => setShowPassword((visible) => !visible)}>
                {showPassword ? "Hide" : "Show"}
              </button>
            </div>
          </div>
          {error && <p role="alert" className={styles.error}>{error}</p>}
          <Button type="submit" className={styles.submit} disabled={loading}>
            {loading ? "Creating…" : "Create account"}
          </Button>
        </form>
        <p className={styles.formRegister}>
          Already have an account?{" "}
          <Link href="/login">
            Log in
          </Link>
        </p>
    </div>
  );
}

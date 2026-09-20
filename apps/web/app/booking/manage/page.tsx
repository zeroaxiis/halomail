"use client";

import { FormEvent, useEffect, useState } from "react";
import Link from "next/link";
import { rpc } from "@/lib/api";
import { Button } from "@/components/ui/button";

export default function ManageBookingPage() {
  const [token, setToken] = useState("");
  const [reason, setReason] = useState("");
  const [busy, setBusy] = useState(false);
  const [done, setDone] = useState(false);
  const [error, setError] = useState("");
  useEffect(() => { setToken(new URLSearchParams(window.location.search).get("cancel") || ""); }, []);
  async function cancel(event: FormEvent) {
    event.preventDefault(); setBusy(true); setError("");
    try { await rpc("halomail.scheduling.v1.BookingService/CancelBooking", { cancelToken: token, reason }); setDone(true); }
    catch (failure) { setError(failure instanceof Error ? failure.message : "Cancellation failed"); }
    finally { setBusy(false); }
  }
  return <main className="mx-auto max-w-lg space-y-6 px-6 py-20">
    <Link href="/" className="font-semibold">HaloMail</Link>
    <h1 className="text-2xl font-semibold">{done ? "Booking cancelled" : "Cancel your booking"}</h1>
    {done ? <p className="text-muted-foreground">Your time is released. The calendar update and cancellation emails are queued.</p> : <form onSubmit={cancel} className="space-y-4">
      <p className="text-muted-foreground">Confirm below to cancel. Opening this link alone does not change your booking.</p>
      <label className="block text-sm">Reason (optional)<textarea value={reason} onChange={(event) => setReason(event.target.value)} maxLength={1000} className="mt-2 block w-full rounded-md border bg-background p-3" /></label>
      {error && <p role="alert" className="text-sm text-destructive">{error}</p>}
      <Button type="submit" disabled={!token || busy}>{busy ? "Cancelling…" : "Cancel booking"}</Button>
    </form>}
  </main>;
}

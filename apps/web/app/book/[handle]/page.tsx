"use client";

import { Suspense, useEffect, useState } from "react";
import { useSearchParams } from "next/navigation";
import Link from "next/link";
import { Logo } from "@/components/logo";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { rpc } from "@/lib/api";

interface EventType { id: string; title: string; description: string; durationMinutes: number }
interface Slot { start: string; end: string }

function Booking() {
  const key = useSearchParams().get("key") || "";
  const [info, setInfo] = useState<{ owner: { name: string }; events: EventType[] } | null>(null);
  const [eventID, setEventID] = useState("");
  const [slots, setSlots] = useState<Slot[]>([]);
  const [picked, setPicked] = useState<Slot | null>(null);
  const [offset, setOffset] = useState(0);
  const [timezone, setTimezone] = useState("UTC");
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [notes, setNotes] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [done, setDone] = useState(false);

  useEffect(() => {
    setTimezone(Intl.DateTimeFormat().resolvedOptions().timeZone);
    if (!key) { setError("Open the booking button shared by the host. Its meeting access key is required."); setLoading(false); return; }
    rpc<{ owner: { name: string }; events: EventType[] }>("v1/meetings/info", {}, null, key)
      .then(result => { setInfo(result); setEventID(result.events[0]?.id || ""); })
      .catch(failure => setError(failure.message)).finally(() => setLoading(false));
  }, [key]);

  useEffect(() => {
    if (!eventID || !key) return;
    let current = true;
    setLoading(true); setError(""); setPicked(null);
    const from = new Date(Date.now() + offset * 86400000);
    const to = new Date(from.getTime() + 6 * 86400000);
    rpc<{ slots?: Slot[] }>("halomail.scheduling.v1.BookingService/ListSlots", { eventTypeId: eventID, fromDate: from.toISOString().slice(0, 10), toDate: to.toISOString().slice(0, 10), inviteeTimezone: timezone }, null, key)
      .then(result => { if (current) setSlots(result.slots || []); })
      .catch(failure => { if (current) { setSlots([]); setError(failure.message); } })
      .finally(() => { if (current) setLoading(false); });
    return () => { current = false; };
  }, [eventID, key, offset, timezone]);

  async function book(event: React.FormEvent) {
    event.preventDefault();
    if (!picked || busy) return;
    setBusy(true); setError("");
    try {
      await rpc("halomail.scheduling.v1.BookingService/CreateBooking", { eventTypeId: eventID, inviteeName: name, inviteeEmail: email, inviteeTimezone: timezone, start: picked.start, notes }, null, key);
      setDone(true);
    } catch (failure) { setError(failure instanceof Error ? failure.message : "Could not book"); }
    finally { setBusy(false); }
  }

  return <div className="mx-auto max-w-2xl space-y-6 px-4 py-10">
    <Link href="/"><Logo /></Link>
    {done ? <Card className="space-y-3 p-8"><h1 className="text-2xl font-semibold">Your time is reserved</h1><p>We are adding the meeting to the host’s calendar. Your confirmation and Google Meet link will arrive at {email} after the calendar confirms it.</p></Card> : <>
      <h1 className="text-3xl font-semibold">{info ? "Book with " + info.owner.name : "Book a meeting"}</h1>
      {error && <p role="alert" className="text-sm text-destructive">{error}</p>}
      {!!info?.events.length && <div className="flex flex-wrap gap-2">{info.events.map(event => <Button key={event.id} variant={eventID === event.id ? "default" : "outline"} onClick={() => setEventID(event.id)}>{event.title} · {event.durationMinutes} min</Button>)}</div>}
      {info && !info.events.length && <p>No meeting types have been published yet.</p>}
      {eventID && !picked && <Card className="space-y-4 p-6">
        <p className="text-sm text-muted-foreground">Times shown in {timezone}</p>
        <div className="flex justify-between"><Button variant="outline" disabled={offset === 0 || loading} onClick={() => setOffset(value => value - 7)}>Previous week</Button><Button variant="outline" disabled={offset >= 84 || loading} onClick={() => setOffset(value => value + 7)}>Next week</Button></div>
        {loading ? <p>Checking availability…</p> : slots.length ? <div className="grid max-h-96 grid-cols-2 gap-2 overflow-auto">{slots.map(slot => <Button variant="outline" key={slot.start} onClick={() => setPicked(slot)}>{new Date(slot.start).toLocaleString(undefined, { month: "short", day: "numeric", hour: "2-digit", minute: "2-digit" })}</Button>)}</div> : !error && <p>No open times this week.</p>}
      </Card>}
      {picked && <Card className="p-6"><form onSubmit={book} className="space-y-4">
        <h2 className="text-xl font-medium">Your booking details</h2><p>{new Date(picked.start).toLocaleString()}</p>
        <div><Label htmlFor="guest-name">Your name</Label><Input id="guest-name" value={name} onChange={event => setName(event.target.value)} required maxLength={200} /></div>
        <div><Label htmlFor="guest-email">Email for your meeting invitation</Label><Input id="guest-email" type="email" value={email} onChange={event => setEmail(event.target.value)} required /></div>
        <div><Label htmlFor="guest-notes">What would you like to discuss?</Label><textarea id="guest-notes" className="min-h-28 w-full rounded border bg-background p-3" value={notes} onChange={event => setNotes(event.target.value)} required maxLength={10000} /></div>
        <div className="flex gap-3"><Button type="button" variant="outline" disabled={busy} onClick={() => setPicked(null)}>Back</Button><Button disabled={busy} type="submit">{busy ? "Booking…" : "Book meeting"}</Button></div>
      </form></Card>}
    </>}
  </div>;
}

export default function BookingPage() {
  return <Suspense fallback={<p className="p-10">Loading…</p>}><Booking /></Suspense>;
}
